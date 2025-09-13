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
        setupUI()
    }
    
    required init?(coder: NSCoder) {
        fatalError("init(coder:) has not been implemented")
    }
    
    func setupUI() {
        self.backgroundColor = UIColor.black.withAlphaComponent(0.4)
        
        let tap = UITapGestureRecognizer(target: self, action: #selector(backgroundTapped))
        tap.delegate = self
        self.addGestureRecognizer(tap)
        
        setupContentView()
        setupLabels()
        setupButtons()
    }
    
    private func setupContentView() {
        contentView.backgroundColor = .white
        contentView.layer.cornerRadius = 16
        contentView.clipsToBounds = true
        addSubview(contentView)
        
        contentView.snp.makeConstraints { make in
            make.left.equalToSuperview().offset(30.adapt)
            make.right.equalToSuperview().offset(-30.adapt)
            make.center.equalToSuperview()
        }
    }
    
    private func setupLabels() {
        titleLabel.text = "Notice"
        titleLabel.numberOfLines = 0
        titleLabel.textAlignment = .center
        titleLabel.font = UIFont.boldSystemFont(ofSize: 18.adapt)
        titleLabel.textColor = .black
        contentView.addSubview(titleLabel)
        
        titleLabel.snp.makeConstraints { make in
            make.top.equalToSuperview().offset(20.adaptH)
            make.left.equalToSuperview().offset(20.adaptW)
            make.right.equalToSuperview().offset(-20.adaptW)
        }
        
        contentLabel.numberOfLines = 0
        contentLabel.textAlignment = .center
        contentLabel.font = UIFont.systemFont(ofSize: 14.adapt)
        contentLabel.textColor = .darkGray
        contentView.addSubview(contentLabel)
        
        switch popType {
        case .mixedContent:
            contentLabel.text = "Folders will not be transferred.\nOnly the selected files will be sent."
        case .foldersOnly:
            contentLabel.text = "Please select files only.\nFolder transfer is not supported."
        }
        
        contentLabel.snp.makeConstraints { make in
            make.top.equalTo(titleLabel.snp.bottom).offset(15.adaptH)
            make.left.equalToSuperview().offset(20.adaptW)
            make.right.equalToSuperview().offset(-20.adaptW)
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
        continueButton.setTitle("Continue", for: .normal)
        continueButton.setTitleColor(.white, for: .normal)
        continueButton.backgroundColor = UIColor.systemBlue
        continueButton.layer.cornerRadius = 8.adapt
        continueButton.titleLabel?.font = UIFont.boldSystemFont(ofSize: 16.adapt)
        continueButton.addTarget(self, action: #selector(continueAction), for: .touchUpInside)
        contentView.addSubview(continueButton)
        
        continueButton.snp.makeConstraints { make in
            make.top.equalTo(contentLabel.snp.bottom).offset(25.adaptH)
            make.right.equalToSuperview().offset(-20.adaptW)
            make.left.equalTo(contentView.snp.centerX).offset(10.adaptW)
            make.height.equalTo(44.adaptH)
            make.bottom.equalToSuperview().offset(-20.adaptH)
        }
        
        cancelButton.setTitle("Cancel", for: .normal)
        cancelButton.setTitleColor(.black, for: .normal)
        cancelButton.backgroundColor = UIColor.white
        cancelButton.layer.cornerRadius = 8.adapt
        cancelButton.layer.borderWidth = 1
        cancelButton.layer.borderColor = UIColor.lightGray.cgColor
        cancelButton.titleLabel?.font = UIFont.systemFont(ofSize: 16.adapt)
        cancelButton.addTarget(self, action: #selector(cancelAction), for: .touchUpInside)
        contentView.addSubview(cancelButton)
        
        cancelButton.snp.makeConstraints { make in
            make.top.width.height.equalTo(continueButton)
            make.right.equalTo(contentView.snp.centerX).offset(-10.adaptW)
        }
    }
    
    private func setupFoldersOnlyButton() {
        confirmButton.setTitle("Confirm", for: .normal)
        confirmButton.setTitleColor(.white, for: .normal)
        confirmButton.backgroundColor = UIColor.systemBlue
        confirmButton.layer.cornerRadius = 8.adapt
        confirmButton.titleLabel?.font = UIFont.boldSystemFont(ofSize: 16.adapt)
        confirmButton.addTarget(self, action: #selector(confirmAction), for: .touchUpInside)
        contentView.addSubview(confirmButton)
        
        confirmButton.snp.makeConstraints { make in
            make.top.equalTo(contentLabel.snp.bottom).offset(25.adaptH)
            make.left.equalToSuperview().offset(20.adaptW)
            make.right.equalToSuperview().offset(-20.adaptW)
            make.height.equalTo(44.adaptH)
            make.bottom.equalToSuperview().offset(-20.adaptH)
        }
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
