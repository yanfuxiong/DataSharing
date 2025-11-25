//
//  P2PManager.swift
//  CrossShare
//
//  Created by user00 on 2025/3/12.
//

import UIKit
import SwiftyJSON

public let AppMainDirectory = FileManager.default.urls(for: .documentDirectory, in: .userDomainMask).first?.appendingPathComponent("CrossShare")
public let ReceiveFuleSuccessNotification = Notification.Name("ReceiveFuleSuccessNotification")
public let UpdateClientListSuccessNotification = Notification.Name("UpdateClientListSuccessNotification")
public let UpdateDiassStatusChangedNotification = Notification.Name("UpdateDiassStatusChangedNotification")
public let UpdateMonitorNameChangedNotification = Notification.Name("UpdateMonitorNameChangedNotification")
public let UpdateLanserviceChangedNotification = Notification.Name("UpdateLanserviceChangedNotification")
public let UpdateLanserviceListNotification = Notification.Name("UpdateLanserviceListNotification")
public let UpdateVersionNotification = Notification.Name("UpdateVersionNotification")
public let UpdateVerifyClientIndexChangedNotification = Notification.Name("UpdateVerifyClientIndexChangedNotification")

private weak var globalP2PManager: P2PManager?

// Global C compatible callback function
private let pasteXClipDataCallback: @convention(c) (UnsafeMutablePointer<CChar>?, UnsafeMutablePointer<CChar>?, UnsafeMutablePointer<CChar>?) -> Void = { cText, cImage, cHtml in
    globalP2PManager?.handlePasteXClipData(cText: cText, cImage: cImage, cHtml: cHtml)
}

private let startBrowseMdnsCallback: @convention(c) (UnsafeMutablePointer<CChar>?, UnsafeMutablePointer<CChar>?) -> Void = { cinstance, cserviceType in
    globalP2PManager?.handleStartBrowseMdns(cinstance: cinstance, cserviceType: cserviceType)
}

private let stopBrowseMdnsCallback: @convention(c) () -> Void = {
    globalP2PManager?.handleStopBrowseMdns()
}

private let multipleProgressBarCallback: @convention(c) (UnsafeMutablePointer<CChar>?, UnsafeMutablePointer<CChar>?, UnsafeMutablePointer<CChar>?, UnsafeMutablePointer<CChar>?, UInt32, UInt32, UInt64, UInt64, UInt64, UInt64) -> Void = { ip, id, deviceName, currentfileName, recvFileCnt, totalFileCnt, currentFileSize, totalSize, recvSize, timestamp in
    let currentfileNameString = currentfileName != nil ? String(cString: currentfileName!) : "nil"
    Logger.info("[multipleProgressBarCallback] currentfileName: \(currentfileNameString), timestamp: \(timestamp)")
    globalP2PManager?.handleMultipleProgressBar(ip: ip, id: id, deviceName: deviceName, currentfileName: currentfileName, recvFileCnt: recvFileCnt, totalFileCnt: totalFileCnt, currentFileSize: currentFileSize, totalSize: totalSize, recvSize: recvSize, timestamp: timestamp)
}

private let authDataCallback: @convention(c) (UInt32) -> UnsafeMutablePointer<CChar>? = { clientIndex in
    return globalP2PManager?.handleGetAuthData(clientIndex: clientIndex)
}

private let diasStatusCallback: @convention(c) (UInt32) -> Void = { status in
    globalP2PManager?.handleDIASStatus(status: status)
}

private let monitorNameCallback: @convention(c) (UnsafeMutablePointer<CChar>?) -> Void = { name in
    globalP2PManager?.handleMonitorName(name: name)
}

private let browseResultCallback: @convention(c) (UnsafeMutablePointer<CChar>?, UnsafeMutablePointer<CChar>?, UnsafeMutablePointer<CChar>?, UnsafeMutablePointer<CChar>?, UInt64) -> Void = { monitorName, instance, ip, version, timestamp in
    globalP2PManager?.handleBrowseResult(monitorName: monitorName, instance: instance, ip: ip, version: version, timestamp: timestamp)
}

