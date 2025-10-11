//
//  UIHelper.swift
//  CrossShare
//
//  Created by user00 on 2025/3/5.
//

import Cocoa

class UIHelper: NSObject {
    
    func createNormalLabelWithString(text:String,color:NSColor?,fontSize:CGFloat) -> NSTextField {
        let textFiled = NSTextField(frame: .zero)
        textFiled.isEditable = false
        textFiled.isSelectable = false
        textFiled.drawsBackground = false
        if let color = color {
            textFiled.textColor = color
        }
        textFiled.font = NSFont.systemFont(ofSize: fontSize)
        return textFiled
    }

}
