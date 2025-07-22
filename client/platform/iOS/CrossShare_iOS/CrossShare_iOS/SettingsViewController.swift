//
//  SettingsViewController.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/6/16.
//

import UIKit

class SettingsViewController: BaseViewController {
    
    override func viewDidLoad() {
        super.viewDidLoad()
        setupUI()
        initialize()
        addNotifications()
    }
    
    func setupUI() {
        self.title = "Cross Share"
        self.view.addSubview(self.deviceView)
        
        self.deviceView.snp.makeConstraints { make in
            make.left.right.equalToSuperview()
            make.top.equalTo(self.view.safeAreaLayoutGuide.snp.top)
            make.bottom.equalToSuperview()
        }
        self.deviceView.refreshUI()
        self.deviceView.dosomothingBlock = { [weak self] in
            guard let self = self else { return }
            let vc = PrivacyViewController()
            self.navigationController?.pushViewController(vc, animated: true)
        }
    }
    
    func initialize() {
        
    }
    
    private func addNotifications() {
        NotificationCenter.default.addObserver(self, selector: #selector(netWorkChanged(_:)), name: CSNetworkAccessibilityChangedNotification, object: nil)
    }
    
    @objc private func netWorkChanged(_ ntf: Notification) {
        Logger.info("SettingsViewController - Network notification received: \(ntf)")
        
        let status = CSNetworkAccessibility.sharedInstance().currentState()
        Logger.info("SettingsViewController - Current network status: \(status.rawValue)")
        
        switch status {
        case .checking:
            Logger.info("Network status: checking")
        case .unknown:
            Logger.info("Network status: unknown")
        case .accessible, .accessibleWiFi, .accessibleCellular:
            Logger.info("Network status: accessible")
            self.deviceView.refreshUI()
        case .restricted:
            Logger.info("Network status: restricted")
        }
    }
    
    lazy var deviceView: DeviceMainView = {
        let view = DeviceMainView(frame: .zero)
        view.backgroundColor = UIColor.clear
        view.isUserInteractionEnabled = true
        return view
    }()
    
}
