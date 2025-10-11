//
//  ScreenMonitor.swift
//  CrossShare
//
//  Created by zorobeyond on 2025/9/12.
//

import Cocoa

class ScreenMonitor {
    private var lastScreenCount: Int = 0
    // Use enumeration to represent the state of screen quantity changes
    private enum ScreenCountChange {
        case unknown
        case increased
        case decreased
    }
    private var screenCountChangeStatus: ScreenCountChange = .unknown

    init() {
        // Record the current number of screens during initialization
        lastScreenCount = NSScreen.screens.count
        print("init number: \(lastScreenCount)")
        if(lastScreenCount > 0){
            gainMacAdress()
        }
        // Register for notifications of screen parameter changes
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(screenParametersChanged(_:)),
            name: NSApplication.didChangeScreenParametersNotification,
            object: nil
        )
    }
    
    @objc private func screenParametersChanged(_ notification: Notification) {
        // Obtain the latest screen count
        let currentScreenCount = NSScreen.screens.count
        
        // Compare whether the number of screens has changed.
        if currentScreenCount != lastScreenCount {
            if currentScreenCount > lastScreenCount {
                screenCountChangeStatus = .increased
            } else {
                screenCountChangeStatus = .decreased
            }
            let changeDescription: String
            switch screenCountChangeStatus {
            case.increased:
                changeDescription = "increased"
            case.decreased:
                changeDescription = "decreased"
            case.unknown:
                changeDescription = "unknown"
            }
            print("The number of screens has \(changeDescription). from \(lastScreenCount) to \(currentScreenCount) ")
            lastScreenCount = currentScreenCount
            // We need to delay for 1 or 2 seconds. Otherwise, the quantity returned by the getAllDisplays function won't have had a chance to change yet.
            DispatchQueue.main.asyncAfter(deadline: .now() + 2) {
                self.gainMacAdress()
            }
        }
    }
    
    func gainMacAdress(){
        let displays = DisplayManager.shared.getAllDisplays()
        if displays.isEmpty {
            print("No displays found ")
            DispatchQueue.main.asyncAfter(deadline: .now() + 0.5) {
                self.gainMacAdress()
            }
            return
        }
        print("displays count:\(displays.count)")
        let macQuery = OptimizedDisplayMacQuery.shared()
        macQuery.startQuerying(displays: displays)
    }

    
    deinit {
        // Remove the listener
        NotificationCenter.default.removeObserver(self)
        print("ScreenMonitor dealloc")
    }
}
