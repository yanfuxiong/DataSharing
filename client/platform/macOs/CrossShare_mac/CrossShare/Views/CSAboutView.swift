//
//  CSAboutView.swift
//  CrossShare
//
//  Created by Assistant on 2025/10/15.
//

import Cocoa
import SnapKit

class CSAboutView: NSView {
    
    // MARK: - UI Components
    
    // Top icon
    private lazy var iconImageView: NSImageView = {
        let imageView = NSImageView()
        imageView.image = NSImage(named: "info_about")
        imageView.imageScaling = .scaleProportionallyDown
        return imageView
    }()
    
    // About title
    private lazy var titleLabel: NSTextField = {
        let label = NSTextField(labelWithString: "About")
        label.font = NSFont.boldSystemFont(ofSize: 18)
        label.textColor = NSColor.labelColor
        label.alignment = .left
        label.isEditable = false
        label.isBordered = false
        label.backgroundColor = .clear
        return label
    }()
    
    // Separator line
    private lazy var separatorLine: NSBox = {
        let box = NSBox()
        box.boxType = .separator
        return box
    }()
    
    // Info container
    private lazy var infoContainerView: NSView = {
        let view = NSView()
        return view
    }()
    
    // APP Version label
    private lazy var appVersionTitleLabel: NSTextField = {
        let label = NSTextField(labelWithString: "APP Version:")
        label.font = NSFont.systemFont(ofSize: 13, weight: .medium)
        label.textColor = NSColor.labelColor
        label.alignment = .left
        label.isEditable = false
        label.isBordered = false
        label.backgroundColor = .clear
        return label
    }()
    
    private lazy var appVersionValueLabel: NSTextField = {
        let label = NSTextField(labelWithString: UtilsHelper.getVersionNumber())
        label.font = NSFont.systemFont(ofSize: 13)
        label.textColor = NSColor.secondaryLabelColor
        label.alignment = .left
        label.isEditable = false
        label.isBordered = false
        label.backgroundColor = .clear
        return label
    }()
    
    // Service Version label
    private lazy var serviceVersionTitleLabel: NSTextField = {
        let label = NSTextField(labelWithString: "Service Version:")
        label.font = NSFont.systemFont(ofSize: 13, weight: .medium)
        label.textColor = NSColor.labelColor
        label.alignment = .left
        label.isEditable = false
        label.isBordered = false
        label.backgroundColor = .clear
        return label
    }()
    
    private lazy var serviceVersionValueLabel: NSTextField = {
        let label = NSTextField(labelWithString: "")
        label.font = NSFont.systemFont(ofSize: 13)
        label.textColor = NSColor.secondaryLabelColor
        label.alignment = .left
        label.isEditable = false
        label.isBordered = false
        label.backgroundColor = .clear
        return label
    }()
    
    // Name label
    private lazy var nameTitleLabel: NSTextField = {
        let label = NSTextField(labelWithString: "Name:")
        label.font = NSFont.systemFont(ofSize: 13, weight: .medium)
        label.textColor = NSColor.labelColor
        label.alignment = .left
        label.isEditable = false
        label.isBordered = false
        label.backgroundColor = .clear
        return label
    }()
    
    private lazy var nameValueLabel: NSTextField = {
        let label = NSTextField(labelWithString: Host.current().localizedName ?? "Mac")
        label.font = NSFont.systemFont(ofSize: 13)
        label.textColor = NSColor.secondaryLabelColor
        label.alignment = .left
        label.isEditable = false
        label.isBordered = false
        label.backgroundColor = .clear
        return label
    }()
    
    // IP label
    private lazy var ipTitleLabel: NSTextField = {
        let label = NSTextField(labelWithString: "IP:")
        label.font = NSFont.systemFont(ofSize: 13, weight: .medium)
        label.textColor = NSColor.labelColor
        label.alignment = .left
        label.isEditable = false
        label.isBordered = false
        label.backgroundColor = .clear
        return label
    }()
    
    private lazy var ipValueLabel: NSTextField = {
        let label = NSTextField(labelWithString: "Fetching...")
        label.font = NSFont.systemFont(ofSize: 13)
        label.textColor = NSColor.secondaryLabelColor
        label.alignment = .left
        label.isEditable = false
        label.isBordered = false
        label.backgroundColor = .clear
        return label
    }()
    
