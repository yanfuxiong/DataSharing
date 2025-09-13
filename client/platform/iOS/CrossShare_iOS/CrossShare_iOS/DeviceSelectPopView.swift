//
//  DeviceSelectPopView.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/5/9.
//

import UIKit
import SnapKit

let SOURCE_HDMI1 = "HDMI1";
let SOURCE_HDMI2 = "HDMI2";
let SOURCE_USBC1 = "USBC1";
let SOURCE_USBC2 = "USBC2";
let SOURCE_DP1 = "DP1";
let SOURCE_DP2 = "DP2";
let SOURCE_MIRACAST = "Miracast";

class DeviceSelectPopView: UIView, UICollectionViewDataSource, UICollectionViewDelegateFlowLayout {
    var fileNames: [String]
    var clients: [ClientInfo]
    var onSelect: ((ClientInfo) -> Void)?
    var onCancel: (() -> Void)?
    var onSure: (() -> Void)?
    
    private let contentView = UIView()
    private var collectionView:UICollectionView
    private let titleLabel = UILabel()
    private let crossShareLabel = UILabel()
    private let cancelButton = UIButton(type: .custom)
    private let lineView = UIView()
    private let transportButton = UIButton(type: .custom)
    
    override init(frame: CGRect) {
        self.fileNames = []
        self.clients = []
        let layout = UICollectionViewFlowLayout()
        layout.scrollDirection = .horizontal
        layout.minimumLineSpacing = 5
        self.collectionView = UICollectionView(frame: .zero, collectionViewLayout: layout)
        super.init(frame: frame)
    }
    
    init(fileNames: [String], clients: [ClientInfo]) {
        self.fileNames = fileNames
        self.clients = clients
        let layout = UICollectionViewFlowLayout()
        layout.scrollDirection = .horizontal
        layout.minimumLineSpacing = 5
        self.collectionView = UICollectionView(frame: .zero, collectionViewLayout: layout)
        super.init(frame: .zero)
        setupUI()
    }
    
    required init?(coder: NSCoder) { fatalError("init(coder:) has not been implemented") }
    
