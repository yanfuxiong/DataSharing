//
//  DeviceSelectPopView.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/5/9.
//

import UIKit
import SnapKit

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
            make.height.equalTo(420)
        }
        
        cancelButton.setImage(UIImage(named: "back"), for: .normal)
        cancelButton.addTarget(self, action: #selector(cancelAction), for: .touchUpInside)
        contentView.addSubview(cancelButton)
        
        cancelButton.snp.makeConstraints { make in
            make.top.equalToSuperview().offset(20)
            make.left.equalToSuperview().offset(14)
        }
        
        crossShareLabel.text = "Cross Share"
        crossShareLabel.textAlignment = .center
        crossShareLabel.font = UIFont.boldSystemFont(ofSize: 18)
        contentView.addSubview(crossShareLabel)
        crossShareLabel.snp.makeConstraints { make in
            make.centerY.equalTo(cancelButton)
            make.centerX.equalToSuperview()
        }
        
        titleLabel.text = !fileNames.isEmpty ? fileNames.first : ""
        titleLabel.textAlignment = .center
        titleLabel.font = UIFont.systemFont(ofSize: 13)
        contentView.addSubview(titleLabel)
        titleLabel.snp.makeConstraints { make in
            make.top.equalTo(crossShareLabel.snp.bottom).offset(60)
            make.centerX.equalToSuperview()
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
    }
    
    @objc func tapAction(_ sender:UIButton) {
        selButton.isSelected = false
        sender.isSelected = true
        selButton = sender
    }
}

