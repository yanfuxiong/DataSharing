//
//  HomeViewController.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/6/16.
//

import UIKit
import AVKit
import AVFoundation
import SnapKit
import MBProgressHUD

enum DeviceStatus {
    case `default`
    case online
    case offline
}

class HomeViewController: BaseViewController {
    private var lastViewType: DeviceStatus? = nil
    var deviceStatus:DeviceStatus = .default {
        didSet {
            if lastViewType != nil {
                guard lastViewType != deviceStatus else {
                    return
                }
            }
            lastViewType = deviceStatus
            switch deviceStatus {
            case .default:
                Logger.info("switch main")
                DispatchQueue.main.async { [weak self] in
                    guard let self = self else { return  }
                    self.configContenView.removeFromSuperview()
                    self.fileContennerView.removeFromSuperview()
                    self.view.addSubview(self.deviceView)
                    
                    self.deviceView.snp.makeConstraints { make in
                        make.left.right.equalToSuperview()
                        make.top.equalTo(self.videoContainerView.snp.bottom)
                        make.bottom.equalToSuperview()
                    }
                    self.view.setNeedsLayout()
                    self.view.layoutIfNeeded()
                }
            case .offline:
                Logger.info("switch config")
                DispatchQueue.main.async { [weak self] in
                    guard let self = self else { return  }
                    self.aheadButton.isSelected = true
                    self.fileContennerView.removeFromSuperview()
                    self.deviceView.removeFromSuperview()
                    self.view.addSubview(self.configContenView)
                    
                    self.configContenView.snp.makeConstraints { make in
                        make.left.right.equalToSuperview()
                        make.top.equalTo(self.videoContainerView.snp.bottom)
                        make.bottom.equalToSuperview()
                    }
                    
                    self.configContenView.refreshUI()
                    self.configContenView.submitConfigBlock = {[weak self] (ddcciText,deviceSourceText,devicePortText) in
                        guard let self = self else { return  }
                        guard ddcciText.isEmpty == false else {
                            MBProgressHUD.showTips(.error,"Please enter device ddcci id", toView: self.view)
                            return
                        }
                        guard deviceSourceText.isEmpty == false else {
                            MBProgressHUD.showTips(.error,"Please enter device source", toView: self.view)
                            return
                        }
                        guard devicePortText.isEmpty == false else {
                            MBProgressHUD.showTips(.error,"Please enter device port", toView: self.view)
                            return
                        }
                        if let intSource = Int(deviceSourceText), let intPort = Int(devicePortText) {
                            UserDefaults.set(forKey: .DEVICECONFIG_DIAS_ID, value: ddcciText)
                            UserDefaults.setInt(forKey: .DEVICECONFIG_SRC, value: intSource)
                            UserDefaults.setInt(forKey: .DEVICECONFIG_PORT, value: intPort)
                            P2PManager.shared.setupDeviceConfig(ddcciText, intSource, intPort)
                            self.configContenView.refreshUI()
                        }
                    }
                    self.view.setNeedsLayout()
                    self.view.layoutIfNeeded()
                }
            case .online:
                Logger.info("switch online")
                DispatchQueue.main.async { [weak self] in
                    guard let self = self else { return  }
                    self.deviceView.removeFromSuperview()
                    self.configContenView.removeFromSuperview()
                    self.view.addSubview(self.fileContennerView)
                    
                    self.fileContennerView.snp.makeConstraints { make in
                        make.left.right.equalToSuperview()
                        make.top.equalTo(self.videoContainerView.snp.bottom)
                        make.bottom.equalToSuperview()
                    }
                    self.fileContennerView.refreshUI()
                    self.fileContennerView.dataArray = P2PManager.shared.clientList
                    self.view.setNeedsLayout()
                    self.view.layoutIfNeeded()
                }
            }
        }
    }
    
    override func viewDidLoad() {
        super.viewDidLoad()
        setupUI()
        initialize()
        addNotifications()
    }
    
    func initialize() {
        Logger.info("before PIP init：\(UIApplication.shared.windows)")
        PictureInPictureManager.shared.initializePIP()
    }
    
    deinit {
        NotificationCenter.default.removeObserver(self)
    }
    
