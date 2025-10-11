//
//  FileLogger.swift
//  CrossShareHelper
//
//  Created by Assistant on 2025/9/11.
//  文件日志记录器
//

import Foundation

class FileLogger {
    private let logFileURL: URL
    private let fileManager = FileManager.default
    private let queue = DispatchQueue(label: "com.crossshare.filelogger", qos: .background)
    private var fileHandle: FileHandle?
    
    init() {
        let appSupport = fileManager.urls(for: .applicationSupportDirectory, in: .userDomainMask).first!
        let logDir = appSupport.appendingPathComponent("CrossShare/Logs/Helper")
        
        try? fileManager.createDirectory(at: logDir, withIntermediateDirectories: true)
        
        let dateFormatter = DateFormatter()
        dateFormatter.dateFormat = "yyyy-MM-dd"
        dateFormatter.timeZone = TimeZone(secondsFromGMT: 8 * 3600) // GMT+8
        let dateStr = dateFormatter.string(from: Date())
        let fileName = "helper-\(dateStr).log"
        
        logFileURL = logDir.appendingPathComponent(fileName)
        
        if !fileManager.fileExists(atPath: logFileURL.path) {
            fileManager.createFile(atPath: logFileURL.path, contents: nil, attributes: nil)
        }
        
        print("FileLogger: Creating log file at \(logFileURL.path)")
        
        do {
            fileHandle = try FileHandle(forWritingTo: logFileURL)
            fileHandle?.seekToEndOfFile()
            
            let startFormatter = DateFormatter()
            startFormatter.dateFormat = "yyyy-MM-dd HH:mm:ss"
            startFormatter.timeZone = TimeZone(secondsFromGMT: 8 * 3600) // GMT+8
            let startTime = startFormatter.string(from: Date())
            let startMessage = "\n\n========== Helper Started at \(startTime) ==========\n\n"
            if let data = startMessage.data(using: .utf8) {
                fileHandle?.write(data)
                fileHandle?.synchronizeFile()
            }
            print("FileLogger: Successfully opened log file")
        } catch {
            print("FileLogger: Failed to open log file: \(error)")
            let fallbackFormatter = DateFormatter()
            fallbackFormatter.dateFormat = "yyyy-MM-dd HH:mm:ss"
            fallbackFormatter.timeZone = TimeZone(secondsFromGMT: 8 * 3600) // GMT+8
            let fallbackTime = fallbackFormatter.string(from: Date())
            let startMessage = "\n\n========== Helper Started at \(fallbackTime) ==========\n\n"
            try? startMessage.write(to: logFileURL, atomically: true, encoding: .utf8)
        }
    }
    
    deinit {
        fileHandle?.closeFile()
    }
    
    func write(_ message: String, level: XPCLogger.LogLevel, file: String, function: String, line: Int) {
        queue.async { [weak self] in
            guard let self = self, let handle = self.fileHandle else { return }
            
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
    
    func getRecentLogs(lines: Int = 100) -> [String] {
        do {
            let content = try String(contentsOf: logFileURL, encoding: .utf8)
            let allLines = content.components(separatedBy: .newlines)
            let startIndex = max(0, allLines.count - lines)
            return Array(allLines[startIndex..<allLines.count])
        } catch {
            return ["Error reading log file: \(error)"]
        }
    }
}
