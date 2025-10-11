//
//  HelperLogViewer.swift
//  CrossShare
//
//  Created by Assistant on 2025/9/11.
//

import Foundation
import AppKit

class HelperLogViewer {
    static let shared = HelperLogViewer()
    
    private init() {}
    
    func getLogPath() -> String {
        let appSupport = FileManager.default.urls(for: .applicationSupportDirectory, in: .userDomainMask).first!
        let logDir = appSupport.appendingPathComponent("CrossShare/Logs/Helper")
        return logDir.path
    }
    
    func getTodayLogFile() -> URL? {
        let appSupport = FileManager.default.urls(for: .applicationSupportDirectory, in: .userDomainMask).first!
        let logDir = appSupport.appendingPathComponent("CrossShare/Logs/Helper")
        
        let dateFormatter = DateFormatter()
        dateFormatter.dateFormat = "yyyy-MM-dd"
        dateFormatter.timeZone = TimeZone(secondsFromGMT: 8 * 3600) // GMT+8
        let dateStr = dateFormatter.string(from: Date())
        let fileName = "helper-\(dateStr).log"
        
        return logDir.appendingPathComponent(fileName)
    }
    
    func getRecentLogs(lines: Int = 100) -> [String] {
        guard let logFile = getTodayLogFile() else {
            return ["Log file not found"]
        }
        
        do {
            let content = try String(contentsOf: logFile, encoding: .utf8)
            let allLines = content.components(separatedBy: .newlines)
            let startIndex = max(0, allLines.count - lines)
            return Array(allLines[startIndex..<allLines.count].filter { !$0.isEmpty })
        } catch {
            return ["Error reading log file: \(error)"]
        }
    }
    
    func openLogFolder() {
        let logPath = getLogPath()
        let url = URL(fileURLWithPath: logPath)
        NSWorkspace.shared.open(url)
    }
    
    func viewInTerminal() {
        guard let logFile = getTodayLogFile() else { return }
        
        let script =
        """
        tell application "Terminal"
            activate
            do script "tail -f '\(logFile.path)'"
        end tell
        """
        
        if let scriptObject = NSAppleScript(source: script) {
            var error: NSDictionary?
            scriptObject.executeAndReturnError(&error)
            if let error = error {
                print("Error opening Terminal: \(error)")
            }
        }
    }
}
