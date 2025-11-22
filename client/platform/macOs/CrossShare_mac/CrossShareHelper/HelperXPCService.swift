//
//  HelperXPCService.swift
//  CrossShareHelper
//
//  Created by TS on 2025/9/11.
//

import Foundation
import AppKit

class HelperXPCService: NSObject, CrossShareHelperXPCProtocol, CrossShareHelperXPCDelegate {
    
    func updateCount(_ count: Int, completion: @escaping (Int) -> Void) {
        logger.log("Received count update from main app: \(count)", level: .info)
        
        let updatedCount = count + 10
        logger.log("Returning updated count to main app: \(updatedCount)", level: .info)
        
        completion(updatedCount)
        
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.1) { [weak self] in
            self?.notifyDelegates { delegate in
                delegate.didUpdateCount?(updatedCount)
            }
        }
    }
    
    func setDIASID(_ diasID: String, completion: @escaping (Bool, String?) -> Void) {
        logger.log("Setting DiasID: \(diasID)", level: .info)
        GoCallbackManager.shared.setCurrentDIASMac(diasID)
        goBridge.setDIASID(diasID: diasID) { [weak self] success, error in
            if success {
                self?.logger.log("DiasID set successfully to: \(diasID)", level: .info)
            } else {
                self?.logger.log("Failed to set DiasID: \(error ?? "Unknown error")", level: .error)
            }
            completion(success, error)
        }
    }
    
    func setExtractDIAS(completion: @escaping (Bool, String?) -> Void) {
        logger.log("Setting extract DIAS", level: .info)
        goBridge.setExtractDIAS { [weak self] success, error in
            if success {
                self?.logger.log("Extract DIAS set successfully", level: .info)
            } else {
                self?.logger.log("Failed to set extract DIAS: \(error ?? "Unknown error")", level: .error)
            }
            completion(success, error)
        }
    }

    private let logger = CSLogger.shared
    private let goBridge = GoServiceBridge.shared
    private var xpcConnections: [NSXPCConnection] = []
    private let connectionsLock = NSLock()
    private var screenMonitor: HelperScreenMonitor?

    private var deviceList: [CrossShareDevice] = []
    private let deviceListLock = NSLock()
    
    override init() {
        super.init()
        setupGoCallbacks()
        logger.log("HelperXPCService initialized", level: .info)

        HelperDisplayManager.shared.startMonitoring()

        screenMonitor = HelperScreenMonitor(xpcService: self)

        startClipboardMonitoringAutomatically()
    }

    private func startClipboardMonitoringAutomatically() {
        logger.log("Auto-starting clipboard monitoring in Helper", level: .info)
        ClipboardMonitor.shareInstance().setDelegate(self)
        ClipboardMonitor.shareInstance().startMonitoring()
        logger.log("Clipboard monitoring started automatically", level: .info)
    }
    
    private func setupGoCallbacks() {
        GoCallbackManager.shared.delegate = self
    }
    
    
    func initializeGoService(config: [String: Any], completion: @escaping (Bool, String?) -> Void) {
        logger.log("Initializing Go service with config: \(config)", level: .info)
        
        goBridge.startService(config: config) { [weak self] success, error in
            if success {
                self?.logger.log("Go service initialized successfully", level: .info)
            } else {
                self?.logger.log("Failed to initialize Go service: \(error ?? "Unknown error")", level: .error)
            }
            completion(success, error)
        }
    }
    
    func getGoServiceHealth(completion: @escaping (Bool, [String: Any]) -> Void) {
        goBridge.healthCheck { isHealthy, healthInfo in
            completion(isHealthy, healthInfo)
        }
    }
    
    func getServiceStatus(completion: @escaping (Bool, [String: Any]?) -> Void) {
        logger.log("Getting service status", level: .info)
        
        let isRunning = goBridge.isRunning
        
        logger.log("Service running state: \(isRunning)", level: .info)
        if isRunning {
            logger.log("Fetching service info", level: .info)
            goBridge.getServiceInfo { [weak self] serviceInfo in
                var status = serviceInfo ?? [:]
                status["isRunning"] = true
                status["helperPID"] = ProcessInfo.processInfo.processIdentifier
                self?.logger.log("Service is running with info: \(status)", level: .info)
                completion(true, status)
            }
        } else {
            logger.log("Service is not running", level: .info)
            completion(false, ["isRunning": false])
        }
    }
    
    func checkPortAvailability(port: Int, completion: @escaping (Bool) -> Void) {
        goBridge.checkPortAvailability(port: port) { available in
            completion(available)
        }
    }
    
    func addXPCConnection(_ connection: NSXPCConnection) {
        connectionsLock.lock()
        defer { connectionsLock.unlock() }
        xpcConnections.append(connection)
        logger.log("Added XPC connection, total connections: \(xpcConnections.count)", level: .info)
    }

    func removeXPCConnection(_ connection: NSXPCConnection) {
        connectionsLock.lock()
        defer { connectionsLock.unlock() }
        xpcConnections.removeAll { $0 === connection }
        logger.log("Removed XPC connection, remaining connections: \(xpcConnections.count)", level: .info)
    }
    
    func notifyDelegates(_ block: @escaping (CrossShareHelperXPCDelegate) -> Void) {
        connectionsLock.lock()
        let connections = xpcConnections
        connectionsLock.unlock()

        for connection in connections {
            if let delegate = connection.remoteObjectProxyWithErrorHandler({ error in
                self.logger.log("Error calling delegate: \(error)", level: .error)
            }) as? CrossShareHelperXPCDelegate {
                block(delegate)
            }
        }
    }
    
    func performHealthCheck(completion: @escaping ([String: Any]) -> Void) {
        logger.log("Performing health check", level: .info)
        
        var healthInfo: [String: Any] = [
            "timestamp": Date().timeIntervalSince1970,
            "helperPID": ProcessInfo.processInfo.processIdentifier,
            "goServiceRunning": goBridge.isRunning
        ]
        
        goBridge.healthCheck { isHealthy, goHealthInfo in
            healthInfo["goServiceHealthy"] = isHealthy
            healthInfo["goServiceInfo"] = goHealthInfo
            completion(healthInfo)
        }
    }
    
    func didReceiveAuthRequest(index: UInt32) {

        handleAuthenticationRequest(index: index)
    }
    
    func didReceiveDeviceData(deviceData: [String: Any]) {
        logger.log("Helper XPC Service received device data: \(deviceData)", level: .info)

        guard let deviceId = deviceData["ID"] as? String else {
            logger.log("Helper XPC Service: Missing device ID in data", level: .error)
            return
        }

        let status = deviceData["Status"] as? Int ?? 0
        logger.log("Helper XPC Service: Device \(deviceId) status: \(status)", level: .info)

        deviceListLock.lock()
        defer { deviceListLock.unlock() }

        var deviceListChanged = false

        if status == 0 {
            if let index = deviceList.firstIndex(where: { $0.id == deviceId }) {
                deviceList.remove(at: index)
                logger.log("Helper XPC Service: Removed offline device with ID: \(deviceId)", level: .info)
                deviceListChanged = true
            } else {
                logger.log("Helper XPC Service: Device \(deviceId) not found in list for removal", level: .warn)
            }
        } else if status == 1 {
            if let newDevice = CrossShareDevice(from: deviceData) {
                if let existingIndex = deviceList.firstIndex(where: { $0.id == newDevice.id }) {
                    deviceList[existingIndex] = newDevice
                    logger.log("Helper XPC Service: Updated existing device with ID: \(newDevice.id)", level: .info)
                    deviceListChanged = true
                } else {
                    deviceList.append(newDevice)
                    logger.log("Helper XPC Service: Added new device with ID: \(newDevice.id)", level: .info)
                    deviceListChanged = true
                }
            } else {
                logger.log("Helper XPC Service: Failed to parse device data into model", level: .error)
            }
        } else {
            logger.log("Helper XPC Service: Unknown device status: \(status) for device \(deviceId)", level: .warn)
        }

        if deviceListChanged {
            logger.log("Helper XPC Service: Total devices in list: \(deviceList.count)", level: .info)
            let deviceDictionaries = deviceList.map { $0.toDictionary() }
            notifyDelegates { [self] delegate in
                delegate.didReceiveDeviceData?(deviceData: [
                    "deviceList": deviceDictionaries,
                    "deviceCount": deviceList.count
                ])
            }
        }
    }
    
    func didReceiveFilesData(_ userInfo: [String: Any]) {
        logger.log("didReceiveFilesData start", level: .info)
        notifyDelegates { delegate in
            delegate.didReceiveFilesData?(userInfo)
        }
    }
    
    func didReceiveTransferFilesDataUpdate(_ userInfo: [String: Any]) {
        logger.log("didReceiveTransferFilesDataUpdate ", level: .info)
        notifyDelegates { delegate in
            delegate.didReceiveTransferFilesDataUpdate?(userInfo)
        }
    }
    
    func didReceiveDIASStatus(_ status: Int){
        logger.log("didReceiveDIASStatus:\(status)", level: .info)
        notifyDelegates { delegate in
            delegate.didReceiveDIASStatus?(status)
        }
    }
    
    func didReceiveErrorEvent(_ errorInfo: [String: Any]) {
        logger.log("didReceiveErrorEvent: \(errorInfo)", level: .info)
        notifyDelegates { delegate in
            delegate.didReceiveErrorEvent?(errorInfo)
        }
    }

    func didReceiveSystemInfoUpdate(_ systemInfo: [String: Any]) {
        logger.log("didReceiveSystemInfoUpdate - systemInfo:\(systemInfo)", level: .info)
        notifyDelegates { delegate in
            delegate.didReceiveSystemInfoUpdate?(systemInfo)
        }
    }

    func didDetectScreenCountChange(change: String, currentCount: Int, previousCount: Int) {
        logger.log("Screen count changed - \(change) (\(previousCount) -> \(currentCount))", level: .info)

        if change == "decreased" {
            logger.log("Screen count decreased, clearing device list", level: .info)
            clearDeviceList()
        }

        notifyDelegates { delegate in
            delegate.didDetectScreenCountChange?(change: change, currentCount: currentCount, previousCount: previousCount)
        }
    }

    private func handleAuthenticationRequest(index: UInt32) {
        guard let displayID = GoCallbackManager.shared.getCurrentActiveDisplayID() else {
            logger.log("No valid display ID found for auth request", level: .error)
            SetAuthStatusCode(0)
            return
        }
        HelperDisplayManager.shared.queryAuthStatus(for: displayID, index: UInt16(index)) { [weak self] success in
            self?.logger.log("Auth status query result: \(success)", level: .info)
            let authResult: UInt32 = success ? 1 : 0
            SetAuthStatusCode(authResult)
            if success {
                self?.logger.log("Authentication successful for display \(displayID)", level: .info)
            } else {
                self?.logger.log("Authentication failed for display \(displayID)", level: .error)
            }
        }
    }

    private func handleUpdateMousePosRequest(width: UInt16, height: UInt16, posX: Int16, posY: Int16) {
        guard let displayID = GoCallbackManager.shared.getCurrentActiveDisplayID() else {
            logger.log("No valid display ID found for auth request", level: .error)
            return
        }
        HelperDisplayManager.shared.updateMousePos(for: displayID, width: width, height: height, posX: posX, posY: posY)
    }

    func updateDisplayMapping(mac: String, displayID: UInt32, completion: @escaping (Bool) -> Void) {
        logger.log("Updating display mapping: \(mac) -> \(displayID)", level: .info)
        GoCallbackManager.shared.updateDisplayMapping(mac: mac, displayID: CGDirectDisplayID(displayID))
        completion(true)
    }
    
    func rescanDisplays(completion: @escaping (Bool) -> Void) {
        logger.log("Rescanning displays", level: .info)
        HelperDisplayManager.shared.checkDisplaysNow()
        completion(true)
    }
    
    func sendTextToRemote(_ text: String, completion: @escaping (Bool, String?) -> Void) {
        logger.log("Sending text to remote: \(text.prefix(100))...", level: .info)
        SendXClipData(text.toGoStringXPC(), "".toGoStringXPC(), "".toGoStringXPC())
        completion(true, nil)
    }
    
    func sendImageToRemote(_ imageData: Data, completion: @escaping (Bool, String?) -> Void) {
        logger.log("Sending image to remote, size: \(imageData.count) bytes", level: .info)
        let base64String = imageData.base64EncodedString()
        SendXClipData("".toGoStringXPC(), base64String.toGoStringXPC(), "".toGoStringXPC())
        completion(true, nil)
    }
    
    func startClipboardMonitoring(completion: @escaping (Bool) -> Void) {
        logger.log("Clipboard monitoring start requested (already auto-started)", level: .info)
        // Monitoring is already started in init(), this is just for compatibility
        completion(true)
    }
    
    func stopClipboardMonitoring(completion: @escaping (Bool) -> Void) {
        logger.log("Stopping clipboard monitoring", level: .info)

        ClipboardMonitor.shareInstance().stopMonitoring()

        completion(true)
    }

    func getDeviceList(completion: @escaping ([[String: Any]]) -> Void) {
        logger.log("GUI requested device list", level: .info)

        deviceListLock.lock()
        let devices = deviceList
        deviceListLock.unlock()

        let deviceDictionaries = devices.map { device in
            return device.toDictionary()
        }

        logger.log("Returning \(deviceDictionaries.count) devices to GUI", level: .info)
        completion(deviceDictionaries)
    }

    func sendMultiFilesDropRequest(multiFilesData: String, completion: @escaping (Bool, String?) -> Void) {
        logger.log("Sending multi-files drop request", level: .info)
        goBridge.sendMultiFilesDropRequest(multiFilesData: multiFilesData, completion: completion)
    }
    
    func setCancelFileTransfer(ipPort: String, clientID: String, timeStamp: UInt64, completion: @escaping (Bool, String?) -> Void) {
        logger.log("Cancelling file transfer - IPPort: \(ipPort), ClientID: \(clientID), TimeStamp: \(timeStamp)", level: .info)
        goBridge.setCancelFileTransfer(ipPort: ipPort, clientID: clientID, timeStamp: timeStamp, completion: completion)
    }
    
    func setDragFileListRequest(multiFilesData: String, timestamp: UInt64, width: UInt16, height: UInt16, posX: Int16, posY: Int16, completion: @escaping (Bool, String?) -> Void) {
        logger.log("Sending multi-files drag request", level: .info)
        goBridge.setDragFileListRequest(multiFilesData: multiFilesData, timestamp: timestamp, completion: completion)
        handleUpdateMousePosRequest(width: width, height: height, posX: posX, posY: posY)
    }
    
    func requestUpdateDownloadPath(downloadPath: String, completion: @escaping (Bool, String?) -> Void) {
        logger.log("Updating download path: \(downloadPath)", level: .info)
        goBridge.requestUpdateDownloadPath(downloadPath: downloadPath, completion: completion)
    }

    func clearDeviceList() {
        deviceListLock.lock()
        let wasEmpty = deviceList.isEmpty
        deviceList.removeAll()
        deviceListLock.unlock()

        if !wasEmpty {
            logger.log("Helper XPC Service: Cleared device list", level: .info)
            notifyDelegates { delegate in
                delegate.didReceiveDeviceData?(deviceData: [
                    "deviceList": [],
                    "deviceCount": 0
                ])
            }
        }
    }

    func removeDevice(byId deviceId: String) {
        deviceListLock.lock()
        defer { deviceListLock.unlock() }

        if let index = deviceList.firstIndex(where: { $0.id == deviceId }) {
            let removedDevice = deviceList.remove(at: index)
            logger.log("Helper XPC Service: Removed device with ID: \(deviceId), remaining count: \(deviceList.count)", level: .info)
        }
    }
}

