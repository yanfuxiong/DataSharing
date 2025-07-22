//
//  ViewController.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/4/15.
//

import UIKit
import AVKit
import AVFoundation
import SnapKit
import MBProgressHUD

enum ViewType {
    case main
    case config
    case file
}

class ViewController: UIViewController {
    private let _pipContent = VideoProvider()
    private var _pipController: AVPictureInPictureController?
    private let _bufferDisplayLayer = AVSampleBufferDisplayLayer()
    private var _pipPossibleObservation: NSKeyValueObservation?
    private var dataArray:[DownloadItem] = []
    private var devicePopView: DeviceSelectPopView?
    private var selectedClient:ClientInfo?
    private var userManuallySwitchedFromFile = false

    private var lastViewType: ViewType? = nil
    private var viewType:ViewType = ViewType.main {
        didSet {
            if lastViewType != nil {
                guard lastViewType != viewType else {
                    return
                }
            }

            lastViewType = viewType
            switch viewType {
            case .main:
                Logger.info("switch main")
                DispatchQueue.main.async { [weak self] in
                    guard let self = self else { return  }
                    self.configContenView.removeFromSuperview()
                    self.fileContennerView.removeFromSuperview()
                    self.view.addSubview(self.deviceView)
                    
                    self.deviceView.snp.makeConstraints { make in
                        make.left.right.equalToSuperview()
                        make.top.equalTo(self.videoContainerView.snp.bottom)
                        make.bottom.equalToSuperview()
                    }
                    
                    self.deviceView.refreshUI()
//                    self.deviceView.dosomothingBlock = { [weak self] index in
//                        guard let self = self else { return  }
//                        switch index {
//                        case 1:
//                            self.aheadButton.isSelected = true
//                            self.viewType = .config
//                        case 2:
//                            self.pickDocument()
//                        case 3:
//                            self.aheadButton.isSelected = true
//                            self.viewType = .file
//                        default:
//                            Logger.info("do nothing but else")
//                        }
//                    }
                    self.view.setNeedsLayout()
                    self.view.layoutIfNeeded()
                }
            case .config:
                Logger.info("switch config")
                DispatchQueue.main.async { [weak self] in
                    guard let self = self else { return  }
                    self.fileContennerView.removeFromSuperview()
                    self.deviceView.removeFromSuperview()
                    self.view.addSubview(self.configContenView)
                    
                    self.configContenView.snp.makeConstraints { make in
                        make.left.right.equalToSuperview()
                        make.top.equalTo(self.videoContainerView.snp.bottom)
                        make.bottom.equalToSuperview()
                    }
                    
                    self.configContenView.refreshUI()
                    self.configContenView.submitConfigBlock = {[weak self] (ddcciText,deviceSourceText,devicePortText) in
                        guard let self = self else { return  }
                        guard ddcciText.isEmpty == false else {
                            MBProgressHUD.showTips(.error,"Please enter device ddcci id", toView: self.view)
                            return
                        }
                        guard deviceSourceText.isEmpty == false else {
                            MBProgressHUD.showTips(.error,"Please enter device source", toView: self.view)
                            return
                        }
                        guard devicePortText.isEmpty == false else {
                            MBProgressHUD.showTips(.error,"Please enter device port", toView: self.view)
                            return
                        }
                        if let intSource = Int(deviceSourceText), let intPort = Int(devicePortText) {
                            UserDefaults.set(forKey: .DEVICECONFIG_DIAS_ID, value: ddcciText)
                            UserDefaults.setInt(forKey: .DEVICECONFIG_SRC, value: intSource)
                            UserDefaults.setInt(forKey: .DEVICECONFIG_PORT, value: intPort)
                            P2PManager.shared.setupDeviceConfig(ddcciText, intSource, intPort)
                            self.configContenView.refreshUI()
                        }
                    }
        
                    self.view.setNeedsLayout()
                    self.view.layoutIfNeeded()
                }
            case .file:
                Logger.info("switch file")
                self.userManuallySwitchedFromFile = false
                DispatchQueue.main.async { [weak self] in
                    guard let self = self else { return  }
                    self.deviceView.removeFromSuperview()
                    self.configContenView.removeFromSuperview()
                    self.view.addSubview(self.fileContennerView)
                    
                    self.fileContennerView.snp.makeConstraints { make in
                        make.left.right.equalToSuperview()
                        make.top.equalTo(self.videoContainerView.snp.bottom)
                        make.bottom.equalToSuperview()
                    }
                    self.view.setNeedsLayout()
                    self.view.layoutIfNeeded()
                }
            }
        }
    }
    
