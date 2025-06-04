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
    
    override func viewDidLoad() {
        super.viewDidLoad()
        handleInput()
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
                        print("Got file URL directly: \(url)")
                        filePaths.append(url)
                    } else if let urlData = item as? Data, let url = URL(dataRepresentation: urlData, relativeTo: nil) {
                        print("Got file URL from data: \(url)")
                        filePaths.append(url)
                    } else if let urlString = item as? String, let url = URL(string: urlString) {
                        print("Got file URL from string: \(url)")
                        filePaths.append(url)
                    }
                }
            } else if provider.hasItemConformingToTypeIdentifier(UTType.data.identifier) {
                provider.loadItem(forTypeIdentifier: UTType.data.identifier, options: nil) { (item, error) in
                    defer { dispatchGroup.leave() }
                    if let url = item as? URL {
                        print("Got URL from data type: \(url)")
                        filePaths.append(url)
                    } else if let data = item as? Data {
                        let tempDirectory = FileManager.default.temporaryDirectory
                        let fileName = UUID().uuidString
                        let tempFileURL = tempDirectory.appendingPathComponent(fileName)
                        do {
                            try data.write(to: tempFileURL)
                            print("Saved data to temporary file: \(tempFileURL)")
                            filePaths.append(tempFileURL)
                        } catch {
                            print("Error saving data to temporary file: \(error)")
                        }
                    }
                }
            } else {
                let availableTypeIdentifiers = provider.registeredTypeIdentifiers
                if let firstType = availableTypeIdentifiers.first {
                    provider.loadItem(forTypeIdentifier: firstType, options: nil) { (item, error) in
                        defer { dispatchGroup.leave() }
                        print("Processing item of type: \(firstType)")
                        if let url = item as? URL {
                            print("Got URL from generic type: \(url)")
                            filePaths.append(url)
                        } else if let data = item as? Data {
                            let tempDirectory = FileManager.default.temporaryDirectory
                            let fileName = UUID().uuidString
                            let fileExtension = self.getFileExtension(from: firstType)
                            let tempFileURL = tempDirectory.appendingPathComponent(fileName + fileExtension)
                            do {
                                try data.write(to: tempFileURL)
                                print("Saved generic data to temporary file: \(tempFileURL)")
                                filePaths.append(tempFileURL)
                            } catch {
                                print("Error saving generic data: \(error)")
                            }
                        } else {
                            print("Unknown item type: \(String(describing: item))")
                        }
                    }
                } else {
                    dispatchGroup.leave()
                    print("No available type identifiers for provider")
                }
            }
        }
        
        dispatchGroup.notify(queue: .main) {
            if filePaths.isEmpty {
                print("No files found to share.")
                self.cancelExtensionContext(with: "No files selected.")
                return
            }
            self.showDeviceSelectPopView(urls: filePaths)
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
                        name: item["name"].stringValue
                    )
                }
            }
        }
        if urls.isEmpty {
            self.cancelExtensionContext(with: "Error: No URLs provided")
            return
        }
        
        for url in urls {
            if url.startAccessingSecurityScopedResource() {
                fileNames.append(url.lastPathComponent)
                url.stopAccessingSecurityScopedResource()
            } else {
                print("Warning: Could not access security-scoped resource for \(url.lastPathComponent).")
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
                print("select device：\(client.name)")
                selectedClient = client
                MBProgressHUD.showTips(.success,"Device: \(client.name)", toView: self.view, duration: 1.0)
            }
            popView.onCancel = { [weak self] in
                self?.dismissDevicePopView()
                self?.cancelExtensionContext(with: "cancel transport files",duration: 1)
            }
            popView.onSure = { [weak self] in
                guard let self = self else { return }
                guard let currentSelectedClient = selectedClient else {
                    MBProgressHUD.showTips(.warn,"Please select a device", toView: self.view)
                    return
                }
                self.dismissDevicePopView()
                let timestamp = String(Int(Date().timeIntervalSince1970))
                var urlSchemeString: String?
                
                if urls.count == 1, let singleUrl = urls.first {
                    guard let finalFilePath = self.copyDocumentToTemp(singleUrl, timestamp) else {
                        self.cancelExtensionContext(with: "Failed to prepare file.")
                        return
                    }
                    guard let fileSize = self.getFileSize(atPath: finalFilePath) else {
                        self.cancelExtensionContext(with: "Invalid file selected.")
                        return
                    }
                    guard let encodedFilePath = finalFilePath.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) else {
                        print("Error: Could not percent-encode file path")
                        self.cancelExtensionContext(with: "Failed to prepare file path.")
                        return
                    }
                    urlSchemeString = "crossshare://import?filePath=\(encodedFilePath)&clientId=\(currentSelectedClient.id)&clientIp=\(currentSelectedClient.ip)&fileSize=\(fileSize)&isMup=false"
                } else {
                    self.cancelExtensionContext(with: "No files to share.")
                    return
                }
                
                if let scheme = urlSchemeString, let openURL = URL(string: scheme) {
                    self.openURL(openURL)
                    self.extensionContext?.completeRequest(returningItems: [], completionHandler: nil)
                } else {
                    print("Error: Could not create URL from scheme: \(urlSchemeString ?? "nil")")
                    self.cancelExtensionContext(with: "URL creation failed")
                }
            }
            // move -40 for safe area and margin
            popView.frame.origin.y -= 40
            self.view.addSubview(popView)
            UIView.animate(withDuration: 0.3) {
                popView.alpha = 1
                if let contentView = popView.subviews.first {
                    contentView.transform = .identity
                }
            }
        } else {
            print("No clients available to share with.")
            self.cancelExtensionContext(with: "No clients available to share with.")
        }
    }
    
    func openURL(_ url: URL) {
        var responder = self as UIResponder?
        let selector = NSSelectorFromString("openURL:")
        while responder != nil {
            if responder?.responds(to: selector) == true {
                responder?.perform(selector, with: url)
                break
            }
            responder = responder?.next
        }
    }
    
    func getFileSize(atPath path: String) -> Int64? {
        let fileManager = FileManager.default
        do {
            let attributes = try fileManager.attributesOfItem(atPath: path)
            if let fileSize = attributes[FileAttributeKey.size] as? NSNumber {
                return fileSize.int64Value
            }
        } catch {
            print("Error fetching file size: \(error)")
        }
        return nil
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
            print("[DocPicker][Err] Could not get App Group container URL. GroupID: \(UserDefaults.groupId)")
            return nil
        }
        
        var securityScopedAccess = false
        if url.scheme == "file" {
            securityScopedAccess = url.startAccessingSecurityScopedResource()
            if !securityScopedAccess {
                print("[DocPicker][Warn] Could not start security-scoped access for: \(url.path). Will attempt copy anyway.")
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
            print("[DocPicker] Copied to App Group container: \(destFullPath.path)")
            return destFullPath.path
        } catch {
            print("[DocPicker][Err] Copy to App Group container failed for \(url.path) to \(destFullPath.path): \(error)")
            return nil
        }
    }
    
    private func removeTempFolderByTimestamp(_ timestamp: String) {
        guard let groupContainerURL = FileManager.default.containerURL(forSecurityApplicationGroupIdentifier: UserDefaults.groupId) else {
            print("[DocPicker][Err] Could not get App Group container URL for cleanup. GroupID: \(UserDefaults.groupId)")
            return
        }
        let root = groupContainerURL.appendingPathComponent("SharedFiles").appendingPathComponent(timestamp)
        if FileManager.default.fileExists(atPath: root.path) {
            do {
                try FileManager.default.removeItem(at: root)
                print("[DocPicker] Cleaned files from App Group successfully: \(root.path)")
            } catch {
                print("[DocPicker][Err] Clean files from App Group failed: \(error)")
            }
        } else {
            print("[DocPicker][Info] Clean files from App Group: Path not found, nothing to clean: \(root.path)")
        }
    }
    
    private func cancelExtensionContext(with errorMsg: String,duration: TimeInterval = 2.0) {
        MBProgressHUD.showTips(.error,errorMsg, toView: self.view,duration: duration)
        DispatchQueue.main.asyncAfter(deadline: .now() + duration) {
            self.extensionContext?.cancelRequest(withError: NSError(domain: errorMsg, code: 500))
        }
    }
}
