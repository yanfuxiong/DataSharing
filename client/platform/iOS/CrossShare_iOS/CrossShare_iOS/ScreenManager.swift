//
//  ScrenManager.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/6/30.
//

import UIKit

class ScreenManager: NSObject {
    
    static let shared = ScreenManager()
    private var discoveryTimer: Timer?
    
    var isStartMonitoring: Bool = false
    var currentScreen: UIScreen?
    var currentScreenBounds: CGRect {
        return currentScreen?.bounds ?? UIScreen.main.bounds
    }
    
    private var lastAuthData: AuthData?
    
    override init() {
        super.init()
        setupScreenConnectionObservers()
        checkForExistingExternalScreen()
    }
    
    public func startMonitoring() {
        guard !isStartMonitoring else { return }
        isStartMonitoring = true
        Logger.info("开始监控屏幕连接状态...")
        startPeriodicDiscovery()
    }
    
    func  startPeriodicDiscovery() {
        guard self.discoveryTimer != nil else {
            self.discoveryTimer?.invalidate()
            return
        }
        self.discoveryTimer = Timer(timeInterval: 5,
                                    target: self,
                                    selector: #selector(checkForExistingExternalScreen),
                                    userInfo: nil,
                                    repeats: true)
        RunLoop.current.add(self.discoveryTimer!, forMode: RunLoop.Mode.common)
    }
    
    func stopPeriodicDiscovery() {
        self.discoveryTimer?.invalidate()
        self.discoveryTimer = nil;
    }
    
    func setupScreenConnectionObservers() {
        let center = NotificationCenter.default
        center.addObserver(self, selector: #selector(screenDidConnect(_:)), name: UIScreen.didConnectNotification, object: nil)
        center.addObserver(self, selector: #selector(screenDidDisconnect(_:)), name: UIScreen.didDisconnectNotification, object: nil)
    }
    
    @objc func checkForExistingExternalScreen() {
        if UIScreen.screens.count > 1 {
            Logger.info("App启动时，已发现外接屏幕。")
            if let externalScreen = UIScreen.screens.first(where: { $0 != UIScreen.main }) {
                logScreenDetailsIfChanged(screen: externalScreen)
            }
        } else {
            Logger.info("App启动，未发现外接屏幕。")
        }
    }
    
    @objc func screenDidConnect(_ notification: NSNotification) {
        var externalScreen: Bool = true
        guard let newScreen = notification.object as? UIScreen else {
            externalScreen = false
            P2PManager.shared.detectPluginEventCallback(isPlugin: externalScreen)
            return
        }
        P2PManager.shared.detectPluginEventCallback(isPlugin: externalScreen)
        logScreenDetails(screen: newScreen)
    }
    
    @objc func screenDidDisconnect(_ notification: NSNotification) {
        Logger.info("current screen count : \(UIScreen.screens.count)")
        lastAuthData = nil
        P2PManager.shared.detectPluginEventCallback(isPlugin: false)
    }
    
    func logScreenDetailsIfChanged(screen: UIScreen) {
        let displayWidth = screen.bounds.width
        let displayHeight = screen.bounds.height
        let fps = screen.maximumFramesPerSecond
        
        let newAuthData = AuthData(width: Int(displayWidth),
                                   height: Int(displayHeight),
                                   framerate: Int(fps),
                                   type: 1)
        
        if lastAuthData == newAuthData {
            Logger.info("ScreenDetailsIfChanged: No change detected, skipping log.")
            return
        }
        
        lastAuthData = newAuthData
        logScreenDetails(screen: screen)
    }
    
    func logScreenDetails(screen: UIScreen) {
        Logger.info("screen.description: \(screen.description)")
        Logger.info("Bounds: \(screen.bounds)")
        Logger.info("Native Bounds : \(screen.nativeBounds)")
        Logger.info("Scale : \(screen.scale)")
        Logger.info("Native Scale : \(screen.nativeScale)")
        
        Logger.info("(FPS): \(screen.maximumFramesPerSecond) Hz")
        
        if let currentMode = screen.currentMode {
            Logger.info(" (Current Mode): \(currentMode.size.width)x\(currentMode.size.height), Pixel Aspect Ratio: \(currentMode.pixelAspectRatio)")
        }
        Logger.info("---  (Available Modes) ---")
        // for (index, mode) in screen.availableModes.enumerated() {
        //     Logger.info(" Mode \(index): \(mode.size.width)x\(mode.size.height), PAR: \(mode.pixelAspectRatio)")
        // }
        
        if let preferredMode = screen.preferredMode {
            Logger.info("(Preferred Mode): \(preferredMode.size.width)x\(preferredMode.size.height)")
        }
        Logger.info("is (Mirrored): \(screen.mirrored != nil)")
        // Logger.info("(Brightness): \(screen.brightness)")
        
        let displayWidth = screen.bounds.width
        let displayHeight = screen.bounds.height
        let fps = screen.maximumFramesPerSecond
        
        let authData = AuthData(width: Int(displayWidth),
                                height: Int(displayHeight),
                                framerate: Int(fps),
                                type: 1)
        
        let enCoder = JSONEncoder()
        enCoder.outputFormatting = .prettyPrinted
        if let jsonData = try? enCoder.encode(authData),
           let jsonString = String(data: jsonData, encoding: .utf8) {
            Logger.info("AuthData: \(jsonString)")
            P2PManager.shared.setScreenData(screen: jsonString)
        }
    }
}
