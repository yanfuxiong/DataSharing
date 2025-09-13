//
//  ShareViewController.swift
//  share
//
//  Created by ts on 2025/5/21.
//

import UIKit
import Social
import UniformTypeIdentifiers
import SwiftyJSON
import MBProgressHUD

class ShareViewController: UIViewController {
    
    private var devicePopView: DeviceSelectPopView?
    private let notifier = CrossProcessNotifier(appGroupID: UserDefaults.groupId)
    private let dataUpdateNotification = "com.realtek.crossshare.dataDidUpdate"
    private var status:Int = 1
    private var statusString:String = ""
    
    override func viewDidLoad() {
        super.viewDidLoad()
        handleInput()
        addNotificationObservers()
    }

    override func viewWillDisappear(_ animated: Bool) {
        super.viewWillDisappear(animated)
        notifier.stopObserving(name: dataUpdateNotification)
    }
    
    func handleInput() {
        guard let extensionItem = extensionContext?.inputItems.first as? NSExtensionItem,
              let attachments = extensionItem.attachments else {
            cancelExtensionContext(with: "No input items found")
            return
        }
        
        var filePaths: [URL] = []
        let dispatchGroup = DispatchGroup()
        
        for provider in attachments {
            dispatchGroup.enter()
            if provider.hasItemConformingToTypeIdentifier(UTType.fileURL.identifier) {
                provider.loadItem(forTypeIdentifier: UTType.fileURL.identifier, options: nil) { (item, error) in
                    defer { dispatchGroup.leave() }
                    if let url = item as? URL {
                        Logger.info("Got file URL directly: \(url)")
                        filePaths.append(url)
                    } else if let urlData = item as? Data, let url = URL(dataRepresentation: urlData, relativeTo: nil) {
                        Logger.info("Got file URL from data: \(url)")
                        filePaths.append(url)
                    } else if let urlString = item as? String, let url = URL(string: urlString) {
                        Logger.info("Got file URL from string: \(url)")
                        filePaths.append(url)
                    }
                }
            } else if provider.hasItemConformingToTypeIdentifier(UTType.data.identifier) {
                provider.loadItem(forTypeIdentifier: UTType.data.identifier, options: nil) { (item, error) in
                    defer { dispatchGroup.leave() }
                    if let url = item as? URL {
                        Logger.info("Got URL from data type: \(url)")
                        filePaths.append(url)
                    } else if let data = item as? Data {
                        let tempDirectory = FileManager.default.temporaryDirectory
                        let fileName = UUID().uuidString
                        let tempFileURL = tempDirectory.appendingPathComponent(fileName)
                        do {
                            try data.write(to: tempFileURL)
                            Logger.info("Saved data to temporary file: \(tempFileURL)")
                            filePaths.append(tempFileURL)
                        } catch {
                            Logger.info("Error saving data to temporary file: \(error)")
                        }
                    }
                }
            } else {
                let availableTypeIdentifiers = provider.registeredTypeIdentifiers
                if let firstType = availableTypeIdentifiers.first {
                    provider.loadItem(forTypeIdentifier: firstType, options: nil) { (item, error) in
                        defer { dispatchGroup.leave() }
                        Logger.info("Processing item of type: \(firstType)")
                        if let url = item as? URL {
                            Logger.info("Got URL from generic type: \(url)")
                            filePaths.append(url)
                        } else if let data = item as? Data {
                            let tempDirectory = FileManager.default.temporaryDirectory
                            let fileName = UUID().uuidString
                            let fileExtension = self.getFileExtension(from: firstType)
                            let tempFileURL = tempDirectory.appendingPathComponent(fileName + fileExtension)
                            do {
                                try data.write(to: tempFileURL)
                                Logger.info("Saved generic data to temporary file: \(tempFileURL)")
                                filePaths.append(tempFileURL)
                            } catch {
                                Logger.info("Error saving generic data: \(error)")
                            }
                        } else {
                            Logger.info("Unknown item type: \(String(describing: item))")
                        }
                    }
                } else {
                    dispatchGroup.leave()
                    Logger.info("No available type identifiers for provider")
                }
            }
        }
        
        dispatchGroup.notify(queue: .main) {
            if filePaths.isEmpty {
                Logger.info("No files found to share.")
                self.cancelExtensionContext(with: "No files selected.")
                return
            }
            guard UserDefaults.getBool(forKey: .ISACTIVITY, type: .group) == true else {
                Logger.info("Error: ISACTIVITY not found in UserDefaults.")
                self.cancelExtensionContext(with: "Please open CrossShare app first and wait a moment")
                return
            }
            self.showDeviceSelectPopView(urls: filePaths)
        }
    }
    
