//
//  ClipboardMonitor.swift
//  CrossShare
//
//  Created by user00 on 2025/3/7.
//

import UIKit
import UniformTypeIdentifiers

class ClipboardMonitor {
    private var lastChangeCount = UIPasteboard.general.changeCount
    private var clipboardQueue = DispatchQueue(label: "com.crossShare.clipboardMonitor", qos: .background)
    private var clipboardTimer: DispatchSourceTimer?
    private let intervalTime = 1.0
    
    private static let shared = ClipboardMonitor()
    
    private init() {}
    
    public static func shareInstance() -> ClipboardMonitor {
        return shared
    }
    
    func startMonitoring() {
        clipboardTimer = DispatchSource.makeTimerSource(queue: clipboardQueue)
        clipboardTimer?.schedule(deadline: .now(), repeating: intervalTime)
        clipboardTimer?.setEventHandler { [weak self] in
            self?.checkClipboard()
        }
        clipboardTimer?.resume()
    }
    
    func setupClipboard(text: String?, image: UIImage?, html: String? = nil) -> Bool {
        var result = false

        clipboardQueue.sync {
            if text == nil && image == nil && html == nil {
                result = false
                return
            }

            let pasteboard = UIPasteboard.general
            pasteboard.items = []

            var items: [[String: Any]] = []
            var item: [String: Any] = [:]

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
                    item[UTType.html.identifier] = htmlData
                }
            }

            if let text = text, !text.isEmpty {
                item[UTType.plainText.identifier] = text
            }

            if let image = image {
                item[UTType.png.identifier] = image.pngData()
            }

            items.append(item)
            pasteboard.items = items

            if pasteboard.changeCount == 0 {
                Logger.error("Setup clipboard failed. Invalid changed")
                result = false
                return
            }

