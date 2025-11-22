//
//  CSDeviceManager.swift
//  CrossShare
//
//  Created by TS on 2025/10/15.
//

import Foundation
import Cocoa

/// Data Transmission Manager - Centralized management of file transfer related business logic
class CSDeviceManager {
    static let shared = CSDeviceManager()
    
    // MARK: - Properties
    
    /// 当前设备诊断状态
    private(set) var diasStatus: Int = 0 {
        didSet {
            // 状态改变时触发回调
            onDiasStatusChanged?(diasStatus)
        }
    }
    
    /// 设备诊断状态变化回调
    var onDiasStatusChanged: ((Int) -> Void)?
    
    // MARK: - Initialization
    
    private init() {
        setupNotifications()
    }
    
    // MARK: - Private Methods
    
    private func setupNotifications() {
        // 监听设备诊断状态通知
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(handleDeviceDiasStatusNotification(_:)),
            name: .deviceDiasStatusNotification,
            object: nil
        )
    }
    
    @objc private func handleDeviceDiasStatusNotification(_ notification: Notification) {
        // 处理DIAS状态通知
        if let userInfo = notification.userInfo,
           let status = userInfo["diasStatus"] as? Int {
            logger.info("CSDeviceManager received diasStatus: \(status)")
            DispatchQueue.main.async { [weak self] in
                self?.diasStatus = status
            }
        }
    }
    
    deinit {
        NotificationCenter.default.removeObserver(self)
    }
}
