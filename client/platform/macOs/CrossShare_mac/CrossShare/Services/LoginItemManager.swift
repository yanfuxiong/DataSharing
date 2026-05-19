//
//  LoginItemManager.swift
//  CrossShare
//
//  Created by TS on 2025/8/15.
//

import Foundation
import ServiceManagement
import Cocoa
import UserNotifications

class LoginItemManager {
    static let shared = LoginItemManager()
    
    private let userDefaults = UserDefaults.standard
    private let hasRequestedLoginItemKey = "HasRequestedLoginItem"
    private let isLoginItemEnabledKey = "IsLoginItemEnabled"
    private let helperCommunication = HelperCommunication.shared
    
    private init() {}
    
    func checkAndRequestLoginItemPermissionIfFirstLaunch() {
        let runInBackgroundEnabled = helperCommunication.isRunInBackgroundEnabled
        
        // Default behavior: Run in Background is OFF, so do not auto-enable helper.
        if !hasRequestedLoginItemBefore() {
            logger.info("First launch: keep RunInBackground default OFF, skip auto enabling helper")
            markLoginItemRequested()
            return
        }
        
        // Do not show background permission dialogs when feature is currently OFF.
        guard runInBackgroundEnabled else {
            logger.info("RunInBackground is OFF, skipping login item permission checks")
            return
        }

        // Not first launch, check if user manually disabled
        if !isLoginItemEnabled {
            DispatchQueue.main.async { [weak self] in
                if #available(macOS 13.0, *) {
                    let (status, isUserDisabled) = self?.helperCommunication.getHelperDetailedStatus() ?? (.notFound, false)

                    #if DEBUG
                    logger.info("Helper status: \(status), isUserDisabled: \(isUserDisabled)")
                    #endif

                    // Only show dialog if user manually disabled it
                    if isUserDisabled {
                        self?.showManualEnableDialog()
                    }
                }
            }
        }
        
