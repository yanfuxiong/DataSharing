//
//  XPCLogger.swift
//  CrossShareHelper
//
//  Created by TS on 2025/8/27.
//  XPC Service Logger
//

import Foundation
import os.log

class XPCLogger {
    
    static let shared = XPCLogger()
    
    private let logger = OSLog(subsystem: "com.instance.crossshare", category: "XPCService")
    private let queue = DispatchQueue(label: "com.crossshare.logger.queue")
    private var recentLogs: [String] = []
    private let maxLogCount = 1000
    private var currentLogLevel: LogLevel = .info
    private let fileLogger: FileLogger
    
    enum LogLevel: Int, Comparable {
        case debug = 0
        case info = 1
        case warn = 2
        case error = 3
        case fatal = 4
        
        static func < (lhs: LogLevel, rhs: LogLevel) -> Bool {
            return lhs.rawValue < rhs.rawValue
        }
        
        var prefix: String {
            switch self {
            case .debug: return "[DEBUG]"
            case .info: return "[INFO]"
            case .warn: return "[WARN]"
            case .error: return "[ERROR]"
            case .fatal: return "[FATAL]"
            }
        }
        
        var osLogType: OSLogType {
            switch self {
            case .debug: return .debug
            case .info: return .info
            case .warn: return .default
            case .error: return .error
            case .fatal: return .fault
            }
        }
    }
    
    private init() {
        fileLogger = FileLogger()
        setupLogFile()
    }
    
    private func setupLogFile() {
        let fileManager = FileManager.default
        guard let appSupport = fileManager.urls(for: .applicationSupportDirectory, in: .userDomainMask).first else {
            return
        }
        
        let logDir = appSupport.appendingPathComponent("CrossShare/Logs/Helper")
        do {
            try fileManager.createDirectory(at: logDir, withIntermediateDirectories: true)
        } catch {
            os_log("Failed to create log directory: %{public}@", log: logger, type: .error, error.localizedDescription)
        }
    }
    
    func log(_ message: String, 
             level: LogLevel = .info,
             file: String = #file,
             function: String = #function,
             line: Int = #line) {
        
        guard level >= currentLogLevel else { return }
        
        let fileName = URL(fileURLWithPath: file).lastPathComponent
        let timestamp = ISO8601DateFormatter().string(from: Date())
        let logEntry = "\(timestamp) \(level.prefix) [\(fileName):\(line)] \(function) - \(message)"
        
        // 写入文件日志
        fileLogger.write(message, level: level, file: file, function: function, line: line)
        
        queue.async { [weak self] in
            os_log("%{public}@", log: self?.logger ?? OSLog.default, type: level.osLogType, logEntry)
            self?.recentLogs.append(logEntry)
            if let count = self?.recentLogs.count, count > self?.maxLogCount ?? 1000 {
                self?.recentLogs.removeFirst(count - (self?.maxLogCount ?? 1000))
            }
            self?.writeToFile(logEntry)
        }
    }
    
    func setLogLevel(level: Int) {
        guard let newLevel = LogLevel(rawValue: level) else { return }
        currentLogLevel = newLevel
        log("Log level changed to \(newLevel.prefix)", level: .info)
    }
    
    func getRecentLogs(lines: Int) -> [String] {
        return queue.sync {
            let startIndex = max(0, recentLogs.count - lines)
            return Array(recentLogs[startIndex...])
        }
    }
    
    func clearLogs() {
        queue.async { [weak self] in
            self?.recentLogs.removeAll()
        }
    }
    
    private func writeToFile(_ logEntry: String) {
        let fileManager = FileManager.default
        guard let appSupport = fileManager.urls(for: .applicationSupportDirectory, in: .userDomainMask).first else {
            return
        }
        
        let logFile = appSupport
            .appendingPathComponent("CrossShare/Logs/Helper")
            .appendingPathComponent("xpc-service.log")
        
        let data = "\(logEntry)\n".data(using: .utf8) ?? Data()
        
        if fileManager.fileExists(atPath: logFile.path) {
            if let fileHandle = try? FileHandle(forWritingTo: logFile) {
                fileHandle.seekToEndOfFile()
                fileHandle.write(data)
                fileHandle.closeFile()
            }
        } else {
            try? data.write(to: logFile)
        }
    }
}

extension XPCLogger {
    
    func debug(_ message: String, file: String = #file, function: String = #function, line: Int = #line) {
        log(message, level: .debug, file: file, function: function, line: line)
    }
    
    func info(_ message: String, file: String = #file, function: String = #function, line: Int = #line) {
        log(message, level: .info, file: file, function: function, line: line)
    }
    
    func warn(_ message: String, file: String = #file, function: String = #function, line: Int = #line) {
        log(message, level: .warn, file: file, function: function, line: line)
    }
    
    func error(_ message: String, file: String = #file, function: String = #function, line: Int = #line) {
        log(message, level: .error, file: file, function: function, line: line)
    }
    
    func fatal(_ message: String, file: String = #file, function: String = #function, line: Int = #line) {
        log(message, level: .fatal, file: file, function: function, line: line)
    }
}
