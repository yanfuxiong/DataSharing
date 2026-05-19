//
//  ClipboardMonitor.swift
//  CrossShare
//
//  Created by user00 on 2025/3/7.
//

import Cocoa

protocol ClipboardMonitorDelegate: AnyObject {
    func clipboardDidChange(text: String?, image: NSImage?, html: String?, rtf: String?)
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

    func setupClipboard(text: String?, image: NSImage?, html: String?, rtf: String?) -> Bool {
        guard text != nil || image != nil || html != nil || rtf != nil else {
            logger.log("No valid clipboard content to set", level: .warn)
            return false
        }

        let pasteboard = NSPasteboard.general
        pasteboard.clearContents()

        let pasteboardItem = NSPasteboardItem()
        
        // 预处理RTF：如果有RTF数据，先解析提取图片并转换为 RTFD 格式
        let rtfResult = processRTFData(rtf)
        let rtfImage = rtfResult.image
        let processedRtfData = rtfResult.processedRtfData
        let rtfdData = rtfResult.rtfdData

        // The sequence of PasteboardItem must be IMAGE -> HTML -> RTF -> TEXT (Following Safari's rules)
        // 注意：如果有 RTFD 格式，不要单独写入 Image，避免应用优先使用单独的图片而忽略文本
        let finalImage = rtfImage ?? image
        
        // 只有在没有 RTFD 数据时，才单独写入图片
        if rtfdData == nil, let img = finalImage {
            if let tiffData = img.tiffRepresentation {
                pasteboardItem.setData(tiffData, forType: .tiff)
                logger.log("Added image (TIFF) to pasteboard item, size: \(img.size)", level: .info)
            }
            if let pngData = img.pngData {
                pasteboardItem.setData(pngData, forType: .png)
                logger.log("Added image (PNG) to pasteboard item", level: .info)
            }
        } else if rtfdData != nil {
            logger.log("[Clipboard] RTFD data available, skipping standalone image write", level: .info)
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
                logger.log("Added HTML to pasteboard item, length: \(wrappedHTML.count)", level: .info)
            }
        }
        
        // 写入 RTFD 格式（macOS 原生格式，优先级最高）
        if let rtfdData = rtfdData {
            let rtfdType = NSPasteboard.PasteboardType(rawValue: "com.apple.flat-rtfd")
            pasteboardItem.setData(rtfdData, forType: rtfdType)
            logger.log("[Clipboard] Added RTFD to pasteboard item, data size: \(rtfdData.count) bytes", level: .info)
        }

        // 写入 RTF 格式（通用格式，兼容性好）
        if let rtfData = processedRtfData {
            pasteboardItem.setData(rtfData, forType: .rtf)
            let msRtfType = NSPasteboard.PasteboardType(rawValue: "com.microsoft.rtf")
            pasteboardItem.setData(rtfData, forType: msRtfType)
            
            logger.log("Added RTF to pasteboard item with multiple types, data size: \(rtfData.count) bytes", level: .info)
        }

        // 已写入 RTFD 时勿再写 NSStringPboardType：备忘录等会优先用纯文本，导致只看到字看不到图
        if let text = text, !text.isEmpty, rtfdData == nil {
            pasteboardItem.setString(text, forType: .string)
            logger.log("Added text to pasteboard item", level: .info)
        } else if let text = text, !text.isEmpty, rtfdData != nil {
            logger.log("[Clipboard] Skipping plain string pasteboard type (RTFD present, length \(text.count))", level: .info)
        }

        if !pasteboardItem.types.isEmpty {
            let result = pasteboard.writeObjects([pasteboardItem])
            logger.log("Pasteboard write result: \(result)", level: .info)
            
            // 验证写入后的剪贴板类型
            if let types = pasteboard.pasteboardItems?.first?.types {
                let typeNames = types.map { $0.rawValue }.joined(separator: ", ")
                logger.log("[Clipboard] Pasteboard types: [\(typeNames)]", level: .info)
            }

            if pasteboard.changeCount == 0 {
                logger.log("[ClipboardMonitor] Setup clipboard failed. Invalid changed", level: .error)
                return false
            }

            if lastChangeCount != pasteboard.changeCount {
                logger.log("[ClipboardMonitor] Setup clipboard. ChangeCount (last, current)=(\(lastChangeCount), \(pasteboard.changeCount))", level: .info)
                lastChangeCount = pasteboard.changeCount
                return true
            }
        }
    
