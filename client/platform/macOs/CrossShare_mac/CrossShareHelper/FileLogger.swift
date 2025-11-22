//
//  FileLogger.swift
//  CrossShareHelper
//
//  Created by Assistant on 2025/9/11.
//  File Logger
//

import Foundation

class FileLogger {
    private var logFileURL: URL
    private let fileManager = FileManager.default
    private let queue = DispatchQueue(label: "com.crossshare.filelogger", qos: .background)
    private var fileHandle: FileHandle?
    private let maxLogFileSize: UInt64 = 5 * 1024 * 1024 // 5MB
    private let maxLogFiles = 3 // Maximum 3 log files to keep
    private let maxLogDays = 3 // Keep logs for 3 days
    private var currentDate: String
    private let filePrefix: String // "app" or "helper"
    
    init(processName: String) {
        // 根据进程名确定文件前缀
        self.filePrefix = processName.lowercased()
        let logDir = getLogPath()
        
        try? fileManager.createDirectory(at: logDir, withIntermediateDirectories: true)
        
        let dateFormatter = DateFormatter()
        dateFormatter.dateFormat = "yyyy-MM-dd"
        dateFormatter.timeZone = TimeZone(secondsFromGMT: 8 * 3600) // GMT+8
        currentDate = dateFormatter.string(from: Date())
        let fileName = "\(filePrefix)-\(currentDate).log"
        
        logFileURL = logDir.appendingPathComponent(fileName)
        
        // Compress old uncompressed log files before creating new log
        compressOldUncompressedLogs(currentDate: currentDate, logDir: logDir)
        
        if !fileManager.fileExists(atPath: logFileURL.path) {
            fileManager.createFile(atPath: logFileURL.path, contents: nil, attributes: nil)
        }
        
        print("[\(filePrefix.uppercased())] FileLogger: Creating log file at \(logFileURL.path)")
        
        openLogFile()
        cleanupOldLogs()
    }
    
    deinit {
        fileHandle?.closeFile()
    }
    
    // MARK: - Private Methods
    
    private func openLogFile() {
        do {
            fileHandle = try FileHandle(forWritingTo: logFileURL)
            fileHandle?.seekToEndOfFile()
            
            let startFormatter = DateFormatter()
            startFormatter.dateFormat = "yyyy-MM-dd HH:mm:ss"
            startFormatter.timeZone = TimeZone(secondsFromGMT: 8 * 3600) // GMT+8
            let startTime = startFormatter.string(from: Date())
            let processName = filePrefix.capitalized
            let startMessage = "\n\n========== \(processName) Process Started at \(startTime) ==========\n\n"
            if let data = startMessage.data(using: .utf8) {
                fileHandle?.write(data)
                fileHandle?.synchronizeFile()
            }
            print("[\(filePrefix.uppercased())] FileLogger: Successfully opened log file")
        } catch {
            print("[\(filePrefix.uppercased())] FileLogger: Failed to open log file: \(error)")
            let fallbackFormatter = DateFormatter()
            fallbackFormatter.dateFormat = "yyyy-MM-dd HH:mm:ss"
            fallbackFormatter.timeZone = TimeZone(secondsFromGMT: 8 * 3600) // GMT+8
            let fallbackTime = fallbackFormatter.string(from: Date())
            let processName = filePrefix.capitalized
            let startMessage = "\n\n========== \(processName) Process Started at \(fallbackTime) ==========\n\n"
            try? startMessage.write(to: logFileURL, atomically: true, encoding: .utf8)
        }
    }
    
    private func checkAndRotateLog() {
        // Check if date has changed
        let dateFormatter = DateFormatter()
        dateFormatter.dateFormat = "yyyy-MM-dd"
        dateFormatter.timeZone = TimeZone(secondsFromGMT: 8 * 3600) // GMT+8
        let today = dateFormatter.string(from: Date())
        
        if today != currentDate {
            // Date changed, compress old log and create new log file
            print("[\(filePrefix.uppercased())] FileLogger: Date changed, rotating log file")
            compressCurrentLog(withTimestamp: false)
            createNewLogFile(forDate: today)
            return
        }
        
        // Check file size
        guard let attributes = try? fileManager.attributesOfItem(atPath: logFileURL.path),
              let fileSize = attributes[.size] as? UInt64 else {
            return
        }
        
        if fileSize >= maxLogFileSize {
            // Log file size exceeded limit, compressing and creating new log
            print("[\(filePrefix.uppercased())] FileLogger: Log file size exceeded \(maxLogFileSize) bytes, rotating")
            compressCurrentLog(withTimestamp: true)
            createNewLogFile(forDate: today)
        }
    }
    
