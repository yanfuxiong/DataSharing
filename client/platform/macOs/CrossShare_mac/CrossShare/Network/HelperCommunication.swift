//
//  HelperCommunication.swift
//  CrossShare
//
//  Created by TS on 2025/8/15.
//

import Foundation
import ServiceManagement
import Cocoa

class HelperCommunication {
    static let shared = HelperCommunication()
    
    private let helperBundleIdentifier = "com.realtek.crossshare.macos.helper"
    private let userDefaults = UserDefaults.standard
    private let helperInstalledKey = "HelperInstalled"
    
    private init() {}
    
    func installHelper(completion: @escaping (Bool) -> Void) {
        if #available(macOS 13.0, *) {
            installHelperModern(completion: completion)
        } else {
            installHelperLegacy(completion: completion)
        }
    }
    
    func uninstallHelper(completion: @escaping (Bool) -> Void) {
        if #available(macOS 13.0, *) {
            uninstallHelperModern(completion: completion)
        } else {
            uninstallHelperLegacy(completion: completion)
        }
    }
    
    var isHelperRunning: Bool {
        if #available(macOS 13.0, *) {
            return checkHelperStatusModern()
        } else {
            return checkHelperStatusLegacy()
        }
    }
    
    @available(macOS 13.0, *)
    private func installHelperModern(completion: @escaping (Bool) -> Void) {
        guard isHelperInstalled else {
            logger.info("Helper not found in app bundle at: \(helperPath ?? "unknown")")
            completion(false)
            return
        }
        
        do {
            let helperApp = SMAppService.loginItem(identifier: helperBundleIdentifier)
            let currentStatus = helperApp.status
            logger.info("Current helper status: \(currentStatus)")
            
            if currentStatus == .notFound {
                logger.info("Helper service not found, attempting to register...")
            }
            
            try helperApp.register()
            
            DispatchQueue.main.asyncAfter(deadline: .now() + 1.0) {
                let newStatus = helperApp.status
                logger.info("Helper status after registration: \(newStatus)")
                
                let success = (newStatus == .enabled || newStatus == .requiresApproval)
                if success {
                    self.userDefaults.set(true, forKey: self.helperInstalledKey)
                    logger.info("Helper registered successfully (modern)")
                    logger.info("Helper status: \(newStatus)")
                    
                    if newStatus == .enabled {
                        // IMPORTANT: Do NOT manually launch Helper!
                        // Only system-launched Helper has proper XPC Mach Service registration
                        logger.info("Status is .enabled - Waiting for system to auto-start Helper...")
                        
                        // Use polling instead of fixed wait time for more reliability
                        self.waitForHelperToStart(timeout: 15.0, pollInterval: 1.0) { helperStarted in
                            if helperStarted {
                                logger.info("Helper process detected and ready for XPC connections")
                            } else {
                                logger.warn("Helper process not detected after 15 seconds")
                                logger.warn("Showing manual activation dialog to user...")
                                
                                // Show dialog to guide user
                                LoginItemManager.shared.showHelperNotStartedDialog()
                            }
                            completion(success)
                        }
                    } else if newStatus == .requiresApproval {
                        // User needs to approve
                        logger.warn("Status is .requiresApproval - User needs to approve in System Settings")
                        logger.info("Please go to System Settings → General → Login Items to enable CrossShare")
                        
                        DispatchQueue.main.asyncAfter(deadline: .now() + 1.0) {
                            completion(success)
                        }
                    }
                } else {
                    logger.info("Helper registration may have failed, status: \(newStatus)")
                    completion(false)
                }
            }
        } catch {
            logger.info("Failed to install helper (modern): \(error)")
            completion(false)
        }
    }
    
    @available(macOS 13.0, *)
    private func uninstallHelperModern(completion: @escaping (Bool) -> Void) {
        do {
            let helperApp = SMAppService.loginItem(identifier: helperBundleIdentifier)
            let currentStatus = helperApp.status
            logger.info("Helper status before unregister: \(currentStatus)")
            
            try helperApp.unregister()
            logger.info("Helper unregistered from SMAppService")
            
            // Terminate the Helper process if it's still running
            logger.info("Terminating Helper process...")
            terminateHelperProcess()
            
            // Wait a bit for process to terminate
            DispatchQueue.main.asyncAfter(deadline: .now() + 1.0) {
                let newStatus = helperApp.status
                logger.info("Helper status after unregister: \(newStatus)")
                
                self.userDefaults.set(false, forKey: self.helperInstalledKey)
                logger.info("Helper uninstalled successfully (modern)")
                completion(true)
            }
        } catch {
            logger.info("Failed to uninstall helper (modern): \(error)")
            // Even if unregister failed, try to terminate the process
            terminateHelperProcess()
            completion(false)
        }
    }
    
    /// Wait for Helper to start with polling (non-blocking)
    private func waitForHelperToStart(timeout: TimeInterval, pollInterval: TimeInterval, completion: @escaping (Bool) -> Void) {
        let startTime = Date()
        var pollCount = 0
        
        logger.info("Starting to poll for Helper process (timeout: \(timeout)s, interval: \(pollInterval)s)")
        
        func pollHelper() {
            pollCount += 1
            let elapsed = Date().timeIntervalSince(startTime)
            
            logger.info("Polling for Helper process (attempt \(pollCount), elapsed: \(String(format: "%.1f", elapsed))s)")
            
            DispatchQueue.global(qos: .utility).async {
                let isRunning = self.isHelperProcessRunning()
                
                DispatchQueue.main.async {
                    if isRunning {
                        logger.info("Helper process detected after \(String(format: "%.1f", elapsed))s (attempt \(pollCount))")
                        // Give XPC service a bit more time to be ready
                        DispatchQueue.main.asyncAfter(deadline: .now() + 1.0) {
                            completion(true)
                        }
                    } else if elapsed >= timeout {
                        logger.warn("Timeout reached (\(timeout)s), Helper not detected")
                        completion(false)
                    } else {
                        // Continue polling
                        DispatchQueue.main.asyncAfter(deadline: .now() + pollInterval) {
                            pollHelper()
                        }
                    }
                }
            }
        }
        
        // Start polling
        pollHelper()
    }
    
    /// Check if Helper process is actually running (async-safe)
    private func isHelperProcessRunning() -> Bool {
        // Method 1: Try NSWorkspace.runningApplications (may not work for LSBackgroundOnly apps)
        let runningApps = NSWorkspace.shared.runningApplications
        let foundInWorkspace = runningApps.contains { $0.bundleIdentifier == helperBundleIdentifier }
        
        if foundInWorkspace {
            logger.info("Helper detected via NSWorkspace.runningApplications")
            return true
        }
        
        // Method 2: Use pgrep with bundle identifier (most reliable)
        let task = Process()
        task.launchPath = "/usr/bin/pgrep"
        // Use -f to match full command line (includes bundle identifier)
        task.arguments = ["-f", helperBundleIdentifier]
        
        let pipe = Pipe()
        task.standardOutput = pipe
        task.standardError = pipe
        
        do {
            try task.run()
            task.waitUntilExit()
            
            // pgrep returns 0 if process found, 1 if not found
            if task.terminationStatus == 0 {
                let data = pipe.fileHandleForReading.readDataToEndOfFile()
                if let output = String(data: data, encoding: .utf8), !output.isEmpty {
                    let pids = output.trimmingCharacters(in: .whitespacesAndNewlines).components(separatedBy: "\n")
                    logger.info("Helper detected via pgrep (bundle ID), PIDs: \(pids.joined(separator: ", "))")
                    return true
                }
            }
        } catch {
            logger.info("Failed to run pgrep: \(error)")
        }
        
        logger.info("Helper process not detected")
        return false
    }
    
    /// Terminate the Helper process if it's running (async, non-blocking)
    private func terminateHelperProcess() {
        // NSWorkspace.runningApplications cannot detect background-only apps (LSBackgroundOnly)
        // Use killall command in background thread to avoid blocking
        
        logger.info("Attempting to terminate Helper process using killall...")
        
        DispatchQueue.global(qos: .utility).async {
            let task = Process()
            task.launchPath = "/usr/bin/killall"
            task.arguments = ["-9", "CrossShareHelper"]  // SIGKILL
            
            let pipe = Pipe()
            task.standardOutput = pipe
            task.standardError = pipe
            
            do {
                try task.run()
                task.waitUntilExit()
                
                DispatchQueue.main.async {
                    if task.terminationStatus == 0 {
                        logger.info("Helper process terminated successfully")
                    } else {
                        logger.info("Helper process may not be running (killall exit code: \(task.terminationStatus))")
                    }
                }
            } catch {
                DispatchQueue.main.async {
                    logger.info("Failed to run killall command: \(error)")
                }
            }
            
            // Double check after termination
            DispatchQueue.main.asyncAfter(deadline: .now() + 0.5) {
                DispatchQueue.global(qos: .utility).async {
                    if self.isHelperProcessRunning() {
                        DispatchQueue.main.async {
                            logger.warn("Helper process still running after killall attempt")
                        }
                    } else {
                        DispatchQueue.main.async {
                            logger.info("Helper process confirmed terminated")
                        }
                    }
                }
            }
        }
    }
    
    @available(macOS 13.0, *)
    private func checkHelperStatusModern() -> Bool {
        let helperApp = SMAppService.loginItem(identifier: helperBundleIdentifier)
        let status = helperApp.status
        // SMAppService.Status的rawValue含义：
        // 0 = notRegistered  - 未注册
        // 1 = enabled        - 已启用（系统管理，应该在运行）
        // 2 = requiresApproval - 已注册但需要批准
        // 3 = notFound       - 未找到
        logger.info("Helper status check: \(status)")
        
        // Note: Helper is a background-only app (LSBackgroundOnly + LSUIElement),
        // so NSWorkspace.shared.runningApplications won't detect it.
        // We trust SMAppService status instead.
        
        // If status is .enabled, the system guarantees the Helper is running
        // If status is .requiresApproval, it means registered but user disabled it
        let isRunning = (status == .enabled || status == .requiresApproval)
        
        logger.info("Helper SMAppService status: \(status) → isRunning: \(isRunning)")
        
        return isRunning
    }
    
    @available(macOS 13.0, *)
    func getHelperDetailedStatus() -> (status: SMAppService.Status, isUserDisabled: Bool) {
        let helperApp = SMAppService.loginItem(identifier: helperBundleIdentifier)
        let status = helperApp.status
        logger.info("Detailed helper status: \(status) (rawValue: \(status.rawValue))")
        let wasInstalled = userDefaults.bool(forKey: helperInstalledKey)
        let isUserDisabled = ((status == .requiresApproval) || (status == .notFound && wasInstalled))
        return (status: status, isUserDisabled: isUserDisabled)
    }
    
    private func installHelperLegacy(completion: @escaping (Bool) -> Void) {
        let success = SMLoginItemSetEnabled(helperBundleIdentifier as CFString, true)
        if success {
            userDefaults.set(true, forKey: helperInstalledKey)
            logger.info("Helper installed successfully (legacy)")
        } else {
            logger.info("Failed to install helper (legacy)")
        }
        completion(success)
    }
    
    private func uninstallHelperLegacy(completion: @escaping (Bool) -> Void) {
        let success = SMLoginItemSetEnabled(helperBundleIdentifier as CFString, false)
        if success {
            userDefaults.set(false, forKey: helperInstalledKey)
            logger.info("Helper uninstalled successfully (legacy)")
        } else {
            logger.info("Failed to uninstall helper (legacy)")
        }
        completion(success)
    }
    
    private func checkHelperStatusLegacy() -> Bool {
        guard let jobDicts = SMCopyAllJobDictionaries(kSMDomainUserLaunchd)?.takeRetainedValue() as? [[String: Any]] else {
            return false
        }
        
        return jobDicts.contains { dict in
            return dict["Label"] as? String == helperBundleIdentifier
        }
    }
    
    func sendMessageToHelper(_ message: [String: Any], completion: @escaping (Bool) -> Void) {
        let sharedDefaults = UserDefaults(suiteName: "group.com.instance.crossshare")
        sharedDefaults?.set(message, forKey: "MainAppMessage")
        sharedDefaults?.synchronize()
        completion(true)
    }
    
    func receiveMessageFromHelper() -> [String: Any]? {
        let sharedDefaults = UserDefaults(suiteName: "group.com.instance.crossshare")
        return sharedDefaults?.object(forKey: "HelperMessage") as? [String: Any]
    }
    
    func notifyHelperAppStatus(isActive: Bool) {
        let message = [
            "type": "app_status",
            "active": isActive,
            "timestamp": Date().timeIntervalSince1970
        ] as [String : Any]
        
        sendMessageToHelper(message) { success in
            if success {
                logger.info("Notified helper of app status: \(isActive)")
            }
        }
    }
    
    var helperPath: String? {
        let appPath = Bundle.main.bundlePath
        let bundledPath = "\(appPath)/Contents/Library/LoginItems/CrossShareHelper.app"
        
#if DEBUG
        if !FileManager.default.fileExists(atPath: bundledPath) {
            let buildDir = URL(fileURLWithPath: appPath).deletingLastPathComponent()
            let debugHelperPath = buildDir.appendingPathComponent("CrossShareHelper.app").path
            if FileManager.default.fileExists(atPath: debugHelperPath) {
                logger.info("DEBUG: Using Helper from build directory: \(debugHelperPath)")
                return debugHelperPath
            }
        }
#endif
        
        return bundledPath
    }
    
    var isHelperInstalled: Bool {
        guard let path = helperPath else { return false }
        return FileManager.default.fileExists(atPath: path)
    }
    
    func launchHelper(completion: @escaping (Bool) -> Void) {
        guard isHelperInstalled else {
            logger.info("launchHelper: Helper not installed")
            completion(false)
            return
        }
        
        guard let path = helperPath else {
            logger.info("launchHelper: Helper path is nil")
            completion(false)
            return
        }
        
        let helperURL = URL(fileURLWithPath: path)
        logger.info("launchHelper: Attempting to launch Helper at: \(helperURL.path)")
        
        let workspace = NSWorkspace.shared
        let configuration = NSWorkspace.OpenConfiguration()
        configuration.activates = false
        configuration.hides = true
        
        workspace.openApplication(at: helperURL, configuration: configuration) { app, error in
            DispatchQueue.main.async {
                if let error = error {
                    logger.info("Failed to launch helper: \(error)")
                    completion(false)
                } else {
                    logger.info("Helper launched successfully: \(app?.bundleIdentifier ?? "unknown")")
                    completion(true)
                }
            }
        }
    }
}
