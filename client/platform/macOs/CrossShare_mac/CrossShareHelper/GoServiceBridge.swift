//
//  GoServiceBridge.swift
//  CrossShareHelper
//
//  Created by TS on 2025/8/27.
//  Go Service Bridge - Connects Swift and Go services
//

import Foundation
import AppKit

private weak var globalGoServiceBridge: GoServiceBridge?

private let authViaIndexCallback: @convention(c) (UInt32) -> Void = { index in
    globalGoServiceBridge?.handAuthViaIndex(idx: index)
}

private let requestSourceAndPortCallback: @convention(c) () -> Void = {
    globalGoServiceBridge?.handleRequestSourceAndPort()
}

private let setDIASStatusCallback: @convention(c) (UInt32) -> Void = { status in
    globalGoServiceBridge?.handleSetDIASStatus(status: status)
}

private let updateClientStatusCallback: @convention(c) (UnsafeMutablePointer<CChar>?) -> Void = { clientJsonStr in
    globalGoServiceBridge?.handleUpdateClientStatus(clientJsonStr: clientJsonStr)
}

private let updateSystemInfoCallback: @convention(c) (UnsafeMutablePointer<CChar>?, UnsafeMutablePointer<CChar>?) -> Void = { ipInfo, verInfo in
    globalGoServiceBridge?.handleUpdateSystemInfo(ipInfo: ipInfo, verInfo: verInfo)
}

private let pasteXClipDataCallback: @convention(c) (UnsafeMutablePointer<CChar>?, UnsafeMutablePointer<CChar>?, UnsafeMutablePointer<CChar>?, UnsafeMutablePointer<CChar>?) -> Void = { textPtr, imagePtr, htmlPtr, rtfPtr in
    // Decode text with multiple encoding support
    var text: String? = nil
    if let textPtr = textPtr {
        let textData = Data(bytes: textPtr, count: strlen(textPtr))

        if let utf8Text = String(data: textData, encoding: .utf8),
           !utf8Text.contains("�") && !utf8Text.contains("\u{FFFD}") {
            text = utf8Text
        } else {
            // Try GB18030 encoding
            let encoding = CFStringConvertEncodingToNSStringEncoding(
                CFStringEncoding(CFStringEncodings.GB_18030_2000.rawValue))
            if let gb18030Text = String(data: textData, encoding: String.Encoding(rawValue: encoding)) {
                text = gb18030Text
            } else if let utf16Text = String(data: textData, encoding: .utf16) {
                text = utf16Text
            } else if let latinText = String(data: textData, encoding: .isoLatin1) {
                text = latinText
            } else {
                text = String(cString: textPtr)
            }
        }
    }

    // Decode image base64
    let imageBase64 = imagePtr != nil ? String(cString: imagePtr!) : nil

    // Decode HTML with multiple encoding support
    var html: String? = nil
    if let htmlPtr = htmlPtr {
        let data = Data(bytes: htmlPtr, count: strlen(htmlPtr))

        // Check for UTF-8 BOM
        let hasUTF8BOM = data.count >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF
        let utf8Data = hasUTF8BOM ? data.subdata(in: 3..<data.count) : data

        if let utf8String = String(data: utf8Data, encoding: .utf8),
           !utf8String.contains("�") && !utf8String.contains("\u{FFFD}") {
            html = utf8String
        } else {
            // Try GB18030 encoding
            let encoding = CFStringConvertEncodingToNSStringEncoding(
                CFStringEncoding(CFStringEncodings.GB_18030_2000.rawValue))
            if let gb18030String = String(data: data, encoding: String.Encoding(rawValue: encoding)) {
                html = gb18030String
            } else if let utf16String = String(data: data, encoding: .utf16) {
                html = utf16String
            } else if let utf16LEString = String(data: data, encoding: .utf16LittleEndian) {
                html = utf16LEString
            } else if let utf16BEString = String(data: data, encoding: .utf16BigEndian) {
                html = utf16BEString
            } else if let latinString = String(data: data, encoding: .isoLatin1) {
                html = latinString
            } else {
                html = String(cString: htmlPtr)
            }
        }
    }

    // Decode RTF with multiple encoding support
    var rtf: String? = nil
    if let rtfPtr = rtfPtr {
        let data = Data(bytes: rtfPtr, count: strlen(rtfPtr))

        if let utf8String = String(data: data, encoding: .utf8),
           !utf8String.contains("�") && !utf8String.contains("\u{FFFD}") {
            rtf = utf8String
        } else {
            // Try GB18030 encoding
            let encoding = CFStringConvertEncodingToNSStringEncoding(
                CFStringEncoding(CFStringEncodings.GB_18030_2000.rawValue))
            if let gb18030String = String(data: data, encoding: String.Encoding(rawValue: encoding)) {
                rtf = gb18030String
            } else if let utf16String = String(data: data, encoding: .utf16) {
                rtf = utf16String
            } else if let utf16LEString = String(data: data, encoding: .utf16LittleEndian) {
                rtf = utf16LEString
            } else if let utf16BEString = String(data: data, encoding: .utf16BigEndian) {
                rtf = utf16BEString
            } else if let latinString = String(data: data, encoding: .isoLatin1) {
                rtf = latinString
            } else {
                rtf = String(cString: rtfPtr)
            }
        }
    }

    // Handle clipboard data with new unified method
    if text != nil || imageBase64 != nil || html != nil || rtf != nil {
        GoCallbackManager.shared.handleRemoteData(text: text, imageBase64: imageBase64, html: html, rtf: rtf)
    }
}

