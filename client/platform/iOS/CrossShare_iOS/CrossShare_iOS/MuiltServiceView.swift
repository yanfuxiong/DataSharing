//
//  MuiltServiceView.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/8/13.
//

import UIKit
import MBProgressHUD

class MuiltServiceView: UIView {
    
    var onSelectService:((LanServiceInfo) -> Void)?
    
    public var dataArray:[LanServiceInfo] = [] {
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
        view.register(MuilpServiceViewCell.self,
                      forCellReuseIdentifier: NSStringFromClass(MuilpServiceViewCell.self))
        return view
    }()
    
    lazy var tableViewTopView: UIView = {
        let imgView = UIView()
        imgView.clipsToBounds = true
        imgView.backgroundColor = .white
        return imgView
    }()
    
    lazy var ddccidView: SerchServiceView = {
        let view = SerchServiceView(frame: .zero)
        view.backgroundColor = UIColor.clear
        tableViewTopView.addSubview(view)
        return view
    }()
}

extension MuiltServiceView:UITableViewDelegate,UITableViewDataSource {
    func tableView(_ tableView: UITableView, numberOfRowsInSection section: Int) -> Int {
        return dataArray.count
    }
    
    func tableView(_ tableView: UITableView, cellForRowAt indexPath: IndexPath) -> UITableViewCell {
        guard let cell = tableView.dequeueReusableCell(withIdentifier: NSStringFromClass(MuilpServiceViewCell.self)) as? MuilpServiceViewCell else {
            return UITableViewCell()
        }
        let model = dataArray[indexPath.row]
        cell.configure(with: model)
        return cell
    }
    
    func tableView(_ tableView: UITableView, didSelectRowAt indexPath: IndexPath) {
        tableView.deselectRow(at: indexPath, animated: true)
        let model = dataArray[indexPath.row]
        if let onSelectService = self.onSelectService {
            onSelectService(model)
        }
    }
}

class SerchServiceView: UIView {
    
    private var animationTimer: Timer?
    private var currentImageIndex = 0
    private let totalImages = 12
    
    override init(frame: CGRect) {
        super.init(frame: frame)
        setupUI()
        startImageAnimation()
    }
    
    required init?(coder: NSCoder) {
        fatalError("init(coder:) has not been implemented")
    }
    
    deinit {
        stopImageAnimation()
    }
    
    private func setupUI() {
        addSubviews([waitImageView,DDccidLable,lineView])
        
        waitImageView.snp.makeConstraints { make in
            make.centerX.equalToSuperview()
            make.top.equalToSuperview().offset(34.adaptH)
            make.size.equalTo(CGSize(width: 50, height: 50).adapt)
        }
        
        DDccidLable.snp.makeConstraints { make in
            make.centerX.equalToSuperview()
            make.top.equalTo(waitImageView.snp.bottom).offset(23.adaptH)
        }
        
        lineView.snp.makeConstraints { make in
            make.left.right.bottom.equalToSuperview()
            make.height.equalTo(3.adaptH)
        }
    }
    
    private func startImageAnimation() {
        animationTimer = Timer.scheduledTimer(withTimeInterval: 0.2, repeats: true) { [weak self] _ in
            self?.updateAnimationImage()
        }
    }
    
    private func stopImageAnimation() {
        animationTimer?.invalidate()
        animationTimer = nil
    }
    
    private func updateAnimationImage() {
        currentImageIndex = (currentImageIndex + 1) % totalImages
        let imageName = String(format: "search_anima_%d", currentImageIndex + 1)
        waitImageView.image = UIImage(named: imageName)
    }
    
    lazy var waitImageView: UIImageView = {
        let imageView = UIImageView(frame: .zero)
        imageView.image = UIImage(named: "search_anima_1")
        imageView.clipsToBounds = true
        imageView.contentMode = .scaleAspectFit
        return imageView
    }()
    
    lazy var DDccidLable: UILabel = {
        let text = UILabel(frame: .zero)
        text.numberOfLines = 0
        text.textColor = UIColor.init(hex: 0x007AFF)
        text.font = UIFont.boldSystemFont(ofSize: 15)
        text.text = """
                    Now searching for the monitor...
                    Please select a monitor to connect to
                    """
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

class MuilpServiceViewCell: UITableViewCell {
    
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
        imgView.image = UIImage(named: "Ddccid")
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
    
    func configure(with model:LanServiceInfo) {
        self.deviceNameLab.text = model.monitorName
        self.deviceIpLab.text = model.ip
    }
}
