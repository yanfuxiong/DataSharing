//
//  GoCallBackManager.swift
//  CrossShareHelper
//
//  Created by ts on 2025/9/16.
//

import Foundation

class GoCallbackManager {
    weak var delegate: CrossShareHelperXPCDelegate?
    private let logger = XPCLogger.shared
    
    private var macToDisplayID: [String: CGDirectDisplayID] = [:]
    private var currentDIASMac: String?
    
    init() {
        
    }
    
    static let shared = GoCallbackManager()
    
    func handleAuthIndex(_ index: Int) {
        logger.log("Received auth index from Go service: \(index)", level: .info)
        guard index >= 0 else {
            logger.log("Invalid auth index received: \(index)", level: .error)
            return
        }
        self.delegate?.didReceiveAuthRequest?(index: UInt32(index))
    }
    
    func handleDIASStatus(_ status: Int) {
        logger.log("Received DIAS status from Go service: \(status)", level: .info)
        DispatchQueue.main.async { [weak self] in
            self?.delegate?.didReceiveDIASStatus?(status)
        }
    }
    
    func handReceiveFilesData(_ userInfo: [String: Any]){
        DispatchQueue.main.async { [weak self] in
            self?.delegate?.didReceiveFilesData?(userInfo)
        }
    }
    
    func handleClientStatus(_ clientJsonStr: String) {
        logger.log("Received client status JSON: \(clientJsonStr)", level: .info)
        guard let jsonData = clientJsonStr.data(using: .utf8) else {
            logger.log("Failed to convert client status string to data", level: .error)
            return
        }
        
        do {
            if let deviceData = try JSONSerialization.jsonObject(with: jsonData, options: []) as? [String: Any] {
                logger.log("Parsed device data: \(deviceData)", level: .info)
                DispatchQueue.main.async { [weak self] in
                    self?.delegate?.didReceiveDeviceData?(deviceData: deviceData)
                    self?.logger.log("Sent device data to GUI through XPC", level: .info)
                }
            } else {
                logger.log("Failed to parse JSON as dictionary", level: .error)
            }
        } catch {
            logger.log("Failed to parse client status JSON: \(error)", level: .error)
        }

//        if let clientListData = GetClientList() {
//            let clientListStr = String(cString: clientListData)
//            logger.log("Client list from GetClientList: \(clientListStr)", level: .info)
//        }
    }
    
    func handleRequestSourceAndPort() {
        logger.log("Received request for source and port", level: .info)
        guard let displayID = getCurrentActiveDisplayID() else {
            logger.log("No active display found", level: .error)
            return
        }
        if let portInfo = HelperDisplayManager.shared.getPortInfo(for: displayID) {
            logger.log("Got source: \(portInfo.source), port: \(portInfo.port)", level: .info)
            SetDIASSourceAndPort(UInt32(portInfo.source), UInt32(portInfo.port))
        } else {
            logger.log("Failed to get source and port info", level: .error)
            SetDIASSourceAndPort(0, 0)
        }
    }
    
    func getCurrentActiveDisplayID() -> CGDirectDisplayID? {
        if let diasMac = currentDIASMac,
           let displayID = macToDisplayID[diasMac] {
            return displayID
        }
        return CGMainDisplayID()
    }
    
    func updateDisplayMapping(mac: String, displayID: CGDirectDisplayID) {
        macToDisplayID[mac] = displayID
        logger.log("Updated display mapping: \(mac) -> \(displayID)", level: .info)
        
        HelperOptimizedDisplayMacQuery.shared().macToDisplayID[mac] = displayID
        HelperOptimizedDisplayMacQuery.shared().displayIDToMac[displayID] = mac
    }
    
    func setCurrentDIASMac(_ mac: String) {
        currentDIASMac = mac
        logger.log("Set current DIAS MAC: \(mac)", level: .info)
    }
    
    func handleRemoteText(_ text: String) {
        logger.log("Received remote text: \(text.prefix(100))...", level: .info)
        ClipboardMonitor.shareInstance().sendTextToClipboard(text)
        DispatchQueue.main.async { [weak self] in
            self?.delegate?.didReceiveRemoteText?(text)
        }
    }
    
    func handleRemoteImage(_ base64Data: String) {
        logger.log("Received remote image data, length: \(base64Data.count)", level: .info)
        guard let imageData = Data(base64Encoded: base64Data) else {
            logger.log("Failed to decode base64 image data", level: .error)
            return
        }
        if ClipboardMonitor.shareInstance().sendImageToClipboard(imageData) {
            logger.log("Successfully wrote image to clipboard", level: .info)
        } else {
            logger.log("Failed to write image to clipboard", level: .error)
        }
        DispatchQueue.main.async { [weak self] in
            self?.delegate?.didReceiveRemoteImage?(imageData)
        }
    }
    
    func handleMultipleFileProgress(_ progress: MultipleFileTransferProgress) {
        logger.log("Processing file transfer: \(progress.currentFileName) (\(progress.receivedFileCount)/\(progress.totalFileCount)) - \(progress.receivedSize)/\(progress.totalSize) bytes", level: .info)

        let session = FileTransferSession(from: progress)
        let userInfo: [String: Any] = [
            "session": session.toDictionary(),
            "sessionId": session.sessionId,
            "senderID": progress.senderID,
            "isCompleted": progress.isCompleted,
            "progress": progress.totalProgress
        ]

        DispatchQueue.main.async { [weak self] in
            self?.delegate?.didReceiveTransferFilesDataUpdate?(userInfo)
        }

//        let notificationName: Notification.Name
//        if progress.isCompleted {
//            notificationName = .fileTransferSessionCompleted
//        } else if progress.receivedSize == 0 {
//            notificationName = .fileTransferSessionStarted
//        } else {
//            notificationName = .fileTransferSessionUpdated
//        }
//
//        NotificationCenter.default.post(
//            name: notificationName,
//            object: session,
//            userInfo: [
//                "session": session.toDictionary(),
//                "sessionId": session.sessionId,
//                "senderID": progress.senderID,
//                "isCompleted": progress.isCompleted,
//                "progress": progress.totalProgress
//            ]
//        )

        logger.log("File transfer session notification sent - \(session.sessionId) Progress: \(String(format: "%.1f", progress.totalProgress * 100))%", level: .info)
    }

    func handleMultipleProgressBar(ip: UnsafeMutablePointer<CChar>?, id: UnsafeMutablePointer<CChar>?, deviceName: UnsafeMutablePointer<CChar>?, currentfileName: UnsafeMutablePointer<CChar>?, recvFileCnt: UInt32, totalFileCnt: UInt32, currentFileSize: UInt64, totalSize: UInt64, recvSize: UInt64, timestamp: UInt64) {
        guard let ip = ip, let id = id, let deviceName = deviceName, let currentfileName = currentfileName else {
            logger.warn("Multiple progress bar callback received with null pointers")
            return
        }

        let senderIP = String(cString: ip)
        let senderID = String(cString: id)
        let deviceNameString = String(cString: deviceName)
        let currentFileNameString = String(cString: currentfileName)

        let multipleProgress = MultipleFileTransferProgress(
            senderIP: senderIP,
            senderID: senderID,
            deviceName: deviceNameString,
            currentFileName: currentFileNameString,
            receivedFileCount: recvFileCnt,
            totalFileCount: totalFileCnt,
            currentFileSize: currentFileSize,
            totalSize: totalSize,
            receivedSize: recvSize,
            timestamp: timestamp
        )

        handleMultipleFileProgress(multipleProgress)
    }

}

extension NSNotification.Name {
    static let authViaIndex = NSNotification.Name("CrossShareAuthViaIndex")
}

