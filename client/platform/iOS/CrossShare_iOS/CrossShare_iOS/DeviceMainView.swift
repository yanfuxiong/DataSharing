//
//  DeviceMainView.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/5/13.
//

import UIKit

class DeviceMainView: UIView {
    
    var dosomothingBlock:(() -> ())?

    override init(frame: CGRect) {
        super.init(frame: frame)
        setupUI()
    }
    
    required init?(coder: NSCoder) {
        fatalError("init(coder:) has not been implemented")
    }
    
    private func setupUI() {
        addSubview(ipTitleLable)
        addSubview(ipLable)
        addSubview(verTitleLable)
        addSubview(versionLable)
        addSubview(cliTitleLable)
        addSubview(clientsLable)
        addSubview(ddcTitleLable)
        addSubview(ddcciLable)
        
        ipTitleLable.snp.makeConstraints { make in
            make.left.equalToSuperview().offset(16.adaptW)
            make.top.equalToSuperview().offset(34.adaptH)
        }
        
        ipLable.snp.makeConstraints { make in
            make.left.equalTo(ipTitleLable)
            make.top.equalTo(ipTitleLable.snp.bottom).offset(6.adaptH)
        }
        
        cliTitleLable.snp.makeConstraints { make in
            make.left.equalTo(ipTitleLable)
            make.top.equalTo(ipLable.snp.bottom).offset(24.adaptH)
        }
        
        clientsLable.snp.makeConstraints { make in
            make.left.equalTo(ipTitleLable)
            make.top.equalTo(cliTitleLable.snp.bottom).offset(6)
        }
        
        verTitleLable.snp.makeConstraints { make in
            make.left.equalTo(ipTitleLable)
            make.top.equalTo(clientsLable.snp.bottom).offset(24)
        }
        
        versionLable.snp.makeConstraints { make in
            make.left.equalTo(ipTitleLable)
            make.top.equalTo(verTitleLable.snp.bottom).offset(6)
        }
        
        ddcTitleLable.snp.makeConstraints { make in
            make.left.equalTo(verTitleLable)
            make.top.equalTo(versionLable.snp.bottom).offset(24)
        }
        
        ddcciLable.snp.makeConstraints { make in
            make.left.equalTo(verTitleLable)
            make.top.equalTo(ddcTitleLable.snp.bottom).offset(6)
        }
        
        licenseBtn.snp.makeConstraints { make in
            make.right.equalToSuperview().offset(-20)
            make.height.equalTo(17)
            make.width.equalTo(17)
            make.centerY.equalTo(ddcciLable)
        }
    }
    
    lazy var verTitleLable: UILabel = {
        let text = UILabel(frame: .zero)
        text.textColor = UIColor.init(hex: 0xC6BDBD)
        text.font = UIFont.systemFont(ofSize: 14)
        text.text = "Software version:"
        return text
    }()
    
    lazy var versionLable: UILabel = {
        let text = UILabel(frame: .zero)
        text.textColor = UIColor.black
        text.font = UIFont.systemFont(ofSize: 16)
        text.text = "Software version:"
        return text
    }()
    
    lazy var ipTitleLable: UILabel = {
        let text = UILabel(frame: .zero)
        text.textColor = UIColor.init(hex: 0xC6BDBD)
        text.font = UIFont.systemFont(ofSize: 14)
        text.text = "Current IP"
        return text
    }()
    
    lazy var ipLable: UILabel = {
        let text = UILabel(frame: .zero)
        text.textColor = UIColor.black
        text.font = UIFont.systemFont(ofSize: 16)
        text.text = "My ip & Device Name:"
        addSubview(text)
        return text
    }()
    
    lazy var cliTitleLable: UILabel = {
        let text = UILabel(frame: .zero)
        text.textColor = UIColor.init(hex: 0xC6BDBD)
        text.font = UIFont.systemFont(ofSize: 14)
        text.text = "Device Name"
        addSubview(text)
        return text
    }()
    
    lazy var clientsLable: UILabel = {
        let text = UILabel(frame: .zero)
        text.textColor = UIColor.black
        text.numberOfLines = 0
        text.font = UIFont.systemFont(ofSize: 16)
        text.text = "NA"
        addSubview(text)
        return text
    }()
    
    lazy var ddcTitleLable: UILabel = {
        let text = UILabel(frame: .zero)
        text.textColor = UIColor.init(hex: 0xC6BDBD)
        text.font = UIFont.systemFont(ofSize: 14)
        text.text = "Application info"
        addSubview(text)
        return text
    }()
    
    lazy var ddcciLable: UILabel = {
        let text = UILabel(frame: .zero)
        text.textColor = UIColor.black
        text.font = UIFont.systemFont(ofSize: 16)
        text.text = "Licenses"
        addSubview(text)
        return text
    }()
    
    lazy var licenseBtn: UIButton = {
        let button = UIButton(type: .custom)
        button.setImage(UIImage(named: "license"), for: .normal)
        button.setImage(UIImage(named: "license"), for: .selected)
        button.addTarget(self, action: #selector(dosomething(_:)), for: .touchUpInside)
        addSubview(button)
        return button
    }()
}

extension DeviceMainView {
    
    @objc func dosomething(_ sender:UIButton) {
        if let dosomothingBlock = dosomothingBlock {
            dosomothingBlock()
        }
    }
    
    public func refreshUI() {
        self.versionLable.text = P2PManager.shared.version
        self.ipLable.text = "\(P2PManager.shared.ip)"
        var mclintString = ""
        let clients = P2PManager.shared.clientList
        for client in clients {
            mclintString.append(client.ip)
            mclintString.append(" ")
            mclintString.append(client.name)
            mclintString.append("\n")
        }
        mclintString = clients.isEmpty ? "NA" : mclintString
        self.clientsLable.text = P2PManager.shared.deviceName
//        self.ddcciLable.text = "\(P2PManager.shared.deviceDiasId)"
    }
}
