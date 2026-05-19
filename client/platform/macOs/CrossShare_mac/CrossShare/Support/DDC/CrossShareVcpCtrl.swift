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
    var customerId: UInt16  // 10 bits
    var styleId: UInt8      // 6 bits
    var reserved: UInt16    // 16 bits
    
    init(bytes: [UInt8]) {
        let safeBytes = bytes.count >= 4 ? bytes : [0, 0, 0, 0]
        
        // Little-endian: byte[0] is LSB, byte[1] is MSB
        let firstTwoBytes = UInt16(safeBytes[0]) | (UInt16(safeBytes[1]) << 8)
        customerId = firstTwoBytes & 0x03FF           // bit 0-9:   mask 0000001111111111
        styleId = UInt8((firstTwoBytes >> 10) & 0x3F) // bit 10-15: mask 00111111
        
        // Little-endian for reserved field
        reserved = UInt16(safeBytes[2]) | (UInt16(safeBytes[3]) << 8)
    }
    
    var byte: [UInt8] {
        // Combine customerId (10 bits) and styleId (6 bits) into first 16 bits
        let firstTwoBytes = (UInt16(styleId) << 10) | customerId
        
        // Return bytes in little-endian order
        return [
            UInt8(firstTwoBytes & 0xFF),         // byte[0]: bit 0-7
            UInt8((firstTwoBytes >> 8) & 0xFF),  // byte[1]: bit 8-15
            UInt8(reserved & 0xFF),              // byte[2]: bit 16-23
            UInt8((reserved >> 8) & 0xFF)        // byte[3]: bit 24-31
        ]
    }
}

