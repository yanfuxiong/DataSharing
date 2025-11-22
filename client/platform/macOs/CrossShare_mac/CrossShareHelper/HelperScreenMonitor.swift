//
//  HelperScreenMonitor.swift
//  CrossShareHelper
//
//  Monitor screen changes in Helper process
//

import Foundation
import CoreGraphics

class HelperScreenMonitor {
    private var lastScreenCount: Int = 0
    private let logger = CSLogger.shared
    private var gcdTimer: DispatchSourceTimer?
    private weak var xpcService: HelperXPCService?

    private enum ScreenCountChange {
        case unknown
        case increased
        case decreased
    }
    private var screenCountChangeStatus: ScreenCountChange = .unknown

    init(xpcService: HelperXPCService) {
        self.xpcService = xpcService
        startMonitoring()
    }
    
    private func startMonitoring() {
        let queue = DispatchQueue(label: "com.crossShare.screen.queue", qos: .background)
        gcdTimer = DispatchSource.makeTimerSource(queue: queue)
        gcdTimer?.schedule(deadline: .now(), repeating: 5.0)
        gcdTimer?.setEventHandler { [weak self] in
            self?.checkScreenChanges()
        }
        gcdTimer?.resume()
    }
    
    
    private func checkScreenChanges() {
        let currentScreenCount =  HelperDisplayManager.shared.updateScreenCount()
        logger.log("currentScreenCount:\(currentScreenCount) lastScreenCount:\(lastScreenCount)", level: .info)
        if currentScreenCount != lastScreenCount {
            if currentScreenCount > lastScreenCount {
                screenCountChangeStatus = .increased
                logger.log("Screen count increased from \(lastScreenCount) to \(currentScreenCount)", level: .info)
                notifyScreenCountChange(change: "increased", currentCount: currentScreenCount, previousCount: lastScreenCount)
            } else {
                logger.log("Screen count decreased from \(lastScreenCount) to \(currentScreenCount)", level: .info)
                
                // Check if the removed display is the DIAS display
                let isDIASRemoved = checkIfDIASDisplayRemoved()
                if isDIASRemoved {
                    logger.log("DIAS display was removed, executing handleScreenDecrease", level: .info)
                    screenCountChangeStatus = .decreased
                    handleScreenDecrease()
                    notifyScreenCountChange(change: "decreased", currentCount: currentScreenCount, previousCount: lastScreenCount)
                } else {
                    logger.log("Non-DIAS display was removed, skipping handleScreenDecrease", level: .info)
                }
                
            }

            lastScreenCount = currentScreenCount

            self.handleScreenChange()
        }
    }
    
    private func checkIfDIASDisplayRemoved() -> Bool {
        guard let diasDisplayID = HelperDisplayManager.shared.currentDIASDisplayID else {
            logger.log("No DIAS display ID stored, treating as non-DIAS removal", level: .debug)
            return false
        }
        
        // Get current active displays
        let activeDisplays = HelperDisplayManager.shared.activeDisplays
        let displayCount = HelperDisplayManager.shared.displayCount
        
        // Check if DIAS display is still in the active list
        for i in 0..<Int(displayCount) {
            let displayID = activeDisplays[i]
            if displayID == diasDisplayID {
                logger.log("DIAS display \(diasDisplayID) is still active", level: .debug)
                return false
            }
        }
        
        // DIAS display was removed
        logger.log("DIAS display \(diasDisplayID) was removed from active displays", level: .info)
        return true
    }
    
    private func handleScreenDecrease() {
        xpcService?.setExtractDIAS { [weak self] success, error in
            if success {
                self?.logger.log("SetExtractDIAS succeeded after screen decrease", level: .info)
                // Clear the stored DIAS display ID after successful execution
                HelperDisplayManager.shared.clearCurrentDIASDisplayID()
            } else {
                self?.logger.log("SetExtractDIAS failed after screen decrease: \(error ?? "Unknown error")", level: .error)
            }
        }
    }

    private func notifyScreenCountChange(change: String, currentCount: Int, previousCount: Int) {
        logger.log("Notifying GUI about screen count change: \(change) (\(previousCount) -> \(currentCount))", level: .info)
        xpcService?.notifyDelegates { delegate in
            delegate.didDetectScreenCountChange?(change: change, currentCount: currentCount, previousCount: previousCount)
        }
    }
    
    private func handleScreenChange() {
        HelperDisplayManager.shared.checkDisplaysNow()
    }
    
    deinit {
        logger.log("HelperScreenMonitor deinitialized", level: .info)
    }
}
