//
//  XPCService.swift
//  CrossShareHelper
//
//  Created by TS on 2025/8/27.
//  XPC 服务实现
//

import Foundation
import Network

/// CrossShare XPC 服务实现类
class CrossShareXPCService: NSObject, CrossShareXPCProtocol {
    
    // MARK: - Properties
    
    private var goServiceManager: GoServiceManager?
    private var deviceManager: DeviceManager?
    private var transferManager: TransferManager?
    private var clipboardManager: ClipboardManager?
    private var networkMonitor: NetworkMonitor?
    
    private let queue = DispatchQueue(label: "com.crossshare.xpc.queue", qos: .userInitiated)
    private let logger = XPCLogger.shared
    
    private var isServiceRunning = false
    private var currentConfig: [String: Any] = [:]
    private var delegates: [CrossShareXPCDelegate] = []
    
    override init() {
        super.init()
        print("[XPCService] Initializing...")
        setupManagers()
        print("[XPCService] Managers setup complete")
        logger.log("XPC Service initialized", level: .info)
        print("[XPCService] Initialization complete")
    }
    
    private func setupManagers() {
        logger.log("Setting up managers...", level: .info)
        
        goServiceManager = GoServiceManager(delegate: self)
        logger.log("GoServiceManager created: \(goServiceManager != nil)", level: .info)
        
        deviceManager = DeviceManager(delegate: self)
        logger.log("DeviceManager created: \(deviceManager != nil)", level: .info)
        
        transferManager = TransferManager(delegate: self)
        logger.log("TransferManager created: \(transferManager != nil)", level: .info)
        
        clipboardManager = ClipboardManager(delegate: self)
        logger.log("ClipboardManager created: \(clipboardManager != nil)", level: .info)
        
        networkMonitor = NetworkMonitor(delegate: self)
        logger.log("NetworkMonitor created: \(networkMonitor != nil)", level: .info)
        
        logger.log("All managers setup complete", level: .info)
    }
    
    func startGoService(config: [String: Any], completion: @escaping (Bool, String?) -> Void) {
        print("[XPCService] startGoService called")
        logger.log("startGoService called with config: \(config)", level: .info)
        
        queue.async { [weak self] in
            guard let self = self else {
                print("[XPCService] ERROR: self is nil in startGoService")
                DispatchQueue.main.async {
                    completion(false, "Service unavailable - self is nil")
                }
                return
            }
            
            print("[XPCService] self is valid, checking goServiceManager...")
            self.logger.log("Starting Go service with config: \(config)", level: .info)
            self.currentConfig = config
            
            guard let goManager = self.goServiceManager else {
                print("[XPCService] ERROR: goServiceManager is nil!")
                self.logger.log("GoServiceManager is nil!", level: .error)
                DispatchQueue.main.async {
                    completion(false, "Go service manager not available")
                }
                return
            }
            
            print("[XPCService] goServiceManager is valid, starting service...")
            
            goManager.start(config: config) { [weak self] success, error in
                DispatchQueue.main.async {
                    if success {
                        self?.isServiceRunning = true
                        self?.logger.log("Go service started successfully", level: .info)
                        self?.notifyDelegates { $0.didChangeServiceStatus(isRunning: true, info: config) }
                    } else {
                        self?.logger.log("Failed to start Go service: \(error ?? "Unknown error")", level: .error)
                    }
                    completion(success, error)
                }
            }
        }
    }
    
    func stopGoService(completion: @escaping (Bool) -> Void) {
        queue.async { [weak self] in
            guard let self = self else {
                completion(false)
                return
            }
            
            self.logger.log("Stopping Go service", level: .info)
            
            self.goServiceManager?.stop { [weak self] success in
                DispatchQueue.main.async {
                    if success {
                        self?.isServiceRunning = false
                        self?.logger.log("Go service stopped successfully", level: .info)
                        self?.notifyDelegates { $0.didChangeServiceStatus(isRunning: false, info: nil) }
                    }
                    completion(success)
                }
            }
        }
    }
    
