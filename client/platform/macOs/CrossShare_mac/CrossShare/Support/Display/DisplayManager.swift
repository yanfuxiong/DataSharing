//  Copyright © MonitorControl. @JoniVR, @theOneyouseek, @waydabber and others

import Cocoa
import CoreGraphics
import os.log

class DisplayManager {
    public static let shared = DisplayManager()
    
    var displays: [Display] = []
    var audioControlTargetDisplays: [OtherDisplay] = []
    let globalDDCQueue = DispatchQueue(label: "Global DDC queue")
    private var _gammaActivityEnforcer: NSWindow?
    var gammaInterferenceCounter = 0
    var gammaInterferenceWarningShown = false
    
    // Display monitoring
    private var displayReconfigurationCallback: CGDisplayReconfigurationCallBack?
    private var displayChangeTimer: Timer?
    
    var gammaActivityEnforcer: NSWindow {
        if let window = _gammaActivityEnforcer {
            return window
        }
        
        if Thread.isMainThread {
            let window = NSWindow(contentRect: .init(origin: NSPoint(x: 0, y: 0), size: .init(width: 1, height: 1)), styleMask: [], backing: .buffered, defer: false)
            _gammaActivityEnforcer = window
            return window
        } else {
            return DispatchQueue.main.sync {
                if let window = _gammaActivityEnforcer {
                    return window
                }
                let window = NSWindow(contentRect: .init(origin: NSPoint(x: 0, y: 0), size: .init(width: 1, height: 1)), styleMask: [], backing: .buffered, defer: false)
                _gammaActivityEnforcer = window
                return window
            }
        }
    }
    
    init() {
        setupDisplayMonitoring()
    }
    
    deinit {
        stopDisplayMonitoring()
    }
    
    func moveGammaActivityEnforcer(displayID: CGDirectDisplayID) {
        if let screen = DisplayManager.getByDisplayID(displayID: DisplayManager.resolveEffectiveDisplayID(displayID)) {
            self.gammaActivityEnforcer.setFrameOrigin(screen.frame.origin)
        }
        self.gammaActivityEnforcer.orderFrontRegardless()
    }
    
    var shades: [CGDirectDisplayID: NSWindow] = [:]
    var shadeGrave: [NSWindow] = []
    
    func isDisqualifiedFromShade(_ displayID: CGDirectDisplayID) -> Bool {
        if CGDisplayIsInHWMirrorSet(displayID) != 0 || CGDisplayIsInMirrorSet(displayID) != 0 {
            if displayID == DisplayManager.resolveEffectiveDisplayID(displayID), DisplayManager.isVirtual(displayID: displayID) || DisplayManager.isDummy(displayID: displayID) {
                var displayIDs = [CGDirectDisplayID](repeating: 0, count: 16)
                var displayCount: UInt32 = 0
                guard CGGetOnlineDisplayList(16, &displayIDs, &displayCount) == .success else {
                    return true
                }
                for displayId in displayIDs where CGDisplayMirrorsDisplay(displayId) == displayID && !DisplayManager.isVirtual(displayID: displayID) {
                    return true
                }
                return false
            }
            return true
        }
        return false
    }
    
    func createShadeOnDisplay(displayID: CGDirectDisplayID) -> NSWindow? {
        func createShadeWindow() -> NSWindow? {
            if let screen = DisplayManager.getByDisplayID(displayID: displayID) {
                let shade = NSWindow(contentRect: .init(origin: NSPoint(x: 0, y: 0), size: .init(width: 10, height: 1)), styleMask: [], backing: .buffered, defer: false)
                shade.title = "Monitor Control Window Shade for Display " + String(displayID)
                shade.isMovableByWindowBackground = false
                shade.backgroundColor = .clear
                shade.ignoresMouseEvents = true
                shade.level = NSWindow.Level(rawValue: Int(CGShieldingWindowLevel()))
                shade.orderFrontRegardless()
                shade.collectionBehavior = [.stationary, .canJoinAllSpaces, .ignoresCycle]
                shade.setFrame(screen.frame, display: true)
                shade.contentView?.wantsLayer = true
                shade.contentView?.alphaValue = 0.0
                shade.contentView?.layer?.backgroundColor = .black
                shade.contentView?.setNeedsDisplay(shade.frame)
                os_log("Window shade created for display %{public}@", type: .info, String(displayID))
                return shade
            }
            return nil
        }
        
        if Thread.isMainThread {
            return createShadeWindow()
        } else {
            return DispatchQueue.main.sync {
                return createShadeWindow()
            }
        }
    }
    
