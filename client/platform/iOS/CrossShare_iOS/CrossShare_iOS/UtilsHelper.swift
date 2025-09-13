//
//  UtilsHelper.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/8/14.
//

import UIKit

class UtilsHelper: NSObject {
    
    static let shared = UtilsHelper()
    
    func getTopWindow() -> UIWindow? {
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
}
