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
import SwiftyJSON

enum ServiceStatus:Int {
    case disconnected = 0
    case searching
    case connected
}

class HomeViewController: BaseViewController {
    private var ugradeView: UpgradeView?
    private var lastViewType: ServiceStatus? = nil
    private var selectedService: LanServiceInfo? = nil
    private var statusPopView: ConnectPopView?
    private var diasStatus: DiassStatus? = .WaitConnecting {
        didSet {
            if diasStatus != nil {
                self.deviceStatus = self.determiteDeviceStatus(status: diasStatus!)
            } else {
                self.deviceStatus = .disconnected
            }
            
            showStatusPopView(for: diasStatus)
        }
    }
    private var isAutoConnectTimeout: Bool = false
    var deviceStatus:ServiceStatus = .disconnected {
        didSet {
            if lastViewType != nil {
                guard lastViewType != deviceStatus else {
                    return
                }
            }
            lastViewType = deviceStatus
            switch deviceStatus {
            case .disconnected:
                DispatchQueue.main.async { [weak self] in
                    guard let self = self else { return  }
                    self.searchButton.isHidden = true
                    self.deviceList.removeFromSuperview()
                    self.serviceView.removeFromSuperview()
                    self.view.addSubview(self.disConnectedView)
                    
                    self.disConnectedView.snp.makeConstraints { make in
                        make.left.right.equalToSuperview()
                        make.top.equalTo(self.view.safeAreaLayoutGuide.snp.top)
                        make.bottom.equalToSuperview()
                    }
                    self.view.setNeedsLayout()
                    self.view.layoutIfNeeded()
                }
            case .searching:
                DispatchQueue.main.async { [weak self] in
                    guard let self = self else { return  }
                    self.searchButton.isHidden = true
                    self.disConnectedView.removeFromSuperview()
                    self.deviceList.removeFromSuperview()
                    self.view.addSubview(self.serviceView)
                    
                    self.serviceView.snp.makeConstraints { make in
                        make.left.right.equalToSuperview()
                        make.top.equalTo(self.view.safeAreaLayoutGuide.snp.top)
                        make.bottom.equalToSuperview()
                    }
                    self.isAutoConnectTimeout = false
                    self.serviceView.onSelectService = { [weak self] lanServiceInfo in
                        guard let self = self else { return }
                        self.selectedService = lanServiceInfo
                        self.diasStatus = .CheckingAuthorization
                        //delay to ensure the service is selected before confirming
                        DispatchQueue.main.asyncAfter(deadline: .now() + 0.1) {
                            P2PManager.shared.comfirmLanServer(instance: lanServiceInfo.instance)
                        }
                    }
                    self.view.setNeedsLayout()
                    self.view.layoutIfNeeded()
                }
                
                if let serviceString = UserDefaults.get(forKey: .LAN_SERVICE_INFO,type: .standard) {
                
                    DispatchQueue.main.asyncAfter(deadline: .now() + 5.0) { [weak self] in
                        guard let self = self else { return }
                        self.isAutoConnectTimeout = true
                    }
                }
            case .connected:
                DispatchQueue.main.async { [weak self] in
                    guard let self = self else { return  }
                    self.searchButton.isHidden = false
                    self.disConnectedView.removeFromSuperview()
                    self.serviceView.removeFromSuperview()
                    self.view.addSubview(self.deviceList)
                    
                    self.deviceList.snp.makeConstraints { make in
                        make.left.right.equalToSuperview()
                        make.top.equalTo(self.view.safeAreaLayoutGuide.snp.top)
                        make.bottom.equalToSuperview()
                    }
                    self.deviceList.refreshUI(P2PManager.shared.monitorName)
                    self.deviceList.dataArray = P2PManager.shared.clientList
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
        self.navigationItem.title = "Cross Share"
        self.view.backgroundColor = UIColor.white
        self.navigationItem.rightBarButtonItem = UIBarButtonItem(customView: self.searchButton)

        let session = AVAudioSession.sharedInstance()
        try! session.setCategory(.playback, mode: .moviePlayback)
        try! session.setActive(true)
        
        self.deviceStatus = self.determiteDeviceStatus(status: P2PManager.shared.cStatus)
        self.view.addSubview(videoContainerView)
        self.view.backgroundColor = UIColor.init(hex: 0xF6F6F6)
        
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
    
    private func addNotifications() {
        NotificationCenter.default.addObserver(self, selector: #selector(updateClientList), name: UpdateClientListSuccessNotification, object: nil)
        NotificationCenter.default.addObserver(self, selector: #selector(updateDiasStatus(_:)), name: UpdateDiassStatusChangedNotification, object: nil)
        NotificationCenter.default.addObserver(self, selector: #selector(updateMonitorNameChanged(_:)), name: UpdateMonitorNameChangedNotification, object: nil)
        NotificationCenter.default.addObserver(self, selector: #selector(updateLanserviceStatus(_:)), name: UpdateLanserviceChangedNotification, object: nil)
        NotificationCenter.default.addObserver(self, selector: #selector(updateLanserviceList(_:)), name: UpdateLanserviceListNotification, object: nil)
        NotificationCenter.default.addObserver(self, selector: #selector(updateVersion(_:)), name: UpdateVersionNotification, object: nil)
    }
    
    private func dismissConnectPopView(_ popView: ConnectPopView) {
        UIView.animate(withDuration: 0.3, animations: {
            popView.alpha = 0
        }) { _ in
            popView.removeFromSuperview()
        }
    }

    lazy var videoContainerView: UIView = {
        let view = UIView(frame: .zero)
        view.backgroundColor = UIColor.clear
        view.isUserInteractionEnabled = true
        return view
    }()
    
    lazy var disConnectedView: DefaultView = {
        let view = DefaultView(frame: .zero)
        view.backgroundColor = UIColor.white
        view.isUserInteractionEnabled = true
        return view
    }()
    
    lazy var serviceView: MuiltServiceView = {
        let view = MuiltServiceView(frame: .zero)
        view.backgroundColor = UIColor.white
        view.isUserInteractionEnabled = true
        return view
    }()
    
    lazy var deviceList: MuilpDeviceView = {
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
    
    lazy var searchButton: UIButton = {
        let button = UIButton(type: .custom)
        let buttonSize: CGFloat = 30
        button.frame = CGRect(x: 0, y: 0, width: buttonSize, height: buttonSize)
        if let image = UIImage(named: "nav_search") {
            let resizedImage = image.withRenderingMode(.alwaysOriginal)
            button.setImage(resizedImage, for: .normal)
        }
        button.translatesAutoresizingMaskIntoConstraints = false
        button.widthAnchor.constraint(equalToConstant: buttonSize).isActive = true
        button.heightAnchor.constraint(equalToConstant: buttonSize).isActive = true
        button.imageView?.contentMode = .scaleAspectFit
        button.contentHorizontalAlignment = .center
        button.contentVerticalAlignment = .center
        button.addTarget(self, action: #selector(searchAction(_:)), for: .touchUpInside)
        return button
    }()
}

extension HomeViewController {
    
    private func determiteDeviceStatus (status:DiassStatus) -> ServiceStatus {
        switch status {
        case .ConnectedNoClients, .Connected,. SearchingClients:
            P2PManager.shared.removeServerList()
            return .connected
        case .SearchingService, .CheckingAuthorization, .WaitScreenCasting, .FailedAuthorization, .ConnectedFailed:
            return .searching
        case .WaitConnecting:
            P2PManager.shared.removeServerList()
            return .disconnected
        }
    }
    
    @objc private func updateClientList() {
        DispatchQueue.main.async { [weak self] in
            guard let self = self else {
                return
            }
            self.deviceList.dataArray = P2PManager.shared.clientList
        }
    }
    
    @objc func updateDiasStatus(_ notification: Notification) {
        guard let statusDic = notification.userInfo as? [String:Any],let status = statusDic["status"] as? DiassStatus  else {
            return
        }
        Logger.info("updateDiasStatus: \(status.rawValue)")
        self.diasStatus = status
    }
    
    @objc func updateMonitorNameChanged(_ notification: Notification) {
        guard let monitorDic = notification.userInfo as? [String:Any],let monitorName = monitorDic["monitorName"] as? String  else {
            return
        }
        self.deviceList.refreshUI(monitorName)
    }

    @objc func updateLanserviceStatus(_ notification: Notification) {
        guard let serviceDic = notification.userInfo as? [String:Any],let status = serviceDic["isDetect"] as? Bool else {
            return
        }
        self.deviceStatus = status ? .searching : .disconnected
    }
    
    @objc func updateLanserviceList(_ notification: Notification) {
        guard let serviceList = notification.userInfo?["serviceList"] as? [LanServiceInfo] else {
            return
        }
        DispatchQueue.main.async { [weak self] in
            guard let self = self else { return }
            self.serviceView.dataArray = P2PManager.shared.serviceList
            
            if self.isAutoConnectTimeout {
                return
            }
            
            if let serviceString = UserDefaults.get(forKey: .LAN_SERVICE_INFO,type: .standard),
               let data = serviceString.data(using: .utf8) {
                
                let json = try? JSON(data: data)
                let lanServiceInfo =
                LanServiceInfo(
                    monitorName: json?["monitorName"].stringValue ?? "",
                    instance: json?["instance"].stringValue ?? "",
                    ip: json?["ip"].stringValue ?? "",
                    version: json?["version"].stringValue ?? "",
                    timestamp: json?["timestamp"].uInt64Value ?? UInt64(Date().timeIntervalSince1970 * 1000)
                )

                for service in P2PManager.shared.serviceList {
                    //catch the service that matches the favorite lanServiceInfo
                    if service.instance == lanServiceInfo.instance {
                        isAutoConnectTimeout = true
                        P2PManager.shared.comfirmLanServer(instance: lanServiceInfo.instance)
                        Logger.info("Confirmed service with instance: \(service.instance)")
                        break
                    }
                }
            }
        }
    }
    
    @objc private func searchAction(_ sender: UIButton) {
        self.diasStatus = .SearchingClients
    }

    @objc func updateVersion(_ notification: Notification) {
        guard let versionDic = notification.userInfo as? [String:Any],let _ = versionDic["version"] as? String  else {
            return
        }
        DispatchQueue.main.async { [weak self] in
            UpgradeManager.shared.showDefaultUpgradeAlert()
        }
    }
    
    private func dismissDevicePopView() {
        guard let popView = self.ugradeView else { return }
        UIView.animate(withDuration: 0.3, animations: {
            popView.alpha = 0
            if let contentView = popView.subviews.first {
                contentView.transform = CGAffineTransform(scaleX: 0.8, y: 0.8)
            }
        }) { _ in
            popView.removeFromSuperview()
            self.ugradeView = nil
        }
    }
    
    private func showStatusPopView(for status: DiassStatus?) {
        guard let status = status else { return }
        
        var popType: ConnectPopType = .none
        var title: String = ""
        var content: String = ""
        var shouldShow: Bool = false
        
        switch status {
        case .WaitConnecting: break
        case .SearchingService:
            popType = .waiting
            content = "Searching for available services..."
            shouldShow = false
            
        case .CheckingAuthorization:
            popType = .connecting
            content = "Connecting to the monitor..."
            shouldShow = true
            
        case .WaitScreenCasting:
            popType = .waiting
            content = "Verifying..."
            shouldShow = true
            
        case .FailedAuthorization:
            title = "Verification failed"
            popType = .authFailed
            content = "Please ensure the phone screen is displayed on the monitor"
            shouldShow = true
            
        case .ConnectedNoClients, .Connected:
            shouldShow = false
            if let selectedService = self.selectedService {
                let json = JSON(selectedService.toDictionary())
                if let jsonString = json.rawString() {
                    UserDefaults.set(forKey: .LAN_SERVICE_INFO, value: jsonString, type: .standard)
                }
            }
            
        case .ConnectedFailed:
            title = "Connection failed"
            popType = .connectionFailed
            content = "Please ensure the monitor is on the local network"
            shouldShow = true
            
        case .SearchingClients:
            title = "Warning"
            popType = .multiple
            content =       """
                               Searching for a monitor will disconnect the current connection.
                               
                               Do you want to continue?
                            """
            shouldShow = true
            if let lanService = UserDefaults.get(forKey: .LAN_SERVICE_INFO,type: .standard), !lanService.isEmpty {
                UserDefaults.del(forKey: .LAN_SERVICE_INFO, type: .standard)
            }
        }
        
        if shouldShow {
            showConnectPopView(type: popType, title: title, content: content)
        } else {
            hideStatusPopView()
        }
    }
    
    private func showConnectPopView(type: ConnectPopType,title:String, content: String) {
        DispatchQueue.main.async { [weak self] in
            guard let self = self else { return }
            
            if self.statusPopView == nil {
                self.statusPopView = ConnectPopView(frame: self.view.bounds, type: type,tittle: title, content: content)
                self.setupPopViewCallbacks()
            } else {
                self.statusPopView?.updateContent(type: type,tittle: title, content: content)
            }
            if let window = UtilsHelper.shared.getTopWindow() {
                self.statusPopView?.show(in: window)
            }
        }
    }
    
    private func hideStatusPopView() {
        DispatchQueue.main.async { [weak self] in
            self?.statusPopView?.hide { [weak self] in
                self?.statusPopView = nil
            }
        }
    }
    
    private func setupPopViewCallbacks() {
        statusPopView?.onContinue = { [weak self] in
            self?.hideStatusPopView()
        }
        
        statusPopView?.onCancel = { [weak self] in
            self?.hideStatusPopView()
        }
        
        statusPopView?.onConfirm = { [weak self] in
            self?.hideStatusPopView()
            P2PManager.shared.browseLanService()
        }
    }
}
