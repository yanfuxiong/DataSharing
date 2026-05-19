//
//  HelperCommunication.swift
//  CrossShare
//
//  Created by TS on 2025/8/15.
//

import Foundation
import ServiceManagement
import Cocoa

/// `installHelper` 的结果类型（面向 Release 语义）。
/// 只有 `readyForXPC` 才表示“可以立刻发起 NSXPCConnection”。
enum HelperInstallOutcome {
    /// 登录项注册成功，且 Helper 进程已确认在运行。
    case readyForXPC
    /// `SMAppService` 为 `.requiresApproval`，需要用户去系统设置手动批准。
    case needsUserApproval
    /// 系统状态显示已启用（或旧系统开启成功），但超时前仍未检测到进程。
    case registrationEnabledButProcessMissing
    /// Helper 包缺失、`register()` 失败，或注册后状态异常。
    case failed
}

class HelperCommunication {
    static let shared = HelperCommunication()
    
    private let helperBundleIdentifier = "com.realtek.crossshare.macos.helper"
    private let userDefaults = UserDefaults.standard
    private let helperInstalledKey = "HelperInstalled"
    private let sharedSuiteName = "group.com.instance.crossshare"
    private let runInBackgroundKey = "RunInBackgroundEnabled"
    
    private var installCompletionWaiters: [(HelperInstallOutcome) -> Void] = []
    private var installInProgress = false
    private let installStateLock = NSLock()
    
    /// 兜底拉起限频：避免监控定时器每 2 秒都触发一次 launchctl
    private var lastReviveAttemptAt: Date?
    private var reviveInProgress = false
    private let reviveStateLock = NSLock()
    
    private init() {}
    
    var isRunInBackgroundEnabled: Bool {
        let sharedDefaults = UserDefaults(suiteName: sharedSuiteName)
        if sharedDefaults?.object(forKey: runInBackgroundKey) == nil {
            logger.info("RunInBackground setting not found, use default: false (OFF)")
            return false
        }
        let enabled = sharedDefaults?.bool(forKey: runInBackgroundKey) ?? false
        logger.info("RunInBackground setting loaded: \(enabled)")
        return enabled
    }
    
    func setRunInBackgroundEnabled(_ enabled: Bool) {
        let sharedDefaults = UserDefaults(suiteName: sharedSuiteName)
        sharedDefaults?.set(enabled, forKey: runInBackgroundKey)
        sharedDefaults?.synchronize()
        logger.info("RunInBackground setting persisted to App Group: \(enabled)")
    }
    
    /// Ensure RunInBackground has an explicit persisted default value.
    func ensureRunInBackgroundDefaultOffIfNeeded() {
        let sharedDefaults = UserDefaults(suiteName: sharedSuiteName)
        if sharedDefaults?.object(forKey: runInBackgroundKey) == nil {
            sharedDefaults?.set(false, forKey: runInBackgroundKey)
            sharedDefaults?.synchronize()
            logger.info("RunInBackground default initialized to false (OFF)")
        }
    }
    
    /// Keep login-item registration consistent with "Run in Background" switch.
    /// - Note: Turning OFF only disables auto-launch; it does not kill current helper process.
    func syncLoginItemRegistrationForRunInBackground(_ enabled: Bool, completion: @escaping (Bool) -> Void) {
        logger.info("Sync login-item with RunInBackground target=\(enabled)")
        
        let wrappedCompletion: (Bool) -> Void = { success in
            self.logLoginItemStateSnapshot(context: "post-sync target=\(enabled), success=\(success)")
            completion(success)
        }
        
        if enabled {
            enableLoginItemEnsuringSingleHelper(completion: wrappedCompletion)
        } else {
            disableHelperAutoLaunchOnly(completion: wrappedCompletion)
        }
    }
    
    /// OFF -> ON may have a manually launched helper process still alive.
    /// Registering login-item in this state can create a duplicate helper.
    private func enableLoginItemEnsuringSingleHelper(completion: @escaping (Bool) -> Void) {
        if #available(macOS 13.0, *) {
            let status = SMAppService.loginItem(identifier: helperBundleIdentifier).status
            let isRegistered = (status == .enabled || status == .requiresApproval)
            let helperRunning = queryHelperProcessRunning()
            
            if !isRegistered && helperRunning {
                logger.warn("Detected manual helper while login-item not registered, restarting helper before register to avoid duplicates")
                terminateHelperProcess()
                DispatchQueue.main.asyncAfter(deadline: .now() + 1.0) {
                    self.installHelperWithLegacyBooleanCompletion(completion)
                }
                return
            }
        }
        
