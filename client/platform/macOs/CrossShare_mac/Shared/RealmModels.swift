//
//  RealmModels.swift
//  CrossShare
//
//  Created by TS on 2025/9/29.
//

import Foundation
import RealmSwift

/// Realm版本号
enum RealmSchemaVersion {
    static let currentVersion: UInt64 = 2  // 添加 errCode 字段，版本升级到 2
}

/// 文件传输信息模型（Realm版本）
class RealmCSFileInfo: Object {
    @Persisted(primaryKey: true) var sessionId: String
    @Persisted var senderID: String
    @Persisted var isCompleted: Bool
    @Persisted var progress: Double
    @Persisted var errCode: Int?
    @Persisted var session: RealmFileTransfer?
    @Persisted var createdAt: Date = Date()
    @Persisted var updatedAt: Date = Date()
    
    // 从CSFileInfo创建RealmCSFileInfo
    convenience init(from csFileInfo: CSFileInfo) {
        self.init()
        self.sessionId = csFileInfo.sessionId
        self.senderID = csFileInfo.senderID
        self.isCompleted = csFileInfo.isCompleted
        self.progress = csFileInfo.progress
        self.errCode = csFileInfo.errCode
        self.session = RealmFileTransfer(from: csFileInfo.session)
    }
    
    // 转换回CSFileInfo
    func toCSFileInfo() -> CSFileInfo {
        guard let session = session else {
            fatalError("Session cannot be nil")
        }
        return CSFileInfo(
            session: session.toFileTransfer(),
            sessionId: sessionId,
            senderID: senderID,
            isCompleted: isCompleted,
            progress: progress,
            errCode: errCode
        )
    }
}

/// 文件传输会话模型（Realm版本）
class RealmFileTransfer: Object {
    @Persisted var sessionId: String
    @Persisted var senderIP: String
    @Persisted var senderID: String
    @Persisted var deviceName: String
    @Persisted var direction: String // 使用字符串存储枚举值
    @Persisted var status: String // 使用字符串存储枚举值
    @Persisted var startTime: Date
    @Persisted var endTime: Date?
    @Persisted var lastUpdateTime: Date

    @Persisted var totalFileCount: Int32
    @Persisted var receivedFileCount: Int32
    @Persisted var totalSize: Int64
    @Persisted var receivedSize: Int64

    @Persisted var currentFileName: String
    @Persisted var currentFileSize: Int64

    @Persisted var error: String?
    
    // 从FileTransfer创建RealmFileTransfer
    convenience init(from fileTransfer: FileTransfer) {
        self.init()
        self.sessionId = fileTransfer.sessionId
        self.senderIP = fileTransfer.senderIP
        self.senderID = fileTransfer.senderID
        self.deviceName = fileTransfer.deviceName
        self.direction = fileTransfer.direction.rawValue
        self.status = fileTransfer.status.rawValue
        self.startTime = fileTransfer.startTime
        self.endTime = fileTransfer.endTime
        self.lastUpdateTime = fileTransfer.lastUpdateTime
        self.totalFileCount = Int32(fileTransfer.totalFileCount)
        self.receivedFileCount = Int32(fileTransfer.receivedFileCount)
        self.totalSize = Int64(fileTransfer.totalSize)
        self.receivedSize = Int64(fileTransfer.receivedSize)
        self.currentFileName = fileTransfer.currentFileName
        self.currentFileSize = Int64(fileTransfer.currentFileSize)
        self.error = fileTransfer.error
    }
    
    // 转换回FileTransfer
    func toFileTransfer() -> FileTransfer {
        // 安全地转换枚举值
        guard let direction = FileTransfer.TransferDirection(rawValue: self.direction),
              let status = FileTransfer.TransferStatus(rawValue: self.status) else {
            fatalError("Invalid enum values")
        }
        
        // 创建一个可变的FileTransfer实例
        let fileTransfer = FileTransfer(
            sessionId: sessionId,
            senderIP: senderIP,
            senderID: senderID,
            deviceName: deviceName,
            direction: direction,
            status: status,
            startTime: startTime,
            endTime: endTime,
            lastUpdateTime: lastUpdateTime,
            totalFileCount: UInt32(totalFileCount),
            receivedFileCount: UInt32(receivedFileCount),
            totalSize: UInt64(totalSize),
            receivedSize: UInt64(receivedSize),
            currentFileName: currentFileName,
            currentFileSize: UInt64(currentFileSize),
            error: error
        )
        
        return fileTransfer
    }
}

// 扩展FileTransfer，添加一个初始化方法来支持从RealmFileTransfer创建
private extension FileTransfer {
    // 添加一个私有初始化方法来支持从RealmFileTransfer创建
    init(
        sessionId: String,
        senderIP: String,
        senderID: String,
        deviceName: String,
        direction: TransferDirection,
        status: TransferStatus,
        startTime: Date,
        endTime: Date?,
        lastUpdateTime: Date,
        totalFileCount: UInt32,
        receivedFileCount: UInt32,
        totalSize: UInt64,
        receivedSize: UInt64,
        currentFileName: String,
        currentFileSize: UInt64,
        error: String?
    ) {
        self.sessionId = sessionId
        self.senderIP = senderIP
        self.senderID = senderID
        self.deviceName = deviceName
        self.direction = direction
        self.status = status
        self.startTime = startTime
        self.endTime = endTime
        self.lastUpdateTime = lastUpdateTime
        self.totalFileCount = totalFileCount
        self.receivedFileCount = receivedFileCount
        self.totalSize = totalSize
        self.receivedSize = receivedSize
        self.currentFileName = currentFileName
        self.currentFileSize = currentFileSize
        self.error = error
    }
}