class CrossShareVcpCtrl {
    private let logger = CSLogger.shared
    func getSmartMonitorMacAddr(for displayID: CGDirectDisplayID) -> String? {
        var macAddr: [UInt8] = [0, 0, 0, 0, 0, 0]
        guard displayID != 0 else {
            logger.info("The input display ID is invalid.")
            return nil
        }
        
        let displays = DisplayManager.shared.getAllDisplays()
        if displays.isEmpty {
            logger.info("No displays found ")
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
            logger.info("can not find target display or it's not an OtherDisplay")
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
            logger.info("The first part of the MAC address cannot be obtained. Display ID: \(displayID) name:\(targetDisplay.name), Service might be invalid after reconnection")
            return nil
        }
        
        // Delay of 50ms (50,000 microseconds)
        usleep(50000)

        let _ = universalDDc.readDDCValues(command: DDCCI_CrossShare_CMD_GET_MAC_2)
        guard let result2 = universalDDc.readDDCValues(command: DDCCI_CrossShare_CMD_GET_MAC_2) else {
            logger.info("Unable to obtain MAC address (Part 2)")
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
        logger.info("The obtained MAC address: \(macString)")
        return macString
    }
    
    /// Get MAC address and Port from DDC/CI 0xE1/0xE2 commands
    /// 0xE1: First 4 bytes of MAC (max[2] + current[2])
    /// 0xE2: First 2 bytes = MAC bytes 5-6 (max[2]), Last 2 bytes = Port (current[2])
    func getMacAddressAndPort(for displayID: CGDirectDisplayID) -> (macAddress: String, source: UInt8, port: UInt8)? {
        var macAddr: [UInt8] = [0, 0, 0, 0, 0, 0]
        guard displayID != 0 else {
            logger.info("[getMacAddressAndPort] The input display ID is invalid.")
            return nil
        }
        
        let displays = DisplayManager.shared.getAllDisplays()
        if displays.isEmpty {
            logger.info("[getMacAddressAndPort] No displays found")
            return nil
        }
        
        var currentDisplay: Display?
        for display in displays {
            if display.identifier == displayID {
                currentDisplay = display
            }
        }
        guard let targetDisplay = currentDisplay as? OtherDisplay else {
            logger.info("[getMacAddressAndPort] Cannot find target display or it's not an OtherDisplay for ID: \(displayID)")
            return nil
        }
        
        let universalDDc = UniversalDDC()
        if targetDisplay.arm64ddc {
            universalDDc.initArm64Service(arm64avService: targetDisplay.arm64avService)
        } else {
            universalDDc.initIntelIdentifier(identifier: displayID)
        }
        
        // Warm-up read for 0xE1
        let _ = universalDDc.readDDCValues(command: DDCCI_CrossShare_CMD_GET_MAC_1)
        guard let result1 = universalDDc.readDDCValues(command: DDCCI_CrossShare_CMD_GET_MAC_1) else {
            logger.info("[getMacAddressAndPort] Cannot get MAC part 1 for display: \(displayID), name: \(targetDisplay.name)")
            return nil
        }
        
        // Delay 50ms
        usleep(50000)
        
        // Warm-up read for 0xE2
        let _ = universalDDc.readDDCValues(command: DDCCI_CrossShare_CMD_GET_MAC_2)
        guard let result2 = universalDDc.readDDCValues(command: DDCCI_CrossShare_CMD_GET_MAC_2) else {
            logger.info("[getMacAddressAndPort] Cannot get MAC part 2 for display: \(displayID)")
            return nil
        }
        
        // Parse MAC address
        // 0xE1: max(2 bytes) = macAddr[0:1], current(2 bytes) = macAddr[2:3]
        macAddr[0] = UInt8((result1.max & 0xff00) >> 8)
        macAddr[1] = UInt8(result1.max & 0xff)
        macAddr[2] = UInt8((result1.current & 0xff00) >> 8)
        macAddr[3] = UInt8(result1.current & 0xff)
        // 0xE2: max(2 bytes) = macAddr[4:5]
        macAddr[4] = UInt8((result2.max & 0xff00) >> 8)
        macAddr[5] = UInt8(result2.max & 0xff)
        
        let macString = macAddr.map { String(format: "%02X", $0) }.joined(separator: ":")
        
        // 0xE2: current(2 bytes) = Source(high byte) + Port(low byte)
        let source = UInt8((result2.current & 0xFF00) >> 8)
        let port = UInt8(result2.current & 0xFF)
        
        logger.info("[getMacAddressAndPort] MAC: \(macString), Source: \(source), Port: \(port)")
        
        return (macAddress: macString, source: source, port: port)
    }

    //    -- Verification successful
    func querySmartMonitorAuthStatus(for displayID: CGDirectDisplayID,index: UInt16,completed: @escaping (Bool) -> Void) {
        let processedIndex = (index & 0xFF) << 8

        guard displayID != 0 else {
            logger.info("The input display ID is invalid.")
            completed(false)
            return
        }
        
        let displays = DisplayManager.shared.getAllDisplays()
        if displays.isEmpty {
            logger.info("No displays found ")
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
            logger.info("can not find target display")
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
            logger.info("query AuthStatus sucess")
            
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
            logger.info("query AuthStatus fail")
        }
    }

    // -- Verification successful
    func getConnectedPortInfo(for displayID: CGDirectDisplayID) -> (source: UInt8, port: UInt8)? {
        guard displayID != 0 else {
            logger.info("The input display ID is invalid.")
            return (source: 0, port: 0)
        }
        
        let displays = DisplayManager.shared.getAllDisplays()
        if displays.isEmpty {
            logger.info("No displays found ")
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
            logger.info("can not find target display")
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
            logger.info("fail to get")
            return (source: 0, port: 0)
        }
        
        logger.info("getConnectedPortInfo result:\(result)")
        let source = UInt8((result.max & 0xFF00) >> 8)
        let port = UInt8(result.max & 0xFF)
        return (source: source, port:port)
    }
    
    //  --- Verification successful
    func updateMousePos(for displayID: CGDirectDisplayID,
                              width: UInt16,
                              height: UInt16,
                              posX: Int16,
                              posY: Int16){
        guard displayID != 0 else {
            logger.info("The input display ID is invalid.")
            return
        }
        
        let displays = DisplayManager.shared.getAllDisplays()
        if displays.isEmpty {
            logger.info("No displays found ")
            return
        }
        
        var currentDisplay: Display?
        for display in displays {
            // Debug display info
            logger.info("---Display --- : \(display.name)")
            if display.identifier == displayID {
                currentDisplay = display;
            }
        }
        if currentDisplay == nil {
            logger.info("can not find target display")
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
            logger.info("write width sucess")
        }else{
            logger.info("write width fail")
        }
        
        let writeHeightResult = universalDDc.writeDDCValues(command: DDCCI_CrossShare_CMD_SET_DESKTOP_RESOLUTION_H, value: height)
        if(writeHeightResult){
            logger.info("write height sucess")
        }else{
            logger.info("write height fail")
        }
        
        // TODO: posX should be signed integer
        let posX: UInt16 = (posX < 0) ? (width/2) : UInt16(posX)
        let writePosXResult = universalDDc.writeDDCValues(command: DDCCI_CrossShare_CMD_SET_CURSOR_POSITION_X, value: posX)
        if(writePosXResult){
            logger.info("write posX sucess")
        }else{
            logger.info("write posX fail")
        }

        // TODO: posY should be signed integer
        let posY: UInt16 = (posY < 0) ? (height/2) : UInt16(posY)
        let writePosYResult = universalDDc.writeDDCValues(command: DDCCI_CrossShare_CMD_SET_CURSOR_POSITION_Y, value: posY)
        if(writePosYResult){
            logger.info("write posY sucess")
        }else{
            logger.info("write posY fail")
        }
    }
    
    func getCustomerThemeCode(for displayID: CGDirectDisplayID) -> CrossShareThemeCode? {
        var themeBytes: [UInt8] = [0x00, 0x00, 0x00, 0x00]
        guard displayID != 0 else {
            logger.info("The input display ID is invalid.")
            return CrossShareThemeCode(bytes: themeBytes)
        }
        
        let displays = DisplayManager.shared.getAllDisplays()
        if displays.isEmpty {
            logger.info("No displays found ")
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
            logger.info("can not find target display")
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
            logger.info("getCustomerThemeCode fail to get")
            return nil
        }
        themeBytes = [
            UInt8((result.max & 0xFF00) >> 8),
            UInt8(result.max & 0xFF),
            UInt8((result.current & 0xFF00) >> 8),
            UInt8(result.current & 0xFF)
        ]
        
        let themeCode = CrossShareThemeCode(bytes: themeBytes)
        print("getCustomerThemeCode success:")
        print("Raw bytes: \(themeBytes.map { String(format: "0x%02X", $0) }.joined(separator: ", "))")
        print("customerId: \(themeCode.customerId)")
        print("styleId: \(themeCode.styleId)")
        print("reserved: \(themeCode.reserved)")
        
        return themeCode
    }
}

