//
//  Managers.swift
//  CrossShareHelper
//
//  Created by TS on 2025/8/27.
//  各种管理器的基础实现
//

import Foundation
import Network
import AppKit

// MARK: - Go Service Manager

/// Go 服务管理器
/// 负责启动、停止和管理 Go 服务进程
class GoServiceManager {
    
    weak var delegate: GoServiceManagerDelegate?
    private let logger = XPCLogger.shared
    private let goBridge = GoServiceBridge.shared
    
    init(delegate: GoServiceManagerDelegate?) {
        self.delegate = delegate
    }
    
    func start(config: [String: Any], completion: @escaping (Bool, String?) -> Void) {
        guard !goBridge.isRunning else {
            completion(false, "Service is already running")
            return
        }
        
        logger.info("Starting Go service with config: \(config)")
        
        goBridge.startService(config: config) { [weak self] success, error in
            if success {
                self?.delegate?.goServiceDidStart()
                self?.logger.info("Go service started successfully")
            } else {
                self?.delegate?.goServiceDidEncounterError(error ?? "Unknown error")
            }
            completion(success, error)
        }
    }
    
    func stop(completion: @escaping (Bool) -> Void) {
        guard goBridge.isRunning else {
            completion(true)
            return
        }
        
        logger.info("Stopping Go service")
        
        goBridge.stopService { [weak self] success in
            if success {
                self?.delegate?.goServiceDidStop()
                self?.logger.info("Go service stopped successfully")
            } else {
                self?.logger.error("Failed to stop Go service")
            }
            completion(success)
        }
    }
    
    func updateConfig(config: [String: Any], completion: @escaping (Bool, String?) -> Void) {
        logger.info("Updating Go service config")
        
        goBridge.updateConfig(config: config, completion: completion)
    }
    
    func isHealthy() -> Bool {
        return goBridge.isRunning
    }
}

class DeviceManager {
    
    weak var delegate: DeviceManagerDelegate?
    private let logger = XPCLogger.shared
    private let goBridge = GoServiceBridge.shared
    private var discoveredDevices: [CrossShareDevice] = []
    private var isDiscovering = false
    private var discoveryTimer: Timer?
    
    init(delegate: DeviceManagerDelegate?) {
        self.delegate = delegate
        setupPeriodicDeviceUpdate()
    }
    
    func startDiscovery(completion: @escaping (Bool) -> Void) {
        guard !isDiscovering else {
            completion(true)
            return
        }
        
        logger.info("Starting device discovery")
        
        goBridge.startDeviceDiscovery { [weak self] success, error in
            if success {
                self?.isDiscovering = true
                self?.startPeriodicUpdate()
                self?.logger.info("Device discovery started successfully")
            } else {
                self?.logger.error("Failed to start device discovery: \(error ?? "Unknown error")")
            }
            completion(success)
        }
    }
    
    func stopDiscovery(completion: @escaping (Bool) -> Void) {
        logger.info("Stopping device discovery")
        
        goBridge.stopDeviceDiscovery { [weak self] success, error in
            if success {
                self?.isDiscovering = false
                self?.stopPeriodicUpdate()
                self?.logger.info("Device discovery stopped successfully")
            } else {
                self?.logger.error("Failed to stop device discovery: \(error ?? "Unknown error")")
            }
            completion(success)
        }
    }
    
    func getDiscoveredDevices() -> [CrossShareDevice] {
        return discoveredDevices
    }
    
    func connectToDevice(deviceId: String, completion: @escaping (Bool, String?) -> Void) {
        logger.info("Connecting to device: \(deviceId)")
        
        goBridge.connectToDevice(deviceId: deviceId) { [weak self] success, error in
            if success {
                self?.delegate?.deviceManager(self!, didChangeConnectionStatus: deviceId, connected: true)
                self?.logger.info("Connected to device: \(deviceId)")
            } else {
                self?.logger.error("Failed to connect to device \(deviceId): \(error ?? "Unknown error")")
            }
            completion(success, error)
        }
    }
    
