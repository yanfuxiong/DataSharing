//
//  GoCallBackManager.swift
//  CrossShareHelper
//
//  Created by ts on 2025/9/16.
//

import Foundation
import AppKit
import UserNotifications

class GoCallbackManager {
    weak var delegate: CrossShareHelperXPCDelegate?
    private let logger = CSLogger.shared
    
    private var macToDisplayID: [String: CGDirectDisplayID] = [:]
    private var currentDIASMac: String?
    var diasStatus: Int = 0
    
    init() {
        // Set up notification manager delegate
        HelperNotificationManager.shared.delegate = self
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
        diasStatus = status
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
    }
    
    func handleSystemInfoUpdate(ipInfo: String, verInfo: String) {
        logger.log("Received system info update - IP: \(ipInfo), Version: \(verInfo)", level: .info)
        
        let systemInfoData: [String: Any] = [
            "ipInfo": ipInfo,
            "verInfo": verInfo,
        ]
        
        logger.log("Parsed system info data: \(systemInfoData)", level: .info)
        DispatchQueue.main.async { [weak self] in
            self?.delegate?.didReceiveSystemInfoUpdate?(systemInfoData)
            self?.logger.log("Sent system info update to GUI through XPC", level: .info)
        }
    }
    
    func handleThemeInfoUpdate(_ themeInfo: [String: Any]) {
        logger.log("Received theme info:\(themeInfo)", level: .info)
        DispatchQueue.main.async { [weak self] in
            self?.delegate?.didReceiveThemeInfoUpdate?(themeInfo)
            self?.logger.log("Sent theme info update to GUI through XPC", level: .info)
        }
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
    
    func handleRemoteData(text: String?, imageBase64: String?, html: String?, rtf: String?) {
        logger.log("Received remote clipboard data - text: \(!text.isNilOrEmpty), image: \(!imageBase64.isNilOrEmpty), html: \(!html.isNilOrEmpty), rtf: \(!rtf.isNilOrEmpty)", level: .info)

        var imageData: Data? = nil
        if let base64 = imageBase64, !base64.isEmpty {
            imageData = Data(base64Encoded: base64)
        }

        var image: NSImage? = nil
        if let data = imageData {
            image = NSImage(data: data)
        }

        // setupClipboard updates lastChangeCount internally, so ClipboardMonitor
        // won't detect this write. We must send the notification here explicitly.
        HelperNotificationManager.shared.sendClipboardNotification(text: text, image: image, html: html, rtf: rtf)

        let success = ClipboardMonitor.shareInstance().setupClipboard(text: text, image: image, html: html, rtf: rtf)

        if success {
            logger.log("Successfully wrote remote data to clipboard with multiple types", level: .info)
        } else {
            logger.log("Failed to write remote data to clipboard", level: .error)
        }
    }
    
    // MARK: - Clipboard Notification (Delegate to HelperNotificationManager)
    
    /// Send clipboard copy notification
    func sendClipboardNotification(text: String?, image: NSImage?, html: String?, rtf: String? = nil) {
        HelperNotificationManager.shared.sendClipboardNotification(text: text, image: image, html: html, rtf: rtf)
    }
    
    // MARK: - File Transfer
    
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
    
    // MARK: - Error Handling
    
    func handleErrEvent(id: String, errCode: UInt32, arg1: String, arg2: String, arg3: String, arg4: String) {
        logger.log("Error event - ID: \(id), ErrCode: \(errCode), Args: [\(arg1), \(arg2), \(arg3), \(arg4)]", level: .error)
        
        let errorInfo: [String: Any] = [
            "id": id,
            "errCode": errCode,
            "ipaddr": arg1,
            "timestamp": arg2,
            "arg3": arg3,
            "arg4": arg4,
        ]
        
        DispatchQueue.main.async { [weak self] in
            self?.delegate?.didReceiveErrorEvent?(errorInfo)
        }
    }
    
    // MARK: - System Notification
    
    func handleNotiMessage(timestamp: UInt64, notiCode: UInt32, notiParam: [String]) {
        logger.log("NotiMessage event - timestamp: \(timestamp), notiCode: \(notiCode), params: \(notiParam)", level: .info)
        HelperNotificationManager.shared.sendSystemNotification(notiCode: notiCode, notiParam: notiParam)
    }
    
    /// Check notification permission status for Helper process
    /// - Returns: true if notification permission is authorized, false otherwise
    func checkHelperNotiStatus() -> Bool {
        return HelperNotificationManager.shared.checkNotificationStatus()
    }
}

// MARK: - HelperNotificationManagerDelegate

extension GoCallbackManager: HelperNotificationManagerDelegate {
    func notificationManagerRequestOpenNotiAlert() {
        DispatchQueue.main.async { [weak self] in
            self?.delegate?.requestOpenNotiAlert?()
        }
    }
}

// MARK: - Extensions

extension NSNotification.Name {
    static let authViaIndex = NSNotification.Name("CrossShareAuthViaIndex")
}

private extension Optional where Wrapped == String {
    var isNilOrEmpty: Bool {
        return self == nil || self?.isEmpty == true
    }
}
