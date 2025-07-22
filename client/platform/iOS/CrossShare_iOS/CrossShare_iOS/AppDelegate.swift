//
//  AppDelegate.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/4/15.
//

import UIKit

@main
class AppDelegate: UIResponder, UIApplicationDelegate, UNUserNotificationCenterDelegate {
    
    func application(_ application: UIApplication, didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?) -> Bool {
        UNUserNotificationCenter.current().delegate = self
        PushNotiManager.shared.initNoti()
        CSNetworkAccessibility.sharedInstance().start()
        CSNetworkAccessibility.sharedInstance().setAlertEnable(true)
        NotificationCenter.default.addObserver(self, selector: #selector(netWorkChanged(_:)), name: CSNetworkAccessibilityChangedNotification, object: nil)
        ClipboardMonitor.shareInstance().startMonitoring()
        ScreenManager.shared.startMonitoring()
        return true
    }

    // MARK: UISceneSession Lifecycle
    func application(_ application: UIApplication, configurationForConnecting connectingSceneSession: UISceneSession, options: UIScene.ConnectionOptions) -> UISceneConfiguration {
        // Called when a new scene session is being created.
        // Use this method to select a configuration to create the new scene with.
        return UISceneConfiguration(name: "Default Configuration", sessionRole: connectingSceneSession.role)
    }

    func application(_ application: UIApplication, didDiscardSceneSessions sceneSessions: Set<UISceneSession>) {
        // Called when the user discards a scene session.
        // If any sessions were discarded while the application was not running, this will be called shortly after application:didFinishLaunchingWithOptions.
        // Use this method to release any resources that were specific to the discarded scenes, as they will not return.
    }
}

extension AppDelegate {
    @objc private func netWorkChanged(_ ntf: Notification) {
        Logger.info("AppDelegate - Network notification received: \(ntf)")
        
        let status = CSNetworkAccessibility.sharedInstance().currentState()
        Logger.info("AppDelegate - Current network status: \(status.rawValue)")
        
        switch status {
        case .checking:
            Logger.info("Network status: checking")
        case .unknown:
            Logger.info("Network status: unknown")
        case .accessible, .accessibleWiFi, .accessibleCellular:
            Logger.info("Network status: accessible")
            CSNetworkAccessibility.sharedInstance().initializeApp { success in
                Logger.info("App initialization result: \(success)")
            }
        case .restricted:
            Logger.info("Network status: restricted")
        }
    }
}

extension AppDelegate {
    func userNotificationCenter(_ center: UNUserNotificationCenter,
                                willPresent notification: UNNotification,
                                withCompletionHandler completionHandler: @escaping (UNNotificationPresentationOptions) -> Void) {
        completionHandler([.banner, .list, .sound])
    }
}
