//
//  Utils.swift
//  CrossShare
//
//  Created by user00 on 2025/3/5.
//

import Foundation
import Alamofire
import UniformTypeIdentifiers

let networkInterface = "en0"
let kBtnHeight: CGFloat = 44

extension GoString {
    func toString() -> String {
        guard let pointer = self.p else { return "" }
        let data = Data(bytes: pointer, count: Int(self.n))
        Logger.info("\(self.n)")
        return String(data: data, encoding: .utf8) ?? ""
    }
}

extension String {
    func toGoString() -> _GoString_ {
        let cString = strdup(self)
        return _GoString_(p: cString, n: Int(Int64(self.utf8.count)))
    }
    
    func base64ToImage() -> UIImage? {
        guard let imageData = Data(base64Encoded: self, options: .ignoreUnknownCharacters) else {
            Logger.info("Failed to decode Base64 string")
            return nil
        }
        let image = UIImage(data: imageData)
        return image
    }
    
    func toDictionary() -> [String: Any]? {
        guard let data = self.data(using: .utf8) else {
            Logger.info("Error converting JSON string to Data")
            return nil
        }
        do {
            let jsonObject = try JSONSerialization.jsonObject(with: data, options: .allowFragments)
            return jsonObject as? [String: Any]
        } catch {
            Logger.info("Error converting JSON string to dictionary: \(error)")
            return nil
        }
    }
}

extension Dictionary where Key == String, Value: Any {
    func toJsonString() -> String? {
        do {
            let jsonData = try JSONSerialization.data(withJSONObject: self, options: .prettyPrinted)
            return String(data: jsonData, encoding: .utf8)
        } catch {
            Logger.info("Error converting dictionary to JSON string: \(error)")
            return nil
        }
    }
}

extension UIImage {
    func imageToBase64() -> String? {
        guard let imageData = self.jpegData(compressionQuality: 1.0) else {
            Logger.info("Failed to convert image to data")
            return nil
        }
        let base64String = imageData.base64EncodedString()
        return base64String
    }
}

extension URL {
    var queryParameters: [String: String]? {
        guard let components = URLComponents(url: self, resolvingAgainstBaseURL: false),
              let queryItems = components.queryItems else { return nil }
        var params = [String: String]()
        for item in queryItems {
            params[item.name] = item.value
        }
        return params
    }
}

extension UserDefaults {
    
    static var groupId: String {
        if let appGroups = Bundle.main.object(forInfoDictionaryKey: "com.apple.security.application-groups") as? [String],
           let firstGroup = appGroups.first {
            return firstGroup
        }
        Logger.info("Warning: App group ID not found in entitlements, using fallback value")
        return "group.com.realtek.crossshare"
    }
    
    enum KEY: String {
        case DEVICECONFIG_DIAS_ID
        case DEVICECONFIG_SRC
        case DEVICECONFIG_PORT
        case DEVICE_CLIENTS
        case FILE_PATH
        case ISACTIVITY
        case SEND_TO_DEVICE
        case LAN_SERVICE_INFO
    }
    
    enum DefaultsType {
        case standard
        case group
        case both
    }
    
    static func userDefaults(type: DefaultsType = .standard) -> UserDefaults {
        switch type {
        case .standard:
            return UserDefaults.standard
        case .group, .both:
            return UserDefaults(suiteName: groupId) ?? UserDefaults.standard
        }
    }
    
    static func setBool(forKey key: KEY, value: Bool, type: DefaultsType = .standard) {
        if type == .both {
            UserDefaults.standard.set(value, forKey: key.rawValue)
            userDefaults(type: .group).set(value, forKey: key.rawValue)
        } else {
            userDefaults(type: type).set(value, forKey: key.rawValue)
        }
    }
    
    static func setStandardBool(forKey key: String, value: Bool, type: DefaultsType = .standard) {
        if type == .both {
            UserDefaults.standard.set(value, forKey: key)
            userDefaults(type: .group).set(value, forKey: key)
        } else {
            userDefaults(type: type).set(value, forKey: key)
        }
    }
    
    static func getBool(forKey key: KEY, type: DefaultsType = .standard) -> Bool {
        return userDefaults(type: type).bool(forKey: key.rawValue)
    }
    
    static func set(forKey key: KEY, value: String, type: DefaultsType = .standard) {
        if type == .both {
            UserDefaults.standard.set(value, forKey: key.rawValue)
            userDefaults(type: .group).set(value, forKey: key.rawValue)
        } else {
            userDefaults(type: type).set(value, forKey: key.rawValue)
        }
    }
    
