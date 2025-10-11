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
    private let logger = XPCLogger.shared
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
        updateScreenCount()
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
    
    private func updateScreenCount() -> Int {
        let maxDisplays: UInt32 = 10
        var activeDisplays = [CGDirectDisplayID](repeating: 0, count: Int(maxDisplays))
        var displayCount: UInt32 = 0
        
        let result = CGGetActiveDisplayList(maxDisplays, &activeDisplays, &displayCount)
        guard result == .success else {
            logger.log("Failed to get active display list", level: .error)
            return 0
        }
        return Int(displayCount)
    }
    
    private func checkScreenChanges() {
        let currentScreenCount = updateScreenCount()
        
        if currentScreenCount != lastScreenCount {
            if currentScreenCount > lastScreenCount {
                screenCountChangeStatus = .increased
                logger.log("Screen count increased from \(lastScreenCount) to \(currentScreenCount)", level: .info)
                notifyScreenCountChange(change: "increased", currentCount: currentScreenCount, previousCount: lastScreenCount)
            } else {
                screenCountChangeStatus = .decreased
                logger.log("Screen count decreased from \(lastScreenCount) to \(currentScreenCount)", level: .info)
                handleScreenDecrease()
                notifyScreenCountChange(change: "decreased", currentCount: currentScreenCount, previousCount: lastScreenCount)
            }

            lastScreenCount = currentScreenCount

            self.handleScreenChange()
        }
    }
    
    private func handleScreenDecrease() {
        xpcService?.setExtractDIAS { [weak self] success, error in
            if success {
                self?.logger.log("SetExtractDIAS succeeded after screen decrease", level: .info)
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
