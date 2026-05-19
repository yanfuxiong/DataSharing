//
//  DebugUtils.swift
//  CrossShareHelper
//
//  调试工具类 - 用于存放调试相关的方法
//  后续删除调试代码时，直接删除此文件即可
//

import Foundation
import AppKit

/// 调试工具单例类
/// 所有调试相关的方法都放在这里，方便统一管理和删除
class DebugUtils {
    
    static let shared = DebugUtils()
    private let logger = CSLogger.shared
    
    private init() {}
    
    // MARK: - 基础文件写入方法
    
    /// 将RTF数据写入桌面文件
    func writeRtfToDesktop(_ rtfData: String?, fileName: String) {
        guard let rtfDataStr = rtfData, !rtfDataStr.isEmpty else { return }
        
        if let desktopURL = FileManager.default.urls(for: .desktopDirectory, in: .userDomainMask).first {
            let rtfURL = desktopURL.appendingPathComponent(fileName)
            do {
                try rtfDataStr.write(to: rtfURL, atomically: true, encoding: .utf8)
                logger.info("[DebugUtils] Written to desktop/\(fileName), size: \(rtfDataStr.count) bytes")
            } catch {
                logger.error("[DebugUtils] Failed to write \(fileName): \(error)")
            }
        }
    }
    
    /// 将二进制数据写入桌面文件
    func writeDataToDesktop(_ data: Data?, fileName: String) {
        guard let data = data, !data.isEmpty else { return }
        
        if let desktopURL = FileManager.default.urls(for: .desktopDirectory, in: .userDomainMask).first {
            let fileURL = desktopURL.appendingPathComponent(fileName)
            do {
                try data.write(to: fileURL)
                logger.info("[DebugUtils] Written to desktop/\(fileName), size: \(data.count) bytes")
            } catch {
                logger.error("[DebugUtils] Failed to write \(fileName): \(error)")
            }
        }
    }
}
