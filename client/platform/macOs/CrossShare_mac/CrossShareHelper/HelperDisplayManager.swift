//
//  HelperDisplayManager.swift
//  CrossShareHelper
//
//  Manages display detection and DDCCI communication in Helper process
//

import Foundation
import CoreGraphics

class HelperDisplayManager {
    static let shared = HelperDisplayManager()
    
    private let logger = XPCLogger.shared
    private let vcpCtrl = CrossShareVcpCtrl()
    private var gcdTimer: DispatchSourceTimer?
    
    // Cache for display state to avoid unnecessary DIAS checks
    private var lastDisplayIDs: Set<CGDirectDisplayID> = []
    private var checkedDIASDisplays: Set<CGDirectDisplayID> = []
    
    private init() {
        DisplayManager.shared.configureDisplays()
        DisplayManager.shared.updateArm64AVServices()
        
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(displayConfigurationChanged(_:)),
            name: .displayConfigurationChangedNotification,
            object: nil
        )
    }
    
    deinit {
        NotificationCenter.default.removeObserver(self)
        stopMonitoring()
    }
    
    func startMonitoring() {
        logger.log("Starting display monitoring in Helper", level: .info)
        
        ensureDisplayManagerConfigured {
            self.checkDisplays()
            
            let queue = DispatchQueue(label: "com.crossShare.display.queue", qos: .background)
            self.gcdTimer = DispatchSource.makeTimerSource(queue: queue)
            self.gcdTimer?.schedule(deadline: .now(), repeating: 5.0)
            self.gcdTimer?.setEventHandler { [weak self] in
                self?.checkDisplays()
            }
            self.gcdTimer?.resume()
        }
    }
    
    func stopMonitoring() {
        logger.log("Stopping display monitoring in Helper", level: .info)
        gcdTimer?.cancel()
        gcdTimer = nil
    }
    
    func checkDisplaysNow() {
        checkDisplays()
    }
    
    func resetDIASCache() {
        logger.log("Resetting DIAS cache", level: .info)
        checkedDIASDisplays.removeAll()
        lastDisplayIDs.removeAll()
    }
    
    func forceCheckAllDisplaysForDIAS() {
        logger.log("Forcing DIAS check for all displays", level: .info)
        resetDIASCache()
        checkDisplays()
    }
    
    private func checkDisplays() {
        let maxDisplays: UInt32 = 10
        var activeDisplays = [CGDirectDisplayID](repeating: 0, count: Int(maxDisplays))
        var displayCount: UInt32 = 0
        
        // Use online display list so mirrored displays are also enumerated (CGGetActiveDisplayList collapses mirrors)
        let result = CGGetOnlineDisplayList(maxDisplays, &activeDisplays, &displayCount)
        guard result == .success else {
            logger.log("Failed to get active display list", level: .error)
            return
        }
        
        if displayCount == 0 {
            logger.log("No displays found", level: .debug)
            lastDisplayIDs.removeAll()
            checkedDIASDisplays.removeAll()
            return
        }
        
        var currentDisplayIDs: Set<CGDirectDisplayID> = []
        for i in 0..<Int(displayCount) {
            let displayID = activeDisplays[i]
            if displayID != 0 {
                currentDisplayIDs.insert(displayID)
            }
        }

        if currentDisplayIDs != lastDisplayIDs {
            logger.log("Display configuration changed, performing DIAS checks", level: .info)
            let newDisplays = currentDisplayIDs.subtracting(lastDisplayIDs)
            let removedDisplays = lastDisplayIDs.subtracting(currentDisplayIDs)
            HelperOptimizedDisplayMacQuery.shared().foundMacAddress = false
            HelperOptimizedDisplayMacQuery.shared().stopAllTimers()
            if !newDisplays.isEmpty {
                logger.log("New displays detected: \(Array(newDisplays))", level: .info)
                refreshDisplayConfiguration()
                for displayID in newDisplays {
                    checkDisplayForDIAS(displayID: displayID)
                }
            }
            
            if !removedDisplays.isEmpty {
                logger.log("Displays removed: \(Array(removedDisplays))", level: .info)
                checkedDIASDisplays.subtract(removedDisplays)
            }
            lastDisplayIDs = currentDisplayIDs
        } else {
            logger.log("Display configuration unchanged, skipping DIAS checks", level: .debug)
        }
    }
    
    private func checkDisplayForDIAS(displayID: CGDirectDisplayID) {
        guard displayID != 0 else {
            logger.log("Invalid display ID (0), skipping DIAS check", level: .warn)
            return
        }
        
        if CGDisplayIsBuiltin(displayID) != 0 {
            logger.log("Display \(displayID) is built-in, skipping DIAS check", level: .debug)
            return
        }
        
        if checkedDIASDisplays.contains(displayID) {
            logger.log("Display \(displayID) already checked for DIAS, skipping", level: .debug)
            return
        }
        
        checkedDIASDisplays.insert(displayID)
        
        DispatchQueue.main.async { [weak self] in
            guard let self = self else { return }
            
            var displayName = "Display \(displayID)"
            
            if let displayInfoRef = CoreDisplay_DisplayCreateInfoDictionary(displayID) {
                let displayInfo = displayInfoRef.takeRetainedValue() as NSDictionary
                if let nameDict = displayInfo["DisplayProductName"] as? [String: String],
                   let name = nameDict["en_US"] ?? nameDict.first?.value {
                    displayName = name
                }
            }
            HelperOptimizedDisplayMacQuery.shared().queryDisplay(
                displayID: displayID,
                displayName: displayName
            )
        }
    }
    
    func queryAuthStatus(for displayID: CGDirectDisplayID, index: UInt16, completion: @escaping (Bool) -> Void) {
        logger.log("Querying auth status for display \(displayID) with index \(index)", level: .info)
        vcpCtrl.querySmartMonitorAuthStatus(for: displayID, index: index, completed: completion)
    }
    
    func getPortInfo(for displayID: CGDirectDisplayID) -> (source: UInt8, port: UInt8)? {
        logger.log("Getting port info for display \(displayID)", level: .info)
        return vcpCtrl.getConnectedPortInfo(for: displayID)
    }
    
    private func ensureDisplayManagerConfigured(completion: @escaping () -> Void) {
        if !DisplayManager.shared.getAllDisplays().isEmpty {
            completion()
            return
        }
        
        DisplayManager.shared.configureDisplays()
        DisplayManager.shared.updateArm64AVServices()
        
        completion()
    }
    
    private func refreshDisplayConfiguration() {
        DisplayManager.shared.configureDisplays()
        DisplayManager.shared.updateArm64AVServices()
    }
    
    @objc private func displayConfigurationChanged(_ notification: Notification) {
        if let userInfo = notification.userInfo {
            if let addedDisplays = userInfo["addedDisplays"] as? [CGDirectDisplayID] {
                for displayID in addedDisplays {
                    checkedDIASDisplays.remove(displayID)
                }
            }
            if let removedDisplays = userInfo["removedDisplays"] as? [CGDirectDisplayID] {
                logger.log("Removed displays from notification: \(removedDisplays)", level: .info)
                checkedDIASDisplays.subtract(removedDisplays)
            }
        }
        DispatchQueue.main.asyncAfter(deadline: .now() + 1.0) { [weak self] in
            self?.checkDisplays()
        }
    }
}
