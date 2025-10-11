//
//  CSFileInfo.swift
//  CrossShare
//
//  Created by TS on 2025/9/29.
//

import Foundation

/// 文件传输信息模型
struct CSFileInfo {
    let session: FileTransfer
    let sessionId: String
    let senderID: String
    let isCompleted: Bool
    let progress: Double
}

/// 文件传输会话模型
struct FileTransfer {
    let sessionId: String
    let senderIP: String
    let senderID: String
    let deviceName: String
    let direction: TransferDirection
    var status: TransferStatus
    let startTime: Date
    var endTime: Date?
    var lastUpdateTime: Date

    let totalFileCount: UInt32
    var receivedFileCount: UInt32
    let totalSize: UInt64
    var receivedSize: UInt64

    var currentFileName: String
    var currentFileSize: UInt64

    var error: String?

    /// 传输方向枚举
    enum TransferDirection: String, CaseIterable {
        case send = "send"
        case receive = "receive"
    }

    /// 传输状态枚举
    enum TransferStatus: String, CaseIterable {
        case pending = "pending"       // 等待中
        case inProgress = "inProgress" // 传输中
        case completed = "completed"   // 已完成
        case failed = "failed"         // 失败
        case cancelled = "cancelled"   // 已取消
        case paused = "paused"         // 已暂停
    }

    /// 计算总进度
    var totalProgress: Double {
        guard totalSize > 0 else { return 0.0 }
        return Double(receivedSize) / Double(totalSize)
    }

    /// 检查是否已完成
    var isCompleted: Bool {
        return status == .completed && receivedFileCount >= totalFileCount && receivedSize >= totalSize
    }

    /// 检查是否为多文件传输
    var isMultipleFiles: Bool {
        return totalFileCount > 1
    }

    /// 计算估计速度
    var estimatedSpeed: UInt64 {
        let timeElapsed = Date().timeIntervalSince(startTime)
        guard timeElapsed > 0 else { return 0 }
        return UInt64(Double(receivedSize) / timeElapsed)
    }

    /// 计算估计剩余时间
    var estimatedTimeRemaining: TimeInterval? {
        let speed = estimatedSpeed
        guard speed > 0 else { return nil }
        let remainingBytes = totalSize - receivedSize
        return Double(remainingBytes) / Double(speed)
    }

    /// 格式化总大小
    var formattedTotalSize: String {
        return ByteCountFormatter.string(fromByteCount: Int64(totalSize), countStyle: .file)
    }

    /// 格式化已接收大小
    var formattedReceivedSize: String {
        return ByteCountFormatter.string(fromByteCount: Int64(receivedSize), countStyle: .file)
    }

    /// 格式化速度
    var formattedSpeed: String {
        return ByteCountFormatter.string(fromByteCount: Int64(estimatedSpeed), countStyle: .file) + "/s"
    }

    /// 从字典创建FileTransferSession实例
    init?(from dict: [String: Any]) {
        guard let sessionId = dict["sessionId"] as? String,
              let senderIP = dict["senderIP"] as? String,
              let senderID = dict["senderID"] as? String,
              let deviceName = dict["deviceName"] as? String,
              let directionString = dict["direction"] as? String,
              let direction = TransferDirection(rawValue: directionString),
              let statusString = dict["status"] as? String,
              let status = TransferStatus(rawValue: statusString),
              let totalFileCount = dict["totalFileCount"] as? UInt32,
              let totalSize = dict["totalSize"] as? UInt64,
              let currentFileName = dict["currentFileName"] as? String,
              let currentFileSize = dict["currentFileSize"] as? UInt64
        else {
            print("缺少创建FileTransferSession所需的数据")
            return nil
        }
        
        self.sessionId = sessionId
        self.senderIP = senderIP
        self.senderID = senderID
        self.deviceName = deviceName
        self.direction = direction
        self.status = status
        
        // 处理时间
        if let startTimeInterval = dict["startTime"] as? Double {
            self.startTime = Date(timeIntervalSince1970: startTimeInterval)
        } else {
            self.startTime = Date()
        }
        
        if let endTimeInterval = dict["endTime"] as? Double {
            self.endTime = Date(timeIntervalSince1970: endTimeInterval)
        } else {
            self.endTime = nil
        }
        
        if let lastUpdateTimeInterval = dict["lastUpdateTime"] as? Double {
            self.lastUpdateTime = Date(timeIntervalSince1970: lastUpdateTimeInterval)
        } else {
            self.lastUpdateTime = Date()
        }
        
        self.totalFileCount = totalFileCount
        self.receivedFileCount = dict["receivedFileCount"] as? UInt32 ?? 0
        self.totalSize = totalSize
        self.receivedSize = dict["receivedSize"] as? UInt64 ?? 0
        
        self.currentFileName = currentFileName
        self.currentFileSize = currentFileSize
        self.error = dict["error"] as? String
    }

    /// 将FileTransferSession转换为字典
    func toDictionary() -> [String: Any] {
        var dict: [String: Any] = [
            "sessionId": sessionId,
            "senderIP": senderIP,
            "senderID": senderID,
            "deviceName": deviceName,
            "direction": direction.rawValue,
            "status": status.rawValue,
            "startTime": startTime.timeIntervalSince1970,
            "lastUpdateTime": lastUpdateTime.timeIntervalSince1970,
            "totalFileCount": totalFileCount,
            "receivedFileCount": receivedFileCount,
            "totalSize": totalSize,
            "receivedSize": receivedSize,
            "currentFileName": currentFileName,
            "currentFileSize": currentFileSize,
            "totalProgress": totalProgress,
            "isCompleted": isCompleted,
            "isMultipleFiles": isMultipleFiles
        ]
        
        if let endTime = endTime {
            dict["endTime"] = endTime.timeIntervalSince1970
        }
        
        if let error = error {
            dict["error"] = error
        }
        
        return dict
    }

    /// 更新进度
    mutating func updateProgress(from dict: [String: Any]) {
        if let receivedFileCount = dict["receivedFileCount"] as? UInt32 {
            self.receivedFileCount = receivedFileCount
        }
        
        if let receivedSize = dict["receivedSize"] as? UInt64 {
            self.receivedSize = receivedSize
        }
        
        if let currentFileName = dict["currentFileName"] as? String {
            self.currentFileName = currentFileName
        }
        
        if let currentFileSize = dict["currentFileSize"] as? UInt64 {
            self.currentFileSize = currentFileSize
        }
        
        self.lastUpdateTime = Date()
        
        if let isCompleted = dict["isCompleted"] as? Bool, isCompleted {
            self.status = .completed
            self.endTime = Date()
        } else if self.status == .pending {
            self.status = .inProgress
        }
    }

    /// 设置错误信息
    mutating func setError(_ error: String) {
        self.error = error
        self.status = .failed
        self.endTime = Date()
    }

    /// 取消传输
    mutating func cancel() {
        self.status = .cancelled
        self.endTime = Date()
    }
}