    func getShade(displayID: CGDirectDisplayID) -> NSWindow? {
        guard !self.isDisqualifiedFromShade(displayID) else {
            return nil
        }
        if let shade = shades[displayID] {
            return shade
        } else {
            if let shade = self.createShadeOnDisplay(displayID: displayID) {
                self.shades[displayID] = shade
                return shade
            }
        }
        return nil
    }
    
    func destroyAllShades() -> Bool {
        var ret = false
        for displayID in self.shades.keys {
            os_log("Attempting to destory shade for display  %{public}@", type: .info, String(displayID))
            if self.destroyShade(displayID: displayID) {
                ret = true
            }
        }
        if ret {
            os_log("Destroyed all shades.", type: .info)
        } else {
            os_log("No shades were found to be destroyed.", type: .info)
        }
        return ret
    }
    
    func destroyShade(displayID: CGDirectDisplayID) -> Bool {
        if let shade = shades[displayID] {
            os_log("Destroying shade for display %{public}@", type: .info, String(displayID))
            self.shadeGrave.append(shade)
            self.shades.removeValue(forKey: displayID)
            shade.close()
            return true
        }
        return false
    }
    
    func updateShade(displayID: CGDirectDisplayID) -> Bool {
        guard !self.isDisqualifiedFromShade(displayID) else {
            return false
        }
        if let screen = DisplayManager.getByDisplayID(displayID: displayID) {
            if let shade = getShade(displayID: displayID) {
                shade.setFrame(screen.frame, display: true)
                return true
            }
        }
        return false
    }
    
    func getShadeAlpha(displayID: CGDirectDisplayID) -> Float? {
        guard !self.isDisqualifiedFromShade(displayID) else {
            return 1
        }
        if let shade = getShade(displayID: displayID) {
            return Float(shade.contentView?.alphaValue ?? 1)
        } else {
            return 1
        }
    }
    
    func setShadeAlpha(value: Float, displayID: CGDirectDisplayID) -> Bool {
        guard !self.isDisqualifiedFromShade(displayID) else {
            return false
        }
        if let shade = getShade(displayID: displayID) {
            shade.contentView?.alphaValue = CGFloat(value)
            return true
        }
        return false
    }
    
    func configureDisplays() {
        self.clearDisplays()
        var onlineDisplayIDs = [CGDirectDisplayID](repeating: 0, count: 16)
        var displayCount: UInt32 = 0
        guard CGGetOnlineDisplayList(16, &onlineDisplayIDs, &displayCount) == .success else {
            os_log("Unable to get display list.", type: .info)
            return
        }
        for onlineDisplayID in onlineDisplayIDs where onlineDisplayID != 0 {
            let name = DisplayManager.getDisplayNameByID(displayID: onlineDisplayID)
            let id = onlineDisplayID
            let vendorNumber = CGDisplayVendorNumber(onlineDisplayID)
            let modelNumber = CGDisplayModelNumber(onlineDisplayID)
            let serialNumber = CGDisplaySerialNumber(onlineDisplayID)
            let isDummy: Bool = DisplayManager.isDummy(displayID: onlineDisplayID)
            let isVirtual: Bool = DisplayManager.isVirtual(displayID: onlineDisplayID)
            if DisplayManager.isAppleDisplay(displayID: onlineDisplayID) { // MARK: (point of interest for testing)
                let appleDisplay = AppleDisplay(id, name: name, vendorNumber: vendorNumber, modelNumber: modelNumber, serialNumber: serialNumber, isVirtual: isVirtual, isDummy: isDummy)
                os_log("Apple display found - %{public}@", type: .info, "ID: \(appleDisplay.identifier), Name: \(appleDisplay.name) (Vendor: \(appleDisplay.vendorNumber ?? 0), Model: \(appleDisplay.modelNumber ?? 0))")
                self.addDisplay(display: appleDisplay)
            } else {
                let otherDisplay = OtherDisplay(id, name: name, vendorNumber: vendorNumber, modelNumber: modelNumber, serialNumber: serialNumber, isVirtual: isVirtual, isDummy: isDummy)
                os_log("Other display found - %{public}@", type: .info, "ID: \(otherDisplay.identifier), Name: \(otherDisplay.name) (Vendor: \(otherDisplay.vendorNumber ?? 0), Model: \(otherDisplay.modelNumber ?? 0))")
                self.addDisplay(display: otherDisplay)
            }
        }
    }
    