            if lastChangeCount != pasteboard.changeCount {
                lastChangeCount = pasteboard.changeCount
                result = true
                return
            }
        }

        return result
    }
    
    private var printErrCheckClipboard = true
    private func checkClipboard() {
        let pasteboard = UIPasteboard.general
        if pasteboard.changeCount == lastChangeCount {
            return
        }
        
        if pasteboard.changeCount == 0 {
            if printErrCheckClipboard {
                Logger.error("Detect clipboard invalid changed")
                printErrCheckClipboard = false
            }
            return
        }
        
        Logger.info("Detect clipboard changed. ChangeCount (last, current)=(\(lastChangeCount), \(pasteboard.changeCount))")
        Logger.info("Detect clipboard changed. Pasteboard.types=\(pasteboard.types)")
        lastChangeCount = pasteboard.changeCount
        printErrCheckClipboard = true
        
        DispatchQueue.main.async { [weak self] in
            self?.handleClipboardChange()
        }
    }
    
    private func handleClipboardChange() {
        let (text, image, html) = readMultipleClipboardTypes()

        if html != nil || text != nil || image != nil {
            var processedHTML = html ?? ""

            if let htmlString = html, htmlString.contains("file:///") {
                processedHTML = convertLocalImagesToBase64(html: htmlString, pasteboard: UIPasteboard.general)
            }

            let textToSend = text ?? ""
            let imageBase64 = image?.imageToBase64() ?? ""

            SendXClipData(textToSend.toGoString(), imageBase64.toGoString(), processedHTML.toGoString())

            if let copiedImage = image {
                PictureInPictureManager.shared.showImageReceived(copiedImage)
            } else if !textToSend.isEmpty {
                PictureInPictureManager.shared.showTextReceived(textToSend)
            }
        }
    }

    func readMultipleClipboardTypes() -> (text: String?, image: UIImage?, html: String?) {
        let pasteboard = UIPasteboard.general

        if pasteboard.types.count == 0 {
            return (nil, nil, nil)
        }

        var text: String? = nil
        var image: UIImage? = nil
        var html: String? = nil

        let htmlType = UTType.html.identifier
        if pasteboard.contains(pasteboardTypes: [htmlType]) {
            if let htmlData = pasteboard.data(forPasteboardType: htmlType) {
                html = decodeHTMLData(htmlData)
            }
        }

        if let pasteboardString = pasteboard.string {
            // Ignore URL string
            if let url = URL(string: pasteboardString), UIApplication.shared.canOpenURL(url) {
                Logger.info("[ClipboardMonitor]: Ignore URL string")
            } else {
                text = pasteboardString
            }
        }

        if pasteboard.hasImages {
            image = pasteboard.image
        }

        Logger.info("Read multiple types - text: \(text != nil), image: \(image != nil), html: \(html != nil)")

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
    
    func stopMonitoring() {
        clipboardTimer?.cancel()
        clipboardTimer = nil
    }

    private func convertLocalImagesToBase64(html: String, pasteboard: UIPasteboard) -> String {
        var processedHTML = html
        var images: [UIImage] = []

        if pasteboard.types.contains("com.apple.webarchive") {
            images = extractImagesFromWebArchive(pasteboard: pasteboard)
        }

        if images.isEmpty {
            for (_, item) in pasteboard.items.enumerated() {
                if let imageData = item[UTType.png.identifier] as? Data,
                   let image = UIImage(data: imageData) {
                    images.append(image)
                } else if let imageData = item[UTType.jpeg.identifier] as? Data,
                          let image = UIImage(data: imageData) {
                    images.append(image)
                } else if let imageData = item[UTType.tiff.identifier] as? Data,
                          let image = UIImage(data: imageData) {
                    images.append(image)
                } else if let image = item[UTType.image.identifier] as? UIImage {
                    images.append(image)
                }
            }
        }

        guard let regex = try? NSRegularExpression(pattern: "<img[^>]*src=\"file:///(.*?)\"[^>]*>", options: [.caseInsensitive]) else {
            return html
        }

        let matches = regex.matches(in: html, range: NSRange(html.startIndex..., in: html))
        var replacements: [(range: Range<String.Index>, newTag: String)] = []

        for (index, match) in matches.enumerated() {
            if index < images.count, let range = Range(match.range, in: html) {
                let image = images[index]

                let compressedData: Data?
                let mimeType: String

                if let jpegData = image.jpegData(compressionQuality: 0.7) {
                    compressedData = jpegData
                    mimeType = "image/jpeg"
                } else if let pngData = image.pngData() {
                    compressedData = pngData
                    mimeType = "image/png"
                } else {
                    compressedData = nil
                    mimeType = "image/png"
                }

                if let imageData = compressedData,
                   let base64String = imageData.base64EncodedString(options: []) as String? {
                    let originalTag = String(html[range])
                    let dataURL = "data:\(mimeType);base64,\(base64String)"

                    let displayWidth = image.size.width / UIScreen.main.scale
                    let displayHeight = image.size.height / UIScreen.main.scale

                    let styleAttr = "style=\"max-width: 100%; height: auto;\""
                    let widthAttr = "width=\"\(Int(displayWidth))\""
                    let heightAttr = "height=\"\(Int(displayHeight))\""

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

    private func extractImagesFromWebArchive(pasteboard: UIPasteboard) -> [UIImage] {
        var images: [UIImage] = []

        guard let webArchiveData = pasteboard.data(forPasteboardType: "com.apple.webarchive"),
              let webArchive = try? PropertyListSerialization.propertyList(from: webArchiveData, format: nil) as? [String: Any] else {
            return images
        }

        if let subresources = webArchive["WebSubresources"] as? [[String: Any]] {
            for resource in subresources {
                if let mimeType = resource["WebResourceMIMEType"] as? String,
                   mimeType.hasPrefix("image/"),
                   let imageData = resource["WebResourceData"] as? Data,
                   let image = UIImage(data: imageData) {
                    images.append(image)
                }
            }
        }

        if images.isEmpty,
           let mainResource = webArchive["WebMainResource"] as? [String: Any],
           let mimeType = mainResource["WebResourceMIMEType"] as? String,
           mimeType.hasPrefix("image/"),
           let imageData = mainResource["WebResourceData"] as? Data,
           let image = UIImage(data: imageData) {
            images.append(image)
        }

        return images
    }
}
