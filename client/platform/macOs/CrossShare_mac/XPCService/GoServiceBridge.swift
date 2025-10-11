//
//  GoServiceBridge.swift
//  CrossShareHelper
//
//  Created by TS on 2025/8/27.
//  Go 服务桥梁层 - 连接 Swift 和 Go 服务
//

import Foundation

class GoServiceBridge {

    static let shared = GoServiceBridge()
    
    private let logger = XPCLogger.shared
    private var isServiceRunning = false
    private var currentConfig: [String: Any] = [:]
    
    private init() {
        logger.info("Go Service Bridge initialized")
    }
    
    func startService(config: [String: Any], completion: @escaping (Bool, String?) -> Void) {
        guard !isServiceRunning else {
            completion(false, "Service is already running")
            return
        }
        
        logger.info("Starting P2P Go service with MainInit - config: \(config)")
        
        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            guard let deviceName = config["deviceName"] as? String,
                  let serverId = config["serverId"] as? String,
                  let serverIpInfo = config["serverIpInfo"] as? String,
                  let listenHost = config["listenHost"] as? String,
                  let listenPort = config["listenPort"] as? Int32 else {
                
                DispatchQueue.main.async {
                    completion(false, "Invalid config parameters for MainInit")
                }
                return
            }
            
            self?.logger.info("Calling MainInit with: deviceName=\(deviceName), serverId=\(serverId), serverIpInfo=\(serverIpInfo), listenHost=\(listenHost), listenPort=\(listenPort)")
            
            let success = self?.callMainInit(
                deviceName: deviceName,
                serverId: serverId,
                serverIpInfo: serverIpInfo,
                listenHost: listenHost,
                listenPort: listenPort
            ) ?? false
            
            DispatchQueue.main.async {
                if success {
                    self?.isServiceRunning = true
                    self?.currentConfig = config
                    self?.logger.info("P2P MainInit completed successfully")
                    completion(true, nil)
                } else {
                    self?.logger.error("Failed to initialize P2P service with MainInit")
                    completion(false, "MainInit failed")
                }
            }
        }
    }
    
    private func callMainInit(deviceName: String, serverId: String, serverIpInfo: String, listenHost: String, listenPort: Int32) -> Bool {
        logger.info("Converting parameters for MainInit...")
        let deviceNameGo = deviceName.toGoStringXPC()
        let serverIdGo = serverId.toGoStringXPC()
        let serverIpInfoGo = serverIpInfo.toGoStringXPC()
        let listenHostGo = listenHost.toGoStringXPC()
        
        logger.info("Calling MainInit with Go strings...")
        MainInit(deviceNameGo, serverIdGo, serverIpInfoGo, listenHostGo, GoInt(listenPort))
        
        logger.info("MainInit call completed successfully")
        return true
    }
    
    func stopService(completion: @escaping (Bool) -> Void) {
        guard isServiceRunning else {
            completion(true)
            return
        }
        
        logger.info("Stopping Go service")
        
        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            let result = self?.callGoFunction {
                return StopCrossShareService()
            }
            
            DispatchQueue.main.async {
                let success = result?.success ?? true
                if success {
                    self?.isServiceRunning = false
                    self?.currentConfig = [:]
                    self?.logger.info("Go service stopped successfully")
                } else {
                    self?.logger.error("Failed to stop Go service")
                }
                completion(success)
            }
        }
    }
    
    func updateConfig(config: [String: Any], completion: @escaping (Bool, String?) -> Void) {
        logger.info("Updating Go service config")
        
        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            do {
                let configData = try self?.prepareConfigData(config: config)
                
                guard let configData = configData,
                      let configString = configData.toGoString() else {
                    DispatchQueue.main.async {
                        completion(false, "Failed to prepare config data")
                    }
                    return
                }
                
                let result = self?.callGoFunction {
                    return UpdateCrossShareConfig(configString)
                }
                
                let success = result?.success ?? false
                
                DispatchQueue.main.async {
                    if success {
                        self?.currentConfig = config
                        self?.logger.info("Go service config updated successfully")
                        completion(true, nil)
                    } else {
                        let errorMsg = result?.errorMessage ?? "Unknown error"
                        self?.logger.error("Failed to update config: \(errorMsg)")
                        completion(false, errorMsg)
                    }
                }
                
            } catch {
                DispatchQueue.main.async {
                    let errorMsg = "Failed to prepare config: \(error.localizedDescription)"
                    self?.logger.error(errorMsg)
                    completion(false, errorMsg)
                }
            }
        }
    }
    
    func startDeviceDiscovery(completion: @escaping (Bool, String?) -> Void) {
        logger.info("Starting device discovery")
        
        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            let result = self?.callGoFunction {
                return StartDeviceDiscovery()
            }
            
            let success = result?.success ?? false
            
            DispatchQueue.main.async {
                if success {
                    self?.logger.info("Device discovery started successfully")
                    completion(true, nil)
                } else {
                    let errorMsg = result?.errorMessage ?? "Unknown error"
                    self?.logger.error("Failed to start device discovery: \(errorMsg)")
                    completion(false, errorMsg)
                }
            }
        }
    }
    
    func stopDeviceDiscovery(completion: @escaping (Bool, String?) -> Void) {
        logger.info("Stopping device discovery")
        
        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            let result = self?.callGoFunction {
                return StopDeviceDiscovery()
            }
            
            let success = result?.success ?? false
            
            DispatchQueue.main.async {
                completion(success, success ? nil : (result?.errorMessage ?? "Unknown error"))
            }
        }
    }
    
    func getDiscoveredDevices(completion: @escaping ([CrossShareDevice], String?) -> Void) {
        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            let result = self?.callGoFunction {
                return GetDiscoveredDevices()
            }
            
            DispatchQueue.main.async {
                if let success = result?.success, success {
                    if let devicesData = result?.data,
                       let devices = self?.parseDeviceList(from: devicesData) {
                        completion(devices, nil)
                    } else {
                        completion([], nil)
                    }
                } else {
                    let errorMsg = result?.errorMessage ?? "Unknown error"
                    completion([], errorMsg)
                }
            }
        }
    }
    
    func connectToDevice(deviceId: String, completion: @escaping (Bool, String?) -> Void) {
        logger.info("Connecting to device: \(deviceId)")
        
        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            let result = self?.callGoFunction {
                return ConnectToDevice(deviceId.toCStringXPC())
            }
            
            let success = result?.success ?? false
            
            DispatchQueue.main.async {
                if success {
                    self?.logger.info("Connected to device: \(deviceId)")
                    completion(true, nil)
                } else {
                    let errorMsg = result?.errorMessage ?? "Unknown error"
                    self?.logger.error("Failed to connect to device \(deviceId): \(errorMsg)")
                    completion(false, errorMsg)
                }
            }
        }
    }
    
    func disconnectFromDevice(deviceId: String, completion: @escaping (Bool, String?) -> Void) {
        logger.info("Disconnecting from device: \(deviceId)")
        
        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            let result = self?.callGoFunction {
                return DisconnectFromDevice(deviceId.toCStringXPC())
            }
            
            let success = result?.success ?? false
            
            DispatchQueue.main.async {
                completion(success, success ? nil : (result?.errorMessage ?? "Unknown error"))
            }
        }
    }

    func sendFile(filePath: String, toDevice deviceId: String, completion: @escaping (String?, String?) -> Void) {
        logger.info("Sending file: \(filePath) to device: \(deviceId)")
        
        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            let result = self?.callGoFunction {
                return SendFile(filePath.toCStringXPC(), deviceId.toCStringXPC())
            }
            
            DispatchQueue.main.async {
                if let success = result?.success, success {
                    let transferId = result?.data ?? ""
                    self?.logger.info("File transfer started: \(transferId)")
                    completion(transferId, nil)
                } else {
                    let errorMsg = result?.errorMessage ?? "Unknown error"
                    self?.logger.error("Failed to send file: \(errorMsg)")
                    completion(nil, errorMsg)
                }
            }
        }
    }
    
    func sendFolder(folderPath: String, toDevice deviceId: String, completion: @escaping (String?, String?) -> Void) {
        logger.info("Sending folder: \(folderPath) to device: \(deviceId)")
        
        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            let result = self?.callGoFunction {
                return SendFolder(folderPath.toCStringXPC(), deviceId.toCStringXPC())
            }
            
            DispatchQueue.main.async {
                if let success = result?.success, success {
                    let transferId = result?.data ?? ""
                    completion(transferId, nil)
                } else {
                    let errorMsg = result?.errorMessage ?? "Unknown error"
                    completion(nil, errorMsg)
                }
            }
        }
    }
    
    func cancelTransfer(transferId: String, completion: @escaping (Bool) -> Void) {
        logger.info("Cancelling transfer: \(transferId)")
        
        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            let result = self?.callGoFunction {
                return CancelTransfer(transferId.toCStringXPC())
            }
            
            DispatchQueue.main.async {
                completion(result?.success ?? false)
            }
        }
    }
    
    func getLocalIPAddress(completion: @escaping (String?) -> Void) {
        NetworkUtils.shared.getLocalIPAddress(completion: completion)
    }
    
    func checkPortAvailability(port: Int, completion: @escaping (Bool) -> Void) {
        NetworkUtils.shared.checkPortAvailability(port: port, completion: completion)
    }
    
    var isRunning: Bool {
        return isServiceRunning
    }
    
    var config: [String: Any] {
        return currentConfig
    }
    
    func healthCheck(completion: @escaping (Bool, [String: Any]) -> Void) {
        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            let result = self?.callGoFunction {
                return HealthCheck()
            }
            
            DispatchQueue.main.async {
                let isHealthy = result?.success ?? false
                var healthInfo: [String: Any] = [
                    "isRunning": self?.isServiceRunning ?? false,
                    "timestamp": Date().timeIntervalSince1970
                ]
                
                if let data = result?.data {
                    healthInfo["details"] = data
                }
                
                completion(isHealthy, healthInfo)
            }
        }
    }
}

