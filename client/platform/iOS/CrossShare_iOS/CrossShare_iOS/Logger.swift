//
//  Logger.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/5/30.
//

import Foundation
import os.log

class Logger {
    static let shared = Logger()
    private let fileManager = FileManager.default
    private let logFileURL: URL
    private let queue = DispatchQueue(label: "com.crossshare.logger.queue", qos: .utility)
    private let dateFormatter: DateFormatter
    
    #if !EXTENSION
    private var stdoutPipe: Pipe?
    private var stderrPipe: Pipe?
    private var originalStdout: Int32 = -1
    private var originalStderr: Int32 = -1
    private var isRedirecting = false
    #endif
    
    private init() {
        let documentsDirectory = fileManager.urls(for: .documentDirectory, in: .userDomainMask).first!
        let logDirectory = documentsDirectory.appendingPathComponent("Log")
        
        if !fileManager.fileExists(atPath: logDirectory.path) {
            try? fileManager.createDirectory(at: logDirectory, withIntermediateDirectories: true)
        }
        
        let fileDateFormatter = DateFormatter()
        fileDateFormatter.dateFormat = "yyyy-MM-dd"
        let dateString = fileDateFormatter.string(from: Date())
        logFileURL = logDirectory.appendingPathComponent("log-\(dateString).txt")
        
        dateFormatter = DateFormatter()
        dateFormatter.dateFormat = "yyyy-MM-dd HH:mm:ss.SSS"
        
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
