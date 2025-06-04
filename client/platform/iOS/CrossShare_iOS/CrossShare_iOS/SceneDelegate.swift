//
//  SceneDelegate.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/4/15.
//

import UIKit
import SwiftyJSON
import MBProgressHUD

class SceneDelegate: UIResponder, UIWindowSceneDelegate {
    
    var window: UIWindow?
    
    func scene(_ scene: UIScene, willConnectTo session: UISceneSession, options connectionOptions: UIScene.ConnectionOptions) {
        // Use this method to optionally configure and attach the UIWindow `window` to the provided UIWindowScene `scene`.
        // If using a storyboard, the `window` property will automatically be initialized and attached to the scene.
        // This delegate does not imply the connecting scene or session are new (see `application:configurationForConnectingSceneSession` instead).
        
        setupNavigationBarAppearance()
        guard let _ = (scene as? UIWindowScene) else { return }
    }
    
    func sceneDidDisconnect(_ scene: UIScene) {
        // Called as the scene is being released by the system.
        // This occurs shortly after the scene enters the background, or when its session is discarded.
        // Release any resources associated with this scene that can be re-created the next time the scene connects.
        // The scene may re-connect later, as its session was not necessarily discarded (see `application:didDiscardSceneSessions` instead).
    }
    
    func sceneDidBecomeActive(_ scene: UIScene) {
        // Called when the scene has moved from an inactive state to an active state.
        // Use this method to restart any tasks that were paused (or not yet started) when the scene was inactive.
        print("sceneDidBecomeActive")
        BackgroundTaskManager.shared.stopPlay()
    }
    
    func sceneWillResignActive(_ scene: UIScene) {
        // Called when the scene will move from an active state to an inactive state.
        // This may occur due to temporary interruptions (ex. an incoming phone call).
    }
    
    func sceneWillEnterForeground(_ scene: UIScene) {
        // Called as the scene transitions from the background to the foreground.
        // Use this method to undo the changes made on entering the background.
    }
    
    func sceneDidEnterBackground(_ scene: UIScene) {
        // Called as the scene transitions from the foreground to the background.
        // Use this method to save data, release shared resources, and store enough scene-specific state information
        // to restore the scene back to its current state.
        
        print("sceneDidEnterBackground")
        BackgroundTaskManager.shared.startPlay()
        
    }
    
    func scene(_ scene: UIScene, openURLContexts URLContexts: Set<UIOpenURLContext>) {
        guard let url = URLContexts.first?.url else {
            print("Error: No URL found in URLContexts.")
            return
        }
        
        guard url.scheme == "crossshare", url.host == "import" else {
            print("Error: URL scheme or host does not match. URL: \(url.absoluteString)")
            return
        }
        
        guard let components = URLComponents(url: url, resolvingAgainstBaseURL: true),
              let queryItems = components.queryItems else {
            print("Error: Could not get URL components or query items from URL: \(url.absoluteString)")
            return
        }
        
        var clientId: String?
        var clientIp: String?
        var filePath: String?
        var fileSizeString: String?
        
        for item in queryItems {
            switch item.name {
            case "clientId":
                clientId = item.value
            case "clientIp":
                clientIp = item.value
            case "filePath":
                filePath = item.value?.removingPercentEncoding ?? item.value
            case "fileSize":
                fileSizeString = item.value
            default:
                print("Warning: Unknown query parameter '\(item.name)' in URL: \(url.absoluteString)")
            }
        }
        
        guard let finalClientId = clientId else {
            print("Error: Missing 'clientId' parameter in URL: \(url.absoluteString)")
            return
        }
        
        guard let finalClientIp = clientIp else {
            print("Error: Missing 'clientIp' parameter in URL: \(url.absoluteString)")
            return
        }
        
        guard let finalFilePath = filePath,
              let finalFileSizeString = fileSizeString,
              let finalFileSize = Int64(finalFileSizeString) else {
            var missingParams: [String] = []
            if filePath == nil { missingParams.append("'filePath'") }
            if fileSizeString == nil { missingParams.append("'fileSize'") }
            if filePath != nil && fileSizeString != nil && Int64(fileSizeString!) == nil {
                missingParams.append("valid 'fileSize'")
            }
            print("Error: Missing or invalid parameters (\(missingParams.joined(separator: ", "))) for single file in URL: \(url.absoluteString)")
            return
        }
        print("Processing Single File: Path: \(finalFilePath), Client ID: \(finalClientId), File Size: \(finalFileSize)")
        // TODO: consider multiple files transfer
        let fileMsg = "1 file"
        let clientName = P2PManager.shared.getClientInfo(id: clientId ?? "")?.name ?? ""
        let params: [String] = [fileMsg, clientName]
        PushNotiManager.shared.sendLocalNotification(code: .sendStart, with: params)
        P2PManager.shared.setFileDropRequest(filePath: finalFilePath, id: finalClientId, fileSize: finalFileSize)
        print("Successfully processed single file URL: \(url.absoluteString)")
        //   if let view = window?.rootViewController?.view {
        //       MBProgressHUD.showSuccess("Sended file from \(finalClientId)", toView: view)
        //   }
    }
}

extension SceneDelegate {
    func setupNavigationBarAppearance() {
        if #available(iOS 13.0, *) {
            let appearance = UINavigationBarAppearance()
            appearance.configureWithOpaqueBackground()
            appearance.backgroundColor = UIColor.systemBlue
            appearance.titleTextAttributes = [.foregroundColor: UIColor.white]
            appearance.largeTitleTextAttributes = [.foregroundColor: UIColor.white]
            
            UINavigationBar.appearance().standardAppearance = appearance
            UINavigationBar.appearance().scrollEdgeAppearance = appearance
            UINavigationBar.appearance().compactAppearance = appearance
            UINavigationBar.appearance().tintColor = UIColor.white
        } else {
            UINavigationBar.appearance().barTintColor = UIColor.systemBlue
            UINavigationBar.appearance().tintColor = UIColor.white
            UINavigationBar.appearance().titleTextAttributes = [NSAttributedString.Key.foregroundColor: UIColor.white]
            UINavigationBar.appearance().isTranslucent = false
        }
    }
}
