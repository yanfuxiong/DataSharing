//
//  main.swift
//  CrossShareHelper
//
//  Created by ts on 2025/8/15.
//  Background running permission helper
//

import Foundation
import Cocoa
import ApplicationServices

class CrossShareHelperApp: NSObject {

    enum MainAppRestartState: Int {
        case restarting = 1
        case killByRestart = 2
        case restartDone = 3
    }
    static var restartState: MainAppRestartState = .restartDone

    private enum TerminateSource: String {
        case monitorTick = "monitor_tick"
        case appTerminateEvent = "app_terminate_event"
    }
    
    private var xpcListener: NSXPCListener?
    private var xpcService: HelperXPCService?
    private let logger = CSLogger.shared
    private var mainAppMonitorTimer: Timer?
    private var gcdTimer: DispatchSourceTimer?
    private var isTerminating = false
    private let sharedSuiteName = "group.com.instance.crossshare"
    private let runInBackgroundKey = "RunInBackgroundEnabled"
    private let mainAppBundleIdentifier = "com.realtek.crossshare.macos"
    
    func run() {
        let logPath = getLogPath()
        logger.log("Helper started - Log path: \(logPath)", level: .info)
        // Check accessibility permission on startup
        checkAndRequestAccessibilityPermission()
        
        setupXPCService()
        
        initializeGoService()
        
        setupSystemEventMonitoring()
        
        setupMainAppMonitoring()
        RunLoop.current.run()
    }
    
    private func checkAndRequestAccessibilityPermission() {
        logger.log("Checking accessibility permission on startup...", level: .info)
        
        let isTrusted = AXIsProcessTrusted()
        
        if isTrusted {
            logger.log("Accessibility permission is GRANTED", level: .info)
        } else {
            logger.log("Accessibility permission is NOT granted", level: .warn)
            // Request permission with system dialog
            let options = [kAXTrustedCheckOptionPrompt.takeUnretainedValue() as String: true] as CFDictionary
            let trusted = AXIsProcessTrustedWithOptions(options)
            
            if trusted {
                logger.log("Accessibility permission granted after request", level: .info)
            } else {
                logger.log("Opening System Settings...", level: .info)
                // Open System Settings to Accessibility pane
                DispatchQueue.main.asyncAfter(deadline: .now() + 1.0) {
                    if let url = URL(string: "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility") {
                        NSWorkspace.shared.open(url)
                    }
                }
            }
        }
    }
    
    private func setupXPCService() {
        logger.log("Setting up XPC service", level: .info)
        xpcService = HelperXPCService()
        let serviceName = "com.realtek.crossshare.macos.helper"
        
        logger.log("Creating NSXPCListener with machServiceName: \(serviceName)", level: .info)
        xpcListener = NSXPCListener(machServiceName: serviceName)
        
        guard let listener = xpcListener else {
            logger.log("Check Info.plist MachServices configuration", level: .error)
            return
        }
        
        logger.log("XPC listener created successfully", level: .info)
        listener.delegate = self
        logger.log("Delegate set to self", level: .info)
        
        listener.resume()
        logger.log("XPC listener RESUMED and ready for connections", level: .info)
        logger.log("Service name: \(serviceName)", level: .info)
        logger.log("Waiting for XPC connections from main app...", level: .info)
    }
    
    private func initializeGoService() {
        logger.log("Initializing Go service in Helper", level: .info)
        
        let config = P2PConfig.defaultConfig()
        
        xpcService?.initializeGoService(config: config.toXPCDict()) { [weak self] success, error in
            if success {
                self?.logger.log("Go service initialized successfully in Helper", level: .info)
                self?.logger.log("Service config: device=\(config.deviceName), host=\(config.listenHost), port=\(config.listenPort)", level: .info)
            } else {
                self?.logger.log("Failed to initialize Go service: \(error ?? "Unknown error")", level: .error)
            }
        }
    }
    