private let updateClientStatusCallback: @convention(c) (UnsafeMutablePointer<CChar>?) -> Void = { clientJsonStr in
    globalP2PManager?.handleUpdateClientStatus(clientJsonStr: clientJsonStr)
}

enum DiassStatus: Int,CaseIterable {
    case WaitConnecting            = 1
    case SearchingService          = 2
    case CheckingAuthorization     = 3
    case WaitScreenCasting         = 4
    case FailedAuthorization       = 5
    case ConnectedNoClients        = 6
    case Connected                 = 7
    case ConnectedFailed           = 8
    case SearchingClients          = 99
}

class P2PManager {
    static let shared = P2PManager()
    private var p2pQueue: DispatchQueue?
    private var isRunning: Bool = false
    private var mClientList: [ClientInfo] = []
    private var mServiceList: [LanServiceInfo] = []
    private var screenData: String = ""
    public var cStatus: DiassStatus = .WaitConnecting {
        didSet {
            let isLastStatusConnected = (oldValue == .ConnectedNoClients || oldValue == .Connected)
            if cStatus == .ConnectedNoClients || cStatus == .Connected {
                if !isLastStatusConnected {
                    PictureInPictureManager.shared.updatePIPStatus(contentType: .idle)
                }
            } else {
                if isLastStatusConnected {
                    PictureInPictureManager.shared.updatePIPStatus(contentType: .disconnect)
                }
            }
        }
    }
    public var monitorName: String = ""
    public var newVersion: String = ""
    private var transferCallbacks:[String: TransferCallbacks] = [:]
    private var updateClientListWorkItem: DispatchWorkItem?
    struct TransferCallbacks {
        let onProgress: (Float, String) -> Void
        let onSuccess: () -> Void
        let onError: (String) -> Void
    }
    
    struct P2PConfig {
        let deviceName: String
        let serverId: String
        let serverIpInfo: String
        let listenHost: String
        let listenPort: Int32
        static func defaultConfig() -> P2PConfig {
            let host = WifiManager.shareInstance().getWiFiIPAddress() ?? "192.168.2.93"
            let port = WifiManager.shareInstance().getAvailablePort() ?? Int32(8080)
            return P2PConfig(
                deviceName: UIDevice.current.name,
                serverId: host,
                serverIpInfo: UUID().uuidString,
                listenHost: host,
                listenPort:port
            )
        }
    }
    
    private init() {
        globalP2PManager = self
        goSetCallBackMethod()
        SetupRootPath(getRootPaths().toGoString())
    }
    
    private func getRootPaths() -> String {
        let fileManager = FileManager.default
        guard let documentsDirectory = fileManager.urls(for: .documentDirectory, in: .userDomainMask).first else {
            Logger.info("Cannot access Documents directory.")
            return ""
        }
        
        return documentsDirectory.path
    }
    
    private func goSetCallBackMethod() {
        autoreleasepool {
            SetCallbackPasteXClipData(pasteXClipDataCallback)
        }

        autoreleasepool {
            SetCallbackUpdateClientStatus(updateClientStatusCallback)
        }

        autoreleasepool {
            SetCallbackMethodStartBrowseMdns(startBrowseMdnsCallback)
        }
        autoreleasepool {
            SetCallbackMethodStopBrowseMdns(stopBrowseMdnsCallback)
        }

        autoreleasepool {
            SetCallbackUpdateMultipleProgressBar(multipleProgressBarCallback)
        }

        autoreleasepool {
            SetCallbackGetAuthData(authDataCallback)
        }

        autoreleasepool {
            SetCallbackDIASStatus(diasStatusCallback)
        }

        autoreleasepool {
            SetCallbackMonitorName(monitorNameCallback)
        }

        autoreleasepool {
            SetCallbackNotifyBrowseResult(browseResultCallback)
        }

        autoreleasepool {
            SetCallbackRequestUpdateClientVersion { (version:UnsafeMutablePointer<CChar>?) in
                if let version = version {
                    let versionString = String(cString: version)
                    P2PManager.shared.newVersion = versionString
                    Logger.info("[GO][Callback] SetCallbackRequestUpdateClientVersion: \(versionString)")
                    DispatchQueue.main.async {
                        let userInfo: [String: Any] = ["version": versionString]
                        NotificationCenter.default.post(name: UpdateVersionNotification, object: nil, userInfo: userInfo)
                    }
                }
            }
        }
        
        Logger.info("[P2PManager] All callbacks have been set up successfully")
    }

