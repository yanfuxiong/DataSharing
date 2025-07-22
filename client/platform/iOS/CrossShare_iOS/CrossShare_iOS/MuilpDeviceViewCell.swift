//
//  MuilpDeviceViewCell.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/6/17.
//

import UIKit

class MuilpDeviceViewCell: UITableViewCell {
    
    override init(style: UITableViewCell.CellStyle, reuseIdentifier: String?) {
        super.init(style: style, reuseIdentifier: reuseIdentifier)
        selectionStyle = .none
        
        self.cornerView.snp.makeConstraints { make in
            make.edges.equalToSuperview()
        }
        
        self.fileIconView.snp.makeConstraints { make in
            make.left.equalTo(17.adaptW)
            make.centerY.equalToSuperview()
            make.width.height.equalTo(52.adaptW)
        }
        
        self.deviceNameLab.snp.makeConstraints { make in
            make.left.equalTo(fileIconView.snp.right).offset(23.adaptW)
            make.top.equalTo(fileIconView).offset(5)
            make.right.lessThanOrEqualTo(-16)
        }
        
        self.deviceIpLab.snp.makeConstraints { make in
            make.left.equalTo(deviceNameLab)
            make.bottom.equalTo(fileIconView.snp.bottom).offset(-5)
            make.right.lessThanOrEqualTo(-16)
        }
    }
    
    required init?(coder: NSCoder) {
        fatalError("init(coder:) has not been implemented")
    }
    
    override func awakeFromNib() {
        super.awakeFromNib()
    }
    
    override func setSelected(_ selected: Bool, animated: Bool) {
        super.setSelected(selected, animated: animated)
        
        // Configure the view for the selected state
    }
    
    lazy var cornerView: UIView = {
        let imgView = UIView()
        imgView.backgroundColor = UIColor.white
        imgView.layer.borderWidth = 1
        imgView.layer.borderColor = UIColor.init(hex: 0xD3D3D3).cgColor
        imgView.isUserInteractionEnabled = true
        contentView.addSubview(imgView)
        return imgView
    }()
    
    lazy var fileIconView: UIImageView = {
        let imgView = UIImageView()
        imgView.isUserInteractionEnabled = true
        imgView.image = UIImage(named: "Device")
        cornerView.addSubview(imgView)
        return imgView
    }()
    
    lazy var deviceNameLab: UILabel = {
        let label = UILabel()
        label.textColor = .init(hex: 0x201B13)
        label.font = UIFont.boldSystemFont(ofSize: 15)
        cornerView.addSubview(label)
        return label
    }()
    
    lazy var deviceIpLab: UILabel = {
        let label = UILabel()
        label.textColor = .init(hex: 0xABABAB)
        label.font = UIFont.systemFont(ofSize: 12)
        cornerView.addSubview(label)
        return label
    }()
    
    //    lazy var lineView: UIView = {
    //        let imgView = UIView()
    //        imgView.backgroundColor = UIColor.init(hex: 0xD3D3D3)
    //        cornerView.addSubview(imgView)
    //        return imgView
    //    }()
    
    func configure(with model:ClientInfo) {
        self.deviceNameLab.text = model.name
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
        case SOURCE_MIRACAST:
            imageName = "miracast"
        default:
            imageName = "computer"
        }
        self.fileIconView.image = UIImage(named: imageName)
        self.deviceIpLab.text = model.ip
    }
}
