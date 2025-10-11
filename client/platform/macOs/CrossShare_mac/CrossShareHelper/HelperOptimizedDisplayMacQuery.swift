//
//  HelperOptimizedDisplayMacQuery.swift
//  CrossShareHelper
//
//  Helper version of OptimizedDisplayMacQuery without HelperClient dependencies
//

import Foundation
import CoreGraphics

class HelperOptimizedDisplayMacQuery {
    private static let sharedInstance = HelperOptimizedDisplayMacQuery()
    private var timers: [Timer] = []
    var foundMacAddress = false
    private let displayManager = CrossShareVcpCtrl()
    private let logger = XPCLogger.shared
    
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
        if(self.foundMacAddress){
            stopAllTimers()
            return
        }
        if let macAddress = displayManager.getSmartMonitorMacAddr(for: displayID) {
            foundMacAddress = true
            stopAllTimers()
            handleMacAddressFound(mac: macAddress, displayID: displayID, displayName: displayName)
        } else {
            startRetryTimer(for: displayID, displayName: displayName)
        }
    }
    
    private func handleMacAddressFound(mac: String, displayID: CGDirectDisplayID, displayName: String) {
        logger.log("Successfully obtained MAC address: \(mac) for display: \(displayName)", level: .info)
        updateMacMapping(mac: mac, displayID: displayID)
        
        GoCallbackManager.shared.updateDisplayMapping(mac: mac, displayID: displayID)
        GoCallbackManager.shared.setCurrentDIASMac(mac)
        
        GoServiceBridge.shared.setDIASID(diasID: mac) { [weak self] success, error in
            if success {
                self?.logger.log("Successfully set DIAS ID in Go service", level: .info)
            } else {
                self?.logger.log("Failed to set DIAS ID: \(error ?? "Unknown"), but request may be queued", level: .warn)
            }
        }
//        stopTimer(for: displayID)
    }
    
    private func startRetryTimer(for displayID: CGDirectDisplayID, displayName: String) {
        logger.log("Starting retry timer for display: \(displayName)", level: .info)
        
        var retryCount = 0
        let maxRetries = 5
        
        let timer = Timer(timeInterval: 2.0, repeats: true) { [weak self] timer in
            guard let strongSelf = self else {
                timer.invalidate()
                return
            }
            
//            retryCount += 1
//            
//            if retryCount == 3 {
//                strongSelf.logger.log("Forcing display refresh after \(retryCount) retries", level: .info)
//                HelperDisplayManager.shared.forceCheckAllDisplaysForDIAS()
//            }
            
            if let mac = strongSelf.displayManager.getSmartMonitorMacAddr(for: displayID) {
                strongSelf.handleMacAddressFound(mac: mac, displayID: displayID, displayName: displayName)
                self?.foundMacAddress = true
                self?.stopAllTimers()
//                timer.invalidate()
            } else {
//                if retryCount >= maxRetries {
//                    strongSelf.logger.log("Max retries reached for display: \(displayName)", level: .warn)
//                    timer.invalidate()
//                }
            }
        }
        
        RunLoop.main.add(timer, forMode: .common)
        timers.append(timer)
    }
    
    private func stopTimer(for displayID: CGDirectDisplayID) {
        timers.removeAll { !$0.isValid }
    }
    
    func stopAllTimers() {
        timers.forEach { $0.invalidate() }
        timers.removeAll()
    }
    
    deinit {
        stopAllTimers()
    }
}
