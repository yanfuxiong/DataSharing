//
//  MuilpDeviceView.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/6/17.
//

import UIKit
import MBProgressHUD

class MuilpDeviceView: UIView {
    
    var mFileOpener: FileOpener? = nil
    public var dataArray:[ClientInfo] = [] {
        didSet {
            self.tableView.reloadData()
        }
    }
    
    override init(frame: CGRect) {
        super.init(frame: frame)
        setupUI()
    }
    
    required init?(coder: NSCoder) {
        fatalError("init(coder:) has not been implemented")
    }
    
    func setupUI() {
        self.addSubview(tableView)
        
        self.tableView.tableHeaderView = self.tableViewTopView
        self.tableView.tableHeaderView?.size = CGSize(width: UIScreen.main.bounds.width, height: 169.adaptH)
        
        self.tableView.setNeedsLayout()
        self.tableView.layoutIfNeeded()
        
        self.ddccidView.snp.makeConstraints { make in
            make.edges.equalToSuperview()
        }
        
        self.tableView.snp.makeConstraints { make in
            make.top.bottom.equalToSuperview()
            make.left.equalToSuperview()
            make.centerX.equalToSuperview()
        }
    }
    
    lazy var tableView: UITableView = {
        let view = UITableView()
        view.backgroundColor = .clear
        view.delegate = self
        view.dataSource = self
        view.showsVerticalScrollIndicator = false
        view.rowHeight = 80
        view.separatorStyle = .none
        view.register(MuilpDeviceViewCell.self,
                      forCellReuseIdentifier: NSStringFromClass(MuilpDeviceViewCell.self))
        return view
    }()
    
    lazy var tableViewTopView: UIView = {
        let imgView = UIView()
        imgView.clipsToBounds = true
        imgView.backgroundColor = .white
        return imgView
    }()
    
    lazy var ddccidView: DDccidView = {
        let view = DDccidView(frame: .zero)
        view.backgroundColor = UIColor.clear
        tableViewTopView.addSubview(view)
        return view
    }()
}

extension MuilpDeviceView {
    func refreshUI(_ name:String) {
        self.ddccidView.updateDDccidView(with: name, ipAddress: P2PManager.shared.deviceDiasId)
    }
}

extension MuilpDeviceView:UITableViewDelegate,UITableViewDataSource {
    func tableView(_ tableView: UITableView, numberOfRowsInSection section: Int) -> Int {
        return dataArray.count
    }
    
    func tableView(_ tableView: UITableView, cellForRowAt indexPath: IndexPath) -> UITableViewCell {
        guard let cell = tableView.dequeueReusableCell(withIdentifier: NSStringFromClass(MuilpDeviceViewCell.self)) as? MuilpDeviceViewCell else {
            return UITableViewCell()
        }
        let model = dataArray[indexPath.row]
        cell.configure(with: model)
        return cell
    }
}

class DDccidView: UIView {
    
    override init(frame: CGRect) {
        super.init(frame: frame)
        setupUI()
    }
    
    required init?(coder: NSCoder) {
        fatalError("init(coder:) has not been implemented")
    }
    
    private func setupUI() {
        addSubviews([waitImageView,DDccidNameLable,lineView])
        
        waitImageView.snp.makeConstraints { make in
            make.centerX.equalToSuperview()
            make.top.equalToSuperview().offset(34.adaptH)
            make.size.equalTo(CGSize(width: 74, height: 74).adapt)
        }
        
        DDccidNameLable.snp.makeConstraints { make in
            make.centerX.equalToSuperview()
            make.top.equalTo(waitImageView.snp.bottom).offset(23.adaptH)
        }
        
        lineView.snp.makeConstraints { make in
            make.left.right.bottom.equalToSuperview()
            make.height.equalTo(3.adaptH)
        }
    }
    
    lazy var waitImageView: UIImageView = {
        let imageView = UIImageView(frame: .zero)
        imageView.image = UIImage(named: "Ddccid")
        imageView.clipsToBounds = true
        imageView.contentMode = .scaleAspectFit
        return imageView
    }()
    
    lazy var DDccidNameLable: UILabel = {
        let text = UILabel(frame: .zero)
        text.textColor = UIColor.init(hex: 0x201B13)
        text.font = UIFont.systemFont(ofSize: 12)
        text.text = "My ip & Device Name:"
        addSubview(text)
        return text
    }()
    
    lazy var lineView: UIView = {
        let imgView = UIView()
        imgView.backgroundColor = UIColor.init(hex: 0xD3D3D3)
        addSubview(imgView)
        return imgView
    }()
    
}

extension DDccidView {
    func updateDDccidView(with deviceName: String, ipAddress: String) {
        self.DDccidNameLable.text = deviceName
    }
}

