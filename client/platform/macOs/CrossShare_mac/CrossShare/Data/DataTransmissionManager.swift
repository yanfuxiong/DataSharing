//
//  DataTransmissionManager.swift
//  CrossShare
//
//  Created by TS on 2025/10/11.
//

import Foundation

/// Data Transmission Manager - Centralized management of file transfer related business logic
class DataTransmissionManager {
    static let shared = DataTransmissionManager()
    
    // Callback definition - Notify UI when data is updated
    typealias DataUpdateCallback = ([CSFileInfo], Bool, Int?) -> Void
    
    // Save callback and ViewController reference
    private var dataUpdateCallback: DataUpdateCallback?
    private weak var viewController: MainHomeViewController?
    
    private init() {}
    
    // MARK: - Public Methods
    
    /// Start listening for transfer events
    /// - Parameters:
    ///   - viewController: MainHomeViewController reference (for accessing deviceList and bottomTableData)
    ///   - callback: Callback when data is updated
    func startListening(
        viewController: MainHomeViewController,
        callback: @escaping DataUpdateCallback
    ) {
        self.viewController = viewController
        self.dataUpdateCallback = callback
        
        // Register all transfer-related notifications
        setupNotifications()
    }
    
    /// Stop listening
    func stopListening() {
        NotificationCenter.default.removeObserver(self)
        viewController = nil
        dataUpdateCallback = nil
    }
    
    // MARK: - Setup Notification Listening
    
    private func setupNotifications() {
        // 1. Listen for file transfer start
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(onTransferStart(_:)),
            name: .fileTransferSessionStarted,
            object: nil
        )
        