private let multipleProgressBarCallback: @convention(c) (Optional<UnsafeMutablePointer<Int8>>, Optional<UnsafeMutablePointer<Int8>>, Optional<UnsafeMutablePointer<Int8>>, UInt32, UInt32, UInt64, UInt64, UInt64, UInt64) -> Void = { ip, id, currentfileName, recvFileCnt, totalFileCnt, currentFileSize, totalSize, recvSize, timestamp in
    globalGoServiceBridge?.handleMultipleProgressBar(ip: ip, id: id, deviceName: nil, currentfileName: currentfileName, recvFileCnt: recvFileCnt, totalFileCnt: totalFileCnt, currentFileSize: currentFileSize, totalSize: totalSize, recvSize: recvSize, timestamp: timestamp)
}

private let fileListNotifyCallback: @convention(c) (Optional<UnsafeMutablePointer<Int8>>, Optional<UnsafeMutablePointer<Int8>>, Optional<UnsafeMutablePointer<Int8>>, UInt32, UInt64, UInt64, Optional<UnsafeMutablePointer<Int8>>, UInt64) -> Void = { ip, id, platform, fileCnt, totalSize, timestamp, firstFileName, firstFileSize in
    globalGoServiceBridge?.handleFileListNotify(ip: ip, id: id, platform: platform, fileCnt: fileCnt, totalSize: totalSize, timestamp: timestamp, firstFileName: firstFileName, firstFileSize: firstFileSize)
}

private let notifyErrEventCallback: @convention(c) (Optional<UnsafeMutablePointer<Int8>>, UInt32, Optional<UnsafeMutablePointer<Int8>>, Optional<UnsafeMutablePointer<Int8>>, Optional<UnsafeMutablePointer<Int8>>, Optional<UnsafeMutablePointer<Int8>>) -> Void = { clientID, errCode, ipaddr, timestamp, arg3, arg4 in
    globalGoServiceBridge?.handleNotifyErrEvent(clientID: clientID, errCode: errCode, ipaddr: ipaddr, timestamp: timestamp, arg3: arg3, arg4: arg4)
}

private let notiMessageCallback: @convention(c) (UInt64, UInt32, Optional<UnsafeMutablePointer<Optional<UnsafeMutablePointer<Int8>>>>, Int32) -> Void = { timestamp, notiCode, notiParam, paramCount in
    globalGoServiceBridge?.handleNotiMessage(timestamp: timestamp, notiCode: notiCode, notiParam: notiParam, paramCount: paramCount)
}


class GoServiceBridge: CrossShareHelperXPCDelegate {
    
//    private let logger = CSLogger.shared
    private var isServiceRunning = false
    private var currentConfig: [String: Any] = [:]
    
