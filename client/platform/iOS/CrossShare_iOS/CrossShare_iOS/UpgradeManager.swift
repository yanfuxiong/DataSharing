import UIKit

class UpgradeManager {
    
    static let shared = UpgradeManager()
    
    private var upgradeView: UpgradeView?
    private var currentWindow: UIWindow?
    
    private init() {}
    
    func shouldShowUpgrade() -> Bool {
        guard !P2PManager.shared.newVersion.isEmpty else {
            return false
        }
        return P2PManager.shared.majorVersion != P2PManager.shared.newVersion
    }
    
    func checkAndShowUpgradeIfNeeded(forceCheck: Bool = false, completion: ((Bool) -> Void)? = nil) {
        if shouldShowUpgrade() || forceCheck {
            showDefaultUpgradeAlert()
            completion?(true)
        } else {
            completion?(false)
        }
    }
    
    func showUpgradeAlert(title: String = "Unsupported Version",
                          message: String = "This version is no longer supported. Please update the app to continue using this feature.",
                          onUpgrade: @escaping () -> Void,
                          onCancel: (() -> Void)? = nil) {
        
        guard upgradeView == nil else { return }
        
        DispatchQueue.main.async { [weak self] in
            self?.createAndShowUpgradeView(title: title, message: message, onUpgrade: onUpgrade, onCancel: onCancel)
        }
    }
    
    func hideUpgradeAlert() {
        DispatchQueue.main.async { [weak self] in
            self?.dismissUpgradeView()
        }
    }
    
    var isShowing: Bool {
        return upgradeView != nil
    }
    
    func showDefaultUpgradeAlert() {
        showUpgradeAlert(
            onUpgrade: { [weak self] in
                self?.openAppStore()
            }
        )
    }
    
    private func createAndShowUpgradeView(title: String, message: String, onUpgrade: @escaping () -> Void, onCancel: (() -> Void)?) {
        guard let window = getTopWindow() else { return }
        
        currentWindow = window
        
        upgradeView = UpgradeView(frame: window.bounds)
        
        upgradeView?.onSure = {
            onUpgrade()
        }
        
        upgradeView?.onCancel = { [weak self] in
            guard let self = self else { return }
            onCancel?()
            self.hideUpgradeAlert()
        }
        
        window.addSubview(upgradeView!)
        
        upgradeView?.alpha = 0
        UIView.animate(withDuration: 0.3) {
            self.upgradeView?.alpha = 1
        }
    }
    
    private func dismissUpgradeView() {
        guard let upgradeView = upgradeView else { return }
        
        UIView.animate(withDuration: 0.3, animations: {
            upgradeView.alpha = 0
        }) { [weak self] _ in
            upgradeView.removeFromSuperview()
            self?.upgradeView = nil
            self?.currentWindow = nil
        }
    }
    
    private func getTopWindow() -> UIWindow? {
        if #available(iOS 13.0, *) {
            for scene in UIApplication.shared.connectedScenes {
                if let windowScene = scene as? UIWindowScene {
                    for window in windowScene.windows {
                        if window.isKeyWindow {
                            return window
                        }
                    }
                }
            }
            for scene in UIApplication.shared.connectedScenes {
                if let windowScene = scene as? UIWindowScene {
                    return windowScene.windows.first
                }
            }
        } else {
            return UIApplication.shared.keyWindow
        }
        return nil
    }
    
    private func openAppStore() {
        guard let url = URL(string: appStoreURL) else { return }
        if UIApplication.shared.canOpenURL(url) {
            UIApplication.shared.open(url, options: [:], completionHandler: nil)
        }
    }
}

extension UpgradeManager {
    
    func showForceUpgradeAlert(title: String = "App Update Required",
                               message: String = "A new version is available. Please update to continue using the app.") {
        showUpgradeAlert(title: title, message: message) { [weak self] in
            self?.openAppStore()
        }
    }
    
    func showOptionalUpgradeAlert(title: String = "New Version Available",
                                  message: String = "A new version is available with improvements and bug fixes.",
                                  onSkip: (() -> Void)? = nil) {
        showUpgradeAlert(title: title, message: message, onUpgrade: { [weak self] in
            self?.openAppStore()
            self?.hideUpgradeAlert()
        }, onCancel: { [weak self] in
            onSkip?()
            self?.hideUpgradeAlert()
        })
    }
}
