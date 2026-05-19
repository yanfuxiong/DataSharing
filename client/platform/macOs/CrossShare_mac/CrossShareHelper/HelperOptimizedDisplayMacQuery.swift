//
//  HelperOptimizedDisplayMacQuery.swift
//  CrossShareHelper
//
//  Helper version of OptimizedDisplayMacQuery without HelperClient dependencies
//

import Foundation
import CoreGraphics
import AppKit

class HelperOptimizedDisplayMacQuery {
    private static let sharedInstance = HelperOptimizedDisplayMacQuery()
    /// Per-display retry timers, keyed by CGDirectDisplayID
    private var displayTimers: [CGDirectDisplayID: Timer] = [:]
    /// Displays that have already been found (MAC obtained successfully)
    private var foundDisplays: Set<CGDirectDisplayID> = []
    private let displayManager = CrossShareVcpCtrl()
    private let logger = CSLogger.shared
    
    public var macToDisplayID: [String: CGDirectDisplayID] = [:]
    public var displayIDToMac: [CGDirectDisplayID: String] = [:]
    
    private init() {}
    
    class func shared() -> HelperOptimizedDisplayMacQuery {
        return sharedInstance
    }
    
    func getDisplayID(forMac mac: String) -> CGDirectDisplayID? {
        return macToDisplayID[mac]
    }
    
    func getMacAddress(forDisplayID displayID: CGDirectDisplayID) -> String? {
        return displayIDToMac[displayID]
    }
    
    private func updateMacMapping(mac: String, displayID: CGDirectDisplayID) {
        macToDisplayID[mac] = displayID
        displayIDToMac[displayID] = mac
        logger.log("Updated MAC mapping: \(mac) <-> \(displayID)", level: .info)
    }
    
    func queryDisplay(displayID: CGDirectDisplayID, displayName: String) {
        logger.log("Querying display: \(displayName) (ID: \(displayID))", level: .info)
        // Skip if this display has already been found
        if foundDisplays.contains(displayID) {
            logger.log("Display \(displayID) already found, skipping query", level: .debug)
            return
        }
        if let macAddress = displayManager.getSmartMonitorMacAddr(for: displayID) {
            foundDisplays.insert(displayID)
            stopTimer(for: displayID)
            // Get Source + Port via 0xE1/0xE2
            let macPortResult = displayManager.getMacAddressAndPort(for: displayID)
            let source = macPortResult?.source ?? 0
            let port = macPortResult?.port ?? 0
            handleMacAddressFound(mac: macAddress, displayID: displayID, displayName: displayName, source: source, port: port)
        } else {
            startRetryTimer(for: displayID, displayName: displayName)
        }
    }
    
    private func handleMacAddressFound(mac: String, displayID: CGDirectDisplayID, displayName: String, source: UInt8, port: UInt8) {
        logger.log("Successfully obtained MAC: \(mac), Source: \(source), Port: \(port) for display: \(displayName)", level: .info)
        updateMacMapping(mac: mac, displayID: displayID)
        
        GoCallbackManager.shared.updateDisplayMapping(mac: mac, displayID: displayID)
        GoCallbackManager.shared.setCurrentDIASMac(mac)
        
        GoServiceBridge.shared.setDIASID(diasID: mac) { [weak self] success, error in
            if success {
                self?.logger.log("Successfully set DIAS ID in Go service", level: .info)
                // Store the DIAS display ID when successfully set
                HelperDisplayManager.shared.setCurrentDIASDisplayID(displayID)
                // Get theme code and update config
                self?.getThemeAndUpdateConfig(mac: mac, displayID: displayID)
            } else {
                self?.logger.log("Failed to set DIAS ID: \(error ?? "Unknown"), but request may be queued", level: .warn)
            }
        }
        
        // Trigger multi-display event handling: allocate UDP ports, start servers, call SetDisplayEventInfo
        HelperDisplayManager.shared.handleDIASDisplayPlugIn(displayID: displayID, macAddress: mac, source: source, port: port)
    }
    
    /// Get theme code from monitor and update config file
    private func getThemeAndUpdateConfig(mac: String, displayID: CGDirectDisplayID) {
        let themeCode = getThemeCode(displayID: displayID)
        logger.log("Retrieved theme code: \(themeCode) for MAC: \(mac)", level: .info)
        updateAppConfig(themeCode: themeCode)
    }
    
    /// Get theme code from displayID using CrossShareVcpCtrl
    private func getThemeCode(displayID: CGDirectDisplayID) -> String {
        if let themeCode = displayManager.getCustomerThemeCode(for: displayID) {
            let customerIdString = String(themeCode.customerId)
            logger.log("Got theme code from monitor - customerId: \(customerIdString), styleId: \(themeCode.styleId)", level: .info)
            return customerIdString
        } else {
            logger.log("Failed to get theme code for displayID: \(displayID)", level: .warn)
            return "0"
        }
    }
    
