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
        // Priority 1: Find keyWindow in foreground active scene
        for scene in UIApplication.shared.connectedScenes {
            if let windowScene = scene as? UIWindowScene,
               windowScene.activationState == .foregroundActive {
                for window in windowScene.windows {
                    if window.isKeyWindow {
                        return window
                    }
                }
            }
        }
        
        // Priority 2: Return first window from foreground active scene
        for scene in UIApplication.shared.connectedScenes {
            if let windowScene = scene as? UIWindowScene,
               windowScene.activationState == .foregroundActive {
                return windowScene.windows.first
            }
        }
        
        // Priority 3: Fallback to any keyWindow
        for scene in UIApplication.shared.connectedScenes {
            if let windowScene = scene as? UIWindowScene {
                for window in windowScene.windows {
                    if window.isKeyWindow {
                        return window
                    }
                }
            }
        }
        
        // Priority 4: Return first available window
        for scene in UIApplication.shared.connectedScenes {
            if let windowScene = scene as? UIWindowScene {
                return windowScene.windows.first
            }
        }
        
        return nil
    }
}