    func disconnectFromDevice(deviceId: String, completion: @escaping (Bool) -> Void) {
        logger.info("Disconnecting from device: \(deviceId)")
        
        goBridge.disconnectFromDevice(deviceId: deviceId) { [weak self] success, error in
            if success {
                self?.delegate?.deviceManager(self!, didChangeConnectionStatus: deviceId, connected: false)
                self?.logger.info("Disconnected from device: \(deviceId)")
            } else {
                self?.logger.error("Failed to disconnect from device \(deviceId): \(error ?? "Unknown error")")
            }
            completion(success)
        }
    }
    
    func isHealthy() -> Bool {
        return true
    }
    
    private func setupPeriodicDeviceUpdate() {
        discoveryTimer = Timer.scheduledTimer(withTimeInterval: 5.0, repeats: true) { [weak self] _ in
            guard let self = self, self.isDiscovering else { return }
            self.updateDeviceList()
        }
    }
    
    private func startPeriodicUpdate() {
        discoveryTimer?.invalidate()
        setupPeriodicDeviceUpdate()
    }
    
    private func stopPeriodicUpdate() {
        discoveryTimer?.invalidate()
        discoveryTimer = nil
    }
    
    private func updateDeviceList() {
        goBridge.getDiscoveredDevices { [weak self] devices, error in
            guard let self = self else { return }
            
            if let error = error {
                self.logger.error("Failed to get device list: \(error)")
                return
            }
            
            let oldDeviceIds = Set(self.discoveredDevices.map { $0.id })
            let newDeviceIds = Set(devices.map { $0.id })
            
            let addedDeviceIds = newDeviceIds.subtracting(oldDeviceIds)
            for device in devices where addedDeviceIds.contains(device.id) {
                self.delegate?.deviceManager(self, didDiscoverDevice: device)
            }
            
            let removedDeviceIds = oldDeviceIds.subtracting(newDeviceIds)
            for deviceId in removedDeviceIds {
                self.delegate?.deviceManager(self, didLoseDevice: deviceId)
            }
            
            self.discoveredDevices = devices
        }
    }
}

class TransferManager {
    
    weak var delegate: TransferManagerDelegate?
    private let logger = XPCLogger.shared
    private let goBridge = GoServiceBridge.shared
    private var activeTransfers: [CrossShareTransfer] = []
    
    init(delegate: TransferManagerDelegate?) {
        self.delegate = delegate
    }
    
    func sendFile(filePath: String, toDevice deviceId: String, completion: @escaping (Bool, String?) -> Void) {
        logger.info("Sending file: \(filePath) to device: \(deviceId)")
        
        guard FileManager.default.fileExists(atPath: filePath) else {
            completion(false, "File not found")
            return
        }
        
        goBridge.sendFile(filePath: filePath, toDevice: deviceId) { [weak self] transferId, error in
            if let transferId = transferId {
                let fileName = URL(fileURLWithPath: filePath).lastPathComponent
                let fileSize: Int64
                do {
                    let attributes = try FileManager.default.attributesOfItem(atPath: filePath)
                    fileSize = attributes[.size] as? Int64 ?? 0
                } catch {
                    fileSize = 0
                }
                
                let transfer = CrossShareTransfer(
                    id: transferId,
                    fileName: fileName,
                    fileSize: fileSize,
                    fromDevice: "local",
                    toDevice: deviceId,
                    progress: 0.0,
                    status: .transferring,
                    startTime: Date(),
                    estimatedTime: nil
                )
                
                self?.activeTransfers.append(transfer)
                self?.logger.info("File transfer started: \(transferId)")
                completion(true, nil)
            } else {
                self?.logger.error("Failed to send file: \(error ?? "Unknown error")")
                completion(false, error)
            }
        }
    }
    
