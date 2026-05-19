//
//  CustomMenuItem.swift
//  CrossShare
//
//  Created by TS on 2025/10/9.
//

import Cocoa
import SnapKit

class CustomMenuItem:NSObject{
    // Create custom "No online devices" menu item
    static func createNoDeviceMenuItem(_ status: Int) -> NSMenuItem {
        // Create menu item
        let menuItem = NSMenuItem()
        menuItem.isEnabled = false
        
        // Create menu item view (width adapts to status text)
        let statusText = gainStatusTextDescribe(status)
        let textAttributes: [NSAttributedString.Key: Any] = [
            .font: NSFont.systemFont(ofSize: 13, weight: .medium)
        ]
        let textWidth = ceil((statusText as NSString).size(withAttributes: textAttributes).width)
        let contentWidth = 8 + 20 + 8 + textWidth + 8 + 10

        let menuItemView = NSView(frame: NSRect(x: 0, y: 0, width: contentWidth, height: 27))
        menuItemView.wantsLayer = true
        menuItemView.layer?.backgroundColor = NSColor.white.cgColor
        
        // Add placeholder icon
        let placeholderIcon = NSImageView()
        if let deviceIcon = NSImage(named: gainStatusImageName(status)) {
            deviceIcon.size = NSSize(width: 20, height: 20) // Set icon size
            placeholderIcon.image = deviceIcon
        }
        menuItemView.addSubview(placeholderIcon)

        
        placeholderIcon.snp.makeConstraints { make in
            make.left.equalTo(8)
            make.centerY.equalToSuperview()
            make.width.equalTo(20)
            make.height.equalTo(20)
        }
        
        // Add "No online devices" text
        let noDeviceLabel = NSTextField()
        noDeviceLabel.stringValue = statusText
        noDeviceLabel.font = NSFont.systemFont(ofSize: 13, weight: .medium)
        noDeviceLabel.textColor = NSColor.hex("666666")
        noDeviceLabel.isEditable = false
        noDeviceLabel.isSelectable = false
        noDeviceLabel.isBordered = false
        noDeviceLabel.backgroundColor = .clear
        menuItemView.addSubview(noDeviceLabel)
        noDeviceLabel.snp.makeConstraints {
            $0.left.equalTo(placeholderIcon.snp.right).offset(8)
            $0.centerY.equalToSuperview()
            $0.right.equalToSuperview().offset(-8)
            $0.height.equalTo(20)
        }
        
        // Set menu item view
        menuItem.view = menuItemView
        
        return menuItem
    }
    
    private static func gainStatusTextDescribe(_ status:Int) -> String{
        switch status {
        case 1:
            return "Detecting monitor..."
        case 2:
            return "Searching for service..."
        case 3:
            return "Checking authorization..."
        case 5:
            return "Authorization failed!"
        case 6,7:
            return "Connected"
        case 8:
            return "The connection is interrupted"
        default:
            return "Detecting monitor..."
        }
    }
    
    
    private static func gainStatusImageName(_ status:Int) -> String{
        if(status == 6 || status == 7){
            return "connTextStatus6"
        }else if(status >= 1){
            return "connTextStatus\(status)"
        }else{
            return "connTextStatus1"
        }
    }
    
    // Create device menu item (icon on left, text and IP address on right)
    static func createDeviceMenuItem(title: String, imageName: String, target: AnyObject?, tag: Int, ipAddr: String) -> NSMenuItem {
        // Create menu item
        let menuItem = NSMenuItem()
        menuItem.target = target
        menuItem.tag = tag
        
        // Create menu item view
        let menuItemView = NSView(frame: NSRect(x: 0, y: 0, width: 250, height: 50))
        
        // Add device icon
        let imageView = NSImageView()
        if let deviceIcon = NSImage(named: imageName) {
            deviceIcon.size = NSSize(width: 35, height: 35) // Set icon size
            imageView.image = deviceIcon
        }
        menuItemView.addSubview(imageView)
        imageView.snp.makeConstraints { make in
            make.left.equalTo(8)
            make.centerY.equalToSuperview()
            make.width.equalTo(35)
            make.height.equalTo(35)
        }
        
        // Create text container view for titleLabel and ipAddrLabel
        let textContainerView = NSView()
        menuItemView.addSubview(textContainerView)
        textContainerView.snp.makeConstraints { make in
            make.left.equalTo(imageView.snp.right).offset(10)
            make.centerY.equalToSuperview()
            make.right.equalToSuperview().offset(-8)
        }
        
        // Add device name text
        let titleLabel = NSTextField()
        titleLabel.stringValue = title
        titleLabel.font = NSFont.systemFont(ofSize: 13, weight: .medium)
        titleLabel.textColor =  NSColor.hex("666666")
        titleLabel.lineBreakMode = .byTruncatingTail
        titleLabel.isEditable = false
        titleLabel.isSelectable = false
        titleLabel.isBordered = false
        titleLabel.backgroundColor = .clear
        textContainerView.addSubview(titleLabel)
        titleLabel.snp.makeConstraints { make in
            make.top.equalTo(imageView.snp.top)
            make.left.right.equalToSuperview()
            // make.height.equalTo(16)
        }
        
        // Add IP address text (12pt font, light gray)
        let ipAddrLabel = NSTextField()
        ipAddrLabel.stringValue = ipAddr
        ipAddrLabel.font = NSFont.systemFont(ofSize: 12)
        ipAddrLabel.textColor = NSColor.secondaryLabelColor
        ipAddrLabel.textColor =  NSColor.hex("bababa")
        ipAddrLabel.isEditable = false
        ipAddrLabel.isSelectable = false
        ipAddrLabel.isBordered = false
        ipAddrLabel.backgroundColor = .clear
        textContainerView.addSubview(ipAddrLabel)
        ipAddrLabel.snp.makeConstraints { make in
            make.top.equalTo(titleLabel.snp.bottom) // Spacing is 8
            make.left.right.equalToSuperview()
            // make.height.equalTo(12)
            make.bottom.equalTo(imageView.snp.bottom)
        }
        
        // Set menu item view
        menuItem.view = menuItemView
        return menuItem
    }

}