    func getServiceStatus(completion: @escaping (Bool, [String: Any]?) -> Void) {
        queue.async { [weak self] in
            guard let self = self else {
                completion(false, nil)
                return
            }
            
            let status = self.isServiceRunning
            let info = status ? self.currentConfig : nil
            
            DispatchQueue.main.async {
                completion(status, info)
            }
        }
    }
    
    func startDeviceDiscovery(completion: @escaping (Bool) -> Void) {
        queue.async { [weak self] in
            self?.deviceManager?.startDiscovery { success in
                DispatchQueue.main.async {
                    completion(success)
                }
            }
        }
    }
    
    func stopDeviceDiscovery(completion: @escaping (Bool) -> Void) {
        queue.async { [weak self] in
            self?.deviceManager?.stopDiscovery { success in
                DispatchQueue.main.async {
                    completion(success)
                }
            }
        }
    }
    
    func getDiscoveredDevices(completion: @escaping ([[String: Any]]) -> Void) {
        queue.async { [weak self] in
            let devices = self?.deviceManager?.getDiscoveredDevices() ?? []
            DispatchQueue.main.async {
                completion(devices.map { $0.toDictionary() })
            }
        }
    }
    
    func connectToDevice(deviceId: String, completion: @escaping (Bool, String?) -> Void) {
        queue.async { [weak self] in
            self?.deviceManager?.connectToDevice(deviceId: deviceId, completion: { success, error in
                DispatchQueue.main.async {
                    completion(success, error)
                }
            })
        }
    }
    
    func disconnectFromDevice(deviceId: String, completion: @escaping (Bool) -> Void) {
        queue.async { [weak self] in
            self?.deviceManager?.disconnectFromDevice(deviceId: deviceId) { success in
                DispatchQueue.main.async {
                    completion(success)
                }
            }
        }
    }
    

    func sendFile(filePath: String, toDevice deviceId: String, completion: @escaping (Bool, String?) -> Void) {
        queue.async { [weak self] in
            self?.transferManager?.sendFile(filePath: filePath, toDevice: deviceId) { success, error in
                DispatchQueue.main.async {
                    completion(success, error)
                }
            }
        }
    }
    
    func sendFolder(folderPath: String, toDevice deviceId: String, completion: @escaping (Bool, String?) -> Void) {
        queue.async { [weak self] in
            self?.transferManager?.sendFolder(folderPath: folderPath, toDevice: deviceId) { success, error in
                DispatchQueue.main.async {
                    completion(success, error)
                }
            }
        }
    }
    
    func cancelTransfer(transferId: String, completion: @escaping (Bool) -> Void) {
        queue.async { [weak self] in
            self?.transferManager?.cancelTransfer(transferId: transferId) { success in
                DispatchQueue.main.async {
                    completion(success)
                }
            }
        }
    }
    
    func getTransferProgress(completion: @escaping ([[String: Any]]) -> Void) {
        queue.async { [weak self] in
            let transfers = self?.transferManager?.getActiveTransfers() ?? []
            DispatchQueue.main.async {
                completion(transfers.map { $0.toDictionary() })
            }
        }
    }
    
    // MARK: - 剪贴板同步
    
    func startClipboardMonitoring(completion: @escaping (Bool) -> Void) {
        queue.async { [weak self] in
            self?.clipboardManager?.startMonitoring { success in
                DispatchQueue.main.async {
                    completion(success)
                }
            }
        }
    }
    
    func stopClipboardMonitoring(completion: @escaping (Bool) -> Void) {
        queue.async { [weak self] in
            self?.clipboardManager?.stopMonitoring { success in
                DispatchQueue.main.async {
                    completion(success)
                }
            }
        }
    }
    
    func syncClipboard(toDevice deviceId: String, completion: @escaping (Bool) -> Void) {
        queue.async { [weak self] in
            self?.clipboardManager?.syncToDevice(deviceId: deviceId) { success in
                DispatchQueue.main.async {
                    completion(success)
                }
            }
        }
    }
    
    // MARK: - 网络状态
    
    func getNetworkInfo(completion: @escaping ([String: Any]) -> Void) {
        queue.async { [weak self] in
            let networkInfo = self?.networkMonitor?.getCurrentNetworkInfo() ?? [:]
            DispatchQueue.main.async {
                completion(networkInfo)
            }
        }
    }
    
