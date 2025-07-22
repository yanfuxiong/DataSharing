//
//  PictureInPictureManager.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/4/17.
//

import UIKit
import AVKit
import AVFoundation

class PictureInPictureManager: NSObject {
    static let shared = PictureInPictureManager()
    
    private let _pipContent = VideoProvider()
    private var _pipController: AVPictureInPictureController?
    private var _pipPossibleObservation: NSKeyValueObservation?
    private var backgroundTask: UIBackgroundTaskIdentifier = .invalid
    
    private var globalVideoContainer: UIView?
    private var hasSetAutoStartProperty = false
    private weak var currentActiveView: UIView?
    private var hasSetupActiveView = false
    
    private override init() {
        super.init()
        setupAudioSession()
        addNotifications()
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.1) {
            self.setupGlobalContainer()
        }
    }
    
    deinit {
        NotificationCenter.default.removeObserver(self)
        endBackgroundTask()
        globalVideoContainer?.removeFromSuperview()
    }
    
    private func setupGlobalContainer() {
        DispatchQueue.main.async { [weak self] in
            guard let self = self else { return }
            var targetWindow: UIWindow? = getCurrentActiveWindow()
            guard let window = targetWindow else {
                Logger.info("No available window found")
                return
            }
            self.globalVideoContainer = UIView(frame: CGRect(x: 0, y: 0, width: 320, height: 240))
            self.globalVideoContainer?.isHidden = true
            self.globalVideoContainer?.alpha = 0
            window.addSubview(self.globalVideoContainer!)
            Logger.info("Global video container created in active window")
        }
    }
    
    private func setupAudioSession() {
        if AVPictureInPictureController.isPictureInPictureSupported() {
            do {
                try AVAudioSession.sharedInstance().setCategory(.playback, options: .mixWithOthers)
                try AVAudioSession.sharedInstance().setActive(true)
            } catch {
                Logger.info("Audio session setup error: \(error.localizedDescription)")
            }
        }
    }
    
    func initializePIP() {
        guard AVPictureInPictureController.isPictureInPictureSupported() else {
            Logger.info("PIP not supported")
            return
        }
        Logger.info("Initializing PIP")
        _pipContent.start()
        DispatchQueue.main.async { [weak self] in
            self?.setupPIPController()
        }
        startBackgroundTask()
    }
    
    private func setupPIPController() {
        guard AVPictureInPictureController.isPictureInPictureSupported() else { return }
        if _pipController != nil && hasSetAutoStartProperty {
            return
        }
        if globalVideoContainer == nil {
            setupGlobalContainer()
        }
        if _pipController == nil {
            _pipController = AVPictureInPictureController(
                contentSource: .init(
                    sampleBufferDisplayLayer: _pipContent.bufferDisplayLayer,
                    playbackDelegate: self
                )
            )
            _pipController?.delegate = self
            _pipController?.setValue(2, forKey: "controlsStyle")
            if #available(iOS 14.2, *), !hasSetAutoStartProperty {
                _pipController?.canStartPictureInPictureAutomaticallyFromInline = true
                hasSetAutoStartProperty = true
            }
            Logger.info("PIP controller created")
        }
    }
    
    func getDisplayLayer() -> AVSampleBufferDisplayLayer {
        return _pipContent.bufferDisplayLayer
    }
    
    func setupDisplayLayer(for view: UIView) {
        DispatchQueue.main.async { [weak self] in
            guard let self = self else { return }
            Logger.info("Setting up display layer for view")
            self.currentActiveView = view
            self.hasSetupActiveView = true
            self._pipContent.bufferDisplayLayer.removeFromSuperlayer()
            self._pipContent.bufferDisplayLayer.frame = view.bounds
            self._pipContent.bufferDisplayLayer.videoGravity = .resizeAspect
            view.layer.addSublayer(self._pipContent.bufferDisplayLayer)
        }
    }
    
    private func getCurrentActiveWindow() -> UIWindow? {
        if #available(iOS 13.0, *) {
            return UIApplication.shared.connectedScenes
                .compactMap { $0 as? UIWindowScene }
                .filter { $0.activationState == .foregroundActive }
                .first?.windows
                .first { $0.isKeyWindow } ?? UIApplication.shared.windows.first
        } else {
            return UIApplication.shared.keyWindow ?? UIApplication.shared.windows.first
        }
    }
    
    private func ensureDisplayLayerInActiveWindow() {
        guard let activeWindow = getCurrentActiveWindow() else {
            Logger.info("No active window found")
            return
        }
        globalVideoContainer?.removeFromSuperview()
        globalVideoContainer = UIView(frame: CGRect(x: 0, y: 0, width: 320, height: 240))
        globalVideoContainer?.isHidden = true
        globalVideoContainer?.alpha = 0
        activeWindow.addSubview(globalVideoContainer!)
        _pipContent.bufferDisplayLayer.removeFromSuperlayer()
        _pipContent.bufferDisplayLayer.frame = globalVideoContainer!.bounds
        globalVideoContainer!.layer.addSublayer(_pipContent.bufferDisplayLayer)
        Logger.info("Display layer moved to active window")
    }
    
    func startPIPIfPossible() {
        DispatchQueue.main.async { [weak self] in
            guard let self = self else { return }
            Logger.info("Attempting to start PIP")
            self.ensureDisplayLayerInActiveWindow()
            DispatchQueue.main.asyncAfter(deadline: .now() + 0.2) {
                guard let pipController = self._pipController,
                      pipController.isPictureInPicturePossible else {
                    Logger.info("PIP not possible - controller: \(self._pipController != nil), possible: \(self._pipController?.isPictureInPicturePossible ?? false)")
                    if self._pipController != nil {
                        self._pipController = nil
                        self.hasSetAutoStartProperty = false
                        self.setupPIPController()
                        DispatchQueue.main.asyncAfter(deadline: .now() + 0.3) {
                            if let retryController = self._pipController,
                               retryController.isPictureInPicturePossible {
                                Logger.info("Retry starting PIP")
                                retryController.startPictureInPicture()
                            } else {
                                Logger.info("Retry failed - PIP still not possible")
                            }
                        }
                    }
                    return
                }
                Logger.info("Starting PIP")
                pipController.startPictureInPicture()
            }
        }
    }
    
    private func restoreDisplayLayerToActiveView() {
        DispatchQueue.main.async { [weak self] in
            guard let self = self,
                  let activeView = self.currentActiveView else {
                Logger.info("No active view to restore display layer")
                return
            }
            Logger.info("Restoring display layer to active view")
            self._pipContent.bufferDisplayLayer.removeFromSuperlayer()
            self._pipContent.bufferDisplayLayer.frame = activeView.bounds
            activeView.layer.addSublayer(self._pipContent.bufferDisplayLayer)
        }
    }
    
    func stopPIP() {
        guard let pipController = _pipController,
              pipController.isPictureInPictureActive else { return }
        Logger.info("Stopping PIP")
        pipController.stopPictureInPicture()
    }
    
    var isPictureInPictureActive: Bool {
        return _pipController?.isPictureInPictureActive ?? false
    }
    
    var isPictureInPicturePossible: Bool {
        return _pipController?.isPictureInPicturePossible ?? false
    }
    
    private func startBackgroundTask() {
        endBackgroundTask()
        backgroundTask = UIApplication.shared.beginBackgroundTask { [weak self] in
            self?.endBackgroundTask()
        }
    }
    
    private func endBackgroundTask() {
        if backgroundTask != .invalid {
            UIApplication.shared.endBackgroundTask(backgroundTask)
            backgroundTask = .invalid
        }
    }
    
    private func addNotifications() {
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(handleEnterBackground),
            name: UIApplication.didEnterBackgroundNotification,
            object: nil
        )
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(handleEnterForeground),
            name: UIApplication.willEnterForegroundNotification,
            object: nil
        )
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(handleWillTerminate),
            name: UIApplication.willTerminateNotification,
            object: nil
        )
    }
    
    @objc private func handleEnterBackground() {
        Logger.info("PictureInPictureManager - handleEnterBackground")
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.8) { [weak self] in
            guard let self = self else { return }
            do {
                try AVAudioSession.sharedInstance().setActive(true)
            } catch {
                Logger.info("Failed to activate audio session: \(error)")
            }
            if !self.isPictureInPictureActive {
                self.startPIPIfPossible()
            }
            self.startBackgroundTask()
        }
    }
    
    @objc private func handleEnterForeground() {
        Logger.info("PictureInPictureManager - handleEnterForeground")
        endBackgroundTask()
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.1) { [weak self] in
            guard let self = self else { return }
            if !self.isPictureInPictureActive {
                self.restoreDisplayLayerToActiveView()
            }
        }
    }
    
    @objc private func handleWillTerminate() {
        Logger.info("PictureInPictureManager - handleWillTerminate")
        endBackgroundTask()
    }
}