extension HelperXPCService: ClipboardMonitorDelegate {
    func clipboardDidChange(text: String?, image: NSImage?, html: String?) {
        logger.log("Local clipboard changed", level: .info)

        var textDataStr = ""
        var imageDataStr = ""
        var htmlDataStr = ""

        if let text = text, !text.isEmpty {
            textDataStr = text
            logger.log("Found text data: \(text.prefix(100))...", level: .info)
        }

        if let image = image {
            if let imageData = image.jpegData {
                imageDataStr = imageData.base64EncodedString()
                logger.log("Found image data, size: \(imageData.count) bytes", level: .info)
            } else {
                logger.log("Failed to convert image to base64", level: .error)
            }
        }

        if let html = html, !html.isEmpty {
            let result = ClipboardMonitor.shareInstance().processHTMLForSending(html)
            htmlDataStr = result.processedHTML
            logger.log("Found HTML data, length: \(htmlDataStr.count)", level: .info)
        }

        if !textDataStr.isEmpty || !imageDataStr.isEmpty || !htmlDataStr.isEmpty {
            SendXClipData(
                textDataStr.toGoStringXPC(),
                imageDataStr.toGoStringXPC(),
                htmlDataStr.toGoStringXPC()
            )
            logger.log("Sent clipboard data - text: \(!textDataStr.isEmpty), image: \(!imageDataStr.isEmpty), html: \(!htmlDataStr.isEmpty)", level: .info)
        } else {
            logger.log("No valid clipboard content to send", level: .warn)
        }
    }
}
