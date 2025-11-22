//
//  CSMenuItemWithArrowView.swift
//  CrossShare
//
//  Created by Assistant on 2025/10/20.
//

import Cocoa
import SnapKit

class CSMenuItemWithArrowView: NSView {
    
    // MARK: - Properties
    
    private let titleText: String
    private let arrowIconName: String
    private var clickAction: (() -> Void)?
    
    // MARK: - UI Components
    
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
    
    private lazy var arrowImageView: NSImageView = {
        let imageView = NSImageView()
        if let image = NSImage(named: arrowIconName) {
            image.size = NSSize(width: 12, height: 12)
            imageView.image = image
        }
        imageView.imageScaling = .scaleProportionallyDown
        return imageView
    }()
    
    private var trackingArea: NSTrackingArea?
    
    // MARK: - Initialization
    
    init(title: String, arrowIconName: String = "rightArrow", action: @escaping () -> Void) {
        self.titleText = title
        self.arrowIconName = arrowIconName
        self.clickAction = action
        super.init(frame: NSRect(x: 0, y: 0, width: 190, height: 30))
        setupUI()
        setupTrackingArea()
    }
    
    required init?(coder: NSCoder) {
        self.titleText = ""
        self.arrowIconName = "rightArrow"
        super.init(coder: coder)
        setupUI()
        setupTrackingArea()
    }
    
    // MARK: - UI Setup
    
    private func setupUI() {
        wantsLayer = true
        layer?.backgroundColor = NSColor.clear.cgColor
        
        addSubview(titleLabel)
        addSubview(arrowImageView)
        
        setupConstraints()
    }
    
    private func setupConstraints() {
        titleLabel.snp.makeConstraints { make in
            make.left.equalToSuperview().offset(20)
            make.centerY.equalToSuperview()
        }
        
        arrowImageView.snp.makeConstraints { make in
            make.right.equalToSuperview().offset(-10)
            make.centerY.equalToSuperview()
            make.width.height.equalTo(12)
        }
    }
    
    // MARK: - Tracking Area
    
    private func setupTrackingArea() {
        let options: NSTrackingArea.Options = [.mouseEnteredAndExited, .activeAlways]
        trackingArea = NSTrackingArea(rect: bounds, options: options, owner: self, userInfo: nil)
        addTrackingArea(trackingArea!)
    }
    
    override func updateTrackingAreas() {
        super.updateTrackingAreas()
        if let trackingArea = trackingArea {
            removeTrackingArea(trackingArea)
        }
        setupTrackingArea()
    }
    
    // MARK: - Mouse Events
    
    override func mouseEntered(with event: NSEvent) {
        layer?.backgroundColor = NSColor.selectedContentBackgroundColor.withAlphaComponent(0.1).cgColor
    }
    
    override func mouseExited(with event: NSEvent) {
        layer?.backgroundColor = NSColor.clear.cgColor
    }
    
    override func mouseDown(with event: NSEvent) {
        clickAction?()
        // Close the menu
        self.enclosingMenuItem?.menu?.cancelTracking()
    }
}

