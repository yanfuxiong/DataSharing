//
//  HelperClient.swift
//  CrossShare
//
//  Created by TS on 2025/9/11.
//  Direct client connection to Helper
//

import Foundation
import Cocoa

class HelperClient: NSObject {
    
    static let shared = HelperClient()
    
    private var connection: NSXPCConnection?
    private let queue = DispatchQueue(label: "com.crossshare.helper.client", qos: .userInitiated)
    private let helperServiceName = "com.realtek.crossshare.macos.helper"
    
    private var eventHandlers: XPCEventHandlers = XPCEventHandlers()
    
    private var isConnected = false
    private var isConnecting = false
    
    private var reconnectTimer: Timer?
    private var reconnectAttempts = 0
    private let maxReconnectAttempts = 5

    private var deviceList: [CrossShareDevice] = []
    private let deviceListLock = NSLock()
    
    // UserDefaults key for system info storage
    private let systemInfoKey = "CrossShareSystemInfo"

    // Queue for pending requests when not connected
    private var pendingRequests: [() -> Void] = []
    private let pendingRequestsLock = NSLock()
    
    private override init() {
        super.init()
    }
    
    deinit {
        disconnect()
    }
    
    private func setupConnection(completion: @escaping (Bool) -> Void) {
        if connection != nil {
            logger.info("Cleaning up existing Helper connection...")
            connection?.invalidate()
            connection = nil
            Thread.sleep(forTimeInterval: 0.2)
        }
        
        logger.info("=== Helper Connection Setup ===")
        logger.info("Helper Service Name: \(helperServiceName)")
        logger.info("Checking if Helper is running...")
        
        let helperRunning = HelperCommunication.shared.isHelperRunning
        logger.info("Helper running status: \(helperRunning)")
        
        if !helperRunning {
            logger.info("Helper is not running, attempting to start it...")
            #if DEBUG
            HelperCommunication.shared.launchHelper { success in
                if success {
                    logger.info("Helper launched successfully, waiting for XPC service...")
                    DispatchQueue.main.asyncAfter(deadline: .now() + 2.0) {
                        self.createConnectionWithRetry(completion: completion, attempt: 1, maxAttempts: 3)
                    }
                } else {
                    logger.info("Failed to launch Helper")
                    completion(false)
                }
            }
            #else
            completion(false)
            #endif
        } else {
            createConnectionWithRetry(completion: completion, attempt: 1, maxAttempts: 3)
        }
    }
    
    private func createConnectionWithRetry(completion: @escaping (Bool) -> Void, attempt: Int, maxAttempts: Int) {
        logger.info("Attempting to create XPC connection (attempt \(attempt)/\(maxAttempts))")
        
        createConnection { success in
            if success {
                logger.info("XPC connection established successfully")
                completion(true)
            } else if attempt < maxAttempts {
                logger.info("XPC connection failed, retrying in 1 second...")
                DispatchQueue.main.asyncAfter(deadline: .now() + 1.0) {
                    self.createConnectionWithRetry(completion: completion, attempt: attempt + 1, maxAttempts: maxAttempts)
                }
            } else {
                logger.info("Failed to establish XPC connection after \(maxAttempts) attempts")
                completion(false)
            }
        }
    }
    
    private func createConnection(completion: @escaping (Bool) -> Void) {
        logger.info("Creating NSXPCConnection to Mach Service: \(helperServiceName)")
        connection = NSXPCConnection(machServiceName: helperServiceName)
        
        guard let conn = connection else {
            logger.info("CRITICAL: Failed to create NSXPCConnection!")
            completion(false)
            return
        }
        
        logger.info("NSXPCConnection object created")
        conn.remoteObjectInterface = NSXPCInterface(with: CrossShareHelperXPCProtocol.self)
        logger.info("   - Set remoteObjectInterface: CrossShareHelperXPCProtocol")
        
        conn.exportedInterface = NSXPCInterface(with: CrossShareHelperXPCDelegate.self)
        logger.info("   - Set exportedInterface: CrossShareHelperXPCDelegate")
        
        conn.exportedObject = self
        logger.info("   - Set exportedObject: self")
        
        conn.invalidationHandler = { [weak self] in
            DispatchQueue.main.async {
                logger.info("Helper connection invalidated")
                self?.handleConnectionLost()
            }
        }
        logger.info("   - Set invalidationHandler")
        
        conn.interruptionHandler = { [weak self] in
            DispatchQueue.main.async {
                logger.info("Helper connection interrupted")
                self?.handleConnectionInterrupted()
            }
        }
        logger.info("   - Set interruptionHandler")
        
        conn.resume()
        logger.info("XPC connection.resume() called, waiting for Helper to be ready...")
        // Wait longer (1 second) to ensure Helper's XPC service is fully ready
        DispatchQueue.main.asyncAfter(deadline: .now() + 1.0) { [weak self] in
            logger.info("Testing XPC connection to Helper...")
            self?.testConnection { success in
                if success {
                    logger.info("Helper connection established successfully")
                    logger.info("=============================")
                    self?.isConnected = true
                    completion(true)
                } else {
                    logger.info("Helper connection test failed")
                    self?.connection?.invalidate()
                    self?.connection = nil
                    completion(false)
                }
            }
        }
    }
    
