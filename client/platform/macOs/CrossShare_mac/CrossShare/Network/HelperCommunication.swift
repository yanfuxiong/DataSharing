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
            print("Helper not found in app bundle at: \(helperPath ?? "unknown")")
            completion(false)
            return
        }
        
        do {
            let helperApp = SMAppService.loginItem(identifier: helperBundleIdentifier)
            let currentStatus = helperApp.status
            print("Current helper status: \(currentStatus)")
            
            if currentStatus == .notFound {
                print("Helper service not found, attempting to register...")
            }
            
            try helperApp.register()
            
            DispatchQueue.main.asyncAfter(deadline: .now() + 1.0) {
                let newStatus = helperApp.status
                print("Helper status after registration: \(newStatus)")
                
                let success = (newStatus == .enabled || newStatus == .requiresApproval)
                if success {
                    self.userDefaults.set(true, forKey: self.helperInstalledKey)
                    print("Helper installed successfully (modern)")
                } else {
                    print("Helper registration may have failed, status: \(newStatus)")
                }
                completion(success)
            }
        } catch {
            print("Failed to install helper (modern): \(error)")
            completion(false)
        }
    }
    
    @available(macOS 13.0, *)
    private func uninstallHelperModern(completion: @escaping (Bool) -> Void) {
        do {
            let helperApp = SMAppService.loginItem(identifier: helperBundleIdentifier)
            try helperApp.unregister()
            
            userDefaults.set(false, forKey: helperInstalledKey)
            print("Helper uninstalled successfully (modern)")
            completion(true)
        } catch {
            print("Failed to uninstall helper (modern): \(error)")
            completion(false)
        }
    }
    
    @available(macOS 13.0, *)
    private func checkHelperStatusModern() -> Bool {
        let helperApp = SMAppService.loginItem(identifier: helperBundleIdentifier)
        let status = helperApp.status
        // SMAppService.Status的rawValue含义：
        // 0 = notRegistered  - 未注册
        // 1 = enabled        - 已启用
        // 2 = requiresApproval - 需要用户批准（已注册但被用户禁用）
        // 3 = notFound       - 未找到
        print("Helper status check: \(status)")
        return status == .enabled || status == .requiresApproval
    }
    
    @available(macOS 13.0, *)
    func getHelperDetailedStatus() -> (status: SMAppService.Status, isUserDisabled: Bool) {
        let helperApp = SMAppService.loginItem(identifier: helperBundleIdentifier)
        let status = helperApp.status
        print("Detailed helper status: \(status) (rawValue: \(status.rawValue))")
        let wasInstalled = userDefaults.bool(forKey: helperInstalledKey)
        let isUserDisabled = ((status == .requiresApproval) || (status == .notFound && wasInstalled))
        return (status: status, isUserDisabled: isUserDisabled)
    }
    
    private func installHelperLegacy(completion: @escaping (Bool) -> Void) {
        let success = SMLoginItemSetEnabled(helperBundleIdentifier as CFString, true)
        if success {
            userDefaults.set(true, forKey: helperInstalledKey)
            print("Helper installed successfully (legacy)")
        } else {
            print("Failed to install helper (legacy)")
        }
        completion(success)
    }
    
    private func uninstallHelperLegacy(completion: @escaping (Bool) -> Void) {
        let success = SMLoginItemSetEnabled(helperBundleIdentifier as CFString, false)
        if success {
            userDefaults.set(false, forKey: helperInstalledKey)
            print("Helper uninstalled successfully (legacy)")
        } else {
            print("Failed to uninstall helper (legacy)")
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
                print("Notified helper of app status: \(isActive)")
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
                print("DEBUG: Using Helper from build directory: \(debugHelperPath)")
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
            print("launchHelper: Helper not installed")
            completion(false)
            return
        }
        
        guard let path = helperPath else {
            print("launchHelper: Helper path is nil")
            completion(false)
            return
        }
        
        let helperURL = URL(fileURLWithPath: path)
        print("launchHelper: Attempting to launch Helper at: \(helperURL.path)")
        
        let workspace = NSWorkspace.shared
        let configuration = NSWorkspace.OpenConfiguration()
        configuration.activates = false
        configuration.hides = true
        
        workspace.openApplication(at: helperURL, configuration: configuration) { app, error in
            DispatchQueue.main.async {
                if let error = error {
                    print("Failed to launch helper: \(error)")
                    completion(false)
                } else {
                    print("Helper launched successfully: \(app?.bundleIdentifier ?? "unknown")")
                    completion(true)
                }
            }
        }
    }
}
