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
    let mouseMonitor = MouseMonitor.shared
    var safeMode = false
    
    func macOS10() -> Bool {
        if !DEBUG_MACOS10, #available(macOS 11.0, *) {
            return false
        } else {
            return true
        }
    }
    
    // MARK: - Window Setup
    /// Setup main window - 800x600
    private func setupMainWindow() {
        // Create window - 800x600
        let windowRect = NSRect(x: 0, y: 0, width: 800, height: 600)
        window = NSWindow(
            contentRect: windowRect,
            styleMask: [.titled, .closable, .miniaturizable, .resizable],
            backing: .buffered,
            defer: false
        )
        
        // Configure window properties
        window.title = "CrossShare"
        window.center() // Center on screen
        window.minSize = NSSize(width: 1000, height: 800) // Set minimum size
        window.delegate = self // Set window delegate
        
        // Create main view controller
        mainViewController = MainHomeViewController()
        
        // Set as window's content view controller
        window.contentViewController = mainViewController
        
        // Show window
        window.makeKeyAndOrderFront(nil)
        
        // Activate application (ensure window is frontmost)
        NSApp.activate(ignoringOtherApps: true)
    }
    

    func applicationDidFinishLaunching(_ aNotification: Notification) {
        CSLogger.configure(processName: "App")
        
        // 根据编译模式设置日志级别
        #if DEBUG
        CSLogger.shared.setLogLevel(level: 0) // Debug 模式：显示所有日志（包括 debug）
        #else
        CSLogger.shared.setLogLevel(level: 1) // Release 模式：只显示 info 及以上
        #endif
        
        logger.info("========== CrossShare Application Started ==========")
        logger.info("Version: \(Bundle.main.infoDictionary?["CFBundleShortVersionString"] as? String ?? "Unknown")")
        
        // Setup main window
        setupMainWindow()
        
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
        
        // Check and request background running permissions on startup
        LoginItemManager.shared.checkAndRequestLoginItemPermissionIfFirstLaunch()
        LoginItemManager.shared.checkHelperStatus()
        
        DispatchQueue.main.asyncAfter(deadline: .now() + 1.0) {
            let helperEnabled = LoginItemManager.shared.isLoginItemEnabled
            let helperRunning = HelperCommunication.shared.isHelperRunning
            
            logger.debug("Helper enabled: \(helperEnabled), Helper running: \(helperRunning)")
            
            if !helperRunning {
                logger.warn("Helper not running, attempting to install/start...")
                #if DEBUG
                logger.debug("Attempting to launch Helper directly...")
                HelperCommunication.shared.launchHelper { launchSuccess in
                    if launchSuccess {
                        logger.info("Helper launched successfully")
                        DispatchQueue.main.asyncAfter(deadline: .now() + 2.0) {
                            self.connectToHelperAndStartServices()
                        }
                    } else {
                        logger.warn("Direct launch failed, attempting to install Helper...")
                        HelperCommunication.shared.installHelper { success in
                            if success {
                                logger.info("Helper install succeeded")
                            } else {
                                logger.error("Helper install failed")
                            }
                            if success {
                                DispatchQueue.main.asyncAfter(deadline: .now() + 2.0) {
                                    self.connectToHelperAndStartServices()
                                }
                            }
                        }
                    }
                }
                #else
                HelperCommunication.shared.installHelper { success in
                    if success {
                        logger.info("Helper install succeeded")
                    } else {
                        logger.error("Helper install failed")
                    }
                    if success {
                        DispatchQueue.main.asyncAfter(deadline: .now() + 2.0) {
                            self.connectToHelperAndStartServices()
                        }
                    }
                }
                #endif
            } else {
                self.connectToHelperAndStartServices()
            }
        }
        
        checkAndInstallDaemonIfNeeded()
        
        HelperCommunication.shared.notifyHelperAppStatus(isActive: true)
        
        // Record the location information of the main process log.
        logger.debug("App log directory: \(getLogPath().path)")
        logger.debug("App log file: \(logger.getLogFilePath())")
    }
    
    func applicationWillTerminate(_ aNotification: Notification) {
        // Insert code here to tear down your application
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
}
