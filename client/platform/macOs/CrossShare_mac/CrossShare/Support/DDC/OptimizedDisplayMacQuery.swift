//
//  OptimizedDisplayMacQuery.swift
//  CrossShare
//
//  Created by zorobeyond on 2025/9/12.
//

import Foundation

class OptimizedDisplayMacQuery {
    private static let sharedInstance = OptimizedDisplayMacQuery()
    private var timers: [Timer] = []
    private var foundMacAddress = false
    let displayManager = CrossShareVcpCtrl()
    var lastDisplays:[Display] = []
    
    private init() {}
    
    class func shared() -> OptimizedDisplayMacQuery {
        return sharedInstance
    }
    
    func startQuerying(displays: [Display]) {
        // Reset the status
        foundMacAddress = false
        self.lastDisplays = displays
        stopAllTimers()
        print("----display count:\(displays.count)")
        
        // Process each monitor in sequence
        processDisplay(at: 0, in: self.lastDisplays)
    }
    
    // Process the display in ascending order by index
    private func processDisplay(at index: Int, in displays: [Display]) {
        // Boundary check: All displays have been processed
        guard index < displays.count,!foundMacAddress else {
            if (!self.foundMacAddress) {
                print("All the displays have failed to obtain the MAC address. The system will continue to attempt to retrieve it at regular intervals.")
            }
            return
        }
        
        let currentDisplay = displays[index]
        print("\n--- Handling the display screen: \(currentDisplay.name) (ID: \(currentDisplay.identifier)) ---")
        
        // 1. Immediately carry out a query first.
        let macAddress = displayManager.getSmartMonitorMacAddr(for: currentDisplay.identifier)
        
        if let mac = macAddress {
            // Successfully acquired, stop all timers and terminate the process
            foundMacAddress = true
            stopAllTimers()
            print("✅ Successfully obtained the MAC address: \(mac)")
            print("✅ DIAS monitor name:\(currentDisplay.name) ---")
            print("✅ DIAS monitor ID:\(currentDisplay.identifier) ---")
            return
        } else {
            // 2. The initial query failed. It will be retried every 2 seconds）
            print("❌ The initial query failed. It will be retried every 2 seconds")
            
            // Create a timer (without automatically joining the running loop)
            let timer = Timer(timeInterval: 2.0, repeats: true) { [weak self] timer in
                guard let strongSelf = self,!strongSelf.foundMacAddress else {
                    timer.invalidate()
                    return
                }
                
                // Scheduled retry query
                let retryMac = strongSelf.displayManager.getSmartMonitorMacAddr(for: currentDisplay.identifier)
                if let mac = retryMac {
                    strongSelf.foundMacAddress = true
                    strongSelf.stopAllTimers()
                    print("✅ Retry successful, MAC address: \(mac) (from: \(currentDisplay.name))")
                } else {
                    print("🔍 display \(currentDisplay.name) retry...")
                }
            }
            
            // Add the timer to the common modes of the main running loop (key optimization)
            RunLoop.main.add(timer, forMode: .common)
            
            timers.append(timer)
            
            // 3. Continue to process the next display (no need to wait for the current timer; execute immediately in sequence)）
            processDisplay(at: index + 1, in: self.lastDisplays)
        }
    }
    
    // Stop all timers
    private func stopAllTimers() {
        timers.forEach { $0.invalidate() }
        timers.removeAll()
    }
}
