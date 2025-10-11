//
//  SharedModels.swift
//  CrossShareHelper
//
//  Created by TS on 2025/9/11.
//  共享的数据模型定义
//

import Foundation

struct CrossShareDevice {
    let timeStamp: Int64
    let id: String
    let ipAddr: String
    var status: Int = 0
    var platform: String = ""
    var deviceName: String = ""
    var sourcePortType: String = ""
    var version: String = ""
    
    enum CodingKeys: String, CodingKey {
        case timeStamp = "TimeStamp"
        case status = "Status"
        case id = "ID"
        case ipAddr = "IpAddr"
        case platform = "Platform"
        case deviceName = "DeviceName"
        case sourcePortType = "SourcePortType"
        case version = "Version"
    }
    
    init(timeStamp: Int64, id: String, ipAddr: String, status: Int = 0, platform: String = "", deviceName: String = "", sourcePortType: String = "", version: String = "") {
        self.timeStamp = timeStamp
        self.id = id
        self.ipAddr = ipAddr
        self.status = status
        self.platform = platform
        self.deviceName = deviceName
        self.sourcePortType = sourcePortType
        self.version = version
    }
    
    init?(from dictionary: [String: Any]) {
        guard let timeStamp = dictionary["TimeStamp"] as? Int64,
              let id = dictionary["ID"] as? String,
              let ipAddr = dictionary["IpAddr"] as? String,
              let platform = dictionary["Platform"] as? String,
              let deviceName = dictionary["DeviceName"] as? String,
              let sourcePortType = dictionary["SourcePortType"] as? String,
              let version = dictionary["Version"] as? String,
              let status = dictionary["Status"] as? Int else {
                  return nil
              }
        
        self.timeStamp = timeStamp
        self.id = id
        self.ipAddr = ipAddr
        self.status = status
        self.platform = platform
        self.deviceName = deviceName
        self.sourcePortType = sourcePortType
        self.version = version
    }
    
    func toDictionary() -> [String: Any] {
        return [
            "TimeStamp": timeStamp,
            "ID": id,
            "IpAddr": ipAddr,
            "Status": status,
            "Platform": platform,
            "DeviceName": deviceName,
            "SourcePortType": sourcePortType,
            "Version": version
        ]
    }
}

enum DevicePlatform: String {
    case android = "android"
    case ios = "ios"
    case macos = "macos"
    case windows = "windows"
    case unknown = "unknown"
    
    init(from string: String) {
        self = DevicePlatform(rawValue: string.lowercased()) ?? .unknown
    }
}

enum SourcePortType: String {
    case SOURCE_HDMI1 = "HDMI1";
    case SOURCE_HDMI2 = "HDMI2";
    case SOURCE_USBC1 = "USBC1";
    case SOURCE_USBC2 = "USBC2";
    case SOURCE_DP1 = "DP1";
    case SOURCE_DP2 = "DP2";
    case SOURCE_MIRACAST = "Miracast";
    case unknown = "Unknown"
    
    init(from string: String) {
        self = SourcePortType(rawValue: string) ?? .unknown
    }
}

extension CrossShareDevice {
    var platformType: DevicePlatform {
        return DevicePlatform(from: platform)
    }
    
    var portType: SourcePortType {
        return SourcePortType(from: sourcePortType)
    }
    
    var displayName: String {
        return deviceName.isEmpty ? "Unknown Device" : deviceName
    }
    
    var ipAddress: String? {
        let components = ipAddr.split(separator: ":")
        return components.first.map(String.init)
    }
    
    var port: Int? {
        let components = ipAddr.split(separator: ":")
        guard components.count == 2,
              let portStr = components.last,
              let port = Int(portStr) else {
            return nil
        }
        return port
    }
}

extension CrossShareDevice {
    var date: Date {
        return Date(timeIntervalSince1970: Double(timeStamp) / 1000.0)
    }
    
    var formattedDate: String {
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy-MM-dd HH:mm:ss"
        return formatter.string(from: date)
    }
}

struct MultipleFileTransferProgress {
    let senderIP: String           // 发送端IP地址(包含port)
    let senderID: String           // 发送端ID
    let deviceName: String         // 发送端设备名称
    let currentFileName: String    // 当前正在接收的文件名
    let receivedFileCount: UInt32  // 当前已经接收完毕的文件个数
    let totalFileCount: UInt32     // 总的文件数量
    let currentFileSize: UInt64    // 当前正在接收的文件size
    let totalSize: UInt64          // 当前一批多文件的所有文档的总的size
    let receivedSize: UInt64       // 当前已经接收的总size
    let timestamp: UInt64          // 时间戳

    var totalProgress: Double {
        guard totalSize > 0 else { return 0.0 }
        return Double(receivedSize) / Double(totalSize)
    }

    var currentFileProgress: Double {
        guard currentFileSize > 0 else { return 0.0 }
        return totalProgress
    }

    var isCompleted: Bool {
        return receivedFileCount >= totalFileCount && receivedSize >= totalSize
    }

    var formattedTotalSize: String {
        return ByteCountFormatter.string(fromByteCount: Int64(totalSize), countStyle: .file)
    }

    var formattedReceivedSize: String {
        return ByteCountFormatter.string(fromByteCount: Int64(receivedSize), countStyle: .file)
    }

    var formattedCurrentFileSize: String {
        return ByteCountFormatter.string(fromByteCount: Int64(currentFileSize), countStyle: .file)
    }