extension PictureInPictureManager: AVPictureInPictureControllerDelegate {
    func pictureInPictureController(
        _ pictureInPictureController: AVPictureInPictureController,
        failedToStartPictureInPictureWithError error: Error
    ) {
        Logger.info("PIP failed to start: \(error)")
    }
    
    func pictureInPictureControllerWillStartPictureInPicture(
        _ pictureInPictureController: AVPictureInPictureController
    ) {
        Logger.info("PIP will start")
    }
    
    func pictureInPictureControllerDidStartPictureInPicture(
        _ pictureInPictureController: AVPictureInPictureController
    ) {
        Logger.info("PIP did start")
    }
    
    func pictureInPictureControllerWillStopPictureInPicture(
        _ pictureInPictureController: AVPictureInPictureController
    ) {
        Logger.info("PIP will stop")
    }
    
    func pictureInPictureControllerDidStopPictureInPicture(
        _ pictureInPictureController: AVPictureInPictureController
    ) {
        Logger.info("PIP did stop")
        restoreDisplayLayerToActiveView()
    }
    
    func pictureInPictureController(
        _ pictureInPictureController: AVPictureInPictureController,
        restoreUserInterfaceForPictureInPictureStopWithCompletionHandler completionHandler: @escaping (Bool) -> Void
    ) {
        Logger.info("PIP restore user interface")
        DispatchQueue.main.async { [weak self] in
            if let tabBarController = UIApplication.shared.windows.first?.rootViewController as? UITabBarController {
                tabBarController.selectedIndex = 0
            }
            self?.restoreDisplayLayerToActiveView()
            completionHandler(true)
        }
    }
}

