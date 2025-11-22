//
//  CustomMenuItem.swift
//  CrossShare
//
//  Created by TS on 2025/10/9.
//

import Cocoa
import SnapKit

class CustomMenuItem:NSObject{
    // 创建"No online devices"自定义菜单项
    static func createNoDeviceMenuItem(_ status: Int) -> NSMenuItem {
        // 创建菜单项
        let menuItem = NSMenuItem()
        menuItem.isEnabled = false
        
        // 创建菜单项视图
        let menuItemView = NSView(frame: NSRect(x: 0, y: 0, width: 220, height: 36))
        
        // 添加占位图标
        let placeholderIcon = NSImageView()
        if let deviceIcon = NSImage(named: gainStatusImageName(status)) {
            deviceIcon.size = NSSize(width: 20, height: 20) // 设置图标大小
            placeholderIcon.image = deviceIcon
        }
        menuItemView.addSubview(placeholderIcon)

        
        placeholderIcon.snp.makeConstraints { make in
            make.left.equalTo(8)
            make.centerY.equalToSuperview()
            make.width.equalTo(20)
            make.height.equalTo(20)
        }
        
        // 添加"No online devices"文本
        let noDeviceLabel = NSTextField()
        noDeviceLabel.stringValue = gainStatusTextDescribe(status)
        noDeviceLabel.font = NSFont.systemFont(ofSize: 13, weight: .medium)
        noDeviceLabel.textColor = NSColor.black
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
        
        // 设置菜单项的视图
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
    
    // 创建设备菜单项（左边图标，右边文字和IP地址）
    static func createDeviceMenuItem(title: String, imageName: String, target: AnyObject?, tag: Int, ipAddr: String) -> NSMenuItem {
        // 创建菜单项
        let menuItem = NSMenuItem()
        menuItem.target = target
        menuItem.tag = tag
        
        // 创建菜单项视图
        let menuItemView = NSView(frame: NSRect(x: 0, y: 0, width: 220, height: 50))
        
        // 添加设备图标
        let imageView = NSImageView()
        if let deviceIcon = NSImage(named: imageName) {
            deviceIcon.size = NSSize(width: 35, height: 35) // 设置图标大小
            imageView.image = deviceIcon
        }
        menuItemView.addSubview(imageView)
        imageView.snp.makeConstraints { make in
            make.left.equalTo(8)
            make.centerY.equalToSuperview()
            make.width.equalTo(35)
            make.height.equalTo(35)
        }
        
        // 创建文本容器视图，用于放置titleLabel和ipAddrLabel
        let textContainerView = NSView()
        menuItemView.addSubview(textContainerView)
        textContainerView.snp.makeConstraints { make in
            make.left.equalTo(imageView.snp.right).offset(10)
            make.centerY.equalToSuperview()
            make.right.equalToSuperview().offset(-8)
        }
        
        // 添加设备名称文本
        let titleLabel = NSTextField()
        titleLabel.stringValue = title
        titleLabel.font = NSFont.systemFont(ofSize: 13, weight: .medium)
        titleLabel.textColor = NSColor.black
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
        
        // 添加IP地址文本（12号字体，浅灰色）
        let ipAddrLabel = NSTextField()
        ipAddrLabel.stringValue = ipAddr
        ipAddrLabel.font = NSFont.systemFont(ofSize: 12)
        ipAddrLabel.textColor = NSColor.secondaryLabelColor
        ipAddrLabel.isEditable = false
        ipAddrLabel.isSelectable = false
        ipAddrLabel.isBordered = false
        ipAddrLabel.backgroundColor = .clear
        textContainerView.addSubview(ipAddrLabel)
        ipAddrLabel.snp.makeConstraints { make in
            make.top.equalTo(titleLabel.snp.bottom) // 间距为8
            make.left.right.equalToSuperview()
            // make.height.equalTo(12)
            make.bottom.equalTo(imageView.snp.bottom)
        }
        
        // 设置菜单项的视图
        menuItem.view = menuItemView
        return menuItem
    }

}