        return false
    }

    private func checkClipboard() {
        let pasteboard = NSPasteboard.general
        if pasteboard.changeCount != lastChangeCount {
            lastChangeCount = pasteboard.changeCount
            handleClipboardChange()
        }
    }
    
    private func handleClipboardChange() {
        let (text, image, html, rtf) = readMultipleClipboardTypes()
//        DebugUtils.shared.writeRtfToDesktop(rtf, fileName: "test2.rtf")
        delegate?.clipboardDidChange(text: text, image: image, html: html, rtf: rtf)
    }

    func readMultipleClipboardTypes() -> (text: String?, image: NSImage?, html: String?, rtf: String?) {
        let pasteboard = NSPasteboard.general

        var text: String? = nil
        var image: NSImage? = nil
        var html: String? = nil
        var rtf: String? = nil

        if let items = pasteboard.pasteboardItems {
            for item in items {
                // 提取纯文本
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

                // 提取图片
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

                // 提取 HTML
                if html == nil {
                    if item.types.contains(.html),
                       let htmlData = item.data(forType: .html) {
                        html = decodeHTMLData(htmlData)
                    }
                }
                
                // 提取 RTF（优先使用 com.apple.flat-rtfd，包含图片）
                if rtf == nil {
                    let rtfdType = NSPasteboard.PasteboardType(rawValue: "com.apple.flat-rtfd")
                    if item.types.contains(rtfdType),
                       let rtfdData = item.data(forType: rtfdType) {
                        logger.info("[ClipboardMonitor] Found com.apple.flat-rtfd, size: \(rtfdData.count) bytes")
                        
                        // 使用 RTFBuilder 构建带图片的 RTF
                        let (builtRtf, extractedImage, _) = RTFBuilder.shared.buildRTFFromRTFD(rtfdData)
                        
                        if let builtRtf = builtRtf, !builtRtf.isEmpty {
                            rtf = builtRtf
                            logger.info("[ClipboardMonitor] Built RTF from RTFD, length: \(builtRtf.count) chars")
                        }
                        
                        // 如果没有图片，使用从 RTFD 提取的图片
                        if image == nil, let extractedImage = extractedImage {
                            image = extractedImage
                            logger.info("[ClipboardMonitor] Extracted image from RTFD, size: \(extractedImage.size)")
                        }
                    }
                    // 如果没有 RTFD，尝试普通 RTF
                    else if item.types.contains(.rtf),
                            let rtfData = item.data(forType: .rtf),
                            let rtfString = String(data: rtfData, encoding: .utf8) {
                        rtf = rtfString
                        logger.info("[ClipboardMonitor] Found public.rtf, length: \(rtfString.count) chars")
                    }
                }
            }
        }

        if image == nil {
            image = NSImage(pasteboard: pasteboard)
        }

        logger.info("[ClipboardMonitor] Read multiple types - text: \(text != nil), image: \(image != nil), html: \(html != nil), rtf: \(rtf != nil)")

        return (text, image, html, rtf)
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
    
    // MARK: - RTF 预处理方法
    
    /// 预处理 RTF 数据：提取所有图片、生成 RTFD 格式
    /// - Parameter rtf: 原始 RTF 字符串
    /// - Returns: (第一张图片用于兼容, 处理后的RTF数据, RTFD数据)
    private func processRTFData(_ rtf: String?) -> (image: NSImage?, processedRtfData: Data?, rtfdData: Data?) {
        guard let rtf = rtf, !rtf.isEmpty, let rtfData = rtf.data(using: .utf8) else {
            return (nil, nil, nil)
        }
        
        var firstImage: NSImage? = nil
        var processedRtfData: Data? = nil
        var rtfdData: Data? = nil
        
        logger.log("[RTF] Received RTF data, size: \(rtfData.count) bytes", level: .info)
        
        // 从 RTF 中提取所有图片（anchorInRTF = 该图 \pngblip 起点，用于算插入位置；比 last \pict 更稳，避免多图位置算反）
        var extractedImages: [(image: NSImage, anchorInRTF: String.Index)] = []
        if let rtfString = String(data: rtfData, encoding: .utf8) {
            extractedImages = extractAllImagesFromRTF(rtfString)
            firstImage = extractedImages.first?.image
        }
        
        // 解析 RTF 为 NSAttributedString
        guard let attributedString = try? NSAttributedString(
            data: rtfData,
            options: [.documentType: NSAttributedString.DocumentType.rtf],
            documentAttributes: nil
        ) else {
            processedRtfData = rtfData
            logger.log("[RTF] Failed to parse RTF, using original data", level: .warn)
            return (firstImage, processedRtfData, rtfdData)
        }
        
        logger.log("[RTF] Parsed RTF successfully, text length: \(attributedString.length) chars", level: .info)
        
        let mutableAttrString = NSMutableAttributedString(attributedString: attributedString)
        
        // 系统解析器对 WordPad 的 {\pict\pngblip} 会创建 NSTextAttachment，但常常只设 image 属性、
        // 缺少 fileWrapper 数据，导致 RTFD 输出无图。这里先尝试就地补全 fileWrapper；
        // 补不全的走手动提取路径。
        var totalAttachmentCount = 0
        var repairedCount = 0
        mutableAttrString.enumerateAttribute(.attachment, in: NSRange(location: 0, length: mutableAttrString.length)) { value, _, _ in
            if let attachment = value as? NSTextAttachment {
                totalAttachmentCount += 1
                let hasFileData = attachment.fileWrapper?.regularFileContents?.isEmpty == false
                if !hasFileData, let img = attachment.image,
                   let tiff = img.tiffRepresentation,
                   let rep = NSBitmapImageRep(data: tiff),
                   let pngData = rep.representation(using: .png, properties: [:]),
                   !pngData.isEmpty {
                    let fw = FileWrapper(regularFileWithContents: pngData)
                    fw.preferredFilename = "image\(totalAttachmentCount - 1).png"
                    attachment.fileWrapper = fw
                    repairedCount += 1
                }
            }
        }
        
        // 重新校验：只有 fileWrapper 带实际数据才算有效
        var validAttachmentCount = 0
        mutableAttrString.enumerateAttribute(.attachment, in: NSRange(location: 0, length: mutableAttrString.length)) { value, _, _ in
            if let attachment = value as? NSTextAttachment,
               let fw = attachment.fileWrapper,
               let contents = fw.regularFileContents,
               !contents.isEmpty {
                validAttachmentCount += 1
            }
        }
        let hasValidAttachment = validAttachmentCount > 0 && validAttachmentCount >= totalAttachmentCount
        logger.log("[RTF] Attachment check: total=\(totalAttachmentCount), repaired=\(repairedCount), valid=\(validAttachmentCount), hasValid=\(hasValidAttachment)", level: .info)
        
        // 系统附件全部有效（含修复后的）则沿用系统定位，否则走手动提取
        if !hasValidAttachment, !extractedImages.isEmpty, let rtfString = String(data: rtfData, encoding: .utf8) {
            logger.log("[RTF] Manually adding \(extractedImages.count) extracted images (system attachments incomplete)", level: .info)
            
            // 查找 NSAttributedString 中的所有占位符 \u{FFFC}
            let nsString = mutableAttrString.string as NSString
            var placeholderPositions: [Int] = []
            for i in 0..<nsString.length {
                if nsString.character(at: i) == 0xFFFC {
                    placeholderPositions.append(i)
                }
            }
            
            if placeholderPositions.count == extractedImages.count {
                // 按「在正文中的插入位置」排序后再与占位符顺序对应（Word 里 blip 字节序未必与 FFFC 序一致）
                var ranked: [(image: NSImage, anchor: String.Index, sortPos: Int)] = extractedImages.map { ext in
                    let sortPos = calculateImagePosition(
                        rtfString: rtfString,
                        pictIndex: ext.anchorInRTF,
                        attributedString: mutableAttrString
                    )
                    return (ext.image, ext.anchorInRTF, sortPos)
                }
                ranked.sort { a, b in
                    let aFirstByAnchor = a.anchor < b.anchor
                    let aFirstByPos = a.sortPos < b.sortPos
                    if aFirstByAnchor != aFirstByPos {
                        return aFirstByAnchor
                    }
                    if a.sortPos != b.sortPos { return a.sortPos < b.sortPos }
                    return a.anchor < b.anchor
                }
                for (index, position) in placeholderPositions.enumerated().reversed() {
                    let img = ranked[index].image
                    if let attachment = createImageAttachment(image: img, index: index) {
                        let attachmentString = NSAttributedString(attachment: attachment)
                        mutableAttrString.replaceCharacters(in: NSRange(location: position, length: 1), with: attachmentString)
                        logger.log("[RTF] Replaced placeholder at position \(position) with image (rank \(index + 1))", level: .info)
                    }
                }
            } else {
                // 占位符不匹配，计算每张图片的插入位置，从后往前插入
                var imageInsertions: [(position: Int, imageIndex: Int, anchor: String.Index)] = []
                for (index, extracted) in extractedImages.enumerated() {
                    let position = calculateImagePosition(
                        rtfString: rtfString,
                        pictIndex: extracted.anchorInRTF,
                        attributedString: mutableAttrString
                    )
                    imageInsertions.append((position: position, imageIndex: index, anchor: extracted.anchorInRTF))
                }
                
                imageInsertions.sort { a, b in
                    if a.position != b.position { return a.position > b.position }
                    return a.anchor > b.anchor
                }
                
                for insertion in imageInsertions {
                    let img = extractedImages[insertion.imageIndex].image
                    if let attachment = createImageAttachment(image: img, index: insertion.imageIndex) {
                        let attachmentString = NSAttributedString(attachment: attachment)
                        if insertion.position >= 0 && insertion.position <= mutableAttrString.length {
                            mutableAttrString.insert(attachmentString, at: insertion.position)
                            logger.log("[RTF] Inserted image #\(insertion.imageIndex + 1) at position \(insertion.position)", level: .info)
                        } else {
                            mutableAttrString.append(NSAttributedString(string: "\n"))
                            mutableAttrString.append(attachmentString)
                            logger.log("[RTF] Appended image #\(insertion.imageIndex + 1) at end (position \(insertion.position) out of range)", level: .warn)
                        }
                    }
                }
            }
        }
        
        // 生成 RTFD 格式（macOS 原生格式，支持图片附件）
        if let rtfdDataGenerated = try? mutableAttrString.data(
            from: NSRange(location: 0, length: mutableAttrString.length),
            documentAttributes: [.documentType: NSAttributedString.DocumentType.rtfd]
        ) {
            rtfdData = rtfdDataGenerated
            logger.log("[RTF] Generated RTFD data, size: \(rtfdDataGenerated.count) bytes", level: .info)
        } else {
            logger.log("[RTF] Failed to generate RTFD", level: .warn)
        }
        
        // 同时生成普通 RTF（不包含图片）
        if let rtfDataGenerated = try? mutableAttrString.data(
            from: NSRange(location: 0, length: mutableAttrString.length),
            documentAttributes: [.documentType: NSAttributedString.DocumentType.rtf]
        ) {
            processedRtfData = rtfDataGenerated
            logger.log("[RTF] Generated RTF data, size: \(rtfDataGenerated.count) bytes", level: .info)
        } else {
            processedRtfData = rtfData
            logger.log("[RTF] Failed to generate RTF, using original data", level: .warn)
        }
        
        return (firstImage, processedRtfData, rtfdData)
    }
    
    /// 创建图片附件（RTFD 需要每张图片有唯一文件名）
    private func createImageAttachment(image: NSImage, index: Int = 0) -> NSTextAttachment? {
        let attachment = NSTextAttachment()
        attachment.image = image
        
        guard let imageData = image.tiffRepresentation,
              let bitmapRep = NSBitmapImageRep(data: imageData),
              let pngData = bitmapRep.representation(using: .png, properties: [:]) else {
            logger.log("[RTF] Failed to create FileWrapper for image #\(index + 1)", level: .error)
            return nil
        }
        
        let fileWrapper = FileWrapper(regularFileWithContents: pngData)
        fileWrapper.preferredFilename = "image\(index).png"
        attachment.fileWrapper = fileWrapper
        logger.log("[RTF] Created FileWrapper for image #\(index + 1), data size: \(pngData.count) bytes", level: .info)
        
        return attachment
    }
    
    /// 计算图片在文本中的插入位置
    private func calculateImagePosition(rtfString: String, pictIndex: String.Index, attributedString: NSMutableAttributedString) -> Int {
        // 找到 \pict 所在的最外层 { 的起始位置
        var groupStart = pictIndex
        var braceDepth = 0
        var searchIdx = rtfString.index(before: pictIndex)
        
        while searchIdx > rtfString.startIndex {
            let ch = rtfString[searchIdx]
            if ch == "}" { braceDepth += 1 }
            else if ch == "{" {
                if braceDepth == 0 {
                    groupStart = searchIdx
                    break
                }
                braceDepth -= 1
            }
            searchIdx = rtfString.index(before: searchIdx)
        }
        
        // 截取图片组之前的 RTF 内容
        var beforePictRTF = String(rtfString[..<groupStart])
        
        // 平衡括号：计算未闭合的 { 数量，补齐 }
        var openBraces = 0
        for ch in beforePictRTF {
            if ch == "{" { openBraces += 1 }
            else if ch == "}" { openBraces -= 1 }
        }
        if openBraces > 0 {
            beforePictRTF += String(repeating: "}", count: openBraces)
        }
        
        var imageInsertPosition = 0
        var positionDetermined = false
        
        // 方法1：解析平衡后的截断 RTF
        if let beforeData = beforePictRTF.data(using: .utf8),
           let beforeAttrString = try? NSAttributedString(
            data: beforeData,
            options: [.documentType: NSAttributedString.DocumentType.rtf],
            documentAttributes: nil
           ) {
            imageInsertPosition = beforeAttrString.length
            positionDetermined = true
            logger.log("[RTF] Calculated image position by parsing truncated RTF: \(imageInsertPosition)", level: .info)
        }
        
        // 方法2：正确计数 \par（不计 \pard、\pararsid 等）
        if !positionDetermined {
            let beforePictStr = String(rtfString[..<groupStart])
            var parCount = 0
            var sIdx = beforePictStr.startIndex
            
            while let range = beforePictStr.range(of: "\\par", range: sIdx..<beforePictStr.endIndex) {
                let afterIdx = range.upperBound
                if afterIdx >= beforePictStr.endIndex {
                    parCount += 1
                } else {
                    let nextChar = beforePictStr[afterIdx]
                    if !nextChar.isLetter {
                        parCount += 1
                    }
                }
                sIdx = range.upperBound
            }
            
            let plainText = attributedString.string
            if parCount > 0 {
                var newlineCount = 0
                for (index, char) in plainText.enumerated() {
                    if char == "\n" {
                        newlineCount += 1
                        if newlineCount == parCount {
                            imageInsertPosition = index + 1
                            positionDetermined = true
                            break
                        }
                    }
                }
            }
            logger.log("[RTF] Calculated image position by \\par count: \(parCount), position: \(imageInsertPosition)", level: .info)
        }
        
        return imageInsertPosition
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
            logger.info("[ClipboardMonitor] Converted \(replacements.count) local images to base64")
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

// MARK: - RTF 图片解析扩展
extension ClipboardMonitor {
    
    /// 从 RTF 中提取图片：按 `\pngblip` / `\jpegblip` 遍历，每个 blip 只取一张图。
    /// 不能按 `\pict` 遍历后在前向窗口找 blip：Word 常在同一图前嵌套多个 `\pict`，会反复命中第一个 `\pngblip` 导致图数量暴涨。
    /// - Returns: (图片, anchorInRTF)：锚点用 blip 起点，供 `calculateImagePosition` 从图在流中的真实位置回溯 `{`，比「blip 前最后一个 \\pict」更不易把相邻两图算反
    private func extractAllImagesFromRTF(_ rtfString: String) -> [(image: NSImage, anchorInRTF: String.Index)] {
        logger.info("[RTF] Starting extraction of all images from RTF (by blip)...")
        
        var results: [(image: NSImage, anchorInRTF: String.Index)] = []
        var searchStart = rtfString.startIndex
        let end = rtfString.endIndex
        
        while searchStart < end {
            let pngR = rtfString.range(of: "\\pngblip", range: searchStart..<end)
            let jpegR = rtfString.range(of: "\\jpegblip", range: searchStart..<end)
            let blipRange: Range<String.Index>?
            let fileHeader: String
            if let p = pngR, let j = jpegR {
                if p.lowerBound < j.lowerBound {
                    blipRange = p
                    fileHeader = "89504e47"
                } else {
                    blipRange = j
                    fileHeader = "ffd8ff"
                }
            } else if let p = pngR {
                blipRange = p
                fileHeader = "89504e47"
            } else if let j = jpegR {
                blipRange = j
                fileHeader = "ffd8ff"
            } else {
                break
            }
            guard let blip = blipRange else { break }
            
            let anchor = blip.lowerBound
            let hexStart = blip.upperBound
            let afterType = rtfString[hexStart..<end]
            guard let headerInSlice = afterType.range(of: fileHeader, options: .caseInsensitive) else {
                logger.warn("[RTF] Could not find image file header after blip #\(results.count + 1)")
                searchStart = blip.upperBound
                continue
            }
            // Substring 与父 String 共用 String.Index，可直接用于 rtfString
            let headerLower = headerInSlice.lowerBound
            
            var hexString = ""
            var idx = headerLower
            while idx < end {
                let char = rtfString[idx]
                if char.isHexDigit {
                    hexString.append(char)
                    idx = rtfString.index(after: idx)
                } else if char.isWhitespace || char.isNewline {
                    // Swift 把 \r\n (CRLF) 当成单个 Character，不匹配 =="\n" 或 =="\r"；
                    // 用 isWhitespace 统一处理所有空白/换行
                    idx = rtfString.index(after: idx)
                } else if char == "}" || char == "\\" {
                    break
                } else {
                    break
                }
            }
            
            searchStart = idx
            
            if hexString.count < 100 {
                logger.warn("[RTF] Image hex after blip too small (\(hexString.count) chars), skipping")
                continue
            }
            guard let imageData = hexString.hexData else {
                logger.error("[RTF] Failed to convert hex to Data for image #\(results.count + 1)")
                continue
            }
            if let image = NSImage(data: imageData) {
                results.append((image: image, anchorInRTF: anchor))
                let kind = fileHeader == "89504e47" ? "png" : "jpeg"
                logger.info("[RTF] Extracted image #\(results.count), type: \(kind), size: \(image.size), data: \(imageData.count) bytes")
            } else {
                logger.error("[RTF] Failed to create NSImage from Data for image #\(results.count + 1)")
            }
        }
        
        logger.info("[RTF] Total images extracted: \(results.count)")
        return results
    }
}

// MARK: - String 十六进制转换扩展
extension String {
    var hexData: Data? {
        var data = Data(capacity: count / 2)
        var buffer = ""
        
        for char in self {
            buffer.append(char)
            if buffer.count == 2 {
                guard let byte = UInt8(buffer, radix: 16) else { return nil }
                data.append(byte)
                buffer = ""
            }
        }
        
        return data
    }
}

