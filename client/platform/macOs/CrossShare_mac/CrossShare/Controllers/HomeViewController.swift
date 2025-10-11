//
//  HomeViewController.swift
//  CrossShare
//
//  Created by user00 on 2025/3/5.
//

import Cocoa
import SnapKit
import Foundation

class HomeViewController: NSViewController {
    
    lazy var contentView: NSView = {
        let cview = NSView(frame: .zero)
        cview.wantsLayer = true
        cview.layer?.backgroundColor = NSColor.lightGray.cgColor
        return cview
    }()
    
    lazy var topView: HomeHeaderView = {
        let cview = HomeHeaderView(frame: .zero)
        cview.wantsLayer = true
        cview.layer?.backgroundColor = NSColor(hex: 0x377AF6).cgColor
        return cview
    }()
    
    lazy var middleView: SelectFilesBorderView = {
        let cview = SelectFilesBorderView(frame: .zero)
        cview.wantsLayer = true
        cview.layer?.backgroundColor = NSColor.clear.cgColor
        return cview
    }()
    
    lazy var centerYview: NSView = {
        let cview = NSView(frame: .zero)
        cview.wantsLayer = true
        cview.layer?.backgroundColor = NSColor.clear.cgColor
        return cview
    }()
    
    lazy var versionLabel: NSTextField = {
        let label = NSTextField(labelWithString: "V1.0.0")
        label.frame = CGRect()
        label.font = NSFont.systemFont(ofSize: 13)
        label.alignment = .center
        label.textColor = NSColor.black
        return label
    }()
    
    lazy var controlView: NSView = {
        let view = NSView()
        view.wantsLayer = true
        view.layer?.backgroundColor = NSColor.controlBackgroundColor.cgColor
        view.layer?.cornerRadius = 8
        return view
    }()
    
    override func viewDidLoad() {
        super.viewDidLoad()
        setupUI()
        setupData()
    }
    
    deinit {
        // Remove notification observers
        NotificationCenter.default.removeObserver(self)
    }
    
    func setupUI() {
        self.view.wantsLayer = true
        self.view.layer?.backgroundColor = NSColor.clear.cgColor
        view.addSubview(contentView)
        contentView.addSubview(centerYview)
        contentView.addSubview(topView)
        contentView.addSubview(middleView)
        contentView.addSubview(controlView)
        contentView.addSubview(versionLabel)
    
        contentView.snp.makeConstraints { make in
            make.edges.equalToSuperview()
        }
        
        centerYview.snp.makeConstraints { make in
            make.width.equalTo(2)
            make.centerX.top.bottom.equalToSuperview()
        }
        
        topView.snp.makeConstraints { make in
            make.left.right.top.equalToSuperview()
            make.height.equalTo(80)
        }
        
        middleView.snp.makeConstraints { make in
            make.left.right.equalTo(topView)
            make.top.equalTo(topView.snp.bottom).offset(16)
            make.height.equalTo(330)
        }
        
        versionLabel.snp.makeConstraints { make in
            make.left.equalToSuperview().offset(14)
            make.bottom.equalToSuperview().offset(-14)
        }
    }
    
    func setupData() {
        self.versionLabel.stringValue = UtilsHelper.getVersionNumber()
        self.topView.tapMoreBtnBlock = {
            UtilsHelper.checkClipboardFileCategory()
        }
    }
    
    override func viewDidAppear() {
        super.viewDidAppear()
    }
}
