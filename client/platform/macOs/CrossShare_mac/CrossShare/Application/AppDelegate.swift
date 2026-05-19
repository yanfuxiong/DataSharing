//
//  AppDelegate.swift
//  CrossShare
//
//  Created by user00 on 2025/3/5.
//

import Cocoa
import AVFoundation
import Foundation
import os.log
import ServiceManagement
import CoreGraphics

func displayReconfigurationCallback(display: CGDirectDisplayID, flags: CGDisplayChangeSummaryFlags, userInfo: UnsafeMutableRawPointer?) {
    // 全局函数中也可以直接使用 logger
    logger.debug("Display reconfiguration detected (display: \(display), flags: \(flags.rawValue))")
}

class AppDelegate: NSObject, NSApplicationDelegate {
    
    var window: NSWindow!
    var mainViewController: MainHomeViewController!
    var splashViewController: SplashViewController!
    let mouseMonitor = MouseMonitor.shared
    var safeMode = false
    
    // Debounce timer for window resize save
    private var windowResizeSaveTimer: Timer?
    /// 记录 Helper 启动流程开始时间，用于排查“+1s 实际触发延迟”问题。
    private var helperLaunchFlowStartAt: Date?
    
    func macOS10() -> Bool {
        if !DEBUG_MACOS10, #available(macOS 11.0, *) {
            return false
        } else {
            return true
        }
    }
    
    // MARK: - Window Setup
    /// Setup window with splash screen - 500x600
    private func setupWindowWithSplash() {
        // Create window with splash size - 500x600
        let splashRect = NSRect(x: 0, y: 0, width: 600, height: 500)
        window = NSWindow(
            contentRect: splashRect,
            styleMask: [.borderless], // Borderless for splash screen
            backing: .buffered,
            defer: false
        )
        
        // Configure window properties for splash
        window.backgroundColor = .clear  // Transparent background for rounded corners
        window.center() // Center on screen
        window.minSize = NSSize(width: 600, height: 500) // Set minimum size
        window.isMovable = true  // Allow window to be moved
        window.isMovableByWindowBackground = true  // Allow dragging by window background
        window.level = .floating // Keep on top during splash
        window.isOpaque = false  // Must be false to show rounded corners
        window.hasShadow = true
        
        // Add rounded corners to window itself
        if let contentView = window.contentView {
            contentView.wantsLayer = true
            contentView.layer?.cornerRadius = 10  // macOS standard window corner radius
            contentView.layer?.masksToBounds = true
        }
        
        // Add rounded corners and shadow effect to window
        window.contentView?.layer?.cornerRadius = 10
        window.contentView?.layer?.masksToBounds = true
        
        // Create splash view controller
        splashViewController = SplashViewController()
        
        // Set completion handler
        splashViewController.onSplashComplete = { [weak self] in
            self?.transitionToMainView()
        }
        
        // Set as window's content view controller
        window.contentViewController = splashViewController
        
        // Ensure rounded corners take effect
        DispatchQueue.main.async {
            self.window.contentView?.wantsLayer = true
            self.window.contentView?.layer?.cornerRadius = 10
            self.window.contentView?.layer?.masksToBounds = true
        }
        
        // Show window
        window.makeKeyAndOrderFront(nil)
        
        // Activate application (ensure window is frontmost)
        NSApp.activate(ignoringOtherApps: true)
    }
    
    /// Setup main window - 800x600 (original method, kept for quick debugging)
    func setupMainWindow() {
        // Load saved window settings or use defaults
        let savedSettings = SharedDataManager.shared.getWindowSettings()
        let windowWidth = savedSettings?.width ?? 800
        let windowHeight = savedSettings?.height ?? 600
        
        logger.info("setupMainWindow - Loading window settings: width=\(windowWidth), height=\(windowHeight), ratio=\(savedSettings?.fileBrowserHeightRatio ?? 0)")
        
        // Ensure saved size is at least the minimum size
        let minWidth: CGFloat = 1000
        let minHeight: CGFloat = 800
        let actualWidth = max(windowWidth, minWidth)
        let actualHeight = max(windowHeight, minHeight)
        
        if actualWidth != windowWidth || actualHeight != windowHeight {
            logger.info("setupMainWindow - Adjusted size to minimum: width=\(actualWidth), height=\(actualHeight)")
        }
        
        // Create window with MINIMUM size first
        let windowRect = NSRect(x: 0, y: 0, width: minWidth, height: minHeight)
        window = NSWindow(
            contentRect: windowRect,
            styleMask: [.titled, .closable, .miniaturizable, .resizable],
            backing: .buffered,
            defer: false
        )
        
        // Configure window properties
        window.title = "CrossShare"
        window.minSize = NSSize(width: minWidth, height: minHeight) // Set minimum size
        window.delegate = self // Set window delegate
        
        // Create main view controller
        mainViewController = MainHomeViewController()
        
        // Set as window's content view controller
        window.contentViewController = mainViewController
        
        // NOW set the actual saved size (after contentViewController is set)
        if actualWidth > minWidth || actualHeight > minHeight {
            let savedRect = NSRect(x: 0, y: 0, width: actualWidth, height: actualHeight)
            window.setFrame(savedRect, display: false)
            logger.info("setupMainWindow - Resized window to saved size: \(actualWidth)x\(actualHeight)")
        }
        
        window.center() // Center on screen AFTER everything is set up
        
        // Show window
        window.makeKeyAndOrderFront(nil)
        
        // Activate application (ensure window is frontmost)
        NSApp.activate(ignoringOtherApps: true)
        
        logger.info("setupMainWindow - Window created with final frame: \(window.frame)")
    }
    
    /// Transition from splash view to main view
    private func transitionToMainView() {
        // Create main view controller
        mainViewController = MainHomeViewController()
        
        // Change window style mask to normal window
        window.styleMask = [.titled, .closable, .miniaturizable, .resizable]
        window.level = .normal
        window.isMovable = true
        
        // Update window properties
        window.title = "CrossShare"
        let minWidth: CGFloat = 1000
        let minHeight: CGFloat = 800
        window.minSize = NSSize(width: minWidth, height: minHeight) // Set minimum size
        window.delegate = self // Set window delegate
        
        // Load saved window settings or use defaults
        let savedSettings = SharedDataManager.shared.getWindowSettings()
        let windowWidth = savedSettings?.width ?? 800
        let windowHeight = savedSettings?.height ?? 600
        
        // Ensure saved size is at least the minimum size
        let actualWidth = max(windowWidth, minWidth)
        let actualHeight = max(windowHeight, minHeight)
        
        logger.info("transitionToMainView - Loading window settings: width=\(windowWidth), height=\(windowHeight), ratio=\(savedSettings?.fileBrowserHeightRatio ?? 0)")
        if actualWidth != windowWidth || actualHeight != windowHeight {
            logger.info("transitionToMainView - Adjusted size to minimum: width=\(actualWidth), height=\(actualHeight)")
        }
        
        // Switch to main view controller FIRST
        window.contentViewController = mainViewController
        
        // THEN resize to saved size
        let mainRect = NSRect(x: 0, y: 0, width: actualWidth, height: actualHeight)
        window.setFrame(mainRect, display: true, animate: false)  // No animation
        window.center() // Center again after resize
        
        // Force window appearance update to fix title bar color
        window.invalidateShadow()
        window.display()
        window.makeKeyAndOrderFront(nil)
        
        // Force appearance refresh
        if let contentView = window.contentView {
            contentView.needsDisplay = true
        }
        splashViewController = nil
        
        logger.info("transitionToMainView - Window transitioned with final frame: \(window.frame)")
    }
    

    func applicationDidFinishLaunching(_ aNotification: Notification) {
        startRecordLog()
        HelperCommunication.shared.ensureRunInBackgroundDefaultOffIfNeeded()
        // Load configuration file
        loadConfigAndDecideStartup()
        // Start all services
        launchCrossShareHelper()
    }
    
    private func startRecordLog(){
        CSLogger.configure(processName: "App")
        // 根据编译模式设置日志级别
        #if DEBUG
        CSLogger.shared.setLogLevel(level: 0) // Debug 模式：显示所有日志（包括 debug）
        #else
        CSLogger.shared.setLogLevel(level: 1) // Release 模式：只显示 info 及以上
        #endif
        
        logger.info("========== CrossShare Application Started ==========")
        logger.info("Version: \(Bundle.main.infoDictionary?["CFBundleShortVersionString"] as? String ?? "Unknown")")

    }
    
    // MARK: - Config Loading
    
    /// Load config file and decide startup flow
    private func loadConfigAndDecideStartup() {
        let appConfig = SharedDataManager.shared.loadConfig()
        if let config = appConfig, config.uiTheme.isInitedBool {
            setupMainWindow()
        } else {
            setupWindowWithSplash()
        }
    }
    
    /// Release 路径：仅当 `installHelper` 明确返回进程已就绪时，才发起 XPC 连接。
    private func handleHelperInstallOutcomeAtStartup(_ outcome: HelperInstallOutcome) {
        switch outcome {
        case .readyForXPC:
            logger.info("Helper install outcome: ready for XPC — scheduling connection")
            DispatchQueue.main.asyncAfter(deadline: .now() + 2.0) {
                self.requestConnectToHelperIfNeeded(reason: "install_outcome_ready")
            }
        case .needsUserApproval:
            logger.info("Helper install outcome: needs user approval — skip XPC; HelperConnectionMonitor will retry after approval")
        case .registrationEnabledButProcessMissing:
            logger.warn("Helper install outcome: registered but process missing — skip XPC; user was prompted; monitor will retry")
        case .failed:
            logger.error("Helper install outcome: failed")
        }
    }
    
    /// Start all services (called after main window is shown)
    private func launchCrossShareHelper() {
        helperLaunchFlowStartAt = Date()
        CGDisplayRegisterReconfigurationCallback(displayReconfigurationCallback, nil)
        
        NSWorkspace.shared.notificationCenter.addObserver(
            self, 
            selector: #selector(sleepNotification), 
            name: NSWorkspace.willSleepNotification, 
            object: nil
        )
        NSWorkspace.shared.notificationCenter.addObserver(
            self, 
            selector: #selector(wakeNotification), 
            name: NSWorkspace.didWakeNotification, 
            object: nil
        )
        
        LoginItemManager.shared.checkAndRequestLoginItemPermissionIfFirstLaunch()
        
        let runInBackgroundEnabled = HelperCommunication.shared.isRunInBackgroundEnabled
        logger.info("Startup RunInBackground: \(runInBackgroundEnabled)")
        if runInBackgroundEnabled {
            HelperCommunication.shared.syncLoginItemRegistrationForRunInBackground(true) { success in
                logger.info(
                    "Startup login-item reconcile finished for ON mode, success=\(success)"
                )
            }
        } else {
            logger.info("Startup OFF mode: ensure helper is system-launched for stable XPC, unregister on app termination")
            HelperCommunication.shared.installHelper { outcome in
                logger.info("Startup OFF mode helper prepare result (session): \(outcome)")
            }
        }
        
        DispatchQueue.main.asyncAfter(deadline: .now() + 1.0) {
            if let startAt = self.helperLaunchFlowStartAt {
                let elapsed = Date().timeIntervalSince(startAt)
                logger.info("Helper startup delayed check fired after \(String(format: "%.2f", elapsed))s (expected ~1.0s)")
            }
            
            let loginItemOn = LoginItemManager.shared.isLoginItemEnabled
            let helperRunning = HelperCommunication.shared.isHelperProcessRunning
            
            logger.debug("Login item allowed by system: \(loginItemOn), Helper process running: \(helperRunning)")
            
            if !helperRunning {
                logger.warn("Helper not running, attempting to install/start...")
                #if DEBUG
                logger.debug("Attempting to launch Helper directly...")
                HelperCommunication.shared.launchHelper { launchSuccess in
                    if launchSuccess {
                        logger.info("Helper launched successfully")
                        DispatchQueue.main.asyncAfter(deadline: .now() + 2.0) {
                            self.requestConnectToHelperIfNeeded(reason: "debug_direct_launch")
                        }
                    } else {
                        logger.warn("Direct launch failed, attempting to install Helper...")
                        HelperCommunication.shared.installHelper { outcome in
                            self.handleHelperInstallOutcomeAtStartup(outcome)
                        }
                    }
                }
                #else
                HelperCommunication.shared.installHelper { outcome in
                    self.handleHelperInstallOutcomeAtStartup(outcome)
                }
                #endif
            } else {
                self.requestConnectToHelperIfNeeded(reason: "helper_already_running")
            }
        }
        
        checkAndInstallDaemonIfNeeded()
        
        HelperCommunication.shared.notifyHelperAppStatus(isActive: true)
        HelperConnectionMonitor.shared.startMonitoring()
        // Record the location information of the main process log.
        logger.debug("App log directory: \(getLogPath().path)")
        logger.debug("App log file: \(logger.getLogFilePath())")
#if DEBUG
        // 在DEBUG模式下, 使主线程的 unCaughtException 不被自动捕获,触发崩溃逻辑,方便定位问题.(默认逻辑不会崩溃,只是打印 log).
        UserDefaults.standard.register(defaults: ["NSApplicationCrashOnExceptions":true])
        
        print("Helper log location: \(getLogPath().path)")
        print("To view Helper logs, run: tail -f \(HelperLogViewer.shared.getTodayLogFile()?.path ?? "no log file")")
#endif

    }
    
    func applicationWillTerminate(_ aNotification: Notification) {
        logRunInBackgroundShutdownDiagnostics()
        unregisterLoginItemIfNeededForOffMode()
        HelperConnectionMonitor.shared.stopMonitoring()
    }
    
    func applicationSupportsSecureRestorableState(_ app: NSApplication) -> Bool {
        return true
    }
    
    func applicationShouldHandleReopen(_ sender: NSApplication, hasVisibleWindows flag: Bool) -> Bool {
        // When Dock icon is clicked, show window if not visible
        if !flag {
            window?.makeKeyAndOrderFront(nil)
        }
        return true
    }
    
    @objc private func sleepNotification() {
        logger.info("System going to sleep - preparing DDC services")
        // Store current state before sleep
    }
    
    @objc private func wakeNotification() {
        logger.info("System waking up - reinitializing DDC services")
    }
    
    private func connectToHelperAndStartServices() {
        if HelperClient.shared.isConnectionBusy {
            logger.info("Skip duplicate connect request: HelperClient is already connected/connecting")
            return
        }
        
        logger.info("Connecting to Helper and starting services...")
        
        HelperClient.shared.connect { success, error in
            DispatchQueue.main.async {
                if success {
                    logger.info("HelperClient connected successfully")
                    
                    P2PServiceManager.shared.checkServiceStatus { isRunning, info in
                        DispatchQueue.main.async {
                            if isRunning {
                                logger.info("P2P service is running in Helper")
                                if let info = info {
                                    logger.debug("Service info: \(info)")
                                }
                            } else {
                                logger.warn("P2P service is not running in Helper")
                            }
                        }
                    }
                } else {
                    logger.error("Failed to connect to Helper: \(error ?? "Unknown error")")
                }
            }
        }
    }
    
    /// 统一的连接入口：用于在多个触发源（启动流程、安装回调、Monitor）并发时做去重。
    private func requestConnectToHelperIfNeeded(reason: String) {
        if HelperClient.shared.isConnectionBusy {
            logger.info("Skip connect trigger (\(reason)): HelperClient is already connected/connecting")
            return
        }
        logger.info("Trigger connectToHelperAndStartServices from: \(reason)")
        connectToHelperAndStartServices()
    }
    
    private func checkAndInstallDaemonIfNeeded() {
        let installFlag = "/tmp/crossshare-needs-daemon-install"
        let validationFlag = "/tmp/crossshare-needs-daemon-validation"
        
        if FileManager.default.fileExists(atPath: installFlag) {
            logger.info("Installing Launch Daemon...")
            installLaunchDaemon()
            try? FileManager.default.removeItem(atPath: installFlag)
        } else if FileManager.default.fileExists(atPath: validationFlag) {
            logger.info("Validating Launch Daemon...")
            validateLaunchDaemon()
            try? FileManager.default.removeItem(atPath: validationFlag)
        }
    }
    
    /// Log key state at app shutdown for helper auto-restart diagnosis.
    private func logRunInBackgroundShutdownDiagnostics() {
        let runInBackgroundEnabled = HelperCommunication.shared.isRunInBackgroundEnabled
        let helperRunning = HelperCommunication.shared.isHelperProcessRunning
        
        if #available(macOS 13.0, *) {
            let detail = HelperCommunication.shared.getHelperDetailedStatus()
            logger.info(
                "App terminating diagnostics - RunInBackground=\(runInBackgroundEnabled), " +
                "helperRunning=\(helperRunning), helperStatus=\(detail.status), isUserDisabled=\(detail.isUserDisabled)"
            )
        } else {
            logger.info(
                "App terminating diagnostics - RunInBackground=\(runInBackgroundEnabled), helperRunning=\(helperRunning)"
            )
        }
    }
    
    /// OFF mode: unregister when main app terminates so helper can keep running until then.
    private func unregisterLoginItemIfNeededForOffMode() {
        let runInBackgroundEnabled = HelperCommunication.shared.isRunInBackgroundEnabled
        guard !runInBackgroundEnabled else {
            logger.info("App terminating in ON mode, keep login-item registration")
            return
        }
        
        logger.info("App terminating in OFF mode, unregister login-item from main app")
        HelperCommunication.shared.disableHelperAutoLaunchOnly { success in
            logger.info("Main app OFF-mode unregister result: \(success)")
            HelperCommunication.shared.logLoginItemStateSnapshot(context: "app-will-terminate-off-mode")
        }
    }
    
    private func installLaunchDaemon() {
        let bundlePath = Bundle.main.bundlePath 
        
        let scriptPath = "\(bundlePath)/../../install-daemon.sh"
        
        DispatchQueue.global(qos: .background).async {
            let task = Process()
            task.launchPath = "/bin/bash"
            task.arguments = [scriptPath, "--install"]
            
            let pipe = Pipe()
            task.standardOutput = pipe
            task.standardError = pipe
            
            do {
                try task.run()
                task.waitUntilExit()
                
                let data = pipe.fileHandleForReading.readDataToEndOfFile()
                let output = String(data: data, encoding: .utf8) ?? ""
                
                DispatchQueue.main.async {
                    if task.terminationStatus == 0 {
                        logger.info("Launch Daemon installed successfully")
                        if !output.isEmpty {
                            logger.debug("Install output: \(output)")
                        }
                    } else {
                        logger.error("Launch Daemon installation failed: \(task.terminationStatus)")
                        if !output.isEmpty {
                            logger.error("Install error: \(output)")
                        }
                    }
                }
            } catch {
                DispatchQueue.main.async {
                    logger.error("Failed to run daemon installation: \(error)")
                }
            }
        }
    }
    
    private func validateLaunchDaemon() {
        let bundlePath = Bundle.main.bundlePath
        
        let scriptPath = "\(bundlePath)/../../install-daemon.sh"
        
        DispatchQueue.global(qos: .background).async {
            let task = Process()
            task.launchPath = "/bin/bash"
            task.arguments = [scriptPath, "--dry-run"]
            
            let pipe = Pipe()
            task.standardOutput = pipe
            task.standardError = pipe
            
            do {
                try task.run()
                task.waitUntilExit()
                
                let data = pipe.fileHandleForReading.readDataToEndOfFile()
                let output = String(data: data, encoding: .utf8) ?? ""
                
                DispatchQueue.main.async {
                    logger.info("Launch Daemon validation result:")
                    if !output.isEmpty {
                        logger.debug("\(output)")
                    }
                }
            } catch {
                DispatchQueue.main.async {
                    logger.error("Failed to run daemon validation: \(error)")
                }
            }
        }
    }
}

