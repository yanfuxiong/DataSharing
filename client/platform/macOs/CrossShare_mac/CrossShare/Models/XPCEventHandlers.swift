//
//  XPCEventHandlers.swift
//  CrossShare
//
//  Created by TS on 2025/9/11.
//

import Foundation

struct XPCEventHandlers {
    
    var onConnected: (() -> Void)?
    var onConnectionLost: (() -> Void)?
    var onConnectionInterrupted: (() -> Void)?
    var onReconnected: (() -> Void)?
    var onReconnectionFailed: (() -> Void)?
    var onCountUpdated: ((Int) -> Void)?
    
    var onAuthRequested: ((UInt32) -> Void)?
    var onDeviceDataReceived: (([String: Any]) -> Void)?
}
