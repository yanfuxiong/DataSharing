//
//  NotificationPermissionManager.swift
//  CrossShare
//
//  Created by TS on 2025/12/XX.
//  Manager for handling notification permission alerts
//

import Foundation
import Cocoa

class NotificationPermissionManager {
    static let shared = NotificationPermissionManager()
    
    private let userDefaults = UserDefaults.standard
    private let dontRemindKey = "CrossShareNotificationPermissionDontRemind"
    
    private init() {}
    
    /// Check if "Don't remind" is set
    var shouldShowAlert: Bool {
        return !userDefaults.bool(forKey: dontRemindKey)
    }
    
    /// Show notification permission alert dialog
    func showNotificationPermissionAlert() {
        // If user has chosen "Don't remind", don't show the alert
        guard shouldShowAlert else {
            logger.info("User has chosen 'Don't remind', skipping notification permission alert")
            return
        }
        
        DispatchQueue.main.async {
            let alert = NSAlert()
            alert.messageText = "Notification Permission Required"
            alert.informativeText = "To receive file transfer and system notifications in time, please enable notification permission for CrossShareHelper in System Settings.\n\nGo to: System Settings → Notifications → CrossShareHelper → Enable Notifications"
            alert.alertStyle = .informational
            
            // First button: Open notification settings
            alert.addButton(withTitle: "Open System Settings")
            // Second button: Skip for now
            alert.addButton(withTitle: "Skip")
            // Third button: Don't show again
            alert.addButton(withTitle: "Don't Show Again")
            
            if let appIcon = NSApp.applicationIconImage {
                alert.icon = appIcon
            }
            
            let response = alert.runModal()
            
            switch response {
            case .alertFirstButtonReturn:
                // Open notification settings
                self.openNotificationSettings()
            case .alertSecondButtonReturn:
                // Later - do nothing, just close the dialog
                logger.info("User chose 'Later', will show alert again next time")
            case .alertThirdButtonReturn:
                // Don't show again - save user preference
                self.setDontRemind()
            default:
                break
            }
        }
    }
    
    /// Open system notification settings page
    private func openNotificationSettings() {
        logger.info("Opening system notification settings")
        
        // macOS 13.0+ uses new settings URL
        if #available(macOS 13.0, *) {
            // Try to open notification settings page
            if let url = URL(string: "x-apple.systempreferences:com.apple.preference.notifications") {
                NSWorkspace.shared.open(url)
                logger.info("Opened notification settings (macOS 13+)")
            } else {
                // If URL scheme fails, try using command line
                let task = Process()
                task.launchPath = "/usr/bin/open"
                task.arguments = ["x-apple.systempreferences:com.apple.preference.notifications"]
                task.launch()
                logger.info("Opened notification settings via command line")
            }
        } else {
            // macOS 12 and below use old System Preferences
            if let url = URL(string: "x-apple.systempreferences:com.apple.preference.notifications") {
                NSWorkspace.shared.open(url)
                logger.info("Opened notification settings (macOS 12)")
            }
        }
        
        logger.info("Notification settings opened, user needs to manually find CrossShareHelper in the list")
    }
    
    /// Set "Don't remind" flag
    private func setDontRemind() {
        userDefaults.set(true, forKey: dontRemindKey)
        userDefaults.synchronize()
        logger.info("User chose 'Don't remind', saved preference")
    }
    
    /// Clear "Don't remind" flag (for testing or reset)
    func clearDontRemind() {
        userDefaults.removeObject(forKey: dontRemindKey)
        userDefaults.synchronize()
        logger.info("Cleared 'Don't remind' preference")
    }
}