    func getLocalIPAddress(completion: @escaping (String?) -> Void) {
        queue.async { [weak self] in
            let ipAddress = self?.networkMonitor?.getLocalIPAddress()
            DispatchQueue.main.async {
                completion(ipAddress)
            }
        }
    }
    
    func checkPortAvailability(port: Int, completion: @escaping (Bool) -> Void) {
        queue.async { [weak self] in
            let available = self?.networkMonitor?.isPortAvailable(port: port) ?? false
            DispatchQueue.main.async {
                completion(available)
            }
        }
    }
    
    // MARK: - 配置管理
    
    func updateConfiguration(config: [String: Any], completion: @escaping (Bool, String?) -> Void) {
        queue.async { [weak self] in
            guard let self = self else {
                completion(false, "Service unavailable")
                return
            }
            
            self.currentConfig = config
            
            // 如果服务正在运行，重启服务以应用新配置
            if self.isServiceRunning {
                self.goServiceManager?.updateConfig(config: config) { success, error in
                    DispatchQueue.main.async {
                        completion(success, error)
                    }
                }
            } else {
                DispatchQueue.main.async {
                    completion(true, nil)
                }
            }
        }
    }
    
    func getCurrentConfiguration(completion: @escaping ([String: Any]) -> Void) {
        queue.async { [weak self] in
            let config = self?.currentConfig ?? [:]
            DispatchQueue.main.async {
                completion(config)
            }
        }
    }
    
    // MARK: - 日志与调试
    
    func getServiceLogs(lines: Int, completion: @escaping ([String]) -> Void) {
        queue.async { [weak self] in
            let logs = self?.logger.getRecentLogs(lines: lines) ?? []
            DispatchQueue.main.async {
                completion(logs)
            }
        }
    }
    
    func setLogLevel(level: Int, completion: @escaping (Bool) -> Void) {
        queue.async { [weak self] in
            self?.logger.setLogLevel(level: level)
            DispatchQueue.main.async {
                completion(true)
            }
        }
    }
    
    func healthCheck(completion: @escaping (Bool, [String: Any]) -> Void) {
        queue.async { [weak self] in
            guard let self = self else {
                completion(false, ["error": "Service unavailable"])
                return
            }
            
            var healthInfo: [String: Any] = [
                "serviceRunning": self.isServiceRunning,
                "timestamp": Date().timeIntervalSince1970,
                "memoryUsage": ProcessInfo.processInfo.physicalMemory
            ]
            
            // 检查各个组件状态
            healthInfo["goService"] = self.goServiceManager?.isHealthy() ?? true  // 默认为 true，因为可能还未初始化
            healthInfo["deviceManager"] = self.deviceManager?.isHealthy() ?? true
            healthInfo["transferManager"] = self.transferManager?.isHealthy() ?? true
            healthInfo["clipboardManager"] = self.clipboardManager?.isHealthy() ?? true
            healthInfo["networkMonitor"] = self.networkMonitor?.isHealthy() ?? true
            
            // XPC Service 本身能响应就认为是健康的
            // 不要求所有子服务都必须运行
            let isHealthy = true
            
            DispatchQueue.main.async {
                completion(isHealthy, healthInfo)
            }
        }
    }
    
    // MARK: - Delegate Management
    
    func addDelegate(_ delegate: CrossShareXPCDelegate) {
        delegates.append(delegate)
    }
    
    func removeDelegate(_ delegate: CrossShareXPCDelegate) {
        // 由于 CrossShareXPCDelegate 是 @objc 协议，无法直接比较
        // 这里需要主应用端管理连接
    }
    
    private func notifyDelegates(_ block: (CrossShareXPCDelegate) -> Void) {
        delegates.forEach { delegate in
            block(delegate)
        }
    }
}

// MARK: - Manager Delegates

extension CrossShareXPCService: GoServiceManagerDelegate {
    func goServiceDidStart() {
        logger.log("Go service started", level: .info)
        notifyDelegates { $0.didChangeServiceStatus(isRunning: true, info: currentConfig) }
    }
    
    func goServiceDidStop() {
        logger.log("Go service stopped", level: .info)
        notifyDelegates { $0.didChangeServiceStatus(isRunning: false, info: nil) }
    }
    