        #if DEBUG
        logger.info("=== LoginItemManager Debug Info ===")
        logger.info("Has requested before: \(hasRequestedLoginItemBefore())")
        logger.info("Login item enabled (system): \(isLoginItemEnabled)")
        logger.info("Helper installed: \(helperCommunication.isHelperInstalled)")
        logger.info("Helper registered: \(helperCommunication.isHelperRegistered)")
        logger.info("Helper process running: \(helperCommunication.isHelperRunning)")
        logger.info("Helper path: \(helperCommunication.helperPath ?? "nil")")
        if #available(macOS 13.0, *) {
            let (status, _) = helperCommunication.getHelperDetailedStatus()
            logger.info("Helper SMAppService status: \(status)")
        }
        logger.info("==================================")
        #endif
    }
    
    private func showManualEnableDialog() {
        let alert = NSAlert()
        alert.messageText = "Enable Background Permission in System Settings"
        alert.informativeText = "CrossShare's background permission has been disabled by the system. Please manually enable it in System Settings:\n\n1. Open System Settings → General → Login Items & Extensions\n2. Find \"CrossShare\"\n3. Toggle the switch on the right\n4. Return to the app and restart\n\nYou can also choose to quit the app."
        alert.addButton(withTitle: "Open System Settings")
        alert.addButton(withTitle: "Later")
        alert.addButton(withTitle: "Quit App")
        alert.alertStyle = .informational
        
        if let appIcon = NSApp.applicationIconImage {
            alert.icon = appIcon
        }
        
        let response = alert.runModal()
        
        switch response {
        case .alertFirstButtonReturn:
            openLoginItemsSettings()
            DispatchQueue.main.asyncAfter(deadline: .now() + 2.0) { [weak self] in
                self?.recheckPermissionStatus()
            }
        case .alertSecondButtonReturn:
            DispatchQueue.main.asyncAfter(deadline: .now() + 3.0) { [weak self] in
                self?.recheckPermissionStatus()
            }
        default:
            NSApp.terminate(nil)
        }
    }
    
    private func showMandatoryPermissionDialog() {
        let alert = NSAlert()
        alert.messageText = "CrossShare Requires Background Running Permission"
        alert.informativeText = "To provide the best file sharing experience, CrossShare needs to run in the background. This ensures you can quickly share files at any time, even when the main window is closed.\n\nPlease allow CrossShare to run in the background, or quit the app."
        alert.addButton(withTitle: "Allow and Run")
        alert.addButton(withTitle: "Quit App")
        alert.alertStyle = .critical
        
        if let appIcon = NSApp.applicationIconImage {
            alert.icon = appIcon
        }
        
        let response = alert.runModal()
        
        if response == .alertFirstButtonReturn {
            enableLoginItem { [weak self] success in
                DispatchQueue.main.async {
                    if success {
                        self?.showSuccessNotification()
                        self?.markLoginItemRequested()
                    } else {
                        self?.showPermissionFailedDialog()
                    }
                }
            }
        } else {
            NSApp.terminate(nil)
        }
    }
    
    private func showPermissionFailedDialog() {
        let alert = NSAlert()
        alert.messageText = "Background Permission Setup Failed"
        alert.informativeText = "Failed to set CrossShare's background running permission. Please check System Settings or try again.\n\nYou can retry or quit the app."
        alert.addButton(withTitle: "Retry")
        alert.addButton(withTitle: "Quit App")
        alert.alertStyle = .warning
        
        if let appIcon = NSApp.applicationIconImage {
            alert.icon = appIcon
        }
        
        let response = alert.runModal()
        
        if response == .alertFirstButtonReturn {
            showMandatoryPermissionDialog()
        } else {
            NSApp.terminate(nil)
        }
    }
    
    func requestLoginItemPermission() {
        showLoginItemPermissionDialog { [weak self] granted in
            if granted {
                self?.enableLoginItem { success in
                    DispatchQueue.main.async {
                        if success {
                            self?.showSuccessNotification()
                        }
                    }
                }
            }
            self?.markLoginItemRequested()
        }
    }
    
    /// Show notification for successfully enabling background running
    private func showSuccessNotification() {
        let center = UNUserNotificationCenter.current()
        
        // Request notification permission
        center.requestAuthorization(options: [.alert, .sound]) { granted, error in
            if let error = error {
                logger.error("Notification permission request failed: \(error.localizedDescription)")
                return
            }
            
            guard granted else {
                logger.info("User did not grant notification permission")
                return
            }
            
            // Create notification content
            let content = UNMutableNotificationContent()
            content.title = "CrossShare Background Running Enabled"
            content.body = "The app can now run continuously in the background, allowing quick file sharing even when the main window is closed."
            content.sound = .default
            
            // Create immediate notification request
            let request = UNNotificationRequest(
                identifier: UUID().uuidString,
                content: content,
                trigger: nil // nil means trigger immediately
            )
            
            // Send notification
            center.add(request) { error in
                if let error = error {
                    logger.error("Notification delivery failed: \(error.localizedDescription)")
                } else {
                    logger.info("Background running permission successfully enabled")
                }
            }
        }
    }
    
    /// manual action dialog
    func showHelperNotStartedDialog() {
        DispatchQueue.main.async {
            let alert = NSAlert()
            alert.messageText = "CrossShare service needs to be started manually"
            alert.informativeText = "CrossShare has been registered, but manual action is required to ensure it starts properly:\n\n1. Open 「System Settings」 → 「General」 → 「Login Items」\n2. Find 「CrossShare」\n3. Turn the switch on, or toggle it off and back on\n\nThis ensures CrossShare service functions correctly."
            alert.addButton(withTitle: "Open 「System Settings」 now")
            alert.addButton(withTitle: "Set it up later")
            alert.alertStyle = .informational
            
            if let appIcon = NSApp.applicationIconImage {
                alert.icon = appIcon
            }
            
            let response = alert.runModal()
            
            if response == .alertFirstButtonReturn {
                self.openLoginItemsSettings()
            }
        }
    }
    
    /// SMAppService 登录项是否已注册（macOS 13+ 含 `.enabled` 与 `.requiresApproval`）。
    /// 与 Helper 进程是否在跑不同；进程状态用 `HelperCommunication.isHelperRunning` / `isHelperProcessRunning`。
    var isLoginItemEnabled: Bool {
        get {
            return helperCommunication.isHelperRegistered
        }
        set {
            if newValue {
                enableLoginItem()
            } else {
                disableLoginItem()
            }
        }
    }
    
    // MARK: - Private Methods
    
    func hasRequestedLoginItemBefore() -> Bool {
        return userDefaults.bool(forKey: hasRequestedLoginItemKey)
    }
    
    private func markLoginItemRequested() {
        userDefaults.set(true, forKey: hasRequestedLoginItemKey)
    }
    
    private func showLoginItemPermissionDialog(completion: @escaping (Bool) -> Void) {
        let alert = NSAlert()
        alert.messageText = "Allow CrossShare to Run in the Background"
        alert.informativeText = "To provide the best file sharing experience, CrossShare needs to launch automatically at system startup and remain active in the background. This ensures you can quickly share files at any time.\n\nThe system will install a background helper program. You can change this option in System Settings at any time."
        alert.addButton(withTitle: "Allow")
        alert.addButton(withTitle: "Not Now")
        alert.alertStyle = .informational
        
        // Set app icon if available
        if let appIcon = NSApp.applicationIconImage {
            alert.icon = appIcon
        }
        
        let response = alert.runModal()
        completion(response == .alertFirstButtonReturn)
    }
    
    private func enableLoginItem(completion: @escaping (Bool) -> Void = { _ in }) {
        helperCommunication.installHelper { [weak self] outcome in
            DispatchQueue.main.async {
                guard let self = self else { return }
                switch outcome {
                case .readyForXPC:
                    self.userDefaults.set(true, forKey: self.isLoginItemEnabledKey)
                    logger.info("Helper installed and process ready for XPC")
                    self.helperCommunication.notifyHelperAppStatus(isActive: true)
                    completion(true)
                case .needsUserApproval:
                    self.userDefaults.set(true, forKey: self.isLoginItemEnabledKey)
                    logger.info("Helper registered; user approval required in System Settings")
                    completion(true)
                case .registrationEnabledButProcessMissing:
                    self.userDefaults.set(true, forKey: self.isLoginItemEnabledKey)
                    logger.warn("Helper registered but process did not start in time")
                    completion(true)
                case .failed:
                    logger.info("Failed to install helper and enable login item")
                    completion(false)
                }
            }
        }
    }
    
    private func disableLoginItem() {
        helperCommunication.uninstallHelper { [weak self] success in
            DispatchQueue.main.async {
                if success {
                    self?.userDefaults.set(false, forKey: self?.isLoginItemEnabledKey ?? "")
                    logger.info("Helper uninstalled and login item disabled successfully")
                } else {
                    logger.info("Failed to uninstall helper and disable login item")
                }
            }
        }
    }
}

