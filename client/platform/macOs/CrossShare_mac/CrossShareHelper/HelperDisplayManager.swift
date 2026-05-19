//
//  HelperDisplayManager.swift
//  CrossShareHelper
//
//  Manages display detection and DDCCI communication in Helper process
//

import Foundation
import CoreGraphics

// MARK: - DIAS Display Info

/// Information about a connected DIAS display
struct DIASDisplayInfo: CustomStringConvertible {
    let displayID: CGDirectDisplayID
    var macAddress: String          // Format: "XX:XX:XX:XX:XX:XX"
    var macAddressNoColon: String   // Format: "XXXXXXXXXXXX" (for JSON)
    var monitorId: String           // Display UUID
    var source: UInt8               // Source type from DDC/CI 0xE3
    var port: UInt8                 // Port from DDC/CI 0xE2
    var udpMousePort: UInt16        // Allocated UDP port for mouse
    var udpKeyboardPort: UInt16     // Allocated UDP port for keyboard
    var isServerRunning: Bool       // Whether UDP servers are active
    
    var description: String {
        return "DIASDisplayInfo(DisplayID=\(displayID), MacAddr=\(macAddressNoColon), MonitorId=\(monitorId), Source=\(source), Port=\(port), UdpMousePort=\(udpMousePort), UdpKeyboardPort=\(udpKeyboardPort), ServerRunning=\(isServerRunning))"
    }
    
    init(displayID: CGDirectDisplayID, macAddress: String, monitorId: String, source: UInt8, port: UInt8, udpMousePort: UInt16, udpKeyboardPort: UInt16) {
        self.displayID = displayID
        self.macAddress = macAddress
        self.macAddressNoColon = macAddress.replacingOccurrences(of: ":", with: "")
        self.monitorId = monitorId
        self.source = source
        self.port = port
        self.udpMousePort = udpMousePort
        self.udpKeyboardPort = udpKeyboardPort
        self.isServerRunning = false
    }
}

class HelperDisplayManager {
    static let shared = HelperDisplayManager()
    private let vcpCtrl = CrossShareVcpCtrl()
    private var gcdTimer: DispatchSourceTimer?
    private var pollingTimer: DispatchSourceTimer?
    
    // Thread safety: Use lock to protect display state
    private let displayStateLock = NSLock()
    
    // Cache for display state to avoid unnecessary DIAS checks
    private var lastDisplayIDs: Set<CGDirectDisplayID> = []
    private var checkedDIASDisplays: Set<CGDirectDisplayID> = []
    private var _activeDisplays = [CGDirectDisplayID](repeating: 0, count: 10)
    private var _displayCount: UInt32 = 0
    
    // Thread-safe accessors
    var activeDisplays: [CGDirectDisplayID] {
        get {
            displayStateLock.lock()
            defer { displayStateLock.unlock() }
            return _activeDisplays
        }
        set {
            displayStateLock.lock()
            defer { displayStateLock.unlock() }
            _activeDisplays = newValue
        }
    }
    
    var displayCount: UInt32 {
        get {
            displayStateLock.lock()
            defer { displayStateLock.unlock() }
            return _displayCount
        }
        set {
            displayStateLock.lock()
            defer { displayStateLock.unlock() }
            _displayCount = newValue
        }
    }
    
    // Store DIAS display ID when successfully set (backward compatibility)
    var currentDIASDisplayID: CGDirectDisplayID? = nil
    
    // Polling interval for DIAS software hotplug detection (in seconds)
    private let pollingInterval: TimeInterval = 2.0
    private var pollingCheckCount: Int = 0

    // MARK: - Multi-Display DIAS Tracking
    
    /// All connected DIAS displays, keyed by displayID
    private(set) var diasDisplays: [CGDirectDisplayID: DIASDisplayInfo] = [:]
    private let diasDisplaysLock = NSLock()
    