    private func testConnection(completion: @escaping (Bool) -> Void) {
        guard let connection = connection else {
            logger.info("Test connection failed: no connection")
            completion(false)
            return
        }
        
        var hasCompleted = false
        let completionLock = NSLock()
        
        // Timeout after 3 seconds
        DispatchQueue.main.asyncAfter(deadline: .now() + 3.0) {
            completionLock.lock()
            if !hasCompleted {
                hasCompleted = true
                completionLock.unlock()
                logger.info("Helper connection test timed out after 3 seconds")
                completion(false)
            } else {
                completionLock.unlock()
            }
        }
        
        let proxy = connection.remoteObjectProxyWithErrorHandler { error in
            logger.info("Failed to get Helper proxy: \(error)")
            completionLock.lock()
            if !hasCompleted {
                hasCompleted = true
                completionLock.unlock()
                completion(false)
            } else {
                completionLock.unlock()
            }
        }
        
        guard let helperProxy = proxy as? CrossShareHelperXPCProtocol else {
            logger.info("Failed to cast proxy to CrossShareHelperXPCProtocol")
            completionLock.lock()
            if !hasCompleted {
                hasCompleted = true
                completionLock.unlock()
                completion(false)
            } else {
                completionLock.unlock()
            }
            return
        }
        
        logger.info("Calling Helper test method (getServiceStatus)...")
        helperProxy.getServiceStatus { isRunning, info in
            logger.info("Helper test call response received: service running: \(isRunning)")
            completionLock.lock()
            if !hasCompleted {
                hasCompleted = true
                completionLock.unlock()
                completion(true)
            } else {
                completionLock.unlock()
                logger.info("Helper test completed but response came too late")
            }
        }
    }
    
    func connect(completion: @escaping (Bool, String?) -> Void) {
        guard !isConnected && !isConnecting else {
            completion(isConnected, isConnected ? nil : "Already connecting")
            return
        }

        isConnecting = true

        queue.async { [weak self] in
            self?.setupConnection { success in
                DispatchQueue.main.async {
                    self?.isConnecting = false
                    if success {
                        self?.isConnected = true
                        self?.setupCallbacks()
                        logger.info("Helper client connected successfully")

                        self?.processPendingRequests()

                        completion(true, nil)
                    } else {
                        logger.info("Failed to connect to Helper")
                        completion(false, "Failed to establish Helper connection")
                    }
                }
            }
        }
    }
    
    private func setupCallbacks() {
        logger.info("Callback delegate already configured via XPC connection setup")
        if eventHandlers.onAuthRequested == nil {
            eventHandlers.onAuthRequested = { index in
                logger.info("Default handler: Auth requested for index: \(index)")
            }
        }
    }
    
    func disconnect() {
        queue.async { [weak self] in
            self?.isConnected = false
            self?.connection?.invalidate()
            self?.connection = nil
            self?.reconnectTimer?.invalidate()
            self?.reconnectTimer = nil
        }
    }
    
    private func handleConnectionLost() {
        isConnected = false
        connection = nil
        
        eventHandlers.onConnectionLost?()
        
        if NSApp.isRunning {
            attemptReconnect()
        }
    }
    
    private func handleConnectionInterrupted() {
        eventHandlers.onConnectionInterrupted?()
    
        if NSApp.isRunning {
            attemptReconnect()
        }
    }
    