    // All callback processing methods
    fileprivate func handlePasteXClipData(cText: UnsafeMutablePointer<CChar>?, cImage: UnsafeMutablePointer<CChar>?, cHtml: UnsafeMutablePointer<CChar>?) {
        var text = ""
        if let cText = cText {
            let textData = Data(bytes: cText, count: strlen(cText))

            if let utf8Text = String(data: textData, encoding: .utf8), !utf8Text.contains("�") && !utf8Text.contains("\u{FFFD}") {
                text = utf8Text
            } else {
                let encoding = CFStringConvertEncodingToNSStringEncoding(CFStringEncoding(CFStringEncodings.GB_18030_2000.rawValue))
                if let gb18030Text = String(data: textData, encoding: String.Encoding(rawValue: encoding)) {
                    text = gb18030Text
                } else if let utf16Text = String(data: textData, encoding: .utf16) {
                    text = utf16Text
                } else if let latinText = String(data: textData, encoding: .isoLatin1) {
                    text = latinText
                } else {
                    text = String(cString: cText)
                }
            }
        }

        let imageBase64 = cImage != nil ? String(cString: cImage!) : ""

        var html = ""
        if let cHtml = cHtml {
            let data = Data(bytes: cHtml, count: strlen(cHtml))
            let hasUTF8BOM = data.count >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF
            let utf8Data = hasUTF8BOM ? data.subdata(in: 3..<data.count) : data

            if let utf8String = String(data: utf8Data, encoding: .utf8), !utf8String.contains("�") && !utf8String.contains("\u{FFFD}") {
                html = utf8String
            } else {
                let encoding = CFStringConvertEncodingToNSStringEncoding(CFStringEncoding(CFStringEncodings.GB_18030_2000.rawValue))
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
                    html = String(cString: cHtml)
                }
            }
        }

