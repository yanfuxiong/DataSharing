//
//  CSNetworkAccessibility.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/4/27.
//

import Foundation
import SystemConfiguration
import CoreTelephony
import UIKit

public let CSNetworkAccessibilityChangedNotification = Notification.Name("CSNetworkAccessibilityChangedNotification")

public enum CSNetworkAccessibleState: Int {
    case checking = 0
    case unknown
    case accessible
    case restricted
    case accessibleWiFi
    case accessibleCellular
}

public typealias NetworkAccessibleStateNotifier = (CSNetworkAccessibleState) -> Void

public class CSNetworkAccessibility {
    
    private var reachabilityRef: SCNetworkReachability?
    private var cellularData: CTCellularData?
    private var becomeActiveCallbacks: [() -> Void] = []
    private var previousState: CSNetworkAccessibleState = .checking
    private var alertController: UIAlertController?
    private var automaticallyAlert: Bool = false
    private var networkAccessibleStateDidUpdateNotifier: NetworkAccessibleStateNotifier?
    private var checkActiveLaterWhenDidBecomeActive = false
    private var checkingActiveLater = false
    private var isInitialized = false
    private static let shared = CSNetworkAccessibility()
    
    private init() { }
    
    public static func sharedInstance() -> CSNetworkAccessibility{
        return shared
    }
    
    public func start() {
        CSNetworkAccessibility.sharedInstance().setupNetworkAccessibility()
    }
    
    public func stop() {
        CSNetworkAccessibility.sharedInstance().cleanNetworkAccessibility()
    }
    
    public func setAlertEnable(_ setAlertEnable: Bool) {
        CSNetworkAccessibility.sharedInstance().automaticallyAlert = setAlertEnable
    }
    
    public func setStateDidUpdateNotifier(_ block: @escaping NetworkAccessibleStateNotifier) {
        CSNetworkAccessibility.sharedInstance().monitorNetworkAccessibleState(with: block)
    }
    
    public func currentState() -> CSNetworkAccessibleState {
        return CSNetworkAccessibility.sharedInstance().previousState
    }
    