    private func attemptReconnect() {
        guard reconnectAttempts < maxReconnectAttempts else {
            logger.info("Max reconnection attempts reached for Helper")
            return
        }
        
        reconnectAttempts += 1
        let delay = Double(reconnectAttempts) * 2.0
        
        logger.info("Attempting to reconnect to Helper in \(delay) seconds (attempt \(reconnectAttempts)/\(maxReconnectAttempts))")
        
        reconnectTimer?.invalidate()
        reconnectTimer = Timer.scheduledTimer(withTimeInterval: delay, repeats: false) { [weak self] _ in
            self?.connect { success, error in
                if success {
                    logger.info("Reconnected to Helper successfully")
                    self?.reconnectAttempts = 0
                } else {
                    logger.info("Failed to reconnect to Helper: \(error ?? "Unknown error")")
                }
            }
        }
    }
    
    func startGoService(config: [String: Any], completion: @escaping (Bool, String?) -> Void) {
        executeWhenConnected("Start Go Service") { [weak self] in
            guard let proxy = self?.getRemoteProxy(completion: { completion(false, "No connection") }) else { return }
            proxy.initializeGoService(config: config, completion: completion)
        }
    }
    
    func getServiceStatus(completion: @escaping (Bool, [String: Any]?) -> Void) {
        executeWhenConnected("Get Service Status") { [weak self] in
            guard let proxy = self?.getRemoteProxy(completion: { completion(false, nil) }) else { return }
            proxy.getServiceStatus(completion: completion)
        }
    }
    
    func updateCount(_ count: Int, completion: @escaping (Int) -> Void) {
        executeWhenConnected("Update Count") { [weak self] in
            guard let proxy = self?.getRemoteProxy(completion: { completion(0) }) else { return }
            proxy.updateCount(count, completion: completion)
        }
    }

    func setDIASID(_ diasID: String, completion: @escaping (Bool, String?) -> Void) {
        executeWhenConnected("Set DIAS ID") { [weak self] in
            guard let proxy = self?.getRemoteProxy(completion: { completion(false, "unKnow Error") }) else { return }
            proxy.setDIASID(diasID, completion: completion)
        }
    }
    
    func setExtractDIAS(completion: @escaping (Bool, String?) -> Void) {
        executeWhenConnected("Set Extract DIAS") { [weak self] in
            guard let proxy = self?.getRemoteProxy(completion: { completion(false, "unKnow Error") }) else { return }
            proxy.setExtractDIAS(completion: completion)
        }
    }
    
    func updateDisplayMapping(mac: String, displayID: CGDirectDisplayID, completion: @escaping (Bool) -> Void) {
        executeWhenConnected("Update Display Mapping") { [weak self] in
            guard let proxy = self?.getRemoteProxy(completion: { completion(false) }) else { return }
            proxy.updateDisplayMapping(mac: mac, displayID: UInt32(displayID), completion: completion)
        }
    }
    
    func rescanDisplays(completion: @escaping (Bool) -> Void) {
        executeWhenConnected("Rescan Displays") { [weak self] in
            guard let proxy = self?.getRemoteProxy(completion: { completion(false) }) else { return }
            proxy.rescanDisplays(completion: completion)
        }
    }

    func getDeviceListFromHelper(completion: @escaping ([[String: Any]]) -> Void) {
        logger.info("Requesting device list from Helper...")

        executeWhenConnected("Get Device List") { [weak self] in
            guard let proxy = self?.getRemoteProxy(completion: { completion([]) }) else { return }
            proxy.getDeviceList(completion: completion)
            logger.info("Device list request sent to Helper")
        }
    }

    func sendMultiFilesDropRequest(multiFilesData: String, completion: @escaping (Bool, String?) -> Void) {
        executeWhenConnected("Send Multi-Files") { [weak self] in
            guard let proxy = self?.getRemoteProxy(completion: { completion(false, "No connection") }) else { return }
            proxy.sendMultiFilesDropRequest(multiFilesData: multiFilesData, completion: completion)
        }
    }
    
    func setCancelFileTransfer(ipPort: String, clientID: String, timeStamp: UInt64, completion: @escaping (Bool, String?) -> Void) {
        executeWhenConnected("Cancel File Transfer") { [weak self] in
            guard let proxy = self?.getRemoteProxy(completion: { completion(false, "No connection") }) else { return }
            proxy.setCancelFileTransfer(ipPort: ipPort, clientID: clientID, timeStamp: timeStamp, completion: completion)
        }
    }
    