    private var pendingDIASRequests: [(diasID: String, completion: (Bool, String?) -> Void)] = []
    private var pendingDisplayEventRequests: [(jsonString: String, completion: (Bool, String?) -> Void)] = []
    private let pendingRequestsQueue = DispatchQueue(label: "com.crossshare.pending.requests", qos: .userInitiated)
    private var pendingRequestsTimer: Timer?
    
    static let shared = GoServiceBridge()
    
    private init() {
        logger.info("Go Service Bridge initialized")
        initManagers()
        globalGoServiceBridge = self
        setupGoCallbacks()
    }
    
    func initManagers() {
        ClipboardMonitor.shareInstance().startMonitoring()
        logger.info("Clipboard monitoring started")
    }
    
    private func setupGoCallbacks() {
        SetCallbackAuthViaIndex(authViaIndexCallback)
        SetCallbackDIASStatus(setDIASStatusCallback)
        SetCallbackUpdateClientStatus(updateClientStatusCallback)
        SetCallbackUpdateSystemInfo(updateSystemInfoCallback)
        SetCallbackRequestSourceAndPort(requestSourceAndPortCallback)
        SetCallbackPasteXClipData(pasteXClipDataCallback)
        SetCallbackUpdateMultipleProgressBar(multipleProgressBarCallback)
        SetCallbackMethodFileListNotify(fileListNotifyCallback)
        SetCallbackNotifyErrEvent(notifyErrEventCallback)
        SetCallbackNotiMessage(notiMessageCallback)
    }
    
