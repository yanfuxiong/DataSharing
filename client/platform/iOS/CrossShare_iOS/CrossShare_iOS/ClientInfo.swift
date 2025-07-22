//
//  ClientInfo.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/5/9.
//

import UIKit
import SwiftyJSON

struct ClientInfo {
    var ip: String
    var id: String
    var name: String
    var deviceType: String
    
    func toDictionary() -> [String: Any] {
        return ["ip": ip, "id": id, "name": name,"deviceType": deviceType]
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