    func requestUpdateDownloadPath(downloadPath: String, completion: @escaping (Bool, String?) -> Void) {
        executeWhenConnected("Update Download Path") { [weak self] in
            guard let proxy = self?.getRemoteProxy(completion: { completion(false, "No connection") }) else { return }
            proxy.requestUpdateDownloadPath(downloadPath: downloadPath, completion: completion)
        }
    }

    func setDragFileListRequest(multiFilesData: String, timestamp: UInt64, width: UInt16, height: UInt16, posX: Int16, posY: Int16, completion: @escaping (Bool, String?) -> Void) {
        executeWhenConnected("Drag Multi-Files") { [weak self] in
            guard let proxy = self?.getRemoteProxy(completion: { completion(false, "No connection") }) else { return }
            proxy.setDragFileListRequest(multiFilesData: multiFilesData, timestamp: timestamp, width: width, height: height, posX: posX, posY: posY, completion: completion)
        }
    }

    func sendTextToRemote(_ text: String, completion: @escaping (Bool, String?) -> Void) {
        executeWhenConnected("Send Text to Remote") { [weak self] in
            guard let proxy = self?.getRemoteProxy(completion: { completion(false, "No connection") }) else { return }
            proxy.sendTextToRemote(text, completion: completion)
        }
    }

    func sendImageToRemote(_ imageData: Data, completion: @escaping (Bool, String?) -> Void) {
        executeWhenConnected("Send Image to Remote") { [weak self] in
            guard let proxy = self?.getRemoteProxy(completion: { completion(false, "No connection") }) else { return }
            proxy.sendImageToRemote(imageData, completion: completion)
        }
    }

    func startClipboardMonitoring(completion: @escaping (Bool) -> Void) {
        executeWhenConnected("Start Clipboard Monitoring") { [weak self] in
            guard let proxy = self?.getRemoteProxy(completion: { completion(false) }) else { return }
            proxy.startClipboardMonitoring(completion: completion)
        }
    }

    func stopClipboardMonitoring(completion: @escaping (Bool) -> Void) {
        executeWhenConnected("Stop Clipboard Monitoring") { [weak self] in
            guard let proxy = self?.getRemoteProxy(completion: { completion(false) }) else { return }
            proxy.stopClipboardMonitoring(completion: completion)
        }
    }

    private func executeWhenConnected(_ description: String = "Request", action: @escaping () -> Void) {
        if isConnected {
            action()
            return
        }

        pendingRequestsLock.lock()
        logger.info("Helper not connected, queuing: \(description)")
        pendingRequests.append(action)
        let shouldConnect = !isConnecting
        pendingRequestsLock.unlock()

        if shouldConnect {
            logger.info("Attempting to connect for: \(description)")
            connect { success, error in
                if !success {
                    logger.info("Failed to connect for queued request: \(description), error: \(error ?? "Unknown")")
                    self.pendingRequestsLock.lock()
                    self.pendingRequests.removeAll()
                    self.pendingRequestsLock.unlock()
                }
            }
        }
    }

    private func processPendingRequests() {
        pendingRequestsLock.lock()
        guard !pendingRequests.isEmpty else {
            pendingRequestsLock.unlock()
            return
        }

        logger.info("Processing \(pendingRequests.count) pending requests")
        let requests = pendingRequests
        pendingRequests.removeAll()
        pendingRequestsLock.unlock()

        DispatchQueue.main.async {
            for request in requests {
                request()
            }
        }
    }


    func getCurrentDeviceList() -> [CrossShareDevice] {
        deviceListLock.lock()
        defer { deviceListLock.unlock() }
        return deviceList
    }

    func clearDeviceList() {
        deviceListLock.lock()
        defer { deviceListLock.unlock() }

        let wasEmpty = deviceList.isEmpty
        deviceList.removeAll()

        if !wasEmpty {
            DispatchQueue.main.async {
                NotificationCenter.default.post(
                    name: .deviceDataReceived,
                    object: [],
                    userInfo: [
                        "deviceList": [],
                        "deviceCount": 0,
                        "action": "cleared"
                    ]
                )
            }
            logger.info("Helper: Cleared device list")
        }
    }

    func removeDevice(byId deviceId: String) {
        deviceListLock.lock()
        defer { deviceListLock.unlock() }

        if let index = deviceList.firstIndex(where: { $0.id == deviceId }) {
            let removedDevice = deviceList.remove(at: index)
            let devices = deviceList

            DispatchQueue.main.async {
                NotificationCenter.default.post(
                    name: .deviceDataReceived,
                    object: devices,
                    userInfo: [
                        "deviceList": devices,
                        "removedDevice": removedDevice,
                        "deviceCount": devices.count,
                        "action": "removed"
                    ]
                )
            }
            logger.info("Helper: Removed device with ID: \(deviceId), remaining count: \(devices.count)")
        }
    }

