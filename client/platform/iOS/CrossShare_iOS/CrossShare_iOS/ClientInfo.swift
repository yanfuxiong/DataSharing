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
    
    func toDictionary() -> [String: Any] {
        return ["ip": ip, "id": id, "name": name]
    }
}
