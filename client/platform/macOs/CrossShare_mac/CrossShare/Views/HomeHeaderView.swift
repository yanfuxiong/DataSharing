//
//  HomeHeaderView.swift
//  CrossShare
//
//  Created by user00 on 2025/3/6.
//

import Cocoa

class HomeHeaderView: NSView {
    
    var tapMoreBtnBlock:(() -> ())?
    var tapDeviceListBtnBlock:(() -> ())?
    private var imageName: String = "connStatus1"
    
    override init(frame frameRect: NSRect) {
        super.init(frame: frameRect)
        setupUI()
        setupNotifications()
    }
    
    required init?(coder: NSCoder) {
        fatalError("init(coder:) has not been implemented")
    }
    
    func setupUI(){
        addSubview(leftImgView)
        addSubview(countBtn)
        addSubview(iconImgView)
        addSubview(titleImgView)
        addSubview(moreBtn)
        addSubview(deviceBtn)
        deviceBtn.addSubview(badgeLabel)
        
        leftImgView.isHidden = true
//        countBtn.isHidden = true
        leftImgView.snp.makeConstraints { make in
            make.centerY.equalToSuperview()
            make.left.equalToSuperview().offset(24)
            make.size.equalTo(CGSize(width: 44, height: 44))
        }
        
        countBtn.snp.makeConstraints { make in
            make.centerY.equalToSuperview()
            make.left.equalTo(leftImgView.snp.right).offset(12)
            make.size.equalTo(CGSize(width: 40, height: 20))
        }
        
        iconImgView.snp.makeConstraints { make in
            make.centerY.equalToSuperview()
            make.size.equalTo(CGSizeMake(50, 31.5))
            make.right.equalTo(self.snp.centerX).offset(-55)
        }
        
        titleImgView.snp.makeConstraints { make in
            make.centerY.equalToSuperview()
            make.size.equalTo(CGSizeMake(130.5, 16.5))
            make.left.equalTo(iconImgView.snp.right).offset(10)
        }
        
        moreBtn.snp.makeConstraints { make in
            make.centerY.equalToSuperview()
            make.right.equalToSuperview().offset(-16)
            make.size.equalTo(CGSize(width: 80, height: 30))
        }
        
        deviceBtn.snp.makeConstraints { make in
            make.centerY.equalToSuperview()
            make.right.equalTo(moreBtn.snp.left).offset(-14)
        }
        
        badgeLabel.snp.makeConstraints { make in
            make.right.equalTo(deviceBtn.snp.right).offset(-8)
            make.bottom.equalTo(deviceBtn.snp.bottom).offset(-7)
        }
    }

    override func draw(_ dirtyRect: NSRect) {
        super.draw(dirtyRect)

    }
    
    lazy var leftImgView: NSImageView = {
        let cview = NSImageView(frame: .zero)
        cview.wantsLayer = true
        cview.image = NSImage(named: "Icons")
        cview.imageScaling = .scaleAxesIndependently
        cview.layer?.backgroundColor = NSColor.clear.cgColor
        return cview
    }()
    
