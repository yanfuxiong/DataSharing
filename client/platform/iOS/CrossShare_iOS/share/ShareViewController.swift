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
                    if let url = item as? URL {
                        Logger.info("Got file URL directly: \(url)")
                        self.convertURLToLocalFile(url: url) { localFileURL in
                            if let localFileURL = localFileURL {
                                filePaths.append(localFileURL)
                            }
                            dispatchGroup.leave()
                        }
                    } else if let urlData = item as? Data, let url = URL(dataRepresentation: urlData, relativeTo: nil) {
                        Logger.info("Got file URL from data: \(url)")
                        self.convertURLToLocalFile(url: url) { localFileURL in
                            if let localFileURL = localFileURL {
                                filePaths.append(localFileURL)
                            }
                            dispatchGroup.leave()
                        }
                    } else if let urlString = item as? String, let url = URL(string: urlString) {
                        Logger.info("Got file URL from string: \(url)")
                        self.convertURLToLocalFile(url: url) { localFileURL in
                            if let localFileURL = localFileURL {
                                filePaths.append(localFileURL)
                            }
                            dispatchGroup.leave()
                        }
                    } else {
                        dispatchGroup.leave()
                    }
                }
            } else if provider.hasItemConformingToTypeIdentifier(UTType.data.identifier) {
                provider.loadItem(forTypeIdentifier: UTType.data.identifier, options: nil) { (item, error) in
                    if let url = item as? URL {
                        Logger.info("Got URL from data type: \(url)")
                        self.convertURLToLocalFile(url: url) { localFileURL in
                            if let localFileURL = localFileURL {
                                filePaths.append(localFileURL)
                            }
                            dispatchGroup.leave()
                        }
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
                        dispatchGroup.leave()
                    } else {
                        dispatchGroup.leave()
                    }
                }
            } else {
                let availableTypeIdentifiers = provider.registeredTypeIdentifiers
                if let firstType = availableTypeIdentifiers.first {
                    provider.loadItem(forTypeIdentifier: firstType, options: nil) { (item, error) in
                        Logger.info("Processing item of type: \(firstType)")
                        if let url = item as? URL {
                            Logger.info("Got URL from generic type: \(url)")
                            self.convertURLToLocalFile(url: url) { localFileURL in
                                if let localFileURL = localFileURL {
                                    filePaths.append(localFileURL)
                                }
                                dispatchGroup.leave()
                            }
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
                            dispatchGroup.leave()
                        } else {
                            Logger.info("Unknown item type: \(String(describing: item))")
                            dispatchGroup.leave()
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
                MBProgressHUD.showTips(.error, "There are no valid files available for transfer", toView: self.view, duration: 1.5)
                DispatchQueue.main.asyncAfter(deadline: .now() + 1.5) {
                    self.cancelExtensionContext(with: "No files selected.")
                }
                return
            }
            if UserDefaults.getBool(forKey: .ISACTIVITY, type: .group) == false {
                Logger.info("Error: ISACTIVITY not found in UserDefaults.")
                MBProgressHUD.showTips(.error, "Please launch the app and connect to the display", toView: self.view, duration: 1.5)
                DispatchQueue.main.asyncAfter(deadline: .now() + 1.5) {
                    self.cancelExtensionContext(with: "Please open CrossShare app first and wait a moment")
                }
                return
            }
            self.showDeviceSelectPopView(urls: filePaths)
        }
    }
    
    /// Convert URL to local file path
    /// - Parameters:
    ///   - url: URL to convert (could be local file, remote URL, or iCloud URL)
    ///   - completion: Completion callback, returns local file URL, or nil if conversion fails (callback executes on main thread)
    private func convertURLToLocalFile(url: URL, completion: @escaping (URL?) -> Void) {
        // Check if it's a local file URL
        if url.scheme == "file" && !url.path.contains("iCloud") {
            // Check if file exists
            if FileManager.default.fileExists(atPath: url.path) {
                Logger.info("[URLConverter] Local file exists: \(url.path)")
                DispatchQueue.main.async {
                    completion(url)
                }
                return
            } else {
                Logger.info("[URLConverter] Local file not found: \(url.path)")
                DispatchQueue.main.async {
                    completion(nil)
                }
                return
            }
        }
        
        // Check if it's a remote URL (http/https)
        if url.scheme == "http" || url.scheme == "https" {
            Logger.info("[URLConverter] Detected remote URL, downloading: \(url.absoluteString)")
            downloadRemoteURL(url: url, completion: completion)
            return
        }
        
        // Check if it's an iCloud URL (contains Mobile Documents or iCloud)
        if url.scheme == "file" && (url.path.contains("iCloud") || url.path.contains("Mobile Documents")) {
            Logger.info("[URLConverter] Detected iCloud URL: \(url.path)")
            handleiCloudURL(url: url, completion: completion)
            return
        }
        
        // Try to handle as file URL (could be other format of file URL)
        if url.scheme == "file" || url.isFileURL {
            var securityScopedAccess = false
            if url.startAccessingSecurityScopedResource() {
                securityScopedAccess = true
            }
            defer {
                if securityScopedAccess {
                    url.stopAccessingSecurityScopedResource()
                }
            }
            
            if FileManager.default.fileExists(atPath: url.path) {
                Logger.info("[URLConverter] File URL accessible: \(url.path)")
                DispatchQueue.main.async {
                    completion(url)
                }
            } else {
                Logger.info("[URLConverter] File URL not accessible: \(url.path)")
                DispatchQueue.main.async {
                    completion(nil)
                }
            }
            return
        }
        
        // Unknown URL type
        Logger.info("[URLConverter] Unknown URL scheme: \(url.scheme ?? "nil")")
        DispatchQueue.main.async {
            completion(nil)
        }
    }
    
    /// Download remote URL and save as temporary file
    private func downloadRemoteURL(url: URL, completion: @escaping (URL?) -> Void) {
        let tempDirectory = FileManager.default.temporaryDirectory
        let fileName = UUID().uuidString
        // Try to get file extension from URL
        let pathExtension = url.pathExtension.isEmpty ? "" : ".\(url.pathExtension)"
        let tempFileURL = tempDirectory.appendingPathComponent(fileName + pathExtension)
        
        let task = URLSession.shared.downloadTask(with: url) { (tempURL, response, error) in
            if let error = error {
                Logger.info("[URLConverter][Err] Download failed: \(error.localizedDescription)")
                DispatchQueue.main.async {
                    completion(nil)
                }
                return
            }
            
            guard let tempURL = tempURL else {
                Logger.info("[URLConverter][Err] No temporary file URL from download")
                DispatchQueue.main.async {
                    completion(nil)
                }
                return
            }
            
            do {
                // If target file already exists, remove it first
                if FileManager.default.fileExists(atPath: tempFileURL.path) {
                    try FileManager.default.removeItem(at: tempFileURL)
                }
                
                // Move downloaded file to target location
                try FileManager.default.moveItem(at: tempURL, to: tempFileURL)
                Logger.info("[URLConverter] Downloaded and saved to: \(tempFileURL.path)")
                DispatchQueue.main.async {
                    completion(tempFileURL)
                }
            } catch {
                Logger.info("[URLConverter][Err] Failed to move downloaded file: \(error)")
                DispatchQueue.main.async {
                    completion(nil)
                }
            }
        }
        
        task.resume()
    }
    
    /// Handle iCloud URL (requires security-scoped resource access)
    private func handleiCloudURL(url: URL, completion: @escaping (URL?) -> Void) {
        guard url.startAccessingSecurityScopedResource() else {
            Logger.info("[URLConverter][Err] Cannot access security-scoped resource for iCloud URL: \(url.path)")
            DispatchQueue.main.async {
                completion(nil)
            }
            return
        }
        
        let fileManager = FileManager.default
        var isDirectory: ObjCBool = false
        
        guard fileManager.fileExists(atPath: url.path, isDirectory: &isDirectory) else {
            url.stopAccessingSecurityScopedResource()
            Logger.info("[URLConverter][Err] iCloud file does not exist: \(url.path)")
            DispatchQueue.main.async {
                completion(nil)
            }
            return
        }
        
        // If it's a directory, not supported
        if isDirectory.boolValue {
            url.stopAccessingSecurityScopedResource()
            Logger.info("[URLConverter][Err] iCloud item is a directory, not supported: \(url.path)")
            DispatchQueue.main.async {
                completion(nil)
            }
            return
        }
        
        // Copy iCloud file to temporary directory
        let tempDirectory = FileManager.default.temporaryDirectory
        let fileName = url.lastPathComponent
        let tempFileURL = tempDirectory.appendingPathComponent(fileName)
        
        do {
            // If target file already exists, remove it first
            if fileManager.fileExists(atPath: tempFileURL.path) {
                try fileManager.removeItem(at: tempFileURL)
            }
            
            // Copy file
            try fileManager.copyItem(at: url, to: tempFileURL)
            url.stopAccessingSecurityScopedResource() // Close security-scoped access after copy completes
            Logger.info("[URLConverter] Copied iCloud file to temporary location: \(tempFileURL.path)")
            DispatchQueue.main.async {
                completion(tempFileURL)
            }
        } catch {
            url.stopAccessingSecurityScopedResource() // Also close security-scoped access on error
            Logger.info("[URLConverter][Err] Failed to copy iCloud file: \(error)")
            DispatchQueue.main.async {
                completion(nil)
            }
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
        self.view.addSubview(popView)
        popView.setupUI()
        
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
                        ip: item[ClientInfo.CodingKeys.ip.rawValue].stringValue,
                        id: item[ClientInfo.CodingKeys.id.rawValue].stringValue,
                        name: item[ClientInfo.CodingKeys.name.rawValue].stringValue,
                        deviceType: item[ClientInfo.CodingKeys.deviceType.rawValue].stringValue,
                        sourcePortType: item[ClientInfo.CodingKeys.sourcePortType.rawValue].stringValue
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
            self.view.addSubview(popView)
            popView.setupUI()
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
            
            self.view.addSubview(popView)
            UIView.animate(withDuration: 0.3) {
                popView.alpha = 1
                if let contentView = popView.subviews.first {
                    contentView.transform = .identity
                }
            }
        } else {
            print("No clients available to share with.")
            MBProgressHUD.showTips(.error, "No devices available for file transfer", toView: self.view)
            DispatchQueue.main.asyncAfter(deadline: .now() + 1.5) {
                self.cancelExtensionContext(with: "No clients found")
            }
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
        
        // Check if source file is in system temporary directory (temporary file from URL conversion)
        let isSystemTempFile = url.path.hasPrefix(FileManager.default.temporaryDirectory.path)
        
        do {
            try FileManager.default.createDirectory(at: destFolder, withIntermediateDirectories: true, attributes: nil)
            if FileManager.default.fileExists(atPath: destFullPath.path) {
                try FileManager.default.removeItem(at: destFullPath)
            }
            try FileManager.default.copyItem(at: url, to: destFullPath)
            Logger.info("[DocPicker] Copied to App Group container: \(destFullPath.path)")
            
            // If source file is a system temporary file (from URL conversion), delete it after successful copy
            if isSystemTempFile {
                do {
                    try FileManager.default.removeItem(at: url)
                    Logger.info("[DocPicker] Cleaned up temporary file from system temp directory: \(url.path)")
                } catch {
                    Logger.info("[DocPicker][Warn] Failed to clean up temporary file: \(url.path), error: \(error)")
                }
            }
            
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
