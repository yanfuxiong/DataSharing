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
    static func createNoDeviceMenuItem(_ title: String) -> NSMenuItem {
        // 创建菜单项
        let menuItem = NSMenuItem()
        menuItem.isEnabled = false
        
        // 创建菜单项视图
        let menuItemView = NSView(frame: NSRect(x: 0, y: 0, width: 180, height: 36))
        
        // 添加占位图标
    //        let placeholderIcon = NSImageView(frame: NSRect(x: 8, y: 8, width: 24, height: 24))
    //        placeholderIcon.image = NSImage(systemSymbolName: "wifi.slash", accessibilityDescription: "No devices")
    //        placeholderIcon.image?.isTemplate = true
    //        placeholderIcon.contentTintColor = NSColor.secondaryLabelColor
    //        menuItemView.addSubview(placeholderIcon)
        
        // 添加"No online devices"文本
        let noDeviceLabel = NSTextField()
        noDeviceLabel.stringValue = title
        noDeviceLabel.font = NSFont.systemFont(ofSize: 13, weight: .medium)
        noDeviceLabel.textColor = NSColor.black
        noDeviceLabel.isEditable = false
        noDeviceLabel.isSelectable = false
        noDeviceLabel.isBordered = false
        noDeviceLabel.backgroundColor = .clear
        menuItemView.addSubview(noDeviceLabel)
        noDeviceLabel.snp.makeConstraints {
            $0.left.equalTo(40)
            $0.centerY.equalToSuperview()
            $0.right.equalToSuperview().offset(-8)
            $0.height.equalTo(20)
        }
        
        // 设置菜单项的视图
        menuItem.view = menuItemView
        
        return menuItem
    }
    
    // 创建设备菜单项（左边图标，右边文字和IP地址）
    static func createDeviceMenuItem(title: String, imageName: String, target: AnyObject?, tag: Int, ipAddr: String) -> NSMenuItem {
        // 创建菜单项
        let menuItem = NSMenuItem()
        menuItem.target = target
        menuItem.tag = tag
        
        // 创建菜单项视图
        let menuItemView = NSView(frame: NSRect(x: 0, y: 0, width: 180, height: 44))
        
        // 添加设备图标
        let imageView = NSImageView()
        if let deviceIcon = NSImage(named: imageName) {
            deviceIcon.size = NSSize(width: 16, height: 16) // 设置图标大小
            imageView.image = deviceIcon
        }
        menuItemView.addSubview(imageView)
        imageView.snp.makeConstraints { make in
            make.left.equalTo(8)
            make.centerY.equalToSuperview()
            make.width.height.equalTo(16)
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
        titleLabel.isEditable = false
        titleLabel.isSelectable = false
        titleLabel.isBordered = false
        titleLabel.backgroundColor = .clear
        textContainerView.addSubview(titleLabel)
        titleLabel.snp.makeConstraints { make in
            make.top.equalTo(2)
            make.left.right.equalToSuperview()
            make.height.equalTo(16)
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
            make.top.equalTo(titleLabel.snp.bottom).offset(4) // 间距为8
            make.left.right.equalToSuperview()
            make.height.equalTo(12)
            make.bottom.equalToSuperview()
        }
        
        // 设置菜单项的视图
        menuItem.view = menuItemView
        return menuItem
    }

}

