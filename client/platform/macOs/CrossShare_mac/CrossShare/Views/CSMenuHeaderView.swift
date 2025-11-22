//
//  CSMenuHeaderView.swift
//  CrossShare
//
//  Created by Assistant on 2025/10/20.
//

import Cocoa
import SnapKit

class CSMenuHeaderView: NSView {
    
    // MARK: - Properties
    
    private let titleText: String
    private let iconName: String?
    
    // MARK: - UI Components
    
    private lazy var iconImageView: NSImageView? = {
        guard let iconName = iconName else { return nil }
        let imageView = NSImageView()
        if let image = NSImage(named: iconName) {
            image.size = NSSize(width: 16, height: 16)
            imageView.image = image
        }
        imageView.imageScaling = .scaleProportionallyDown
        return imageView
    }()
    
    private lazy var titleLabel: NSTextField = {
        let label = NSTextField(labelWithString: titleText)
        label.font = NSFont.systemFont(ofSize: 13)
        label.textColor = NSColor.black
        label.alignment = .left
        label.isEditable = false
        label.isBordered = false
        label.backgroundColor = .clear
        return label
    }()
    
    // MARK: - Initialization
    
    init(title: String, iconName: String? = nil) {
        self.titleText = title
        self.iconName = iconName
        super.init(frame: NSRect(x: 0, y: 0, width: 190, height: 30))
        setupUI()
    }
    
    required init?(coder: NSCoder) {
        self.titleText = ""
        self.iconName = nil
        super.init(coder: coder)
        setupUI()
    }
    
    // MARK: - UI Setup
    
    private func setupUI() {
        wantsLayer = true
        layer?.backgroundColor = NSColor.clear.cgColor
        
        if let iconImageView = iconImageView {
            addSubview(iconImageView)
            addSubview(titleLabel)
            
            iconImageView.snp.makeConstraints { make in
                make.left.equalToSuperview().offset(20)
                make.centerY.equalToSuperview()
                make.width.height.equalTo(16)
            }
            
            titleLabel.snp.makeConstraints { make in
                make.left.equalTo(iconImageView.snp.right).offset(8)
                make.centerY.equalToSuperview()
                make.right.lessThanOrEqualToSuperview().offset(-10)
            }
        } else {
            addSubview(titleLabel)
            
            titleLabel.snp.makeConstraints { make in
                make.left.equalToSuperview().offset(20)
                make.centerY.equalToSuperview()
                make.right.lessThanOrEqualToSuperview().offset(-10)
            }
        }
    }
}

