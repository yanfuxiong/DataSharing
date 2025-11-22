//
//  CSLogger.swift
//  CrossShareHelper
//
//  Created by TS on 2025/8/27.
//  XPC Service Logger
//

import Foundation
import os.log

class CSLogger {
    
    // 使用静态变量存储单例,支持配置
    private static var _shared: CSLogger?
    static var shared: CSLogger {
        if _shared == nil {
            // 默认为 Helper 进程
            _shared = CSLogger(processName: "Helper")
        }
        return _shared!
    }
    
    // 配置方法：在应用启动时调用
    static func configure(processName: String) {
        _shared = CSLogger(processName: processName)
    }
    
    private let processName: String
    private let logger: OSLog
    private let queue = DispatchQueue(label: "com.crossshare.logger.queue")
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
    
    private init(processName: String) {
        self.processName = processName
        // 根据进程名设置不同的 category
        self.logger = OSLog(subsystem: "com.instance.crossshare", category: processName)
        // 创建 FileLogger 时传入进程名
        self.fileLogger = FileLogger(processName: processName)
        setupLogFile()
    }
    
    private func setupLogFile() {
        let logDir = getLogPath()
        do {
            try FileManager.default.createDirectory(at: logDir, withIntermediateDirectories: true)
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
        let logEntry = "\(timestamp) \(level.prefix) [\(processName)] [\(fileName):\(line)] \(function) - \(message)"
        
        // write to log files
        fileLogger.write(message, level: level, file: file, function: function, line: line)

        queue.async { [weak self] in
            guard let self = self else { return }
            // output to OSLog
            os_log("%{public}@", log: self.logger, type: level.osLogType, logEntry)
        }
    }
    
    func setLogLevel(level: Int) {
        guard let newLevel = LogLevel(rawValue: level) else { return }
        currentLogLevel = newLevel
        log("Log level changed to \(newLevel.prefix)", level: .info)
    }
    

    
    func getLogFilePath() -> String {
        return fileLogger.getLogFilePath()
    }
    
}

extension CSLogger {
    
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
