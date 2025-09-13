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

private weak var globalP2PManager: P2PManager?

// Global C compatible callback function
private let textCallback: @convention(c) (UnsafeMutablePointer<CChar>?) -> Void = { cText in
    globalP2PManager?.handleTextReceived(cText: cText)
}

private let imageCallback: @convention(c) (UnsafeMutablePointer<CChar>?) -> Void = { cText in
    globalP2PManager?.handleImageReceived(cText: cText)
}

private let startBrowseMdnsCallback: @convention(c) (UnsafeMutablePointer<CChar>?, UnsafeMutablePointer<CChar>?) -> Void = { cinstance, cserviceType in
    globalP2PManager?.handleStartBrowseMdns(cinstance: cinstance, cserviceType: cserviceType)
}

private let stopBrowseMdnsCallback: @convention(c) () -> Void = {
    globalP2PManager?.handleStopBrowseMdns()
}

private let foundPeerCallback: @convention(c) () -> Void = {
    globalP2PManager?.handleFoundPeer()
}

private let fileConfirmCallback: @convention(c) (UnsafeMutablePointer<CChar>?, UnsafeMutablePointer<CChar>?, UnsafeMutablePointer<CChar>?, Int64) -> Void = { cid, cplatform, cfileName, cfileSize in
    globalP2PManager?.handleFileConfirm(cid: cid, cplatform: cplatform, cfileName: cfileName, cfileSize: cfileSize)
}

private let eventCallback: @convention(c) (Int32) -> Void = { event in
    globalP2PManager?.handleEvent(event: event)
}

private let progressBarCallback: @convention(c) (UnsafeMutablePointer<CChar>?, UnsafeMutablePointer<CChar>?, UInt64, UInt64, UInt64) -> Void = { clientId, fileName, recvSize, total, timeStamp in
    globalP2PManager?.handleProgressBar(clientId: clientId, fileName: fileName, recvSize: recvSize, total: total, timeStamp: timeStamp)
}

private let multipleProgressBarCallback: @convention(c) (UnsafeMutablePointer<CChar>?, UnsafeMutablePointer<CChar>?, UnsafeMutablePointer<CChar>?, UnsafeMutablePointer<CChar>?, UInt32, UInt32, UInt64, UInt64, UInt64, UInt64) -> Void = { ip, id, deviceName, currentfileName, recvFileCnt, totalFileCnt, currentFileSize, totalSize, recvSize, timestamp in
    globalP2PManager?.handleMultipleProgressBar(ip: ip, id: id, deviceName: deviceName, currentfileName: currentfileName, recvFileCnt: recvFileCnt, totalFileCnt: totalFileCnt, currentFileSize: currentFileSize, totalSize: totalSize, recvSize: recvSize, timestamp: timestamp)
}