    private func setupUI() {
        self.backgroundColor = UIColor.black.withAlphaComponent(0.4)
        self.layer.cornerRadius = 16
        self.clipsToBounds = true
        
        let tap = UITapGestureRecognizer(target: self, action: #selector(cancelAction))
        tap.delegate = self
        self.addGestureRecognizer(tap)
        
        contentView.backgroundColor = .white
        contentView.layer.cornerRadius = 16
        contentView.clipsToBounds = true
        addSubview(contentView)
        contentView.snp.makeConstraints { make in
            make.left.right.bottom.equalToSuperview()
            make.height.equalTo(500).priority(.low)
        }
        
        cancelButton.setImage(UIImage(named: "back"), for: .normal)
        cancelButton.addTarget(self, action: #selector(cancelAction), for: .touchUpInside)
        contentView.addSubview(cancelButton)
        
        cancelButton.snp.makeConstraints { make in
            make.top.equalToSuperview().offset(20)
            make.left.equalToSuperview().offset(14)
            make.size.equalTo(CGSize(width: 30, height: 30))
        }
        
        crossShareLabel.text = "Cross Share"
        crossShareLabel.textAlignment = .center
        crossShareLabel.font = UIFont.boldSystemFont(ofSize: 18)
        contentView.addSubview(crossShareLabel)
        crossShareLabel.snp.makeConstraints { make in
            make.centerY.equalTo(cancelButton)
            make.centerX.equalToSuperview()
        }
        
        titleLabel.attributedText = fileNameShow()
        titleLabel.numberOfLines = 0
        titleLabel.textAlignment = .center
        titleLabel.font = UIFont.systemFont(ofSize: 13)
        contentView.addSubview(titleLabel)
        titleLabel.snp.makeConstraints { make in
            make.top.equalTo(crossShareLabel.snp.bottom).offset(60)
            make.centerX.equalToSuperview()
            make.left.equalTo(cancelButton.snp.right)
        }
        
        lineView.backgroundColor = .lightGray
        contentView.addSubview(lineView)
        lineView.snp.makeConstraints { make in
            make.top.equalTo(titleLabel.snp.bottom).offset(60)
            make.centerX.equalToSuperview()
            make.left.equalTo(cancelButton)
            make.height.equalTo(1)
        }
        
        collectionView.dataSource = self
        collectionView.delegate = self
        collectionView.register(DeviceCollectionCell.self, forCellWithReuseIdentifier: "DeviceCollectionCell")
        collectionView.showsHorizontalScrollIndicator = false
        contentView.addSubview(collectionView)
        
        collectionView.snp.makeConstraints { make in
            make.top.equalTo(lineView.snp.bottom).offset(16)
            make.height.equalTo(100)
            make.left.equalTo(cancelButton)
            make.centerX.equalToSuperview()
            make.bottom.equalToSuperview().offset(-80)
        }
        
        transportButton.setTitle("Transport Files", for: .normal)
        transportButton.setTitleColor(UIColor.white, for: .normal)
        transportButton.backgroundColor = UIColor.systemBlue
        transportButton.layerCornerRadius = 5
        transportButton.addTarget(self, action: #selector(transportFiles), for: .touchUpInside)
        contentView.addSubview(transportButton)
        
        transportButton.snp.makeConstraints { make in
            make.top.equalTo(collectionView.snp.bottom).offset(25)
            make.height.equalTo(40)
            make.left.equalTo(cancelButton)
            make.centerX.equalToSuperview()
        }
    }
    
    private func fileNameShow() -> NSAttributedString {
        if fileNames.isEmpty {
            return NSAttributedString(string: "")
        }
        
        let testLabel = UILabel()
        testLabel.font = UIFont.systemFont(ofSize: 13)
        testLabel.numberOfLines = 0
        
        let screenWidth = UIScreen.main.bounds.width
        let availableWidth = screenWidth - 64
        
        var fullText = ""
        for (index, fileName) in fileNames.enumerated() {
            fullText.append(fileName)
            if index < fileNames.count - 1 {
                fullText.append("\n")
            }
        }
        
        testLabel.text = fullText
        let fullSize = testLabel.sizeThatFits(CGSize(width: availableWidth, height: CGFloat.greatestFiniteMagnitude))
        
        if fullSize.height > 320 {
            return createTruncatedTextWithStyledDots(availableWidth: availableWidth)
        } else {
            return NSAttributedString(string: fullText, attributes: [
                .font: UIFont.systemFont(ofSize: 13)
            ])
        }
    }
    
    private func createTruncatedTextWithStyledDots(availableWidth: CGFloat) -> NSAttributedString {
        let testLabel = UILabel()
        testLabel.font = UIFont.systemFont(ofSize: 13)
        testLabel.numberOfLines = 0
        
        let reservedHeight: CGFloat = 100
        let availableHeightForFiles = 320 - reservedHeight
        
        var displayText = ""
        
        for (index, fileName) in fileNames.enumerated() {
            let testText = displayText.isEmpty ? fileName : displayText + "\n" + fileName
            testLabel.text = testText
            
            let testSize = testLabel.sizeThatFits(CGSize(width: availableWidth, height: CGFloat.greatestFiniteMagnitude))
            
            if testSize.height > availableHeightForFiles {
                break
            }
            
            displayText = testText
            
            if index == fileNames.count - 1 {
                return NSAttributedString(string: displayText, attributes: [
                    .font: UIFont.systemFont(ofSize: 13)
                ])
            }
        }
        
        let attributedString = NSMutableAttributedString()
        
        if !displayText.isEmpty {
            let fileNamesAttr = NSAttributedString(string: displayText + "\n\n\n", attributes: [
                .font: UIFont.systemFont(ofSize: 13),
                .foregroundColor: UIColor.label
            ])
            attributedString.append(fileNamesAttr)
        }
        
        if let dotsImage = UIImage(named: "more") {
            let textAttachment = NSTextAttachment()
            textAttachment.image = dotsImage
            
            textAttachment.bounds = CGRect(x: 0, y: -5, width: 30, height: 30)
            
            let imageString = NSAttributedString(attachment: textAttachment)
            attributedString.append(imageString)
            attributedString.append(NSAttributedString(string: "\n"))
        }
        
        let totalAttr = NSAttributedString(string: "Total \(fileNames.count) files", attributes: [
            .font: UIFont.boldSystemFont(ofSize: 16),
            .foregroundColor: UIColor.black
        ])
        attributedString.append(totalAttr)
        
        return attributedString
    }
    
    @objc func transportFiles() {
        onSure?()
    }
    
    @objc private func cancelAction() {
        onCancel?()
    }
    
    func collectionView(_ collectionView: UICollectionView, numberOfItemsInSection section: Int) -> Int { clients.count }
    
    func collectionView(_ collectionView: UICollectionView, cellForItemAt indexPath: IndexPath) -> UICollectionViewCell {
        let cell = collectionView.dequeueReusableCell(withReuseIdentifier: "DeviceCollectionCell", for: indexPath) as! DeviceCollectionCell
        cell.contentView.backgroundColor = UIColor.white
        cell.configure(with: clients[indexPath.item])
        return cell
    }
    
    func collectionView(_ collectionView: UICollectionView, didSelectItemAt indexPath: IndexPath) {
        if let selectedIndexPaths = collectionView.indexPathsForSelectedItems {
            for selectedIndexPath in selectedIndexPaths {
                if selectedIndexPath != indexPath {
                    if let cell = collectionView.cellForItem(at: selectedIndexPath) as? DeviceCollectionCell {
                        collectionView.deselectItem(at: selectedIndexPath, animated: false)
                    }
                }
            }
        }

        onSelect?(clients[indexPath.item])
    }
    
    func collectionView(_ collectionView: UICollectionView, layout collectionViewLayout: UICollectionViewLayout, sizeForItemAt indexPath: IndexPath) -> CGSize {
        return CGSize(width: 70, height: 100)
    }
}

extension DeviceSelectPopView: UIGestureRecognizerDelegate {
    func gestureRecognizer(_ gestureRecognizer: UIGestureRecognizer, shouldReceive touch: UITouch) -> Bool {
        return touch.view == self
    }
}

class DeviceCollectionCell: UICollectionViewCell {
    
    private var selButton:UIButton = UIButton(type: .custom)
    
    override init(frame: CGRect) {
        super.init(frame: frame)
        
        icoImgView.snp.makeConstraints { make in
            make.top.equalToSuperview().offset(10)
            make.size.equalTo(CGSize(width: 56, height: 56))
            make.centerX.equalToSuperview()
        }
        
        selectedButton.snp.makeConstraints { make in
            make.bottom.equalTo(icoImgView).offset(-2)
            make.right.equalToSuperview()
            make.size.equalTo(CGSize(width: 20, height: 20))
        }
        
        fileNameLab.snp.makeConstraints { make in
            make.top.equalTo(icoImgView.snp.bottom).offset(4)
            make.centerX.equalTo(icoImgView)
            make.left.equalTo(icoImgView)
            make.height.equalTo(30)
        }
    }
    
    required init?(coder: NSCoder) {
        fatalError("init(coder:) has not been implemented")
    }
    
    override func awakeFromNib() {
        super.awakeFromNib()
        // Initialization code
    }
    
    override func layoutSubviews() {
        super.layoutSubviews()
        
    }
    
    lazy var selectedButton: UIButton = {
        let button = UIButton(type: .custom)
        button.isHidden = true
        button.setImage(UIImage(named: "unselected"), for: .normal)
        button.setImage(UIImage(named: "selected"), for: .selected)
        button.addTarget(self, action: #selector(tapAction(_:)), for: .touchUpInside)
        contentView.addSubview(button)
        return button
    }()
    
    lazy var icoImgView: UIImageView = {
        let imgView = UIImageView()
        imgView.isUserInteractionEnabled = true
        imgView.image = UIImage(named: "computer")
        contentView.addSubview(imgView)
        return imgView
    }()
    
    lazy var fileNameLab: UILabel = {
        let label = UILabel()
        label.textColor = .lightGray
        label.numberOfLines = 0
        label.font = UIFont.systemFont(ofSize: 10)
        contentView.addSubview(label)
        return label
    }()
    
    func configure(with model:ClientInfo) {
        self.fileNameLab.text = model.name
        var imageName = ""
        switch model.deviceType {
        case SOURCE_HDMI1:
            imageName = "hdmi"
        case SOURCE_HDMI2:
            imageName = "hdmi2"
        case SOURCE_USBC1:
            imageName = "usb_c1"
        case SOURCE_USBC2:
            imageName = "usb_c2"
        case SOURCE_DP1:
            imageName = "dp1"
        case SOURCE_DP2:
            imageName = "dp2"
        case SOURCE_MIRACAST:
            imageName = "miracast"
        default:
            imageName = "computer"
        }
        self.icoImgView.image = UIImage(named: imageName)
    }
    
    @objc func tapAction(_ sender:UIButton) {
        selButton.isSelected = false
        sender.isSelected = true
        selButton = sender
    }

    override var isSelected: Bool {
        didSet {
            if isSelected {
                self.icoImgView.layer.borderWidth = 3
                self.icoImgView.layer.borderColor = UIColor(red: 51/255.0, green: 51/255.0, blue: 51/255.0, alpha: 0.6).cgColor
                self.icoImgView.layer.cornerRadius = 12
                self.icoImgView.clipsToBounds = true
            } else {
                self.icoImgView.layer.borderWidth = 0
                self.icoImgView.layer.borderColor = UIColor.clear.cgColor
                self.icoImgView.layer.cornerRadius = 0
                self.icoImgView.clipsToBounds = false
            }
        }
    }

    override func prepareForReuse() {
        super.prepareForReuse()
        self.layer.borderWidth = 0
        self.layer.borderColor = UIColor.clear.cgColor
        self.layer.cornerRadius = 0
        self.clipsToBounds = false
    }
}

