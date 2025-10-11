//
//  ClipboardMonitor.swift
//  CrossShare
//
//  Created by user00 on 2025/3/7.
//

import Cocoa

enum ClipboardContentType {
    case text(String)
    case image(NSImage)
    case files([URL])
    case unknown
}

struct ClipboardContent {
    let type: ClipboardContentType
    let timestamp: Date
    
    var description: String {
        switch type {
        case .text(let content):
            return "Text: \(content)"
        case .image(_):
            return "Image"
        case .files(let urls):
            return "Files: \(urls.map { $0.lastPathComponent }.joined(separator: ", "))"
        case .unknown:
            return "Unknown content type"
        }
    }
    
    func toDictionary() -> [String: Any] {
        var dict: [String: Any] = [
            "timestamp": timestamp.timeIntervalSince1970
        ]
        
        switch type {
        case .text(let content):
            dict["type"] = "text"
            dict["content"] = content
            
        case .image(let image):
            dict["type"] = "image"
            dict["width"] = image.size.width
            dict["height"] = image.size.height
            if let imageData = image.tiffRepresentation {
                dict["dataSize"] = imageData.count
                dict["base64Data"] = imageData.base64EncodedString()
            }
            
        case .files(let urls):
            dict["type"] = "files"
            dict["count"] = urls.count
            dict["paths"] = urls.map { $0.path }
            dict["fileInfo"] = urls.map { url in
                var info: [String: Any] = ["path": url.path]
                if let attributes = try? FileManager.default.attributesOfItem(atPath: url.path) {
                    info["size"] = attributes[.size] ?? 0
                    info["modificationDate"] = (attributes[.modificationDate] as? Date)?.timeIntervalSince1970 ?? 0
                    info["isDirectory"] = (attributes[.type] as? FileAttributeType) == .typeDirectory
                }
                return info
            }
        case .unknown:
            dict["type"] = "unknown"
        }
        return dict
    }
}

protocol ClipboardMonitorDelegate: AnyObject {
    func clipboardDidChange(_ content: ClipboardContent)
}

class ClipboardMonitor {
    private var lastChangeCount = NSPasteboard.general.changeCount
    private var clipboardQueue: DispatchQueue?
    private var clipboardTimer: DispatchSourceTimer?
    private weak var delegate: ClipboardMonitorDelegate?

    private let remoteContentQueue = DispatchQueue(label: "com.crossShare.remoteContent", qos: .userInitiated)

    private static let shared = ClipboardMonitor()

    private init() {}
    
    public static func shareInstance() -> ClipboardMonitor {
        return shared
    }
    
    func startMonitoring() {
        clipboardQueue = DispatchQueue(label: "com.crossShare.clipboardMonitor", qos: .background)
        clipboardTimer = DispatchSource.makeTimerSource(queue: clipboardQueue)
        clipboardTimer?.schedule(deadline: .now(), repeating: 1.0)
        clipboardTimer?.setEventHandler { [weak self] in
            self?.checkClipboard()
        }
        clipboardTimer?.resume()
    }
    
    func stopMonitoring() {
        clipboardTimer?.cancel()
        clipboardTimer = nil
        clipboardQueue = nil
    }
    
    func setDelegate(_ delegate: ClipboardMonitorDelegate) {
        self.delegate = delegate
    }
    
    func getCurrentClipboardContent() -> [String: Any]? {
        let content = readClipboardContent()
        return content.toDictionary()
    }
    
    func sendTextToClipboard(_ text: String) {
        _ = setupClipboard(text: text, image: nil)
    }
    
    func sendImageToClipboard(_ imageData: Data) -> Bool {
        guard let image = NSImage(data: imageData) else { return false }
        return setupClipboard(text: nil, image: image)
    }

    func setupClipboard(text: String?, image: NSImage?) -> Bool {
        var result = false

        remoteContentQueue.sync {
            if text == nil && image == nil {
                result = false
                return
            }

            let pasteboard = NSPasteboard.general
            pasteboard.clearContents()

            if let text = text {
                pasteboard.setString(text, forType: .string)
            }
            if let image = image {
                pasteboard.writeObjects([image])
            }

            if pasteboard.changeCount == 0 {
                print("[ClipboardMonitor] Setup clipboard failed. Invalid changed")
                result = false
                return
            }

            if lastChangeCount != pasteboard.changeCount {
                print("[ClipboardMonitor] Setup clipboard. ChangeCount (last, current)=(\(lastChangeCount), \(pasteboard.changeCount))")
                lastChangeCount = pasteboard.changeCount
                result = true
                return
            }
        }

        return result
    }