    /// Next port to try when allocating UDP ports
    private var nextPortStart: UInt16 = 10000
    
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
        stopPolling()
    }
    
    func startMonitoring() {
        logger.log("Starting display monitoring in Helper", level: .info)
        _ = updateScreenCount()
        ensureDisplayManagerConfigured {
            self.checkDisplays()
        }
        startPolling()
    }
    
    func stopMonitoring() {
        logger.log("Stopping display monitoring in Helper", level: .info)
        gcdTimer?.cancel()
        gcdTimer = nil
        stopPolling()
        // Stop all per-display servers
        CSWifiRemoteControlServer.shared.stopAllDisplayServers()
    }
    
    // MARK: - Polling mechanism for DIAS software hotplug detection
    
    /// Start active polling to detect DIAS software-triggered display changes
    /// This complements the system notification mechanism which may not fire for software hotplug events
    private func startPolling() {
        stopPolling()
        
        logger.log("Starting polling for DIAS software hotplug detection (interval: \(pollingInterval)s)", level: .info)
        
        let queue = DispatchQueue.global(qos: .utility)
        let timer = DispatchSource.makeTimerSource(queue: queue)
        timer.schedule(deadline: .now() + pollingInterval, repeating: pollingInterval)
        
        timer.setEventHandler { [weak self] in
            guard let self = self else { return }
            
            self.pollingCheckCount += 1
            let checkCount = self.pollingCheckCount
            
            logger.log("[Polling #\(checkCount)] Checking for display changes (DIAS software hotplug)", level: .debug)
            
            let previousCount = self.displayCount
            let currentCount = self.updateScreenCount()
            
            if currentCount != previousCount {
                logger.log("[Polling #\(checkCount)] Display count changed: \(previousCount) -> \(currentCount)", level: .info)
            }
            
            self.checkDisplays()
        }
        
        timer.resume()
        pollingTimer = timer
    }
    
    /// Stop the polling timer
    private func stopPolling() {
        if pollingTimer != nil {
            logger.log("Stopping polling for DIAS software hotplug detection (total checks: \(pollingCheckCount))", level: .info)
            pollingTimer?.cancel()
            pollingTimer = nil
            pollingCheckCount = 0
        }
    }
    
    func checkDisplaysNow() {
        _ = updateScreenCount()
        checkDisplays()
    }
    
    func resetDIASCache() {
        logger.log("Resetting DIAS cache", level: .info)
        checkedDIASDisplays.removeAll()
        lastDisplayIDs.removeAll()
    }
    
    func setCurrentDIASDisplayID(_ displayID: CGDirectDisplayID) {
        currentDIASDisplayID = displayID
        logger.log("Set current DIAS display ID: \(displayID)", level: .info)
    }
    
    func clearCurrentDIASDisplayID() {
        if let id = currentDIASDisplayID {
            logger.log("Clearing current DIAS display ID: \(id)", level: .info)
            currentDIASDisplayID = nil
        }
    }
        
