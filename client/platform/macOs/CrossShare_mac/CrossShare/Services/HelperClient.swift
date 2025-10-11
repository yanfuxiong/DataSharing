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
            print("Cleaning up existing Helper connection...")
            connection?.invalidate()
            connection = nil
            Thread.sleep(forTimeInterval: 0.2)
        }
        
        print("=== Helper Connection Setup ===")
        print("Helper Service Name: \(helperServiceName)")
        print("Checking if Helper is running...")
        
        let helperRunning = HelperCommunication.shared.isHelperRunning
        print("Helper running status: \(helperRunning)")
        
        if !helperRunning {
            print("Helper is not running, attempting to start it...")
            #if DEBUG
            HelperCommunication.shared.launchHelper { success in
                if success {
                    print("Helper launched successfully")
                    DispatchQueue.main.asyncAfter(deadline: .now() + 2.0) {
                        self.createConnection(completion: completion)
                    }
                } else {
                    print("Failed to launch Helper")
                    completion(false)
                }
            }
            #else
            completion(false)
            #endif
        } else {
            createConnection(completion: completion)
        }
    }
    
    private func createConnection(completion: @escaping (Bool) -> Void) {
        connection = NSXPCConnection(machServiceName: helperServiceName)
        connection?.remoteObjectInterface = NSXPCInterface(with: CrossShareHelperXPCProtocol.self)
        connection?.exportedInterface = NSXPCInterface(with: CrossShareHelperXPCDelegate.self)
        connection?.exportedObject = self
        connection?.invalidationHandler = { [weak self] in
            DispatchQueue.main.async {
                print("Helper connection invalidated")
                self?.handleConnectionLost()
            }
        }
        connection?.interruptionHandler = { [weak self] in
            DispatchQueue.main.async {
                print("Helper connection interrupted")
                self?.handleConnectionInterrupted()
            }
        }
        connection?.resume()
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.1) { [weak self] in
            self?.testConnection { success in
                if success {
                    print("Helper connection established successfully")
                    print("=============================")
                    self?.isConnected = true
                    completion(true)
                } else {
                    print("Helper connection test failed")
                    self?.connection?.invalidate()
                    self?.connection = nil
                    completion(false)
                }
            }
        }
    }
    
    private func testConnection(completion: @escaping (Bool) -> Void) {
        guard let connection = connection else {
            completion(false)
            return
        }
        
        let proxy = connection.remoteObjectProxyWithErrorHandler { error in
            print("Failed to get Helper proxy: \(error)")
            completion(false)
        }
        
        guard let helperProxy = proxy as? CrossShareHelperXPCProtocol else {
            completion(false)
            return
        }
        
        helperProxy.getServiceStatus { isRunning, info in
            print("Helper test call successful, service running: \(isRunning)")
            completion(true)
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
                        print("Helper client connected successfully")

                        self?.processPendingRequests()

                        completion(true, nil)
                    } else {
                        print("Failed to connect to Helper")
                        completion(false, "Failed to establish Helper connection")
                    }
                }
            }
        }
    }
    
    private func setupCallbacks() {
        print("Callback delegate already configured via XPC connection setup")
        if eventHandlers.onAuthRequested == nil {
            eventHandlers.onAuthRequested = { index in
                print("Default handler: Auth requested for index: \(index)")
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
            print("Max reconnection attempts reached for Helper")
            return
        }
        
        reconnectAttempts += 1
        let delay = Double(reconnectAttempts) * 2.0
        
        print("Attempting to reconnect to Helper in \(delay) seconds (attempt \(reconnectAttempts)/\(maxReconnectAttempts))")
        
        reconnectTimer?.invalidate()
        reconnectTimer = Timer.scheduledTimer(withTimeInterval: delay, repeats: false) { [weak self] _ in
            self?.connect { success, error in
                if success {
                    print("Reconnected to Helper successfully")
                    self?.reconnectAttempts = 0
                } else {
                    print("Failed to reconnect to Helper: \(error ?? "Unknown error")")
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
        print("Requesting device list from Helper...")

        executeWhenConnected("Get Device List") { [weak self] in
            guard let proxy = self?.getRemoteProxy(completion: { completion([]) }) else { return }
            proxy.getDeviceList(completion: completion)
            print("Device list request sent to Helper")
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

    private func executeWhenConnected(_ description: String = "Request", action: @escaping () -> Void) {
        if isConnected {
            action()
            return
        }

        pendingRequestsLock.lock()
        print("Helper not connected, queuing: \(description)")
        pendingRequests.append(action)
        let shouldConnect = !isConnecting
        pendingRequestsLock.unlock()

        if shouldConnect {
            print("Attempting to connect for: \(description)")
            connect { success, error in
                if !success {
                    print("Failed to connect for queued request: \(description), error: \(error ?? "Unknown")")
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

        print("Processing \(pendingRequests.count) pending requests")
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
            print("Helper: Cleared device list")
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
            print("Helper: Removed device with ID: \(deviceId), remaining count: \(devices.count)")
        }
    }

    func getDevice(byId deviceId: String) -> CrossShareDevice? {
        deviceListLock.lock()
        defer { deviceListLock.unlock() }
        return deviceList.first(where: { $0.id == deviceId })
    }

    private func getRemoteProxy<T>(completion: @escaping (T) -> Void) -> CrossShareHelperXPCProtocol? {
        guard let connection = connection, isConnected else {
            print("No active Helper connection")
            return nil
        }
        
        let proxy = connection.remoteObjectProxyWithErrorHandler { error in
            print("Helper remote proxy error: \(error)")
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
            print("Helper: Count updated - \(newCount)")
            self.eventHandlers.onCountUpdated?(newCount)
        }
    }
    
    func didReceiveAuthRequest(index: UInt32) {
        DispatchQueue.main.async {
            print("Helper: Received auth request for index \(index)")
            self.eventHandlers.onAuthRequested?(index)
        }
    }
    
    func didReceiveDeviceData(deviceData: [String: Any]) {
        print("Helper: Received device data - \(deviceData)")
        if let deviceList = deviceData["deviceList"] as? [[String: Any]] {
                self.deviceListLock.lock()
                defer { self.deviceListLock.unlock() }

                self.deviceList = deviceList.compactMap { CrossShareDevice(from: $0) }

                print("Helper: Updated device list with \(self.deviceList.count) devices")

                let devices = self.deviceList
                self.eventHandlers.onDeviceDataReceived?(deviceData)

                NotificationCenter.default.post(
                    name: .deviceDataReceived,
                    object: devices,
                    userInfo: [
                        "deviceList": devices,
                        "deviceCount": devices.count
                    ]
                )
                print("Helper: Total devices in list: \(devices.count)")
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
        print("gui ReceiveDIASStatus:\(status)")
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


    func didDetectScreenCountChange(change: String, currentCount: Int, previousCount: Int) {
        DispatchQueue.main.async {
            print("GUI: Screen count changed - \(change) (\(previousCount) -> \(currentCount))")
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
                print("GUI: Screen count increased, new displays detected")
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
            print("GUI: Received file transfer session update - \(sessionInfo)")

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

                print("File Transfer [\(sessionId)]: \(currentFileName) (\(receivedCount)/\(totalCount)) - \(String(format: "%.1f", progress * 100))%")
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
