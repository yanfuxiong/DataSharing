//
//  LicenseView.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/6/20.
//

import UIKit

class LicenseView: UIView {
    
    var didSelectBlock:((licenseModel) -> ())?
    
    public var dataArray:[licenseModel] = [] {
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
        self.tableView.tableHeaderView?.size = CGSize(width: UIScreen.main.bounds.width, height: 70.adaptH)
        
        self.tableView.setNeedsLayout()
        self.tableView.layoutIfNeeded()
        
        self.licenceLable.snp.makeConstraints { make in
            make.centerY.equalToSuperview()
            make.left.equalToSuperview().offset(16)
        }
        
        lineView.snp.makeConstraints { make in
            make.left.right.equalToSuperview()
            make.bottom.equalToSuperview()
            make.height.equalTo(1)
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
        view.rowHeight = 42.adaptH
        view.separatorStyle = .none
        view.register(LicenseViewCell.self,
                      forCellReuseIdentifier: NSStringFromClass(LicenseViewCell.self))
        return view
    }()
    
    lazy var tableViewTopView: UIView = {
        let imgView = UIView()
        imgView.clipsToBounds = true
        imgView.backgroundColor = .white
        return imgView
    }()
    
    lazy var licenceLable: UILabel = {
        let text = UILabel(frame: .zero)
        text.textColor = UIColor.init(hex: 0xC6BDBD)
        text.font = UIFont.boldSystemFont(ofSize: 12)
        text.text = "Licenses"
        tableViewTopView.addSubview(text)
        return text
    }()
    
    lazy var lineView: UIView = {
        let imgView = UIView()
        imgView.backgroundColor = UIColor.init(hex: 0xD7D7D7)
        tableViewTopView.addSubview(imgView)
        return imgView
    }()
}

extension LicenseView:UITableViewDelegate,UITableViewDataSource {
    func tableView(_ tableView: UITableView, numberOfRowsInSection section: Int) -> Int {
        return dataArray.count
    }
    
    func tableView(_ tableView: UITableView, cellForRowAt indexPath: IndexPath) -> UITableViewCell {
        guard let cell = tableView.dequeueReusableCell(withIdentifier: NSStringFromClass(LicenseViewCell.self)) as? LicenseViewCell else {
            return UITableViewCell()
        }
        let model = dataArray[indexPath.row]
        cell.configure(with: model)
        return cell
    }
    
    func tableView(_ tableView: UITableView, didSelectRowAt indexPath: IndexPath) {
        tableView.deselectRow(at: indexPath, animated: true)
        let model = dataArray[indexPath.row]
        self.didSelectBlock?(model)
    }
}

class LicenseViewCell: UITableViewCell {
    
    override init(style: UITableViewCell.CellStyle, reuseIdentifier: String?) {
        super.init(style: style, reuseIdentifier: reuseIdentifier)
        selectionStyle = .none
        
        self.cornerView.snp.makeConstraints { make in
            make.edges.equalToSuperview()
        }
        
        self.deviceNameLab.snp.makeConstraints { make in
            make.left.equalToSuperview().offset(14)
            make.centerY.equalToSuperview()
        }
        
        self.fileIconView.snp.makeConstraints { make in
            make.right.equalToSuperview().offset(-9)
            make.centerY.equalToSuperview()
            make.width.height.equalTo(24.adaptW)
        }
        
        self.lineView.snp.makeConstraints { make in
            make.left.right.equalToSuperview()
            make.bottom.equalToSuperview()
            make.height.equalTo(1)
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
        //        imgView.layer.borderWidth = 1
        //        imgView.layer.borderColor = UIColor.init(hex: 0xD3D3D3).cgColor
        imgView.isUserInteractionEnabled = true
        contentView.addSubview(imgView)
        return imgView
    }()
    
    lazy var fileIconView: UIImageView = {
        let imgView = UIImageView()
        imgView.isUserInteractionEnabled = true
        imgView.image = UIImage(named: "keyboard_arrow_down")
        cornerView.addSubview(imgView)
        return imgView
    }()
    
    lazy var deviceNameLab: UILabel = {
        let label = UILabel()
        label.textColor = .init(hex: 0x201B13)
        label.font = UIFont.systemFont(ofSize: 13)
        cornerView.addSubview(label)
        return label
    }()
    
    lazy var lineView: UIView = {
        let imgView = UIView()
        imgView.backgroundColor = UIColor.init(hex: 0xD7D7D7)
        cornerView.addSubview(imgView)
        return imgView
    }()
    
    func configure(with model:licenseModel) {
        self.deviceNameLab.text = model.tittle
    }
}
