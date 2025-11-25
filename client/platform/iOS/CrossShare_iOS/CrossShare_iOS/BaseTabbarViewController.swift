//
//  BaseTabbarViewController.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/6/16.
//

import UIKit

class BaseTabbarViewController: UITabBarController {
    
    override func viewDidLoad() {
        super.viewDidLoad()
        
        setupUI()
        addNotifications()
    }
    
    func addNotifications() {
        FileTransferDataManager.shared.addObserver(self)
    }
    
    deinit {
        FileTransferDataManager.shared.removeObserver(self)
    }
    
    func  setupUI() {
        if #available(iOS 13.0, *) {
            let tabbar = UITabBar.appearance()
            tabbar.tintColor = UIColor.init(hex:0x007AFF)
            tabbar.backgroundColor = .white
            tabbar.unselectedItemTintColor = UIColor.init(hex:0x999999,alpha: 1.0)
        } else {
            let appearance = UITabBarItem.appearance()
            appearance.setTitleTextAttributes([.foregroundColor: UIColor.init(hex:0x999999,alpha: 1.0) as Any,
                                               .font: UIFont.systemFont(ofSize: 10)],
                                              for: .normal)
            appearance.setTitleTextAttributes([.foregroundColor: UIColor.init(hex:0x007AFF) as Any,
                                               .font: UIFont.systemFont(ofSize: 10)],
                                              for: .selected)
        }
        let homeVc = HomeViewController()
        setupItem(vc: homeVc, title: "Share", imageName: "tabbar_home")
        
        let transformVc = TransformViewController()
        setupItem(vc: transformVc, title: "Record", imageName: "tabbar_record")
        
        let settingVc = SettingsViewController()
        setupItem(vc: settingVc, title: "Info", imageName: "tabbar_info")
    }
    
    private func setupItem(vc: BaseViewController, title: String, imageName: String) {
        vc.title = title
        vc.tabBarItem.image = resizeTabBarImage(named: imageName)
        vc.tabBarItem.selectedImage = resizeTabBarImage(named: "\(imageName)_selected")
        
        vc.tabBarItem.setTitleTextAttributes([.foregroundColor: UIColor.init(hex:0x999999,alpha: 1.0) as Any,
                                              .font: UIFont.systemFont(ofSize: 10)],
                                             for: .normal)
        vc.tabBarItem.setTitleTextAttributes([.foregroundColor: UIColor.init(hex:0x007AFF) as Any,
                                              .font: UIFont.systemFont(ofSize: 10)],
                                             for: .selected)
        let nav = BaseNavViewController(rootViewController: vc)
        addChild(nav)
    }
    
    private func resizeTabBarImage(named imageName: String) -> UIImage? {
        guard let originalImage = UIImage(named: imageName) else {
            return nil
        }
        let targetSize = CGSize(width: 25, height: 25)
        UIGraphicsBeginImageContextWithOptions(targetSize, false, UIScreen.main.scale)
        originalImage.draw(in: CGRect(origin: .zero, size: targetSize))
        let resizedImage = UIGraphicsGetImageFromCurrentImageContext()
        UIGraphicsEndImageContext()
        return resizedImage?.withRenderingMode(.alwaysOriginal)
    }
}

extension BaseTabbarViewController: FileTransferDataObserver {
    func dataDidUpdate(_ data: [DownloadItem]) {
        DispatchQueue.main.async { [weak self] in
            guard let self = self else { return }
            if self.selectedIndex != 1 && FileTransferDataManager.shared.isNewDataTransfering {
                self.selectedIndex = 1
            }
        }
    }
}