        // 2. Listen for file transfer update
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(onTransferUpdate(_:)),
            name: .fileTransferSessionUpdated,
            object: nil
        )
        
        // 3. Listen for file transfer complete
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(onTransferComplete(_:)),
            name: .fileTransferSessionCompleted,
            object: nil
        )
        
        // 4. Listen for error events
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(onError(_:)),
            name: .didReceiveErrorEventNotification,
            object: nil
        )
    }
    
    // MARK: - Notification Handling Methods
    
    /// Handle transfer start notification
    @objc private func onTransferStart(_ notification: Notification) {
        // 1. Get notification data
        guard let userInfo = notification.userInfo as? [String: Any] else {
            logger.error("❌ Failed to parse transfer start data")
            return
        }
        logger.info("✅ Received transfer start notification")
        
        // 2. Get device list and current data
        guard let deviceList = viewController?.deviceList,
              let currentData = viewController?.bottomTableData else {
            logger.error("❌ Failed to get device list or current data")
            return
        }
        
        // 3. Create CSFileInfo
        guard let fileInfo = createFileInfo(from: userInfo, deviceList: deviceList) else {
            logger.error("❌ Failed to create CSFileInfo")
            return
        }
        
        // 4. Save to database and notify UI update
        saveToDatabase(fileInfo: fileInfo, currentData: currentData)
    }
    
    /// Handle transfer update notification
    @objc private func onTransferUpdate(_ notification: Notification) {
        // 1. Get notification data
        guard let userInfo = notification.userInfo as? [String: Any] else {
            logger.error("❌ Failed to parse transfer update data")
            return
        }
        
        // 2. Get device list and current data
        guard let deviceList = viewController?.deviceList,
              let currentData = viewController?.bottomTableData else {
            logger.error("❌ Failed to get device list or current data")
            return
        }
        
        // 3. Create CSFileInfo
        guard let fileInfo = createFileInfo(from: userInfo, deviceList: deviceList) else {
            logger.error("❌ Failed to create CSFileInfo")
            return
        }
        
        logger.info("📝 Transfer update - sessionId: \(fileInfo.sessionId), progress: \(fileInfo.progress)")
        
        // 4. Save to database and notify UI update
        saveToDatabase(fileInfo: fileInfo, currentData: currentData)
    }
    
    /// Handle transfer complete notification
    @objc private func onTransferComplete(_ notification: Notification) {
        // 1. Get notification data
        guard let userInfo = notification.userInfo as? [String: Any] else {
            logger.error("❌ Failed to parse transfer complete data")
            return
        }
        
        // 2. Get device list and current data
        guard let deviceList = viewController?.deviceList,
              let currentData = viewController?.bottomTableData else {
            logger.error("❌ Failed to get device list or current data")
            return
        }
        
        // 3. Create CSFileInfo
        guard let fileInfo = createFileInfo(from: userInfo, deviceList: deviceList) else {
            logger.error("❌ Failed to create CSFileInfo")
            return
        }
        
        logger.info("✅ Transfer completed - sessionId: \(fileInfo.sessionId)")
        
        // 4. Save to database and notify UI update
        saveToDatabase(fileInfo: fileInfo, currentData: currentData)
    }
    
    /// Handle error notification
    @objc private func onError(_ notification: Notification) {
        // 1. Get error information
        guard let errorInfo = notification.userInfo as? [String: Any] else {
            logger.error("❌ Failed to parse error data")
            return
        }
        logger.warn("⚠️ Received error notification: \(errorInfo)")
        
        // 2. Extract id and timestamp
        guard let id = errorInfo["id"] as? String,
              let timestamp = errorInfo["timestamp"] as? String else {
            logger.error("❌ Failed to extract id or timestamp")
            return
        }
        
        // 3. Extract error code
        let errCode = errorInfo["errCode"] as? Int
        
        // 4. Assemble sessionId (format: id-timestamp)
        let sessionId = "\(id)-\(timestamp)"
        logger.info("🔍 Searching for sessionId: \(sessionId)")
        
        // 5. Get current data
        guard let currentData = viewController?.bottomTableData else {
            logger.error("❌ Failed to get current data")
            return
        }
        
        // 6. Find matching transfer record
        guard var matchedFileInfo = currentData.first(where: { $0.sessionId == sessionId }) else {
            logger.error("❌ No matching transfer record found")
            return
        }
        
        logger.info("✅ Found matching record, device: \(matchedFileInfo.session.deviceName)")
        
        // 7. Update error code
        matchedFileInfo.errCode = errCode
        logger.info("📝 Updated error code: \(errCode ?? -1)")
        
        // 8. Save to database and notify UI update
        saveToDatabase(fileInfo: matchedFileInfo, currentData: currentData)
    }
    
    // MARK: - Helper Methods
    
    /// Create CSFileInfo from dictionary
    private func createFileInfo(from userInfo: [String: Any], deviceList: [CrossShareDevice]) -> CSFileInfo? {
        // 1. Extract required fields
        guard let sessionDict = userInfo["session"] as? [String: Any],
              let sessionId = userInfo["sessionId"] as? String,
              let senderID = userInfo["senderID"] as? String,
              let progress = userInfo["progress"] as? Double,
              let isCompletedInt = userInfo["isCompleted"] as? Int else {
            logger.error("❌ Required fields are missing")
            return nil
        }
        
        // 2. Extract optional fields
        let errCode = userInfo["errCode"] as? Int
        
        // 3. Create FileTransfer
        guard let fileTransfer = createFileTransfer(from: sessionDict, deviceList: deviceList) else {
            logger.error("❌ Failed to create FileTransfer")
            return nil
        }
        
        // 4. Create and return CSFileInfo
        return CSFileInfo(
            session: fileTransfer,
            sessionId: sessionId,
            senderID: senderID,
            isCompleted: isCompletedInt != 0,
            progress: progress,
            errCode: errCode
        )
    }
    
    /// Create FileTransfer object
    private func createFileTransfer(from dict: [String: Any], deviceList: [CrossShareDevice]) -> FileTransfer? {
        // 1. First create basic FileTransfer with dictionary
        guard let fileTransfer = FileTransfer(from: dict) else {
            return nil
        }
        
        // 2. Find matching device and update device name
        return updateDeviceName(fileTransfer: fileTransfer, dict: dict, deviceList: deviceList)
    }
    
    /// Find device and update device name
    private func updateDeviceName(
        fileTransfer: FileTransfer,
        dict: [String: Any],
        deviceList: [CrossShareDevice]
    ) -> FileTransfer? {
        // 1. Extract senderID and senderIP
        guard let senderID = dict["senderID"] as? String,
              let senderIP = dict["senderIP"] as? String else {
            return fileTransfer
        }
        
        // 2. Find matching device in device list
        for device in deviceList {
            if device.id == senderID && device.ipAddr == senderIP {
                logger.info("✅ Found matching device: \(device.deviceName)")
                
                // 3. Create new dictionary and add device name
                var newDict = dict
                newDict["deviceName"] = device.deviceName
                
                // 4. Recreate FileTransfer with new dictionary
                if let newFileTransfer = FileTransfer(from: newDict) {
                    return newFileTransfer
                }
                
                // If creation fails, return the original
                return fileTransfer
            }
        }
        
        logger.warn("⚠️ No matching device found, senderID: \(senderID), senderIP: \(senderIP)")
        return fileTransfer
    }
    
    /// Save to database and notify UI update
    private func saveToDatabase(fileInfo: CSFileInfo, currentData: [CSFileInfo]) {
        // Call RealmDataManager to save data
        RealmDataManager.shared.updateCSFileInfo(fileInfo, bottomTableData: currentData) { [weak self] result in
            switch result {
            case .success(let data):
                // Data saved successfully, notify UI update
                self?.dataUpdateCallback?(
                    data.updatedData,  // Updated data
                    data.isNewRecord,  // Whether it's a new record
                    data.index         // Index of update (if updating existing record)
                )
                
            case .failure(let error):
                logger.error("❌ Failed to save data: \(error.localizedDescription)")
            }
        }
    }
}
