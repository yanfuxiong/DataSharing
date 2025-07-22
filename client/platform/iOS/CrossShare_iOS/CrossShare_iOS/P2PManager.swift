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

class P2PManager {
    static let shared = P2PManager()
    private var p2pQueue: DispatchQueue?
    private var isRunning: Bool = false
    private var mClientList: [ClientInfo] = []
    private var screenData: String = ""

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
        goSetCallBackMethod()
        SetupRootPath(getRootPaths().toGoString())
    }

    private func getRootPaths() -> String {
        let fileManager = FileManager.default
        guard let documentsDirectory = fileManager.urls(for: .documentDirectory, in: .userDomainMask).first else {
            Logger.info("❌ Cannot access Documents directory.")
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
            SetCallbackMethodText { (cText: UnsafeMutablePointer<CChar>?) in
                if let cText = cText {
                    let text = String(cString: cText)
                    Logger.info("go service send Text callback: \(text)")
                    let pasteboard = P2PManager.getPasteboard()
                    pasteboard.string = text
                }
            }
        }
        autoreleasepool {
            SetCallbackMethodImage { (cText: UnsafeMutablePointer<CChar>?) in
                if let cText = cText {
                    let copiedText = String(cString: cText)
                    DispatchQueue.main.async {
                        do {
                            if let imageData = Data(base64Encoded: copiedText, options: .ignoreUnknownCharacters) {
                                if let image = UIImage(data: imageData) {
                                    let pasteboard = P2PManager.getPasteboard()
                                    pasteboard.image = image
                                } else {
                                    Logger.info("Failed to decode Base64 string into image")
                                }
                            } else {
                                Logger.info("Failed to decode Base64 string")
                            }
                        } catch {
                            Logger.info("Failed to base64ToImage: \(error.localizedDescription)")
                        }
                    }
                }
            }
        }
        autoreleasepool {
            SetCallbackMethodStartBrowseMdns { (cinstance: UnsafeMutablePointer<CChar>?, cserviceType: UnsafeMutablePointer<CChar>?) in
                Logger.info("[GO][Callback] SetCallbackMethodStartBrowseMdns")
                if let cinstance = cinstance, let cserviceType = cserviceType {
                    let instanceName = String(cString: cinstance)
                    let serviceType = String(cString: cserviceType)
                    DispatchQueue.main.async {
                        BonjourService.shared.start(instanceName: instanceName, serviceType: serviceType)
                    }
                }
            }
        }
        autoreleasepool {
            SetCallbackMethodStopBrowseMdns {
                Logger.info("[GO][Callback] SetCallbackMethodStopBrowseMdns")
                DispatchQueue.main.async {
                    BonjourService.shared.stop()
                }
            }
        }
        autoreleasepool {
            SetCallbackMethodFoundPeer {
                Logger.info("[GO][Callback] SetCallbackMethodFoundPeer")
                // TODO: update client list on UI
                P2PManager.shared.updateClientList()
            }
        }
        autoreleasepool {
            SetCallbackMethodFileConfirm { (cid: UnsafeMutablePointer<CChar>?, cplatform: UnsafeMutablePointer<CChar>?, cfileName: UnsafeMutablePointer<CChar>?, cfileSize: Int64) in
                Logger.info("[GO][Callback] SetCallbackMethodFileConfirm")
                if let cid = cid, let cplatform = cplatform, let cfileName = cfileName {
                    let id = String(cString: cid)
                    let platform = String(cString: cplatform)
                    let fileName = String(cString: cfileName)

                    // TODO: Show dialog to request receive file or not
                    Logger.info("[P2PManager][FileDrop] ID:[\(id)], Platform:[\(platform)], FileNmae:[\(fileName)], FileSize:[\(cfileSize)]")
                    let isReceive = 1
                    SetFileDropResponse(fileName.toGoString(), id.toGoString(), GoUint8(isReceive))
                }
            }
        }
        autoreleasepool {
            SetEventCallback { event in
                Logger.info("[GO][Callback] SetLogMessageCallback \(event)")
            }
        }
        
        autoreleasepool {
            SetCallbackUpdateProgressBar { (clientId: UnsafeMutablePointer<CChar>?, fileName: UnsafeMutablePointer<CChar>?, recvSize: UInt64, total: UInt64,timeStamp: UInt64) in
                if let clientId = clientId,let fileName = fileName {
                    let clientIdString = String(cString: clientId)
                    let fileNameString = String(cString: fileName)
//                    Logger.info("[GO][Callback] SetCallbackUpdateProgressBar id:\(fileIdString) fileName:\(fileNameString) recvSize:\(recvSize) bytes total:\(total) bytes")
                    var downloadParams = [:]
                    let downloadItem = DownloadItem()
                    downloadItem.timestamp = TimeInterval(timeStamp)
                    downloadItem.uuid = "\(timeStamp)"
                    downloadItem.deviceName = P2PManager.shared.getClientInfo(id: clientIdString)?.name
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
        }
        
        autoreleasepool {
            SetCallbackUpdateMultipleProgressBar {(ip: UnsafeMutablePointer<CChar>?, id: UnsafeMutablePointer<CChar>?, deviceName: UnsafeMutablePointer<CChar>?, currentfileName: UnsafeMutablePointer<CChar>?, recvFileCnt:UInt32, totalFileCnt:UInt32, currentFileSize :UInt64, totalSize: UInt64, recvSize: UInt64, timestamp: UInt64) in
                if let ip = ip,let id = id,let deviceName = deviceName,let currentfileName = currentfileName {
                    let ipString = String(cString: ip)
                    let idString = String(cString: id)
                    let deviceNameString = String(cString: deviceName)
                    let currentfileNameString = String(cString: currentfileName)
//                    Logger.info("[GO][Callback] SetCallbackUpdateProgressBar ip:\(ipString) id:\(idString) deviceName:\(deviceNameString) currentfileName:\(currentfileNameString) recvFileCnt:\(recvFileCnt) totalFileCnt:\(totalFileCnt) currentFileSize:\(currentFileSize) totalSize:\(totalSize) recvSize:\(recvSize) timestamp:\(timestamp)")
                    var downloadParams = [:]
                    let downloadItem = DownloadItem()
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
        }
        
        autoreleasepool {
            SetCallbackGetAuthData {
                Logger.info("[GO][Callback] SetCallbackGetAuthData")
                let authString = P2PManager.shared.screenData.isEmpty ? "" : P2PManager.shared.screenData
                return authString.withCString { cString in
                    let length = strlen(cString) + 1
                    let buffer = UnsafeMutablePointer<CChar>.allocate(capacity: length)
                    strcpy(buffer, cString)
                    return buffer
                }
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

    var clientList: [ClientInfo] {
        return mClientList
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

    private func updateClientList() {
        guard let cstr = GetClientList() else {
            return
        }
        let str = String(cString: cstr)
        FreeCString(cstr)
        handleClientList(str)
    }

    private func handleClientList(_ clientListStr: String) {
        mClientList.removeAll()
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

            let clientInfo = ClientInfo(ip: ip, id: id, name: name,deviceType: deviceType)
            mClientList.append(clientInfo)
            Logger.info("[P2PManager] update client list[\(index)]: IP:[\(ip)], Name:[\(name)], ID:[\(id)],deviceType:[\(deviceType)]")
        }
        let dictArray = mClientList.map { $0.toDictionary() }
        let json = JSON(dictArray)
        if let jsonString = json.rawString() {
            UserDefaults.set(forKey: .DEVICE_CLIENTS, value: jsonString,type: .group)
        }
        NotificationCenter.default.post(name: UpdateClientListSuccessNotification, object: nil, userInfo: nil)
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
        DispatchQueue.main.async {
            SendFileDropRequest(filePath.toGoString(), id.toGoString(), GoInt(fileSize))
        }
    }
    
    func setFileListsDropRequest(filePath: String) {
        if filePath.isEmpty{
            Logger.info("[P2PManager] file drop request failed: Invalid params")
            return
        }
        Logger.info("[P2PManager] file drop request json: \(filePath)")
        DispatchQueue.main.async {
            SendMultiFilesDropRequest(filePath.toGoString())
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
        }
    }
}
