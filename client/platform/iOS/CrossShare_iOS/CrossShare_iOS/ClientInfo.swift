//
//  ClientInfo.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/5/9.
//

import UIKit
import SwiftyJSON

struct ClientInfo {
    let ip: String
    let id: String
    let name: String
    let deviceType: String
    let sourcePortType: String
    let version: String

    enum CodingKeys: String, CodingKey {
        case ip = "IpAddr"
        case id = "ID"
        case name = "DeviceName"
        case deviceType = "Platform"
        case sourcePortType = "SourcePortType"
        case version = "Version"
    }

    init(ip: String, id: String, name: String, deviceType: String, sourcePortType: String = "", version: String = "") {
        self.ip = ip
        self.id = id
        self.name = name
        self.deviceType = deviceType
        self.sourcePortType = sourcePortType
        self.version = version
    }

    init?(json: JSON) {
        guard let ip = json[CodingKeys.ip.rawValue].string,
              let id = json[CodingKeys.id.rawValue].string,
              let name = json[CodingKeys.name.rawValue].string,
              let platform = json[CodingKeys.deviceType.rawValue].string else {
            return nil
        }
        self.ip = ip
        self.id = id
        self.name = name
        self.deviceType = platform
        self.sourcePortType = json[CodingKeys.sourcePortType.rawValue].string ?? ""
        self.version = json[CodingKeys.version.rawValue].string ?? ""
    }

    init?(dict: [String: Any]) {
        guard let ip = dict[CodingKeys.ip.rawValue] as? String,
              let id = dict[CodingKeys.id.rawValue] as? String,
              let name = dict[CodingKeys.name.rawValue] as? String,
              let platform = dict[CodingKeys.deviceType.rawValue] as? String else {
            return nil
        }
        self.ip = ip
        self.id = id
        self.name = name
        self.deviceType = platform
        self.sourcePortType = dict[CodingKeys.sourcePortType.rawValue] as? String ?? ""
        self.version = dict[CodingKeys.version.rawValue] as? String ?? ""
    }

    func toDictionary() -> [String: Any] {
        return [
            CodingKeys.ip.rawValue: ip,
            CodingKeys.id.rawValue: id,
            CodingKeys.name.rawValue: name,
            CodingKeys.deviceType.rawValue: deviceType,
            CodingKeys.sourcePortType.rawValue: sourcePortType,
            CodingKeys.version.rawValue: version
        ]
    }

    var deviceIconName: String {
        switch sourcePortType {
        case "HDMI1":
            return "hdmi"
        case "HDMI2":
            return "hdmi2"
        case "USBC1":
            return "usb_c1"
        case "USBC2":
            return "usb_c2"
        case "DP1":
            return "dp1"
        case "DP2":
            return "dp2"
        case "Miracast":
            return "miracast"
        default:
            return platformIconName
        }
    }

    var platformIconName: String {
        switch deviceType.lowercased() {
        case "windows":
            return "computer"
        case "android":
            return "computer"
        case "ios":
            return "computer"
        case "macos":
            return "computer"
        default:
            return "computer"
        }
    }

    var displayName: String {
        if !version.isEmpty {
            return "\(name) (\(deviceType) \(version))"
        } else if !deviceType.isEmpty {
            return "\(name) (\(deviceType))"
        } else {
            return name
        }
    }

    var isWirelessConnection: Bool {
        return sourcePortType.lowercased() == "miracast"
    }
}

struct AuthData:Equatable,Codable {
    var width: Int
    var height: Int
    var framerate: Int
    var type: Int
    var displayName: String? = ""
    
    enum CodingKeys: String, CodingKey {
        case width = "Width"
        case height = "Height"
        case framerate = "Framerate"
        case type = "Type"
        case displayName = "DisplayName"
    }
}

struct TransferTask {
    let taskId: String
    let clientId: String
    let clientIp: String
    let clientName: String
    let filePaths: [String]
    let startTime: Date
    var status: TransferStatus
    
    enum TransferStatus {
        case preparing
        case started
        case inProgress
        case completed
        case failed
        case cancel
    }
}

struct LanServiceInfo {
    var monitorName: String
    var instance: String
    var ip: String
    var version: String = ""
    var timestamp: UInt64 = 0
    
    func toDictionary() -> [String: Any] {
        return ["monitorName": monitorName, "instance": instance, "ip": ip, "version": version, "timestamp": timestamp]
    }
}
    
