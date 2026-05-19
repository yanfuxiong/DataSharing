//
//  RTFBuilder.swift
//  CrossShareHelper
//
//  RTF 构建工具类 - 从 RTFD 数据构建带图片的 RTF
//

import Foundation
import AppKit

/// RTF 构建器单例类
/// 用于从 RTFD 数据提取并构建带图片的 RTF 字符串
class RTFBuilder {
    
    static let shared = RTFBuilder()
    private let logger = CSLogger.shared
    
    private init() {}
    
    // MARK: - 十六进制查表（预生成 ASCII 字节，性能优化）
    
    private static let hexTable: [UInt8] = Array("0123456789abcdef".utf8)
    
    // MARK: - 公开方法
    
    /// 从 RTFD 数据构建带图片的 RTF
    /// - Parameter rtfdData: RTFD 格式的二进制数据
    /// - Returns: (rtf: 带图片的RTF字符串, image: 提取的第一张图片, text: 纯文本)
    func buildRTFFromRTFD(_ rtfdData: Data) -> (rtf: String?, image: NSImage?, text: String?) {
        // 解析 RTFD 为 NSAttributedString
        guard let attributedString = try? NSAttributedString(
            data: rtfdData,
            options: [.documentType: NSAttributedString.DocumentType.rtfd],
            documentAttributes: nil
        ) else {
            logger.warn("[RTFBuilder] Failed to parse RTFD data")
            return (nil, nil, nil)
        }
        
        logger.info("[RTFBuilder] Parsed NSAttributedString, length: \(attributedString.length)")
        
        // 提取图片信息
        var imageDataList: [(position: Int, data: Data, size: NSSize)] = []
        var firstImage: NSImage? = nil
        
        attributedString.enumerateAttribute(.attachment, in: NSRange(location: 0, length: attributedString.length)) { value, range, _ in
            if let attachment = value as? NSTextAttachment {
                var imageData: Data? = nil
                var imageSize = NSSize.zero
                
                if let fileWrapper = attachment.fileWrapper,
                   let data = fileWrapper.regularFileContents {
                    imageData = data
                    if let image = NSImage(data: data) {
                        imageSize = image.size
                        if firstImage == nil {
                            firstImage = image
                        }
                    }
                } else if let image = attachment.image {
                    imageSize = image.size
                    if firstImage == nil {
                        firstImage = image
                    }
                    if let tiffData = image.tiffRepresentation,
                       let bitmapRep = NSBitmapImageRep(data: tiffData),
                       let pngData = bitmapRep.representation(using: .png, properties: [:]) {
                        imageData = pngData
                    }
                }
                
                if let data = imageData {
                    imageDataList.append((position: range.location, data: data, size: imageSize))
                    logger.info("[RTFBuilder] Found image at position \(range.location), size: \(imageSize), data: \(data.count) bytes")
                }
            }
        }
        
        logger.info("[RTFBuilder] Total images: \(imageDataList.count)")
        
        // 如果没有图片，返回普通 RTF
        if imageDataList.isEmpty {
            if let rtfData = try? attributedString.data(
                from: NSRange(location: 0, length: attributedString.length),
                documentAttributes: [.documentType: NSAttributedString.DocumentType.rtf]
            ), let rtfString = String(data: rtfData, encoding: .utf8) {
                return (rtfString, nil, attributedString.string)
            }
            return (nil, nil, attributedString.string)
        }
        
        // 构建带图片的 RTF
        let rtf = buildRTFWithImages(attributedString: attributedString, imageDataList: imageDataList)
        return (rtf, firstImage, attributedString.string)
    }
    
    // MARK: - 内部方法
    