    private func compressCurrentLog(withTimestamp: Bool) {
        fileHandle?.closeFile()
        fileHandle = nil
        
        compressLogFile(at: logFileURL, withTimestamp: withTimestamp)
    }
    
    private func createNewLogFile(forDate date: String) {
        currentDate = date
        let fileName = "\(filePrefix)-\(date).log"
        logFileURL = logFileURL.deletingLastPathComponent().appendingPathComponent(fileName)
        
        // Create new log file
        if !fileManager.fileExists(atPath: logFileURL.path) {
            fileManager.createFile(atPath: logFileURL.path, contents: nil, attributes: nil)
        }
        
        print("[\(filePrefix.uppercased())] FileLogger: Created new log file at \(logFileURL.path)")
        openLogFile()
        cleanupOldLogs()
    }
    
    private func compressOldUncompressedLogs(currentDate: String, logDir: URL) {
        guard let fileURLs = try? fileManager.contentsOfDirectory(
            at: logDir,
            includingPropertiesForKeys: nil,
            options: .skipsHiddenFiles
        ) else {
            return
        }
        
        let dateFormatter = DateFormatter()
        dateFormatter.dateFormat = "yyyy-MM-dd"
        dateFormatter.timeZone = TimeZone(secondsFromGMT: 8 * 3600) // GMT+8
        
        for fileURL in fileURLs {
            let fileName = fileURL.lastPathComponent
            
            // Skip directories
            var isDirectory: ObjCBool = false
            if fileManager.fileExists(atPath: fileURL.path, isDirectory: &isDirectory), isDirectory.boolValue {
                continue
            }
            
            // Only process uncompressed log files matching current prefix (e.g., app-*.log or helper-*.log)
            guard fileName.hasPrefix("\(filePrefix)-") && fileName.hasSuffix(".log") else {
                continue
            }
            
            // Extract date from filename (prefix-yyyy-MM-dd.log)
            let prefixLength = filePrefix.count + 1
            let nameWithoutPrefix = String(fileName.dropFirst(prefixLength))
            let datePart = String(nameWithoutPrefix.prefix(10)) // yyyy-MM-dd
            
            // If this log file is from a different date (older), compress it without timestamp
            if datePart != currentDate {
                print("[\(filePrefix.uppercased())] FileLogger: Found old uncompressed log file: \(fileName), compressing...")
                compressLogFile(at: fileURL, withTimestamp: false)
            }
        }
    }
    
    private func compressLogFile(at fileURL: URL, withTimestamp: Bool) {
        // Generate filename based on whether we need timestamp
        let compressedFileName: String
        if withTimestamp {
            // Use HHmmss format for same-day rotation (e.g. app-2025-10-22-143130.tar.gz)
            let timeFormatter = DateFormatter()
            timeFormatter.dateFormat = "HHmmss"
            timeFormatter.timeZone = TimeZone(secondsFromGMT: 8 * 3600) // GMT+8
            let timeString = timeFormatter.string(from: Date())
            compressedFileName = fileURL.deletingPathExtension().lastPathComponent + "-\(timeString).tar.gz"
        } else {
            compressedFileName = fileURL.deletingPathExtension().lastPathComponent + ".tar.gz"
        }
        
        let compressedFileURL = fileURL.deletingLastPathComponent().appendingPathComponent(compressedFileName)
        
        // If compressed file already exists, delete it first (only for non-timestamp version)
        if !withTimestamp && fileManager.fileExists(atPath: compressedFileURL.path) {
            try? fileManager.removeItem(at: compressedFileURL)
        }
        
        // Using tar command to compress log file
        let process = Process()
        process.executableURL = URL(fileURLWithPath: "/usr/bin/tar")
        process.arguments = [
            "-czf",
            compressedFileURL.path,
            "-C",
            fileURL.deletingLastPathComponent().path,
            fileURL.lastPathComponent
        ]
        
        do {
            try process.run()
            process.waitUntilExit()
            
            if process.terminationStatus == 0 {
                print("[\(filePrefix.uppercased())] FileLogger: Successfully compressed log to \(compressedFileName)")
                // Delete original log file
                try? fileManager.removeItem(at: fileURL)
            } else {
                print("[\(filePrefix.uppercased())] FileLogger: Failed to compress log file: \(fileURL.lastPathComponent)")
            }
        } catch {
            print("[\(filePrefix.uppercased())] FileLogger: Error compressing log: \(error)")
        }
    }
    
