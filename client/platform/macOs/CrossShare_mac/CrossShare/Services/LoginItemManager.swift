//
//  LoginItemManager.swift
//  CrossShare
//
//  Created by TS on 2025/8/15.
//

import Foundation
import ServiceManagement
import Cocoa

class LoginItemManager {
    static let shared = LoginItemManager()
    
    private let userDefaults = UserDefaults.standard
    private let hasRequestedLoginItemKey = "HasRequestedLoginItem"
    private let isLoginItemEnabledKey = "IsLoginItemEnabled"
    private let helperCommunication = HelperCommunication.shared
    
    private init() {}
    
    func checkAndRequestLoginItemPermissionIfFirstLaunch() {
        // 如果是首次启动，直接静默安装 Helper
        if !hasRequestedLoginItemBefore() {
            #if DEBUG
            print("首次启动，自动请求后台运行权限")
            #endif

            // 直接启用登录项，不显示对话框
            self.enableLoginItem { [weak self] success in
                if success {
                    self?.showSuccessNotification()
                    self?.markLoginItemRequested()
                    #if DEBUG
                    print("后台权限请求成功")
                    #endif
                } else {
                    #if DEBUG
                    print("后台权限请求失败")
                    #endif
                }
            }
            return
        }

        // 非首次启动，检查是否被用户手动禁用
        if !isLoginItemEnabled {
            DispatchQueue.main.async { [weak self] in
                if #available(macOS 13.0, *) {
                    let (status, isUserDisabled) = self?.helperCommunication.getHelperDetailedStatus() ?? (.notFound, false)

                    #if DEBUG
                    print("Helper status: \(status), isUserDisabled: \(isUserDisabled)")
                    #endif

                    // 只有在用户手动禁用的情况下才显示对话框
                    if isUserDisabled {
                        self?.showManualEnableDialog()
                    }
                }
            }
        }
        
        #if DEBUG
        print("=== LoginItemManager Debug Info ===")
        print("Has requested before: \(hasRequestedLoginItemBefore())")
        print("Is login item enabled: \(isLoginItemEnabled)")
        print("Helper installed: \(helperCommunication.isHelperInstalled)")
        print("Helper running: \(helperCommunication.isHelperRunning)")
        print("Helper path: \(helperCommunication.helperPath ?? "nil")")
        if #available(macOS 13.0, *) {
            let (status, _) = helperCommunication.getHelperDetailedStatus()
            print("Helper detailed status: \(status)")
        }
        print("==================================")
        #endif
    }
    
    private func showManualEnableDialog() {
        let alert = NSAlert()
        alert.messageText = "需要在系统设置中启用后台权限"
        alert.informativeText = "CrossShare 的后台权限已被系统禁用，需要您手动在系统设置中重新启用：\n\n1. 打开「系统设置」→「通用」→「登录项与扩展」\n2. 找到「CrossShare Helper」\n3. 开启右侧的开关\n4. 返回应用重新启动\n\n您也可以选择退出应用。"
        alert.addButton(withTitle: "打开系统设置")
        alert.addButton(withTitle: "稍后设置")
        alert.addButton(withTitle: "退出应用")
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
        alert.messageText = "CrossShare 需要后台运行权限"
        alert.informativeText = "为了提供最佳的文件分享体验，CrossShare 需要在后台运行。这样可以确保您随时可以快速分享文件，即使主窗口关闭也能正常工作。\n\n请允许 CrossShare 在后台运行，或选择退出应用。"
        alert.addButton(withTitle: "允许并运行")
        alert.addButton(withTitle: "退出应用")
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
        alert.messageText = "后台权限设置失败"
        alert.informativeText = "无法设置 CrossShare 的后台运行权限。请检查系统设置或重新尝试。\n\n您可以重新尝试或退出应用。"
        alert.addButton(withTitle: "重新尝试")
        alert.addButton(withTitle: "退出应用")
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
    
    /// 显示成功启用后台运行的通知
    private func showSuccessNotification() {
        let notification = NSUserNotification()
        notification.title = "CrossShare 后台运行已启用"
        notification.informativeText = "应用现在可以在后台持续运行，即使主窗口关闭也能快速分享文件。"
        notification.soundName = NSUserNotificationDefaultSoundName
        NSUserNotificationCenter.default.deliver(notification)
        print("后台运行权限已成功启用")
    }
    
    /// Check if login item is currently enabled
    var isLoginItemEnabled: Bool {
        get {
            return helperCommunication.isHelperRunning
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
        alert.messageText = "允许 CrossShare 在后台运行"
        alert.informativeText = "为了提供最佳的文件分享体验，CrossShare 需要在系统启动时自动运行，并在后台保持活跃状态。这样可以确保您随时可以快速分享文件。\n\n系统将安装一个后台帮助程序，您可以随时在系统设置中更改此选项。"
        alert.addButton(withTitle: "允许")
        alert.addButton(withTitle: "暂不")
        alert.alertStyle = .informational
        
        // Set app icon if available
        if let appIcon = NSApp.applicationIconImage {
            alert.icon = appIcon
        }
        
        let response = alert.runModal()
        completion(response == .alertFirstButtonReturn)
    }
    
    private func enableLoginItem(completion: @escaping (Bool) -> Void = { _ in }) {
        helperCommunication.installHelper { [weak self] success in
            DispatchQueue.main.async {
                if success {
                    self?.userDefaults.set(true, forKey: self?.isLoginItemEnabledKey ?? "")
                    print("Helper installed and login item enabled successfully")
                    
                    self?.helperCommunication.notifyHelperAppStatus(isActive: true)
                    completion(true)
                } else {
                    print("Failed to install helper and enable login item")
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
                    print("Helper uninstalled and login item disabled successfully")
                } else {
                    print("Failed to uninstall helper and disable login item")
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
        print("已重置权限请求状态，下次启动将重新请求")
    }
    

    func forceRequestPermission() {
        print("强制显示权限请求对话框")
        requestLoginItemPermission()
    }
    
    /// Get a user-friendly status description
    var statusDescription: String {
        if isLoginItemEnabled {
            return "CrossShare 后台助手已启用"
        } else {
            return "CrossShare 后台助手未启用"
        }
    }
    
    /// 打开系统设置的登录项页面
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
    
    func checkHelperStatus() {
        if !helperCommunication.isHelperInstalled && userDefaults.bool(forKey: isLoginItemEnabledKey) {
            print("Helper was enabled but not found, reinstalling...")
            enableLoginItem()
        }
    }
}
