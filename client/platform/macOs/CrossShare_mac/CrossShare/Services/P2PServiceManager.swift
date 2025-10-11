//
//  P2PServiceManager.swift
//  CrossShare
//
//  Created by TS on 2025/9/8.
//  P2P Service Manager - Manages P2P service via XPC
//

import Foundation

class P2PServiceManager: NSObject {
    static let shared = P2PServiceManager()

    private let helperClient = HelperClient.shared
    private var isServiceRunning = false
    private var currentConfig: P2PConfig?
    
    private override init() {
        super.init()
    }
    
    func startP2PService(config: P2PConfig? = nil, completion: @escaping (Bool, String?) -> Void) {
        guard !isServiceRunning else {
            completion(false, "P2P service is already running")
            return
        }
        
        let serviceConfig = config ?? P2PConfig.defaultConfig()
        currentConfig = serviceConfig
        
        print("Starting P2P service via Helper with config: \(serviceConfig)")
        
        helperClient.connect { [weak self] (helperConnected: Bool, error: String?) in
            guard helperConnected else {
                completion(false, "Failed to connect to Helper: \(error ?? "Unknown error")")
                return
            }
            
            self?.helperClient.startGoService(config: serviceConfig.toXPCDict()) { success, errorMsg in
                if success {
                    self?.isServiceRunning = true
                    print("P2P service started successfully")
                }
            }
        }
    }
    
    func connectToDevice(deviceId: String, completion: @escaping (Bool, String?) -> Void) {
        // TODO: Implement connectToDevice in Helper
        completion(false, "Not implemented")
    }
    
    func checkServiceStatus(completion: @escaping (Bool, [String: Any]?) -> Void) {
        helperClient.connect { [weak self] (helperConnected: Bool, error: String?) in
            guard helperConnected else {
                completion(false, ["error": "Failed to connect to Helper: \(error ?? "Unknown error")"])
                return
            }
            self?.helperClient.getServiceStatus { (isRunning: Bool, info: [String: Any]?) in
                if isRunning {
                    var statusInfo = info ?? [:]
                    statusInfo["isServiceRunning"] = true
                    if let config = self?.currentConfig {
                        statusInfo["deviceName"] = config.deviceName
                        statusInfo["listenPort"] = config.listenPort
                    }
                    completion(true, statusInfo)
                } else {
                    completion(false, info)
                }
            }
        }
    }
}

extension P2PServiceManager {
    
    var isRunning: Bool {
        return isServiceRunning
    }
    
    var configuration: P2PConfig? {
        return currentConfig
    }
}