    override func viewWillAppear(_ animated: Bool) {
        super.viewWillAppear(animated)
        
        setupVideoDisplayLayer()
    }
    
    private func setupVideoDisplayLayer() {
        self.videoView.setNeedsLayout()
        self.videoView.layoutIfNeeded()
        
        PictureInPictureManager.shared.setupDisplayLayer(for: self.videoView)
    }
    
    func setupUI() {
        self.title = "Cross Share"
        self.navigationItem.rightBarButtonItem = UIBarButtonItem(customView: addButton)
        self.navigationItem.leftBarButtonItem = UIBarButtonItem(customView: self.aheadButton)
        
        self.view.backgroundColor = UIColor.white
        
        let session = AVAudioSession.sharedInstance()
        try! session.setCategory(.playback, mode: .moviePlayback)
        try! session.setActive(true)
        
        self.deviceStatus = P2PManager.shared.clientList.count != 0 || !P2PManager.shared.deviceDiasId.isEmpty ? .online : .default
        self.aheadButton.isSelected = self.deviceStatus == .offline
        self.view.addSubview(videoContainerView)
        self.view.backgroundColor = UIColor.init(hex: 0xF6F6F6)
        
        self.videoContainerView.snp.makeConstraints { make in
            make.left.right.equalToSuperview()
            make.top.equalTo(self.view.safeAreaLayoutGuide.snp.top).offset(10)
            make.height.equalTo(30)
        }
        
        self.videoView.snp.makeConstraints { make in
            make.edges.equalToSuperview()
        }
        
        setupVideoDisplayLayer()
    }
    
    private func addNotifications() {
        NotificationCenter.default.addObserver(self, selector: #selector(updateClientList), name: UpdateClientListSuccessNotification, object: nil)
    }
    
    lazy var videoContainerView: UIView = {
        let view = UIView(frame: .zero)
        view.backgroundColor = UIColor.clear
        view.isUserInteractionEnabled = true
        return view
    }()
    
    lazy var deviceView: DefaultView = {
        let view = DefaultView(frame: .zero)
        view.backgroundColor = UIColor.white
        view.isUserInteractionEnabled = true
        return view
    }()
    
    lazy var configContenView: DeviceConfigView = {
        let view = DeviceConfigView(frame: .zero)
        view.backgroundColor = UIColor.white
        view.isUserInteractionEnabled = true
        return view
    }()
    
    lazy var fileContennerView: MuilpDeviceView = {
        let view = MuilpDeviceView(frame: .zero)
        view.backgroundColor = UIColor.white
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
    
    lazy var addButton: UIButton = {
        let button = UIButton(type: .custom)
        button.frame = CGRect(x: 0, y: 0, width: 44, height: 44)
        button.setImage(UIImage(named: "add"), for: .normal)
        button.setImage(UIImage(named: "add"), for: .selected)
        button.contentHorizontalAlignment = .right
        button.addTarget(self, action: #selector(addDeviceConfig(_:)), for: .touchUpInside)
        return button
    }()
    
    lazy var aheadButton: UIButton = {
        let button = UIButton(type: .custom)
        button.frame = CGRect(x: 0, y: 0, width: 44, height: 44)
        button.setImage(nil, for: .normal)
        button.setImage(UIImage(named: "arrow_left"), for: .selected)
        button.contentHorizontalAlignment = .left
        button.addTarget(self, action: #selector(changeView(_:)), for: .touchUpInside)
        return button
    }()
}

extension HomeViewController {
    @objc func addDeviceConfig(_ sender: UIButton) {
        self.deviceStatus = .offline
    }
    
    @objc private func updateClientList() {
        DispatchQueue.main.async { [weak self] in
            guard let self = self else {
                return
            }
            self.deviceStatus = P2PManager.shared.clientList.count != 0 ? .online : .default
            self.aheadButton.isSelected = false
            self.fileContennerView.refreshUI()
            self.fileContennerView.dataArray = P2PManager.shared.clientList
        }
    }
    
    @objc func changeView(_ sender:UIButton) {
        sender.isSelected = !sender.isSelected
        if (P2PManager.shared.clientList.count != 0 || !P2PManager.shared.deviceDiasId.isEmpty) {
            self.deviceStatus = .online
        } else {
            self.deviceStatus = .default
        }
    }
}
