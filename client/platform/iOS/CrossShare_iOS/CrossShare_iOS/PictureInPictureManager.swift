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
    
    private let _pipContent = PIPStatusView()
    private var _pipController: AVPictureInPictureController?
    private var _pipPossibleObservation: NSKeyValueObservation?
    private var backgroundTask: UIBackgroundTaskIdentifier = .invalid
    
    private var globalVideoContainer: UIView?
    private var hasSetAutoStartProperty = false
    private weak var currentActiveView: UIView?
    private var hasSetupActiveView = false
    private var mPIPContentType: PIPContentType = .idle
    
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
            let targetWindow: UIWindow? = UtilsHelper.shared.getTopWindow()
            guard let window = targetWindow else {
                Logger.info("No available window found")
                return
            }
            self.globalVideoContainer = UIView(frame: CGRect(x: 0, y: 0, width: 320, height: 50))
            self.globalVideoContainer?.isHidden = true
            self.globalVideoContainer?.alpha = 0
            window.addSubview(self.globalVideoContainer!)
            // Logger.info("Global video container created in active window")
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
        // Logger.info("Initializing PIP")
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
        
        // Make sure the global container exists
        if globalVideoContainer == nil {
            setupGlobalContainer()
        }
        
        // Make sure the display layer is generating content
        if !_pipContent.isRunning() {
            _pipContent.start()
        }
        
        if _pipController == nil {
            // Set display layer properties
            _pipContent.bufferDisplayLayer.videoGravity = .resizeAspectFill
            _pipContent.bufferDisplayLayer.backgroundColor = UIColor.clear.cgColor
            
            // Logger.info("Creating PIP controller with display layer")
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
            Logger.info("PIP controller created successfully")
        }
    }
    
    func getDisplayLayer() -> AVSampleBufferDisplayLayer {
        return _pipContent.bufferDisplayLayer
    }
    
    func setupDisplayLayer(for view: UIView) {
        DispatchQueue.main.async { [weak self] in
            guard let self = self else { return }
            // Logger.info("Setting up display layer for view")
            self.currentActiveView = view
            self.hasSetupActiveView = true
            self._pipContent.bufferDisplayLayer.removeFromSuperlayer()
            self._pipContent.bufferDisplayLayer.frame = view.bounds
            self._pipContent.bufferDisplayLayer.videoGravity = .resizeAspect
            view.layer.addSublayer(self._pipContent.bufferDisplayLayer)
        }
    }
    
    private func ensureDisplayLayerInActiveWindow() {
        guard let activeWindow = UtilsHelper.shared.getTopWindow() else {
            Logger.info("No active window found")
            return
        }
        globalVideoContainer?.removeFromSuperview()
        globalVideoContainer = UIView(frame: CGRect(x: 0, y: 0, width: 320, height: 50))
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
            // Logger.info("Attempting to start PIP")
            
            // Make sure the display layer is in the correct window
            self.ensureDisplayLayerInActiveWindow()
            
            // Check PIP controller status
            guard let pipController = self._pipController else {
                Logger.info("PIP controller is nil, recreating...")
                self.setupPIPController()
                DispatchQueue.main.asyncAfter(deadline: .now() + 0.2) {
                    self.startPIPIfPossible()
                }
                return
            }
            
            // Logger.info("PIP controller status - isPictureInPicturePossible: \(pipController.isPictureInPicturePossible)")
            // Logger.info("PIP controller status - isPictureInPictureActive: \(pipController.isPictureInPictureActive)")
            // Logger.info("Audio session status - isActive: \(AVAudioSession.sharedInstance().isOtherAudioPlaying)")
            
            guard pipController.isPictureInPicturePossible else {
                Logger.info("PIP not possible, retrying...")
                self._pipController = nil
                self.hasSetAutoStartProperty = false
                self.setupPIPController()
                if let retryController = self._pipController,
                   retryController.isPictureInPicturePossible {
                    Logger.info("Retry starting PIP after recreation")
                    retryController.startPictureInPicture()
                } else {
                    Logger.info("Retry failed - PIP still not possible")
                }
                return
            }
            
            Logger.info("Starting PIP now...")
            pipController.startPictureInPicture()
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
    
    // MARK: - PIP Content Update Methods
    
    func updatePIPStatus(contentType: PIPContentType, text: String? = nil, image: UIImage? = nil) {
        _pipContent.updateStatus(contentType: contentType, text: text, image: image)
        mPIPContentType = contentType
    }
    
    func showTextReceived(_ text: String) {
        guard mPIPContentType == .idle else {
            return
        }
        _pipContent.updateStatus(contentType: .textReceived, text: text)
    }
    
    func showImageReceived(_ image: UIImage) {
        guard mPIPContentType == .idle else {
            return
        }
        _pipContent.updateStatus(contentType: .imageReceived, image: image)
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
        // Logger.info("PictureInPictureManager - handleEnterBackground - Start PIP")
        do {
            try AVAudioSession.sharedInstance().setCategory(.playback, options: .mixWithOthers)
            try AVAudioSession.sharedInstance().setActive(true)
            Logger.info("Audio session activated successfully")
        } catch {
            Logger.info("Failed to activate audio session: \(error)")
        }
        
        // Make sure the PIP content is running
        if !self._pipContent.isRunning() {
            Logger.info("Starting PIP content generation")
            self._pipContent.start()
        }
        
        // Wait for the content to be generated and start PIP
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.2) {
            if !self.isPictureInPictureActive {
                self.startPIPIfPossible()
            }
        }
        
        self.startBackgroundTask()
    }
    
    @objc private func handleEnterForeground() {
        Logger.info("PictureInPictureManager - handleEnterForeground - Stop PIP")
        endBackgroundTask()
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.5) { [weak self] in
            guard let self = self else { return }
            // Stop PIP in the foreground and hide the PIP window
            if self.isPictureInPictureActive {
                self.stopPIP()
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
        Logger.info("Error domain: \((error as NSError).domain)")
        Logger.info("Error code: \((error as NSError).code)")
        Logger.info("Error description: \(error.localizedDescription)")
        
        if (error as NSError).domain == "PGPegasusErrorDomain" && (error as NSError).code == -1003 {
            Logger.info("Content source error detected, restarting content generation...")
            DispatchQueue.main.async { [weak self] in
                guard let self = self else { return }
                
                self._pipContent.stop()
                DispatchQueue.main.asyncAfter(deadline: .now() + 0.3) {
                    self._pipContent.start()
                    
                    self._pipController = nil
                    self.hasSetAutoStartProperty = false
                    DispatchQueue.main.asyncAfter(deadline: .now() + 0.3) {
                        self.setupPIPController()
                        if let controller = self._pipController,
                           controller.isPictureInPicturePossible {
                            Logger.info("Retrying PIP start after content restart")
                            controller.startPictureInPicture()
                        }
                    }
                }
            }
        }
    }
    
    func pictureInPictureControllerWillStartPictureInPicture(
        _ pictureInPictureController: AVPictureInPictureController
    ) {
        // Logger.info("PIP will start")
    }
    
    func pictureInPictureControllerDidStartPictureInPicture(
        _ pictureInPictureController: AVPictureInPictureController
    ) {
        // Logger.info("PIP did start")
    }
    
    func pictureInPictureControllerWillStopPictureInPicture(
        _ pictureInPictureController: AVPictureInPictureController
    ) {
        // Logger.info("PIP will stop")
    }
    
    func pictureInPictureControllerDidStopPictureInPicture(
        _ pictureInPictureController: AVPictureInPictureController
    ) {
        // Logger.info("PIP did stop")
        restoreDisplayLayerToActiveView()
    }
    
    func pictureInPictureController(
        _ pictureInPictureController: AVPictureInPictureController,
        restoreUserInterfaceForPictureInPictureStopWithCompletionHandler completionHandler: @escaping (Bool) -> Void
    ) {
        // Logger.info("PIP restore user interface")
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
