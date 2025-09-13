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
        FileTransferDataManager.shared.addObserver(self)
        initialize()
        setupShareExtensionMonitoring()
     }

     private func setupShareExtensionMonitoring() {
         SharedCommunicationManager.shared.onTransferProgress = { [weak self] payload in
             self?.handleShareExtensionProgress(payload: payload)
         }
     }

     private func handleShareExtensionProgress(payload: [String: Any]) {
         guard let status = payload["status"] as? String else { return }
         
         Logger.info("TransformViewController: Share Extension status update: \(status)")
         
         switch status {
         case "transfer_started":
             DispatchQueue.main.async { [weak self] in
                 self?.ensureDataSync()
             }
         default:
             break
         }
     }
    
    override func viewDidLayoutSubviews() {
        super.viewDidLayoutSubviews()
        
        DispatchQueue.main.async { [weak self] in
            self?.ensureDataSync()
        }
    }
    
    private func ensureDataSync() {
        let latestData = FileTransferDataManager.shared.getCurrentData()
        if latestData.count != dataArray.count || dataArray.isEmpty {
            Logger.info("TransformViewController: syncing data, latest count: \(latestData.count)")
            dataArray = latestData
            fileContennerView.dataArray = latestData
        }
    }
    
    deinit {
        FileTransferDataManager.shared.removeObserver(self)
    }
    
    private func setupVideoDisplayLayer() {
        self.videoView.setNeedsLayout()
        self.videoView.layoutIfNeeded()

        PictureInPictureManager.shared.setupDisplayLayer(for: self.videoView)
    }
    
    func setupUI() {
        self.navigationItem.title = "Cross Share"
        self.view.addSubview(videoContainerView)
        self.view.addSubview(self.fileContennerView)
        
        self.fileContennerView.snp.makeConstraints { make in
            make.left.right.equalToSuperview()
            make.top.equalTo(self.view.safeAreaLayoutGuide.snp.top)
            make.bottom.equalToSuperview()
        }
        self.view.setNeedsLayout()
        self.view.layoutIfNeeded()
        
        self.videoContainerView.snp.makeConstraints { make in
            make.left.right.equalToSuperview()
            make.top.equalTo(self.view.safeAreaLayoutGuide.snp.top).offset(10)
            make.height.equalTo(1)
        }

        self.videoView.snp.makeConstraints { make in
            make.edges.equalToSuperview()
        }

        setupVideoDisplayLayer()
    }
    
    func initialize() {
        dataArray = FileTransferDataManager.shared.getCurrentData()
        Logger.info("TransformViewController: initialize with \(dataArray.count) items")
        
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.1) {
            self.fileContennerView.dataArray = self.dataArray
        }
    }
    
    override func viewWillAppear(_ animated: Bool) {
        super.viewWillAppear(animated)
        ensureDataSync()
        setupVideoDisplayLayer()
    }
    
    lazy var videoContainerView: UIView = {
        let view = UIView(frame: .zero)
        view.backgroundColor = UIColor.clear
        view.isUserInteractionEnabled = true
        return view
    }()
    
    lazy var videoView: UIView = {
        let view = UIView(frame: .zero)
        view.backgroundColor = UIColor.white
        view.isUserInteractionEnabled = true
        videoContainerView.addSubview(view)
        return view
    }()
    
    lazy var fileContennerView: DeviceTransportView = {
        let view = DeviceTransportView(frame: .zero)
        view.backgroundColor = UIColor.clear
        view.isUserInteractionEnabled = true
        return view
    }()
}

extension TransformViewController: FileTransferDataObserver {
    func dataDidUpdate(_ data: [DownloadItem]) {
//        Logger.info("TransformViewController: dataDidUpdate called with \(data.count) items")
        
        DispatchQueue.main.async { [weak self] in
            guard let self = self else { return }
            
            let oldInProgressCount = self.dataArray.filter { ($0.receiveSize ?? 0) < ($0.totalSize ?? 0) && $0.error == nil }.count
            let newInProgressCount = data.filter { ($0.receiveSize ?? 0) < ($0.totalSize ?? 0) && $0.error == nil }.count
            
//            Logger.info("TransformViewController: oldInProgress=\(oldInProgressCount), newInProgress=\(newInProgressCount)")
            
            self.dataArray = data
            self.fileContennerView.dataArray = data
            
            if newInProgressCount > oldInProgressCount {
                Logger.info("TransformViewController: switching to InProgress")
                self.fileContennerView.switchToInProgress()
                
                if self.tabBarController?.selectedIndex != 1 {
                    self.tabBarController?.selectedIndex = 1
                }
            }
        }
    }
}
