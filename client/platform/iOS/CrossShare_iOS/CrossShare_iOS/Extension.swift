//
//  Extension.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/4/28.
//

import UIKit
import MBProgressHUD

enum HUDType {
    case error
    case success
    case warn
}

extension MBProgressHUD {
    private static func getKeyWindow(from view: UIView?) -> UIWindow? {
        var window = view?.window
        if window == nil {
            window = findWindowInScene()
        }
        return window
    }

    private static func findWindowInScene() -> UIWindow? {
        let allScenes = UIApplication.value(forKey: "sharedApplication") as? UIApplication
        return allScenes?.connectedScenes
            .compactMap { $0 as? UIWindowScene }
            .flatMap { $0.windows }
            .first { $0.isKeyWindow }
    }

    private static func getBottomOffset(from view: UIView) -> CGFloat {
        guard let window = getKeyWindow(from: view) else {
            return MBProgressMaxOffset
        }
        return window.bounds.height / 2 - window.safeAreaInsets.bottom
    }

    private static func createCustomContentView(imageName: String, message: String) -> UIView {
        let imageView = UIImageView(image: UIImage(named: imageName))
        imageView.contentMode = .scaleAspectFit
        imageView.translatesAutoresizingMaskIntoConstraints = false
        NSLayoutConstraint.activate([
            imageView.widthAnchor.constraint(equalToConstant: 20),
            imageView.heightAnchor.constraint(equalToConstant: 20)
        ])

        let label = UILabel()
        label.text = message
        label.font = UIFont.systemFont(ofSize: 14)
        label.textColor = .white
        label.numberOfLines = 0
        label.textAlignment = .left

        let stackView = UIStackView(arrangedSubviews: [imageView, label])
        stackView.axis = .horizontal
        stackView.spacing = 8
        stackView.alignment = .center

        return stackView
    }

    static func showTips(_ type:HUDType = .success,_ message: String, toView view: UIView?, duration: TimeInterval = 2.0) {
        var tipsImage = "success"
        switch type {
        case .success:
            tipsImage = "success"
        case .error:
            tipsImage = "error"
        case .warn:
            tipsImage = "warn"
        }
        let hud = MBProgressHUD.showAdded(to: view ?? getKeyWindow(from: view)! , animated: true)
        hud.mode = .customView
        hud.customView = createCustomContentView(imageName: tipsImage, message: message)
        hud.label.text = nil
        hud.margin = 12
        hud.offset = .zero
        hud.bezelView.style = .solidColor
        hud.bezelView.color = UIColor.black.withAlphaComponent(0.75)
        hud.hide(animated: true, afterDelay: duration)
    }
}