        installHelperWithLegacyBooleanCompletion(completion)
    }
    
    /// Disable helper auto-launch registration only (keep current helper process alive).
    func disableHelperAutoLaunchOnly(completion: @escaping (Bool) -> Void) {
        if #available(macOS 13.0, *) {
            disableHelperAutoLaunchOnlyModern(completion: completion)
        } else {
            disableHelperAutoLaunchOnlyLegacy(completion: completion)
        }
    }
    
    /// 注册登录项（或旧系统开启 helper）。
    /// 会合并并发调用（如首启与延迟检查同时触发），只执行一次安装流程并回调同一个结果。
    func installHelper(completion: @escaping (HelperInstallOutcome) -> Void) {
        installStateLock.lock()
        installCompletionWaiters.append(completion)
        if installInProgress {
            let waitersCount = installCompletionWaiters.count
            installStateLock.unlock()
            logger.info("installHelper: request queued (in-flight install, \(waitersCount) waiters)")
            return
        }
        installInProgress = true
        installStateLock.unlock()
        
        let finish: (HelperInstallOutcome) -> Void = { outcome in
            self.installStateLock.lock()
            let waiters = self.installCompletionWaiters
            self.installCompletionWaiters.removeAll()
            self.installInProgress = false
            self.installStateLock.unlock()
            waiters.forEach { $0(outcome) }
        }
        
        if #available(macOS 13.0, *) {
            installHelperModern(completion: finish)
        } else {
            installHelperLegacy(completion: finish)
        }
    }
    
    /// 与旧版 `installHelper(completion: (Bool))` 语义对齐，供 RunInBackground 等路径使用。
    func installHelperWithLegacyBooleanCompletion(_ completion: @escaping (Bool) -> Void) {
        installHelper { outcome in
            switch outcome {
            case .readyForXPC, .needsUserApproval, .registrationEnabledButProcessMissing:
                completion(true)
            case .failed:
                completion(false)
            }
        }
    }
    
    func uninstallHelper(completion: @escaping (Bool) -> Void) {
        if #available(macOS 13.0, *) {
            uninstallHelperModern(completion: completion)
        } else {
            uninstallHelperLegacy(completion: completion)
        }
    }
    
    /// Whether helper login-item is registered/enabled in system (`.enabled` or `.requiresApproval` on macOS 13+).
    var isHelperRegistered: Bool {
        if #available(macOS 13.0, *) {
            return checkHelperStatusModern()
        } else {
            return checkHelperStatusLegacy()
        }
    }
    
    /// 仅表示 Helper 进程是否存在（NSWorkspace / pgrep），不使用 `SMAppService` 状态推断。
    var isHelperRunning: Bool {
        queryHelperProcessRunning()
    }
    
    /// macOS 13+：`SMAppService.Status == .enabled`（用户已允许登录项/后台项）。
    /// 低版本：通过 `SMCopyAllJobDictionaries` 判断（兼容旧 `SMLoginItemSetEnabled` 路径）。
    /// 注意：这不等价于 `isHelperRunning`，开关开启时进程仍可能未起或已崩溃。
    var isLoginItemEnabledBySystem: Bool {
        if #available(macOS 13.0, *) {
            let status = SMAppService.loginItem(identifier: helperBundleIdentifier).status
            return status == .enabled
        }
        return checkHelperStatusLegacy()
    }
    
    /// 在「登录项开关已开启，但进程不存在」时的兜底拉起。
    /// - 参数 reason: 触发原因，用于日志排查。
    /// - 参数 cooldown: 最小重试间隔（秒）。
    /// - 参数 completion: 异步回调；true 表示 kickstart 后已检测到进程。
    func tryReviveHelperByLaunchctlIfNeeded(
        reason: String,
        cooldown: TimeInterval = 4.0,
        completion: ((Bool) -> Void)? = nil
    ) {
        guard isLoginItemEnabledBySystem else {
            logger.info("Skip revive Helper: login item not enabled by system (\(reason))")
            DispatchQueue.main.async { completion?(false) }
            return
        }
        
        let now = Date()
        reviveStateLock.lock()
        if reviveInProgress {
            reviveStateLock.unlock()
            logger.debug("Skip revive Helper: previous revive still in progress (\(reason))")
            DispatchQueue.main.async { completion?(false) }
            return
        }
        if let last = lastReviveAttemptAt, now.timeIntervalSince(last) < cooldown {
            let remain = cooldown - now.timeIntervalSince(last)
            reviveStateLock.unlock()
            logger.debug("Skip revive Helper: cooldown \(String(format: "%.1f", remain))s (\(reason))")
            DispatchQueue.main.async { completion?(false) }
            return
        }
        reviveInProgress = true
        lastReviveAttemptAt = now
        reviveStateLock.unlock()
        
        DispatchQueue.global(qos: .utility).async {
            let uid = getuid()
            let launchdService = "gui/\(uid)/\(self.helperBundleIdentifier)"
            let task = Process()
            task.launchPath = "/bin/launchctl"
            task.arguments = ["kickstart", "-k", launchdService]
            
            let pipe = Pipe()
            task.standardOutput = pipe
            task.standardError = pipe
            
            defer {
                self.reviveStateLock.lock()
                self.reviveInProgress = false
                self.reviveStateLock.unlock()
            }
            
            do {
                logger.info("Trying to revive Helper via launchctl (\(reason)): \(launchdService)")
                try task.run()
                task.waitUntilExit()
                
                let data = pipe.fileHandleForReading.readDataToEndOfFile()
                let output = String(data: data, encoding: .utf8)?
                    .trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
                
                if task.terminationStatus == 0 {
                    logger.info("launchctl revive finished successfully (\(reason))")
                    // kickstart 成功后，短暂等待再检查进程是否已出现，供上层立即触发连接
                    DispatchQueue.global(qos: .utility).asyncAfter(deadline: .now() + 1.0) {
                        let running = self.queryHelperProcessRunning()
                        DispatchQueue.main.async { completion?(running) }
                    }
                } else {
                    logger.warn("launchctl revive failed (\(reason)), exit=\(task.terminationStatus), output=\(output)")
                    DispatchQueue.main.async { completion?(false) }
                }
            } catch {
                logger.warn("launchctl revive execution error (\(reason)): \(error)")
                DispatchQueue.main.async { completion?(false) }
            }
        }
    }
    
    /// Whether helper process currently exists.
    var isHelperProcessRunning: Bool { queryHelperProcessRunning() }
    
    @available(macOS 13.0, *)
    private func installHelperModern(completion: @escaping (HelperInstallOutcome) -> Void) {
        guard isHelperInstalled else {
            logger.info("Helper not found in app bundle at: \(helperPath ?? "unknown")")
            completion(.failed)
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
                
                let registered = (newStatus == .enabled || newStatus == .requiresApproval)
                if registered {
                    self.userDefaults.set(true, forKey: self.helperInstalledKey)
                    logger.info("Helper registered successfully (modern)")
                    logger.info("Helper status: \(newStatus)")
                    
                    if newStatus == .enabled {
                        // 重要：Release 下不要手动 openApplication 拉起 Helper。
                        // 由系统拉起的 Helper 才能稳定完成 Mach Service 注册。
                        logger.info("Status is .enabled - Waiting for system to auto-start Helper...")
                        
                        self.waitForHelperToStart(timeout: 15.0, pollInterval: 1.0) { helperStarted in
                            if helperStarted {
                                logger.info("Helper process detected and ready for XPC connections")
                                completion(.readyForXPC)
                            } else {
                                logger.warn("Helper process not detected after 15 seconds")
                                // 兜底再尝试一次系统拉起，后续由 Monitor 持续观察是否恢复。
                                self.tryReviveHelperByLaunchctlIfNeeded(reason: "install_timeout_modern")
                                logger.warn("Showing manual activation dialog to user...")
                                LoginItemManager.shared.showHelperNotStartedDialog()
                                completion(.registrationEnabledButProcessMissing)
                            }
                        }
                    } else if newStatus == .requiresApproval {
                        logger.warn("Status is .requiresApproval - User needs to approve in System Settings")
                        logger.info("Please go to System Settings → General → Login Items to enable CrossShare")
                        
                        DispatchQueue.main.asyncAfter(deadline: .now() + 1.0) {
                            completion(.needsUserApproval)
                        }
                    }
                } else {
                    logger.info("Helper registration may have failed, status: \(newStatus)")
                    completion(.failed)
                }
            }
        } catch {
            logger.info("Failed to install helper (modern): \(error)")
            completion(.failed)
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
    
    @available(macOS 13.0, *)
    private func disableHelperAutoLaunchOnlyModern(completion: @escaping (Bool) -> Void) {
        do {
            let helperApp = SMAppService.loginItem(identifier: helperBundleIdentifier)
            let currentStatus = helperApp.status
            logger.info("Helper status before disable auto-launch: \(currentStatus)")
            
            // Already disabled or not found, treat as success.
            if currentStatus == .notRegistered || currentStatus == .notFound {
                logger.info("Helper auto-launch already disabled")
                completion(true)
                return
            }
            
            try helperApp.unregister()
            logger.info("Helper auto-launch disabled successfully (modern)")
            completion(true)
        } catch {
            logger.error("Failed to disable helper auto-launch (modern): \(error.localizedDescription)")
            completion(false)
        }
    }
    
    /// 轮询等待 Helper 进程出现（非阻塞）。
    private func waitForHelperToStart(timeout: TimeInterval, pollInterval: TimeInterval, completion: @escaping (Bool) -> Void) {
        let startTime = Date()
        var pollCount = 0
        
        logger.info("Starting to poll for Helper process (timeout: \(timeout)s, interval: \(pollInterval)s)")
        
        func pollHelper() {
            pollCount += 1
            let elapsed = Date().timeIntervalSince(startTime)
            
            logger.info("Polling for Helper process (attempt \(pollCount), elapsed: \(String(format: "%.1f", elapsed))s)")
            
            DispatchQueue.global(qos: .utility).async {
                let isRunning = self.queryHelperProcessRunning()
                
                DispatchQueue.main.async {
                    if isRunning {
                        logger.info("Helper process detected after \(String(format: "%.1f", elapsed))s (attempt \(pollCount))")
                        // 额外等待一点时间，给 XPC 服务初始化预留缓冲
                        DispatchQueue.main.asyncAfter(deadline: .now() + 1.0) {
                            completion(true)
                        }
                    } else if elapsed >= timeout {
                        logger.warn("Timeout reached (\(timeout)s), Helper not detected")
                        completion(false)
                    } else {
                        // 继续轮询
                        DispatchQueue.main.asyncAfter(deadline: .now() + pollInterval) {
                            pollHelper()
                        }
                    }
                }
            }
        }
        
        // 开始轮询
        pollHelper()
    }
    
    /// 检查 Helper 进程是否真实存在（线程安全）。
    /// 检测顺序：可执行名优先，其次命令行中的 bundle ID。
    private func queryHelperProcessRunning() -> Bool {
        let runningApps = NSWorkspace.shared.runningApplications
        if runningApps.contains(where: { $0.bundleIdentifier == helperBundleIdentifier }) {
            logger.info("Helper detected via NSWorkspace.runningApplications")
            return true
        }
        
        if pgrep(arguments: ["-x", "CrossShareHelper"], label: "executable name") {
            return true
        }
        
        if pgrep(arguments: ["-f", helperBundleIdentifier], label: "bundle ID in cmdline") {
            return true
        }
        
        logger.info("Helper process not detected")
        return false
    }
    
    private func pgrep(arguments: [String], label: String) -> Bool {
        let task = Process()
        task.launchPath = "/usr/bin/pgrep"
        task.arguments = arguments
        let pipe = Pipe()
        task.standardOutput = pipe
        task.standardError = pipe
        do {
            try task.run()
            task.waitUntilExit()
            if task.terminationStatus == 0 {
                let data = pipe.fileHandleForReading.readDataToEndOfFile()
                if let output = String(data: data, encoding: .utf8), !output.isEmpty {
                    let pids = output.trimmingCharacters(in: .whitespacesAndNewlines).components(separatedBy: "\n")
                    logger.info("Helper detected via pgrep (\(label)), PIDs: \(pids.joined(separator: ", "))")
                    return true
                }
            }
        } catch {
            logger.info("Failed to run pgrep (\(label)): \(error)")
        }
        return false
    }
    
    /// 若 Helper 仍在运行则尝试终止（异步、非阻塞）
    private func terminateHelperProcess() {
        // NSWorkspace.runningApplications 对 LSBackgroundOnly 的识别不稳定
        // 这里改用后台线程执行 killall，避免阻塞主线程
        
        logger.info("Attempting to terminate Helper process using killall...")
        
        DispatchQueue.global(qos: .utility).async {
            let task = Process()
            task.launchPath = "/usr/bin/killall"
            task.arguments = ["-9", "CrossShareHelper"]  // 强制终止
            
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
            
            // 终止后做一次二次确认
            DispatchQueue.main.asyncAfter(deadline: .now() + 0.5) {
                DispatchQueue.global(qos: .utility).async {
                    if self.queryHelperProcessRunning() {
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
    func getHelperDetailedStatus() -> (status: SMAppService.Status, isUserDisabled: Bool) {
        let helperApp = SMAppService.loginItem(identifier: helperBundleIdentifier)
        let status = helperApp.status
        logger.info("Detailed helper status: \(status) (rawValue: \(status.rawValue))")
        let wasInstalled = userDefaults.bool(forKey: helperInstalledKey)
        let isUserDisabled = ((status == .requiresApproval) || (status == .notFound && wasInstalled))
        return (status: status, isUserDisabled: isUserDisabled)
    }
    
    /// Unified debug log for login-item registration state.
    func logLoginItemStateSnapshot(context: String) {
        if #available(macOS 13.0, *) {
            let detail = getHelperDetailedStatus()
            logger.info(
                "Login-item snapshot [\(context)] - status=\(detail.status), isUserDisabled=\(detail.isUserDisabled)"
            )
        } else {
            logger.info(
                "Login-item snapshot [\(context)] - legacyStatus(isHelperRegistered)=\(isHelperRegistered)"
            )
        }
    }
    
    private func installHelperLegacy(completion: @escaping (HelperInstallOutcome) -> Void) {
        let success = SMLoginItemSetEnabled(helperBundleIdentifier as CFString, true)
        guard success else {
            logger.info("Failed to install helper (legacy)")
            completion(.failed)
            return
        }
        userDefaults.set(true, forKey: helperInstalledKey)
        logger.info("Helper installed successfully (legacy), waiting for process...")
        waitForHelperToStart(timeout: 15.0, pollInterval: 1.0) { helperStarted in
            if helperStarted {
                logger.info("Helper process detected (legacy)")
                completion(.readyForXPC)
            } else {
                logger.warn("Helper process not detected after 15 seconds (legacy)")
                self.tryReviveHelperByLaunchctlIfNeeded(reason: "install_timeout_legacy")
                LoginItemManager.shared.showHelperNotStartedDialog()
                completion(.registrationEnabledButProcessMissing)
            }
        }
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
    
    private func disableHelperAutoLaunchOnlyLegacy(completion: @escaping (Bool) -> Void) {
        let success = SMLoginItemSetEnabled(helperBundleIdentifier as CFString, false)
        if success {
            logger.info("Helper auto-launch disabled successfully (legacy)")
        } else {
            logger.error("Failed to disable helper auto-launch (legacy)")
        }
        completion(success)
    }
    
    @available(macOS 13.0, *)
    private func checkHelperStatusModern() -> Bool {
        let helperApp = SMAppService.loginItem(identifier: helperBundleIdentifier)
        let status = helperApp.status
        logger.info("Helper status check: \(status)")
        let isRegistered = (status == .enabled || status == .requiresApproval)
        logger.info("Helper SMAppService status: \(status) → isRegistered: \(isRegistered)")
        return isRegistered
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
        let sharedDefaults = UserDefaults(suiteName: sharedSuiteName)
        sharedDefaults?.set(message, forKey: "MainAppMessage")
        sharedDefaults?.synchronize()
        completion(true)
    }
    
    func receiveMessageFromHelper() -> [String: Any]? {
        let sharedDefaults = UserDefaults(suiteName: sharedSuiteName)
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
            } else {
                logger.warn("Failed to notify helper of app status (App Group may not be configured)")
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