    func goServiceDidEncounterError(_ error: String) {
        logger.log("Go service error: \(error)", level: .error)
        notifyDelegates { $0.didEncounterError(error: ["source": "goService", "message": error]) }
    }
}

extension CrossShareXPCService: DeviceManagerDelegate {
    func deviceManager(_ manager: DeviceManager, didDiscoverDevice device: CrossShareDevice) {
        logger.log("Discovered device: \(device.name) (\(device.id))", level: .info)
        notifyDelegates { $0.didDiscoverDevice(device: device.toDictionary()) }
    }
    
    func deviceManager(_ manager: DeviceManager, didLoseDevice deviceId: String) {
        logger.log("Lost device: \(deviceId)", level: .info)
        notifyDelegates { $0.didLoseDevice(deviceId: deviceId) }
    }
    
    func deviceManager(_ manager: DeviceManager, didChangeConnectionStatus deviceId: String, connected: Bool) {
        logger.log("Device \(deviceId) connection changed: \(connected)", level: .info)
        notifyDelegates { $0.didChangeDeviceConnection(deviceId: deviceId, connected: connected) }
    }
}

extension CrossShareXPCService: TransferManagerDelegate {
    func transferManager(_ manager: TransferManager, didStartReceivingFile transferInfo: [String: Any]) {
        logger.log("Started receiving file: \(transferInfo)", level: .info)
        notifyDelegates { $0.didStartReceivingFile(transferInfo: transferInfo) }
    }
    
    func transferManager(_ manager: TransferManager, didUpdateProgress progress: [String: Any]) {
        notifyDelegates { $0.didUpdateTransferProgress(progress: progress) }
    }
    
    func transferManager(_ manager: TransferManager, didCompleteTransfer result: [String: Any]) {
        logger.log("Transfer completed: \(result)", level: .info)
        notifyDelegates { $0.didCompleteTransfer(result: result) }
    }
    
    func transferManager(_ manager: TransferManager, didFailTransfer error: [String: Any]) {
        logger.log("Transfer failed: \(error)", level: .error)
        notifyDelegates { $0.didFailTransfer(error: error) }
    }
}

extension CrossShareXPCService: ClipboardManagerDelegate {
    func clipboardManager(_ manager: ClipboardManager, didReceiveData data: [String: Any]) {
        logger.log("Received clipboard data", level: .info)
        notifyDelegates { $0.didReceiveClipboardData(clipboardData: data) }
    }
}

extension CrossShareXPCService: NetworkMonitorDelegate {
    func networkMonitor(_ monitor: NetworkMonitor, didChangeNetworkStatus networkInfo: [String: Any]) {
        logger.log("Network status changed: \(networkInfo)", level: .info)
        notifyDelegates { $0.didChangeNetworkStatus(networkInfo: networkInfo) }
    }
}

// MARK: - Delegate Protocols

protocol GoServiceManagerDelegate: AnyObject {
    func goServiceDidStart()
    func goServiceDidStop()
    func goServiceDidEncounterError(_ error: String)
}

protocol DeviceManagerDelegate: AnyObject {
    func deviceManager(_ manager: DeviceManager, didDiscoverDevice device: CrossShareDevice)
    func deviceManager(_ manager: DeviceManager, didLoseDevice deviceId: String)
    func deviceManager(_ manager: DeviceManager, didChangeConnectionStatus deviceId: String, connected: Bool)
}

protocol TransferManagerDelegate: AnyObject {
    func transferManager(_ manager: TransferManager, didStartReceivingFile transferInfo: [String: Any])
    func transferManager(_ manager: TransferManager, didUpdateProgress progress: [String: Any])
    func transferManager(_ manager: TransferManager, didCompleteTransfer result: [String: Any])
    func transferManager(_ manager: TransferManager, didFailTransfer error: [String: Any])
}

protocol ClipboardManagerDelegate: AnyObject {
    func clipboardManager(_ manager: ClipboardManager, didReceiveData data: [String: Any])
}

protocol NetworkMonitorDelegate: AnyObject {
    func networkMonitor(_ monitor: NetworkMonitor, didChangeNetworkStatus networkInfo: [String: Any])
}