private let authDataCallback: @convention(c) () -> UnsafeMutablePointer<CChar>? = {
    return globalP2PManager?.handleGetAuthData()
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
    public var cStatus: DiassStatus = .WaitConnecting
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
    
    private static func getPasteboard() -> UIPasteboard {
        ClipboardMonitor.shareInstance().skipLocalChecking()
        return UIPasteboard.general
    }
    
    private func goSetCallBackMethod() {
        autoreleasepool {
            SetCallbackMethodText(textCallback)
        }
        autoreleasepool {
            SetCallbackMethodImage(imageCallback)
        }
        autoreleasepool {
            SetCallbackMethodStartBrowseMdns(startBrowseMdnsCallback)
        }
        autoreleasepool {
            SetCallbackMethodStopBrowseMdns(stopBrowseMdnsCallback)
        }
        autoreleasepool {
            SetCallbackMethodFoundPeer(foundPeerCallback)
        }
        autoreleasepool {
            SetCallbackMethodFileConfirm(fileConfirmCallback)
        }
        autoreleasepool {
            SetEventCallback(eventCallback)
        }
        
        autoreleasepool {
            SetCallbackUpdateProgressBar(progressBarCallback)
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
    fileprivate func handleTextReceived(cText: UnsafeMutablePointer<CChar>?) {
        if let cText = cText {
            let text = String(cString: cText)
            DispatchQueue.main.async {
                Logger.info("go service send Text callback: \(text)")
                let pasteboard = P2PManager.getPasteboard()
                pasteboard.string = text
                
                PictureInPictureManager.shared.showTextReceived(text)
            }
        }
    }
    
    fileprivate func handleImageReceived(cText: UnsafeMutablePointer<CChar>?) {
        if let cText = cText {
            let copiedText = String(cString: cText)
            DispatchQueue.main.async {
                if let imageData = Data(base64Encoded: copiedText, options: .ignoreUnknownCharacters) {
                    if let image = UIImage(data: imageData) {
                        let pasteboard = P2PManager.getPasteboard()
                        pasteboard.image = image
                        
                        PictureInPictureManager.shared.showImageReceived(image)
                    } else {
                        Logger.info("Failed to decode Base64 string into image")
                     }
                } else {
                    Logger.info("Failed to decode Base64 string")
                 }
             }
        } else {
            Logger.info("[P2PManager] Image callback received with nil cText")
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
    
    fileprivate func handleFoundPeer() {
        Logger.info("[GO][Callback] SetCallbackMethodFoundPeer")
        self.updateClientList()
    }
    
    fileprivate func handleFileConfirm(cid: UnsafeMutablePointer<CChar>?, cplatform: UnsafeMutablePointer<CChar>?, cfileName: UnsafeMutablePointer<CChar>?, cfileSize: Int64) {
        if let cid = cid, let cplatform = cplatform, let cfileName = cfileName {
            let id = String(cString: cid)
            let platform = String(cString: cplatform)
            let fileName = String(cString: cfileName)
            
            Logger.info("[P2PManager][FileDrop] ID:[\(id)], Platform:[\(platform)], FileNmae:[\(fileName)], FileSize:[\(cfileSize)]")
            let isReceive = 1
            SetFileDropResponse(fileName.toGoString(), id.toGoString(), GoUint8(isReceive))
        }
    }
    
    fileprivate func handleEvent(event: Int32) {
        // Logger.info("[GO][Callback] SetLogMessageCallback \(event)")
    }
    
    fileprivate func handleProgressBar(clientId: UnsafeMutablePointer<CChar>?, fileName: UnsafeMutablePointer<CChar>?, recvSize: UInt64, total: UInt64, timeStamp: UInt64) {
        if let clientId = clientId, let fileName = fileName {
            let clientIdString = String(cString: clientId)
            let fileNameString = String(cString: fileName)
            var downloadParams: [String: Any] = [:]
            let downloadItem = DownloadItem()
            downloadItem.timestamp = TimeInterval(timeStamp)
            downloadItem.uuid = "\(timeStamp)"
            downloadItem.deviceName = self.getClientInfo(id: clientIdString)?.name
            downloadItem.fileId = clientIdString
            downloadItem.totalFileCnt = 1
            downloadItem.currentfileName = fileNameString
            downloadItem.receiveSize = recvSize
            downloadItem.totalSize = total
            downloadItem.progress = Float(recvSize) * 100.0 / Float(total)
            if recvSize == total {
                downloadItem.finishTime = Date().timeIntervalSince1970
            }
            downloadParams["download"] = downloadItem
            NotificationCenter.default.post(name: ReceiveFuleSuccessNotification, object: nil, userInfo: downloadParams)
        }
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
    
    fileprivate func handleGetAuthData() -> UnsafeMutablePointer<CChar>? {
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
        guard let cstr = GetClientList() else {
            return
        }
        let str = String(cString: cstr)
        FreeCString(cstr)
        handleClientList(str)
    }
    
    private func handleClientList(_ clientListStr: String) {
        var newClientList: [ClientInfo] = []
        var seenClients = Set<String>()
        
        let clientListArr = clientListStr.components(separatedBy: ",")
        for (index, clientStr) in clientListArr.enumerated() {
            let clientArr = clientStr.components(separatedBy: "#")
            if clientArr.count != 4 {
                Logger.info("[P2PManager][Err] Invalid client count = \(clientArr.count)")
                continue
            }
            
            let ip = clientArr[0]
            let id = clientArr[1]
            let name = clientArr[2]
            let deviceType = clientArr[3]
            if ip.isEmpty || id.isEmpty || name.isEmpty || deviceType.isEmpty {
                continue
            }
            
            let uniqueKey = "\(ip)_\(id)"
            if seenClients.contains(uniqueKey) {
                Logger.info("[P2PManager] Skipping duplicate client: IP:[\(ip)], ID:[\(id)]")
                continue
            }
            seenClients.insert(uniqueKey)
            
            let clientInfo = ClientInfo(ip: ip, id: id, name: name,deviceType: deviceType)
            newClientList.append(clientInfo)
            Logger.info("[P2PManager] update client list[\(index)]: IP:[\(ip)], Name:[\(name)], ID:[\(id)],deviceType:[\(deviceType)]")
        }
        
        if !areClientListsEqual(mClientList, newClientList) {
            mClientList = newClientList
            let dictArray = mClientList.map { $0.toDictionary() }
            let json = JSON(dictArray)
            if let jsonString = json.rawString() {
                UserDefaults.set(forKey: .DEVICE_CLIENTS, value: jsonString,type: .group)
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
            SetSrcAndPort(GoInt(src), GoInt(port))
        }
    }
    
    func setFileDropRequest(filePath: String, id: String, fileSize: Int64) {
        if filePath.isEmpty || id.isEmpty || fileSize <= 0{
            Logger.info("[P2PManager] file drop request failed: Invalid params")
        }
        Logger.info("[P2PManager] file drop request json: \(filePath), id: \(id), fileSize: \(fileSize)")
        DispatchQueue.main.async {
            SendFileDropRequest(filePath.toGoString(), id.toGoString(), GoInt(fileSize))
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