    /// 从 NSAttributedString 构建包含图片的完整 RTF（保留字体、大小、颜色、粗体、斜体等格式）
    private func buildRTFWithImages(attributedString: NSAttributedString, imageDataList: [(position: Int, data: Data, size: NSSize)]) -> String {
        var imageDict: [Int: (data: Data, size: NSSize)] = [:]
        for img in imageDataList {
            imageDict[img.position] = (data: img.data, size: img.size)
        }
        
        let fullRange = NSRange(location: 0, length: attributedString.length)
        let nsString = attributedString.string as NSString
        
        // 第一遍：收集所有使用的字体和颜色
        var fontNames: [String] = []
        var fontNameToIndex: [String: Int] = [:]
        var colorEntries: [(Int, Int, Int)] = []
        var colorKeyToIndex: [String: Int] = [:]
        
        attributedString.enumerateAttributes(in: fullRange, options: []) { attrs, _, _ in
            if let font = attrs[.font] as? NSFont {
                let name = font.familyName ?? font.fontName
                if fontNameToIndex[name] == nil {
                    fontNameToIndex[name] = fontNames.count
                    fontNames.append(name)
                }
            }
            if let color = attrs[.foregroundColor] as? NSColor,
               let rgbColor = color.usingColorSpace(.sRGB) {
                let r = Int(round(rgbColor.redComponent * 255))
                let g = Int(round(rgbColor.greenComponent * 255))
                let b = Int(round(rgbColor.blueComponent * 255))
                let key = "\(r),\(g),\(b)"
                if colorKeyToIndex[key] == nil {
                    colorKeyToIndex[key] = colorEntries.count + 1
                    colorEntries.append((r, g, b))
                }
            }
        }
        
        if fontNames.isEmpty {
            fontNames.append("Helvetica")
            fontNameToIndex["Helvetica"] = 0
        }
        
        var parts: [String] = []
        parts.reserveCapacity(attributedString.length * 2 + imageDataList.count + 20)
        
        // RTF 头部
        parts.append("{\\rtf1\\ansi\\ansicpg936\\deff0\\uc0\n")
        
        // 字体表（包含所有实际使用的字体）
        parts.append("{\\fonttbl")
        for (index, name) in fontNames.enumerated() {
            parts.append("{\\f\(index)\\fnil\\fcharset0 \(name);}")
        }
        parts.append("}\n")
        
        // 颜色表（索引 0 = auto/default，从索引 1 开始存储实际颜色）
        parts.append("{\\colortbl;")
        for (r, g, b) in colorEntries {
            parts.append("\\red\(r)\\green\(g)\\blue\(b);")
        }
        parts.append("}\n")
        
        parts.append("\\pard")
        
        // 格式状态追踪，避免重复输出相同控制词
        var curFontIdx = -1
        var curFontSize = -1
        var curBold = false
        var curItalic = false
        var curUnderline = false
        var curStrike = false
        var curColorIdx = 0
        
        // 第二遍：按属性 run 遍历，保留每段文字的格式
        attributedString.enumerateAttributes(in: fullRange, options: []) { attrs, range, _ in
            var fontIdx = 0
            var fontSize = 24
            var isBold = false
            var isItalic = false
            
            if let font = attrs[.font] as? NSFont {
                let name = font.familyName ?? font.fontName
                fontIdx = fontNameToIndex[name] ?? 0
                fontSize = Int(round(font.pointSize * 2))
                
                let traits = NSFontManager.shared.traits(of: font)
                isBold = traits.contains(.boldFontMask)
                isItalic = traits.contains(.italicFontMask)
            }
            
            var colorIdx = 0
            if let color = attrs[.foregroundColor] as? NSColor,
               let rgbColor = color.usingColorSpace(.sRGB) {
                let r = Int(round(rgbColor.redComponent * 255))
                let g = Int(round(rgbColor.greenComponent * 255))
                let b = Int(round(rgbColor.blueComponent * 255))
                let key = "\(r),\(g),\(b)"
                colorIdx = colorKeyToIndex[key] ?? 0
            }
            
            let isUnderline = (attrs[.underlineStyle] as? Int ?? 0) != 0
            let isStrike = (attrs[.strikethroughStyle] as? Int ?? 0) != 0
            
            // 生成格式变更控制词
            var fmt = ""
            if fontIdx != curFontIdx { fmt += "\\f\(fontIdx)"; curFontIdx = fontIdx }
            if fontSize != curFontSize { fmt += "\\fs\(fontSize)"; curFontSize = fontSize }
            if isBold != curBold { fmt += isBold ? "\\b" : "\\b0"; curBold = isBold }
            if isItalic != curItalic { fmt += isItalic ? "\\i" : "\\i0"; curItalic = isItalic }
            if isUnderline != curUnderline { fmt += isUnderline ? "\\ul" : "\\ulnone"; curUnderline = isUnderline }
            if isStrike != curStrike { fmt += isStrike ? "\\strike" : "\\strike0"; curStrike = isStrike }
            if colorIdx != curColorIdx { fmt += "\\cf\(colorIdx)"; curColorIdx = colorIdx }
            
            if !fmt.isEmpty {
                parts.append("\(fmt) ")
            }
            
            // 逐字符处理此 run 的内容
            for i in range.location..<NSMaxRange(range) {
                if let imageInfo = imageDict[i] {
                    let pictBlock = createRTFPictBlock(imageData: imageInfo.data, size: imageInfo.size)
                    parts.append(pictBlock)
                    continue
                }
                
                let ch = nsString.character(at: i)
                
                if UTF16.isTrailSurrogate(ch) { continue }
                if ch == 0xFFFC { continue }
                
                var codePoint: UInt32
                if UTF16.isLeadSurrogate(ch) && i + 1 < NSMaxRange(range) {
                    let trail = nsString.character(at: i + 1)
                    if UTF16.isTrailSurrogate(trail) {
                        codePoint = 0x10000 + UInt32(ch - 0xD800) * 0x400 + UInt32(trail - 0xDC00)
                    } else {
                        codePoint = UInt32(ch)
                    }
                } else {
                    codePoint = UInt32(ch)
                }
                
                if codePoint == 0x0A {
                    parts.append("\\par\n")
                } else if codePoint == 0x09 {
                    parts.append("\\tab ")
                } else if codePoint == 0x5C {
                    parts.append("\\\\")
                } else if codePoint == 0x7B {
                    parts.append("\\{")
                } else if codePoint == 0x7D {
                    parts.append("\\}")
                } else if codePoint < 128 {
                    parts.append(String(Character(Unicode.Scalar(codePoint)!)))
                } else if codePoint <= 0xFFFF {
                    let signedVal = codePoint > 32767 ? Int(codePoint) - 65536 : Int(codePoint)
                    parts.append("\\u\(signedVal) ")
                } else {
                    let hi = 0xD800 + (codePoint - 0x10000) / 0x400
                    let lo = 0xDC00 + (codePoint - 0x10000) % 0x400
                    parts.append("\\u\(Int(hi) - 65536)\\u\(Int(lo) - 65536) ")
                }
            }
        }
        
        parts.append("}")
        
        let rtf = parts.joined()
        logger.info("[RTFBuilder] Built RTF with formatting, length: \(rtf.count)")
        return rtf
    }
    
