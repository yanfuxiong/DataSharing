//
//  FIleCancelPopView.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/7/11.
//

import UIKit

class FileCancelPopView: UIView {
    
    var onCancel: (() -> Void)?
    var onSure: (() -> Void)?
    
    private let contentView = UIView()
    private let titleLabel = UILabel()
    private let subTitleLabel = UILabel()
    private let cancelButton = UIButton(type: .custom)
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
        
        let tap = UITapGestureRecognizer(target: self, action: #selector(cancelAction))
        tap.delegate = self
        self.addGestureRecognizer(tap)
        
        contentView.backgroundColor = .white
        contentView.layer.cornerRadius = 16
        contentView.clipsToBounds = true
        addSubview(contentView)
        contentView.snp.makeConstraints { make in
            make.left.equalToSuperview().offset(20)
            make.right.equalToSuperview().offset(-20)
            make.center.equalToSuperview()
            make.height.equalTo(220)
        }
        
        titleLabel.text = "Cancel all transfers in progress"
        titleLabel.numberOfLines = 0
        titleLabel.textAlignment = .center
        titleLabel.font = UIFont.boldSystemFont(ofSize: 16)
        contentView.addSubview(titleLabel)
        titleLabel.snp.makeConstraints { make in
            make.top.equalToSuperview().offset(10)
            make.centerX.equalToSuperview()
        }
        
        subTitleLabel.text = "All you sure want to cancel all transfers ?"
        subTitleLabel.numberOfLines = 0
        subTitleLabel.textAlignment = .center
        subTitleLabel.font = UIFont.boldSystemFont(ofSize: 13)
        contentView.addSubview(subTitleLabel)
        subTitleLabel.snp.makeConstraints { make in
            make.top.equalTo(titleLabel.snp.bottom).offset(20)
            make.centerX.equalToSuperview()
        }
        
        sureButton.setTitle("Comfirm", for: .normal)
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
        }
        
        cancelButton.setTitle("Cancel", for: .normal)
        cancelButton.setTitleColor(UIColor.black, for: .normal)
        cancelButton.backgroundColor = UIColor.white
        cancelButton.layerCornerRadius = 5
        cancelButton.layer.shadowColor = UIColor.black.cgColor
        cancelButton.layer.shadowOpacity = 0.3
        cancelButton.layer.shadowOffset = CGSize(width: 0, height: 2)
        cancelButton.layer.masksToBounds = false
        cancelButton.addTarget(self, action: #selector(cancelAction), for: .touchUpInside)
        contentView.addSubview(cancelButton)
        
        cancelButton.snp.makeConstraints { make in
            make.top.equalTo(sureButton.snp.bottom).offset(15)
            make.height.centerX.equalTo(sureButton)
            make.right.equalToSuperview().offset(-40)
        }
    }
    
    @objc func transportFiles() {
        onSure?()
    }
    
    @objc private func cancelAction() {
        onCancel?()
    }
    
}

extension FileCancelPopView: UIGestureRecognizerDelegate {
    func gestureRecognizer(_ gestureRecognizer: UIGestureRecognizer, shouldReceive touch: UITouch) -> Bool {
        return touch.view == self
    }
}
