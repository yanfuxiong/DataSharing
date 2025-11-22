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
    case html(String)
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
        case .html(let html):
            return "HTML (\(html.count) chars)"
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

        case .html(let html):
            dict["type"] = "html"
            dict["html"] = html
            dict["htmlLength"] = html.count

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
    func clipboardDidChange(text: String?, image: NSImage?, html: String?)
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

    func setupClipboard(text: String?, image: NSImage?, html: String? = nil) -> Bool {
        var result = false

        remoteContentQueue.sync {
            if text == nil && image == nil && html == nil {
                result = false
                return
            }

            let pasteboard = NSPasteboard.general
            pasteboard.clearContents()

            if let html = html, !html.isEmpty {
                var wrappedHTML = html
                if !wrappedHTML.lowercased().contains("<meta charset=\"UTF-8\">") {
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
                    pasteboard.setData(htmlData, forType: .html)
                }

                print("[ClipboardMonitor] Setup clipboard with HTML content, length: \(wrappedHTML.count)")
            }

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
        let (text, image, html) = readMultipleClipboardTypes()
        delegate?.clipboardDidChange(text: text, image: image, html: html)
    }

    func readMultipleClipboardTypes() -> (text: String?, image: NSImage?, html: String?) {
        let pasteboard = NSPasteboard.general

        var text: String? = nil
        var image: NSImage? = nil
        var html: String? = nil

        if let items = pasteboard.pasteboardItems {
            for item in items {
                if text == nil {
                    if item.types.contains(.string) {
                        if let textTmp = item.string(forType: .string){
                            let filteredText = textTmp.unicodeScalars.filter { $0.value != 0xFFFC }
                            let finalText = String(String.UnicodeScalarView(filteredText))
                            
                            if !finalText.isEmpty {
                                text = finalText
                            }
                        }
                    }
                }

                if image == nil {
                    if item.types.contains(.tiff),
                       let imageData = item.data(forType: .tiff) {
                        image = NSImage(data: imageData)
                    } else if item.types.contains(.png),
                              let imageData = item.data(forType: .png) {
                        image = NSImage(data: imageData)
                    } else if item.types.contains(NSPasteboard.PasteboardType(rawValue: "public.jpeg")),
                              let imageData = item.data(forType: NSPasteboard.PasteboardType(rawValue: "public.jpeg")) {
                        image = NSImage(data: imageData)
                    }
                }

                if html == nil {
                    if item.types.contains(.html),
                       let htmlData = item.data(forType: .html) {
                        html = decodeHTMLData(htmlData)
                    }
                }
            }
        }

        if image == nil {
            image = NSImage(pasteboard: pasteboard)
        }

        print("[ClipboardMonitor] Read multiple types - text: \(text != nil), image: \(image != nil), html: \(html != nil)")

        return (text, image, html)
    }

    private func decodeHTMLData(_ htmlData: Data) -> String? {
        if let utf8String = String(data: htmlData, encoding: .utf8),
           !utf8String.contains("�") && !utf8String.contains("\u{FFFD}") {
            return utf8String
        }

        let encoding = CFStringConvertEncodingToNSStringEncoding(
            CFStringEncoding(CFStringEncodings.GB_18030_2000.rawValue))
        if let gb18030String = String(data: htmlData, encoding: String.Encoding(rawValue: encoding)) {
            return gb18030String
        }

        if let utf16String = String(data: htmlData, encoding: .utf16) {
            return utf16String
        }

        if let utf16LEString = String(data: htmlData, encoding: .utf16LittleEndian) {
            return utf16LEString
        }

        if let utf16BEString = String(data: htmlData, encoding: .utf16BigEndian) {
            return utf16BEString
        }

        if let latinString = String(data: htmlData, encoding: .isoLatin1) {
            return latinString
        }

        return nil
    }
    
    func processHTMLForSending(_ html: String) -> (text: String, imageBase64: String,processedHTML: String) {
        let pasteboard = NSPasteboard.general
        var processedHTML = html

        if html.contains("file:///") {
            processedHTML = convertLocalImagesToBase64(html: html, pasteboard: pasteboard)
        }
        return ("", "",processedHTML)
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

    private func convertImageToBase64FromPasteboard(_ pasteboard: NSPasteboard) -> String? {
        if let image = NSImage(pasteboard: pasteboard) {
            return convertImageToBase64(image)
        }
        return nil
    }

    private func convertLocalImagesToBase64InHTML(html: String, images: [NSImage]) -> String {
        var processedHTML = html

        guard let regex = try? NSRegularExpression(
            pattern: "<img[^>]*src=\"file:///(.*?)\"[^>]*>",
            options: [.caseInsensitive]
        ) else {
            return html
        }

        let matches = regex.matches(in: html, range: NSRange(html.startIndex..., in: html))
        var replacements: [(range: Range<String.Index>, newTag: String)] = []

        for (index, match) in matches.enumerated() {
            if index < images.count, let range = Range(match.range, in: html) {
                let image = images[index]

                guard let tiffData = image.tiffRepresentation,
                      let bitmapRep = NSBitmapImageRep(data: tiffData) else {
                    continue
                }

                let compressedData: Data?
                let mimeType: String

                if let jpegData = bitmapRep.representation(
                    using: .jpeg,
                    properties: [.compressionFactor: 0.7]
                ) {
                    compressedData = jpegData
                    mimeType = "image/jpeg"
                } else if let pngData = bitmapRep.representation(using: .png, properties: [:]) {
                    compressedData = pngData
                    mimeType = "image/png"
                } else {
                    continue
                }

                if let imageData = compressedData {
                    let base64String = imageData.base64EncodedString()
                    let originalTag = String(html[range])
                    let dataURL = "data:\(mimeType);base64,\(base64String)"

                    let displayWidth = Int(image.size.width)
                    let displayHeight = Int(image.size.height)

                    let styleAttr = "style=\"max-width: 100%; height: auto;\""
                    let widthAttr = "width=\"\(displayWidth)\""
                    let heightAttr = "height=\"\(displayHeight)\""

                    let newTag = originalTag.replacingOccurrences(
                        of: "src=\"file:///(.*?)\"",
                        with: "src=\"\(dataURL)\" \(widthAttr) \(heightAttr) \(styleAttr)",
                        options: .regularExpression
                    )
                    replacements.append((range: range, newTag: newTag))
                }
            }
        }

        for replacement in replacements.reversed() {
            processedHTML.replaceSubrange(replacement.range, with: replacement.newTag)
        }

        return processedHTML
    }

    private func convertLocalImagesToBase64(html: String, pasteboard: NSPasteboard) -> String {
        var processedHTML = html
        var images: [NSImage] = []

        if images.isEmpty, let pasteboardItems = pasteboard.pasteboardItems {
            for item in pasteboardItems {
                if let imageData = item.data(forType: .png) ??
                                  item.data(forType: .tiff) ??
                                  item.data(forType: NSPasteboard.PasteboardType(rawValue: "public.jpeg")),
                   let image = NSImage(data: imageData) {
                    images.append(image)
                }
            }
        }

        if images.isEmpty, let image = NSImage(pasteboard: pasteboard) {
            images.append(image)
        }

        guard let regex = try? NSRegularExpression(
            pattern: "<img[^>]*src=\"file:///(.*?)\"[^>]*>",
            options: [.caseInsensitive]
        ) else {
            return html
        }

        let matches = regex.matches(in: html, range: NSRange(html.startIndex..., in: html))
        var replacements: [(range: Range<String.Index>, newTag: String)] = []

        for (index, match) in matches.enumerated() {
            if index < images.count, let range = Range(match.range, in: html) {
                let image = images[index]

                guard let tiffData = image.tiffRepresentation,
                      let bitmapRep = NSBitmapImageRep(data: tiffData) else {
                    continue
                }

                let compressedData: Data?
                let mimeType: String

                if let jpegData = bitmapRep.representation(
                    using: .jpeg,
                    properties: [.compressionFactor: 0.7]
                ) {
                    compressedData = jpegData
                    mimeType = "image/jpeg"
                } else if let pngData = bitmapRep.representation(using: .png, properties: [:]) {
                    compressedData = pngData
                    mimeType = "image/png"
                } else {
                    continue
                }

                if let imageData = compressedData {
                    let base64String = imageData.base64EncodedString()
                    let originalTag = String(html[range])
                    let dataURL = "data:\(mimeType);base64,\(base64String)"

                    let displayWidth = Int(image.size.width)
                    let displayHeight = Int(image.size.height)

                    let styleAttr = "style=\"max-width: 100%; height: auto;\""
                    let widthAttr = "width=\"\(displayWidth)\""
                    let heightAttr = "height=\"\(displayHeight)\""

                    let newTag = originalTag.replacingOccurrences(
                        of: "src=\"file:///(.*?)\"",
                        with: "src=\"\(dataURL)\" \(widthAttr) \(heightAttr) \(styleAttr)",
                        options: .regularExpression
                    )
                    replacements.append((range: range, newTag: newTag))
                }
            }
        }

        for replacement in replacements.reversed() {
            processedHTML.replaceSubrange(replacement.range, with: replacement.newTag)
        }

        if !replacements.isEmpty {
            print("[ClipboardMonitor] Converted \(replacements.count) local images to base64")
        }

        return processedHTML
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