    /// 创建 RTF \pict 块
    private func createRTFPictBlock(imageData: Data, size: NSSize) -> String {
        // 检测图片格式
        var format = "pngblip"
        if imageData.count >= 3 {
            let header = [UInt8](imageData.prefix(3))
            if header[0] == 0xFF && header[1] == 0xD8 && header[2] == 0xFF {
                format = "jpegblip"
            }
        }
        
        // 获取像素尺寸
        var pixelsWide = Int(size.width)
        var pixelsHigh = Int(size.height)
        
        // 尝试从 PNG 头部读取实际像素尺寸
        if format == "pngblip" && imageData.count >= 24 {
            let width = (Int(imageData[16]) << 24) | (Int(imageData[17]) << 16) | (Int(imageData[18]) << 8) | Int(imageData[19])
            let height = (Int(imageData[20]) << 24) | (Int(imageData[21]) << 16) | (Int(imageData[22]) << 8) | Int(imageData[23])
            if width > 0 && height > 0 {
                pixelsWide = width
                pixelsHigh = height
            }
        }
        
        // 转换为 twips (1 point = 20 twips)
        let widthTwips = Int(size.width * 20)
        let heightTwips = Int(size.height * 20)
        
        // 转换为十六进制
        let hexString = dataToHexString(imageData)
        
        return "{\\pict\\\(format)\\picw\(pixelsWide)\\pich\(pixelsHigh)\\picwgoal\(widthTwips)\\pichgoal\(heightTwips) \(hexString)}"
    }
    
    /// 快速将 Data 转换为十六进制字符串（优化版：使用 UInt8 数组）
    private func dataToHexString(_ data: Data) -> String {
        let count = data.count
        var hexBytes = [UInt8](repeating: 0, count: count * 2)
        
        data.withUnsafeBytes { (bytes: UnsafeRawBufferPointer) in
            let srcPtr = bytes.bindMemory(to: UInt8.self)
            for i in 0..<count {
                let byte = srcPtr[i]
                hexBytes[i * 2] = RTFBuilder.hexTable[Int(byte >> 4)]
                hexBytes[i * 2 + 1] = RTFBuilder.hexTable[Int(byte & 0x0F)]
            }
        }
        
        return String(bytes: hexBytes, encoding: .ascii) ?? ""
    }
}
