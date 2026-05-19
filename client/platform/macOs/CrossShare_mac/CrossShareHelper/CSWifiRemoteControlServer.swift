//
//  CSWifiRemoteControlServer.swift
//  CrossShareHelper
//
//  Created by TS on 2026/1/26.
//
//  WiFi Remote Control Server (using BSD sockets)
//  - Port 8080: Mouse control
//  - Port 8081: Keyboard control
//

import Foundation
import CoreGraphics
import AppKit
import Carbon.HIToolbox
import ApplicationServices

// MARK: - Data Packet Structures

/// Mouse control packet (16 bytes)
struct MouseDataPacket {
    var x: Float           // 4 bytes - normalized coordinate 0.0-1.0
    var y: Float           // 4 bytes - normalized coordinate 0.0-1.0
    var mouseKey: Int32    // 4 bytes - key type
    var mouseValue: Int32  // 4 bytes - key value (0=release, 1=press, scroll amount for wheel)
    
    static let size = 16
}

/// Keyboard control packet (8 bytes)
struct KeyboardDataPacket {
    var keyboardKey: Int32    // 4 bytes - keycode
    var keyboardValue: Int32  // 4 bytes - key value (0=release, 1=press)
    
    static let size = 8
}

// MARK: - Mouse Key Constants

enum MouseKey: Int32 {
    case polling = -128       // Polling packet
    case pollingEnd = -127    // Polling end
    case wheel = 8            // Scroll wheel
    case leftButton = 272     // Left button
    case rightButton = 273    // Right button
    case middleButton = 274   // Middle button
}

// MARK: - Per-Display Server Info

/// Stores UDP server state for a specific display
struct DisplayUDPServerInfo {
    let displayID: CGDirectDisplayID
    var mouseSocketFD: Int32 = -1
    var keyboardSocketFD: Int32 = -1
    var mouseReceiveThread: Thread?
    var keyboardReceiveThread: Thread?
    var isMouseRunning = false
    var isKeyboardRunning = false
    var mousePort: UInt16 = 0
    var keyboardPort: UInt16 = 0
}

// MARK: - CSWifiRemoteControlServer

class CSWifiRemoteControlServer {
    
    // MARK: - Singleton
    static let shared = CSWifiRemoteControlServer()
    
    // MARK: - Properties
    
    private let logger = CSLogger.shared
    
    // Per-display server tracking
    private var displayServers: [CGDirectDisplayID: DisplayUDPServerInfo] = [:]
    private let displayServersLock = NSLock()
    
    // Mouse button states
    private var prevLeftButton: Int32 = 0
    private var prevRightButton: Int32 = 0
    private var prevMiddleButton: Int32 = 0
    
    // CapsLock toggle state (macOS CapsLock is a toggle, not hold)
    private var capsLockOn = false
    
    // Arrow keys and navigation keys require numericPad + secondaryFn flags on macOS
    private static let navigationKeyCodes: Set<CGKeyCode> = [
        CGKeyCode(kVK_LeftArrow), CGKeyCode(kVK_RightArrow),
        CGKeyCode(kVK_UpArrow), CGKeyCode(kVK_DownArrow),
        CGKeyCode(kVK_Home), CGKeyCode(kVK_End),
        CGKeyCode(kVK_PageUp), CGKeyCode(kVK_PageDown),
        CGKeyCode(kVK_ForwardDelete), CGKeyCode(kVK_Help),
    ]
    private static let navigationKeyFlags = CGEventFlags(rawValue: 0xA00000)
    
    // Click count tracking for double/triple click
    private var leftClickCount: Int64 = 0
    private var rightClickCount: Int64 = 0
    private var lastLeftClickTime: Date?
    private var lastRightClickTime: Date?
    private var lastLeftClickPoint: CGPoint = .zero
    private var lastRightClickPoint: CGPoint = .zero
    private let multiClickInterval: TimeInterval = 0.3  // 300ms for multi-click
    private let multiClickRadius: CGFloat = 3.0  // pixels tolerance for multi-click
    
    // Mouse movement range (screen)
    private var mouseRange: CGRect = .zero
    
    // Statistics
    private var mousePacketCount: Int = 0
    private var keyboardPacketCount: Int = 0
    private var lastStatTime: Date = Date()
    
    // MARK: - Initialization
    
    private init() {
        updateMouseRange()
        
        // Monitor screen changes
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(screenConfigurationChanged),
            name: NSApplication.didChangeScreenParametersNotification,
            object: nil
        )
        
