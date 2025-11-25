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
    private let clientIndexBmView = UIView()
    private let clientIndexInfoView = UIView()
    
    private let colorIdxEncode0 = UIColor(red: 243/255.0, green: 255/255.0, blue: 248/255.0, alpha: 1.0)
    private let colorIdxEncode1 = UIColor(red: 250/255.0, green: 245/255.0, blue: 255/255.0, alpha: 1.0)
    
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
        contentView.backgroundColor = UIColor(red: 250/255.0, green: 250/255.0, blue: 250/255.0, alpha: 1.0)
        contentView.layer.cornerRadius = 25
        contentView.clipsToBounds = true
        addSubview(contentView)
        
        contentView.snp.remakeConstraints { make in
            if popType == .connecting || popType == .waiting {
                make.width.equalTo(220)
                make.height.equalTo(140)
            } else {
                make.left.equalToSuperview().offset(30)
                make.right.equalToSuperview().offset(-30)
            }
            make.center.equalToSuperview()
        }
    }
    
    private func pixel2Point(_ pixel: Int) -> CGFloat {
        return CGFloat(pixel) / UIScreen.main.scale
    }
    
    private func setupClientIndexUI(clientIndex: UInt32) {
        guard clientIndex > 0, popType == .waiting else {
            return
        }
        setupClientIndexBenchmarkView()
        setupClientIndexInfoView(index: UInt8(clientIndex))
    }
    private func setupClientIndexBenchmarkView() {
        let posLeftShift = pixel2Point(-30)
        let width = pixel2Point(10)
        let height = pixel2Point(20)
        let padding = pixel2Point(10)

        clientIndexBmView.backgroundColor = .clear
        clientIndexBmView.isOpaque = false
        addSubview(clientIndexBmView)

        clientIndexBmView.snp.makeConstraints { make in
            make.width.equalTo(width*2 + padding*1)
            make.height.equalTo(height)
            make.right.equalTo(self.snp.centerX).offset(posLeftShift)
            make.centerY.equalToSuperview()
        }

        let view_encode0 = UIView()
        view_encode0.backgroundColor = colorIdxEncode0
        clientIndexBmView.addSubview(view_encode0)
        view_encode0.snp.makeConstraints { make in
            make.width.equalTo(width)
            make.height.equalTo(height)
            make.right.equalToSuperview()
            make.centerY.equalToSuperview()
        }
        
        let view_encode1 = UIView()
        view_encode1.backgroundColor = colorIdxEncode1
        clientIndexBmView.addSubview(view_encode1)
        view_encode1.snp.makeConstraints { make in
            make.width.equalTo(width)
            make.height.equalTo(height)
            make.left.equalToSuperview()
            make.centerY.equalToSuperview()
        }
    }

    private func setupClientIndexInfoView(index: UInt8) {
        let posRightShift = pixel2Point(30)
        let width = pixel2Point(10)
        let height = pixel2Point(20)
        let padding = pixel2Point(10)

        clientIndexInfoView.backgroundColor = .clear
        clientIndexInfoView.isOpaque = false
        addSubview(clientIndexInfoView)

        clientIndexInfoView.snp.makeConstraints { make in
            make.width.equalTo(width*8 + padding*7)
            make.height.equalTo(height)
            make.left.equalTo(self.snp.centerX).offset(posRightShift)
            make.centerY.equalToSuperview()
        }

        let maxBit = 7
        for i in stride(from: maxBit, through: 0, by: -1) {
            let bit = (index >> i) & 1
            
            let view_encode = UIView()
            if bit == 0 {
                view_encode.backgroundColor = colorIdxEncode0
            } else {
                view_encode.backgroundColor = colorIdxEncode1
            }
            
            clientIndexInfoView.addSubview(view_encode)
            let posX: CGFloat = CGFloat(maxBit-i)*width + CGFloat(maxBit-i)*padding
            view_encode.snp.remakeConstraints { make in
                make.width.equalTo(width)
                make.height.equalTo(height)
                make.left.equalToSuperview().offset(posX)
                make.centerY.equalToSuperview()
            }
        }
    }
    
    private func setupLabels() {
        titleLabel.text = title
        titleLabel.numberOfLines = 0
        titleLabel.textAlignment = .center
        titleLabel.font = UIFont.boldSystemFont(ofSize: 18)
        titleLabel.textColor = .black
        contentView.addSubview(titleLabel)
        
        titleLabel.snp.makeConstraints { make in
            make.top.equalToSuperview().offset(20)
            make.left.equalToSuperview().offset(20)
            make.right.equalToSuperview().offset(-20)
        }
        
        contentLabel.numberOfLines = 0
        contentLabel.textAlignment = .center
        contentLabel.font = UIFont.systemFont(ofSize: 14)
        contentLabel.textColor = .darkGray
        contentLabel.text = content
        contentView.addSubview(contentLabel)
        
        contentLabel.snp.remakeConstraints { make in
            if popType == .connecting || popType == .waiting {
                make.centerY.equalToSuperview().offset(20)
                make.left.equalToSuperview().offset(20)
                make.right.equalToSuperview().offset(-20)
            } else {
                make.top.equalTo(titleLabel.snp.bottom).offset(10)
                make.left.equalToSuperview().offset(10)
                make.right.equalToSuperview().offset(-10)
            }
        }
    }
    
    private func setupButtons() {
        switch popType {
        case .multiple:
            setupMixedContentButtons()
        case .single, .authFailed, .connectionFailed:
            setupOkOnlyButton()
        case .none, .connecting, .waiting:
            break
        }
    }
    
    private func setupMixedContentButtons() {
        confirmButton.setTitle("Confirm", for: .normal)
        confirmButton.setTitleColor(.white, for: .normal)
        confirmButton.backgroundColor = UIColor.systemBlue
        confirmButton.layer.cornerRadius = kBtnHeight / 2
        confirmButton.titleLabel?.font = UIFont.boldSystemFont(ofSize: 16)
        confirmButton.addTarget(self, action: #selector(confirmAction), for: .touchUpInside)
        contentView.addSubview(confirmButton)
        
        confirmButton.snp.makeConstraints { make in
            make.top.equalTo(contentLabel.snp.bottom).offset(30)
            make.right.equalToSuperview().offset(-10)
            make.left.equalTo(contentView.snp.centerX).offset(5)
            make.height.equalTo(kBtnHeight)
            make.bottom.equalToSuperview().offset(-15)
        }
        
        cancelButton.setTitle("Cancel", for: .normal)
        cancelButton.setTitleColor(.systemBlue, for: .normal)
        cancelButton.backgroundColor = UIColor.white
        cancelButton.backgroundColor = UIColor(red: 240/255.0, green: 240/255.0, blue: 240/255.0, alpha: 1.0)
        cancelButton.layer.cornerRadius = kBtnHeight / 2
        cancelButton.titleLabel?.font = UIFont.boldSystemFont(ofSize: 16)
        cancelButton.addTarget(self, action: #selector(cancelAction), for: .touchUpInside)
        contentView.addSubview(cancelButton)
        
        cancelButton.snp.makeConstraints { make in
            make.top.width.height.equalTo(confirmButton)
            make.right.equalTo(contentView.snp.centerX).offset(-5)
        }
    }
    
    private func setupOkOnlyButton() {
        continueButton.setTitle("Ok", for: .normal)
        continueButton.setTitleColor(.white, for: .normal)
        continueButton.backgroundColor = UIColor.systemBlue
        continueButton.layer.cornerRadius = kBtnHeight / 2
        continueButton.titleLabel?.font = UIFont.boldSystemFont(ofSize: 16)
        continueButton.addTarget(self, action: #selector(continueAction), for: .touchUpInside)
        contentView.addSubview(continueButton)
        
        continueButton.snp.makeConstraints { make in
            make.top.equalTo(contentLabel.snp.bottom).offset(30)
            make.left.equalToSuperview().offset(20)
            make.right.equalToSuperview().offset(-20)
            make.height.equalTo(kBtnHeight)
            make.bottom.equalToSuperview().offset(-15)
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

    func updateClientIndex(clientIndex: UInt32) {
        setupClientIndexUI(clientIndex: clientIndex)
    }
    private func setupLoadingUI() {
        guard popType == .connecting || popType == .waiting else { return }
        
        loadingIndicator.color = UIColor.systemBlue
        loadingIndicator.hidesWhenStopped = true
        contentView.addSubview(loadingIndicator)
        
        loadingIndicator.snp.makeConstraints { make in
            make.centerX.equalToSuperview()
            make.bottom.equalTo(contentLabel.snp.top).offset(-10)
            make.width.height.equalTo(40)
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
        clientIndexBmView.removeFromSuperview()
        clientIndexInfoView.removeFromSuperview()
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
