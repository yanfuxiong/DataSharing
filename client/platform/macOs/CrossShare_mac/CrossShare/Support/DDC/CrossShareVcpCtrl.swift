//
//  OptimizedDisplayMacQuery.swift
//  CrossShare
//
//  Created by zorobeyond on 2025/9/11.
//
import Foundation
import IOKit
import IOKit.graphics


// Define DDC/CI command constants
let DDCCI_CrossShare_CMD_GET_MAC_1: UInt8 = 0xE1 // read only
let DDCCI_CrossShare_CMD_GET_MAC_2: UInt8 = 0xE2 // read only
let DDCCI_CrossShare_CMD_AUTH_DEVICE:UInt8 = 0xE0 // do set then get
let DDCCI_CrossShare_CMD_GET_TV_SRC:UInt8 = 0xE3 // read only
let DDCCI_CrossShare_CMD_SET_DESKTOP_RESOLUTION_W:UInt8 = 0xE4 // write only
let DDCCI_CrossShare_CMD_SET_DESKTOP_RESOLUTION_H:UInt8 = 0xE5 // write only
let DDCCI_CrossShare_CMD_SET_CURSOR_POSITION_X:UInt8 = 0xE6 // write only
let DDCCI_CrossShare_CMD_SET_CURSOR_POSITION_Y:UInt8 = 0xE7 // write only
let DDCCI_CrossShare_CMD_GET_CUSTOMIZED_THEME:UInt8 = 0xE8 // read only

//let DDCCI_CrossShare_TimeInterval: Float = 0.05

struct CrossShareThemeCode {
    var customerId: UInt16  // 10
    var styleId: UInt8      // 6
    var reserved: UInt16    // 16
    
    init(bytes: [UInt8]) {
        let safeBytes = bytes.count >= 4 ? bytes : [0, 0, 0, 0]
        
        let firstTwoBytes = UInt16(safeBytes[0]) << 8 | UInt16(safeBytes[1])
        customerId = firstTwoBytes & 0x03FF
        styleId = UInt8((firstTwoBytes >> 10) & 0x3F)
        reserved = UInt16(safeBytes[2]) << 8 | UInt16(safeBytes[3])
    }
    
    var byte: [UInt8] {
        let firstTwoBytes = (UInt16(styleId) << 10) | customerId
        return [
            UInt8((firstTwoBytes >> 8) & 0xFF),
            UInt8(firstTwoBytes & 0xFF),
            UInt8((reserved >> 8) & 0xFF),
            UInt8(reserved & 0xFF)
        ]
    }
}

class CrossShareVcpCtrl {
    func getSmartMonitorMacAddr(for displayID: CGDirectDisplayID) -> String? {
        var macAddr: [UInt8] = [0, 0, 0, 0, 0, 0]
        guard displayID != 0 else {
            print("The input display ID is invalid.")
            return nil
        }
        
        let displays = DisplayManager.shared.getAllDisplays()
        if displays.isEmpty {
            print("No displays found ")
            return nil
        }
        
        var currentDisplay: Display?
        for display in displays {
            // Debug display info
            print("--display name: \(display.name) ID:\(display.identifier) --" )
            if display.identifier == displayID {
                currentDisplay = display;
                print("---currentDisplay --- : \(display.name)")}
        }
        guard let targetDisplay = currentDisplay as? OtherDisplay else {
            print("can not find target display or it's not an OtherDisplay")
            return nil
        }
        let universalDDc = UniversalDDC()
        if targetDisplay.arm64ddc{
            universalDDc.initArm64Service(arm64avService: targetDisplay.arm64avService)
        }else{
            universalDDc.initIntelIdentifier(identifier: displayID)
        }
        let _ = universalDDc.readDDCValues(command: DDCCI_CrossShare_CMD_GET_MAC_1)
        guard let result1 = universalDDc.readDDCValues(command: DDCCI_CrossShare_CMD_GET_MAC_1) else {
            print("The first part of the MAC address cannot be obtained. Display ID: \(displayID) name:\(targetDisplay.name), Service might be invalid after reconnection")
            return nil
        }
        
        // Delay of 50ms (50,000 microseconds)
        usleep(50000)

        let _ = universalDDc.readDDCValues(command: DDCCI_CrossShare_CMD_GET_MAC_2)
        guard let result2 = universalDDc.readDDCValues(command: DDCCI_CrossShare_CMD_GET_MAC_2) else {
            print("Unable to obtain MAC address (Part 2)")
            return nil
        }
        
        // Parse MAC address
        macAddr[0] = UInt8((result1.max & 0xff00) >> 8)
        macAddr[1] = UInt8(result1.max & 0xff)
        macAddr[2] = UInt8((result1.current & 0xff00) >> 8)
        macAddr[3] = UInt8(result1.current & 0xff)
        macAddr[4] = UInt8((result2.max & 0xff00) >> 8)
        macAddr[5] = UInt8(result2.max & 0xff)
        
        // Format the MAC address into the "AA:BB:CC:DD:EE:FF" format
        let macString = macAddr.map { String(format: "%02X", $0) }.joined(separator: ":")
        print("The obtained MAC address: \(macString)")
        return macString
    }