//    func forceCheckAllDisplaysForDIAS() {
//        logger.log("Forcing DIAS check for all displays", level: .info)
//        resetDIASCache()
//        checkDisplays()
//    }
    
   func updateScreenCount() -> Int {
        let maxDisplays: UInt32 = 10
        var localActiveDisplays = [CGDirectDisplayID](repeating: 0, count: Int(maxDisplays))
        var localDisplayCount: UInt32 = 0
        
        let result = CGGetOnlineDisplayList(maxDisplays, &localActiveDisplays, &localDisplayCount)
        
        guard result == .success else {
            logger.log("Failed to get active display list from CGGetOnlineDisplayList", level: .error)
            return 0
        }
        
        // Log display IDs for debugging
        let displayIDsStr = localActiveDisplays.prefix(Int(localDisplayCount)).map { String($0) }.joined(separator: ", ")
        logger.log("CGGetOnlineDisplayList returned \(localDisplayCount) displays: [\(displayIDsStr)]", level: .debug)
        
        // Update shared state atomically
        displayStateLock.lock()
        self._activeDisplays = localActiveDisplays
        self._displayCount = localDisplayCount
        displayStateLock.unlock()

        return Int(localDisplayCount)
    }

    
    private func checkDisplays() {
        let activeDisplays = self.activeDisplays
        let displayCount = self.displayCount
        
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
            
            if !removedDisplays.isEmpty {
                logger.log("Displays removed: \(Array(removedDisplays))", level: .info)
                checkedDIASDisplays.subtract(removedDisplays)
                HelperOptimizedDisplayMacQuery.shared().cleanupDisplays(removedDisplays)
                handleRemovedDisplays(removedDisplays)
            }
            
            if !newDisplays.isEmpty {
                logger.log("New displays detected: \(Array(newDisplays))", level: .info)
                
                // Ensure DisplayManager operations are on main thread
                DispatchQueue.main.async {
                    self.refreshDisplayConfiguration()
                }
                
                for displayID in newDisplays {
                    checkDisplayForDIAS(displayID: displayID)
                }
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
            guard let _ = self else { return }
            
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

    func updateMousePos(for displayID: CGDirectDisplayID, width: UInt16, height: UInt16, posX: Int16, posY: Int16) {
        logger.log("Update mouse position for display \(displayID) with resolution: \(width)x\(height), (x,y): (\(posX), \(posY))", level: .info)
        vcpCtrl.updateMousePos(for: displayID, width: width, height: height, posX: posX, posY: posY)
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
    
    // MARK: - Display UUID
    
    /// Get a unique Display identifier string for a given display ID
    /// Constructs UUID from vendor/model/serial via CoreGraphics, with IODisplayUUID fallback from CoreDisplay
    func getDisplayUUID(for displayID: CGDirectDisplayID) -> String {
        // Try to get IODisplayUUID from CoreDisplay info dictionary
        if let displayInfoRef = CoreDisplay_DisplayCreateInfoDictionary(displayID) {
            let displayInfo = displayInfoRef.takeRetainedValue() as NSDictionary
            if let uuidValue = displayInfo["IODisplayUUID"] as? String, !uuidValue.isEmpty {
                logger.log("Display UUID (IODisplayUUID) for \(displayID): \(uuidValue)", level: .info)
                return uuidValue
            }
        }
        
        // Construct a unique identifier from vendor/model/serial
        let vendor = CGDisplayVendorNumber(displayID)
        let model = CGDisplayModelNumber(displayID)
        let serial = CGDisplaySerialNumber(displayID)
        
        let constructed = String(format: "%08X-%08X-%08X-%08X", vendor, model, serial, displayID)
        logger.log("Display UUID (constructed) for \(displayID): vendor=\(vendor), model=\(model), serial=\(serial) -> \(constructed)", level: .info)
        return constructed
    }
    
    // MARK: - Multi-Display DIAS Event Management
    
    /// Handle a DIAS display plug-in event
    /// Called when a display is detected as DIAS (MAC address obtained successfully)
    /// - Parameters:
    ///   - displayID: The CGDirectDisplayID of the display
    ///   - macAddress: MAC address obtained from DDC/CI 0xE1/0xE2
    ///   - source: Source value from DDC/CI 0xE2 high byte
    ///   - port: Port value from DDC/CI 0xE2 low byte
    func handleDIASDisplayPlugIn(displayID: CGDirectDisplayID, macAddress: String, source: UInt8, port: UInt8) {
        logger.log("[DisplayEvent] DIAS display plug-in detected - DisplayID: \(displayID), MAC: \(macAddress), Source: \(source), Port: \(port)", level: .info)
        
        // Get Display UUID
        let monitorId = getDisplayUUID(for: displayID)
        
        // Allocate UDP ports for this display
        guard let portPair = NetworkUtils.shared.findAvailableUDPPortPair(startPort: nextPortStart) else {
            logger.log("[DisplayEvent] Failed to allocate UDP port pair for display \(displayID)", level: .error)
            return
        }
        
        // Update next port start for future allocations
        nextPortStart = portPair.keyboardPort + 1
        if nextPortStart > 60000 { nextPortStart = 10000 }
        
        logger.log("[DisplayEvent] Allocated UDP ports - Mouse: \(portPair.mousePort), Keyboard: \(portPair.keyboardPort)", level: .info)
        
        // Create DIAS display info
        var displayInfo = DIASDisplayInfo(
            displayID: displayID,
            macAddress: macAddress,
            monitorId: monitorId,
            source: source,
            port: port,
            udpMousePort: portPair.mousePort,
            udpKeyboardPort: portPair.keyboardPort
        )
        
        // Start UDP servers for this display
        CSWifiRemoteControlServer.shared.startForDisplay(displayID, mousePort: portPair.mousePort, keyboardPort: portPair.keyboardPort)
        displayInfo.isServerRunning = true
        
        // Track this display
        diasDisplaysLock.lock()
        diasDisplays[displayID] = displayInfo
        diasDisplaysLock.unlock()
        
        // Set backward-compatible single DIAS ID (first one found)
        if currentDIASDisplayID == nil {
            currentDIASDisplayID = displayID
        }
        
        logger.log("[DisplayEvent] Plug-In: \(displayInfo)", level: .info)
        
        // Send plug-in event to Go
        sendDisplayEventInfo(displayInfo: displayInfo, plugEvent: 1)
    }
    
    /// Handle a DIAS display plug-out event
    func handleDIASDisplayPlugOut(displayID: CGDirectDisplayID) {
        diasDisplaysLock.lock()
        guard let displayInfo = diasDisplays[displayID] else {
            diasDisplaysLock.unlock()
            logger.log("[DisplayEvent] No DIAS info found for removed display \(displayID)", level: .debug)
            return
        }
        diasDisplays.removeValue(forKey: displayID)
        diasDisplaysLock.unlock()
        
        logger.log("[DisplayEvent] Plug-Out: \(displayInfo)", level: .info)
        
        // Stop UDP servers for this display
        CSWifiRemoteControlServer.shared.stopForDisplay(displayID)
        
        // Send plug-out event to Go
        sendDisplayEventInfo(displayInfo: displayInfo, plugEvent: 0)
        
        // Update backward-compatible single DIAS tracking
        if currentDIASDisplayID == displayID {
            diasDisplaysLock.lock()
            currentDIASDisplayID = diasDisplays.keys.first
            diasDisplaysLock.unlock()
            
            if currentDIASDisplayID == nil {
                logger.log("[DisplayEvent] All DIAS displays removed", level: .info)
            }
        }
    }
    
    /// Send display event info JSON to Go via SetDisplayEventInfo
    private func sendDisplayEventInfo(displayInfo: DIASDisplayInfo, plugEvent: Int) {
        let json: [String: Any] = [
            "MacAddr": displayInfo.macAddressNoColon,
            "MonitorId": displayInfo.monitorId,
            "PlugEvent": plugEvent,
            "Source": Int(displayInfo.source),
            "Port": Int(displayInfo.port),
            "UdpMousePort": Int(displayInfo.udpMousePort),
            "UdpKeyboardPort": Int(displayInfo.udpKeyboardPort)
        ]
        
        guard let jsonData = try? JSONSerialization.data(withJSONObject: json, options: []),
              let jsonString = String(data: jsonData, encoding: .utf8) else {
            logger.log("[DisplayEvent] Failed to serialize display event JSON", level: .error)
            return
        }
        
        logger.log("[DisplayEvent] Sending SetDisplayEventInfo: \(jsonString)", level: .info)
        
        GoServiceBridge.shared.setDisplayEventInfo(jsonString: jsonString) {success, error in
            if success {
                logger.log("[DisplayEvent] SetDisplayEventInfo succeeded for display \(displayInfo.displayID)", level: .info)
            } else {
               logger.log("[DisplayEvent] SetDisplayEventInfo failed: \(error ?? "Unknown")", level: .error)
            }
        }
    }
    
    /// Check if a display ID is a tracked DIAS display
    func isDIASDisplay(_ displayID: CGDirectDisplayID) -> Bool {
        diasDisplaysLock.lock()
        defer { diasDisplaysLock.unlock() }
        return diasDisplays[displayID] != nil
    }
    
    /// Get all tracked DIAS display IDs
    func getAllDIASDisplayIDs() -> [CGDirectDisplayID] {
        diasDisplaysLock.lock()
        defer { diasDisplaysLock.unlock() }
        return Array(diasDisplays.keys)
    }
    
    /// Handle removal of displays - check if any DIAS displays were removed
    func handleRemovedDisplays(_ removedDisplayIDs: Set<CGDirectDisplayID>) {
        for displayID in removedDisplayIDs {
            if isDIASDisplay(displayID) {
                handleDIASDisplayPlugOut(displayID: displayID)
            }
        }
    }
    
    @objc private func displayConfigurationChanged(_ notification: Notification) {
        logger.log("[Notification] Display configuration changed notification received", level: .info)
        
        if let userInfo = notification.userInfo {
            if let addedDisplays = userInfo["addedDisplays"] as? [CGDirectDisplayID] {
                logger.log("[Notification] Added displays: \(addedDisplays)", level: .info)
                for displayID in addedDisplays {
                    checkedDIASDisplays.remove(displayID)
                }
            }
            if let removedDisplays = userInfo["removedDisplays"] as? [CGDirectDisplayID] {
                logger.log("[Notification] Removed displays: \(removedDisplays)", level: .info)
                checkedDIASDisplays.subtract(removedDisplays)
            }
        }
        
        // Reduced delay from 1.0s to 0.5s for faster response
        // Still keep some delay to batch multiple rapid changes
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.3) { [weak self] in
            guard let self = self else { return }
            logger.log("[Notification] Executing delayed checkDisplays() call", level: .info)
            _ = self.updateScreenCount()
            self.checkDisplays()
        }
    }
}