// MARK: - NSWindowDelegate
extension AppDelegate: NSWindowDelegate {
    func windowShouldClose(_ sender: NSWindow) -> Bool {
        // QQ-like behavior: hide window instead of quitting app when close button is clicked
        sender.orderOut(nil)
        return false // Return false to prevent actual window closure
    }
    
    func windowWillClose(_ notification: Notification) {
        logger.debug("Window will close")
    }
    
    func windowDidResize(_ notification: Notification) {
        // Save window size when user resizes
        guard let window = notification.object as? NSWindow else { return }
        
        // Cancel any pending save timer
        windowResizeSaveTimer?.invalidate()
        
        // Use debounce - only save after user stops resizing for 0.5 seconds
        windowResizeSaveTimer = Timer.scheduledTimer(withTimeInterval: 0.5, repeats: false) { [weak self] _ in
            guard let self = self else { return }
            
            // Check if mainViewController is ready
            guard let viewController = self.mainViewController else {
                logger.debug("windowDidResize - mainViewController is nil, skipping save")
                return
            }
            
            // If initial ratio hasn't been set yet, wait a bit longer and try again
            if !viewController.hasSetInitialRatio {
                logger.debug("windowDidResize - Initial ratio not set yet, will retry in 1 second")
                self.windowResizeSaveTimer = Timer.scheduledTimer(withTimeInterval: 1.0, repeats: false) { [weak self] _ in
                    self?.saveWindowSettings(window: window, viewController: viewController)
                }
                return
            }
            
            self.saveWindowSettings(window: window, viewController: viewController)
        }
    }
    
    private func saveWindowSettings(window: NSWindow, viewController: MainHomeViewController) {
        let frame = window.frame
        
        // Get current fileBrowserHeightRatio from mainViewController
        let currentRatio = viewController.getCurrentFileBrowserHeightRatio()
        
        logger.info("Saving window settings: width: \(frame.width), height: \(frame.height), ratio: \(currentRatio)")
        
        // Save window settings
        SharedDataManager.shared.saveWindowSettings(
            width: Double(frame.width),
            height: Double(frame.height),
            fileBrowserHeightRatio: currentRatio
        )
    }
}
