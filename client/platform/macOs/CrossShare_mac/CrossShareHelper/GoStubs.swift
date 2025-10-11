//
//  GoStubs.swift
//  CrossShareHelper
//
//  Created by TS on 2025/9/11.
//

import Foundation

struct GoResult {
    var success: UInt8
    var data: UnsafeMutablePointer<CChar>?
    var error_message: UnsafeMutablePointer<CChar>?
}

func HealthCheck() -> UnsafeMutablePointer<GoResult>? {
    let result = UnsafeMutablePointer<GoResult>.allocate(capacity: 1)
    result.pointee.success = 1
    result.pointee.data = strdup("{\"status\": \"healthy\"}")
    result.pointee.error_message = nil
    return result
}

func GetLocalIPAddress() -> UnsafeMutablePointer<GoResult>? {
    let result = UnsafeMutablePointer<GoResult>.allocate(capacity: 1)
    result.pointee.success = 1
    result.pointee.data = strdup("127.0.0.1")
    result.pointee.error_message = nil
    return result
}

func CheckPortAvailability(_ port: Int32) -> UnsafeMutablePointer<GoResult>? {
    let result = UnsafeMutablePointer<GoResult>.allocate(capacity: 1)
    result.pointee.success = 1
    result.pointee.data = nil
    result.pointee.error_message = nil
    return result
}