    // Main content
    private lazy var contentTextView: NSTextView = {
        let textView = NSTextView()
        textView.isEditable = false
        textView.isSelectable = true
        textView.backgroundColor = .clear
        textView.textColor = NSColor.labelColor
        textView.font = NSFont.systemFont(ofSize: 12)
        textView.textContainerInset = NSSize(width: 0, height: 0)
        textView.string = """
        CrossShare simplifies your workflow by enabling cross-platform copy-paste, file transfers, and multi-device control —all on one screen.
        With seamless multi-screen integration, you can manage documents across systems effortlessly, boosting productivity and streamlining everyday tasks.
        """
        return textView
    }()
    
    // Bottom copyright info
    private lazy var copyrightLabel: NSTextField = {
        let label = NSTextField(labelWithString: "© 2025 Realtek Semiconductor Corp. All rights reserved")
        label.font = NSFont.systemFont(ofSize: 11)
        label.textColor = NSColor.tertiaryLabelColor
        label.alignment = .center
        label.isEditable = false
        label.isBordered = false
        label.backgroundColor = .clear
        return label
    }()
    
    // MARK: - Initialization
    
    override init(frame frameRect: NSRect) {
        super.init(frame: frameRect)
        setupUI()
        loadIPAddress()
        loadSystemInfo()
    }
    
    required init?(coder: NSCoder) {
        super.init(coder: coder)
        setupUI()
        loadIPAddress()
        loadSystemInfo()
    }
    
    // MARK: - UI Setup
    
    private func setupUI() {
        wantsLayer = true
        layer?.backgroundColor = NSColor.controlBackgroundColor.cgColor
        
        // Add subviews
        addSubview(iconImageView)
        addSubview(titleLabel)
        addSubview(separatorLine)
        addSubview(infoContainerView)
        addSubview(contentTextView)
        addSubview(copyrightLabel)
        
        // Add info labels to container
        infoContainerView.addSubview(appVersionTitleLabel)
        infoContainerView.addSubview(appVersionValueLabel)
        infoContainerView.addSubview(serviceVersionTitleLabel)
        infoContainerView.addSubview(serviceVersionValueLabel)
        infoContainerView.addSubview(nameTitleLabel)
        infoContainerView.addSubview(nameValueLabel)
        infoContainerView.addSubview(ipTitleLabel)
        infoContainerView.addSubview(ipValueLabel)
        
        setupConstraints()
    }
    
    private func setupConstraints() {
        // Icon constraints - 8px from left
        iconImageView.snp.makeConstraints { make in
            make.top.equalToSuperview().offset(10)
            make.left.equalToSuperview().offset(8)
            make.width.height.equalTo(24)
        }
        
        // Title constraints - to the right of icon
        titleLabel.snp.makeConstraints { make in
            make.centerY.equalTo(iconImageView)
            make.left.equalTo(iconImageView.snp.right).offset(8)
            make.right.equalToSuperview().offset(-8)
        }
        
        // Separator line constraints - 8px below title
        separatorLine.snp.makeConstraints { make in
            make.top.equalTo(titleLabel.snp.bottom).offset(8)
            make.left.equalToSuperview().offset(8)
            make.right.equalToSuperview().offset(-8)
            make.height.equalTo(1)
        }
        
        // Info container constraints
        infoContainerView.snp.makeConstraints { make in
            make.top.equalTo(separatorLine.snp.bottom).offset(16)
            make.left.equalToSuperview().offset(8)
            make.right.equalToSuperview().offset(-8)
            make.height.equalTo(120)
        }
        
        // APP Version
        appVersionTitleLabel.snp.makeConstraints { make in
            make.top.equalToSuperview()
            make.left.equalToSuperview()
            make.width.equalTo(120)
        }
        
        appVersionValueLabel.snp.makeConstraints { make in
            make.centerY.equalTo(appVersionTitleLabel)
            make.left.equalTo(appVersionTitleLabel.snp.right).offset(8)
            make.right.equalToSuperview()
        }
        
        // Service Version
        serviceVersionTitleLabel.snp.makeConstraints { make in
            make.top.equalTo(appVersionTitleLabel.snp.bottom).offset(12)
            make.left.equalToSuperview()
            make.width.equalTo(120)
        }
        
        serviceVersionValueLabel.snp.makeConstraints { make in
            make.centerY.equalTo(serviceVersionTitleLabel)
            make.left.equalTo(serviceVersionTitleLabel.snp.right).offset(8)
            make.right.equalToSuperview()
        }
        
        // Name
        nameTitleLabel.snp.makeConstraints { make in
            make.top.equalTo(serviceVersionTitleLabel.snp.bottom).offset(12)
            make.left.equalToSuperview()
            make.width.equalTo(120)
        }
        
        nameValueLabel.snp.makeConstraints { make in
            make.centerY.equalTo(nameTitleLabel)
            make.left.equalTo(nameTitleLabel.snp.right).offset(8)
            make.right.equalToSuperview()
        }
        
        // IP
        ipTitleLabel.snp.makeConstraints { make in
            make.top.equalTo(nameTitleLabel.snp.bottom).offset(12)
            make.left.equalToSuperview()
            make.width.equalTo(120)
        }
        
        ipValueLabel.snp.makeConstraints { make in
            make.centerY.equalTo(ipTitleLabel)
            make.left.equalTo(ipTitleLabel.snp.right).offset(8)
            make.right.equalToSuperview()
        }
        
        // Main content constraints
        contentTextView.snp.makeConstraints { make in
            make.top.equalTo(infoContainerView.snp.bottom).offset(20)
            make.left.equalToSuperview().offset(8)
            make.right.equalToSuperview().offset(-8)
            make.bottom.equalTo(copyrightLabel.snp.top).offset(-20)
        }
        
        // Copyright info constraints - centered at bottom
        copyrightLabel.snp.makeConstraints { make in
            make.left.equalToSuperview().offset(8)
            make.right.equalToSuperview().offset(-8)
            make.bottom.equalToSuperview().offset(-20)
        }
    }
    
