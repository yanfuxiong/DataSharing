//
//  ShareFilesPopView.swift
//  share
//
//  Created by ts on 2025/8/6.
//

import UIKit

enum ShareFilesPopType {
    case mixedContent
    case foldersOnly
}

class ShareFilesPopView: UIView {
    var onCancel: (() -> Void)?
    var onContinue: (() -> Void)?
    var onConfirm: (() -> Void)?
    
    private let popType: ShareFilesPopType
    private let contentView = UIView()
    private let titleLabel = UILabel()
    private let contentLabel = UILabel()
    private let continueButton = UIButton(type: .custom)
    private let cancelButton = UIButton(type: .custom)
    private let confirmButton = UIButton(type: .custom)
    
    init(frame: CGRect, type: ShareFilesPopType) {
        self.popType = type
        super.init(frame: frame)
    }
    
    required init?(coder: NSCoder) {
        fatalError("init(coder:) has not been implemented")
    }
    
    func setupUI() {
        let tap = UITapGestureRecognizer(target: self, action: #selector(backgroundTapped))
        tap.delegate = self
        self.addGestureRecognizer(tap)
        
        self.snp.makeConstraints { make in
            make.left.right.equalToSuperview()
            make.centerY.equalToSuperview()
        }
        setupContentView()
        setupLabels()
        setupButtons()
    }
    
    private func setupContentView() {
        contentView.backgroundColor = .white
        contentView.layer.cornerRadius = 25
        contentView.clipsToBounds = true
        addSubview(contentView)
        
        contentView.snp.makeConstraints { make in
            make.width.equalToSuperview().multipliedBy(0.8)
            make.height.equalToSuperview().multipliedBy(0.5)
            make.center.equalToSuperview()
        }
    }
    
    private func setupLabels() {
        titleLabel.text = "Notice"
        titleLabel.numberOfLines = 0
        titleLabel.textAlignment = .center
        titleLabel.font = UIFont.boldSystemFont(ofSize: 20)
        titleLabel.textColor = .black
        contentView.addSubview(titleLabel)
        
        titleLabel.snp.makeConstraints { make in
            make.top.equalToSuperview().offset(20)
            make.left.right.equalToSuperview()
        }
        
        contentLabel.numberOfLines = 0
        contentLabel.textAlignment = .center
        contentLabel.font = UIFont.systemFont(ofSize: 16)
        contentLabel.textColor = .darkGray
        contentView.addSubview(contentLabel)
        
        switch popType {
        case .mixedContent:
            contentLabel.text = "Folders will not be transferred.\nOnly the selected files will be sent."
        case .foldersOnly:
            contentLabel.text = "Please select files only.\nFolder transfer is not supported."
        }
        
        contentLabel.snp.makeConstraints { make in
            make.top.equalTo(titleLabel.snp.bottom).offset(25)
            make.left.equalToSuperview().offset(20)
            make.right.equalToSuperview().offset(-20)
        }
    }
    
    private func setupButtons() {
        switch popType {
        case .mixedContent:
            setupMixedContentButtons()
        case .foldersOnly:
            setupFoldersOnlyButton()
        }
    }
    
    private func setupMixedContentButtons() {
        contentView.addSubview(continueButton)
        continueButton.snp.makeConstraints { make in
            make.top.equalTo(contentLabel.snp.bottom).offset(30)
            make.right.equalToSuperview().offset(-10)
            make.left.equalTo(contentView.snp.centerX).offset(5)
            make.height.equalTo(kBtnHeight)
            make.bottom.equalToSuperview().offset(-15)
        }
        continueButton.setTitle("Continue", for: .normal)
        continueButton.setTitleColor(.white, for: .normal)
        continueButton.backgroundColor = UIColor.systemBlue
        continueButton.layer.cornerRadius = kBtnHeight / 2
        continueButton.titleLabel?.font = UIFont.boldSystemFont(ofSize: 16)
        continueButton.addTarget(self, action: #selector(continueAction), for: .touchUpInside)
        
        contentView.addSubview(cancelButton)
        cancelButton.snp.makeConstraints { make in
            make.top.width.height.equalTo(continueButton)
            make.right.equalTo(contentView.snp.centerX).offset(-5)
        }
        cancelButton.setTitle("Cancel", for: .normal)
        cancelButton.setTitleColor(.systemBlue, for: .normal)
        cancelButton.backgroundColor = UIColor(red: 240/255.0, green: 240/255.0, blue: 240/255.0, alpha: 1.0)
        cancelButton.layer.cornerRadius = kBtnHeight / 2
        cancelButton.titleLabel?.font = UIFont.boldSystemFont(ofSize: 16)
        cancelButton.addTarget(self, action: #selector(cancelAction), for: .touchUpInside)
    }
    
    private func setupFoldersOnlyButton() {
        contentView.addSubview(confirmButton)
        confirmButton.snp.makeConstraints { make in
            make.top.equalTo(contentLabel.snp.bottom).offset(30)
            make.left.equalToSuperview().offset(20)
            make.right.equalToSuperview().offset(-20)
            make.height.equalTo(kBtnHeight)
            make.bottom.equalToSuperview().offset(-15)
        }
        confirmButton.setTitle("Confirm", for: .normal)
        confirmButton.setTitleColor(.white, for: .normal)
        confirmButton.backgroundColor = UIColor.systemBlue
        confirmButton.layer.cornerRadius = kBtnHeight / 2
        confirmButton.titleLabel?.font = UIFont.boldSystemFont(ofSize: 16)
        confirmButton.addTarget(self, action: #selector(confirmAction), for: .touchUpInside)
    }
    
    @objc private func continueAction() {
        onContinue?()
    }
    
    @objc private func cancelAction() {
        onCancel?()
    }
    
    @objc private func confirmAction() {
        onConfirm?()
    }
    
    @objc private func backgroundTapped() {
        switch popType {
        case .mixedContent:
            cancelAction()
        case .foldersOnly:
            confirmAction()
        }
    }
}

extension ShareFilesPopView: UIGestureRecognizerDelegate {
    func gestureRecognizer(_ gestureRecognizer: UIGestureRecognizer, shouldReceive touch: UITouch) -> Bool {
        return touch.view == self
    }
}