    func sendFolder(folderPath: String, toDevice deviceId: String, completion: @escaping (Bool, String?) -> Void) {
        logger.info("Sending folder: \(folderPath) to device: \(deviceId)")
        
        guard FileManager.default.fileExists(atPath: folderPath) else {
            completion(false, "Folder not found")
            return
        }
        
        goBridge.sendFolder(folderPath: folderPath, toDevice: deviceId) { [weak self] transferId, error in
            if let transferId = transferId {
                let folderName = URL(fileURLWithPath: folderPath).lastPathComponent
                
                let transfer = CrossShareTransfer(
                    id: transferId,
                    fileName: folderName,
                    fileSize: 0,
                    fromDevice: "local",
                    toDevice: deviceId,
                    progress: 0.0,
                    status: .transferring,
                    startTime: Date(),
                    estimatedTime: nil
                )
                
                self?.activeTransfers.append(transfer)
                self?.logger.info("Folder transfer started: \(transferId)")
                completion(true, nil)
            } else {
                self?.logger.error("Failed to send folder: \(error ?? "Unknown error")")
                completion(false, error)
            }
        }
    }
    
    func cancelTransfer(transferId: String, completion: @escaping (Bool) -> Void) {
        logger.info("Cancelling transfer: \(transferId)")
        
        if let index = activeTransfers.firstIndex(where: { $0.id == transferId }) {
            activeTransfers.remove(at: index)
        }
        
        completion(true)
    }
    
    func getActiveTransfers() -> [CrossShareTransfer] {
        return activeTransfers
    }
    
    func isHealthy() -> Bool {
        return true
    }
    
    private func simulateTransfer(transfer: CrossShareTransfer) {
        let totalSteps = 10
        var currentStep = 0
        
        Timer.scheduledTimer(withTimeInterval: 0.5, repeats: true) { [weak self] timer in
            currentStep += 1
            let progress = Double(currentStep) / Double(totalSteps)
            
            let progressInfo: [String: Any] = [
                "transferId": transfer.id,
                "fileName": transfer.fileName,
                "progress": progress,
                "speed": "1.2 MB/s"
            ]
            
            self?.delegate?.transferManager(self!, didUpdateProgress: progressInfo)
            
            if currentStep >= totalSteps {
                timer.invalidate()
                
                let result: [String: Any] = [
                    "transferId": transfer.id,
                    "fileName": transfer.fileName,
                    "success": true
                ]
                
                self?.delegate?.transferManager(self!, didCompleteTransfer: result)
                
                if let index = self?.activeTransfers.firstIndex(where: { $0.id == transfer.id }) {
                    self?.activeTransfers.remove(at: index)
                }
            }
        }
    }
}

class ClipboardManager {
    
    weak var delegate: ClipboardManagerDelegate?
    private let logger = XPCLogger.shared
    private var isMonitoring = false
    private var monitoringTimer: Timer?
    private var lastClipboardChangeCount = NSPasteboard.general.changeCount
    
    init(delegate: ClipboardManagerDelegate?) {
        self.delegate = delegate
    }
    
    func startMonitoring(completion: @escaping (Bool) -> Void) {
        guard !isMonitoring else {
            completion(true)
            return
        }
        
        logger.info("Starting clipboard monitoring")
        isMonitoring = true
        
        monitoringTimer = Timer.scheduledTimer(withTimeInterval: 1.0, repeats: true) { [weak self] _ in
            self?.checkClipboardChanges()
        }
        
        completion(true)
    }
    
    func stopMonitoring(completion: @escaping (Bool) -> Void) {
        logger.info("Stopping clipboard monitoring")
        isMonitoring = false
        monitoringTimer?.invalidate()
        monitoringTimer = nil
        completion(true)
    }
    
    func syncToDevice(deviceId: String, completion: @escaping (Bool) -> Void) {
        logger.info("Syncing clipboard to device: \(deviceId)")
        
        let pasteboard = NSPasteboard.general
        
        if let string = pasteboard.string(forType: .string) {
            let clipboardData: [String: Any] = [
                "type": "text",
                "content": string,
                "timestamp": Date().timeIntervalSince1970
            ]
            
            delegate?.clipboardManager(self, didReceiveData: clipboardData)
        }
        
        completion(true)
    }
    
