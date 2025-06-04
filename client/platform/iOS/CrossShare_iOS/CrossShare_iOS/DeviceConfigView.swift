//
//  DeviceConfigView.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/5/14.
//

import UIKit

class DeviceConfigView: UIView {
    
    var submitConfigBlock:((String,String,String) -> ())?

    override init(frame: CGRect) {
        super.init(frame: frame)
        setupUI()
    }
    
    required init?(coder: NSCoder) {
        fatalError("init(coder:) has not been implemented")
    }
    
    private func setupUI() {
        ddcciLable.snp.makeConstraints { make in
            make.left.equalToSuperview().offset(30)
            make.top.equalToSuperview().offset(10)
        }
        
        ddcciTextFiled.snp.makeConstraints { make in
            make.centerY.equalTo(ddcciLable)
            make.left.equalToSuperview().offset(140)
            make.right.equalToSuperview().offset(-20)
            make.height.equalTo(40)
        }
        
        deviceSourceLable.snp.makeConstraints { make in
            make.left.equalTo(ddcciLable)
            make.top.equalTo(ddcciLable.snp.bottom).offset(40)
        }
        
        deviceSourceTextFiled.snp.makeConstraints { make in
            make.centerY.equalTo(deviceSourceLable)
            make.left.equalToSuperview().offset(140)
            make.right.equalToSuperview().offset(-20)
            make.height.equalTo(40)
        }
        
        devicePortLable.snp.makeConstraints { make in
            make.left.equalTo(ddcciLable)
            make.top.equalTo(deviceSourceLable.snp.bottom).offset(40)
        }
        
        devicePortTextFiled.snp.makeConstraints { make in
            make.centerY.equalTo(devicePortLable)
            make.left.equalToSuperview().offset(140)
            make.right.equalToSuperview().offset(-20)
            make.height.equalTo(40)
        }
        
        submitButton.snp.makeConstraints { (make) in
            make.centerX.equalToSuperview()
            make.top.equalTo(devicePortTextFiled.snp.bottom).offset(40)
            make.width.equalTo(200)
            make.height.equalTo(40)
        }
    }
    
    lazy var ddcciLable: UILabel = {
        let text = UILabel(frame: .zero)
        text.textColor = UIColor.black
        text.font = UIFont.systemFont(ofSize: 14)
        text.text = "diasID"
        addSubview(text)
        return text
    }()
    
    lazy var ddcciTextFiled: UITextField = {
        let textFiled = UITextField(frame: .zero)
        textFiled.textColor = UIColor.black
        textFiled.font = UIFont.systemFont(ofSize: 14)
        textFiled.borderStyle = .roundedRect
        textFiled.text = ""
        textFiled.placeholder = "Pleas enter ddcci Id"
        addSubview(textFiled)
        return textFiled
    }()
    
    lazy var deviceSourceLable: UILabel = {
        let text = UILabel(frame: .zero)
        text.textColor = UIColor.black
        text.font = UIFont.systemFont(ofSize: 14)
        text.text = "deviceSource"
        addSubview(text)
        return text
    }()
    
    lazy var deviceSourceTextFiled: UITextField = {
        let textFiled = UITextField(frame: .zero)
        textFiled.textColor = UIColor.black
        textFiled.font = UIFont.systemFont(ofSize: 14)
        textFiled.borderStyle = .roundedRect
        textFiled.placeholder = "Pleas enter deviceSource"
        textFiled.keyboardType = .numberPad
        addSubview(textFiled)
        return textFiled
    }()
    
    lazy var devicePortLable: UILabel = {
        let text = UILabel(frame: .zero)
        text.textColor = UIColor.black
        text.font = UIFont.systemFont(ofSize: 14)
        text.text = "devicePort"
        addSubview(text)
        return text
    }()
    
    lazy var devicePortTextFiled: UITextField = {
        let textFiled = UITextField(frame: .zero)
        textFiled.textColor = UIColor.black
        textFiled.font = UIFont.systemFont(ofSize: 14)
        textFiled.borderStyle = .roundedRect
        textFiled.placeholder = "Pleas enter devicePort"
        textFiled.keyboardType = .numberPad
        addSubview(textFiled)
        return textFiled
    }()
    
    lazy var submitButton: UIButton = {
        let button = UIButton(type: .custom)
        button.setTitle("Set Device Config", for: .normal)
        button.setTitleColor(UIColor.systemBlue, for: .normal)
        button.addTarget(self, action: #selector(submitConfig), for: .touchUpInside)
        addSubview(button)
        return button
    }()
}

extension DeviceConfigView {
    
    @objc func submitConfig() {
        if let submitConfigBlock = submitConfigBlock {
            let ddcciText = self.ddcciTextFiled.text ?? ""
            let deviceSourceText = self.deviceSourceTextFiled.text ?? ""
            let devicePortText = self.devicePortTextFiled.text ?? ""
            submitConfigBlock(ddcciText,deviceSourceText,devicePortText)
        }
    }
    
    public func refreshUI() {
        if let diasId = UserDefaults.get(forKey: .DEVICECONFIG_DIAS_ID) {
            ddcciTextFiled.text = diasId
        }
        if let src = UserDefaults.getInt(forKey: .DEVICECONFIG_SRC) {
            deviceSourceTextFiled.text = String(src)
        }
        if let port = UserDefaults.getInt(forKey: .DEVICECONFIG_PORT) {
            devicePortTextFiled.text = String(port)
        }
    }
}