    static func get(forKey key: KEY, type: DefaultsType = .standard) -> String? {
        return userDefaults(type: type).string(forKey: key.rawValue)
    }
    
    static func setInt(forKey key: KEY, value: Int, type: DefaultsType = .standard) {
        if type == .both {
            UserDefaults.standard.set(value, forKey: key.rawValue)
            userDefaults(type: .group).set(value, forKey: key.rawValue)
        } else {
            userDefaults(type: type).set(value, forKey: key.rawValue)
        }
    }
    
    static func setStandardInt(forKey key: String, value: Int, type: DefaultsType = .standard) {
        if type == .both {
            UserDefaults.standard.set(value, forKey: key)
            userDefaults(type: .group).set(value, forKey: key)
        } else {
            userDefaults(type: type).set(value, forKey: key)
        }
    }
    
    static func getInt(forKey key: KEY, type: DefaultsType = .standard) -> Int? {
        if userDefaults(type: type).object(forKey: key.rawValue) == nil {
            return nil
        }
        return userDefaults(type: type).integer(forKey: key.rawValue)
    }
    
    static func setArray(forKey key: KEY, value: [String], type: DefaultsType = .standard) {
        if type == .both {
            UserDefaults.standard.set(value, forKey: key.rawValue)
            userDefaults(type: .group).set(value, forKey: key.rawValue)
        } else {
            userDefaults(type: type).set(value, forKey: key.rawValue)
        }
    }
    
    static func getArray(forKey key: KEY, type: DefaultsType = .standard) -> [String]? {
        return userDefaults(type: type).array(forKey: key.rawValue) as? [String]
    }
    
    static func setURLArray(forKey key: KEY, value: [URL], type: DefaultsType = .standard) {
        let urlStrings = value.map { $0.absoluteString }
        if type == .both {
            UserDefaults.standard.set(urlStrings, forKey: key.rawValue)
            userDefaults(type: .group).set(urlStrings, forKey: key.rawValue)
        } else {
            userDefaults(type: type).set(urlStrings, forKey: key.rawValue)
        }
    }
    
    static func getURLArray(forKey key: KEY, type: DefaultsType = .standard) -> [URL]? {
        guard let urlStrings = userDefaults(type: type).array(forKey: key.rawValue) as? [String] else {
            return nil
        }
        return urlStrings.compactMap { URL(string: $0) }
    }
    
    static func del(forKey key: KEY, type: DefaultsType = .standard) {
        if type == .both {
            UserDefaults.standard.removeObject(forKey: key.rawValue)
            userDefaults(type: .group).removeObject(forKey: key.rawValue)
        } else {
            userDefaults(type: type).removeObject(forKey: key.rawValue)
        }
    }
    
    static func delArray(forKey key: KEY, type: DefaultsType = .standard) {
        del(forKey: key, type: type)
    }
    
    static func synchronize(type: DefaultsType = .standard) -> Bool {
        if type == .both {
            let standardResult = UserDefaults.standard.synchronize()
            let groupResult = userDefaults(type: .group).synchronize()
            return standardResult && groupResult
        } else {
            return userDefaults(type: type).synchronize()
        }
    }
}

extension UIColor {
    convenience init(hex: Int, alpha: CGFloat = 1.0) {
        let red   = CGFloat((hex >> 16) & 0xFF) / 255.0
        let green = CGFloat((hex >> 8) & 0xFF) / 255.0
        let blue  = CGFloat(hex & 0xFF) / 255.0
        self.init(red: red, green: green, blue: blue, alpha: alpha)
    }
}

// 添加对microsoft文件的支持
extension UTType {
    // Word 文档
    static let word = UTType(importedAs: "com.microsoft.word.doc")
    static let wordLegacy = UTType(importedAs: "com.microsoft.word.binary")
    // PowerPoint 演示文稿
    static let powerpoint = UTType(importedAs: "com.microsoft.powerpoint.presentation")
    static let powerpointLegacy = UTType(importedAs: "com.microsoft.powerpoint.ppt")
    // Excel 电子表格
    static let excel = UTType(importedAs: "com.microsoft.excel.xlsx")
    static let excelLegacy = UTType(importedAs: "com.microsoft.excel.xls")
}

extension Data {
    mutating func append(_ string: String) {
        if let data = string.data(using: .utf8) {
            append(data)
        }
    }
}