    //    -- Verification successful
    func querySmartMonitorAuthStatus(for displayID: CGDirectDisplayID,index: UInt16,completed: @escaping (Bool) -> Void) {
        let processedIndex = (index & 0xFF) << 8

        guard displayID != 0 else {
            print("The input display ID is invalid.")
            completed(false)
            return
        }
        
        let displays = DisplayManager.shared.getAllDisplays()
        if displays.isEmpty {
            print("No displays found ")
            completed(false)
            return
        }
        
        var currentDisplay: Display?
        for display in displays {
            // Debug display info
            print("---Display --- : \(display.name)")
            if display.identifier == displayID {
                currentDisplay = display;
            }
        }
        if currentDisplay == nil {
            print("can not find target display")
            completed(false)
            return
        }
        let targetDisplay = currentDisplay as! OtherDisplay
        let universalDDc = UniversalDDC()
        if targetDisplay.arm64ddc {
            universalDDc.initArm64Service(arm64avService: targetDisplay.arm64avService)
        }else{
            universalDDc.initIntelIdentifier(identifier: displayID)
        }
        let result = universalDDc.writeDDCValues(command: DDCCI_CrossShare_CMD_AUTH_DEVICE, value: UInt16(processedIndex))
        if(result){
            print("query AuthStatus sucess")
            
            // Delay of 50ms (50,000 microseconds)
            usleep(50000)

            let readResult = universalDDc.readDDCValues(command: DDCCI_CrossShare_CMD_AUTH_DEVICE)
            // Determine the reading result (any non-nil value is regarded as "value read", and additional value verification can be added)）
            let isReadSuccess = readResult != nil

            let result1 = universalDDc.readDDCValues(command: DDCCI_CrossShare_CMD_AUTH_DEVICE)
            if isReadSuccess {
                completed(true)
            }else{
                completed(false)
            }
        }else{
            completed(false)
            print("query AuthStatus fail")
        }
    }

    // -- Verification successful
    func getConnectedPortInfo(for displayID: CGDirectDisplayID) -> (source: UInt8, port: UInt8)? {
        guard displayID != 0 else {
            print("The input display ID is invalid.")
            return (source: 0, port: 0)
        }
        
        let displays = DisplayManager.shared.getAllDisplays()
        if displays.isEmpty {
            print("No displays found ")
            return (source: 0, port: 0)
        }
        
        var currentDisplay: Display?
        for display in displays {
            // Debug display info
            print("---Display --- : \(display.name)")
            if display.identifier == displayID {
                currentDisplay = display;
                print("---currentDisplay --- : \(display.name)")
            }
        }
        if currentDisplay == nil {
            print("can not find target display")
            return (source: 0, port: 0)
        }
        let targetDisplay = currentDisplay as! OtherDisplay
        let universalDDc = UniversalDDC()
        if targetDisplay.arm64ddc{
            universalDDc.initArm64Service(arm64avService: targetDisplay.arm64avService)
        }else{
            universalDDc.initIntelIdentifier(identifier: displayID)
        }
        guard let result = universalDDc.readDDCValues(command: DDCCI_CrossShare_CMD_GET_TV_SRC) else {
            print("fail to get")
            return (source: 0, port: 0)
        }
        
        print("getConnectedPortInfo result:\(result)")
        let source = UInt8((result.max & 0xFF00) >> 8)
        let port = UInt8(result.max & 0xFF)
        return (source: source, port:port)
    }
    