    private func cleanupOldLogs() {
        let logDir = logFileURL.deletingLastPathComponent()
        
        guard let fileURLs = try? fileManager.contentsOfDirectory(
            at: logDir,
            includingPropertiesForKeys: [.creationDateKey],
            options: .skipsHiddenFiles
        ) else {
            return
        }
        
        let dateFormatter = DateFormatter()
        dateFormatter.dateFormat = "yyyy-MM-dd"
        dateFormatter.timeZone = TimeZone(secondsFromGMT: 8 * 3600) // GMT+8
        
        let calendar = Calendar.current
        let today = calendar.startOfDay(for: Date())
        
        // Calculate cutoff date: keep only logs within maxLogDays (e.g. today, yesterday, day before yesterday)
        // If maxLogDays = 3, cutoff = today - 2 days, so we keep today, today-1, today-2
        guard let cutoffDate = calendar.date(byAdding: .day, value: -(maxLogDays - 1), to: today) else {
            return
        }
        
        var logFiles: [(url: URL, date: Date)] = []
        
        for fileURL in fileURLs {
            let fileName = fileURL.lastPathComponent
            
            // Skip directories
            var isDirectory: ObjCBool = false
            if fileManager.fileExists(atPath: fileURL.path, isDirectory: &isDirectory), isDirectory.boolValue {
                continue
            }
            
            // Only process files with matching prefix and ending with '.log' or '.tar.gz'
            guard fileName.hasPrefix("\(filePrefix)-") &&
                  (fileName.hasSuffix(".log") || fileName.hasSuffix(".tar.gz")) else {
                continue
            }
            
            var fileDate: Date?
            
            // Try parsing date from filename
            if fileName.hasPrefix("\(filePrefix)-") {
                // Extract date part (prefix-yyyy-MM-dd.log or prefix-yyyy-MM-dd.tar.gz)
                let prefixLength = filePrefix.count + 1
                let nameWithoutPrefix = String(fileName.dropFirst(prefixLength))
                let datePart = String(nameWithoutPrefix.prefix(10)) // yyyy-MM-dd
                if let parsedDate = dateFormatter.date(from: datePart) {
                    // Use start of day for date comparison
                    fileDate = calendar.startOfDay(for: parsedDate)
                }
            }
            
            // If date parsing from filename fails, use creation date
            if fileDate == nil {
                if let attributes = try? fileManager.attributesOfItem(atPath: fileURL.path),
                   let creationDate = attributes[.creationDate] as? Date {
                    fileDate = calendar.startOfDay(for: creationDate)
                }
            }
            
            if let date = fileDate {
                // Delete logs older than cutoff date (outside retention period)
                if date < cutoffDate {
                    print("[\(filePrefix.uppercased())] FileLogger: Deleting old log file (outside \(maxLogDays) days): \(fileName)")
                    try? fileManager.removeItem(at: fileURL)
                } else {
                    logFiles.append((url: fileURL, date: date))
                }
            }
        }
        
        // Sort by date in descending order (newest first)
        logFiles.sort { $0.date > $1.date }
        
        // Keep only maxLogFiles (delete excess files beyond the limit)
        if logFiles.count > maxLogFiles {
            let filesToDelete = logFiles.suffix(from: maxLogFiles)
            for file in filesToDelete {
                print("[\(filePrefix.uppercased())] FileLogger: Deleting excess log file (beyond \(maxLogFiles) files limit): \(file.url.lastPathComponent)")
                try? fileManager.removeItem(at: file.url)
            }
        }
    }
    
    // MARK: - Public Methods
    
    func write(_ message: String, level: CSLogger.LogLevel, file: String, function: String, line: Int) {
        queue.async { [weak self] in
            guard let self = self else { return }
            
            // Check if log rotation is needed
            self.checkAndRotateLog()
            
            guard let handle = self.fileHandle else { return }
            
            let dateFormatter = DateFormatter()
            dateFormatter.dateFormat = "yyyy-MM-dd HH:mm:ss"
            dateFormatter.timeZone = TimeZone(secondsFromGMT: 8 * 3600) // GMT+8
            let timestamp = dateFormatter.string(from: Date())
            
            let fileName = URL(fileURLWithPath: file).lastPathComponent
            let logEntry = "\(timestamp) \(level.prefix) [\(fileName):\(line)] \(function) - \(message)\n"
            
            if let data = logEntry.data(using: .utf8) {
                handle.write(data)
                handle.synchronizeFile()
            }
        }
    }
    
    func getLogFilePath() -> String {
        return logFileURL.path
    }
}
