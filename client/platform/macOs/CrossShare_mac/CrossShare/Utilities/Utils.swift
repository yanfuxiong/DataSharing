//
//  Utils.swift
//  CrossShare
//
//  Created by user00 on 2025/3/5.
//

import Cocoa
import Alamofire
import UniformTypeIdentifiers


extension UserDefaults {
    enum KEY: String {
        // v2ray-core version
        case xRayCoreVersion
        // v2ray server item list
        case v2rayServerList
    }
    
    static func setBool(forKey key: KEY, value: Bool) {
        UserDefaults.standard.set(value, forKey: key.rawValue)
    }
    
    static func getBool(forKey key: KEY) -> Bool {
        return UserDefaults.standard.bool(forKey: key.rawValue)
    }
    
    static func set(forKey key: KEY, value: String) {
        UserDefaults.standard.set(value, forKey: key.rawValue)
    }
    
    static func get(forKey key: KEY) -> String? {
        return UserDefaults.standard.string(forKey: key.rawValue)
    }
    
    static func del(forKey key: KEY) {
        UserDefaults.standard.removeObject(forKey: key.rawValue)
    }
    
    static func setArray(forKey key: KEY, value: [String]) {
        UserDefaults.standard.set(value, forKey: key.rawValue)
    }
    
    static func getArray(forKey key: KEY) -> [String]? {
        return UserDefaults.standard.array(forKey: key.rawValue) as? [String]
    }
    
    static func delArray(forKey key: KEY) {
        UserDefaults.standard.removeObject(forKey: key.rawValue)
    }
}

extension NSColor {
    convenience init(hex: Int, alpha: CGFloat = 1.0) {
        let red   = CGFloat((hex >> 16) & 0xFF) / 255.0
        let green = CGFloat((hex >> 8) & 0xFF) / 255.0
        let blue  = CGFloat(hex & 0xFF) / 255.0
        self.init(red: red, green: green, blue: blue, alpha: alpha)
    }
}

extension UTType {
    static let word = UTType(importedAs: "com.microsoft.word.doc")
    static let wordLegacy = UTType(importedAs: "com.microsoft.word.binary")
    
    static let powerpoint = UTType(importedAs: "com.microsoft.powerpoint.presentation")
    static let powerpointLegacy = UTType(importedAs: "com.microsoft.powerpoint.ppt")
    
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