    func startService(config: [String: Any], completion: @escaping (Bool, String?) -> Void) {
        guard !isServiceRunning else {
            completion(false, "Service is already running")
            return
        }
        
        logger.info("Starting P2P Go service with MainInit - config: \(config)")
        
        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            guard let deviceName = config["deviceName"] as? String,
                  let rootPath = config["rootPath"] as? String,
                  let downloadPath = config["downloadPath"] as? String,
                  let serverId = config["serverId"] as? String,
                  let serverIpInfo = config["serverIpInfo"] as? String,
                  let listenHost = config["listenHost"] as? String,
                  let listenPort = config["listenPort"] as? Int32 else {
                
                DispatchQueue.main.async {
                    completion(false, "Invalid config parameters for MainInit")
                }
                return
            }
            
            logger.info("Gathering network interface information before MainInit...")
            
            WifiManager.shareInstance().getNetInfoFromLocalIp { netName, mac, mtu, index, flags in
                DispatchQueue.global(qos: .userInitiated).async { [weak self] in
                    if let netName = netName, let mac = mac, let mtu = mtu, let index = index, let flags = flags {
                        logger.info("Network interface: name=\(netName), mac=\(mac), mtu=\(mtu), index=\(index), flags=\(flags)")
                        self?.sendNetInterfaces(name: netName, mac: mac, mtu: mtu, index: index, flags: flags)
                    } else {
                        logger.warn("Failed to get network interface information, proceeding without it")
                    }
                    
                    logger.info("Calling MainInit with: deviceName=\(deviceName),rootPath = \(rootPath) serverId=\(serverId), serverIpInfo=\(serverIpInfo), listenHost=\(listenHost), listenPort=\(listenPort)")
                    
                    self?.callMainInit(
                        deviceName: deviceName,
                        rootPath: rootPath,
                        downloadPath: downloadPath,
                        serverId: serverId,
                        serverIpInfo: serverIpInfo,
                        listenHost: listenHost,
                        listenPort: listenPort,
                        completion: completion
                    )
                }
            }
        }
    }
    
    private func callMainInit(deviceName: String,rootPath:String,downloadPath:String,serverId: String, serverIpInfo: String, listenHost: String, listenPort: Int32, completion: @escaping (Bool, String?) -> Void) {
        logger.info("Converting parameters for MainInit...")
        let deviceNameGo = deviceName.toGoStringXPC()
        let rootPathGo = rootPath.toGoStringXPC()
        let downloadPathGo = downloadPath.toGoStringXPC()
        
        logger.info("Calling MainInit with Go strings...")
        logger.info("Parameters: deviceName=\(deviceName), rootPath=\(rootPath), serverId=\(serverId), serverIpInfo=\(serverIpInfo), listenHost=\(listenHost), listenPort=\(listenPort)")
        
        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            logger.info("Calling MainInit in background thread...")
            guard let _ = self else { return }
            MainInit(deviceNameGo, rootPathGo,downloadPathGo)
            logger.info("MainInit call returned (this may not be reached if MainInit runs indefinitely)")
        }
        
        DispatchQueue.main.asyncAfter(deadline: .now() + 1.0) { [weak self] in
            logger.info("Starting pending requests monitoring")
            self?.startPendingRequestsMonitoring()
            self?.isServiceRunning = true
            completion(true, nil)
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
    
    
    func getServiceInfo(completion: @escaping ([String: Any]?) -> Void) {
        
        guard isServiceRunning else {
            logger.warn("Service not running, returning basic info")
            let basicInfo: [String: Any] = [
                "isRunning": false,
                "version": "Unknown",
                "buildDate": "Unknown"
            ]
            completion(basicInfo)
            return
        }
        
        var info = currentConfig
        info["isRunning"] = isServiceRunning
        info["startTime"] = Date().timeIntervalSince1970
        
        if info["version"] == nil {
            let versionPtr = GetVersion()
            info["version"] = versionPtr != nil ? String(cString: versionPtr!) : "Unknown"
        }
        if info["buildDate"] == nil {
            let buildDatePtr = GetBuildDate()
            info["buildDate"] = buildDatePtr != nil ? String(cString: buildDatePtr!) : "Unknown"
        }
        
        completion(info)
    }
    
    private func sendNetInterfaces(name: String, mac: String, mtu: Int32, index: UInt32, flags: UInt32) {
        logger.info("Sending network interface info to Go: name=\(name), mac=\(mac), mtu=\(mtu), index=\(index), flags=\(flags)")
        
        let nameGo = name.toGoStringXPC()
        let macGo = mac.toGoStringXPC()
        
        DispatchQueue.main.async {
            SendNetInterfaces(nameGo, macGo, GoInt(mtu), GoInt(index), GoUint(flags))
        }
        
        logger.info("Network interface information sent to Go service")
    }
    
    func setDIASID(diasID: String, completion: @escaping (Bool, String?) -> Void) {
        logger.info("Setting DIASID to: \(diasID)")
        
        guard !diasID.isEmpty else {
            let error = "DIASID cannot be empty"
            logger.error(error)
            completion(false, error)
            return
        }
        
        pendingRequestsQueue.async { [weak self] in
            guard let self = self else { return }
            if self.isServiceRunning {
                self.executeDIASIDRequest(diasID: diasID, completion: completion)
            } else {
                logger.info("Go service not ready, queuing DIAS ID request: \(diasID)")
                self.pendingDIASRequests.append((diasID: diasID, completion: completion))
            }
        }
    }
    
    private func executeDIASIDRequest(diasID: String, completion: @escaping (Bool, String?) -> Void) {
        logger.info("Executing DIAS ID request: \(diasID)")
        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            if let macBytes = self?.parseMacAddress(diasID) {
                macBytes.withUnsafeBytes { bytesPtr in
                    if let baseAddress = bytesPtr.baseAddress {
                        SetMacAddress(UnsafeMutablePointer(mutating: baseAddress.assumingMemoryBound(to: CChar.self)), 6)
                    }
                }
                logger.info("SetMacAddress called with 6-byte MAC: \(diasID)")
            } else {
                logger.error("Failed to parse MAC address: \(diasID)")
            }
            
            DispatchQueue.main.async {
                completion(true, nil)
            }
        }
    }
    
    private func parseMacAddress(_ macString: String) -> [UInt8]? {
        let components = macString.components(separatedBy: ":")
        guard components.count == 6 else {
            logger.error("Invalid MAC address format: \(macString), expected 6 components")
            return nil
        }
        
        var macBytes: [UInt8] = []
        for component in components {
            guard let byte = UInt8(component, radix: 16) else {
                logger.error("Invalid hex component in MAC address: \(component)")
                return nil
            }
            macBytes.append(byte)
        }
        
        logger.info("Parsed MAC address \(macString) to bytes: \(macBytes.map { String(format: "%02X", $0) }.joined(separator: ":"))")
        return macBytes
    }
    
    private func processPendingRequests() {
        pendingRequestsQueue.async { [weak self] in
            guard let self = self else { return }
            
            // Process pending DIAS ID requests
            let diasRequests = self.pendingDIASRequests
            self.pendingDIASRequests.removeAll()
            if !diasRequests.isEmpty {
                logger.info("Processing \(diasRequests.count) pending DIAS requests")
                for request in diasRequests {
                    self.executeDIASIDRequest(diasID: request.diasID, completion: request.completion)
                }
            }
            
            // Process pending display event requests
            let displayEventRequests = self.pendingDisplayEventRequests
            self.pendingDisplayEventRequests.removeAll()
            if !displayEventRequests.isEmpty {
                logger.info("Processing \(displayEventRequests.count) pending DisplayEvent requests")
                for request in displayEventRequests {
                    self.executeDisplayEventInfoRequest(jsonString: request.jsonString, completion: request.completion)
                }
            }
            
            // Stop monitoring only when both queues are empty
            if !diasRequests.isEmpty || !displayEventRequests.isEmpty {
                DispatchQueue.main.async { [weak self] in
                    guard let self = self else { return }
                    if self.pendingDIASRequests.isEmpty && self.pendingDisplayEventRequests.isEmpty {
                        self.stopPendingRequestsMonitoring()
                    }
                }
            }
        }
    }
    
    private func startPendingRequestsMonitoring() {
        stopPendingRequestsMonitoring()
        
        pendingRequestsTimer = Timer.scheduledTimer(withTimeInterval: 2.0, repeats: true) { [weak self] _ in
            self?.tryProcessPendingRequests()
        }
        logger.info("Started pending requests monitoring timer")
    }
    
    private func stopPendingRequestsMonitoring() {
        pendingRequestsTimer?.invalidate()
        pendingRequestsTimer = nil
        logger.info("Stopped pending requests monitoring timer")
    }
    
    private func tryProcessPendingRequests() {
        pendingRequestsQueue.async { [weak self] in
            guard let self = self else { return }
            let totalPending = self.pendingDIASRequests.count + self.pendingDisplayEventRequests.count
            if totalPending > 0 {
                logger.info("Attempting to process \(totalPending) pending requests (DIAS: \(self.pendingDIASRequests.count), DisplayEvent: \(self.pendingDisplayEventRequests.count))")
                self.processPendingRequests()
            }
        }
    }
    
    func setExtractDIAS(completion: @escaping (Bool, String?) -> Void) {
        logger.info("Calling SetExtractDIAS")
        guard isServiceRunning else {
            let error = "Go service is not running"
            logger.error(error)
            completion(false, error)
            return
        }
        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            SetExtractDIAS()
            logger.info("SetExtractDIAS called")
            guard let _ = self else { return }
            DispatchQueue.main.async {
                completion(true, nil)
            }
        }
    }
    
    /// Send display event info JSON to Go service via SetDisplayEventInfo
    /// If Go service is not ready, the request is queued and will be sent once the service starts.
    func setDisplayEventInfo(jsonString: String, completion: @escaping (Bool, String?) -> Void) {
        logger.info("Calling SetDisplayEventInfo with JSON: \(jsonString)")
        
        pendingRequestsQueue.async { [weak self] in
            guard let self = self else { return }
            if self.isServiceRunning {
                self.executeDisplayEventInfoRequest(jsonString: jsonString, completion: completion)
            } else {
                logger.info("Go service not ready, queuing SetDisplayEventInfo request")
                self.pendingDisplayEventRequests.append((jsonString: jsonString, completion: completion))
            }
        }
    }
    
    private func executeDisplayEventInfoRequest(jsonString: String, completion: @escaping (Bool, String?) -> Void) {
        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            guard let _ = self else { return }
            let cString = jsonString.withCString { ptr -> UnsafeMutablePointer<CChar> in
                let len = strlen(ptr)
                let buf = UnsafeMutablePointer<CChar>.allocate(capacity: len + 1)
                strcpy(buf, ptr)
                return buf
            }
            SetDisplayEventInfo(cString)
            cString.deallocate()
            logger.info("SetDisplayEventInfo called successfully")
            DispatchQueue.main.async {
                completion(true, nil)
            }
        }
    }

    func sendMultiFilesDropRequest(multiFilesData: String, completion: @escaping (Bool, String?) -> Void) {
        guard isServiceRunning else {
            let error = "Go service is not running"
            logger.error(error)
            completion(false, error)
            return
        }
        logger.info("Sending multi-files drop request with data length: \(multiFilesData.count)")
        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            guard let _ = self else { return }
            let result = SendMultiFilesDropRequest(multiFilesData.toGoStringXPC())
            var statusString = ""
            switch result {
            case 1:
                statusString = "Transfer request sent successfully"
            case 2:
                statusString = "Cannot send: Parameter error"
            case 3:
                statusString = "Cannot send: Sending in progress"
            case 4:
                statusString = "Cannot send: Receiving in progress"
            case 5:
                statusString = "Cannot send: Callback not set in Go"
            case 6:
                statusString = "Cannot send: Message length exceeds limit"
            case 7:
                statusString = "Cannot send: Total file size exceeds 10GB"
            case 8:
                statusString = "Cannot send: Too many pending requests"
            default:
                logger.info("Transfer request sent successfully:\(result))")
            }
            let success = result == 1
            logger.info("Multi-files drop request result: \(success) - \(statusString)")
            DispatchQueue.main.async {
                completion(success, success ? nil : statusString)
            }
        }
    }
    
    func setCancelFileTransfer(ipPort: String, clientID: String, timeStamp: UInt64, completion: @escaping (Bool, String?) -> Void) {
        logger.info("Cancelling file transfer - IPPort: \(ipPort), ClientID: \(clientID), TimeStamp: \(timeStamp)")
        
        guard isServiceRunning else {
            let error = "Go service is not running"
            logger.error(error)
            completion(false, error)
            return
        }
        
        guard !ipPort.isEmpty, !clientID.isEmpty, timeStamp > 0 else {
            let error = "Invalid parameters for cancel file transfer"
            logger.error(error)
            completion(false, error)
            return
        }
        
        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            guard let _ = self else { return }
            let ipPortGo = ipPort.toGoStringXPC()
            let clientIDGo = clientID.toGoStringXPC()
            
            SetCancelFileTransfer(ipPortGo, clientIDGo, GoUint64(timeStamp))
            
            logger.info("SetCancelFileTransfer called successfully for \(ipPort)")
            
            DispatchQueue.main.async {
                completion(true, nil)
            }
        }
    }

    func setDragFileListRequest(multiFilesData: String, timestamp: UInt64, completion: @escaping (Bool, String?) -> Void) {
        guard isServiceRunning else {
            let error = "Go service is not running"
            logger.error(error)
            completion(false, error)
            return
        }
        logger.info("Sending multi-files drag request with data length: \(multiFilesData.count)")
        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            guard let _ = self else { return }
            SetDragFileListRequest(multiFilesData.toGoStringXPC(), timestamp)
            DispatchQueue.main.async {
                completion(true, nil)
            }
        }
    }
    
    func getIsSupportFileDrag(completion: @escaping (Bool) -> Void) {
        guard isServiceRunning else {
            logger.warn("Go service is not running, GetIsSupportFileDrag defaults to false")
            completion(false)
            return
        }
        
        DispatchQueue.global(qos: .userInitiated).async {
            let result = Int(GetIsSupportFileDrag())
            let isSupported = (result != 0)
            logger.info("GetIsSupportFileDrag() returned: \(result), interpreted as: \(isSupported)")
            DispatchQueue.main.async {
                completion(isSupported)
            }
        }
    }
    
    func requestUpdateDownloadPath(downloadPath: String, completion: @escaping (Bool, String?) -> Void) {
        logger.info("Updating download path: \(downloadPath)")
        
        guard isServiceRunning else {
            let error = "Go service is not running"
            logger.error(error)
            completion(false, error)
            return
        }
        
        guard !downloadPath.isEmpty else {
            let error = "Invalid download path parameter"
            logger.error(error)
            completion(false, error)
            return
        }
        
        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            guard let _ = self else { return }
            let downloadPathGo = downloadPath.toGoStringXPC()
            RequestUpdateDownloadPath(downloadPathGo)
            logger.info("RequestUpdateDownloadPath called successfully with path: \(downloadPath)")
            DispatchQueue.main.async {
                completion(true, nil)
            }
        }
    }

    func handAuthViaIndex(idx: UInt32) {
        logger.info("Received auth index from Go: \(idx)")
        guard isServiceRunning else {
            let error = "Go service is not running"
            logger.error(error)
            return
        }
        GoCallbackManager.shared.handleAuthIndex(Int(idx))
    }
    
    func handleSetDIASStatus(status: UInt32) {
        logger.info("Received DIAS status from Go: \(status)")
        GoCallbackManager.shared.handleDIASStatus(Int(status))
    }
    
    func handleUpdateClientStatus(clientJsonStr: UnsafeMutablePointer<CChar>?) {
        guard let clientJsonStr = clientJsonStr else {
            logger.error("Received nil client status")
            return
        }
        let clientStatus = String(cString: clientJsonStr)
        logger.info("Received client status from Go: \(clientStatus)")
        guard isServiceRunning else {
            logger.error("Go service is not running")
            return
        }
        GoCallbackManager.shared.handleClientStatus(clientStatus)
    }
    
    func handleUpdateSystemInfo(ipInfo: UnsafeMutablePointer<CChar>?, verInfo: UnsafeMutablePointer<CChar>?) {
        guard let ipInfo = ipInfo else {
            logger.error("Received nil ipInfo")
            return
        }
        
        let ipInfoString = String(cString: ipInfo)
        let verInfoString = verInfo != nil ? String(cString: verInfo!) : ""
        
        logger.info("Received system info update from Go - IP: \(ipInfoString), Version: \(verInfoString)")
        guard isServiceRunning else {
            logger.error("Go service is not running")
            return
        }
        GoCallbackManager.shared.handleSystemInfoUpdate(ipInfo: ipInfoString, verInfo: verInfoString)
    }
    
    func handleRequestSourceAndPort() {
        logger.info("Received request for source and port from Go")
        guard isServiceRunning else {
            logger.error("Go service is not running")
            return
        }
        GoCallbackManager.shared.handleRequestSourceAndPort()
    }

    func handleMultipleProgressBar(ip: Optional<UnsafeMutablePointer<Int8>>, id: Optional<UnsafeMutablePointer<Int8>>, deviceName: Optional<UnsafeMutablePointer<Int8>>, currentfileName: Optional<UnsafeMutablePointer<Int8>>, recvFileCnt: UInt32, totalFileCnt: UInt32, currentFileSize: UInt64, totalSize: UInt64, recvSize: UInt64, timestamp: UInt64) {

        guard let ipPtr = ip, let idPtr = id, let currentFileNamePtr = currentfileName else {
            logger.warn("Multiple progress bar callback received with null pointers")
            return
        }

        let senderIP = String(cString: ipPtr)
        let senderID = String(cString: idPtr)
        let currentFileName = String(cString: currentFileNamePtr)
        let deviceName = senderID

        logger.info("Multiple file progress: \(currentFileName) (\(recvFileCnt)/\(totalFileCnt)) - \(recvSize)/\(totalSize) bytes from \(senderIP)")

        let multipleProgress = MultipleFileTransferProgress(
            senderIP: senderIP,
            senderID: senderID,
            deviceName: deviceName,
            currentFileName: currentFileName,
            receivedFileCount: recvFileCnt,
            totalFileCount: totalFileCnt,
            currentFileSize: currentFileSize,
            totalSize: totalSize,
            receivedSize: recvSize,
            timestamp: timestamp
        )

        GoCallbackManager.shared.handleMultipleFileProgress(multipleProgress)
    }

    func handleFileListNotify(ip: Optional<UnsafeMutablePointer<Int8>>, id: Optional<UnsafeMutablePointer<Int8>>, platform: Optional<UnsafeMutablePointer<Int8>>, fileCnt: UInt32, totalSize: UInt64, timestamp: UInt64, firstFileName: Optional<UnsafeMutablePointer<Int8>>, firstFileSize: UInt64) {

        guard let ipPtr = ip, let idPtr = id, let platformPtr = platform, let firstFileNamePtr = firstFileName else {
            logger.warn("File list notify callback received with null pointers")
            return
        }

        let senderIP = String(cString: ipPtr)
        let senderID = String(cString: idPtr)
        let platformString = String(cString: platformPtr)
        let firstFileNameString = String(cString: firstFileNamePtr)

        logger.info("File transfer started: \(firstFileNameString) from \(senderIP) (\(senderID)) - Platform: \(platformString), Files: \(fileCnt), Total: \(totalSize) bytes")

        // 创建文件传输会话
        let session = FileTransferSession(
            sessionId: "\(senderID)-\(timestamp)",
            senderIP: senderIP,
            senderID: senderID,
            deviceName: platformString,
            direction: .receive,
            totalFileCount: fileCnt,
            totalSize: totalSize,
            currentFileName: firstFileNameString,
            currentFileSize: firstFileSize
        )
        
        // 假设 session、senderID 已经有定义
        let userInfo: [String: Any] = [
            "session": session.toDictionary(),
            "sessionId": session.sessionId,
            "senderID": senderID,
            "isCompleted": false,
            "progress": 0.0
        ]

//        // 通过 GoCallbackManager 处理文件传输开始事件
//        DispatchQueue.main.async {
//            NotificationCenter.default.post(
//                name: .fileTransferSessionStarted,
//                object: session,
//                userInfo: [
//                    "session": session.toDictionary(),
//                    "sessionId": session.sessionId,
//                    "senderID": senderID,
//                    "isCompleted": false,
//                    "progress": 0.0
//                ]
//            )
//        }
        
        GoCallbackManager.shared.handReceiveFilesData(userInfo)
        logger.info("File transfer session started notification sent - \(session.sessionId)")
    }
    
    func handleNotifyErrEvent(clientID: Optional<UnsafeMutablePointer<Int8>>, errCode: UInt32, ipaddr: Optional<UnsafeMutablePointer<Int8>>, timestamp: Optional<UnsafeMutablePointer<Int8>>, arg3: Optional<UnsafeMutablePointer<Int8>>, arg4: Optional<UnsafeMutablePointer<Int8>>) {
        
        let idString = clientID != nil ? String(cString: clientID!) : ""
        let arg1String = ipaddr != nil ? String(cString: ipaddr!) : ""
        let arg2String = timestamp != nil ? String(cString: timestamp!) : ""
        let arg3String = arg3 != nil ? String(cString: arg3!) : ""
        let arg4String = arg4 != nil ? String(cString: arg4!) : ""
        
        logger.error("Error event received - ID: \(idString), ErrCode: \(errCode), Args: [\(arg1String), \(arg2String), \(arg3String), \(arg4String)]")
        
        GoCallbackManager.shared.handleErrEvent(id: idString, errCode: errCode, arg1: arg1String, arg2: arg2String, arg3: arg3String, arg4: arg4String)
    }
    
    func handleNotiMessage(timestamp: UInt64, notiCode: UInt32, notiParam: Optional<UnsafeMutablePointer<Optional<UnsafeMutablePointer<Int8>>>>, paramCount: Int32) {
        var paramStrings: [String] = []
        
        if let paramArray = notiParam, paramCount > 0 {
            for i in 0..<Int(paramCount) {
                if let paramPtr = paramArray[i] {
                    paramStrings.append(String(cString: paramPtr))
                } else {
                    paramStrings.append("")
                }
            }
        }
        
        logger.info("NotiMessage received - timestamp: \(timestamp), notiCode: \(notiCode), paramCount: \(paramCount), params: \(paramStrings)")
        
        GoCallbackManager.shared.handleNotiMessage(timestamp: timestamp, notiCode: notiCode, notiParam: paramStrings)
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

private extension GoServiceBridge {
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

