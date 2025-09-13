//
//  ConnectPopView.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/8/14.
//

import UIKit

enum ConnectPopType {
    case single
    case multiple
    case none
    case connecting
    case authFailed
    case waiting
    case connectionFailed
}

class ConnectPopView: UIView {
    
    var onCancel: (() -> Void)?
    var onContinue: (() -> Void)?
    var onConfirm: (() -> Void)?
    
    private var popType: ConnectPopType
    private var title: String = "Warning"
    private var content: String
    private let contentView = UIView()
    private let titleLabel = UILabel()
    private let contentLabel = UILabel()
    private let continueButton = UIButton(type: .custom)
    private let cancelButton = UIButton(type: .custom)
    private let confirmButton = UIButton(type: .custom)
    private let loadingIndicator = UIActivityIndicatorView(style: .large)
    private let loadingImageView = UIImageView()
    
    init(frame: CGRect, type: ConnectPopType,tittle:String, content: String) {
        self.title = tittle
        self.popType = type
        self.content = content
        super.init(frame: frame)
        setupUI()
    }
    
    required init?(coder: NSCoder) {
        fatalError("init(coder:) has not been implemented")
    }
    
    func setupUI() {
        self.backgroundColor = UIColor.black.withAlphaComponent(0.4)
        
        setupContentView()
        setupLabels()
        setupButtons()
        setupLoadingUI()
    }
    
    private func setupContentView() {
        contentView.backgroundColor = .white
        contentView.layer.cornerRadius = 16
        contentView.clipsToBounds = true
        addSubview(contentView)
        
        contentView.snp.remakeConstraints { make in
            if popType == .connecting || popType == .waiting {
                make.width.equalTo(220.adaptW)
                make.height.equalTo(140.adaptH)
            } else {
                make.left.equalToSuperview().offset(30.adapt)
                make.right.equalToSuperview().offset(-30.adapt)
            }
            make.center.equalToSuperview()
        }
    }
    
    private func setupLabels() {
        titleLabel.text = title
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
        contentLabel.text = content
        contentView.addSubview(contentLabel)
        
        contentLabel.snp.remakeConstraints { make in
            if popType == .connecting || popType == .waiting {
                make.centerY.equalToSuperview().offset(20.adaptH)
                make.left.equalToSuperview().offset(20.adaptW)
                make.right.equalToSuperview().offset(-20.adaptW)
            } else {
                make.top.equalTo(titleLabel.snp.bottom).offset(10.adaptH)
                make.left.equalToSuperview().offset(10.adaptW)
                make.right.equalToSuperview().offset(-10.adaptW)
            }
        }
    }
    
    private func setupButtons() {
        switch popType {
        case .multiple:
            setupMixedContentButtons()
        case .single, .authFailed, .connectionFailed:
            setupFoldersOnlyButton()
        case .none, .connecting, .waiting:
            break
        }
    }
    
    private func setupMixedContentButtons() {
        confirmButton.setTitle("Confirm", for: .normal)
        confirmButton.setTitleColor(.white, for: .normal)
        confirmButton.backgroundColor = UIColor.systemBlue
        confirmButton.layer.cornerRadius = 8.adapt
        confirmButton.layer.borderWidth = 1
        confirmButton.layer.borderColor = UIColor.lightGray.cgColor
        confirmButton.titleLabel?.font = UIFont.systemFont(ofSize: 16.adapt)
        confirmButton.addTarget(self, action: #selector(confirmAction), for: .touchUpInside)
        contentView.addSubview(confirmButton)
        
        confirmButton.snp.makeConstraints { make in
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
        cancelButton.titleLabel?.font = UIFont.boldSystemFont(ofSize: 16.adapt)
        cancelButton.addTarget(self, action: #selector(cancelAction), for: .touchUpInside)
        contentView.addSubview(cancelButton)
        
        cancelButton.snp.makeConstraints { make in
            make.top.width.height.equalTo(confirmButton)
            make.right.equalTo(contentView.snp.centerX).offset(-10.adaptW)
        }
    }
    
    private func setupFoldersOnlyButton() {
        continueButton.setTitle("Ok", for: .normal)
        continueButton.setTitleColor(.white, for: .normal)
        continueButton.backgroundColor = UIColor.systemBlue
        continueButton.layer.cornerRadius = 8.adapt
        continueButton.titleLabel?.font = UIFont.boldSystemFont(ofSize: 16.adapt)
        continueButton.addTarget(self, action: #selector(continueAction), for: .touchUpInside)
        contentView.addSubview(continueButton)
        
        continueButton.snp.makeConstraints { make in
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
    
    func updateContent(type: ConnectPopType,tittle:String, content: String) {
        self.title = tittle
        self.popType = type
        self.content = content
        clearUI()
        setupUI()
    }
    
    private func setupLoadingUI() {
        guard popType == .connecting || popType == .waiting else { return }
        
        loadingIndicator.color = UIColor.systemBlue
        loadingIndicator.hidesWhenStopped = true
        contentView.addSubview(loadingIndicator)
        
        loadingIndicator.snp.makeConstraints { make in
            make.centerX.equalToSuperview()
            make.bottom.equalTo(contentLabel.snp.top).offset(-10.adaptH)
            make.width.height.equalTo(40.adapt)
        }
        
        startLoadingAnimation()
    }
    
    private func startLoadingAnimation() {
        guard popType == .connecting || popType == .waiting else { return }
        
        loadingIndicator.startAnimating()
    }
    
    private func stopLoadingAnimation() {
        loadingIndicator.stopAnimating()
    }
    
    private func clearUI() {
        titleLabel.removeFromSuperview()
        contentLabel.removeFromSuperview()
        continueButton.removeFromSuperview()
        cancelButton.removeFromSuperview()
        confirmButton.removeFromSuperview()
        loadingIndicator.removeFromSuperview()
        loadingImageView.removeFromSuperview()
        stopLoadingAnimation()
    }
    
    func show(in parentView: UIView) {
        parentView.addSubview(self)
        self.alpha = 0
        UIView.animate(withDuration: 0.3) {
            self.alpha = 1
        }
    }
    
    func hide(completion: (() -> Void)? = nil) {
        stopLoadingAnimation()
        UIView.animate(withDuration: 0.3, animations: {
            self.alpha = 0
        }) { _ in
            self.removeFromSuperview()
            completion?()
        }
    }
}

extension ConnectPopView: UIGestureRecognizerDelegate {
    func gestureRecognizer(_ gestureRecognizer: UIGestureRecognizer, shouldReceive touch: UITouch) -> Bool {
        return touch.view == self
    }
}
