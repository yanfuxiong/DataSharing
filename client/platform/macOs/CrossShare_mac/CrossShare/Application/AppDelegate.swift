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
import Sparkle
import CoreGraphics

func displayReconfigurationCallback(display: CGDirectDisplayID, flags: CGDisplayChangeSummaryFlags, userInfo: UnsafeMutableRawPointer?) {
    print("🔧 Display reconfiguration detected (display: \(display), flags: \(flags.rawValue))")
}

class AppDelegate: NSObject, NSApplicationDelegate {
    
    var reconfigureID: Int = 0 // dispatched reconfigure command ID
    var sleepID: Int = 0 // sleep event ID
    
    var window: NSWindow!
    var mainAppWVC:NSWindowController!
    let mouseMonitor = MouseMonitor.shared
    var safeMode = false
    
    func macOS10() -> Bool {
        if !DEBUG_MACOS10, #available(macOS 11.0, *) {
            return false
        } else {
            return true
        }
    }

    func applicationDidFinishLaunching(_ aNotification: Notification) {
        let sb = NSStoryboard(name: NSStoryboard.Name("Main"), bundle: nil)
        mainAppWVC = sb.instantiateController(withIdentifier: NSStoryboard.SceneIdentifier("MainAppWVC")) as? NSWindowController
        mainAppWVC?.showWindow(self)
        
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
            
            print("DEBUG: Helper enabled: \(helperEnabled), Helper running: \(helperRunning)")
            
            if !helperRunning {
                print("Helper not running, attempting to install/start...")
                #if DEBUG
                print("DEBUG: Attempting to launch Helper directly...")
                HelperCommunication.shared.launchHelper { launchSuccess in
                    if launchSuccess {
                        print("DEBUG: Helper launched successfully")
                        DispatchQueue.main.asyncAfter(deadline: .now() + 2.0) {
                            self.connectToHelperAndStartServices()
                        }
                    } else {
                        print("DEBUG: Direct launch failed, attempting to install Helper...")
                        HelperCommunication.shared.installHelper { success in
                            print("Helper install result: \(success)")
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
                    print("Helper install result: \(success)")
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
        
#if DEBUG
        // 在DEBUG模式下, 使主线程的 unCaughtException 不被自动捕获,触发崩溃逻辑,方便定位问题.(默认逻辑不会崩溃,只是打印 log).
        UserDefaults.standard.register(defaults: ["NSApplicationCrashOnExceptions":true])
        
        print("Helper log location: \(HelperLogViewer.shared.getLogPath())")
        print("To view Helper logs, run: tail -f \(HelperLogViewer.shared.getTodayLogFile()?.path ?? "no log file")")
#endif
    }
    
    func applicationWillTerminate(_ aNotification: Notification) {
        // Insert code here to tear down your application
        mouseMonitor.stopMonitoring()
    }
    
    func applicationSupportsSecureRestorableState(_ app: NSApplication) -> Bool {
        return true
    }
    
    @objc private func sleepNotification() {
        print("System going to sleep - preparing DDC services")
        // Store current state before sleep
    }
    
    @objc private func wakeNotification() {
        print("System waking up - reinitializing DDC services")
    }
    
    private func connectToHelperAndStartServices() {
        print("Connecting to Helper and starting services...")
        
        HelperClient.shared.connect { success, error in
            DispatchQueue.main.async {
                if success {
                    print("HelperClient connected successfully")
                    
                    P2PServiceManager.shared.checkServiceStatus { isRunning, info in
                        DispatchQueue.main.async {
                            if isRunning {
                                print("P2P service is running in Helper")
                                if let info = info {
                                    print("Service info: \(info)")
                                }
                            } else {
                                print("P2P service is not running in Helper")
                            }
                        }
                    }
                } else {
                    print("Failed to connect to Helper: \(error ?? "Unknown error")")
                    print("Clipboard monitoring will not work until Helper connection is established")
                }
            }
        }
    }
    
    private func checkAndInstallDaemonIfNeeded() {
        let installFlag = "/tmp/crossshare-needs-daemon-install"
        let validationFlag = "/tmp/crossshare-needs-daemon-validation"
        
        if FileManager.default.fileExists(atPath: installFlag) {
            print("Installing Launch Daemon...")
            installLaunchDaemon()
            try? FileManager.default.removeItem(atPath: installFlag)
        } else if FileManager.default.fileExists(atPath: validationFlag) {
            print("🔍 Validating Launch Daemon...")
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
                        print("Launch Daemon installed successfully")
                        print(output)
                    } else {
                        print("Launch Daemon installation failed: \(task.terminationStatus)")
                        print(output)
                    }
                }
            } catch {
                DispatchQueue.main.async {
                    print("Failed to run daemon installation: \(error)")
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
                    print("Launch Daemon validation result:")
                    print(output)
                }
            } catch {
                DispatchQueue.main.async {
                    print("Failed to run daemon validation: \(error)")
                }
            }
        }
    }
    
}