    func getDevice(byId deviceId: String) -> CrossShareDevice? {
        deviceListLock.lock()
        defer { deviceListLock.unlock() }
        return deviceList.first(where: { $0.id == deviceId })
    }
    
    func getCurrentSystemInfo() -> (ipInfo: String, verInfo: String)? {
        let defaults = UserDefaults.standard
        guard let systemInfoDict = defaults.dictionary(forKey: systemInfoKey),
              let ipInfo = systemInfoDict["ipInfo"] as? String,
              let verInfo = systemInfoDict["verInfo"] as? String else {
            return nil
        }
        return (ipInfo: ipInfo, verInfo: verInfo)
    }
    
    func getCurrentSystemInfoDict() -> [String: Any]? {
        let defaults = UserDefaults.standard
        return defaults.dictionary(forKey: systemInfoKey)
    }

    private func getRemoteProxy<T>(completion: @escaping (T) -> Void) -> CrossShareHelperXPCProtocol? {
        guard let connection = connection, isConnected else {
            logger.info("No active Helper connection")
            return nil
        }
        
        let proxy = connection.remoteObjectProxyWithErrorHandler { error in
            logger.info("Helper remote proxy error: \(error)")
            DispatchQueue.main.async {
                self.handleConnectionLost()
            }
        }
        
        return proxy as? CrossShareHelperXPCProtocol
    }
}

extension HelperClient: CrossShareHelperXPCDelegate {
    
    func didUpdateCount(_ newCount: Int) {
        DispatchQueue.main.async {
            logger.info("Helper: Count updated - \(newCount)")
            self.eventHandlers.onCountUpdated?(newCount)
        }
    }
    
    func didReceiveAuthRequest(index: UInt32) {
        DispatchQueue.main.async {
            logger.info("Helper: Received auth request for index \(index)")
            self.eventHandlers.onAuthRequested?(index)
        }
    }
    
    func didReceiveDeviceData(deviceData: [String: Any]) {
        logger.info("Helper: Received device data - \(deviceData)")
        if let deviceList = deviceData["deviceList"] as? [[String: Any]] {
                self.deviceListLock.lock()
                self.deviceList = deviceList.compactMap { CrossShareDevice(from: $0) }
                logger.info("Helper: Updated device list with \(self.deviceList.count) devices")
                let devices = Array(self.deviceList)
                self.deviceListLock.unlock()

                self.eventHandlers.onDeviceDataReceived?(deviceData)

                NotificationCenter.default.post(
                    name: .deviceDataReceived,
                    object: devices,
                    userInfo: [
                        "deviceList": devices,
                        "deviceCount": devices.count
                    ]
                )
                logger.info("Helper: Total devices in list: \(devices.count)")
            }
    }
    
    
    func didReceiveFilesData(_ userInfo: [String: Any]) {
        DispatchQueue.main.async {
            NotificationCenter.default.post(
                name: .fileTransferSessionStarted,
                object: userInfo["session"],
                userInfo: userInfo
            )
        }
    }
    
    func didReceiveTransferFilesDataUpdate(_ userInfo: [String: Any]) {
        let notificationName: Notification.Name
        let isCompleted = userInfo["isCompleted"] as? Bool ?? false
        if isCompleted {
            notificationName = .fileTransferSessionCompleted
        } else {
            notificationName = .fileTransferSessionUpdated
        }
        DispatchQueue.main.async {
            NotificationCenter.default.post(
                name: notificationName,
                object: userInfo["session"],
                userInfo: userInfo
            )
        }
    }
    
    func didReceiveDIASStatus(_ status: Int){
        logger.info("gui ReceiveDIASStatus:\(status)")
        DispatchQueue.main.async {
            NotificationCenter.default.post(
                name: .deviceDiasStatusNotification,
                object: [],
                userInfo: [
                    "diasStatus": status,
                ]
            )
        }
    }
    
    func didReceiveErrorEvent(_ errorInfo: [String: Any]) {
        logger.info("GUI HelperClient received error event: \(errorInfo)")
        DispatchQueue.main.async {
            NotificationCenter.default.post(
                name: .didReceiveErrorEventNotification,
                object: nil,
                userInfo: errorInfo
            )
        }
    }
    
