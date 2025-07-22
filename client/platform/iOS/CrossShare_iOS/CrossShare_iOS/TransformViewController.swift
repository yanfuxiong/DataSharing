//
//  TransformViewController.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/6/16.
//

import UIKit

class TransformViewController: BaseViewController {
    
    private var dataArray:[DownloadItem] = []
    private var userManuallySwitchedFromFile = false
    
    override func viewDidLoad() {
        super.viewDidLoad()
        
        setupUI()
        initialize()
        addNotifications()
    }
    
    func setupUI() {
        self.title = "Cross Share"
        self.view.addSubview(self.fileContennerView)
        
        self.fileContennerView.snp.makeConstraints { make in
            make.left.right.equalToSuperview()
            make.top.equalTo(self.view.safeAreaLayoutGuide.snp.top)
            make.bottom.equalToSuperview()
        }
        self.view.setNeedsLayout()
        self.view.layoutIfNeeded()
    }
    
    func initialize() {
        
    }
    
    func addNotifications() {
        NotificationCenter.default.addObserver(self, selector: #selector(receriveNewFiles(_:)), name: ReceiveFuleSuccessNotification, object: nil)
    }
    
    lazy var fileContennerView: DeviceTransportView = {
        let view = DeviceTransportView(frame: .zero)
        view.backgroundColor = UIColor.clear
        view.isUserInteractionEnabled = true
        return view
    }()
}

extension TransformViewController {
    @objc private func receriveNewFiles(_ ntf: Notification) {
        DispatchQueue.main.async { [weak self] in
            guard let self = self else { return }
            Logger.info("TransformViewController receiveNewFiles \(self.tabBarController?.selectedIndex)")
            if self.tabBarController?.selectedIndex != 1 {
                self.tabBarController?.selectedIndex = 1
            }
            if let userInfo = ntf.userInfo as? [String: Any],
               let newItem = userInfo["download"] as? DownloadItem {
                var isNewFile = false
                if let index = self.dataArray.firstIndex(where: { $0.uuid == newItem.uuid }) {
                    self.dataArray[index].receiveSize = newItem.receiveSize
                    self.dataArray[index].totalSize = newItem.totalSize
                    self.dataArray[index].finishTime = newItem.finishTime
                    self.dataArray[index].timestamp = newItem.timestamp
                    if let recvFileCnt = newItem.recvFileCnt, let currentfileName = newItem.currentfileName {
                        self.dataArray[index].recvFileCnt = recvFileCnt
                        self.dataArray[index].currentfileName = currentfileName
                    }
                } else {
                    self.dataArray.append(newItem)
                    isNewFile = true
                }
                self.dataArray.sort { ($0.timestamp ?? 0) > ($1.timestamp ?? 0) }
                self.fileContennerView.dataArray = self.dataArray
                if isNewFile {
                    self.userManuallySwitchedFromFile = false
                    if let index = self.dataArray.firstIndex(where: { $0.uuid == newItem.uuid }) {
                        let fileCnt = (self.dataArray[index].totalFileCnt ?? 0)
                        let fileMsg = fileCnt > 1 ? "\(fileCnt) files" : "\(fileCnt) file"
                        let deviceName = self.dataArray[index].deviceName ?? ""
                        let params: [String] = [fileMsg, deviceName]
                        PushNotiManager.shared.sendLocalNotification(code: .receiveStart, with: params)
                    }
                }
            }
        }
    }
}