    //  --- Verification successful
    func updateMousePos(for displayID: CGDirectDisplayID,
                              width: UInt16,
                              height: UInt16,
                              posX: UInt16,
                              posY: UInt16){
        guard displayID != 0 else {
            print("The input display ID is invalid.")
            return
        }
        
        let displays = DisplayManager.shared.getAllDisplays()
        if displays.isEmpty {
            print("No displays found ")
            return
        }
        
        var currentDisplay: Display?
        for display in displays {
            // Debug display info
            print("---Display --- : \(display.name)")
            if display.identifier == displayID {
                currentDisplay = display;
            }
        }
        if currentDisplay == nil {
            print("can not find target display")
            return
        }
        let targetDisplay = currentDisplay as! OtherDisplay
        let universalDDc = UniversalDDC()
        if targetDisplay.arm64ddc{
            universalDDc.initArm64Service(arm64avService: targetDisplay.arm64avService)
        }else{
            universalDDc.initIntelIdentifier(identifier: displayID)
        }
        let writeWidthResult = universalDDc.writeDDCValues(command: DDCCI_CrossShare_CMD_SET_DESKTOP_RESOLUTION_W, value: width)
        if(writeWidthResult){
            print("write width sucess")
        }else{
            print("write width fail")
        }
        
        // 延迟 50ms (50,000 微秒)
        usleep(50000)
        let writeHeightResult = universalDDc.writeDDCValues(command: DDCCI_CrossShare_CMD_SET_DESKTOP_RESOLUTION_H, value: height)
        if(writeHeightResult){
            print("write height sucess")
        }else{
            print("write height fail")
        }
        
        // 延迟 50ms (50,000 微秒)
        usleep(50000)
        let writePosXResult = universalDDc.writeDDCValues(command: DDCCI_CrossShare_CMD_SET_CURSOR_POSITION_X, value: height)
        if(writePosXResult){
            print("write posX sucess")
        }else{
            print("write posX fail")
        }

        
        // 延迟 50ms (50,000 微秒)
        usleep(50000)
        let writePosYResult = universalDDc.writeDDCValues(command: DDCCI_CrossShare_CMD_SET_CURSOR_POSITION_Y, value: height)
        if(writePosYResult){
            print("write posY sucess")
        }else{
            print("write posY fail")
        }
    }
    
    static func getCustomerThemeCode(for displayID: CGDirectDisplayID) -> CrossShareThemeCode? {
        var themeBytes: [UInt8] = [0x00, 0x00, 0x00, 0x00]
        guard displayID != 0 else {
            print("The input display ID is invalid.")
            return CrossShareThemeCode(bytes: themeBytes)
        }
        
        let displays = DisplayManager.shared.getAllDisplays()
        if displays.isEmpty {
            print("No displays found ")
            return CrossShareThemeCode(bytes: themeBytes)
        }
        
        var currentDisplay: Display?
        for display in displays {
            // Debug display info
            print("---Display --- : \(display.name)")
            if display.identifier == displayID {
                currentDisplay = display;
            }
        }
        if currentDisplay == nil {
            print("can not find target display")
            return CrossShareThemeCode(bytes: themeBytes)
        }
        let targetDisplay = currentDisplay as! OtherDisplay
        let universalDDc = UniversalDDC()
        if targetDisplay.arm64ddc {
            universalDDc.initArm64Service(arm64avService: targetDisplay.arm64avService)
        }else{
            universalDDc.initIntelIdentifier(identifier: displayID)
        }

        guard let result = universalDDc.readDDCValues(command:DDCCI_CrossShare_CMD_GET_CUSTOMIZED_THEME) else {
            print("getCustomerThemeCode fail to get")
            return nil
        }
        themeBytes = [
            UInt8((result.current & 0xFF00) >> 8),
            UInt8(result.current & 0xFF),
            UInt8((result.max & 0xFF00) >> 8),
            UInt8(result.max & 0xFF)
        ]
        return CrossShareThemeCode(bytes: themeBytes)
    }
}

