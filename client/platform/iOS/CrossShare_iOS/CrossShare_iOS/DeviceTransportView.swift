//
//  DeviceTransportView.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/5/15.
//

import UIKit
import MBProgressHUD

class DeviceTransportView: UIView {
    
    var mFileOpener: FileOpener? = nil
    public var dataArray:[DownloadItem] = [] {
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
        self.tableView.tableHeaderView?.size = CGSize(width: UIScreen.main.bounds.width, height: 40)
        
        self.tableView.setNeedsLayout()
        self.tableView.layoutIfNeeded()
        
        self.recordLable.snp.makeConstraints { make in
            make.left.equalToSuperview().offset(12)
            make.centerY.equalToSuperview()
        }
        
        self.cleanView.snp.makeConstraints { make in
            make.left.equalTo(self.recordLable.snp.right).offset(20)
            make.centerY.equalTo(self.recordLable)
            make.size.equalTo(CGSize(width: 30, height: 30))
        }
        
        self.tableView.snp.makeConstraints { make in
            make.top.bottom.equalToSuperview()
            make.left.equalToSuperview().offset(12)
            make.centerX.equalToSuperview()
        }
    }
    
    lazy var tableView: UITableView = {
        let view = UITableView()
        view.backgroundColor = .clear
        view.delegate = self
        view.dataSource = self
        view.showsVerticalScrollIndicator = false
        view.rowHeight = 130
        view.separatorStyle = .none
        view.register(DownloadViewCell.self,
                      forCellReuseIdentifier: NSStringFromClass(DownloadViewCell.self))
        return view
    }()
    
    lazy var tableViewTopView: UIView = {
        let imgView = UIView()
        imgView.clipsToBounds = true
        imgView.backgroundColor = .white
        return imgView
    }()
    
    lazy var recordLable: UILabel = {
        let text = UILabel(frame: .zero)
        text.textColor = UIColor.black
        text.font = UIFont.systemFont(ofSize: 14)
        text.text = "Files records"
        tableViewTopView.addSubview(text)
        return text
    }()
    
    lazy var cleanView: UIImageView = {
        let imageView = UIImageView(image: UIImage(named: "clear"))
        imageView.contentMode = .scaleAspectFit
        imageView.clipsToBounds = true
        imageView.isUserInteractionEnabled = true
        let tap = UITapGestureRecognizer(target: self, action: #selector(clearAction))
        imageView.addGestureRecognizer(tap)
        tableViewTopView.addSubview(imageView)
        return imageView
    }()
}

extension DeviceTransportView {
    @objc private func clearAction() {
        DispatchQueue.main.async { [weak self] in
            guard let self = self else {
                return
            }
            self.dataArray.removeAll()
            self.tableView.reloadData()
        }
    }
    
    @objc private func deleteFile(at index: Int) {
        DispatchQueue.main.async { [weak self] in
            guard let self = self else { return }
            if index >= 0 && index < self.dataArray.count {
                self.dataArray.remove(at: index)
                self.tableView.reloadData()
            }
        }
    }

    @objc private func openFile(at index: Int) {
        DispatchQueue.main.async { [weak self] in
            guard let self = self else { return }
            if index >= 0 && index < self.dataArray.count {
                let filesItem = self.dataArray[index]
                guard let fileCnt = filesItem.totalFileCnt else {
                    return
                }

                if fileCnt > 1 {
                    MBProgressHUD.showTips(.error,"Only file can be opened", toView: self)
                    return
                }

                guard let filePath = filesItem.currentfileName else {
                    return
                }

                var fileName = ""
                if filesItem.isMutip {
                    if let nameArray = filesItem.currentfileName?.components(separatedBy: "/") as? [String],nameArray.count > 1 {
                        fileName = nameArray.last ?? ""
                    }
                } else {
                    fileName = filePath
                }

                guard let viewController = self.parentViewController else {
                    return
                }

                mFileOpener = FileOpener(presenter: viewController)
                if (mFileOpener?.openFile(fileName: fileName) == false) {
                    MBProgressHUD.showTips(.error,"Only file can be opened", toView: self)
                }
            }
        }
    }
}

extension DeviceTransportView:UITableViewDelegate,UITableViewDataSource {
    func tableView(_ tableView: UITableView, numberOfRowsInSection section: Int) -> Int {
        return dataArray.count
    }
    
    func tableView(_ tableView: UITableView, cellForRowAt indexPath: IndexPath) -> UITableViewCell {
        guard let cell = tableView.dequeueReusableCell(withIdentifier: NSStringFromClass(DownloadViewCell.self)) as? DownloadViewCell else {
            return UITableViewCell()
        }
        let model = dataArray[indexPath.row]
        cell.refreshUI(with: model)
        cell.deleteBlock = { [weak self] in
            guard let self = self else { return  }
            self.deleteFile(at: indexPath.row)
        }
        cell.openBlock = { [weak self] in
            guard let self = self else { return }
            self.openFile(at: indexPath.row)
        }
        return cell
    }
}