    private func checkClipboard() {
        let pasteboard = NSPasteboard.general
        if pasteboard.changeCount != lastChangeCount {
            lastChangeCount = pasteboard.changeCount
            handleClipboardChange()
        }
    }
    
    private func handleClipboardChange() {
        let content = readClipboardContent()
        logClipboardChange(content)
        delegate?.clipboardDidChange(content)
        sendContentToGoService(content)
    }
    
    private func readClipboardContent() -> ClipboardContent {
        let pasteboard = NSPasteboard.general
        let timestamp = Date()
        
        if let fileURLs = pasteboard.readObjects(forClasses: [NSURL.self], options: nil) as? [URL] {
            let fileOnlyURLs = fileURLs.filter { $0.isFileURL }
            if !fileOnlyURLs.isEmpty {
                return ClipboardContent(type: .files(fileOnlyURLs), timestamp: timestamp)
            }
        }
        
        if let image = NSImage(pasteboard: pasteboard) {
            return ClipboardContent(type: .image(image), timestamp: timestamp)
        }
        
        if let text = pasteboard.string(forType: .string), !text.isEmpty {
            return ClipboardContent(type: .text(text), timestamp: timestamp)
        }
        
        return ClipboardContent(type: .unknown, timestamp: timestamp)
    }
    
    private func logClipboardChange(_ content: ClipboardContent) {
        switch content.type {
        case .text(let text):
            print("[ClipboardMonitor] Text copied: \(text.prefix(100))\(text.count > 100 ? "..." : "")")
        case .image(let image):
            print("[ClipboardMonitor] Image copied: \(image.size.width)x\(image.size.height)")
        case .files(let urls):
            print("[ClipboardMonitor] Files copied:")
            for url in urls {
                print("  - \(url.path)")
            }
        case .unknown:
            print("[ClipboardMonitor] Unknown content type copied")
        }
    }
    
    private func sendContentToGoService(_ content: ClipboardContent) {
        switch content.type {
        case .text(let text):
            print("[ClipboardMonitor] Local copy text event: \(text.prefix(100))\(text.count > 100 ? "..." : "")")
            SendXClipData(text.toGoStringXPC(), "".toGoStringXPC(), "".toGoStringXPC())
        case .image(let image):
            if let imageBase64 = convertImageToBase64(image) {
                print("[ClipboardMonitor] Local copy image event: \(image.size.width)x\(image.size.height)")
                SendXClipData("".toGoStringXPC(), imageBase64.toGoStringXPC(), "".toGoStringXPC())
            } else {
                print("[ClipboardMonitor][Err] Failed to convert image to base64")
            }

        case .files(let urls):
            print("[ClipboardMonitor] Files copied:")
            for url in urls {
                print("  - \(url.path)")
            }
        case .unknown:
            print("[ClipboardMonitor][Err] Unknown content type")
        }
    }

    private func convertImageToBase64(_ image: NSImage) -> String? {
        guard let tiffData = image.tiffRepresentation,
              let bitmapRep = NSBitmapImageRep(data: tiffData) else {
            return nil
        }

        if let jpegData = bitmapRep.representation(using: .jpeg, properties: [.compressionFactor: 0.8]) {
            return jpegData.base64EncodedString()
        }

        return nil
    }
}

extension NSImage {
    var pngData: Data? {
        guard let tiffRepresentation = self.tiffRepresentation,
              let bitmapImage = NSBitmapImageRep(data: tiffRepresentation) else { return nil }
        return bitmapImage.representation(using: .png, properties: [:])
    }
    
    var jpegData: Data? {
        guard let tiffRepresentation = self.tiffRepresentation,
              let bitmapImage = NSBitmapImageRep(data: tiffRepresentation) else { return nil }
        return bitmapImage.representation(using: .jpeg, properties: [.compressionFactor: 0.8])
    }
}
