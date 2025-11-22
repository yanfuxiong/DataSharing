//
//  CSAlertManager.swift
//  CrossShare
//
//  Alert manager for unified alert handling
//

import Cocoa

class CSAlertManager {
    
    static let shared = CSAlertManager()
    
    private init() {}
    
    // MARK: - Basic Alerts
    
    /// Show a basic informational alert
    /// - Parameters:
    ///   - message: The main message text
    ///   - informativeText: Additional informative text
    func showInfo(message: String, informativeText: String) {
        let alert = NSAlert()
        alert.messageText = message
        alert.informativeText = informativeText
        alert.alertStyle = .informational
        alert.addButton(withTitle: "OK")
        alert.runModal()
    }
    
    /// Show a warning alert
    /// - Parameters:
    ///   - message: The main message text
    ///   - informativeText: Additional informative text
    func showWarning(message: String, informativeText: String) {
        let alert = NSAlert()
        alert.messageText = message
        alert.informativeText = informativeText
        alert.alertStyle = .warning
        alert.addButton(withTitle: "OK")
        alert.runModal()
    }
    
    /// Show an error alert
    /// - Parameters:
    ///   - message: The main message text
    ///   - informativeText: Additional informative text
    func showError(message: String, informativeText: String) {
        let alert = NSAlert()
        alert.messageText = message
        alert.informativeText = informativeText
        alert.alertStyle = .critical
        alert.addButton(withTitle: "OK")
        alert.runModal()
    }
    
    /// Show a confirmation alert with Yes/No buttons
    /// - Parameters:
    ///   - message: The main message text
    ///   - informativeText: Additional informative text
    ///   - completion: Completion handler with user's choice (true for Yes, false for No)
    func showConfirmation(message: String, informativeText: String, completion: @escaping (Bool) -> Void) {
        let alert = NSAlert()
        alert.messageText = message
        alert.informativeText = informativeText
        alert.alertStyle = .warning
        alert.addButton(withTitle: "Yes")
        alert.addButton(withTitle: "No")
        
        let response = alert.runModal()
        completion(response == .alertFirstButtonReturn)
    }
    
    // MARK: - App Specific Alerts
    
    /// Show about dialog with app information
    func showAbout() {
        let alert = NSAlert()
        alert.alertStyle = .informational
        
        // Get app info
        let appName = "CrossShare"
        let appVersion = Bundle.main.infoDictionary?["CFBundleShortVersionString"] as? String ?? "1.0.0"
        let buildVersion = Bundle.main.infoDictionary?["CFBundleVersion"] as? String ?? "1"
        
        alert.messageText = appName
        alert.informativeText = "Version: \(appVersion) (\(buildVersion))"
        alert.addButton(withTitle: "OK")
        
        alert.runModal()
    }
    
    /// Show download path updated notification
    /// - Parameter path: The new download path
    func showDownloadPathUpdated(path: String) {
        showInfo(message: "Download path updated", informativeText: "New path: \(path)")
    }
    
    /// Show bug report export success
    /// - Parameter zipFilePath: The path to the exported zip file
    func showBugReportExported(zipFilePath: String) {
        let fileName = URL(fileURLWithPath: zipFilePath).lastPathComponent
        let directory = URL(fileURLWithPath: zipFilePath).deletingLastPathComponent().path
        showInfo(
            message: "Bug report exported successfully",
            informativeText: "File: \(fileName)\nLocation: \(directory)"
        )
    }
    
    /// Show bug report export failure
    /// - Parameter error: The error that occurred
    func showBugReportExportFailed(error: Error) {
        showError(
            message: "Failed to export bug report",
            informativeText: error.localizedDescription
        )
    }
    
    /// Show file not found alert
    /// - Parameter filePath: The path of the missing file
    func showFileNotFound(filePath: String) {
        showWarning(
            message: "File does not exist",
            informativeText: filePath
        )
    }
    
    /// Show invalid file path alert
    func showInvalidFilePath() {
        showWarning(
            message: "Invalid file path",
            informativeText: "The file path is empty or invalid."
        )
    }
}