extension LoginItemManager {
    /// Reset the "has requested" flag (useful for testing)
    func resetRequestedFlag() {
        userDefaults.removeObject(forKey: hasRequestedLoginItemKey)
        userDefaults.removeObject(forKey: isLoginItemEnabledKey)
        logger.info("Permission request status has been reset, will request again on next launch")
    }
    

    func forceRequestPermission() {
        logger.info("Forcing permission request dialog to show")
        requestLoginItemPermission()
    }
    
    /// Get a user-friendly status description
    var statusDescription: String {
        if isLoginItemEnabled {
            return "CrossShare background helper is enabled"
        } else {
            return "CrossShare background helper is disabled"
        }
    }
    
    /// Open the Login Items page in System Settings
    private func openLoginItemsSettings() {
        if #available(macOS 13.0, *) {
            let url = URL(string: "x-apple.systempreferences:com.apple.LoginItems-Settings.extension")!
            NSWorkspace.shared.open(url)
        } else {
            let url = URL(string: "x-apple.systempreferences:com.apple.preference.users")!
            NSWorkspace.shared.open(url)
        }
    }
    
    private func recheckPermissionStatus() {
        if isLoginItemEnabled {
            showSuccessNotification()
        } else {
            showManualEnableDialog()
        }
    }
}