    private func setupNetworkAccessibility() {
        if reachabilityRef != nil || cellularData != nil { return }
        
        if Float(UIDevice.current.systemVersion) ?? 15.0 < 10.0 || isSimulator() {
            notiWithAccessibleState(.accessible)
            return
        }
        
        NotificationCenter.default.addObserver(self, selector: #selector(applicationWillResignActive), name: UIApplication.willResignActiveNotification, object: nil)
        NotificationCenter.default.addObserver(self, selector: #selector(applicationDidBecomeActive), name: UIApplication.didBecomeActiveNotification, object: nil)
        
        reachabilityRef = SCNetworkReachabilityCreateWithName(nil, "223.5.5.5")
        SCNetworkReachabilityScheduleWithRunLoop(reachabilityRef!, CFRunLoopGetCurrent(), CFRunLoopMode.commonModes.rawValue)
        
        becomeActiveCallbacks = []
        
        let firstRun = !UserDefaults.standard.bool(forKey: "CSNetworkAccessibilityRunFlag")
        if firstRun {
            UserDefaults.standard.set(true, forKey: "CSNetworkAccessibilityRunFlag")
            DispatchQueue.main.asyncAfter(deadline: .now() + 3.0) {
                self.waitActive {
                    self.startReachabilityNotifier()
                    self.startCellularDataNotifier()
                }
            }
        } else {
            startReachabilityNotifier()
            startCellularDataNotifier()
        }
    }
    
    private func cleanNetworkAccessibility() {
        NotificationCenter.default.removeObserver(self)
        
        cellularData?.cellularDataRestrictionDidUpdateNotifier = nil
        cellularData = nil
        
        if reachabilityRef != nil {
            SCNetworkReachabilityUnscheduleFromRunLoop(reachabilityRef!, CFRunLoopGetMain(), CFRunLoopMode.commonModes.rawValue)
        }
        reachabilityRef = nil
        
        becomeActiveCallbacks.removeAll()
        previousState = .checking
        
        checkActiveLaterWhenDidBecomeActive = false
        checkingActiveLater = false
    }
    
    private func monitorNetworkAccessibleState(with completionBlock: @escaping NetworkAccessibleStateNotifier) {
        networkAccessibleStateDidUpdateNotifier = completionBlock
    }
    
    func initializeApp(completion: @escaping (Bool) -> Void) {
        switch CSNetworkAccessibility.sharedInstance().currentState() {
        case .checking :
            completion(false)
        case .accessible, .accessibleWiFi, .accessibleCellular:
            initializeSDKs { success in
                if success {
                    self.isInitialized = true
                    completion(true)
                } else {
                    completion(false)
                }
            }
        case .restricted:
            showNetworkRestrictedAlert()
            completion(false)
            
        case .unknown:
            completion(false)
        }
    }
    
    private func initializeSDKs(completion: @escaping (Bool) -> Void) {
        let shareInstance = WifiManager.shareInstance()
        shareInstance.getNetInfoFromLocalIp { netname, index in
            guard let netname = netname , let index = index else {
                return
            }
            SendNetInterfaces(netname.toGoString(),"".toGoString(),GoInt(0), GoInt(index),GoUint(0))
            DispatchQueue.main.asyncAfter(deadline: .now() + 0.5) {
                P2PManager.shared.startP2PService()
                completion(true)
            }
        }
    }
    
    @objc private func applicationWillResignActive() {
        hideNetworkRestrictedAlert()
        
        if checkingActiveLater {
            cancelEnsureActive()
            checkActiveLaterWhenDidBecomeActive = true
        }
    }
    
    @objc private func applicationDidBecomeActive() {
        if checkActiveLaterWhenDidBecomeActive {
            checkActiveLater()
            checkActiveLaterWhenDidBecomeActive = false
        }
        if !isInitialized {
            let newStatus = CSNetworkAccessibility.sharedInstance().currentState()
            if newStatus != CSNetworkAccessibleState.accessible {
                initializeApp { _ in }
            }
        }
    }
    
    private func waitActive(_ block: @escaping () -> Void) {
        becomeActiveCallbacks.append(block)
        if UIApplication.shared.applicationState != .active {
            checkActiveLaterWhenDidBecomeActive = true
        } else {
            checkActiveLater()
        }
    }
    
    private func checkActiveLater() {
        checkingActiveLater = true
        DispatchQueue.main.asyncAfter(wallDeadline: .now() + 2) { [weak self] in
            self?.ensureActive()
        }
    }
    
    @objc private func ensureActive() {
        checkingActiveLater = false
        for callback in becomeActiveCallbacks {
            callback()
        }
        becomeActiveCallbacks.removeAll()
    }
    
    private func cancelEnsureActive() {
        NSObject.cancelPreviousPerformRequests(withTarget: self, selector: #selector(ensureActive), object: nil)
    }
    
    private func startReachabilityNotifier() {
        Logger.info("CSNetworkAccessibility - setting up reachability notifier")
        
        var context = SCNetworkReachabilityContext(
            version: 0,
            info: UnsafeMutableRawPointer(Unmanaged.passUnretained(self).toOpaque()),
            retain: nil,
            release: nil,
            copyDescription: nil
        )
        
        let callbackSet = SCNetworkReachabilitySetCallback(reachabilityRef!, { (_, flags, info) in
            Logger.info("CSNetworkAccessibility - reachability callback triggered with flags: \(flags)")
            let networkAccessibility = Unmanaged<CSNetworkAccessibility>.fromOpaque(info!).takeUnretainedValue()
            DispatchQueue.main.async {
                networkAccessibility.startCheck()
            }
        }, &context)
        
        if !callbackSet {
            Logger.info("CSNetworkAccessibility - failed to set reachability callback")
        }
        
        let scheduled = SCNetworkReachabilityScheduleWithRunLoop(reachabilityRef!, CFRunLoopGetCurrent(), CFRunLoopMode.defaultMode.rawValue)
        if !scheduled {
            Logger.info("CSNetworkAccessibility - failed to schedule reachability with run loop")
        } else {
            Logger.info("CSNetworkAccessibility - reachability scheduled successfully")
        }
    }
    
    private func startCellularDataNotifier() {
        Logger.info("CSNetworkAccessibility - setting up cellular data notifier")
        
        cellularData = CTCellularData()
        cellularData?.cellularDataRestrictionDidUpdateNotifier = { [weak self] state in
            Logger.info("CSNetworkAccessibility - cellular data restriction state: \(state.rawValue)")
            DispatchQueue.main.async {
                self?.startCheck()
            }
        }
    }
    
    private func startCheck() {
        Logger.info("CSNetworkAccessibility - startCheck called")
        
        if currentReachable() {
            let networkType = getCurrentNetworkType()
            Logger.info("Network reachable, type: \(networkType)")
            
            switch networkType {
            case "WiFi":
                notiWithAccessibleState(.accessibleWiFi)
            case "Cellular":
                notiWithAccessibleState(.accessibleCellular)
            default:
                notiWithAccessibleState(.accessible)
            }
        } else {
            Logger.info("Network not reachable")
            checkCellularDataAccess()
        }
    }
    
    private func getCurrentNetworkType() -> String {
        var flags: SCNetworkReachabilityFlags = []
        if SCNetworkReachabilityGetFlags(reachabilityRef!, &flags) {
            if flags.contains(.isWWAN) {
                return "Cellular"
            } else if flags.contains(.reachable) {
                return "WiFi"
            }
        }
        return "Unknown"
    }
    
    private func checkCellularDataAccess() {
        guard let state = cellularData?.restrictedState else { return }
        
        switch state {
        case .restricted:
            notiWithAccessibleState(.restricted)
        case .notRestricted:
            notiWithAccessibleState(.accessible)
        case .restrictedStateUnknown:
            DispatchQueue.main.asyncAfter(deadline: .now() + 0.1) {
                self.startCheck()
            }
        @unknown default:
            break
        }
    }
    
    private func currentReachable() -> Bool {
        var flags: SCNetworkReachabilityFlags = []
        if SCNetworkReachabilityGetFlags(reachabilityRef!, &flags) {
            let isReachable = flags.contains(.reachable)
            Logger.info("Network flags: \(flags), reachable: \(isReachable)")
            return isReachable
        }
        Logger.info("Failed to get network flags")
        return false
    }
    
    private func notiWithAccessibleState(_ state: CSNetworkAccessibleState) {
        Logger.info("CSNetworkAccessibility - state change: \(previousState.rawValue) -> \(state.rawValue)")
        
        if automaticallyAlert {
            if state == .restricted {
                showNetworkRestrictedAlert()
            } else {
                hideNetworkRestrictedAlert()
            }
        }
        
        let shouldNotify = (state != previousState) ||
        (state == .accessibleWiFi && previousState == .accessibleCellular) ||
        (state == .accessibleCellular && previousState == .accessibleWiFi)
        
        if shouldNotify {
            Logger.info("CSNetworkAccessibility - sending notification for state: \(state.rawValue)")
            previousState = state
            networkAccessibleStateDidUpdateNotifier?(state)
            NotificationCenter.default.post(name: CSNetworkAccessibilityChangedNotification, object: nil)
        } else {
            Logger.info("CSNetworkAccessibility - no notification sent, same state")
        }
    }
    
    private func showNetworkRestrictedAlert() {
        let alert = UIAlertController(
            title: "Network permission not authorized",
            message: "The app requires network permissions to run properly, please go to Settings to enable network permissions.",
            preferredStyle: .alert
        )
        
        alert.addAction(UIAlertAction(
            title: "Cancel",
            style: .cancel
        ) { _ in
            exit(0)
        })
        
        alert.addAction(UIAlertAction(
            title: "Go to Settings",
            style: .default
        ) { _ in
            if let url = URL(string: UIApplication.openSettingsURLString) {
                UIApplication.shared.open(url, options: [:], completionHandler: nil)
            }
        })
        
        if let windowScene = UIApplication.shared.connectedScenes.first as? UIWindowScene,
           let rootViewController = windowScene.windows.first?.rootViewController {
            rootViewController.present(alert, animated: true)
        }
    }
    
    private func hideNetworkRestrictedAlert() {
        alertController?.dismiss(animated: true, completion: nil)
    }
    
    private func isSimulator() -> Bool {
#if targetEnvironment(simulator)
        return true
#else
        return false
#endif
    }
}
