//
//  DeviceMainView.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/5/13.
//

import UIKit

class DeviceMainView: UIView {
    
    var dosomothingBlock:((Int) -> ())?

    override init(frame: CGRect) {
        super.init(frame: frame)
        setupUI()
    }
    
    required init?(coder: NSCoder) {
        fatalError("init(coder:) has not been implemented")
    }
    
    private func setupUI() {
        addSubview(verTitleLable)
        addSubview(versionLable)
        addSubview(ipTitleLable)
        addSubview(ipLable)
        addSubview(cliTitleLable)
        addSubview(clientsLable)
        addSubview(ddcTitleLable)
        addSubview(ddcciLable)
        
        verTitleLable.snp.makeConstraints { make in
            make.left.equalToSuperview().offset(16)
            make.top.equalToSuperview().offset(5)
        }
        
        versionLable.snp.makeConstraints { make in
            make.left.equalTo(verTitleLable)
            make.top.equalTo(verTitleLable.snp.bottom).offset(6)
        }
        
        ipTitleLable.snp.makeConstraints { make in
            make.left.equalTo(verTitleLable)
            make.top.equalTo(versionLable.snp.bottom).offset(24)
        }
        
        ipLable.snp.makeConstraints { make in
            make.left.equalTo(verTitleLable)
            make.top.equalTo(ipTitleLable.snp.bottom).offset(6)
        }
        
        cliTitleLable.snp.makeConstraints { make in
            make.left.equalTo(verTitleLable)
            make.top.equalTo(ipLable.snp.bottom).offset(24)
        }
        
        clientsLable.snp.makeConstraints { make in
            make.left.equalTo(verTitleLable)
            make.top.equalTo(cliTitleLable.snp.bottom).offset(6)
        }
        
        ddcTitleLable.snp.makeConstraints { make in
            make.left.equalTo(verTitleLable)
            make.top.equalTo(clientsLable.snp.bottom).offset(24)
        }
        
        ddcciLable.snp.makeConstraints { make in
            make.left.equalTo(verTitleLable)
            make.top.equalTo(ddcTitleLable.snp.bottom).offset(6)
        }
        
        settingButton.snp.makeConstraints { make in
            make.left.equalTo(verTitleLable)
            make.height.equalTo(40)
            make.width.equalTo(200)
            make.top.equalTo(ddcciLable.snp.bottom).offset(40)
        }
        
        uploadButton.snp.makeConstraints { make in
            make.left.equalTo(verTitleLable)
            make.height.equalTo(40)
            make.width.equalTo(200)
            make.top.equalTo(settingButton.snp.bottom).offset(30)
        }
        
        historyFileButton.snp.makeConstraints { make in
            make.left.equalTo(verTitleLable)
            make.height.equalTo(40)
            make.width.equalTo(200)
            make.top.equalTo(uploadButton.snp.bottom).offset(30)
        }
    }
    
    lazy var verTitleLable: UILabel = {
        let text = UILabel(frame: .zero)
        text.textColor = UIColor.black
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
        text.textColor = UIColor.black
        text.font = UIFont.systemFont(ofSize: 14)
        text.text = "My Ip & Device Name:"
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
        text.textColor = UIColor.black
        text.font = UIFont.systemFont(ofSize: 14)
        text.text = "Connection:"
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
        text.textColor = UIColor.black
        text.font = UIFont.systemFont(ofSize: 14)
        text.text = "Dias ID:"
        addSubview(text)
        return text
    }()
    
    lazy var ddcciLable: UILabel = {
        let text = UILabel(frame: .zero)
        text.textColor = UIColor.black
        text.font = UIFont.systemFont(ofSize: 16)
        text.text = "diasID"
        addSubview(text)
        return text
    }()
    
    lazy var settingButton: PaddingButton = {
        let button = PaddingButton(type: .custom)
        button.tag = 1
        button.layerBorderColor = UIColor.systemBlue
        button.layerCornerRadius = 5
        button.layerBorderWidth = 1
        button.titleLabel?.font = UIFont.systemFont(ofSize: 15)
        button.setTitle("Settings", for: .normal)
        button.setTitleColor(UIColor.systemBlue, for: .normal)
        button.setImage(UIImage(named: "settings-fill"), for: .normal)
        button.setImage(UIImage(named: "settings-fill"), for: .selected)
        button.addTarget(self, action: #selector(dosomething(_:)), for: .touchUpInside)
        addSubview(button)
        return button
    }()
    
    lazy var uploadButton: PaddingButton = {
        let button = PaddingButton(type: .custom)
        button.tag = 2
        button.layerBorderColor = UIColor.systemBlue
        button.layerCornerRadius = 5
        button.layerBorderWidth = 1
        button.titleLabel?.font = UIFont.systemFont(ofSize: 15)
        button.setTitle("Transport Files", for: .normal)
        button.setTitleColor(UIColor.systemBlue, for: .normal)
        button.setImage(UIImage(named: "upload"), for: .normal)
        button.setImage(UIImage(named: "upload"), for: .selected)
        button.addTarget(self, action: #selector(dosomething(_:)), for: .touchUpInside)
        addSubview(button)
        return button
    }()
    
    lazy var historyFileButton: PaddingButton = {
        let button = PaddingButton(type: .custom)
        button.tag = 3
        button.layerBorderColor = UIColor.systemBlue
        button.layerCornerRadius = 5
        button.layerBorderWidth = 1
        button.titleLabel?.font = UIFont.systemFont(ofSize: 15)
        button.setTitle("Transport History", for: .normal)
        button.setTitleColor(UIColor.systemBlue, for: .normal)
        button.setImage(UIImage(named: "history"), for: .normal)
        button.setImage(UIImage(named: "history"), for: .selected)
        button.addTarget(self, action: #selector(dosomething(_:)), for: .touchUpInside)
        addSubview(button)
        return button
    }()
}

extension DeviceMainView {
    
    @objc func dosomething(_ sender:PaddingButton) {
        if let dosomothingBlock = dosomothingBlock {
            dosomothingBlock(sender.tag)
        }
    }
    
    public func refreshUI() {
        self.versionLable.text = P2PManager.shared.version
        self.ipLable.text = "\(P2PManager.shared.ip)   \(P2PManager.shared.deviceName)"
        var mclintString = ""
        let clients = P2PManager.shared.clientList
        for client in clients {
            mclintString.append(client.ip)
            mclintString.append(" ")
            mclintString.append(client.name)
            mclintString.append("\n")
        }
        mclintString = clients.isEmpty ? "NA" : mclintString
        self.clientsLable.text = mclintString
        self.ddcciLable.text = "\(P2PManager.shared.deviceDiasId)"
    }
}

class PaddingButton: UIButton {
    override func titleRect(forContentRect contentRect: CGRect) -> CGRect {
        return CGRectMake(50, 10, self.width - 30 - 30, 20)
    }
    
    override func imageRect(forContentRect contentRect: CGRect) -> CGRect {
        return CGRectMake(10, 5, 30, 30)
    }
}