    private func setupSystemEventMonitoring() {
        let nc = NSWorkspace.shared.notificationCenter
        
        nc.addObserver(self,
                      selector: #selector(appDidLaunch(_:)),
                      name: NSWorkspace.didLaunchApplicationNotification,
                      object: nil)
        
        nc.addObserver(self,
                      selector: #selector(appDidTerminate(_:)),
                      name: NSWorkspace.didTerminateApplicationNotification,
                      object: nil)
        
        nc.addObserver(self,
                      selector: #selector(systemWillSleep(_:)),
                      name: NSWorkspace.willSleepNotification,
                      object: nil)
        
        nc.addObserver(self,
                      selector: #selector(systemDidWake(_:)),
                      name: NSWorkspace.didWakeNotification,
                      object: nil)
        
        print("System event monitoring setup completed")
    }
    
    private func setupMainAppMonitoring() {
        let queue = DispatchQueue(label: "com.crossShare.clipboardMonitor", qos: .background)
        gcdTimer = DispatchSource.makeTimerSource(queue: queue)
        gcdTimer?.schedule(deadline: .now(), repeating: 2.0)
        logger.log("Main app monitor started (interval: 2s)", level: .info)
        gcdTimer?.setEventHandler { [weak self] in
            self?.checkMainApp()
        }
        gcdTimer?.resume()
    }
    
    private func checkMainApp() {
        let runningApps = NSWorkspace.shared.runningApplications
        logger.log("Checking main app status at \(Date())", level: .debug)
        let isMainAppRunning = runningApps.contains { app in
            app.bundleIdentifier == mainAppBundleIdentifier
        }
        
        if isMainAppRunning {
            logger.log("Main app is running", level: .debug)
        } else {
            if CrossShareHelperApp.restartState == .restarting || CrossShareHelperApp.restartState == .killByRestart {
                logger.log("Main app is not running, but currently in restart sequence (State: \(CrossShareHelperApp.restartState)). Skip termination.", level: .debug)
                return
            }

            logger.log("Main app is not running", level: .warn)
            if !isRunInBackgroundEnabled() {
                logger.log("RunInBackground is OFF, terminate helper", level: .info)
                terminateHelperBecauseMainAppClosed(source: .monitorTick)
            }
        }
    }
    
    private func isRunInBackgroundEnabled() -> Bool {
        let sharedDefaults = UserDefaults(suiteName: sharedSuiteName)
        if sharedDefaults?.object(forKey: runInBackgroundKey) == nil {
            logger.log("RunInBackground setting missing in App Group, defaulting to false (OFF)", level: .info)
            return false
        }
        let enabled = sharedDefaults?.bool(forKey: runInBackgroundKey) ?? false
        logger.log("RunInBackground setting loaded from App Group: \(enabled)", level: .info)
        return enabled
    }
    
    private func terminateHelperBecauseMainAppClosed(source: TerminateSource) {
        guard !isTerminating else {
            logger.log("Helper terminate already in progress, ignore duplicate source=\(source.rawValue)", level: .warn)
            return
        }
        isTerminating = true
        
        let runInBackgroundEnabled = isRunInBackgroundEnabled()
        logger.log(
            "Start helper termination source=\(source.rawValue), RunInBackground=\(runInBackgroundEnabled), mainAppBundle=\(mainAppBundleIdentifier)",
            level: .info
        )
        DispatchQueue.main.async {
            self.logger.log("Terminating helper process because RunInBackground is OFF and main app is not running", level: .info)
            exit(0)
        }
    }
    
    private func launchMainAppIfNeeded() {
        if let appURL = NSWorkspace.shared.urlForApplication(withBundleIdentifier: mainAppBundleIdentifier) {
            let configuration = NSWorkspace.OpenConfiguration()
            configuration.createsNewApplicationInstance = false
            configuration.activates = false
            configuration.hides = true
            
            NSWorkspace.shared.openApplication(at: appURL,
                                              configuration: configuration,
                                              completionHandler: nil)
            print("Launched main app in background")
        }
    }
    
