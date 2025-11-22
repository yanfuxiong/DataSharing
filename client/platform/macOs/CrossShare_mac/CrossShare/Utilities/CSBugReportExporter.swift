//
//  CSBugReportExporter.swift
//  CrossShare
//
//  Bug report exporter utility
//

import Foundation
import AppKit

class CSBugReportExporter {
    
    static let shared = CSBugReportExporter()
    
    private init() {}
    
    /// Export bug report to user's selected download directory
    /// - Parameter completion: Completion handler with result (success message or error)
    func exportBugReport(completion: @escaping (Result<String, Error>) -> Void) {
        let logPath = getLogPath()
        
        // Get user's selected download path or default
        let downloadPath = CSUserPreferences.shared.getDownloadPathOrDefault()
        
        // Create date formatter for zip filename
        let dateFormatter = DateFormatter()
        dateFormatter.dateFormat = "yyyy-MM-dd"
        let dateString = dateFormatter.string(from: Date())
        let zipFileName = "CrossShareLogs_\(dateString).zip"
        let zipFilePath = downloadPath + "/" + zipFileName
        
        // Create zip file
        do {
            let fileManager = FileManager.default
            
            // Ensure download directory exists
            if !fileManager.fileExists(atPath: downloadPath) {
                try fileManager.createDirectory(atPath: downloadPath, withIntermediateDirectories: true, attributes: nil)
                logger.info("CSBugReportExporter: Created download directory: \(downloadPath)")
            }
            
            // Create temporary directory for collecting logs
            let tempDir = URL(fileURLWithPath:(NSTemporaryDirectory() + "CrossShareLogs_\(UUID().uuidString)"))
            try fileManager.createDirectory(atPath: tempDir.path, withIntermediateDirectories: true, attributes: nil)
            
            var hasLogs = false
            // Copy logs from log directory
            if fileManager.fileExists(atPath: logPath.path) {
                let logs = try fileManager.contentsOfDirectory(atPath: logPath.path)
                for logFile in logs where logFile.hasSuffix(".log") {
                    let sourcePath = logPath.appendingPathComponent(logFile).path
                    let destPath = tempDir.appendingPathComponent(logFile).path
                    try fileManager.copyItem(atPath: sourcePath, toPath: destPath)
                    hasLogs = true
                    logger.info("CSBugReportExporter: Copied log: \(logFile)")
                }
            } else {
                logger.info("CSBugReportExporter: Log directory not found: \(logPath.path)")
            }
            
            // Check if we have any logs to export
            if !hasLogs {
                // Clean up temp directory
                try? fileManager.removeItem(atPath: tempDir.path)
                
                completion(.failure(NSError(domain: "CSBugReportExporter", code: 1002, userInfo: [
                    NSLocalizedDescriptionKey: "No log files found to export"
                ])))
                return
            }
            
            // Create zip file using command line
            let process = Process()
            process.executableURL = URL(fileURLWithPath: "/usr/bin/zip")
            process.arguments = ["-r", "-j", zipFilePath, tempDir.path]
            
            try process.run()
            process.waitUntilExit()
            
            // Clean up temp directory
            try fileManager.removeItem(atPath: tempDir.path)
            
            if process.terminationStatus == 0 {
                logger.info("CSBugReportExporter: Bug report exported to: \(zipFilePath)")
                completion(.success(zipFilePath))
            } else {
                completion(.failure(NSError(domain: "CSBugReportExporter", code: 1003, userInfo: [
                    NSLocalizedDescriptionKey: "Failed to create zip file"
                ])))
            }
            
        } catch {
            logger.info("CSBugReportExporter: Error exporting bug report: \(error)")
            completion(.failure(error))
        }
    }
}