    private func addNotificationObservers() {
        notifier.observe(name: dataUpdateNotification) { userInfo in
            if let data = userInfo,let status = data["status"] as? Int {
                print("received data: \(data)")
                self.status = status
                self.statusString = data["message"] as? String ?? ""
                DispatchQueue.main.async {
                    if self.status == 1 {
                        MBProgressHUD.showTips(.success, self.statusString, toView: self.view, duration: 1.5)
                    } else {
                        MBProgressHUD.showTips(.warn, self.statusString, toView: self.view, duration: 1.5)
                    }
                }
            }
        }
    }
    
    private func getFileExtension(from typeIdentifier: String) -> String {
        switch typeIdentifier {
        case "com.android.package-archive":
            return ".apk"
        case "public.jpeg":
            return ".jpg"
        case "public.png":
            return ".png"
        case "public.mp4":
            return ".mp4"
        case "com.adobe.pdf":
            return ".pdf"
        case "public.zip-archive":
            return ".zip"
        default:
            if #available(iOS 14.0, *) {
                if let uttype = UTType(typeIdentifier) {
                    return uttype.preferredFilenameExtension.map { "." + $0 } ?? ""
                }
            }
            return ""
        }
    }
    
    func showDeviceSelectPopView(urls: [URL]) {
        let (hasFiles, hasFolders) = analyzeURLs(urls)
        if hasFolders && hasFiles {
            showFolderWarningPopup(urls: urls, type: .mixedContent)
            return
        } else if hasFolders && !hasFiles {
            showFolderWarningPopup(urls: urls, type: .foldersOnly)
            return
        }
        continueWithFileSelection(urls: urls,isFolder: false)
    }
    
    private func analyzeURLs(_ urls: [URL]) -> (hasFiles: Bool, hasFolders: Bool) {
        var hasFiles = false
        var hasFolders = false
        
        for url in urls {
            guard url.startAccessingSecurityScopedResource() else {
                continue
            }
            defer { url.stopAccessingSecurityScopedResource() }
            
            var isDirectory: ObjCBool = false
            if FileManager.default.fileExists(atPath: url.path, isDirectory: &isDirectory) {
                if isDirectory.boolValue {
                    hasFolders = true
                } else {
                    hasFiles = true
                }
                if hasFiles && hasFolders {
                    break
                }
            }
        }
        return (hasFiles, hasFolders)
    }
    
    private func showFolderWarningPopup(urls: [URL], type: ShareFilesPopType) {
        let popView = ShareFilesPopView(frame: self.view.bounds, type: type)
        
        switch type {
        case .mixedContent:
            popView.onContinue = { [weak self] in
                self?.dismissFolderPopView(popView)
                let filteredUrls = self?.filterOutFolders(from: urls) ?? []
                if !filteredUrls.isEmpty {
                    self?.continueWithFileSelection(urls: filteredUrls,isFolder: true)
                } else {
                    self?.cancelExtensionContext(with: "No files to transfer after filtering")
                }
            }
            
            popView.onCancel = { [weak self] in
                self?.dismissFolderPopView(popView)
                self?.cancelExtensionContext(with: "User cancelled file transfer")
            }
            
        case .foldersOnly:
            popView.onConfirm = { [weak self] in
                self?.dismissFolderPopView(popView)
                self?.cancelExtensionContext(with: "Only folders selected, transfer not supported")
            }
        }
        
        popView.alpha = 0
        self.view.addSubview(popView)
        
        UIView.animate(withDuration: 0.3) {
            popView.alpha = 1
        }
    }
    
    private func continueWithFileSelection(urls: [URL], isFolder: Bool = false) {
        var clients: [ClientInfo] = []
        var fileNames: [String] = []

        if let clientsString = UserDefaults.get(forKey: .DEVICE_CLIENTS,type: .group),
           let data = clientsString.data(using: .utf8) {
            let json = try? JSON(data: data)
            if let array = json?.array {
                clients = array.map { item in
                    ClientInfo(
                        ip: item["ip"].stringValue,
                        id: item["id"].stringValue,
                        name: item["name"].stringValue,
                        deviceType: item["deviceType"].stringValue
                    )
                }
            }
        }

        if urls.isEmpty {
            print("Error: No URLs provided to continueWithFileSelection.")
            self.cancelExtensionContext(with: "No URLs to process")
            return
        }

        for url in urls {
            if url.startAccessingSecurityScopedResource() {
                fileNames.append(url.lastPathComponent)
                url.stopAccessingSecurityScopedResource()
            } else {
                fileNames.append("Unknown File")
            }
        }
        
        print("Found \(clients.count) clients:")
        for client in clients {
            print("Device: \(client.name), IP:Port: \(client.ip), ID: \(client.id)")
        }
        
        var selectedClient: ClientInfo?
        
        if !clients.isEmpty {
            let popView = DeviceSelectPopView(fileNames: fileNames, clients: clients)
            popView.frame = self.view.bounds
            popView.alpha = 0
            self.devicePopView = popView

            popView.onSelect = { [weak self] client in
                guard let self = self else { return }
                print("choose device：\(client.name)")
                selectedClient = client
            }
            popView.onCancel = { [weak self] in
                self?.dismissDevicePopView()
                self?.cancelExtensionContext(with: "User cancelled device selection")
            }
            popView.onSure = { [weak self] in
                guard let self = self else { return }
                guard let currentSelectedClient = selectedClient else {
                    MBProgressHUD.showTips(.warn,"Please select a device", toView: self.view)
                    return
                }
                
                self.dismissDevicePopView()
                let timestamp = String(Int(Date().timeIntervalSince1970))
                let taskId = UUID().uuidString
                if urls.count > 0 {
                    self.handleMultipleFileShare(
                        urls: urls,
                        client: currentSelectedClient,
                        timestamp: timestamp,
                        taskId: taskId
                    )
                } else {
                    self.cancelExtensionContext(with: "No files to share.")
                    return
                }
            }
            
            if !isFolder {
                popView.y -= 80
            }
            self.view.addSubview(popView)
            UIView.animate(withDuration: 0.3) {
                popView.alpha = 1
                if let contentView = popView.subviews.first {
                    contentView.transform = .identity
                }
            }
        } else {
            print("No clients available to share with.")
            MBProgressHUD.showTips(.error, "No devices found.", toView: self.view)
            self.cancelExtensionContext(with: "No clients found")
        }
    }
    
    private func handleMultipleFileShare(urls: [URL], client: ClientInfo, timestamp: String,taskId : String) {
        var copiedFilePaths: [String] = []
        for url in urls {
            if let copiedPath = self.copyDocumentToTemp(url, timestamp) {
                copiedFilePaths.append(copiedPath)
            }
        }
        
        if copiedFilePaths.isEmpty {
            self.sendTransferError(message: "Failed to prepare files")
            return
        }
        
        let payload: [String: Any] = [
            "taskId":taskId,
            "clientId": client.id,
            "clientIp": client.ip,
            "clientName": client.name,
            "filePaths": copiedFilePaths,
            "timestamp": timestamp,
            "fileNames": urls.map { $0.lastPathComponent }
        ]
        
        SharedCommunicationManager.shared.sendCommunication(
            type: .transferPrepare,
            payload: payload
        )
        
        DispatchQueue.main.asyncAfter(deadline: .now() + 1.5) {
            self.extensionContext?.completeRequest(returningItems: [], completionHandler: nil)
        }
    }
    
    private func sendTransferError(message: String) {
        let payload: [String: Any] = [
            "error": message,
            "timestamp": Date().timeIntervalSince1970
        ]
        
        SharedCommunicationManager.shared.sendCommunication(
            type: .transferError,
            payload: payload
        )
        
        self.cancelExtensionContext(with: message)
    }

    // MARK: - Folder Detection Methods
    private func filterOutFolders(from urls: [URL]) -> [URL] {
        return urls.filter { url in
            guard url.startAccessingSecurityScopedResource() else {
                return false
            }
            defer { url.stopAccessingSecurityScopedResource() }
            
            var isDirectory: ObjCBool = false
            if FileManager.default.fileExists(atPath: url.path, isDirectory: &isDirectory) {
                return !isDirectory.boolValue
            }
            return false
        }
    }

    private func dismissFolderPopView(_ popView: ShareFilesPopView) {
        UIView.animate(withDuration: 0.3, animations: {
            popView.alpha = 0
        }) { _ in
            popView.removeFromSuperview()
        }
    }
    
    private func dismissDevicePopView() {
        guard let popView = self.devicePopView else { return }
        UIView.animate(withDuration: 0.3, animations: {
            popView.alpha = 0
            if let contentView = popView.subviews.first {
                contentView.transform = CGAffineTransform(scaleX: 0.8, y: 0.8)
            }
        }) { _ in
            popView.removeFromSuperview()
            self.devicePopView = nil
        }
    }
    
    private func copyDocumentToTemp(_ url: URL, _ timestamp: String) -> String? {
        guard let groupContainerURL = FileManager.default.containerURL(forSecurityApplicationGroupIdentifier: UserDefaults.groupId) else {
            Logger.info("[DocPicker][Err] Could not get App Group container URL. GroupID: \(UserDefaults.groupId)")
            return nil
        }
        
        var securityScopedAccess = false
        if url.scheme == "file" {
            securityScopedAccess = url.startAccessingSecurityScopedResource()
            if !securityScopedAccess {
                Logger.info("[DocPicker][Warn] Could not start security-scoped access for: \(url.path). Will attempt copy anyway.")
            }
        }
        defer {
            if securityScopedAccess {
                url.stopAccessingSecurityScopedResource()
            }
        }
        
        let fileName = url.lastPathComponent
        let sharedFilesDirectory = groupContainerURL.appendingPathComponent("SharedFiles")
        let destFolder = sharedFilesDirectory.appendingPathComponent(timestamp)
        let destFullPath = destFolder.appendingPathComponent(fileName)
        
        do {
            try FileManager.default.createDirectory(at: destFolder, withIntermediateDirectories: true, attributes: nil)
            if FileManager.default.fileExists(atPath: destFullPath.path) {
                try FileManager.default.removeItem(at: destFullPath)
            }
            try FileManager.default.copyItem(at: url, to: destFullPath)
            Logger.info("[DocPicker] Copied to App Group container: \(destFullPath.path)")
            return destFullPath.path
        } catch {
            Logger.info("[DocPicker][Err] Copy to App Group container failed for \(url.path) to \(destFullPath.path): \(error)")
            return nil
        }
    }
    
    private func removeTempFolderByTimestamp(_ timestamp: String) {
        guard let groupContainerURL = FileManager.default.containerURL(forSecurityApplicationGroupIdentifier: UserDefaults.groupId) else {
            Logger.info("[DocPicker][Err] Could not get App Group container URL for cleanup. GroupID: \(UserDefaults.groupId)")
            return
        }
        let root = groupContainerURL.appendingPathComponent("SharedFiles").appendingPathComponent(timestamp)
        if FileManager.default.fileExists(atPath: root.path) {
            do {
                try FileManager.default.removeItem(at: root)
                Logger.info("[DocPicker] Cleaned files from App Group successfully: \(root.path)")
            } catch {
                Logger.info("[DocPicker][Err] Clean files from App Group failed: \(error)")
            }
        } else {
            Logger.info("[DocPicker][Info] Clean files from App Group: Path not found, nothing to clean: \(root.path)")
        }
    }
    
    private func cancelExtensionContext(with errorMsg: String,duration: TimeInterval = 0.5) {
        // MBProgressHUD.showTips(.error,errorMsg, toView: self.view,duration: duration)
        DispatchQueue.main.asyncAfter(deadline: .now() + duration) {
            self.extensionContext?.cancelRequest(withError: NSError(domain: errorMsg, code: 500))
        }
    }
}