    func isHealthy() -> Bool {
        return true
    }
    
    private func checkClipboardChanges() {
        let currentChangeCount = NSPasteboard.general.changeCount
        
        if currentChangeCount != lastClipboardChangeCount {
            lastClipboardChangeCount = currentChangeCount
            
            if let string = NSPasteboard.general.string(forType: .string) {
                let clipboardData: [String: Any] = [
                    "type": "text",
                    "content": string,
                    "timestamp": Date().timeIntervalSince1970
                ]
                
                delegate?.clipboardManager(self, didReceiveData: clipboardData)
            }
        }
    }
}

class NetworkMonitor {
    
    weak var delegate: NetworkMonitorDelegate?
    private let logger = XPCLogger.shared
    private var monitor: NWPathMonitor?
    private let queue = DispatchQueue(label: "NetworkMonitor")
    
    init(delegate: NetworkMonitorDelegate?) {
        self.delegate = delegate
        setupNetworkMonitoring()
    }
    
    func getCurrentNetworkInfo() -> [String: Any] {
        var info: [String: Any] = [
            "timestamp": Date().timeIntervalSince1970
        ]
        
        if let ipAddress = getLocalIPAddress() {
            info["ipAddress"] = ipAddress
        }
        
        info["isConnected"] = isNetworkAvailable()
        
        return info
    }
    
    func getLocalIPAddress() -> String? {
        var address: String?
        var ifaddr: UnsafeMutablePointer<ifaddrs>?
        
        guard getifaddrs(&ifaddr) == 0 else { return nil }
        guard let firstAddr = ifaddr else { return nil }
        
        for ifptr in sequence(first: firstAddr, next: { $0.pointee.ifa_next }) {
            let interface = ifptr.pointee
            let addrFamily = interface.ifa_addr.pointee.sa_family
            
            if addrFamily == UInt8(AF_INET) || addrFamily == UInt8(AF_INET6) {
                let name = String(cString: interface.ifa_name)
                if name == "en0" || name == "en1" || name.starts(with: "wl") {
                    var hostname = [CChar](repeating: 0, count: Int(NI_MAXHOST))
                    getnameinfo(interface.ifa_addr,
                               socklen_t(interface.ifa_addr.pointee.sa_len),
                               &hostname,
                               socklen_t(hostname.count),
                               nil,
                               socklen_t(0),
                               NI_NUMERICHOST)
                    address = String(cString: hostname)
                    
                    if addrFamily == UInt8(AF_INET) {
                        break
                    }
                }
            }
        }
        
        freeifaddrs(ifaddr)
        return address
    }
    
    func isPortAvailable(port: Int) -> Bool {
        let socket = Darwin.socket(AF_INET, SOCK_STREAM, 0)
        guard socket != -1 else { return false }
        
        defer { close(socket) }
        
        var addr = sockaddr_in()
        addr.sin_family = sa_family_t(AF_INET)
        addr.sin_addr.s_addr = inet_addr("127.0.0.1")
        addr.sin_port = in_port_t(port).bigEndian
        
        let result = withUnsafePointer(to: &addr) {
            $0.withMemoryRebound(to: sockaddr.self, capacity: 1) {
                bind(socket, $0, socklen_t(MemoryLayout<sockaddr_in>.size))
            }
        }
        
        return result == 0
    }
    
    func isHealthy() -> Bool {
        return monitor != nil
    }
    
    private func setupNetworkMonitoring() {
        monitor = NWPathMonitor()
        monitor?.pathUpdateHandler = { [weak self] path in
            let networkInfo: [String: Any] = [
                "isConnected": path.status == .satisfied,
                "isExpensive": path.isExpensive,
                "isConstrained": path.isConstrained,
                "timestamp": Date().timeIntervalSince1970
            ]
            
            self?.delegate?.networkMonitor(self!, didChangeNetworkStatus: networkInfo)
        }
        
        monitor?.start(queue: queue)
    }
    
    private func isNetworkAvailable() -> Bool {
        return monitor?.currentPath.status == .satisfied
    }
}