    func setupOtherDisplays(firstrun: Bool = false) {
        
    }
    
    func restoreOtherDisplays() {
        
    }
    
    func normalizedName(_ name: String) -> String {
        var normalizedName = name.replacingOccurrences(of: "(", with: "")
        normalizedName = normalizedName.replacingOccurrences(of: ")", with: "")
        normalizedName = normalizedName.replacingOccurrences(of: " ", with: "")
        for i in 0 ... 9 {
            normalizedName = normalizedName.replacingOccurrences(of: String(i), with: "")
        }
        return normalizedName
    }
    
    func updateAudioControlTargetDisplays(deviceName: String) -> Int {
        self.audioControlTargetDisplays.removeAll()
        os_log("Detecting displays for audio control via audio device name matching...", type: .info)
        var numOfAddedDisplays = 0
        for ddcCapableDisplay in self.getDdcCapableDisplays() {
            var displayAudioDeviceName = DisplayManager.getDisplayRawNameByID(displayID: ddcCapableDisplay.identifier)
            if self.normalizedName(displayAudioDeviceName) == self.normalizedName(deviceName) {
                self.audioControlTargetDisplays.append(ddcCapableDisplay)
                numOfAddedDisplays += 1
                os_log("Added display for audio control - %{public}@", type: .info, ddcCapableDisplay.name)
            }
        }
        return numOfAddedDisplays
    }
    
    func getOtherDisplays() -> [OtherDisplay] {
        self.displays.compactMap { $0 as? OtherDisplay }
    }
    
    func getAllDisplays() -> [Display] {
        self.displays
    }
    
    func getDdcCapableDisplays() -> [OtherDisplay] {
        self.displays.compactMap { display -> OtherDisplay? in
            if let otherDisplay = display as? OtherDisplay {
                return otherDisplay
            } else { return nil }
        }
    }
    
    func getAppleDisplays() -> [AppleDisplay] {
        self.displays.compactMap { $0 as? AppleDisplay }
    }
    
    func getBuiltInDisplay() -> Display? {
        self.displays.first { CGDisplayIsBuiltin($0.identifier) != 0 }
    }
    
    func getCurrentDisplay(byFocus: Bool = false) -> Display? {
        if byFocus {
            guard let mainDisplayID = NSScreen.main?.displayID else {
                return nil
            }
            return self.displays.first { $0.identifier == mainDisplayID }
        } else {
            let mouseLocation = NSEvent.mouseLocation
            let screens = NSScreen.screens
            if let screenWithMouse = (screens.first { NSMouseInRect(mouseLocation, $0.frame, false) }) {
                return self.displays.first { $0.identifier == screenWithMouse.displayID }
            }
            return nil
        }
    }
    
    func addDisplay(display: Display) {
        self.displays.append(display)
    }
    
    func clearDisplays() {
        self.displays = []
    }
    
    func addDisplayCounterSuffixes() {
        var nameDisplays: [String: [Display]] = [:]
        for display in self.displays {
            if nameDisplays[display.name] != nil {
                nameDisplays[display.name]?.append(display)
            } else {
                nameDisplays[display.name] = [display]
            }
        }
        for nameDisplayKey in nameDisplays.keys where nameDisplays[nameDisplayKey]?.count ?? 0 > 1 {
            for i in 0 ... (nameDisplays[nameDisplayKey]?.count ?? 1) - 1 {
                if let display = nameDisplays[nameDisplayKey]?[i] {
                    display.name = "" + display.name + " (" + String(i + 1) + ")"
                }
            }
        }
    }
    
