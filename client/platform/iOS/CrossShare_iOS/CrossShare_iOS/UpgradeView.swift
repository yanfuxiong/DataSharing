//
//  UpgradeView.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/8/12.
//

import UIKit

class UpgradeView: UIView {
    
    var onCancel: (() -> Void)?
    var onSure: (() -> Void)?
    
    private let contentView = UIView()
    private let titleLabel = UILabel()
    private let subTitleLabel = UILabel()
    private let sureButton = UIButton(type: .custom)
    
    override init(frame: CGRect) {
        super.init(frame: frame)
        setupUI()
    }
    
    required init?(coder: NSCoder) {
        fatalError("init(coder:) has not been implemented")
    }
    
    func setupUI() {
        self.backgroundColor = UIColor.black.withAlphaComponent(0.4)
        
        contentView.backgroundColor = .white
        contentView.layer.cornerRadius = 16
        contentView.clipsToBounds = true
        addSubview(contentView)
        contentView.snp.makeConstraints { make in
            make.left.equalToSuperview().offset(20)
            make.right.equalToSuperview().offset(-20)
            make.center.equalToSuperview()
        }
        
        titleLabel.text = "Version Mismatch Detected"
        titleLabel.numberOfLines = 0
        titleLabel.textAlignment = .center
        titleLabel.font = UIFont.boldSystemFont(ofSize: 16)
        contentView.addSubview(titleLabel)
        titleLabel.snp.makeConstraints { make in
            make.top.equalToSuperview().offset(14)
            make.centerX.equalToSuperview()
        }
        
        subTitleLabel.text = """
                          Your CrossShare version (\(P2PManager.shared.majorVersion)) is lower than another connected client.
                          
                          Please update to continue.
                          """
        subTitleLabel.numberOfLines = 0
        subTitleLabel.textAlignment = .center
        subTitleLabel.font = UIFont.systemFont(ofSize: 13)
        contentView.addSubview(subTitleLabel)
        subTitleLabel.snp.makeConstraints { make in
            make.top.equalTo(titleLabel.snp.bottom).offset(20)
            make.centerX.equalToSuperview()
            make.left.equalToSuperview().offset(20)
        }
        
        sureButton.setTitle("Upgrade now", for: .normal)
        sureButton.setTitleColor(UIColor.white, for: .normal)
        sureButton.backgroundColor = UIColor.systemBlue
        sureButton.layerCornerRadius = 5
        sureButton.addTarget(self, action: #selector(transportFiles), for: .touchUpInside)
        contentView.addSubview(sureButton)
        
        sureButton.snp.makeConstraints { make in
            make.top.equalTo(subTitleLabel.snp.bottom).offset(25)
            make.height.equalTo(40)
            make.left.equalToSuperview().offset(40)
            make.centerX.equalToSuperview()
            make.bottom.equalToSuperview().offset(-20)
        }
    }
    
    @objc func transportFiles() {
        onSure?()
    }
}

extension UpgradeView: UIGestureRecognizerDelegate {
    func gestureRecognizer(_ gestureRecognizer: UIGestureRecognizer, shouldReceive touch: UITouch) -> Bool {
        return touch.view == self
    }
}