    lazy var countBtn: NSButton = {
        let button = NSButton(title: "0", target: self, action: #selector(changeCountAction(_:)))
        button.wantsLayer = true
        button.layer?.backgroundColor = NSColor.lightGray.cgColor
        button.bezelStyle = .regularSquare
        button.isBordered = false
        button.layer?.cornerRadius = 5
        let attributes: [NSAttributedString.Key: Any] = [
            .foregroundColor: NSColor.black,
            .backgroundColor: NSColor.clear,
            .font: NSFont.systemFont(ofSize: 13)
        ]
        let attributedTitle = NSAttributedString(string: "0", attributes: attributes)
        button.attributedTitle = attributedTitle
        return button
    }()
    
    lazy var iconImgView: NSImageView = {
        let cview = NSImageView(frame: .zero)
        cview.wantsLayer = true
        cview.image = NSImage(named: "Group_3")
        cview.imageScaling = .scaleAxesIndependently
        cview.layer?.backgroundColor = NSColor.clear.cgColor
        return cview
    }()
    
    lazy var titleImgView: NSImageView = {
        let cview = NSImageView(frame: .zero)
        cview.wantsLayer = true
        cview.image = NSImage(named: "Cross_Share")
        cview.imageScaling = .scaleAxesIndependently
        cview.layer?.backgroundColor = NSColor.clear.cgColor
        return cview
    }()
    
    lazy var deviceBtn: NSButton = {
        let button = NSButton(title: "", target: self, action: #selector(tapDeviceListBtn(_:)))
        button.wantsLayer = true
        button.isBordered = false
        button.bezelStyle = .texturedRounded
        if let image = NSImage(named: self.imageName) {
            image.size = NSSize(width: 75, height: 66)
            button.image = image
        }
        return button
    }()
    
    // 文字标签的懒加载属性
    lazy var badgeLabel: NSTextField = {
        let label = NSTextField(labelWithString: "")
        label.textColor = NSColor.white
        label.font = NSFont.systemFont(ofSize: 14, weight: .bold)
        label.backgroundColor = .clear
        label.isBordered = false
        label.alignment = .center
        return label
    }()
    
    lazy var moreBtn: NSButton = {
        let button = NSButton(title: "", target: self, action: #selector(tapMoreBtn(_:)))
        button.wantsLayer = true
        button.layer?.backgroundColor = NSColor.clear.cgColor
        button.bezelStyle = .regularSquare
        button.isBordered = false
        button.layer?.cornerRadius = 5
        
        // 设置图标
        if let moreImage = NSImage(named: "moreBtn") {
            button.image = moreImage
            button.imageScaling = .scaleProportionallyDown
        } else {
            print("Warning: moreBtn image not found")
        }
        
        return button
    }()
    
    private func setupNotifications() {
        HelperClient.shared.onCountUpdated = { [weak self] newCount in
            DispatchQueue.main.async {
                let attributes: [NSAttributedString.Key: Any] = [
                    .foregroundColor: NSColor.black,
                    .backgroundColor: NSColor.clear,
                    .font: NSFont.systemFont(ofSize: 13)
                ]
                let attributedTitle = NSAttributedString(string: "\(newCount)", attributes: attributes)
                self?.countBtn.attributedTitle = attributedTitle
            }
        }
        
        // 添加对deviceDiasStatusNotification通知的监听
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(handleDeviceDiasStatusNotification(_:)),
            name: .deviceDiasStatusNotification,
            object: nil
        )
    }
    
    @objc private func handleDeviceDiasStatusNotification(_ notification: Notification) {
        // 处理DIAS状态通知
        if let userInfo = notification.userInfo,
           let diasStatus = userInfo["diasStatus"] as? Int {
            if(diasStatus == 6 || diasStatus == 7){
                imageName = "connStatus6"
                self.badgeLabel.isHidden = false
            }else{
                imageName = "connStatus\(diasStatus)"
                self.badgeLabel.isHidden = true
            }
            updateDeviceBtnImage(imageName)
        }
    }
    
    func updateDeviceBtnImage(_ imageName: String) {
        // Update deviceBtn image on main thread
        DispatchQueue.main.async {
            self.deviceBtn.image = NSImage(named: self.imageName)
        }
    }
}

extension HomeHeaderView {
    
    @objc func changeCountAction(_ sender:NSButton) {
        HelperClient.shared.connect { [weak self] success, error in
            if success {
                print("Helper connected successfully")
                
                let currentCount = Int(sender.title) ?? 0
                let newCount = currentCount + 1
                
                HelperClient.shared.updateCount(newCount) { updatedCount in
                    DispatchQueue.main.async {
                        let attributes: [NSAttributedString.Key: Any] = [
                            .foregroundColor: NSColor.black,
                            .backgroundColor: NSColor.clear,
                            .font: NSFont.systemFont(ofSize: 13)
                        ]
                        let attributedTitle = NSAttributedString(string: "\(updatedCount)", attributes: attributes)
                        self?.countBtn.attributedTitle = attributedTitle
                        
                        print("Count updated: \(currentCount) -> \(updatedCount)")
                    }
                }
            } else {
                print("Failed to connect to Helper: \(error ?? "Unknown error")")
            }
        }
    }

    
    @objc func tapMoreBtn(_ sender:NSButton) {
        if let block = self.tapMoreBtnBlock {
            block()
        }
    }
    
    @objc func tapDeviceListBtn(_ sender:NSButton) {
        if let block = self.tapDeviceListBtnBlock {
            block()
        }
    }
    
    func refreshDeviceList(_ devices: [CrossShareDevice]) {
        let deviceCount = devices.count
        self.badgeLabel.stringValue = "\(deviceCount)"
    }
}
