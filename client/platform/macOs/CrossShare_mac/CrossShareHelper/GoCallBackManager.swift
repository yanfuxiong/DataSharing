//
//  GoCallBackManager.swift
//  CrossShareHelper
//
//  Created by ts on 2025/9/16.
//

import Foundation
import AppKit

class GoCallbackManager {
    weak var delegate: CrossShareHelperXPCDelegate?
    private let logger = CSLogger.shared
    
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
    
    func handleRemoteData(text: String?, imageBase64: String?, html: String?) {
        logger.log("Received remote clipboard data - text: \(!text.isNilOrEmpty), image: \(!imageBase64.isNilOrEmpty), html: \(!html.isNilOrEmpty)", level: .info)

        var imageData: Data? = nil
        if let base64 = imageBase64, !base64.isEmpty {
            imageData = Data(base64Encoded: base64)
        }

        var image: NSImage? = nil
        if let data = imageData {
            image = NSImage(data: data)
        }

        let success = setMultipleTypesToClipboard(text: text, image: image, html: html)

        if success {
            logger.log("Successfully wrote remote data to clipboard with multiple types", level: .info)
        } else {
            logger.log("Failed to write remote data to clipboard", level: .error)
        }

        DispatchQueue.main.async { [weak self] in
            self?.delegate?.didReceiveRemoteClipboard?(text: text, imageData: imageData, html: html)
        }
    }

    private func setMultipleTypesToClipboard(text: String?, image: NSImage?, html: String?) -> Bool {
        guard text != nil || image != nil || html != nil else {
            logger.log("No valid clipboard content to set", level: .warn)
            return false
        }

        let pasteboard = NSPasteboard.general
        pasteboard.clearContents()

        let pasteboardItem = NSPasteboardItem()
        var hasData = false

        if let text = text, !text.isEmpty {
            pasteboardItem.setString(text, forType: .string)
            hasData = true
            logger.log("Added text to pasteboard item", level: .info)
        }

        if let image = image {
            if let tiffData = image.tiffRepresentation {
                pasteboardItem.setData(tiffData, forType: .tiff)
                hasData = true
                logger.log("Added image (TIFF) to pasteboard item", level: .info)
            }
            if let pngData = image.pngData {
                pasteboardItem.setData(pngData, forType: .png)
                logger.log("Added image (PNG) to pasteboard item", level: .info)
            }
        }

        if let html = html, !html.isEmpty {
            var wrappedHTML = html
            if html.lowercased().range(of: #"<meta\s+[^>]*charset\s*=\s*["']?utf-8["']?"#, options: .regularExpression) == nil {
                wrappedHTML = """
                <!DOCTYPE html>
                <html>
                <head>
                <meta charset="UTF-8">
                <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
                </head>
                <body>
                \(html)
                </body>
                </html>
                """
            }

            if let htmlData = wrappedHTML.data(using: .utf8) {
                pasteboardItem.setData(htmlData, forType: .html)
                hasData = true
                logger.log("Added HTML to pasteboard item, length: \(wrappedHTML.count)", level: .info)
            }
        }

        if hasData {
            let result = pasteboard.writeObjects([pasteboardItem])
            logger.log("Pasteboard write result: \(result)", level: .info)
            return result
        }

        return false
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

}

extension NSNotification.Name {
    static let authViaIndex = NSNotification.Name("CrossShareAuthViaIndex")
}

private extension Optional where Wrapped == String {
    var isNilOrEmpty: Bool {
        return self == nil || self?.isEmpty == true
    }
}