private struct SwiftGoResult {
    let success: Bool
    let data: String?
    let errorMessage: String?
    
    init(cResult: UnsafePointer<GoResult>) {
        self.success = cResult.pointee.success != 0
        
        if let dataPtr = cResult.pointee.data {
            self.data = String(cString: dataPtr)
        } else {
            self.data = nil
        }
        
        if let errorPtr = cResult.pointee.error_message {
            self.errorMessage = String(cString: errorPtr)
        } else {
            self.errorMessage = nil
        }
    }
}

// MARK: - Private Methods

private extension GoServiceBridge {
    
    func prepareConfigData(config: [String: Any]) throws -> [String: Any] {
        var processedConfig = config
        
        if processedConfig["device_name"] == nil {
            processedConfig["device_name"] = Host.current().localizedName ?? "Mac Device"
        }
        
        if processedConfig["device_id"] == nil {
            processedConfig["device_id"] = UUID().uuidString
        }
        
        if processedConfig["listen_port"] == nil {
            processedConfig["listen_port"] = 8080
        }
        
        return processedConfig
    }
    
    func callGoFunction(_ goCall: () -> UnsafeMutablePointer<GoResult>?) -> SwiftGoResult? {
        guard let cResult = goCall() else { return nil }
        let swiftResult = SwiftGoResult(cResult: UnsafePointer(cResult))
        if let data = cResult.pointee.data {
            free(data)
        }
        if let errorMessage = cResult.pointee.error_message {
            free(errorMessage)
        }
        free(cResult)
        return swiftResult
    }
    
    func parseDeviceList(from data: String) -> [CrossShareDevice]? {
        guard let jsonData = data.data(using: .utf8) else { return nil }
        
        do {
            if let deviceDicts = try JSONSerialization.jsonObject(with: jsonData) as? [[String: Any]] {
                return deviceDicts.compactMap { CrossShareDevice(dictionary: $0) }
            }
        } catch {
            logger.error("Failed to parse device list: \(error)")
        }
        
        return nil
    }
}

extension Dictionary where Key == String, Value == Any {
    func toGoString() -> UnsafePointer<CChar>? {
        guard let jsonData = try? JSONSerialization.data(withJSONObject: self),
              let jsonString = String(data: jsonData, encoding: .utf8) else {
            return nil
        }
        return jsonString.toCStringXPC()
    }
}