    override func viewDidLoad() {
        super.viewDidLoad()
        Logger.info("before PIP init：\(UIApplication.shared.windows)")
        if AVPictureInPictureController.isPictureInPictureSupported() {
            do {
                try AVAudioSession.sharedInstance().setCategory(.playback, options: .mixWithOthers)
            } catch {
                Logger.info(error.localizedDescription)
            }
            setupUI()
            _pipContent.start()
            DispatchQueue.main.async {
                self.start()
            }
            NotificationCenter.default.addObserver(self, selector: #selector(handleEnterForeground), name: UIApplication.willEnterForegroundNotification, object: nil)
            NotificationCenter.default.addObserver(self, selector: #selector(handleEnterBackground), name: UIApplication.didEnterBackgroundNotification, object: nil)
            NotificationCenter.default.addObserver(self, selector: #selector(receriveNewFiles(_:)), name: ReceiveFuleSuccessNotification, object: nil)
            NotificationCenter.default.addObserver(self, selector: #selector(updateClientList), name: UpdateClientListSuccessNotification, object: nil)
//            NotificationCenter.default.addObserver(self, selector: #selector(receivedFile), name: ReceivedFilesSuccessNotification, object: nil)
        } else {
            Logger.info("not support PIP")
        }
        
        UIApplication.shared.beginBackgroundTask {
            UIApplication.shared.endBackgroundTask(UIBackgroundTaskIdentifier.invalid)
        }
    }

    deinit {
        NotificationCenter.default.removeObserver(self)
    }

    override func touchesBegan(_ touches: Set<UITouch>, with event: UIEvent?) {
        super.touchesBegan(touches, with: event)
        
        self.view.endEditing(true)
    }
    
    
    // 配置UI
    private func setupUI() {
        self.title = "Cross Share"
        self.navigationItem.leftBarButtonItem = UIBarButtonItem(customView: self.aheadButton)
        self.navigationItem.rightBarButtonItem = UIBarButtonItem(customView: self.clientsLable)

        let session = AVAudioSession.sharedInstance()
        try! session.setCategory(.playback, mode: .moviePlayback)
        try! session.setActive(true)

        self.aheadButton.isSelected = true
        self.viewType = .file
        self.view.addSubview(videoContainerView)
        self.view.backgroundColor = UIColor.init(hex: 0xF6F6F6)

        self.videoContainerView.snp.makeConstraints { make in
            make.left.right.equalToSuperview()
            make.top.equalTo(self.view.safeAreaLayoutGuide.snp.top).offset(10)
            make.height.equalTo(50)
        }

        self.videoView.snp.makeConstraints { make in
            make.center.equalToSuperview()
            make.size.equalTo(CGSize(width: 200, height: 30))
        }

        self.videoView.setNeedsLayout()
        self.videoView.layoutIfNeeded()

        let bufferDisplayLayer = _pipContent.bufferDisplayLayer
        bufferDisplayLayer.frame = self.videoView.bounds
        bufferDisplayLayer.videoGravity = .resizeAspect
        self.videoView.layer.addSublayer(bufferDisplayLayer)
    }

    lazy var videoView: UIView = {
        let view = UIView(frame: .zero)
        view.backgroundColor = UIColor.clear
        view.isUserInteractionEnabled = true
        videoContainerView.addSubview(view)
        return view
    }()
    
    lazy var clientsLable: UILabel = {
        let text = UILabel(frame: .zero)
        text.textColor = UIColor.white
        text.font = UIFont.systemFont(ofSize: 14)
        text.text = "Online: 0"
        return text
    }()
    
    lazy var aheadButton: UIButton = {
        let button = UIButton(type: .custom)
        button.frame = CGRect(x: 0, y: 0, width: 44, height: 44)
        button.setImage(nil, for: .normal)
        button.setImage(UIImage(named: "arrow_left"), for: .selected)
        button.contentHorizontalAlignment = .left
        button.addTarget(self, action: #selector(changeView(_:)), for: .touchUpInside)
        return button
    }()
    
    lazy var startButton: UIButton = {
        let button = UIButton(type: .custom)
        button.setTitle("Start PiP", for: .normal)
        button.setTitleColor(UIColor.systemBlue, for: .normal)
        button.addTarget(self, action: #selector(toggle(_:)), for: .touchUpInside)
        videoContainerView.addSubview(button)
        return button
    }()
    
    //main
    lazy var deviceView: DeviceMainView = {
        let view = DeviceMainView(frame: .zero)
        view.backgroundColor = UIColor.clear
        view.isUserInteractionEnabled = true
        return view
    }()
    
    //Config
    lazy var configContenView: DeviceConfigView = {
        let view = DeviceConfigView(frame: .zero)
        view.backgroundColor = UIColor.clear
        view.isUserInteractionEnabled = true
        return view
    }()
    
    //pip
    lazy var videoContainerView: UIView = {
        let view = UIView(frame: .zero)
        view.backgroundColor = UIColor.clear
        view.isUserInteractionEnabled = true
        return view
    }()
    
    //file
    lazy var fileContennerView: DeviceTransportView = {
        let view = DeviceTransportView(frame: .zero)
        view.backgroundColor = UIColor.clear
        view.isUserInteractionEnabled = true
        return view
    }()
    
    func start() {
        if AVPictureInPictureController.isPictureInPictureSupported() {
            do {
                try AVAudioSession.sharedInstance().setCategory(.playback, options: .mixWithOthers)
            } catch {
                Logger.info(error.localizedDescription)
            }
            _pipController = AVPictureInPictureController(
                contentSource: .init(
                    sampleBufferDisplayLayer:
                        _pipContent.bufferDisplayLayer,
                    playbackDelegate: self))
            _pipController?.delegate = self
            _pipController?.setValue(2, forKey: "controlsStyle")
            if #available(iOS 14.2, *) {
                _pipController?.canStartPictureInPictureAutomaticallyFromInline = true
            } else {
                // Fallback on earlier versions
            }
            _pipPossibleObservation = _pipController?.observe(
                \AVPictureInPictureController.isPictureInPicturePossible,
                 options: [.initial, .new]) { [weak self] _, change in
                     guard let self = self else { return }
                     if (change.newValue ?? false) {
                         self.startButton.isEnabled = (change.newValue ?? false)
                     }
                 }
        }
    }
    
    @objc func toggle(_ sender:UIButton) {
        guard let _pipController = _pipController else { return }
        if !_pipController.isPictureInPictureActive {
            sender.setTitle("Stop PiP", for: .normal)
            _pipController.startPictureInPicture()
        } else {
            sender.setTitle("Start PiP", for: .normal)
            _pipController.stopPictureInPicture()
        }
    }
    
    @objc func changeView(_ sender:UIButton) {
        guard self.viewType != .main else {
            return
        }
        sender.isSelected = !sender.isSelected
        if !sender.isSelected {
            self.userManuallySwitchedFromFile = true
            self.viewType = .main
        }
    }

    // MARK: - 进入前后台
    @objc private func handleEnterForeground() {
        Logger.info("handleEnterForeground");
    }

    @objc private func handleEnterBackground() {
        Logger.info("handleEnterBackground");
    }

    @objc private func receriveNewFiles(_ ntf: Notification) {
        DispatchQueue.main.async { [weak self] in
            guard let self = self else { return }
            if let userInfo = ntf.userInfo as? [String: Any],
               let newItem = userInfo["download"] as? DownloadItem {
                var isNewFile = false
                if let index = self.dataArray.firstIndex(where: { $0.uuid == newItem.uuid }) {
                    self.dataArray[index].receiveSize = newItem.receiveSize
                    self.dataArray[index].totalSize = newItem.totalSize
                    self.dataArray[index].finishTime = newItem.finishTime
                    self.dataArray[index].timestamp = newItem.timestamp
                    if let recvFileCnt = newItem.recvFileCnt, let currentfileName = newItem.currentfileName {
                        self.dataArray[index].recvFileCnt = recvFileCnt
                        self.dataArray[index].currentfileName = currentfileName
                    }
                } else {
                    self.dataArray.append(newItem)
                    isNewFile = true
                }
                self.dataArray.sort { ($0.timestamp ?? 0) > ($1.timestamp ?? 0) }
                self.fileContennerView.dataArray = self.dataArray
                if isNewFile {
                    self.userManuallySwitchedFromFile = false

                    if let index = self.dataArray.firstIndex(where: { $0.uuid == newItem.uuid }) {
                        let fileCnt = (self.dataArray[index].totalFileCnt ?? 0)
                        let fileMsg = fileCnt > 1 ? "\(fileCnt) files" : "\(fileCnt) file"
                        let deviceName = self.dataArray[index].deviceName ?? ""
                        let params: [String] = [fileMsg, deviceName]
                        PushNotiManager.shared.sendLocalNotification(code: .receiveStart, with: params)
                    }
                }
                if !self.userManuallySwitchedFromFile {
                    self.viewType = .file
                    self.aheadButton.isSelected = true
                }
           }
        }
    }

    @objc private func updateClientList() {
        DispatchQueue.main.async { [weak self] in
            guard let self = self else {
                return
            }
            self.deviceView.refreshUI()
            self.clientsLable.text = "Online: \(P2PManager.shared.clientList.count)"
        }
    }

    @objc private func pickDocument() {
        let supportedTypes: [UTType] = [.data]
        let documentPicker = UIDocumentPickerViewController(forOpeningContentTypes: supportedTypes, asCopy: false)
        documentPicker.delegate = self
        documentPicker.allowsMultipleSelection = false
        present(documentPicker, animated: true)
    }

    private func dismissDevicePopView() {
        guard let popView = self.devicePopView else { return }
        UIView.animate(withDuration: 0.3, animations: {
            popView.alpha = 0
        }) { _ in
            popView.removeFromSuperview()
            self.devicePopView = nil
        }
    }

    func getFileSize(atPath path: String) -> Int64? {
        let fileManager = FileManager.default
        do {
            let attributes = try fileManager.attributesOfItem(atPath: path)
            if let fileSize = attributes[FileAttributeKey.size] as? NSNumber {
                return fileSize.int64Value
            }
        } catch {
            Logger.info("Error fetching file size: \(error)")
        }
        return nil
    }
    
    private func configClients(with urls: [URL]) {
        guard let url = urls.first else { return }
        if url.startAccessingSecurityScopedResource() {
            defer { url.stopAccessingSecurityScopedResource() }
            let fileName = (url.path as NSString).lastPathComponent
            let clients = P2PManager.shared.clientList
            guard clients.isEmpty == false else {
                Logger.info("not found clients")
                MBProgressHUD.showTips(.error,"Please make sure your device is online", toView: self.view)
                return
            }
            Logger.info("Found \(clients.count) clients:")
            for client in clients {
                Logger.info("Device: \(client.name), IP:Port: \(client.ip), ID: \(client.id)")
            }
            if !clients.isEmpty {
                let popView = DeviceSelectPopView(fileNames: [fileName], clients: clients)
                popView.frame = self.view.bounds
                popView.alpha = 0
                popView.onSelect = { [weak self] client in
                    guard let self = self else { return  }
                    Logger.info("select device：\(client.name)")
                    self.selectedClient = client
                    MBProgressHUD.showTips(.error,"Choose device：\(client.name)", toView: self.view)
                }
                popView.onCancel = { [weak self] in
                    self?.dismissDevicePopView()
                }
                popView.onSure = { [weak self] in
                    guard let self = self else { return }
                    guard let selectedClient = self.selectedClient else {
                        MBProgressHUD.showTips(.error,"Please select a device", toView: self.view)
                        return
                    }
                    self.dismissDevicePopView()
                    
                    // TODO: Create the unique folder means the each file transfer (now use timestamp)
                    // TODO: It should be deleted after the trnsfer finished
                    let timestamp = String(Int(Date().timeIntervalSince1970))
                    guard let finalFilePath = copyDocumentToTemp(url, timestamp) else {
                        return
                    }
                    guard let fileSize = self.getFileSize(atPath: finalFilePath ) else {
                        MBProgressHUD.showTips(.error,"Please select a vaild file", toView: self.view)
                        return
                    }
                    let fileMsg = "1 file"
                    let params: [String] = [fileMsg, selectedClient.name]
                    PushNotiManager.shared.sendLocalNotification(code: .sendStart, with: params)
                    P2PManager.shared.setFileDropRequest(filePath: finalFilePath, id: selectedClient.id, fileSize: fileSize)
                    self.selectedClient = nil
                }
                self.view.addSubview(popView)
                self.devicePopView = popView
                UIView.animate(withDuration: 0.3) {
                    popView.alpha = 1
                    if let contentView = popView.subviews.first {
                        contentView.transform = .identity
                    }
                }
            }
        } else {
            Logger.info("Unable to obtain security permissions")
        }
    }
}

extension ViewController: AVPictureInPictureControllerDelegate {

    func pictureInPictureController(
        _ pictureInPictureController: AVPictureInPictureController,
        failedToStartPictureInPictureWithError error: Error
    ) {
        Logger.info("\(#function)")
        Logger.info("pip error: \(error)")
    }

    func pictureInPictureControllerWillStartPictureInPicture(
        _ pictureInPictureController: AVPictureInPictureController
    ) {
        Logger.info("\(#function)")
    }

    func pictureInPictureControllerWillStopPictureInPicture(
        _ pictureInPictureController: AVPictureInPictureController
    ) {
        Logger.info("\(#function)")
        Logger.info("\(#function)")
        if !pictureInPictureController.isPictureInPictureActive {
            startButton.setTitle("Stop PiP", for: .normal)
        } else {
            startButton.setTitle("Start PiP", for: .normal)
        }
    }
}

extension ViewController: AVPictureInPictureSampleBufferPlaybackDelegate {
    func pictureInPictureController(
        _ pictureInPictureController: AVPictureInPictureController,
        setPlaying playing: Bool
    ) {
        Logger.info("\(#function)")
    }

    func pictureInPictureControllerTimeRangeForPlayback(
        _ pictureInPictureController: AVPictureInPictureController
    ) -> CMTimeRange {
        Logger.info("\(#function)")
        return CMTimeRange(start: .negativeInfinity, duration: .positiveInfinity)
    }

    func pictureInPictureControllerIsPlaybackPaused(
        _ pictureInPictureController: AVPictureInPictureController
    ) -> Bool {
        Logger.info("\(#function)")
        return false
    }

    func pictureInPictureController(
        _ pictureInPictureController: AVPictureInPictureController,
        didTransitionToRenderSize newRenderSize: CMVideoDimensions
    ) {
        Logger.info("\(#function)")
        Logger.info("New render size: \(newRenderSize.width)x\(newRenderSize.height)")
    }

    func pictureInPictureController(
        _ pictureInPictureController: AVPictureInPictureController,
        skipByInterval skipInterval: CMTime,
        completion completionHandler: @escaping () -> Void
    ) {
        Logger.info("\(#function)")
        completionHandler()
    }
}

extension ViewController: UIDocumentPickerDelegate {
    public func documentPicker(_ controller: UIDocumentPickerViewController, didPickDocumentsAt urls: [URL]) {
        self.configClients(with: urls)
    }

    public func documentPickerWasCancelled(_ controller: UIDocumentPickerViewController) {
        Logger.info("[DocPicker] User cancel selection")
    }

    private func copyDocumentToTemp(_ url: URL,_ timestamp: String) -> String? {
        guard url.startAccessingSecurityScopedResource() else {
            Logger.info("[DocPicker][Err] Copy failed: Unauthorized")
            return nil
        }
        defer { url.stopAccessingSecurityScopedResource() }

        let pathComponents = url.pathComponents
        guard let uuidIndex = pathComponents.firstIndex(where: { $0.range(of: #"^[A-F0-9\-]{36}$"#, options: .regularExpression) != nil }) else {
            Logger.info("[DocPicker][Err] Invalid UUID rules")
            return nil
        }

        let uuid = pathComponents[uuidIndex]
        var subPath = pathComponents[(uuidIndex + 1)...].joined(separator: "/")
        if subPath.hasPrefix("/") {
            subPath.removeFirst()
        }

        let destFolder = FileManager.default.temporaryDirectory
            .appendingPathComponent(timestamp)
            .appendingPathComponent(uuid)

        let destFullPath = destFolder.appendingPathComponent(subPath)

        do {
            try FileManager.default.createDirectory(at: destFullPath.deletingLastPathComponent(), withIntermediateDirectories: true)
            try FileManager.default.copyItem(at: url, to: destFullPath)
            Logger.info("[DocPicker] Copy to : \(destFullPath.path)")
        } catch {
            Logger.info("[DocPicker][Err] Copy filed: \(error)")
        }

        return destFullPath.path
    }

    private func removeTempFolderByTimestamp(_ timestamp: String) {
        let root = FileManager.default.temporaryDirectory.appendingPathComponent(timestamp)
        if FileManager.default.fileExists(atPath: root.path) {
            do {
                try FileManager.default.removeItem(at: root)
                Logger.info("[DocPicker] Clean files successfully: \(root.path)")
            } catch {
                Logger.info("[DocPicker][Err] Clean files failed: \(error)")
            }
        } else {
            Logger.info("[DocPicker][Err] Clean files failed: Path not exited: \(root.path)")
        }
    }
}