extension PictureInPictureManager: AVPictureInPictureSampleBufferPlaybackDelegate {
    func pictureInPictureController(
        _ pictureInPictureController: AVPictureInPictureController,
        setPlaying playing: Bool
    ) {
        Logger.info("PIP setPlaying: \(playing)")
    }
    
    func pictureInPictureControllerTimeRangeForPlayback(
        _ pictureInPictureController: AVPictureInPictureController
    ) -> CMTimeRange {
        return CMTimeRange(start: .negativeInfinity, duration: .positiveInfinity)
    }
    
    func pictureInPictureControllerIsPlaybackPaused(
        _ pictureInPictureController: AVPictureInPictureController
    ) -> Bool {
        return false
    }
    
    func pictureInPictureController(
        _ pictureInPictureController: AVPictureInPictureController,
        didTransitionToRenderSize newRenderSize: CMVideoDimensions
    ) {
        Logger.info("PIP render size: \(newRenderSize.width)x\(newRenderSize.height)")
    }
    
    func pictureInPictureController(
        _ pictureInPictureController: AVPictureInPictureController,
        skipByInterval skipInterval: CMTime,
        completion completionHandler: @escaping () -> Void
    ) {
        Logger.info("PIP skip by interval: \(skipInterval)")
        completionHandler()
    }
}
