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

    func getTodayLogFile() -> URL? {
        let logDir = getLogPath()
        
        let dateFormatter = DateFormatter()
        dateFormatter.dateFormat = "yyyy-MM-dd"
        dateFormatter.timeZone = TimeZone(secondsFromGMT: 8 * 3600) // GMT+8
        let dateStr = dateFormatter.string(from: Date())
        let fileName = "helper-\(dateStr).log"
        
        return logDir.appendingPathComponent(fileName)
    }
    
    func openLogFolder() {
        NSWorkspace.shared.open(getLogPath())
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
                logger.info("Error opening Terminal: \(error)")
            }
        }
    }
}
