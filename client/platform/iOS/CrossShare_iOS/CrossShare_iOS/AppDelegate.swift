//
//  AppDelegate.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/4/15.
//

import UIKit

@main
class AppDelegate: UIResponder, UIApplicationDelegate, UNUserNotificationCenterDelegate {
    
    private var activeTransferTasks: [String: TransferTask] = [:]
    
    func application(_ application: UIApplication, didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?) -> Bool {
        UNUserNotificationCenter.current().delegate = self
        PushNotiManager.shared.initNoti()
        CSNetworkAccessibility.sharedInstance().start()
        CSNetworkAccessibility.sharedInstance().setAlertEnable(true)
        NotificationCenter.default.addObserver(self, selector: #selector(netWorkChanged(_:)), name: CSNetworkAccessibilityChangedNotification, object: nil)
        NotificationCenter.default.addObserver(self,selector: #selector(handleWillTerminate),name: UIApplication.willTerminateNotification,object: nil)
        ClipboardMonitor.shareInstance().startMonitoring()
        ScreenManager.shared.startMonitoring()
        _ = FileTransferDataManager.shared
        setupShareExtensionCommunication()
        return true
    }
    
    // MARK: UISceneSession Lifecycle
    func application(_ application: UIApplication, configurationForConnecting connectingSceneSession: UISceneSession, options: UIScene.ConnectionOptions) -> UISceneConfiguration {
        // Called when a new scene session is being created.
        // Use this method to select a configuration to create the new scene with.
        return UISceneConfiguration(name: "Default Configuration", sessionRole: connectingSceneSession.role)
    }
    
    func application(_ application: UIApplication, didDiscardSceneSessions sceneSessions: Set<UISceneSession>) {
        // Called when the user discards a scene session.
        // If any sessions were discarded while the application was not running, this will be called shortly after application:didFinishLaunchingWithOptions.
        // Use this method to release any resources that were specific to the discarded scenes, as they will not return.
    }
}

extension AppDelegate {
    @objc private func netWorkChanged(_ ntf: Notification) {
        Logger.info("AppDelegate - Network notification received: \(ntf)")
        
        let status = CSNetworkAccessibility.sharedInstance().currentState()
        Logger.info("AppDelegate - Current network status: \(status.rawValue)")
        
        switch status {
        case .checking:
            Logger.info("Network status: checking")
        case .unknown:
            Logger.info("Network status: unknown")
        case .accessible:
            Logger.info("Network status: accessible")
        case .accessibleWiFi, .accessibleCellular:
            Logger.info("Network status: accessible WiFi or Cellular")
            CSNetworkAccessibility.sharedInstance().initializeApp { success in
                Logger.info("App initialization result: \(success)")
            }
        case .restricted:
            Logger.info("Network status: restricted")
        }
    }
    
    @objc private func handleWillTerminate() {
        Logger.info("AppDelegate - Application will terminate")
        UserDefaults.setBool(forKey: .ISACTIVITY, value: false,type: .group)
        let success = UserDefaults.synchronize(type: .group)
        Logger.info("UserDefaults synchronize result: \(success)")
    }
}

extension AppDelegate {
    func userNotificationCenter(_ center: UNUserNotificationCenter,
                                willPresent notification: UNNotification,
                                withCompletionHandler completionHandler: @escaping (UNNotificationPresentationOptions) -> Void) {
        completionHandler([.banner, .list, .sound])
    }
}

extension AppDelegate {
    private func setupShareExtensionCommunication() {
        let communicationManager = SharedCommunicationManager.shared
        
        communicationManager.onTransferPrepare = { [weak self] payload in
            guard let self = self else {
                Logger.error("AppDelegate: Invalid transfer prepare payload")
                return
            }
            self.handleTransferPrepare(payload: payload)
        }
        
        communicationManager.onTransferStart = { [weak self] payload in
            guard let self = self else {
                Logger.error("AppDelegate: Invalid transfer prepare payload")
                return
            }
            self.handleTransferStart(payload: payload)
        }
    }
    
    private func handleTransferPrepare(payload: [String: Any]) {
        guard let clientId = payload["clientId"] as? String,
              let clientIp = payload["clientIp"] as? String,
              let clientName = payload["clientName"] as? String,
              let filePaths = payload["filePaths"] as? [String]
        else {
            Logger.error("AppDelegate: Invalid transfer prepare payload")
            return
        }
        
        let taskId = payload["taskId"] as? String ?? UUID().uuidString
        
        Logger.info("AppDelegate: Files prepared by Share Extension, starting transfer... TaskID: \(taskId)")
        
        let transferTask = TransferTask(
            taskId: taskId,
            clientId: clientId,
            clientIp: clientIp,
            clientName: clientName,
            filePaths: filePaths,
            startTime: Date(),
            status: .preparing
        )
        
        self.activeTransferTasks[taskId] = transferTask
        
        self.setupP2PCallbacks(for: taskId)
        
        let startPayload: [String: Any] = [
            "taskId": taskId,
            "clientId": clientId,
            "clientIp": clientIp,
            "clientName": clientName,
            "filePaths": filePaths,
            "status": "transfer_started",
            "timestamp": Date().timeIntervalSince1970
        ]
        
        SharedCommunicationManager.shared.sendCommunication(
            type: .transferStart,
            payload: startPayload
        )
    }
    
    private func handleTransferStart(payload: [String: Any]) {
        guard let clientId = payload["clientId"] as? String,
              let clientIp = payload["clientIp"] as? String,
              let clientName = payload["clientName"] as? String,
              let taskId = payload["taskId"] as? String,
              let filePaths = payload["filePaths"] as? [String]
        else {
            Logger.error("AppDelegate: Invalid transfer complete payload")
            return
        }
        
        Logger.info("AppDelegate: Transfer started from Share Extension - Client ID: \(clientId), IP: \(clientIp), Name: \(clientName)")
        
        if let pathsJsonString = ["Id": clientId,"Ip": clientIp, "PathList": filePaths].toJsonString() {
            DispatchQueue.global(qos: .userInitiated).async {
                P2PManager.shared.setFileListsDropRequest(filePath: pathsJsonString,taskId: taskId)
                Logger.info("AppDelegate: Successfully initiated multiple file drop for URL: \(pathsJsonString)")
            }
        }
        
        let progressPayload: [String: Any] = [
            "status": "transfer_progress",
            "clientId": clientId,
            "clientIp": clientIp,
            "taskId": taskId,
            "timestamp": Date().timeIntervalSince1970
        ]
        
        SharedCommunicationManager.shared.sendCommunication(
            type: .transferProgress,
            payload: progressPayload
        )
    }
    
    private func setupP2PCallbacks(for taskId: String) {
        P2PManager.shared.setTransferCallbacks(
            taskId: taskId,
            onProgress: { [weak self] progress, currentFile in
                self?.handleTransferProgress(taskId: taskId, progress: progress, currentFile: currentFile)
            },
            onSuccess: { [weak self] in
                self?.handleTransferSuccess(taskId: taskId)
            },
            onError: { [weak self] error in
                self?.handleTransferError(taskId: taskId, error: error)
            }
        )
    }
    
    private func handleTransferProgress(taskId: String, progress: Float, currentFile: String) {
        guard var task = self.activeTransferTasks[taskId] else { return }
        
        task.status = .inProgress
        self.activeTransferTasks[taskId] = task
        
        let payload: [String: Any] = [
            "taskId": taskId,
            "status": "in_progress",
            "progress": progress,
            "currentFile": currentFile,
            "timestamp": Date().timeIntervalSince1970
        ]
        
        SharedCommunicationManager.shared.sendCommunication(
            type: .transferProgress,
            payload: payload
        )
        
        Logger.info("Transfer progress for TaskID \(taskId): \(progress * 100)% - \(currentFile)")
    }
    
    private func handleTransferSuccess(taskId: String) {
        guard var task = self.activeTransferTasks[taskId] else { return }
        
        task.status = .completed
        self.activeTransferTasks[taskId] = task
        
        let payload: [String: Any] = [
            "taskId": taskId,
            "status": "completed",
            "clientName": task.clientName,
            "fileCount": task.filePaths.count,
            "timestamp": Date().timeIntervalSince1970
        ]
        
        SharedCommunicationManager.shared.sendCommunication(
            type: .transferComplete,
            payload: payload
        )
        
        DispatchQueue.main.async {
            let fileMsg = task.filePaths.count > 1 ? "\(task.filePaths.count) files" : "\(task.filePaths.count) file"
            let params: [String] = [fileMsg, task.clientName]
            PushNotiManager.shared.sendLocalNotification(code: .sendStart, with: params)
        }
        
        Logger.info("Transfer completed for TaskID: \(taskId)")
        
        DispatchQueue.main.asyncAfter(deadline: .now() + 5) {
            self.activeTransferTasks.removeValue(forKey: taskId)
        }
    }
    
    private func handleTransferError(taskId: String, error: String) {
        guard var task = self.activeTransferTasks[taskId] else { return }
        
        task.status = .failed
        self.activeTransferTasks[taskId] = task
        
        let payload: [String: Any] = [
            "taskId": taskId,
            "status": "failed",
            "error": error,
            "clientName": task.clientName,
            "timestamp": Date().timeIntervalSince1970
        ]
        
        SharedCommunicationManager.shared.sendCommunication(
            type: .transferError,
            payload: payload
        )
        
        Logger.error("Transfer failed for TaskID \(taskId): \(error)")
        
        DispatchQueue.main.asyncAfter(deadline: .now() + 5) {
            self.activeTransferTasks.removeValue(forKey: taskId)
        }
    }
}
