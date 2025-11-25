//
//  DownloadViewCell.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/5/8.
//

import UIKit

class DownloadViewCell: UITableViewCell {
    
    var deleteBlock:(() -> ())?
    var openBlock:(() -> ())?
    var cancelBlock:(() -> ())?
    var refreshBlock:(() -> ())?
    
    override init(style: UITableViewCell.CellStyle, reuseIdentifier: String?) {
        super.init(style: style, reuseIdentifier: reuseIdentifier)
        selectionStyle = .none
        contentView.backgroundColor = .init(hex: 0xF0F0F0)
        
        cornerView.snp.makeConstraints { make in
            make.edges.equalToSuperview().inset(UIEdgeInsets(top: 2, left: 4, bottom: 2, right: 4))
        }
        
        fileIconView.snp.makeConstraints { make in
            make.top.equalToSuperview().offset(18)
            make.left.equalToSuperview().offset(12)
            make.size.equalTo(CGSize(width: 52, height: 52))
        }
        
        totalLab.snp.makeConstraints { make in
            make.top.equalToSuperview().offset(12)
            make.height.equalTo(14)
            make.left.equalToSuperview().offset(72)
        }
        
        fileNameLab.snp.makeConstraints { make in
            make.left.equalTo(totalLab)
            make.height.equalTo(14)
            make.top.equalToSuperview().offset(24)
        }
        
        fileSizeLab.snp.makeConstraints { make in
            make.left.equalTo(totalLab)
            make.height.equalTo(14)
            make.top.equalTo(fileNameLab.snp.bottom).offset(10)
        }
        
        deleteImgView.snp.makeConstraints { make in
            make.right.equalToSuperview().offset(-12)
            make.centerY.equalTo(fileSizeLab)
            make.size.equalTo(CGSize(width: 20, height: 20))
        }
        
        cancelButton.snp.makeConstraints { make in
            make.right.equalToSuperview().offset(-12)
            make.centerY.equalTo(fileSizeLab)
            make.size.equalTo(CGSize(width: 44, height: 44))
        }
        
        cancelImgView.snp.makeConstraints { make in
            make.center.equalTo(cancelButton)
            make.size.equalTo(CGSize(width: 20, height: 20))
        }
        
        refreshImgView.snp.makeConstraints { make in
            make.right.equalToSuperview().offset(-12)
            make.centerY.equalTo(fileSizeLab)
            make.size.equalTo(CGSize(width: 20, height: 20))
        }
        
        openFileView.snp.makeConstraints { make in
            make.right.equalTo(deleteImgView.snp.left).offset(-10)
            make.centerY.equalTo(fileSizeLab)
            make.size.equalTo(CGSize(width: 20, height: 20))
        }
        
        progressView.snp.makeConstraints { make in
            make.left.equalToSuperview().offset(10)
            make.bottom.equalToSuperview().offset(-30)
            make.centerX.equalToSuperview()
            make.height.equalTo(4)
        }
        
        finishLab.snp.makeConstraints { make in
            make.left.equalTo(progressView)
            make.top.equalTo(progressView.snp.bottom).offset(8)
        }
        
        fromClientLab.snp.makeConstraints { make in
            make.left.equalTo(finishLab.snp.right).offset(3)
            make.right.equalTo(fileRatioLab.snp.left).offset(-3)
            make.top.equalTo(progressView.snp.bottom).offset(8)
        }
        
        fileRatioLab.snp.makeConstraints { make in
            make.right.equalToSuperview().offset(-8)
            make.top.equalTo(progressView.snp.bottom).offset(8)
        }
    }
    
    required init?(coder: NSCoder) {
        fatalError("init(coder:) has not been implemented")
    }
    
    override func awakeFromNib() {
        super.awakeFromNib()
        // Initialization code
    }
    
    override func setSelected(_ selected: Bool, animated: Bool) {
        super.setSelected(selected, animated: animated)
        
        // Configure the view for the selected state
    }
    
    override func layoutSubviews() {
        super.layoutSubviews()
        
    }
    
    lazy var cornerView: UIView = {
        let imgView = UIView()
        imgView.backgroundColor = UIColor.white
        imgView.layerCornerRadius = 7
        imgView.layer.shadowColor = UIColor.black.cgColor
        imgView.layer.shadowOpacity = 0.1
        imgView.layer.shadowOffset = CGSize(width: 0, height: 2)
        imgView.layer.masksToBounds = false
        imgView.isUserInteractionEnabled = true
        contentView.addSubview(imgView)
        return imgView
    }()
    
