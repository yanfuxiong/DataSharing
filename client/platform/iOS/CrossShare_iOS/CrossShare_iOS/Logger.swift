//
//  Logger.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/5/30.
//

import Foundation
import os.log
import SWCompression

class Logger {
    static let shared = Logger()
    private let fileManager = FileManager.default
    private var logFileURL: URL
    private let logDirectory: URL
    private let queue = DispatchQueue(label: "com.crossshare.logger.queue", qos: .utility)
    private let dateFormatter: DateFormatter
    private let maxLogFileSize: UInt64 = 5 * 1024 * 1024
    private let maxLogDays: Int = 3
    private let maxLogFiles: Int = 3
    
    #if !EXTENSION
    private var stdoutPipe: Pipe?
    private var stderrPipe: Pipe?
    private var originalStdout: Int32 = -1
    private var originalStderr: Int32 = -1
    private var isRedirecting = false
    #endif
    
    private init() {
        let documentsDirectory = fileManager.urls(for: .documentDirectory, in: .userDomainMask).first!
        logDirectory = documentsDirectory.appendingPathComponent("Log")
        
        if !fileManager.fileExists(atPath: logDirectory.path) {
            try? fileManager.createDirectory(at: logDirectory, withIntermediateDirectories: true)
        }
        
        let fileDateFormatter = DateFormatter()
        fileDateFormatter.dateFormat = "yyyy-MM-dd"
        let dateString = fileDateFormatter.string(from: Date())
        logFileURL = logDirectory.appendingPathComponent("log-\(dateString).txt")
        
        dateFormatter = DateFormatter()
        dateFormatter.dateFormat = "yyyy-MM-dd HH:mm:ss.SSS"
        
        compressOldTxtFiles()
        
        cleanupOldLogs()
        
        log("Logger initialized", type: .info)
        
        #if !EXTENSION
        setupOSLogCapture()
        #endif
    }
    
    deinit {
        #if !EXTENSION
        if isRedirecting {
            cleanupPipes()
        }
        #endif
    }
    
    func log(_ message: String, type: LogType = .info, file: String = #file, function: String = #function, line: Int = #line) {
        let fileName = URL(fileURLWithPath: file).lastPathComponent
        let logMessage = "[\(dateFormatter.string(from: Date()))] [\(type.rawValue)] [\(fileName):\(line)] \(function): \(message)"
        
        print(logMessage)
        
        queue.async { [weak self] in
            guard let self = self else { return }
            
            self.rotateLogIfNeeded()
            
            if let data = (logMessage + "\n").data(using: .utf8) {
                if self.fileManager.fileExists(atPath: self.logFileURL.path) {
                    if let fileHandle = try? FileHandle(forWritingTo: self.logFileURL) {
                        fileHandle.seekToEndOfFile()
                        fileHandle.write(data)
                        fileHandle.closeFile()
                    }
                } else {
                    try? data.write(to: self.logFileURL, options: .atomic)
                }
            }
        }
    }
    
    func exportLogs() -> URL {
        return logFileURL
    }
    
    enum LogType: String {
        case debug = "DEBUG"
        case info = "INFO"
        case warning = "WARNING"
        case error = "ERROR"
        case goLog = "GO_LOG"
    }
    
    #if !EXTENSION
    private func setupOSLogCapture() {
        let subsystem = Bundle.main.bundleIdentifier ?? "com.thundersoft.crossshare.ios"
        let log = OSLog(subsystem: subsystem, category: "Default")
        
        os_log("OSLog capture initialized", log: log, type: .info)
    }
    
    private func cleanupPipes() {
        if originalStdout >= 0 {
            dup2(originalStdout, STDOUT_FILENO)
            close(originalStdout)
            originalStdout = -1
        }
        
        if originalStderr >= 0 {
            dup2(originalStderr, STDERR_FILENO)
            close(originalStderr)
            originalStderr = -1
        }
        
        stdoutPipe?.fileHandleForReading.readabilityHandler = nil
        stderrPipe?.fileHandleForReading.readabilityHandler = nil
        
        stdoutPipe = nil
        stderrPipe = nil
        
        isRedirecting = false
    }
    #endif
    
    func logFromGo(_ message: String) {
        log(message, type: .goLog)
    }
    
