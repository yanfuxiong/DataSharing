//
//  BundleIdentifiers.swift
//  CrossShare
//
//  Created by TS on 2025/9/8.
//  Bundle Identifier utilities and constants
//

import Foundation

struct BundleIdentifiers {
    
    private static let xpcServiceSuffix = "XPCService"
    
    private static let helperSuffix = "helper"
    
    static var mainApp: String {
        return Bundle.main.bundleIdentifier ?? "com.realtek.crossshare.macos"
    }
    
    static var xpcService: String {
        return "\(mainApp).\(xpcServiceSuffix)"
    }
    
    static var helper: String {
        return "\(mainApp).\(helperSuffix)"
    }
    
    static var crossShareHelper: String {
        return "\(mainApp).helper"
    }
    
    static var ddcciHelper: String {
        return "\(mainApp).ddcciHelper"
    }
    
    static func printAllIdentifiers() {
        print("=== Bundle Identifiers ===")
        print("Main App: \(mainApp)")
        print("XPC Service: \(xpcService)")
        print("Helper: \(helper)")
        print("CrossShare Helper: \(crossShareHelper)")
        print("Ddcci Helper: \(ddcciHelper)")
        print("========================")
    }
    
    static func validateXPCService() -> Bool {
        // XPC Services 位于 App.app/Contents/XPCServices/ 目录
        
        let bundlePath = Bundle.main.bundlePath
        let xpcServicePath = "\(bundlePath)/Contents/XPCServices/\(xpcServiceSuffix).xpc"
        
        print("=== Validating XPC Service ===")
        print("Bundle path: \(bundlePath)")
        print("Looking for XPC Service at: \(xpcServicePath)")
        
        guard FileManager.default.fileExists(atPath: xpcServicePath) else {
            print("XPC Service not found at expected path")
            
            // 列出 XPCServices 目录内容用于调试
            let xpcServicesDir = "\(bundlePath)/Contents/XPCServices"
            if let contents = try? FileManager.default.contentsOfDirectory(atPath: xpcServicesDir) {
                print("XPCServices directory contents: \(contents)")
            } else {
                print("XPCServices directory not found or empty")
            }
            return false
        }
        
        print("XPC Service file exists")
        
        // 创建 Bundle 对象并验证 identifier
        guard let xpcBundle = Bundle(path: xpcServicePath) else {
            print("Failed to create Bundle object for XPC Service")
            return false
        }
        
        guard let actualIdentifier = xpcBundle.bundleIdentifier else {
            print("Failed to get bundle identifier from XPC Service")
            return false
        }
        
        print("XPC Service actual identifier: \(actualIdentifier)")
        print("Expected identifier: \(xpcService)")
        
        let isValid = actualIdentifier == xpcService
        print(isValid ? "Bundle identifier matches!" : "Bundle identifier mismatch!")
        print("==============================")
        
        return isValid
    }
}

