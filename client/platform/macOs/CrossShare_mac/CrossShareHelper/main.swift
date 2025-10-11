//
//  main.swift
//  CrossShareHelper
//
//  Created by ts on 2025/8/15.
//  Background running permission helper
//

import Foundation
import Cocoa

class CrossShareHelperApp: NSObject {
    
    private var xpcListener: NSXPCListener?
    private var xpcService: HelperXPCService?
    private let logger = XPCLogger.shared
    private var mainAppMonitorTimer: Timer?
    private var gcdTimer: DispatchSourceTimer?
    
    func run() {
        print("CrossShare Helper started at \(Date())")
        
        let logPath = FileManager.default.urls(for: .applicationSupportDirectory, in: .userDomainMask).first!
            .appendingPathComponent("CrossShare/Logs/Helper").path
        print("Helper log files location: \(logPath)")
        logger.log("Helper started - Log path: \(logPath)", level: .info)
        
        setupXPCService()
        
        initializeGoService()
        
        setupSystemEventMonitoring()
        
        setupMainAppMonitoring()
        
        print("CrossShare Helper setup completed, entering run loop...")
        logger.log("Helper setup completed, entering run loop", level: .info)
        
        RunLoop.current.run()
    }
    
    private func setupXPCService() {
        logger.log("Setting up XPC service", level: .info)
        xpcService = HelperXPCService()
        let serviceName = "com.realtek.crossshare.macos.helper"
        
        xpcListener = NSXPCListener(machServiceName: serviceName)
        
        if let listener = xpcListener {
            listener.delegate = self
            listener.resume()
            logger.log("XPC listener started for service: \(serviceName)", level: .info)
        } else {
            logger.log("Failed to create XPC listener", level: .error)
        }
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
        gcdTimer?.schedule(deadline: .now(), repeating: 30.0)
        gcdTimer?.setEventHandler { [weak self] in
            self?.checkMainApp()
        }
        gcdTimer?.resume()
    }
    
    private func checkMainApp() {
        let mainAppBundleID = "com.realtek.crossshare.macos"
        let runningApps = NSWorkspace.shared.runningApplications
        
        logger.log("Checking main app status at \(Date())", level: .debug)
        
        let isMainAppRunning = runningApps.contains { app in
            app.bundleIdentifier == mainAppBundleID
        }
        
        if isMainAppRunning {
            logger.log("Main app is running", level: .debug)
        } else {
            print("Main app is not running")
            logger.log("Main app is not running", level: .warn)
        }
    }
    
    private func launchMainAppIfNeeded() {
        let mainAppBundleID = "com.realtek.crossshare.macos"
        
        if let appURL = NSWorkspace.shared.urlForApplication(withBundleIdentifier: mainAppBundleID) {
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
            if app.bundleIdentifier == "com.realtek.crossshare.macos" {
                print("Main app launched: \(app.localizedName ?? "CrossShare")")
            }
        }
    }
    
    @objc private func appDidTerminate(_ notification: Notification) {
        if let app = notification.userInfo?[NSWorkspace.applicationUserInfoKey] as? NSRunningApplication {
            if app.bundleIdentifier == "com.realtek.crossshare.macos" {
                print("Main app terminated: \(app.localizedName ?? "CrossShare")")
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
        logger.log("New XPC connection request from pid: \(newConnection.processIdentifier)", level: .info)
        
        newConnection.exportedInterface = NSXPCInterface(with: CrossShareHelperXPCProtocol.self)
        newConnection.exportedObject = xpcService
        
        newConnection.remoteObjectInterface = NSXPCInterface(with: CrossShareHelperXPCDelegate.self)
        
        newConnection.invalidationHandler = { [weak self] in
            self?.logger.log("XPC connection invalidated", level: .info)
            self?.xpcService?.removeXPCConnection(newConnection)
        }

        newConnection.interruptionHandler = { [weak self] in
            self?.logger.log("XPC connection interrupted", level: .warn)
        }

        newConnection.resume()

        // Add the connection to track it for callbacks
        xpcService?.addXPCConnection(newConnection)
        
        return true
    }
}

autoreleasepool {
    let helperApp = CrossShareHelperApp()
    helperApp.run()
}