    func didReceiveSystemInfoUpdate(_ systemInfo: [String: Any]) {
        logger.info("GUI HelperClient received system info update systemInfo:\(systemInfo)")
        // save systemInfo
        let defaults = UserDefaults.standard
        defaults.set(systemInfo, forKey: systemInfoKey)
        defaults.synchronize()
    }
    
    func didReceiveThemeInfoUpdate(_ themeInfo: [String: Any]) {
        logger.info("GUI HelperClient received theme info update: \(themeInfo)")
    }

    func didDetectScreenCountChange(change: String, currentCount: Int, previousCount: Int) {
        DispatchQueue.main.async {
            logger.info("GUI: Screen count changed - \(change) (\(previousCount) -> \(currentCount))")
            if change == "decreased" {
                self.deviceListLock.lock()
                let wasEmpty = self.deviceList.isEmpty
                self.deviceList.removeAll()
                self.deviceListLock.unlock()

                if !wasEmpty {
                    NotificationCenter.default.post(
                        name: .deviceDataReceived,
                        object: [],
                        userInfo: [
                            "deviceList": [],
                            "deviceCount": 0,
                            "action": "screen_decreased_cleared",
                            "reason": "Screen count decreased, Go service communication interrupted"
                        ]
                    )
                }
            } else if change == "increased" {
                logger.info("GUI: Screen count increased, new displays detected")
            }

            NotificationCenter.default.post(
                name: .screenCountChanged,
                object: nil,
                userInfo: [
                    "change": change,
                    "currentCount": currentCount,
                    "previousCount": previousCount
                ]
            )
        }
    }

    func didReceiveFileTransferUpdate(_ sessionInfo: [String: Any]) {
        DispatchQueue.main.async {
            logger.info("GUI: Received file transfer session update - \(sessionInfo)")

            let notificationName: Notification.Name
            if let isCompleted = sessionInfo["isCompleted"] as? Bool, isCompleted {
                notificationName = .fileTransferSessionCompleted
            } else if let progress = sessionInfo["progress"] as? Double, progress == 0.0 {
                notificationName = .fileTransferSessionStarted
            } else {
                notificationName = .fileTransferSessionUpdated
            }

            NotificationCenter.default.post(
                name: notificationName,
                object: sessionInfo,
                userInfo: sessionInfo
            )

            if let sessionId = sessionInfo["sessionId"] as? String,
               let currentFileName = sessionInfo["currentFileName"] as? String,
               let progress = sessionInfo["totalProgress"] as? Double,
               let receivedCount = sessionInfo["receivedFileCount"] as? UInt32,
               let totalCount = sessionInfo["totalFileCount"] as? UInt32 {

                logger.info("File Transfer [\(sessionId)]: \(currentFileName) (\(receivedCount)/\(totalCount)) - \(String(format: "%.1f", progress * 100))%")
            }
        }
    }
}

extension HelperClient {
    var onCountUpdated: ((Int) -> Void)? {
        get { eventHandlers.onCountUpdated }
        set { eventHandlers.onCountUpdated = newValue }
    }
    
    var onConnected: (() -> Void)? {
        get { eventHandlers.onConnected }
        set { eventHandlers.onConnected = newValue }
    }
    
    var onConnectionLost: (() -> Void)? {
        get { eventHandlers.onConnectionLost }
        set { eventHandlers.onConnectionLost = newValue }
    }
    
    var onConnectionInterrupted: (() -> Void)? {
        get { eventHandlers.onConnectionInterrupted }
        set { eventHandlers.onConnectionInterrupted = newValue }
    }
    
    var onReconnected: (() -> Void)? {
        get { eventHandlers.onReconnected }
        set { eventHandlers.onReconnected = newValue }
    }
    
    var onReconnectionFailed: (() -> Void)? {
        get { eventHandlers.onReconnectionFailed }
        set { eventHandlers.onReconnectionFailed = newValue }
    }
    
    var onAuthRequested: ((UInt32) -> Void)? {
        get { eventHandlers.onAuthRequested }
        set { eventHandlers.onAuthRequested = newValue }
    }
    
    var onDeviceDataReceived: (([String: Any]) -> Void)? {
        get { eventHandlers.onDeviceDataReceived }
        set { eventHandlers.onDeviceDataReceived = newValue }
    }
}