    @objc private func appDidLaunch(_ notification: Notification) {
        if let app = notification.userInfo?[NSWorkspace.applicationUserInfoKey] as? NSRunningApplication {
            if app.bundleIdentifier == mainAppBundleIdentifier {
                print("Main app launched: \(app.localizedName ?? "CrossShare")")

                if CrossShareHelperApp.restartState != .restartDone {
                    logger.log("Main app successfully launched. State -> restartDone.", level: .info)
                    CrossShareHelperApp.restartState = .restartDone
                }
            }
        }
    }
    
    @objc private func appDidTerminate(_ notification: Notification) {
        if let app = notification.userInfo?[NSWorkspace.applicationUserInfoKey] as? NSRunningApplication {
            if app.bundleIdentifier == mainAppBundleIdentifier {
                if CrossShareHelperApp.restartState == .restarting {
                    logger.log("Main app terminated due to restart. State -> killByRestart.", level: .info)
                    CrossShareHelperApp.restartState = .killByRestart
                    return
                }

                print("Main app terminated: \(app.localizedName ?? "CrossShare")")
                let runInBackgroundEnabled = isRunInBackgroundEnabled()
                logger.log("Received main app terminated event, RunInBackground: \(runInBackgroundEnabled)", level: .info)
                if !runInBackgroundEnabled {
                    logger.log("Main app terminated and RunInBackground is OFF", level: .info)
                    terminateHelperBecauseMainAppClosed(source: .appTerminateEvent)
                } else {
                    logger.log("Main app terminated but RunInBackground is ON, helper keeps running", level: .info)
                }
                // Optionally restart the app after termination
                // DispatchQueue.main.asyncAfter(deadline: .now() + 2.0) {
                //     self.launchMainAppIfNeeded()
                // }
            }
        }
    }
    
    @objc private func systemWillSleep(_ notification: Notification) {
        print("System will sleep at \(Date())")
    }
    
    @objc private func systemDidWake(_ notification: Notification) {
        print("System did wake at \(Date())")
        DispatchQueue.main.asyncAfter(deadline: .now() + 5.0) { [weak self] in
            self?.checkMainApp()
        }
    }
}

extension CrossShareHelperApp: NSXPCListenerDelegate {
    
    func listener(_ listener: NSXPCListener, shouldAcceptNewConnection newConnection: NSXPCConnection) -> Bool {
        logger.log("NEW XPC CONNECTION REQUEST", level: .info)
        logger.log("From PID: \(newConnection.processIdentifier)", level: .info)
        logger.log("Effective UID: \(newConnection.effectiveUserIdentifier)", level: .info)
        logger.log("Effective GID: \(newConnection.effectiveGroupIdentifier)", level: .info)
        
        newConnection.exportedInterface = NSXPCInterface(with: CrossShareHelperXPCProtocol.self)
        newConnection.exportedObject = xpcService
        
        newConnection.remoteObjectInterface = NSXPCInterface(with: CrossShareHelperXPCDelegate.self)
        
        newConnection.invalidationHandler = { [weak self] in
            self?.logger.log("XPC connection invalidated (PID: \(newConnection.processIdentifier))", level: .info)
            self?.xpcService?.removeXPCConnection(newConnection)
        }

        newConnection.interruptionHandler = { [weak self] in
            self?.logger.log("XPC connection interrupted (PID: \(newConnection.processIdentifier))", level: .warn)
        }

        newConnection.resume()
        logger.log("Connection resumed and accepted", level: .info)

        // Add the connection to track it for callbacks
        xpcService?.addXPCConnection(newConnection)
        
        return true
    }
}

autoreleasepool {
    let helperApp = CrossShareHelperApp()
    helperApp.run()
}
