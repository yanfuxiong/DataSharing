//
//  DefaultView.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/6/17.
//

import UIKit

class DefaultView: UIView {

    override init(frame: CGRect) {
        super.init(frame: frame)
        setupUI()
    }
    
    required init?(coder: NSCoder) {
        fatalError("init(coder:) has not been implemented")
    }
    
    private func setupUI() {
        addSubviews([waitImageView,connectedTipsImageView])
        
        waitImageView.snp.makeConstraints { make in
            make.centerX.equalToSuperview()
            make.top.equalToSuperview()
            make.size.equalTo(CGSize(width: 300, height: 300).adapt)
        }
        
        connectedTipsImageView.snp.makeConstraints { make in
            make.centerX.equalToSuperview()
            make.top.equalTo(waitImageView.snp.bottom).offset(37.adaptH)
            make.size.equalTo(CGSize(width: 267, height: 267).adapt)
        }
    }
    
    lazy var waitImageView: UIImageView = {
        let imageView = UIImageView(frame: .zero)
        imageView.image = UIImage(named: "wait_connected")
        imageView.clipsToBounds = true
        imageView.contentMode = .scaleAspectFit
        return imageView
    }()
    
    lazy var connectedTipsImageView: UIImageView = {
        let imageView = UIImageView(frame: .zero)
        imageView.image = UIImage(named: "connect_tips")
        imageView.clipsToBounds = true
        imageView.contentMode = .scaleAspectFit
        return imageView
    }()

}