    /// Update app config file with theme code (only if changed)
    private func updateAppConfig(themeCode: String) {
        let config = SharedDataManager.shared.loadConfig() ?? SharedDataManager.shared.createDefaultConfig()
        let currentCustomerID = config.uiTheme.customerID
        logger.log("Theme code changed: \(currentCustomerID) → \(themeCode)", level: .info)
        var tconfig = config
        tconfig.uiTheme.customerID = themeCode
        SharedDataManager.shared.saveConfig(tconfig)
        
        logger.info("currentCustomerID:\(currentCustomerID) themeCode:\(themeCode) ")
        let themeInfoData: [String: Any] = [
            "customerId": themeCode,
            "isInited": tconfig.uiTheme.isInited,
        ]
        if currentCustomerID == themeCode {
            GoCallbackManager.shared.handleThemeInfoUpdate(themeInfoData)
        }else{
            restartMainApp()
        }
    }
    
    /// Restart the main CrossShare application
    private func restartMainApp() {
        logger.log("Terminating main CrossShare application...", level: .info)
        
        CrossShareHelperApp.restartState = .restarting

        // 使用命令行方式终止进程，避免 crash
        let process = Process()
        process.executableURL = URL(fileURLWithPath: "/usr/bin/killall")
        process.arguments = ["CrossShare"]
        
        do {
            try process.run()
            process.waitUntilExit()
            
            let exitCode = process.terminationStatus
            if exitCode == 0 {
                self.logger.log("Main app terminated successfully (exit code: \(exitCode))", level: .info)
            } else {
                self.logger.log("killall exit code: \(exitCode) (process may not be running)", level: .warn)
            }
            
            // 等待一秒后重新启动主应用
            DispatchQueue.main.asyncAfter(deadline: .now() + 1.0) {
                self.logger.log("Launching main app...", level: .info)
                self.launchMainApp(bundleID: BundleIdentifiers.mainApp)
            }
            
        } catch {
            self.logger.log("Failed to execute killall: \(error.localizedDescription)", level: .error)
            CrossShareHelperApp.restartState = .restartDone
        }
    }
    
    /// Launch the main CrossShare application
    private func launchMainApp(bundleID: String) {
        let workspace = NSWorkspace.shared
        
        // 优先从 /Applications 目录查找应用
        let applicationsPath = "/Applications/CrossShare.app"
        let appURL = URL(fileURLWithPath: applicationsPath)
        
        if FileManager.default.fileExists(atPath: applicationsPath) {
            // 应用在 /Applications 目录下，直接启动
            self.logger.log("Found app at: \(applicationsPath)", level: .info)
            launchApplication(at: appURL, workspace: workspace)
        } else {
            // 在 /Applications 目录未找到，尝试使用 bundle ID 查找
            self.logger.log("App not found at \(applicationsPath), searching by bundle ID...", level: .warn)
            
            if let foundURL = workspace.urlForApplication(withBundleIdentifier: bundleID) {
                self.logger.log("Found app at: \(foundURL.path)", level: .info)
                launchApplication(at: foundURL, workspace: workspace)
            } else {
                logger.log("Could not find main app with bundle ID: \(bundleID)", level: .error)
            }
        }
    }
    
    /// Helper method to launch application at given URL
    private func launchApplication(at appURL: URL, workspace: NSWorkspace) {
        let configuration = NSWorkspace.OpenConfiguration()
        configuration.activates = true
        
        workspace.openApplication(at: appURL, configuration: configuration) { app, error in
            if let error = error {
                self.logger.log("Failed to launch main app: \(error.localizedDescription)", level: .error)
            } else {
                self.logger.log("Main app launched successfully at: \(appURL.path)", level: .info)
            }
        }
    }
    
    private func startRetryTimer(for displayID: CGDirectDisplayID, displayName: String) {
        // Stop existing timer for this display if any
        stopTimer(for: displayID)
        
        logger.log("Starting retry timer for display: \(displayName) (ID: \(displayID))", level: .info)
        
        let timer = Timer(timeInterval: 2.0, repeats: true) { [weak self] timer in
            guard let strongSelf = self else {
                timer.invalidate()
                return
            }
            
            if let mac = strongSelf.displayManager.getSmartMonitorMacAddr(for: displayID) {
                // Get Source + Port via 0xE1/0xE2
                let macPortResult = strongSelf.displayManager.getMacAddressAndPort(for: displayID)
                let source = macPortResult?.source ?? 0
                let port = macPortResult?.port ?? 0
                strongSelf.foundDisplays.insert(displayID)
                strongSelf.stopTimer(for: displayID)
                strongSelf.handleMacAddressFound(mac: mac, displayID: displayID, displayName: displayName, source: source, port: port)
            }
        }
        
        RunLoop.main.add(timer, forMode: .common)
        displayTimers[displayID] = timer
    }
    
    /// Stop the retry timer for a specific display
    func stopTimer(for displayID: CGDirectDisplayID) {
        displayTimers[displayID]?.invalidate()
        displayTimers.removeValue(forKey: displayID)
    }
    
    /// Stop timers and clear found state for specific displays (called when displays are removed)
    func cleanupDisplays(_ displayIDs: Set<CGDirectDisplayID>) {
        for displayID in displayIDs {
            stopTimer(for: displayID)
            foundDisplays.remove(displayID)
            // Clean up MAC mappings
            if let mac = displayIDToMac.removeValue(forKey: displayID) {
                macToDisplayID.removeValue(forKey: mac)
            }
        }
    }
    
    /// Stop all timers and reset all state
    func stopAllTimers() {
        displayTimers.values.forEach { $0.invalidate() }
        displayTimers.removeAll()
        foundDisplays.removeAll()
    }
    
    deinit {
        stopAllTimers()
    }
}
