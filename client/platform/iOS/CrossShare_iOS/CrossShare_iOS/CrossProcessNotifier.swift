//
//  CrossProcessNotifier.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/7/29.
//

import UIKit

public final class CrossProcessNotifier {
    private let appGroupID: String
    private let sharedDefaults: UserDefaults
    private let center = CFNotificationCenterGetDarwinNotifyCenter()
    
    private lazy var observer: UnsafeMutableRawPointer = {
        return Unmanaged.passUnretained(self).toOpaque()
    }()
    
    private var handlers: [String: ([String: Any]?) -> Void] = [:]
    
    public init(appGroupID: String) {
        self.appGroupID = appGroupID
        guard let sharedDefaults = UserDefaults(suiteName: appGroupID) else {
            fatalError("Failed to initialize UserDefaults. Please check if App Group ID '\(appGroupID)' is correctly configured in the project.")
        }
        self.sharedDefaults = sharedDefaults
    }
    
    public func post(name: String, userInfo: [String: Any]? = nil) {
        let key = dataKey(for: name)
        if let userInfo = userInfo {
            sharedDefaults.set(userInfo, forKey: key)
        } else {
            sharedDefaults.removeObject(forKey: key)
        }
        let cfName = CFNotificationName(name as CFString)
        CFNotificationCenterPostNotification(center, cfName, nil, nil, true)
        print("[Notifier] Signal sent: \(name)")
    }
    
    public func observe(name: String, handler: @escaping ([String: Any]?) -> Void) {
        handlers[name] = handler
        let cfName = CFNotificationName(name as CFString)
        CFNotificationCenterAddObserver(center, observer, notificationCallback, cfName.rawValue, nil, .deliverImmediately)
        print("[Notifier] Started observing: \(name)")
    }
    
    public func stopObserving(name: String) {
        handlers.removeValue(forKey: name)
        let cfName = CFNotificationName(name as CFString)
        CFNotificationCenterRemoveObserver(center, observer, cfName, nil)
        print("[Notifier] Stopped observing: \(name)")
    }
    
    deinit {
        CFNotificationCenterRemoveEveryObserver(center, observer)
        print("[Notifier] All observers cleaned up.")
    }
    
    private func dataKey(for name: String) -> String {
        return "\(appGroupID).\(name).data"
    }
    
    fileprivate func handleNotification(name: String) {
        print("[Notifier] Notification received: \(name)")
        let key = dataKey(for: name)
        let userInfo = sharedDefaults.dictionary(forKey: key)
        if let handler = handlers[name] {
            handler(userInfo)
        }
    }
    
    private let notificationCallback: CFNotificationCallback = { center, observer, name, object, userInfo in
        guard
            let observer = observer,
            let name = name
        else { return }
        
        let notifier = Unmanaged<CrossProcessNotifier>.fromOpaque(observer).takeUnretainedValue()
        let swiftName = name.rawValue as String
        notifier.handleNotification(name: swiftName)
    }
}