    lazy var deleteImgView: UIImageView = {
        let imgView = UIImageView()
        imgView.isHidden = true
        imgView.isUserInteractionEnabled = true
        imgView.image = UIImage(named: "garbage")
        cornerView.addSubview(imgView)
        imgView.addGestureRecognizer(UITapGestureRecognizer(target: self, action: #selector(deleteAction)))
        return imgView
    }()
    
    lazy var cancelButton: UIButton = {
        let button = UIButton(type: .custom)
        button.backgroundColor = .clear
        button.addTarget(self, action: #selector(cancelAction), for: .touchUpInside)
        cornerView.addSubview(button)
        return button
    }()
    
    lazy var cancelImgView: UIImageView = {
        let imgView = UIImageView()
//        imgView.isHidden = true
        imgView.isUserInteractionEnabled = false
        imgView.image = UIImage(named: "cancel")
        cancelButton.addSubview(imgView)
        return imgView
    }()
    
    lazy var refreshImgView: UIImageView = {
        let imgView = UIImageView()
        imgView.isHidden = true
        imgView.isUserInteractionEnabled = true
        imgView.image = UIImage(named: "failure")
        cornerView.addSubview(imgView)
        imgView.addGestureRecognizer(UITapGestureRecognizer(target: self, action: #selector(refreshAction)))
        return imgView
    }()
    
    lazy var openFileView: UIImageView = {
        let imgView = UIImageView()
        imgView.isUserInteractionEnabled = true
        imgView.image = UIImage(named: "open")
        cornerView.addSubview(imgView)
        imgView.addGestureRecognizer(UITapGestureRecognizer(target: self, action: #selector(openAction)))
        return imgView
    }()
    
    lazy var fileIconView: UIImageView = {
        let imgView = UIImageView()
        imgView.isUserInteractionEnabled = true
        imgView.image = UIImage(named: "Device")
        cornerView.addSubview(imgView)
        return imgView
    }()
    
    lazy var fileNameLab: UILabel = {
        let label = UILabel()
        label.textColor = .init(hex: 0x1C2D41)
        label.font = UIFont.boldSystemFont(ofSize: 12)
        cornerView.addSubview(label)
        return label
    }()
    
    lazy var totalLab: UILabel = {
        let label = UILabel()
        label.textColor = .init(hex: 0x1C2D41)
        label.font = UIFont.boldSystemFont(ofSize: 12)
        cornerView.addSubview(label)
        return label
    }()
    
    lazy var fileSizeLab: UILabel = {
        let label = UILabel()
        label.textColor = .init(hex: 0xA4ABB3)
        label.font = UIFont.boldSystemFont(ofSize: 12)
        cornerView.addSubview(label)
        return label
    }()
    
    lazy var fileRatioLab: UILabel = {
        let label = UILabel()
        label.textColor = .init(hex: 0xA4ABB3)
        label.font = UIFont.boldSystemFont(ofSize: 10)
        label.setContentHuggingPriority(.required, for: .horizontal)
        label.setContentCompressionResistancePriority(.required, for: .horizontal)
        cornerView.addSubview(label)
        return label
    }()
    
    lazy var fromClientLab: UILabel = {
        let label = UILabel()
        label.textColor = .init(hex: 0x77818D)
        label.font = UIFont.systemFont(ofSize: 10)
        label.textAlignment = .center
        label.lineBreakMode = .byTruncatingTail
        label.numberOfLines = 1
        cornerView.addSubview(label)
        return label
    }()
    
    lazy var finishLab: UILabel = {
        let label = UILabel()
        label.textColor = .init(hex: 0x77818D)
        label.font = UIFont.systemFont(ofSize: 10)
        label.setContentHuggingPriority(.required, for: .horizontal)
        label.setContentCompressionResistancePriority(.required, for: .horizontal)
        cornerView.addSubview(label)
        return label
    }()
    
    lazy var progressView: UIProgressView = {
        let view = UIProgressView.init(progressViewStyle: .default)
        view.progressTintColor = .systemGreen
        view.trackTintColor = .systemGray4
        view.progress = 0.0
        cornerView.addSubview(view)
        return view
    }()
    
    func refreshUI(with model:DownloadItem) {
        if model.isMutip {
            if let nameArray = model.currentfileName?.components(separatedBy: "/") as? [String],nameArray.count > 1 {
                let componentsPath = nameArray.last
                self.fileNameLab.text = componentsPath
                self.fileNameLab.snp.remakeConstraints { make in
                    make.left.equalTo(totalLab)
                    make.right.equalToSuperview().offset(-8)
                    make.height.equalTo(14)
                    make.centerY.equalTo(fileIconView)
                }
            } else {
                self.fileNameLab.text = model.currentfileName
            }
        } else {
            fileNameLab.snp.remakeConstraints { make in
                make.left.equalTo(totalLab)
                make.height.equalTo(14)
                make.top.equalToSuperview().offset(24)
            }
            self.fileNameLab.text = model.currentfileName
        }
        self.totalLab.isHidden = !model.isMutip
        if let receiveSize = model.receiveSize,let totalSize = model.totalSize {
            if receiveSize < totalSize {
                self.cancelButton.isHidden = false
                self.refreshImgView.isHidden = true
                self.deleteImgView.isHidden = true
            } else {
                self.cancelButton.isHidden = true
                self.refreshImgView.isHidden = true
                self.deleteImgView.isHidden = false
            }
            if model.error != nil {
                self.cancelButton.isHidden = true
                self.refreshImgView.isHidden = false
                self.deleteImgView.isHidden = true
                
                self.finishLab.text = convertDateFormat(timeStamp: Date.now.timeIntervalSince1970)
                if totalSize > 0 {
                    self.fileRatioLab.text = "Failed \(Float(Double(receiveSize) / Double(totalSize)))%"
                    progressView.setProgress(Float(Double(receiveSize) / Double(totalSize)), animated: true)
                    progressView.progressTintColor = .systemRed
                    self.fileRatioLab.textColor = .systemRed
                }
            }
            // TODO: retry API not ready
            self.refreshImgView.isHidden = true
            self.openFileView.isHidden = receiveSize < totalSize
            self.finishLab.isHidden = receiveSize < totalSize
            self.fileRatioLab.text = ((totalSize == 0) ? "0%" : "\(Float(receiveSize * 100 / totalSize))%")
            self.fileSizeLab.text = calculateFile(with: receiveSize)
            if receiveSize == totalSize {
                self.finishLab.text = convertDateFormat(timeStamp: model.finishTime ?? Date.now.timeIntervalSince1970)
                self.progressView.progress = 1.0
                self.fileRatioLab.textColor = UIColor.init(hex: 0x25D366)
                self.fileRatioLab.text = "Complete 100%"
            } else {
                if totalSize > 0 {
                    progressView.setProgress(Float(Double(receiveSize) / Double(totalSize)), animated: true)
                    self.fileRatioLab.textColor = .init(hex: 0x2574ED)
                }
            }
            if let recvFileCnt = model.recvFileCnt,let totalFileCnt = model.totalFileCnt {
                self.totalLab.text = "Total:\(recvFileCnt)/\(totalFileCnt) files(\(calculateFile(with: receiveSize))/\(calculateFile(with: totalSize)))"
            }
        }
        self.fromClientLab.isHidden = !model.isMutip
        if let deviceName = model.deviceName {
            self.fromClientLab.text = "From \(deviceName)"
        }
    }
    
    public func calculateFile(with total:UInt64) -> String {
        if total < 1024 * 1024 {
            let progress = Float(Double(total) / 1024.0)
            return String(format: "%.2fK", progress)
        } else if total < 1024 * 1024 * 1024 {
            let progress = Float(Double(total) / 1024.0 / 1024.0)
            return String(format: "%.2fM", progress)
        } else {
            let progress = Float(Double(total) / 1024.0 / 1024.0 / 1024.0)
            return String(format: "%.2fG", progress)
        }
    }
    
    @objc func cancelAction() {
        if let cancelBlock = self.cancelBlock {
            cancelBlock()
        }
    }
    
    @objc func refreshAction() {
        if let refreshBlock = self.refreshBlock {
            refreshBlock()
        }
    }
    
    @objc func deleteAction() {
        if let deleteBlock = self.deleteBlock {
            deleteBlock()
        }
    }
    
    @objc func openAction() {
        if let openBlock = self.openBlock {
            openBlock()
        }
    }
    
    func convertDateFormat(timeStamp: TimeInterval) -> String? {
        let dateFormatter = DateFormatter()
        dateFormatter.dateFormat = "yyyy.MM.dd HH:mm:ss"
        let date = dateFormatter.string(from: Date(timeIntervalSince1970: timeStamp))
        return date
    }
}