    func updateArm64AVServices() {
        if Arm64DDC.isArm64 {
            os_log("arm64 AVService update requested", type: .info)
            var displayIDs: [CGDirectDisplayID] = []
            for otherDisplay in self.getOtherDisplays() {
                displayIDs.append(otherDisplay.identifier)
            }
            for serviceMatch in Arm64DDC.getServiceMatches(displayIDs: displayIDs) {
                for otherDisplay in self.getOtherDisplays() where otherDisplay.identifier == serviceMatch.displayID && serviceMatch.service != nil {
                    otherDisplay.arm64avService = serviceMatch.service
                    os_log("Display service match successful for display %{public}@", type: .info, String(serviceMatch.displayID))
                    if serviceMatch.discouraged {
                        os_log("Display %{public}@ is flagged as discouraged by Arm64DDC.", type: .info, String(serviceMatch.displayID))
                        otherDisplay.isDiscouraged = true
                    } else if serviceMatch.dummy {
                        os_log("Display %{public}@ is flagged as dummy by Arm64DDC.", type: .info, String(serviceMatch.displayID))
                        otherDisplay.isDiscouraged = true
                        otherDisplay.isDummy = true
                    } else {
                        otherDisplay.arm64ddc = true // MARK: (point of interest when testing)
                    }
                }
            }
            os_log("AVService update done", type: .info)
        }
    }
    
    func resetSwBrightnessForAllDisplays(prefsOnly: Bool = false, noPrefSave: Bool = false, async: Bool = false) {
        
    }
    
    func restoreSwBrightnessForAllDisplays(async: Bool = false) {
        
    }
    
    func getAffectedDisplays(isBrightness: Bool = false, isVolume: Bool = false) -> [Display]? {
        return []
    }
    
    static func isDummy(displayID: CGDirectDisplayID) -> Bool {
        let vendorNumber = CGDisplayVendorNumber(displayID)
        let rawName = DisplayManager.getDisplayRawNameByID(displayID: displayID)
        if rawName.lowercased().contains("dummy") || (self.isVirtual(displayID: displayID) && vendorNumber == UInt32(0xF0F0)) {
            os_log("NOTE: Display is a dummy!", type: .info)
            return true
        }
        return false
    }
    
