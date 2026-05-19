//
//  UtilsHelper.swift
//  CrossShare
//
//  Created by user00 on 2025/3/6.
//

import Cocoa
import UniformTypeIdentifiers

struct GoString {
    let p: UnsafePointer<Int8>
    let n: Int
    private let buffer: UnsafeMutableBufferPointer<Int8>?
    
    init(_ string: String) {
        let stringData = string.utf8
        let buffer = UnsafeMutableBufferPointer<Int8>.allocate(capacity: stringData.count + 1)
        // 将 UTF8 数据复制到缓冲区
        var index = 0
        for byte in stringData {
            buffer[index] = Int8(bitPattern: byte)
            index += 1
        }
        buffer[index] = 0 // null terminator
        
        self.p = UnsafePointer(buffer.baseAddress!)
        self.n = stringData.count
        self.buffer = buffer
    }
    
    // 清理内存
    func deallocate() {
        if let buffer = buffer {
            buffer.deallocate()
        }
    }
}

class UtilsHelper: NSObject {
    static func getVersionNumber() -> String {
        if let version = Bundle.main.infoDictionary?["CFBundleShortVersionString"] as? String,
           let build = Bundle.main.infoDictionary?["CFBundleVersion"] as? String {
            logger.info("App marketing ver: \(version)")
            logger.info("App building ver: \(build)")
            return "v\(version)"
        }
        return "1.0.0"
    }
    
    static func checkClipboardFileCategory() {
        let pasteboard = NSPasteboard.general
        if let fileURLs = pasteboard.readObjects(forClasses: [NSURL.self], options: nil) as? [URL] {
            for fileURL in fileURLs {
                if let fileType = UTType(filenameExtension: fileURL.pathExtension) {
                    if fileType.conforms(to: .text) {
                        logger.info("文件: \(fileURL.lastPathComponent) 是文字")
                    } else if fileType.conforms(to: .image) {
                        logger.info("文件: \(fileURL.lastPathComponent) 是图片")
                    } else if fileType.conforms(to: .video) {
                        logger.info("文件: \(fileURL.lastPathComponent) 是视频")
                    } else if fileType.conforms(to: .audio) {
                        logger.info("文件: \(fileURL.lastPathComponent) 是音频")
                    } else if fileType.conforms(to: .pdf) {
                        logger.info("文件: \(fileURL.lastPathComponent) 是 PDF 文档")
                    } else if fileType.conforms(to: UTType.word) || fileType.conforms(to: UTType.wordLegacy) {
                        logger.info("文件: \(fileURL.lastPathComponent) 是 Word 文件")
                    } else if fileType.conforms(to: UTType.powerpoint) || fileType.conforms(to: UTType.powerpointLegacy) {
                        logger.info("文件: \(fileURL.lastPathComponent) 是 PowerPoint 文件")
                    } else if fileType.conforms(to: UTType.excel) || fileType.conforms(to: UTType.excelLegacy) {
                        logger.info("文件: \(fileURL.lastPathComponent) 是 Excel 文件")
                    } else {
                        logger.info("文件: \(fileURL.lastPathComponent) 类型未知")
                    }
                }
            }
        } else {
            logger.info("剪贴板没有文件")
        }
    }
    
    // 文件大小格式化工具方法
    static func formatFileSize(_ bytes: Int64) -> String {
        guard bytes > 0 else { return "0 B" }
        let units = ["B", "KB", "MB", "GB", "TB"]
        let logValue = log2(Double(bytes)) / log2(1024) // Calculate unit level
        let unitIndex = min(Int(logValue), units.count - 1) // Prevent exceeding unit array range
        let size = Double(bytes) / pow(1024, Double(unitIndex))
        return String(format: "%.1f %@", size, units[unitIndex])
    }
    
    // MARK: - File Size Calculation
    
    /// Calculate total size for file items (recursively for directories)
    /// - Parameter items: Array of FileInfo items
    /// - Returns: Total size in bytes
    static func calculateTotalSize(for items: [FileInfo]) -> Int64 {
        var totalSize: Int64 = 0
        let fileManager = FileManager.default
        
        for item in items {
            if item.isDirectory {
                // Recursively calculate directory size
                totalSize += calculateDirectorySize(at: item.path, fileManager: fileManager)
            } else {
                // Use fileSize from FileInfo, or calculate if nil
                if let size = item.fileSize {
                    totalSize += size
                } else {
                    // Fallback: calculate size from file system
                    if let attributes = try? fileManager.attributesOfItem(atPath: item.path),
                       let size = attributes[.size] as? Int64 {
                        totalSize += size
                    }
                }
            }
        }
        
        return totalSize
    }
    
    /// Recursively calculate directory size
    /// - Parameters:
    ///   - path: Directory path
    ///   - fileManager: FileManager instance
    /// - Returns: Total size of all files in the directory in bytes
    static func calculateDirectorySize(at path: String, fileManager: FileManager) -> Int64 {
        var totalSize: Int64 = 0
        
        guard let enumerator = fileManager.enumerator(atPath: path) else {
            return 0
        }
        
        for case let file as String in enumerator {
            let filePath = (path as NSString).appendingPathComponent(file)
            
            // Check if it's a directory
            var isDirectory: ObjCBool = false
            guard fileManager.fileExists(atPath: filePath, isDirectory: &isDirectory) else {
                continue
            }
            
            // Only count files, not directories
            if !isDirectory.boolValue {
                if let attributes = try? fileManager.attributesOfItem(atPath: filePath),
                   let size = attributes[.size] as? Int64 {
                    totalSize += size
                }
            }
        }
        
        return totalSize
    }

}
