//
//  XPCLogger.swift
//  CrossShareHelper
//
//  Created by TS on 2025/8/27.
//  XPC 服务日志管理器
//

import Foundation
import os.log

/// XPC 服务日志管理器
class XPCLogger {
    
    static let shared = XPCLogger()
    
    // MARK: - Properties
    
    private let logger = OSLog(subsystem: "com.instance.crossshare", category: "XPCService")
    private let queue = DispatchQueue(label: "com.crossshare.logger.queue")
    private var recentLogs: [String] = []
    private let maxLogCount = 1000
    private var currentLogLevel: LogLevel = .info
    
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
    
    // MARK: - Initialization
    
    private init() {
        setupLogFile()
    }
    
    private func setupLogFile() {
        // 创建日志目录
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
    
    // MARK: - Public Methods
    
    /// 记录日志
    /// - Parameters:
    ///   - message: 日志消息
    ///   - level: 日志级别
    ///   - file: 文件名
    ///   - function: 函数名
    ///   - line: 行号
    func log(_ message: String, 
             level: LogLevel = .info,
             file: String = #file,
             function: String = #function,
             line: Int = #line) {
        
        // 检查日志级别过滤
        guard level >= currentLogLevel else { return }
        
        let fileName = URL(fileURLWithPath: file).lastPathComponent
        let timestamp = ISO8601DateFormatter().string(from: Date())
        let logEntry = "\(timestamp) \(level.prefix) [\(fileName):\(line)] \(function) - \(message)"
        
        queue.async { [weak self] in
            // 写入系统日志
            os_log("%{public}@", log: self?.logger ?? OSLog.default, type: level.osLogType, logEntry)
            
            // 保存到内存中的最近日志
            self?.recentLogs.append(logEntry)
            if let count = self?.recentLogs.count, count > self?.maxLogCount ?? 1000 {
                self?.recentLogs.removeFirst(count - (self?.maxLogCount ?? 1000))
            }
            
            // 写入日志文件
            self?.writeToFile(logEntry)
        }
    }
    
    /// 设置日志级别
    /// - Parameter level: 新的日志级别
    func setLogLevel(level: Int) {
        guard let newLevel = LogLevel(rawValue: level) else { return }
        currentLogLevel = newLevel
        log("Log level changed to \(newLevel.prefix)", level: .info)
    }
    
    /// 获取最近的日志
    /// - Parameter lines: 获取的行数
    /// - Returns: 日志数组
    func getRecentLogs(lines: Int) -> [String] {
        return queue.sync {
            let startIndex = max(0, recentLogs.count - lines)
            return Array(recentLogs[startIndex...])
        }
    }
    
    /// 清空日志
    func clearLogs() {
        queue.async { [weak self] in
            self?.recentLogs.removeAll()
        }
    }
    
    // MARK: - Private Methods
    
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

// MARK: - 便利方法扩展

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