        DispatchQueue.main.async {
            var image: UIImage? = nil
            if !imageBase64.isEmpty, let imageData = Data(base64Encoded: imageBase64, options: .ignoreUnknownCharacters) {
                image = UIImage(data: imageData)
            }

            let ret = ClipboardMonitor.shareInstance().setupClipboard(
                text: !text.isEmpty ? text : nil,
                image: image,
                html: !html.isEmpty ? html : nil
            )

            if ret {
                if !html.isEmpty {
                    if !text.isEmpty {
                        PictureInPictureManager.shared.showTextReceived(text)
                    } else if let img = image {
                        PictureInPictureManager.shared.showImageReceived(img)
                    }
                } else if let img = image {
                    PictureInPictureManager.shared.showImageReceived(img)
                } else if !text.isEmpty {
                    PictureInPictureManager.shared.showTextReceived(text)
                }
            }
        }
    }
    
    fileprivate func handleStartBrowseMdns(cinstance: UnsafeMutablePointer<CChar>?, cserviceType: UnsafeMutablePointer<CChar>?) {
        Logger.info("[GO][Callback] SetCallbackMethodStartBrowseMdns")
        if let cinstance = cinstance, let cserviceType = cserviceType {
            let instanceName = String(cString: cinstance)
            let serviceType = String(cString: cserviceType)
            DispatchQueue.main.async {
                BonjourService.shared.start(instanceName: instanceName, serviceType: serviceType)
            }
        }
    }
    
    fileprivate func handleStopBrowseMdns() {
        Logger.info("[GO][Callback] SetCallbackMethodStopBrowseMdns")
        DispatchQueue.main.async {
            BonjourService.shared.stop()
        }
    }
    
    fileprivate func handleUpdateClientStatus(clientJsonStr: UnsafeMutablePointer<CChar>?) {
        guard let clientJsonStr = clientJsonStr else {
            Logger.info("[GO][Callback] SetCallbackUpdateClientStatus received nil")
            return
        }
        let jsonString = String(cString: clientJsonStr)
        Logger.info("[GO][Callback] SetCallbackUpdateClientStatus - JSON length: \(jsonString.count)")

        // Process immediately without debouncing to avoid cancelling previous device updates
        handleSingleClientUpdate(jsonString)
    }

    private func handleSingleClientUpdate(_ jsonString: String) {
        guard let jsonData = jsonString.data(using: .utf8) else {
            Logger.info("[P2PManager][Err] Failed to convert string to data")
            return
        }

        let json = JSON(jsonData)

        let status = json["Status"].int ?? 0
        let timestamp = json["TimeStamp"].int64 ?? 0
        Logger.info("[P2PManager] Client Status Update - Status: \(status), Timestamp: \(timestamp)")

        guard let clientInfo = ClientInfo(json: json) else {
            Logger.info("[P2PManager][Err] Failed to parse client info from status update")
            return
        }

        Logger.info("[P2PManager] Client Update - ID: \(clientInfo.id), IP: \(clientInfo.ip), Name: \(clientInfo.name), Platform: \(clientInfo.deviceType), Status: \(status)")

        let oldCount = mClientList.count
        
        if status == 1 {
            if let existingIndex = mClientList.firstIndex(where: { $0.id == clientInfo.id }) {
                mClientList[existingIndex] = clientInfo
                Logger.info("[P2PManager] Updated existing client: \(clientInfo.name)")
            } else {
                mClientList.append(clientInfo)
                Logger.info("[P2PManager] Added new client: \(clientInfo.name)")
            }
        } else {
            if let existingIndex = mClientList.firstIndex(where: { $0.id == clientInfo.id }) {
                mClientList.remove(at: existingIndex)
                Logger.info("[P2PManager] Removed client: \(clientInfo.name)")
            }
        }
        
        let newCount = mClientList.count
        Logger.info("[P2PManager] Client list updated: \(oldCount) -> \(newCount)")

        let dictArray = mClientList.map { $0.toDictionary() }
        let clientListJson = JSON(dictArray)
        if let jsonString = clientListJson.rawString() {
            UserDefaults.set(forKey: .DEVICE_CLIENTS, value: jsonString, type: .group)
        }
        NotificationCenter.default.post(name: UpdateClientListSuccessNotification, object: nil, userInfo: nil)
    }

    fileprivate func handleFoundPeer() {
        Logger.info("[GO][Callback] SetCallbackMethodFoundPeer (deprecated)")
        self.updateClientList()
    }

    fileprivate func handleEvent(event: Int32) {
        // Logger.info("[GO][Callback] SetLogMessageCallback \(event)")
    }
    
    fileprivate func handleMultipleProgressBar(ip: UnsafeMutablePointer<CChar>?, id: UnsafeMutablePointer<CChar>?, deviceName: UnsafeMutablePointer<CChar>?, currentfileName: UnsafeMutablePointer<CChar>?, recvFileCnt: UInt32, totalFileCnt: UInt32, currentFileSize: UInt64, totalSize: UInt64, recvSize: UInt64, timestamp: UInt64) {
        if let ip = ip, let id = id, let deviceName = deviceName, let currentfileName = currentfileName {
            let ipString = String(cString: ip)
            let idString = String(cString: id)
            let deviceNameString = String(cString: deviceName)
            let currentfileNameString = String(cString: currentfileName)
            var downloadParams: [String: Any] = [:]
            let downloadItem = DownloadItem()
            downloadItem.fileId = idString
            downloadItem.ip = ipString
            downloadItem.timestamp = TimeInterval(timestamp)
            downloadItem.uuid = "\(timestamp)"
            downloadItem.deviceName = deviceNameString
            downloadItem.recvFileCnt = recvFileCnt
            downloadItem.totalFileCnt = totalFileCnt
            downloadItem.currentFileSize = currentFileSize
            downloadItem.currentfileName = currentfileNameString
            downloadItem.receiveSize = recvSize
            downloadItem.totalSize = totalSize
            downloadItem.progress = Float(recvSize) * 100.0 / Float(totalSize)
            downloadItem.isMutip = true
            if recvSize == totalSize {
                downloadItem.finishTime = Date().timeIntervalSince1970
            }
            downloadParams["download"] = downloadItem
            NotificationCenter.default.post(name: ReceiveFuleSuccessNotification, object: nil, userInfo: downloadParams)
        }
    }
    
    fileprivate func handleGetAuthData(clientIndex: UInt32) -> UnsafeMutablePointer<CChar>? {
        Logger.info("[GO][Callback] SetGetAuthDataCallback clientIndex: \(clientIndex)")
        DispatchQueue.main.async {
            NotificationCenter.default.post(name: UpdateVerifyClientIndexChangedNotification, object: nil, userInfo: ["clientIndex": clientIndex])
        }
        let authString = self.screenData.isEmpty ? "" : self.screenData
        return authString.withCString { cString in
            let length = strlen(cString) + 1
            let buffer = UnsafeMutablePointer<CChar>.allocate(capacity: length)
            strcpy(buffer, cString)
            return buffer
        }
    }
    
    fileprivate func handleDIASStatus(status: UInt32) {
        if let status = DiassStatus(rawValue: Int(status)) {
            Logger.info("[GO][Callback] SetSetDIASStatusCallback status: \(status)")
            self.cStatus = status
            DispatchQueue.main.async {
                NotificationCenter.default.post(name: UpdateDiassStatusChangedNotification, object: nil, userInfo: ["status": status])
            }
        }
    }
    
    fileprivate func handleMonitorName(name: UnsafeMutablePointer<CChar>?) {
        if let name = name {
            let monitorName = String(cString: name)
            self.monitorName = monitorName
            Logger.info("[GO][Callback] SetSetMonitorNameCallback monitorName: \(monitorName)")
            DispatchQueue.main.async {
                NotificationCenter.default.post(name: UpdateMonitorNameChangedNotification, object: nil, userInfo: ["monitorName": monitorName])
            }
        }
    }
    
    fileprivate func handleBrowseResult(monitorName: UnsafeMutablePointer<CChar>?, instance: UnsafeMutablePointer<CChar>?, ip: UnsafeMutablePointer<CChar>?, version: UnsafeMutablePointer<CChar>?, timestamp: UInt64) {
        if let monitorName = monitorName, let instance = instance, let ip = ip, let version = version {
            let monitorName = String(cString: monitorName)
            let instance = String(cString: instance)
            let ip = String(cString: ip)
            let version = String(cString: version)
            Logger.info("[GO][Callback] SetCallbackNotifyBrowseResult monitorName: \(monitorName), instance: \(instance), ip: \(ip), version: \(version), timestamp: \(timestamp)")
            let lanServiceInfo = LanServiceInfo(monitorName: monitorName, instance: instance, ip: ip, version: version, timestamp: timestamp)
    
            if let existingIndex = self.mServiceList.firstIndex(where: { $0.instance == lanServiceInfo.instance }) {
                self.mServiceList[existingIndex] = lanServiceInfo
            } else {
                self.mServiceList.append(lanServiceInfo)
            }
            DispatchQueue.main.async { [weak self] in
                guard let self = self else { return }
                NotificationCenter.default.post(name: UpdateLanserviceListNotification, object: nil, userInfo: ["serviceList": self.mServiceList])
            }
        }
    }

    func startP2PService(config: P2PConfig = .defaultConfig()) {
        guard !isRunning else {
            Logger.info("P2P service is already running")
            SetHostListenAddr(config.serverId.toGoString(), GoInt(config.listenPort))
            return
        }
        p2pQueue = DispatchQueue(label: "com.crossshare.p2p", qos: .userInitiated)
        p2pQueue?.async { [weak self] in
            guard let self = self else { return }
            self.isRunning = true
            Logger.info("Version: \(version)")
            Logger.info("Device Name: \(config.deviceName)")
            Logger.info("Server ID: \(config.serverId)")
            Logger.info("Server IP: \(config.serverIpInfo)")
            Logger.info("Listen Host: \(config.listenHost)")
            Logger.info("Starting P2P service on port \(config.listenPort)...")
            do {
                let deviceName = config.deviceName.toGoString()
                let serverId = config.serverId.toGoString()
                let serverIpInfo = config.serverIpInfo.toGoString()
                let listenHost = config.listenHost.toGoString()
                try MainInit(deviceName, serverId, serverIpInfo, listenHost, GoInt(config.listenPort))
            } catch {
                Logger.info("Not me error")
            }
        }
    }
    
    func stopP2PService() {
        guard isRunning else {
            Logger.info("P2P service is not running")
            return
        }
        p2pQueue = nil
        isRunning = false
    }
    
    var serviceStatus: String {
        return isRunning ? "Running" : "Stopped"
    }
    
    var version: String {
        guard let cver = GetVersion(), let cbuildDate = GetBuildDate() else {
            return ""
        }
        
        let ver = String(cString: cver)
        let buildDate = String(cString: cbuildDate)
        FreeCString(cver)
        FreeCString(cbuildDate)
        return "\(ver) (\(buildDate))"
    }
    
    var majorVersion: String {
        guard let cver = GetVersion() else {
            return ""
        }
        let ver = String(cString: cver)
        FreeCString(cver)
        return "\(ver)"
    }
    
    var clientList: [ClientInfo] {
        return mClientList
    }
    
    var serviceList: [LanServiceInfo] {
        return mServiceList
    }
    
    var ip:String {
        guard let ip = WifiManager.shareInstance().getWiFiIPAddress() else { return "0.0.0.0" }
        return ip
    }
    
    var deviceName:String {
        return UIDevice.current.name
    }
    
    var deviceDiasId:String {
        if let diasId = UserDefaults.get(forKey: .DEVICECONFIG_DIAS_ID) {
            return diasId
        }
        return ""
    }
    
    var deviceSrc:String {
        if let src = UserDefaults.get(forKey: .DEVICECONFIG_SRC) {
            return src
        }
        return ""
    }
    
    var devicePort:String {
        if let port = UserDefaults.get(forKey: .DEVICECONFIG_PORT) {
            return port
        }
        return ""
    }
    
    func getClientInfo(id: String) -> ClientInfo? {
        for clientInfo in clientList {
            if clientInfo.id == id {
                return clientInfo
            }
        }
        return nil
    }
    
    private func updateClientListDebounced() {
        updateClientListWorkItem?.cancel()
        let workItem = DispatchWorkItem { [weak self] in
            self?.updateClientList()
        }
        updateClientListWorkItem = workItem
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.2, execute: workItem)
    }
    
    private func updateClientList() {
        guard let cstr = GetClientListEx() else {
            return
        }
        let str = String(cString: cstr)
        FreeCString(cstr)
        handleClientList(str)
    }
    
    private func handleClientList(_ clientListStr: String) {
        var newClientList: [ClientInfo] = []
        var seenClients = Set<String>()

        guard let jsonData = clientListStr.data(using: .utf8) else {
            Logger.info("[P2PManager][Err] Failed to convert string to data")
            return
        }

        let json = JSON(jsonData)

        let selfID = json["ID"].string ?? ""
        let selfIP = json["IpAddr"].string ?? ""
        let timestamp = json["TimeStamp"].int64 ?? 0
        Logger.info("[P2PManager] Self Info - ID: \(selfID), IP: \(selfIP), Timestamp: \(timestamp)")

        guard let clientListArray = json["ClientList"].array else {
            Logger.info("[P2PManager][Err] ClientList is not an array or missing")
            return
        }

        for (index, clientJson) in clientListArray.enumerated() {
            guard let clientInfo = ClientInfo(json: clientJson) else {
                Logger.info("[P2PManager][Err] Failed to parse client at index \(index)")
                continue
            }

            let uniqueKey = "\(clientInfo.ip)_\(clientInfo.id)"
            if seenClients.contains(uniqueKey) {
                Logger.info("[P2PManager] Skipping duplicate client: IP:[\(clientInfo.ip)], ID:[\(clientInfo.id)]")
                continue
            }
            seenClients.insert(uniqueKey)

            newClientList.append(clientInfo)
            Logger.info("[P2PManager] update client list[\(index)]: IP:[\(clientInfo.ip)], Name:[\(clientInfo.name)], ID:[\(clientInfo.id)], Platform:[\(clientInfo.deviceType)], SourcePort:[\(clientInfo.sourcePortType)], Version:[\(clientInfo.version)]")
        }

        if !areClientListsEqual(mClientList, newClientList) {
            mClientList = newClientList
            let dictArray = mClientList.map { $0.toDictionary() }
            let json = JSON(dictArray)
            if let jsonString = json.rawString() {
                UserDefaults.set(forKey: .DEVICE_CLIENTS, value: jsonString, type: .group)
            }
            NotificationCenter.default.post(name: UpdateClientListSuccessNotification, object: nil, userInfo: nil)
        }
    }
    
    private func areClientListsEqual(_ list1: [ClientInfo], _ list2: [ClientInfo]) -> Bool {
        guard list1.count == list2.count else { return false }
        for i in 0..<list1.count {
            if list1[i].ip != list2[i].ip || list1[i].id != list2[i].id {
                return false
            }
        }
        return true
    }
    
    private func postShareExtensionnotifaction(userInfo: [String: Any]? = nil) {
        Logger.info("[P2PManager] postShareExtensionnotifaction")
        let notifier = CrossProcessNotifier(appGroupID: UserDefaults.groupId)
        let dataUpdateNotification = "com.realtek.crossshare.dataDidUpdate"
        notifier.post(name: dataUpdateNotification, userInfo: userInfo)
    }
    
    func setupDeviceConfig(_ diasId: String,_ src: Int,_ port: Int) {
        DispatchQueue.main.async {
            Logger.info("[P2PManager] setup device config. DIAS_ID:[\(diasId)], (Source, Port):[\(src), \(port)]")
            SetDIASID(diasId.toGoString())
        }
    }
    
    func setFileListsDropRequest(filePath: String,taskId:String) {
        if filePath.isEmpty{
            Logger.info("[P2PManager] file drop request failed: Invalid params")
            return
        }
        Logger.info("[P2PManager] file drop request json: \(filePath)")
        DispatchQueue.main.async {
            let status = SendMultiFilesDropRequest(filePath.toGoString())
            var statusString = ""
            switch status {
            case 1:
                statusString = "Starting transfer"
                // Logger.info("[P2PManager] file drop request started")
                self.notifyTransferSuccess(taskId: taskId)
            case 2:
                statusString = "Transfer failed (Error code:2)"
                // Logger.info("[P2PManager] file drop request failed: params error")
                self.notifyTransferError(taskId: taskId, error: "Params error")
            case 3:
                statusString = "Cannot send: Sending in progress"
                // Logger.info("[P2PManager] file drop request failed: sended in progress")
                self.notifyTransferError(taskId: taskId, error: "Sending in progress")
            case 4:
                statusString = "Cannot send: Receiving in progress"
                // Logger.info("[P2PManager] file drop request failed: received in progress")
                self.notifyTransferError(taskId: taskId, error: "Receiving in progress")
            case 5:
                statusString = "Transfer failed (Error code:5)"
                // Logger.info("[P2PManager] file drop request failed: callback not set")
                self.notifyTransferError(taskId: taskId, error: "App not inited")
            default:
                Logger.info("[P2PManager] file drop request failed: Unknown status:\(status))")
                self.notifyTransferError(taskId: taskId, error: "Unknown status:\(status)")
            }
            let data: [String: Any] = [
                "message": statusString,
                "timestamp": Date(),
                "status":Int(status)
            ]
            self.postShareExtensionnotifaction(userInfo: data)
        }
    }
    
    func setScreenData(screen:String) {
        Logger.info("[P2PManager] getScreenData \(screen)")
        self.screenData = screen
    }
    
    func detectPluginEventCallback(isPlugin:Bool, productName:String? = "") {
        Logger.info("[P2PManager] detectPluginEventCallback: \(isPlugin) \(productName ?? "")")
        DispatchQueue.main.async {
            SetDetectPluginEvent(isPlugin ? GoUint8(1) : GoUint8(0))
            NotificationCenter.default.post(name: UpdateLanserviceChangedNotification, object: nil, userInfo: ["isDetect":isPlugin])
        }
    }
    
    func setCancelFileTransfer(ipPort: String,clientID:String,timeStamp:UInt64) {
        if ipPort.isEmpty || clientID.isEmpty || timeStamp <= 0 {
            Logger.info("[P2PManager] Cancel File Transfer failed: Invalid params")
            return
        }
        Logger.info("[P2PManager] Cancel File Transfer: IPPort:[\(ipPort)], ClientID:[\(clientID)], TimeStamp:[\(timeStamp)]")
        DispatchQueue.main.async {
            SetCancelFileTransfer(ipPort.toGoString(), clientID.toGoString(), GoUint64(timeStamp))
        }
    }
    
    func comfirmLanServer(instance: String) {
        Logger.info("[P2PManager] ConfirmLanServer instance: \(instance)")
        if instance.isEmpty {
            Logger.info("[P2PManager] ConfirmLanServer failed: Invalid parameters")
            return
        }
        DispatchQueue.main.async {
            WorkerConnectLanServer(instance.toGoString())
        }
    }
    
    func browseLanService() {
        Logger.info("[P2PManager] BrowseLanService")
        DispatchQueue.main.async {
            BrowseLanServer()
        }
    }
    
    func removeServerList() {
        mServiceList.removeAll()
        DispatchQueue.main.async { [weak self] in
            guard let self = self else { return }
            NotificationCenter.default.post(name: UpdateLanserviceListNotification, object: nil, userInfo: ["serviceList": self.mServiceList])
        }
    }
    
}

extension P2PManager {
    func setTransferCallbacks(
        taskId: String,
        onProgress: @escaping (Float, String) -> Void,
        onSuccess: @escaping () -> Void,
        onError: @escaping (String) -> Void
    ) {
        transferCallbacks[taskId] = TransferCallbacks(
            onProgress: onProgress,
            onSuccess: onSuccess,
            onError: onError
        )
    }
    
    private func notifyTransferProgress(taskId: String, progress: Float, currentFile: String) {
        if let callbacks = transferCallbacks[taskId] {
            DispatchQueue.main.async {
                callbacks.onProgress(progress, currentFile)
            }
        }
    }
    
    private func notifyTransferSuccess(taskId: String) {
        if let callbacks = transferCallbacks[taskId] {
            DispatchQueue.main.async {
                callbacks.onSuccess()
            }
        }
        transferCallbacks.removeValue(forKey: taskId)
    }
    
    private func notifyTransferError(taskId: String, error: String) {
        if let callbacks = transferCallbacks[taskId] {
            DispatchQueue.main.async {
                callbacks.onError(error)
            }
        }
        transferCallbacks.removeValue(forKey: taskId)
    }
}