    // MARK: - Data Loading
    
    private func loadIPAddress() {
        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            var ipAddress = "Not Available"
            
            // Give priority to reading from the local getCurrentSystemInfoDict.
            if let systemInfoDict = HelperClient.shared.getCurrentSystemInfoDict(),
               let ipInfo = systemInfoDict["ipInfo"] as? String, !ipInfo.isEmpty {
                ipAddress = ipInfo
            } else {
                // If it cannot be read, then use the local IP instead.
                ipAddress = self?.getLocalIPAddress() ?? "Not Available"
            }
            
            DispatchQueue.main.async {
                self?.ipValueLabel.stringValue = ipAddress
            }
        }
    }
    
    private func getLocalIPAddress() -> String? {
        var address: String?
        var ifaddr: UnsafeMutablePointer<ifaddrs>?
        
        guard getifaddrs(&ifaddr) == 0 else { return nil }
        guard let firstAddr = ifaddr else { return nil }
        
        for ifptr in sequence(first: firstAddr, next: { $0.pointee.ifa_next }) {
            let interface = ifptr.pointee
            let addrFamily = interface.ifa_addr.pointee.sa_family
            
            if addrFamily == UInt8(AF_INET) {
                let name = String(cString: interface.ifa_name)
                if name == "en0" || name == "en1" {
                    var hostname = [CChar](repeating: 0, count: Int(NI_MAXHOST))
                    getnameinfo(interface.ifa_addr,
                               socklen_t(interface.ifa_addr.pointee.sa_len),
                               &hostname,
                               socklen_t(hostname.count),
                               nil,
                               socklen_t(0),
                               NI_NUMERICHOST)
                    address = String(cString: hostname)
                    break
                }
            }
        }
        
        freeifaddrs(ifaddr)
        return address
    }
    
    // MARK: - System Info Loading
    
    private func loadSystemInfo() {
        if let systemInfoDict = HelperClient.shared.getCurrentSystemInfoDict(),
           let ipInfo = systemInfoDict["ipInfo"] as? String,
           let verInfo = systemInfoDict["verInfo"] as? String {
            logger.info("CSAboutView: Found system info - IP: \(ipInfo), Version: \(verInfo)")
            updateSystemInfo(ipInfo: ipInfo, verInfo: verInfo)
        }
    }
    
    private func updateSystemInfo(ipInfo: String, verInfo: String) {
        DispatchQueue.main.async { [weak self] in
            self?.serviceVersionValueLabel.stringValue = verInfo
            self?.ipValueLabel.stringValue = ipInfo
        }
    }
}