    private func rotateLogIfNeeded() {
        let fileDateFormatter = DateFormatter()
        fileDateFormatter.dateFormat = "yyyy-MM-dd"
        let currentDateString = fileDateFormatter.string(from: Date())
        let expectedLogFileURL = logDirectory.appendingPathComponent("log-\(currentDateString).txt")
        
        let dateChanged = logFileURL.path != expectedLogFileURL.path
        
        var fileSizeExceeded = false
        if fileManager.fileExists(atPath: logFileURL.path),
           let attributes = try? fileManager.attributesOfItem(atPath: logFileURL.path),
           let fileSize = attributes[.size] as? UInt64 {
            fileSizeExceeded = fileSize >= maxLogFileSize
        }
        
        if dateChanged || fileSizeExceeded {
            if dateChanged {
                compressOldTxtFiles()
                logFileURL = expectedLogFileURL
            } else {
                if fileManager.fileExists(atPath: logFileURL.path) {
                    compressLogFile(logFileURL)
                }
                let timestamp = Int(Date().timeIntervalSince1970)
                let baseName = logFileURL.deletingPathExtension().lastPathComponent
                let newFileName = "\(baseName)-\(timestamp).txt"
                logFileURL = logDirectory.appendingPathComponent(newFileName)
            }
            
            cleanupOldLogs()
        }
    }

    private func compressOldTxtFiles() {
        guard let files = try? fileManager.contentsOfDirectory(at: logDirectory, includingPropertiesForKeys: nil, options: [.skipsHiddenFiles]) else {
            return
        }
        
        let fileDateFormatter = DateFormatter()
        fileDateFormatter.dateFormat = "yyyy-MM-dd"
        let todayString = fileDateFormatter.string(from: Date())
        
        let oldTxtFiles = files.filter { url in
            url.pathExtension == "txt" && !url.lastPathComponent.contains(todayString)
        }
        
        for fileURL in oldTxtFiles {
            compressLogFile(fileURL)
        }
    }
    
    private func compressLogFile(_ fileURL: URL) {
        let fileName = fileURL.lastPathComponent
        let tarGzFileName = fileName.replacingOccurrences(of: ".txt", with: ".tar.gz")
        let tarGzFileURL = logDirectory.appendingPathComponent(tarGzFileName)
        
        if fileManager.fileExists(atPath: tarGzFileURL.path) {
            try? fileManager.removeItem(at: tarGzFileURL)
        }
        
        do {
            let inputData = try Data(contentsOf: fileURL)
            let entry = TarEntry(info: TarEntryInfo(name: fileName, type: .regular),
                                data: inputData)
            let tarData = try TarContainer.create(from: [entry])
            let gzipData = try GzipArchive.archive(data: tarData)
            try gzipData.write(to: tarGzFileURL, options: .atomic)
            try fileManager.removeItem(at: fileURL)
            print("Log file compressed: \(tarGzFileName)")
        } catch {
            print("Failed to compress log file: \(error.localizedDescription)")
        }
    }
    
    private func cleanupOldLogs() {
        guard let files = try? fileManager.contentsOfDirectory(at: logDirectory, includingPropertiesForKeys: [.creationDateKey], options: [.skipsHiddenFiles]) else {
            return
        }
        
        let currentDate = Date()
        let calendar = Calendar.current
        
        var logFiles = files.filter { url in
            let pathExtension = url.pathExtension
            return pathExtension == "txt" || pathExtension == "gz"
        }
        
        logFiles = logFiles.filter { url in
            if let creationDate = try? url.resourceValues(forKeys: [.creationDateKey]).creationDate {
                let daysDifference = calendar.dateComponents([.day], from: creationDate, to: currentDate).day ?? 0
                if daysDifference >= maxLogDays {
                    try? fileManager.removeItem(at: url)
                    print("Deleted old log file: \(url.lastPathComponent) (age: \(daysDifference) days)")
                    return false
                }
            }
            return true
        }
        
        logFiles.sort { url1, url2 in
            let date1 = try? url1.resourceValues(forKeys: [.creationDateKey]).creationDate
            let date2 = try? url2.resourceValues(forKeys: [.creationDateKey]).creationDate
            return (date1 ?? Date.distantPast) > (date2 ?? Date.distantPast)
        }
        
        if logFiles.count > maxLogFiles {
            let filesToDelete = logFiles.suffix(from: maxLogFiles)
            for fileURL in filesToDelete {
                try? fileManager.removeItem(at: fileURL)
                print("Deleted excess log file: \(fileURL.lastPathComponent)")
            }
        }
    }
}

extension Logger {
    static func debug(_ message: String, file: String = #file, function: String = #function, line: Int = #line) {
        shared.log(message, type: .debug, file: file, function: function, line: line)
    }
    
    static func info(_ message: String, file: String = #file, function: String = #function, line: Int = #line) {
        shared.log(message, type: .info, file: file, function: function, line: line)
    }
    
    static func warning(_ message: String, file: String = #file, function: String = #function, line: Int = #line) {
        shared.log(message, type: .warning, file: file, function: function, line: line)
    }
    
    static func error(_ message: String, file: String = #file, function: String = #function, line: Int = #line) {
        shared.log(message, type: .error, file: file, function: function, line: line)
    }
    
    static func goLog(_ message: String) {
        shared.logFromGo(message)
    }
}