    init(senderIP: String, senderID: String, deviceName: String, currentFileName: String,
         receivedFileCount: UInt32, totalFileCount: UInt32, currentFileSize: UInt64,
         totalSize: UInt64, receivedSize: UInt64, timestamp: UInt64) {
        self.senderIP = senderIP
        self.senderID = senderID
        self.deviceName = deviceName
        self.currentFileName = currentFileName
        self.receivedFileCount = receivedFileCount
        self.totalFileCount = totalFileCount
        self.currentFileSize = currentFileSize
        self.totalSize = totalSize
        self.receivedSize = receivedSize
        self.timestamp = timestamp
    }

    func toDictionary() -> [String: Any] {
        return [
            "senderIP": senderIP,
            "senderID": senderID,
            "deviceName": deviceName,
            "currentFileName": currentFileName,
            "receivedFileCount": receivedFileCount,
            "totalFileCount": totalFileCount,
            "currentFileSize": currentFileSize,
            "totalSize": totalSize,
            "receivedSize": receivedSize,
            "timestamp": timestamp,
            "totalProgress": totalProgress,
            "isCompleted": isCompleted
        ]
    }
}

struct FileTransferSession {
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

    enum TransferDirection: String, CaseIterable {
        case send = "send"
        case receive = "receive"
    }

    enum TransferStatus: String, CaseIterable {
        case pending = "pending"       // 等待中
        case inProgress = "inProgress" // 传输中
        case completed = "completed"   // 已完成
        case failed = "failed"         // 失败
        case cancelled = "cancelled"   // 已取消
        case paused = "paused"         // 已暂停
    }

    var totalProgress: Double {
        guard totalSize > 0 else { return 0.0 }
        return Double(receivedSize) / Double(totalSize)
    }

    var isCompleted: Bool {
        return status == .completed && receivedFileCount >= totalFileCount && receivedSize >= totalSize
    }

    var isMultipleFiles: Bool {
        return totalFileCount > 1
    }

    var estimatedSpeed: UInt64 {
        let timeElapsed = Date().timeIntervalSince(startTime)
        guard timeElapsed > 0 else { return 0 }
        return UInt64(Double(receivedSize) / timeElapsed)
    }

    var estimatedTimeRemaining: TimeInterval? {
        let speed = estimatedSpeed
        guard speed > 0 else { return nil }
        let remainingBytes = totalSize - receivedSize
        return Double(remainingBytes) / Double(speed)
    }

    var formattedTotalSize: String {
        return ByteCountFormatter.string(fromByteCount: Int64(totalSize), countStyle: .file)
    }

    var formattedReceivedSize: String {
        return ByteCountFormatter.string(fromByteCount: Int64(receivedSize), countStyle: .file)
    }

    var formattedSpeed: String {
        return ByteCountFormatter.string(fromByteCount: Int64(estimatedSpeed), countStyle: .file) + "/s"
    }

    init(from progress: MultipleFileTransferProgress) {
        self.sessionId = "\(progress.senderID)-\(progress.timestamp)"
        self.senderIP = progress.senderIP
        self.senderID = progress.senderID
        self.deviceName = progress.deviceName
        self.direction = .receive
        self.status = progress.isCompleted ? .completed : .inProgress
        self.startTime = Date(timeIntervalSince1970: Double(progress.timestamp) / 1000.0)
        self.endTime = progress.isCompleted ? Date() : nil
        self.lastUpdateTime = Date()

        self.totalFileCount = progress.totalFileCount
        self.receivedFileCount = progress.receivedFileCount
        self.totalSize = progress.totalSize
        self.receivedSize = progress.receivedSize

        self.currentFileName = progress.currentFileName
        self.currentFileSize = progress.currentFileSize

        self.error = nil
    }

    init(sessionId: String, senderIP: String, senderID: String, deviceName: String,
         direction: TransferDirection, totalFileCount: UInt32, totalSize: UInt64,
         currentFileName: String, currentFileSize: UInt64) {
        self.sessionId = sessionId
        self.senderIP = senderIP
        self.senderID = senderID
        self.deviceName = deviceName
        self.direction = direction
        self.status = .pending
        self.startTime = Date()
        self.endTime = nil
        self.lastUpdateTime = Date()

        self.totalFileCount = totalFileCount
        self.receivedFileCount = 0
        self.totalSize = totalSize
        self.receivedSize = 0

        self.currentFileName = currentFileName
        self.currentFileSize = currentFileSize

        self.error = nil
    }

    mutating func updateProgress(from progress: MultipleFileTransferProgress) {
        self.receivedFileCount = progress.receivedFileCount
        self.receivedSize = progress.receivedSize
        self.currentFileName = progress.currentFileName
        self.currentFileSize = progress.currentFileSize
        self.lastUpdateTime = Date()

        if progress.isCompleted {
            self.status = .completed
            self.endTime = Date()
        } else if self.status == .pending {
            self.status = .inProgress
        }
    }

    mutating func setError(_ error: String) {
        self.error = error
        self.status = .failed
        self.endTime = Date()
    }

    mutating func cancel() {
        self.status = .cancelled
        self.endTime = Date()
    }

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
            "isMultipleFiles": isMultipleFiles,
            "estimatedSpeed": estimatedSpeed,
            "formattedTotalSize": formattedTotalSize,
            "formattedReceivedSize": formattedReceivedSize,
            "formattedSpeed": formattedSpeed
        ]

        if let endTime = endTime {
            dict["endTime"] = endTime.timeIntervalSince1970
        }

        if let error = error {
            dict["error"] = error
        }

        if let estimatedTimeRemaining = estimatedTimeRemaining {
            dict["estimatedTimeRemaining"] = estimatedTimeRemaining
        }

        return dict
    }
}