    static func isVirtual(displayID: CGDirectDisplayID) -> Bool {
        var isVirtual = false
        if #available(macOS 11.0, *) {
            if let dictionary = (CoreDisplay_DisplayCreateInfoDictionary(displayID)?.takeRetainedValue() as NSDictionary?) {
                let isVirtualDevice = dictionary["kCGDisplayIsVirtualDevice"] as? Bool
                let displayIsAirplay = dictionary["kCGDisplayIsAirPlay"] as? Bool
                if isVirtualDevice ?? displayIsAirplay ?? false {
                    isVirtual = true
                }
            }
        }
        return isVirtual
    }
    
    static func engageMirror() -> Bool {
        var onlineDisplayIDs = [CGDirectDisplayID](repeating: 0, count: 16)
        var displayCount: UInt32 = 0
        guard CGGetOnlineDisplayList(16, &onlineDisplayIDs, &displayCount) == .success, displayCount > 1 else {
            return false
        }
        // Break display mirror if there is any
        var mirrorBreak = false
        var displayConfigRef: CGDisplayConfigRef?
        for onlineDisplayID in onlineDisplayIDs where onlineDisplayID != 0 {
            if CGDisplayIsInHWMirrorSet(onlineDisplayID) != 0 || CGDisplayIsInMirrorSet(onlineDisplayID) != 0 {
                if mirrorBreak == false {
                    CGBeginDisplayConfiguration(&displayConfigRef)
                }
                CGConfigureDisplayMirrorOfDisplay(displayConfigRef, onlineDisplayID, kCGNullDirectDisplay)
                mirrorBreak = true
            }
        }
        if mirrorBreak {
            CGCompleteDisplayConfiguration(displayConfigRef, CGConfigureOption.permanently)
            return true
        }
        // Build display mirror
        var mainDisplayId = kCGNullDirectDisplay
        for onlineDisplayID in onlineDisplayIDs where onlineDisplayID != 0 {
            if CGDisplayIsBuiltin(onlineDisplayID) == 0, mainDisplayId == kCGNullDirectDisplay {
                mainDisplayId = onlineDisplayID
            }
        }
        guard mainDisplayId != kCGNullDirectDisplay else {
            return false
        }
        CGBeginDisplayConfiguration(&displayConfigRef)
        for onlineDisplayID in onlineDisplayIDs where onlineDisplayID != 0 && onlineDisplayID != mainDisplayId {
            CGConfigureDisplayMirrorOfDisplay(displayConfigRef, onlineDisplayID, mainDisplayId)
        }
        CGCompleteDisplayConfiguration(displayConfigRef, CGConfigureOption.permanently)
        return true
    }
    
    static func resolveEffectiveDisplayID(_ displayID: CGDirectDisplayID) -> CGDirectDisplayID {
        var realDisplayID = displayID
        if CGDisplayIsInHWMirrorSet(displayID) != 0 || CGDisplayIsInMirrorSet(displayID) != 0 {
            let mirroredDisplayID = CGDisplayMirrorsDisplay(displayID)
            if mirroredDisplayID != 0 {
                realDisplayID = mirroredDisplayID
            }
        }
        return realDisplayID
    }
    
    static func isAppleDisplay(displayID: CGDirectDisplayID) -> Bool {
        if #available(macOS 15.0, *) {
            if CGDisplayVendorNumber(displayID) != 1552, CGSIsHDRSupported(displayID), CGSIsHDREnabled(displayID) {
                return CGDisplayIsBuiltin(displayID) != 0
            }
        }
        var brightness: Float = -1
        let ret = DisplayServicesGetBrightness(displayID, &brightness)
        if ret == 0, brightness >= 0 { // If brightness read appears to be successful using DisplayServices then it should be an Apple display
            return true
        }
        return CGDisplayIsBuiltin(displayID) != 0 // If built-in display, it should be Apple
    }
    
    static func getByDisplayID(displayID: CGDirectDisplayID) -> NSScreen? {
        NSScreen.screens.first { $0.displayID == displayID }
    }
    
    static func getDisplayRawNameByID(displayID: CGDirectDisplayID) -> String {
        let defaultName = ""
        if #available(macOS 11.0, *) {
            if let dictionary = (CoreDisplay_DisplayCreateInfoDictionary(displayID)?.takeRetainedValue() as NSDictionary?), let nameList = dictionary["DisplayProductName"] as? [String: String], let name = nameList["en_US"] ?? nameList.first?.value {
                return name
            }
        }
        if let screen = getByDisplayID(displayID: displayID) {
            return screen.displayName ?? defaultName
        }
        return defaultName
    }
    
    static func getDisplayNameByID(displayID: CGDirectDisplayID) -> String {
        let defaultName: String = NSLocalizedString("Unknown", comment: "Unknown display name")
        if #available(macOS 11.0, *) {
            if let dictionary = (CoreDisplay_DisplayCreateInfoDictionary(displayID)?.takeRetainedValue() as NSDictionary?), let nameList = dictionary["DisplayProductName"] as? [String: String], var name = nameList[Locale.current.identifier] ?? nameList["en_US"] ?? nameList.first?.value {
                if CGDisplayIsInHWMirrorSet(displayID) != 0 || CGDisplayIsInMirrorSet(displayID) != 0 {
                    let mirroredDisplayID = CGDisplayMirrorsDisplay(displayID)
                    if mirroredDisplayID != 0, let dictionary = (CoreDisplay_DisplayCreateInfoDictionary(mirroredDisplayID)?.takeRetainedValue() as NSDictionary?), let nameList = dictionary["DisplayProductName"] as? [String: String], let mirroredName = nameList[Locale.current.identifier] ?? nameList["en_US"] ?? nameList.first?.value {
                        name.append(" | " + mirroredName)
                    }
                }
                return name
            }
        }
        if let screen = getByDisplayID(displayID: displayID) { // MARK: This, and NSScreen+Extension.swift will not be needed when we drop MacOS 10 support.
            if #available(macOS 10.15, *) {
                return screen.localizedName
            } else {
                return screen.displayName ?? defaultName
            }
        }
        return defaultName
    }
    
    // MARK: - Display Connection Monitoring
    
    private func setupDisplayMonitoring() {
        // Register for display reconfiguration callbacks
        let callback: CGDisplayReconfigurationCallBack = { (display, flags, userInfo) in
            let manager = Unmanaged<DisplayManager>.fromOpaque(userInfo!).takeUnretainedValue()
            manager.handleDisplayChange(display: display, flags: flags)
        }
        
        let userInfo = Unmanaged.passUnretained(self).toOpaque()
        CGDisplayRegisterReconfigurationCallback(callback, userInfo)
        
        os_log("Display monitoring started", type: .info)
    }
    
    private func stopDisplayMonitoring() {
        if let callback = displayReconfigurationCallback {
            let userInfo = Unmanaged.passUnretained(self).toOpaque()
            CGDisplayRemoveReconfigurationCallback(callback, userInfo)
        }
        
        displayChangeTimer?.invalidate()
        displayChangeTimer = nil
        
        os_log("Display monitoring stopped", type: .info)
    }
    
    private func handleDisplayChange(display: CGDirectDisplayID, flags: CGDisplayChangeSummaryFlags) {
        // Log the change
        if flags.contains(.addFlag) {
            os_log("Display connected: %{public}@", type: .info, String(display))
        } else if flags.contains(.removeFlag) {
            os_log("Display disconnected: %{public}@", type: .info, String(display))
        } else if flags.contains(.movedFlag) {
            os_log("Display moved: %{public}@", type: .info, String(display))
        } else if flags.contains(.setMainFlag) {
            os_log("Display became main: %{public}@", type: .info, String(display))
        } else if flags.contains(.setModeFlag) {
            os_log("Display mode changed: %{public}@", type: .info, String(display))
        } else if flags.contains(.enabledFlag) {
            os_log("Display enabled: %{public}@", type: .info, String(display))
        } else if flags.contains(.disabledFlag) {
            os_log("Display disabled: %{public}@", type: .info, String(display))
        } else if flags.contains(.mirrorFlag) {
            os_log("Display mirroring changed: %{public}@", type: .info, String(display))
        } else if flags.contains(.unMirrorFlag) {
            os_log("Display unmirrored: %{public}@", type: .info, String(display))
        }
        
        // Cancel existing timer
        displayChangeTimer?.invalidate()
        
        // Set a short timer to batch multiple changes
        displayChangeTimer = Timer.scheduledTimer(withTimeInterval: 0.5, repeats: false) { _ in
            self.refreshDisplayConfiguration()
        }
    }
    
    private func refreshDisplayConfiguration() {
        os_log("Refreshing display configuration...", type: .info)
        
        // Store previous display IDs for comparison
        let previousDisplayIDs = Set(displays.map { $0.identifier })
        
        // Reconfigure displays
        configureDisplays()
        addDisplayCounterSuffixes()
        updateArm64AVServices()
        setupOtherDisplays(firstrun: false)
        
        // Get current display IDs
        let currentDisplayIDs = Set(displays.map { $0.identifier })
        
        // Find added and removed displays
        let addedDisplays = currentDisplayIDs.subtracting(previousDisplayIDs)
        let removedDisplays = previousDisplayIDs.subtracting(currentDisplayIDs)
        
        // Post notifications for added displays
        for displayID in addedDisplays {
            os_log("Posting notification for added display: %{public}@", type: .info, String(displayID))
            NotificationCenter.default.post(
                name: .displayConnectedNotification,
                object: nil,
                userInfo: ["displayID": displayID]
            )
        }
        
        // Post notifications for removed displays
        for displayID in removedDisplays {
            os_log("Posting notification for removed display: %{public}@", type: .info, String(displayID))
            NotificationCenter.default.post(
                name: .displayDisconnectedNotification,
                object: nil,
                userInfo: ["displayID": displayID]
            )
        }
        
        // Post general configuration changed notification
        NotificationCenter.default.post(
            name: .displayConfigurationChangedNotification,
            object: nil,
            userInfo: [
                "addedDisplays": Array(addedDisplays),
                "removedDisplays": Array(removedDisplays),
                "allDisplays": Array(currentDisplayIDs)
            ]
        )
        
        os_log("Display configuration refresh complete. Displays: %d", type: .info, displays.count)
    }
    
    // Public method to manually trigger refresh
    func forceRefreshDisplays() {
        os_log("Force refresh requested", type: .info)
        refreshDisplayConfiguration()
    }
}
