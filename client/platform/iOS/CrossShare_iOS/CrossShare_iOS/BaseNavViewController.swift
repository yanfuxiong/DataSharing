//
//  BaseNavViewController.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/6/16.
//

import UIKit

class BaseNavViewController: UINavigationController {
    
    override func viewDidLoad() {
        super.viewDidLoad()
        
        initUIConfig()
    }
    
    override func pushViewController(_ viewController: UIViewController, animated: Bool) {
        if viewControllers.count > 0 {
            viewController.hidesBottomBarWhenPushed = true
            let backBtn = UIButton.init(type: .custom)
            backBtn.setImage(UIImage.init(named: "back"), for: .normal)
            backBtn.size = CGSize.init(width: 44.0, height: 44.0)
            backBtn.addTarget(self, action: #selector(navBackAction), for: .touchUpInside)
            let leftItem = UIBarButtonItem.init(customView: backBtn)
            viewController.navigationItem.leftBarButtonItem = leftItem
        }
        super.pushViewController(viewController, animated: animated)
    }
    
    override func present(_ viewControllerToPresent: UIViewController, animated flag: Bool, completion: (() -> Void)? = nil) {
        guard presentedViewController == nil else { return }
        guard viewControllerToPresent.isBeingPresented == false else { return }
        viewControllerToPresent.modalPresentationStyle = .fullScreen
        super.present(viewControllerToPresent, animated: flag, completion: completion)
    }
    
    func initUIConfig() {
        setupNavigationBarAppearance()
    }
    
    func setupNavigationBarAppearance() {
        if #available(iOS 13.0, *) {
            let appearance = UINavigationBarAppearance()
            appearance.configureWithOpaqueBackground()
            appearance.backgroundColor = UIColor.init(hex: 0x007AFF)
            appearance.titleTextAttributes = [.foregroundColor: UIColor.white]
            appearance.largeTitleTextAttributes = [.foregroundColor: UIColor.white]
            
            UINavigationBar.appearance().standardAppearance = appearance
            UINavigationBar.appearance().scrollEdgeAppearance = appearance
            UINavigationBar.appearance().compactAppearance = appearance
            UINavigationBar.appearance().tintColor = UIColor.white
        } else {
            UINavigationBar.appearance().barTintColor = UIColor.init(hex: 0x007AFF)
            UINavigationBar.appearance().tintColor = UIColor.white
            UINavigationBar.appearance().titleTextAttributes = [NSAttributedString.Key.foregroundColor: UIColor.white]
            UINavigationBar.appearance().isTranslucent = false
        }
    }
    
    @objc
    func navBackAction() {
        popViewController(animated: true)
    }
    
    @objc
    func dismissAction() {
        dismiss(animated: true)
    }
    
    deinit {
        Logger.info(String(describing: object_getClass(self)))
    }
    
}


extension  BaseNavViewController: UIGestureRecognizerDelegate {
    
    func gestureRecognizerShouldBegin(_ gestureRecognizer: UIGestureRecognizer) -> Bool {
        if gestureRecognizer == interactivePopGestureRecognizer {
            if viewControllers.count < 2 || visibleViewController == viewControllers.first {
                return false
            }
        }
        return true
    }
}
