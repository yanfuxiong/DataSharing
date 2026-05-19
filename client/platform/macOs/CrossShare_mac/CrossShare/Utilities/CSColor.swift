//
//  CSColor.swift
//  CrossShare
//
//  Created for color utility functions
//

import Cocoa

// Extension to convert hex string to NSColor
extension NSColor {
    static func hex(_ hexString: String) -> NSColor {
        var hexSanitized = hexString.trimmingCharacters(in: .whitespacesAndNewlines)
        hexSanitized = hexSanitized.replacingOccurrences(of: "#", with: "")
        
        var rgb: UInt64 = 0
        
        Scanner(string: hexSanitized).scanHexInt64(&rgb)
        
        let red = CGFloat((rgb & 0xFF0000) >> 16) / 255.0
        let green = CGFloat((rgb & 0x00FF00) >> 8) / 255.0
        let blue = CGFloat(rgb & 0x0000FF) / 255.0
        
        return NSColor(red: red, green: green, blue: blue, alpha: 1.0)
    }
}