        logger.log("CSWifiRemoteControlServer initialized", level: .info)
    }
    
    deinit {
        NotificationCenter.default.removeObserver(self)
        stopAllDisplayServers()
    }
    
    // MARK: - Screen Configuration
    
    @objc private func screenConfigurationChanged() {
        updateMouseRange()
        logger.log("Screen configuration changed, updating mouse range", level: .info)
    }
    
    private func updateMouseRange() {
        if let screen = NSScreen.main {
            mouseRange = screen.frame
            logger.log("Mouse range updated: x=\(mouseRange.origin.x), y=\(mouseRange.origin.y), w=\(mouseRange.width), h=\(mouseRange.height)", level: .info)
        }
    }
    
    // MARK: - Mouse Data Processing
    
    private func processMouseData(_ data: Data, displayID: CGDirectDisplayID) {
        // Parse packet (little-endian)
        let x = data.withUnsafeBytes { $0.load(fromByteOffset: 0, as: Float.self) }
        let y = data.withUnsafeBytes { $0.load(fromByteOffset: 4, as: Float.self) }
        let mouseKey = data.withUnsafeBytes { $0.load(fromByteOffset: 8, as: Int32.self) }
        let mouseValue = data.withUnsafeBytes { $0.load(fromByteOffset: 12, as: Int32.self) }
        
        // Update statistics
        mousePacketCount += 1
        updateStats()
        
        // Log mouse data (skip polling packets to reduce log volume)
        if mouseKey != MouseKey.polling.rawValue {
            let keyName = getMouseKeyName(mouseKey)
            logger.log("Mouse data: x=\(String(format: "%.4f", x)), y=\(String(format: "%.4f", y)), key=\(keyName)(\(mouseKey)), value=\(mouseValue), display=\(displayID)", level: .info)
        }
        
        // Ignore polling packets (too frequent)
        if mouseKey == MouseKey.polling.rawValue {
            return
        }
        
        // Handle polling end
        if mouseKey == MouseKey.pollingEnd.rawValue {
            logger.log("Polling end signal received", level: .info)
            resetMouseButtonStates()
            return
        }
        
        // Clamp coordinates
        let clampedX = max(0.0, min(1.0, x))
        let clampedY = max(0.0, min(1.0, y))
        
        // Use display-specific bounds (CGDisplayBounds returns CG coordinates with top-left origin)
        let displayBounds = CGDisplayBounds(displayID)
        let targetRange: CGRect
        if displayBounds.width > 0 && displayBounds.height > 0 {
            targetRange = displayBounds
        } else {
            targetRange = mouseRange
            logger.log("CGDisplayBounds empty for display \(displayID), using fallback", level: .warn)
        }
        
        let absX = targetRange.origin.x + CGFloat(clampedX) * targetRange.width
        let absY = targetRange.origin.y + CGFloat(clampedY) * targetRange.height
        
        // Move mouse
        moveMouse(to: CGPoint(x: absX, y: absY))
        
        // Process button events
        processMouseKey(mouseKey, value: mouseValue, at: CGPoint(x: absX, y: absY))
    }
    
    private func moveMouse(to point: CGPoint) {
        let eventType: CGEventType
        let button: CGMouseButton
        
        if prevLeftButton != 0 {
            eventType = .leftMouseDragged
            button = .left
        } else if prevRightButton != 0 {
            eventType = .rightMouseDragged
            button = .right
        } else if prevMiddleButton != 0 {
            eventType = .otherMouseDragged
            button = .center
        } else {
            eventType = .mouseMoved
            button = .left
        }
        
        let moveEvent = CGEvent(mouseEventSource: nil,
                               mouseType: eventType,
                               mouseCursorPosition: point,
                               mouseButton: button)
        moveEvent?.post(tap: .cghidEventTap)
    }
    
    private func processMouseKey(_ key: Int32, value: Int32, at point: CGPoint) {
        switch key {
        case MouseKey.wheel.rawValue:
            // Scroll wheel
            wheelMouse(delta: value)
            
        case MouseKey.leftButton.rawValue:
            // Left button
            if prevLeftButton != value {
                if value == 1 {
                    // Update click count for multi-click detection
                    let clickCount = updateClickCount(for: .left, at: point)
                    sendMouseEvent(.leftMouseDown, at: point, button: .left, clickCount: clickCount)
                } else {
                    sendMouseEvent(.leftMouseUp, at: point, button: .left, clickCount: leftClickCount)
                }
                prevLeftButton = value
            }
            
        case MouseKey.rightButton.rawValue:
            // Right button
            if prevRightButton != value {
                if value == 1 {
                    let clickCount = updateClickCount(for: .right, at: point)
                    sendMouseEvent(.rightMouseDown, at: point, button: .right, clickCount: clickCount)
                } else {
                    sendMouseEvent(.rightMouseUp, at: point, button: .right, clickCount: rightClickCount)
                }
                prevRightButton = value
            }
            
        case MouseKey.middleButton.rawValue:
            // Middle button
            if prevMiddleButton != value {
                if value == 1 {
                    sendMouseEvent(.otherMouseDown, at: point, button: .center, clickCount: 1)
                } else {
                    sendMouseEvent(.otherMouseUp, at: point, button: .center, clickCount: 1)
                }
                prevMiddleButton = value
            }
            
        default:
            // Other cases, release all buttons
            releaseAllMouseButtons(at: point)
        }
    }
    
    private func sendMouseEvent(_ type: CGEventType, at point: CGPoint, button: CGMouseButton, clickCount: Int64 = 1) {
        let event = CGEvent(mouseEventSource: nil,
                           mouseType: type,
                           mouseCursorPosition: point,
                           mouseButton: button)
        // Set click count for double/triple click support
        event?.setIntegerValueField(.mouseEventClickState, value: clickCount)
        event?.post(tap: .cghidEventTap)
    }
    
    /// Update click count for multi-click detection (double/triple click)
    private func updateClickCount(for button: CGMouseButton, at point: CGPoint) -> Int64 {
        let now = Date()
        
        switch button {
        case .left:
            // Check if this is a continuation of previous clicks (within time and distance)
            if let lastTime = lastLeftClickTime,
               now.timeIntervalSince(lastTime) <= multiClickInterval,
               hypot(point.x - lastLeftClickPoint.x, point.y - lastLeftClickPoint.y) <= multiClickRadius {
                leftClickCount += 1
            } else {
                leftClickCount = 1
            }
            lastLeftClickTime = now
            lastLeftClickPoint = point
            return leftClickCount
            
        case .right:
            if let lastTime = lastRightClickTime,
               now.timeIntervalSince(lastTime) <= multiClickInterval,
               hypot(point.x - lastRightClickPoint.x, point.y - lastRightClickPoint.y) <= multiClickRadius {
                rightClickCount += 1
            } else {
                rightClickCount = 1
            }
            lastRightClickTime = now
            lastRightClickPoint = point
            return rightClickCount
            
        default:
            return 1
        }
    }
    
    private func wheelMouse(delta: Int32) {
        let scrollEvent = CGEvent(scrollWheelEvent2Source: nil,
                                 units: .line,
                                 wheelCount: 1,
                                 wheel1: delta,
                                 wheel2: 0,
                                 wheel3: 0)
        scrollEvent?.post(tap: .cghidEventTap)
    }
    
    private func releaseAllMouseButtons(at point: CGPoint) {
        if prevLeftButton != 0 {
            sendMouseEvent(.leftMouseUp, at: point, button: .left)
            prevLeftButton = 0
        }
        if prevRightButton != 0 {
            sendMouseEvent(.rightMouseUp, at: point, button: .right)
            prevRightButton = 0
        }
        if prevMiddleButton != 0 {
            sendMouseEvent(.otherMouseUp, at: point, button: .center)
            prevMiddleButton = 0
        }
    }
    
    private func resetMouseButtonStates() {
        let currentPos = NSEvent.mouseLocation
        if let screen = NSScreen.main {
            let cgPos = CGPoint(x: currentPos.x, y: screen.frame.height - currentPos.y)
            releaseAllMouseButtons(at: cgPos)
        }
    }
    
    // MARK: - Keyboard Data Processing
    
    private func processKeyboardData(_ data: Data) {
        // Log raw bytes for debugging
//        let hexString = data.map { String(format: "%02X", $0) }.joined(separator: " ")
//        logger.log("Keyboard raw data (\(data.count) bytes): \(hexString)", level: .debug)
        
        // Parse packet (little-endian)
        let keyboardKey = data.withUnsafeBytes { $0.load(fromByteOffset: 0, as: Int32.self) }
        let keyboardValue = data.withUnsafeBytes { $0.load(fromByteOffset: 4, as: Int32.self) }
        
        // Update statistics
        keyboardPacketCount += 1
        updateStats()
        
        // Log keyboard data
        // DIAS protocol: keyboardValue 0=press(down), 1=release(up) — matches Windows KEYEVENTF_KEYUP convention
        let isKeyDown = keyboardValue == 0
//        let keyName = getVKCodeName(keyboardKey)
        let action = isKeyDown ? "pressed" : "released"
//        logger.log("Keyboard data: key=\(keyName)(0x\(String(keyboardKey, radix: 16, uppercase: true))), value=\(keyboardValue), action=\(action)", level: .info)
        
        // Convert Windows VK code to macOS keycode
        guard let macKeyCode = windowsVKCodeToMacOS(keyboardKey) else {
            logger.log("Unknown Windows VK code: \(keyboardKey) (0x\(String(keyboardKey, radix: 16, uppercase: true))), cannot convert to macOS keycode", level: .warn)
            return
        }
        
        if macKeyCode == CGKeyCode(kVK_CapsLock) {
            // CapsLock: toggle on press, send flagsChanged for both press and release
            if isKeyDown {
                capsLockOn.toggle()
            }
            let capsFlags: CGEventFlags = capsLockOn ? .maskAlphaShift : []
            if let event = CGEvent(keyboardEventSource: nil, virtualKey: macKeyCode, keyDown: isKeyDown) {
                event.type = .flagsChanged
                event.flags = capsFlags
                event.post(tap: .cghidEventTap)
                logger.log("CapsLock flagsChanged sent: capsLockOn=\(capsLockOn), \(action)", level: .info)
            }
        } else if let keyEvent = CGEvent(keyboardEventSource: nil, virtualKey: macKeyCode, keyDown: isKeyDown) {
            if CSWifiRemoteControlServer.navigationKeyCodes.contains(macKeyCode) {
                keyEvent.flags = CSWifiRemoteControlServer.navigationKeyFlags
            }
            keyEvent.post(tap: .cghidEventTap)
            logger.log("Keyboard event sent: macOS keyCode=\(macKeyCode), \(action)", level: .info)
        }
    }
    
    /// Get mouse key name
    private func getMouseKeyName(_ key: Int32) -> String {
        switch key {
        case MouseKey.polling.rawValue:
            return "Polling"
        case MouseKey.pollingEnd.rawValue:
            return "PollingEnd"
        case MouseKey.wheel.rawValue:
            return "Wheel"
        case MouseKey.leftButton.rawValue:
            return "LeftButton"
        case MouseKey.rightButton.rawValue:
            return "RightButton"
        case MouseKey.middleButton.rawValue:
            return "MiddleButton"
        default:
            return "Unknown"
        }
    }
    
    /// Get Windows VK code name
    private func getVKCodeName(_ key: Int32) -> String {
        let keyNames: [Int32: String] = [
            // Letters (0x41-0x5A)
            0x41: "A", 0x42: "B", 0x43: "C", 0x44: "D", 0x45: "E", 0x46: "F", 0x47: "G",
            0x48: "H", 0x49: "I", 0x4A: "J", 0x4B: "K", 0x4C: "L", 0x4D: "M", 0x4E: "N",
            0x4F: "O", 0x50: "P", 0x51: "Q", 0x52: "R", 0x53: "S", 0x54: "T", 0x55: "U",
            0x56: "V", 0x57: "W", 0x58: "X", 0x59: "Y", 0x5A: "Z",
            // Numbers (0x30-0x39)
            0x30: "0", 0x31: "1", 0x32: "2", 0x33: "3", 0x34: "4",
            0x35: "5", 0x36: "6", 0x37: "7", 0x38: "8", 0x39: "9",
            // Function keys (0x70-0x7B)
            0x70: "F1", 0x71: "F2", 0x72: "F3", 0x73: "F4", 0x74: "F5", 0x75: "F6",
            0x76: "F7", 0x77: "F8", 0x78: "F9", 0x79: "F10", 0x7A: "F11", 0x7B: "F12",
            // Control keys
            0x08: "Backspace", 0x09: "Tab", 0x0D: "Enter", 0x10: "Shift", 0x11: "Ctrl",
            0x12: "Alt", 0x14: "CapsLock", 0x1B: "Escape", 0x20: "Space",
            // Navigation
            0x21: "PageUp", 0x22: "PageDown", 0x23: "End", 0x24: "Home",
            0x25: "Left", 0x26: "Up", 0x27: "Right", 0x28: "Down",
            0x2D: "Insert", 0x2E: "Delete",
            // Numpad
            0x60: "Num0", 0x61: "Num1", 0x62: "Num2", 0x63: "Num3", 0x64: "Num4",
            0x65: "Num5", 0x66: "Num6", 0x67: "Num7", 0x68: "Num8", 0x69: "Num9",
            0x6A: "Num*", 0x6B: "Num+", 0x6D: "Num-", 0x6E: "Num.", 0x6F: "Num/",
            // OEM
            0xBA: ";", 0xBB: "=", 0xBC: ",", 0xBD: "-", 0xBE: ".", 0xBF: "/",
            0xC0: "`", 0xDB: "[", 0xDC: "\\", 0xDD: "]", 0xDE: "'",
            // Windows
            0x5B: "LWin", 0x5C: "RWin",
        ]
        return keyNames[key] ?? "VK_0x\(String(key, radix: 16, uppercase: true))"
    }
    
    /// Convert Windows Virtual Key Code to macOS keycode
    /// Go layer sends Windows VK codes, not Linux keycodes
    private func windowsVKCodeToMacOS(_ vkCode: Int32) -> CGKeyCode? {
        // Windows Virtual Key Code to macOS keycode mapping
        // Reference: https://docs.microsoft.com/en-us/windows/win32/inputdev/virtual-key-codes
        let keyMap: [Int32: CGKeyCode] = [
            // Letter keys (VK_A = 0x41 = 65, VK_Z = 0x5A = 90)
            0x41: CGKeyCode(kVK_ANSI_A),   // VK_A
            0x42: CGKeyCode(kVK_ANSI_B),   // VK_B
            0x43: CGKeyCode(kVK_ANSI_C),   // VK_C
            0x44: CGKeyCode(kVK_ANSI_D),   // VK_D
            0x45: CGKeyCode(kVK_ANSI_E),   // VK_E
            0x46: CGKeyCode(kVK_ANSI_F),   // VK_F
            0x47: CGKeyCode(kVK_ANSI_G),   // VK_G
            0x48: CGKeyCode(kVK_ANSI_H),   // VK_H
            0x49: CGKeyCode(kVK_ANSI_I),   // VK_I
            0x4A: CGKeyCode(kVK_ANSI_J),   // VK_J
            0x4B: CGKeyCode(kVK_ANSI_K),   // VK_K
            0x4C: CGKeyCode(kVK_ANSI_L),   // VK_L
            0x4D: CGKeyCode(kVK_ANSI_M),   // VK_M
            0x4E: CGKeyCode(kVK_ANSI_N),   // VK_N
            0x4F: CGKeyCode(kVK_ANSI_O),   // VK_O
            0x50: CGKeyCode(kVK_ANSI_P),   // VK_P
            0x51: CGKeyCode(kVK_ANSI_Q),   // VK_Q (81)
            0x52: CGKeyCode(kVK_ANSI_R),   // VK_R
            0x53: CGKeyCode(kVK_ANSI_S),   // VK_S
            0x54: CGKeyCode(kVK_ANSI_T),   // VK_T
            0x55: CGKeyCode(kVK_ANSI_U),   // VK_U
            0x56: CGKeyCode(kVK_ANSI_V),   // VK_V
            0x57: CGKeyCode(kVK_ANSI_W),   // VK_W (87)
            0x58: CGKeyCode(kVK_ANSI_X),   // VK_X
            0x59: CGKeyCode(kVK_ANSI_Y),   // VK_Y
            0x5A: CGKeyCode(kVK_ANSI_Z),   // VK_Z
            
            // Number keys (VK_0 = 0x30 = 48, VK_9 = 0x39 = 57)
            0x30: CGKeyCode(kVK_ANSI_0),   // VK_0
            0x31: CGKeyCode(kVK_ANSI_1),   // VK_1
            0x32: CGKeyCode(kVK_ANSI_2),   // VK_2
            0x33: CGKeyCode(kVK_ANSI_3),   // VK_3
            0x34: CGKeyCode(kVK_ANSI_4),   // VK_4
            0x35: CGKeyCode(kVK_ANSI_5),   // VK_5
            0x36: CGKeyCode(kVK_ANSI_6),   // VK_6
            0x37: CGKeyCode(kVK_ANSI_7),   // VK_7
            0x38: CGKeyCode(kVK_ANSI_8),   // VK_8
            0x39: CGKeyCode(kVK_ANSI_9),   // VK_9
            
            // Function keys (VK_F1 = 0x70 = 112, VK_F12 = 0x7B = 123)
            0x70: CGKeyCode(kVK_F1),       // VK_F1
            0x71: CGKeyCode(kVK_F2),       // VK_F2
            0x72: CGKeyCode(kVK_F3),       // VK_F3
            0x73: CGKeyCode(kVK_F4),       // VK_F4
            0x74: CGKeyCode(kVK_F5),       // VK_F5
            0x75: CGKeyCode(kVK_F6),       // VK_F6
            0x76: CGKeyCode(kVK_F7),       // VK_F7
            0x77: CGKeyCode(kVK_F8),       // VK_F8
            0x78: CGKeyCode(kVK_F9),       // VK_F9
            0x79: CGKeyCode(kVK_F10),      // VK_F10
            0x7A: CGKeyCode(kVK_F11),      // VK_F11
            0x7B: CGKeyCode(kVK_F12),      // VK_F12
            
            // Control keys
            0x08: CGKeyCode(kVK_Delete),       // VK_BACK (Backspace)
            0x09: CGKeyCode(kVK_Tab),          // VK_TAB
            0x0D: CGKeyCode(kVK_Return),       // VK_RETURN (Enter)
            0x10: CGKeyCode(kVK_Shift),        // VK_SHIFT
            0x11: CGKeyCode(kVK_Control),      // VK_CONTROL
            0x12: CGKeyCode(kVK_Option),       // VK_MENU (Alt)
            0x14: CGKeyCode(kVK_CapsLock),     // VK_CAPITAL (Caps Lock)
            0x1B: CGKeyCode(kVK_Escape),       // VK_ESCAPE
            0x20: CGKeyCode(kVK_Space),        // VK_SPACE
            
            // Navigation keys
            0x21: CGKeyCode(kVK_PageUp),       // VK_PRIOR (Page Up)
            0x22: CGKeyCode(kVK_PageDown),     // VK_NEXT (Page Down)
            0x23: CGKeyCode(kVK_End),          // VK_END
            0x24: CGKeyCode(kVK_Home),         // VK_HOME
            0x25: CGKeyCode(kVK_LeftArrow),    // VK_LEFT
            0x26: CGKeyCode(kVK_UpArrow),      // VK_UP
            0x27: CGKeyCode(kVK_RightArrow),   // VK_RIGHT
            0x28: CGKeyCode(kVK_DownArrow),    // VK_DOWN
            0x2D: CGKeyCode(kVK_Help),         // VK_INSERT
            0x2E: CGKeyCode(kVK_ForwardDelete),// VK_DELETE
            
            // Numpad keys (VK_NUMPAD0 = 0x60 = 96, VK_NUMPAD9 = 0x69 = 105)
            0x60: CGKeyCode(kVK_ANSI_Keypad0), // VK_NUMPAD0
            0x61: CGKeyCode(kVK_ANSI_Keypad1), // VK_NUMPAD1
            0x62: CGKeyCode(kVK_ANSI_Keypad2), // VK_NUMPAD2
            0x63: CGKeyCode(kVK_ANSI_Keypad3), // VK_NUMPAD3
            0x64: CGKeyCode(kVK_ANSI_Keypad4), // VK_NUMPAD4
            0x65: CGKeyCode(kVK_ANSI_Keypad5), // VK_NUMPAD5
            0x66: CGKeyCode(kVK_ANSI_Keypad6), // VK_NUMPAD6
            0x67: CGKeyCode(kVK_ANSI_Keypad7), // VK_NUMPAD7
            0x68: CGKeyCode(kVK_ANSI_Keypad8), // VK_NUMPAD8
            0x69: CGKeyCode(kVK_ANSI_Keypad9), // VK_NUMPAD9
            0x6A: CGKeyCode(kVK_ANSI_KeypadMultiply), // VK_MULTIPLY
            0x6B: CGKeyCode(kVK_ANSI_KeypadPlus),     // VK_ADD
            0x6D: CGKeyCode(kVK_ANSI_KeypadMinus),    // VK_SUBTRACT
            0x6E: CGKeyCode(kVK_ANSI_KeypadDecimal),  // VK_DECIMAL
            0x6F: CGKeyCode(kVK_ANSI_KeypadDivide),   // VK_DIVIDE
            
            // OEM keys (symbols)
            0xBA: CGKeyCode(kVK_ANSI_Semicolon),    // VK_OEM_1 (;:)
            0xBB: CGKeyCode(kVK_ANSI_Equal),        // VK_OEM_PLUS (=+)
            0xBC: CGKeyCode(kVK_ANSI_Comma),        // VK_OEM_COMMA (,<)
            0xBD: CGKeyCode(kVK_ANSI_Minus),        // VK_OEM_MINUS (-_)
            0xBE: CGKeyCode(kVK_ANSI_Period),       // VK_OEM_PERIOD (.>)
            0xBF: CGKeyCode(kVK_ANSI_Slash),        // VK_OEM_2 (/?)
            0xC0: CGKeyCode(kVK_ANSI_Grave),        // VK_OEM_3 (`~)
            0xDB: CGKeyCode(kVK_ANSI_LeftBracket),  // VK_OEM_4 ([{)
            0xDC: CGKeyCode(kVK_ANSI_Backslash),    // VK_OEM_5 (\|)
            0xDD: CGKeyCode(kVK_ANSI_RightBracket), // VK_OEM_6 (]})
            0xDE: CGKeyCode(kVK_ANSI_Quote),        // VK_OEM_7 ('")
            
            // Windows/Command keys
            0x5B: CGKeyCode(kVK_Command),      // VK_LWIN (Left Windows)
            0x5C: CGKeyCode(kVK_RightCommand), // VK_RWIN (Right Windows)
            
            // Other keys
            0x90: CGKeyCode(kVK_ANSI_KeypadClear), // VK_NUMLOCK
            0x91: CGKeyCode(kVK_ANSI_KeypadClear), // VK_SCROLL (Scroll Lock)
            0xA0: CGKeyCode(kVK_Shift),        // VK_LSHIFT
            0xA1: CGKeyCode(kVK_RightShift),   // VK_RSHIFT
            0xA2: CGKeyCode(kVK_Control),      // VK_LCONTROL
            0xA3: CGKeyCode(kVK_RightControl), // VK_RCONTROL
            0xA4: CGKeyCode(kVK_Option),       // VK_LMENU (Left Alt)
            0xA5: CGKeyCode(kVK_RightOption),  // VK_RMENU (Right Alt)
        ]
        
        return keyMap[vkCode]
    }
    
    // MARK: - Statistics
    
    private func updateStats() {
        let now = Date()
        let elapsed = now.timeIntervalSince(lastStatTime)
        
        if elapsed >= 10.0 {
            let mouseFreq = Double(mousePacketCount) / elapsed
            let keyboardFreq = Double(keyboardPacketCount) / elapsed
            
            if mousePacketCount > 0 || keyboardPacketCount > 0 {
                logger.log("Packet statistics (last \(String(format: "%.1f", elapsed))s): Mouse: \(mousePacketCount) packets (\(String(format: "%.1f", mouseFreq)) pps), Keyboard: \(keyboardPacketCount) packets (\(String(format: "%.1f", keyboardFreq)) pps)", level: .info)
            }
            
            mousePacketCount = 0
            keyboardPacketCount = 0
            lastStatTime = now
        }
    }
    
    // MARK: - Per-Display Server Management
    
    /// Start UDP servers for a specific display
    func startForDisplay(_ displayID: CGDirectDisplayID, mousePort: UInt16, keyboardPort: UInt16) {
        displayServersLock.lock()
        defer { displayServersLock.unlock() }
        
        // Stop existing server for this display if any
        if displayServers[displayID] != nil {
            stopForDisplayLocked(displayID)
        }
        
        var serverInfo = DisplayUDPServerInfo(displayID: displayID)
        serverInfo.mousePort = mousePort
        serverInfo.keyboardPort = keyboardPort
        
        // Create and bind mouse UDP socket
        let mouseFD = socket(AF_INET, SOCK_DGRAM, 0)
        guard mouseFD >= 0 else {
            logger.log("Failed to create mouse UDP socket for display \(displayID): \(String(cString: strerror(errno)))", level: .error)
            return
        }
        
        var reuseAddr: Int32 = 1
        setsockopt(mouseFD, SOL_SOCKET, SO_REUSEADDR, &reuseAddr, socklen_t(MemoryLayout<Int32>.size))
        
        var mouseAddr = sockaddr_in()
        mouseAddr.sin_len = UInt8(MemoryLayout<sockaddr_in>.size)
        mouseAddr.sin_family = sa_family_t(AF_INET)
        mouseAddr.sin_port = mousePort.bigEndian
        mouseAddr.sin_addr.s_addr = INADDR_ANY.bigEndian
        
        let mouseBindResult = withUnsafePointer(to: &mouseAddr) { addrPtr in
            addrPtr.withMemoryRebound(to: sockaddr.self, capacity: 1) { sockaddrPtr in
                bind(mouseFD, sockaddrPtr, socklen_t(MemoryLayout<sockaddr_in>.size))
            }
        }
        
        guard mouseBindResult >= 0 else {
            logger.log("Failed to bind mouse server to port \(mousePort) for display \(displayID): \(String(cString: strerror(errno)))", level: .error)
            close(mouseFD)
            return
        }
        
        serverInfo.mouseSocketFD = mouseFD
        serverInfo.isMouseRunning = true
        
        // Create and bind keyboard UDP socket
        let kbdFD = socket(AF_INET, SOCK_DGRAM, 0)
        guard kbdFD >= 0 else {
            logger.log("Failed to create keyboard UDP socket for display \(displayID): \(String(cString: strerror(errno)))", level: .error)
            close(mouseFD)
            return
        }
        
        setsockopt(kbdFD, SOL_SOCKET, SO_REUSEADDR, &reuseAddr, socklen_t(MemoryLayout<Int32>.size))
        
        var kbdAddr = sockaddr_in()
        kbdAddr.sin_len = UInt8(MemoryLayout<sockaddr_in>.size)
        kbdAddr.sin_family = sa_family_t(AF_INET)
        kbdAddr.sin_port = keyboardPort.bigEndian
        kbdAddr.sin_addr.s_addr = INADDR_ANY.bigEndian
        
        let kbdBindResult = withUnsafePointer(to: &kbdAddr) { addrPtr in
            addrPtr.withMemoryRebound(to: sockaddr.self, capacity: 1) { sockaddrPtr in
                bind(kbdFD, sockaddrPtr, socklen_t(MemoryLayout<sockaddr_in>.size))
            }
        }
        
        guard kbdBindResult >= 0 else {
            logger.log("Failed to bind keyboard server to port \(keyboardPort) for display \(displayID): \(String(cString: strerror(errno)))", level: .error)
            close(mouseFD)
            close(kbdFD)
            return
        }
        
        serverInfo.keyboardSocketFD = kbdFD
        serverInfo.isKeyboardRunning = true
        
        // Store server info before starting threads (threads capture displayID)
        displayServers[displayID] = serverInfo
        
        // Start mouse receive thread
        let mouseThread = Thread { [weak self] in
            self?.displayMouseReceiveLoop(displayID: displayID)
        }
        mouseThread.name = "MouseReceiveThread-\(displayID)"
        mouseThread.qualityOfService = .userInteractive
        mouseThread.start()
        displayServers[displayID]?.mouseReceiveThread = mouseThread
        
        // Start keyboard receive thread
        let kbdThread = Thread { [weak self] in
            self?.displayKeyboardReceiveLoop(displayID: displayID)
        }
        kbdThread.name = "KeyboardReceiveThread-\(displayID)"
        kbdThread.qualityOfService = .userInteractive
        kbdThread.start()
        displayServers[displayID]?.keyboardReceiveThread = kbdThread
        
        logger.log("Per-display servers started for display \(displayID) - Mouse port: \(mousePort), Keyboard port: \(keyboardPort)", level: .info)
    }
    
    /// Stop UDP servers for a specific display
    func stopForDisplay(_ displayID: CGDirectDisplayID) {
        displayServersLock.lock()
        defer { displayServersLock.unlock() }
        stopForDisplayLocked(displayID)
    }
    
    /// Internal stop method (must be called with lock held)
    private func stopForDisplayLocked(_ displayID: CGDirectDisplayID) {
        guard var serverInfo = displayServers[displayID] else {
            logger.log("No server found for display \(displayID)", level: .warn)
            return
        }
        
        logger.log("Stopping per-display servers for display \(displayID)", level: .info)
        
        serverInfo.isMouseRunning = false
        serverInfo.isKeyboardRunning = false
        displayServers[displayID] = serverInfo
        
        // Close mouse socket
        if serverInfo.mouseSocketFD >= 0 {
            shutdown(serverInfo.mouseSocketFD, SHUT_RDWR)
            close(serverInfo.mouseSocketFD)
        }
        
        // Close keyboard socket
        if serverInfo.keyboardSocketFD >= 0 {
            shutdown(serverInfo.keyboardSocketFD, SHUT_RDWR)
            close(serverInfo.keyboardSocketFD)
        }
        
        // Wait for threads to finish
        if let thread = serverInfo.mouseReceiveThread {
            thread.cancel()
            var waitCount = 0
            while thread.isExecuting && waitCount < 50 {
                Thread.sleep(forTimeInterval: 0.01)
                waitCount += 1
            }
        }
        if let thread = serverInfo.keyboardReceiveThread {
            thread.cancel()
            var waitCount = 0
            while thread.isExecuting && waitCount < 50 {
                Thread.sleep(forTimeInterval: 0.01)
                waitCount += 1
            }
        }
        
        let mPort = serverInfo.mousePort
        let kPort = serverInfo.keyboardPort
        displayServers.removeValue(forKey: displayID)
        
        logger.log("Per-display servers stopped for display \(displayID) - freed mouse port: \(mPort), keyboard port: \(kPort)", level: .info)
    }
    
    /// Stop all per-display servers
    func stopAllDisplayServers() {
        displayServersLock.lock()
        let displayIDs = Array(displayServers.keys)
        displayServersLock.unlock()
        
        for displayID in displayIDs {
            stopForDisplay(displayID)
        }
        logger.log("All per-display servers stopped", level: .info)
    }
    
    /// Get ports for a specific display
    func getDisplayPorts(_ displayID: CGDirectDisplayID) -> (mousePort: UInt16, keyboardPort: UInt16)? {
        displayServersLock.lock()
        defer { displayServersLock.unlock() }
        
        guard let serverInfo = displayServers[displayID] else { return nil }
        return (mousePort: serverInfo.mousePort, keyboardPort: serverInfo.keyboardPort)
    }
    
    /// Check if a display has active servers
    func isDisplayServerRunning(_ displayID: CGDirectDisplayID) -> Bool {
        displayServersLock.lock()
        defer { displayServersLock.unlock() }
        
        guard let serverInfo = displayServers[displayID] else { return false }
        return serverInfo.isMouseRunning || serverInfo.isKeyboardRunning
    }
    
    /// Get all active display server IDs
    func getActiveDisplayServerIDs() -> [CGDirectDisplayID] {
        displayServersLock.lock()
        defer { displayServersLock.unlock() }
        return Array(displayServers.keys)
    }
    
    // MARK: - Per-Display Receive Loops
    
    private func displayMouseReceiveLoop(displayID: CGDirectDisplayID) {
        var buffer = [UInt8](repeating: 0, count: 1024)
        var srcAddr = sockaddr_in()
        var addrLen = socklen_t(MemoryLayout<sockaddr_in>.size)
        
        while true {
            displayServersLock.lock()
            guard let serverInfo = displayServers[displayID], serverInfo.isMouseRunning else {
                displayServersLock.unlock()
                break
            }
            let socketFD = serverInfo.mouseSocketFD
            displayServersLock.unlock()
            
            if Thread.current.isCancelled { break }
            
            let bytesRead = withUnsafeMutablePointer(to: &srcAddr) { addrPtr in
                addrPtr.withMemoryRebound(to: sockaddr.self, capacity: 1) { sockaddrPtr in
                    recvfrom(socketFD, &buffer, buffer.count, 0, sockaddrPtr, &addrLen)
                }
            }
            
            if bytesRead == MouseDataPacket.size {
                let data = Data(bytes: buffer, count: bytesRead)
                processMouseData(data, displayID: displayID)
            } else if bytesRead < 0 {
                if errno != EAGAIN && errno != EWOULDBLOCK {
                    displayServersLock.lock()
                    let stillRunning = displayServers[displayID]?.isMouseRunning ?? false
                    displayServersLock.unlock()
                    if stillRunning {
                        logger.log("Error receiving mouse data for display \(displayID): \(String(cString: strerror(errno)))", level: .error)
                    }
                }
                break
            }
        }
        logger.log("Mouse receive loop ended for display \(displayID)", level: .info)
    }
    
    private func displayKeyboardReceiveLoop(displayID: CGDirectDisplayID) {
        var buffer = [UInt8](repeating: 0, count: 1024)
        var srcAddr = sockaddr_in()
        var addrLen = socklen_t(MemoryLayout<sockaddr_in>.size)
        while true {
            displayServersLock.lock()
            guard let serverInfo = displayServers[displayID], serverInfo.isKeyboardRunning else {
                displayServersLock.unlock()
                break
            }
            let socketFD = serverInfo.keyboardSocketFD
            displayServersLock.unlock()
            
            if Thread.current.isCancelled { break }
            
            let bytesRead = withUnsafeMutablePointer(to: &srcAddr) { addrPtr in
                addrPtr.withMemoryRebound(to: sockaddr.self, capacity: 1) { sockaddrPtr in
                    recvfrom(socketFD, &buffer, buffer.count, 0, sockaddrPtr, &addrLen)
                }
            }
            
            if bytesRead == KeyboardDataPacket.size {
                let data = Data(bytes: buffer, count: bytesRead)
                processKeyboardData(data)
            } else if bytesRead < 0 {
                if errno != EAGAIN && errno != EWOULDBLOCK {
                    displayServersLock.lock()
                    let stillRunning = displayServers[displayID]?.isKeyboardRunning ?? false
                    displayServersLock.unlock()
                    if stillRunning {
                        logger.log("Error receiving keyboard data for display \(displayID): \(String(cString: strerror(errno)))", level: .error)
                    }
                }
                break
            }
        }
        logger.log("Keyboard receive loop ended for display \(displayID)", level: .info)
    }
}
