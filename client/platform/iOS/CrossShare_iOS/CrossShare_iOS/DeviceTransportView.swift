//
//  DeviceTransportView.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/5/15.
//

import UIKit
import MBProgressHUD

enum FileTransferState: Int, CaseIterable {
    case inProgress = 0 // 正在进行中
    case recorded = 1   // 已完成记录
    case failed = 2     // 传输失败
    
    var title: String {
        switch self {
        case .inProgress: return "In Progress"
        case .recorded: return "Received"
        case .failed: return "Fail to send"
        }
    }
    
    var emptyMessage: String {
        switch self {
        case .inProgress: return "No transfers in progress"
        case .recorded: return "No files transferred yet"
        case .failed: return "No failed transfers"
        }
    }
    
    var color: UIColor {
        switch self {
        case .inProgress, .recorded: return UIColor(hex: 0x007AFF)
        case .failed: return UIColor.systemRed
        }
    }
}

enum TransportViewState {
    case noFiles
    case downloading
    case fileList
}

class DeviceTransportView: UIView {
    
    var mFileOpener: FileOpener? = nil
    public var dataArray:[DownloadItem] = [] {
        didSet {
            detectCompletedFiles(oldData: oldValue, newData: dataArray)
            updateViewState()
            reloadAllTableViews()
        }
    }
    
    private var currentState: TransportViewState = .noFiles
    private var currentFilterState: FileTransferState = .inProgress
    
    private var fileCancelView: FileCancelPopView?
    
    // Track completed files with their completion time for delayed removal
    private var delayedCompletedFiles: [String: Date] = [:]
    private var delayRemovalTimer: Timer?
    
    private var inProgressData: [DownloadItem] {
        let now = Date()
        return dataArray.filter { item in
            // Files currently being transferred
            let isInProgress = (item.receiveSize ?? 0) < (item.totalSize ?? 0) && item.error == nil
            
            // Completed files within 2-second delay period
            if let completedTime = delayedCompletedFiles[item.uuid] {
                let elapsed = now.timeIntervalSince(completedTime)
                return elapsed < 2.0 && item.error == nil
            }
            
            return isInProgress
        }
    }
    
    private var recordedData: [DownloadItem] {
        return dataArray.filter { item in
            let isCompleted = (item.receiveSize ?? 0) >= (item.totalSize ?? 0) && item.error == nil
            // Exclude files in delayed display period
            let isDelayed = delayedCompletedFiles[item.uuid] != nil
            return isCompleted && !isDelayed
        }
    }
    
    private var failedData: [DownloadItem] {
        return dataArray.filter { $0.error != nil }
    }
    
    private func dataSource(for state: FileTransferState) -> [DownloadItem] {
        switch state {
        case .inProgress: return inProgressData
        case .recorded: return recordedData
        case .failed: return failedData
        }
    }
    
    // Detect newly completed files
    private func detectCompletedFiles(oldData: [DownloadItem], newData: [DownloadItem]) {
        for newItem in newData {
            // Check if transfer just completed
            let isNowCompleted = (newItem.receiveSize ?? 0) >= (newItem.totalSize ?? 0) && newItem.error == nil
            
            if isNowCompleted {
                // Find corresponding item in old data
                if let oldItem = oldData.first(where: { $0.uuid == newItem.uuid }) {
                    let wasInProgress = (oldItem.receiveSize ?? 0) < (oldItem.totalSize ?? 0)
                    
                    // If was in progress and now completed, record completion time
                    if wasInProgress && delayedCompletedFiles[newItem.uuid] == nil {
                        delayedCompletedFiles[newItem.uuid] = Date()
                        Logger.info("File transfer completed, will be removed after 2s delay: \(newItem.currentfileName ?? newItem.uuid)")
                        scheduleDelayedRemoval()
                    }
                } else {
                    // New file that's already completed (fast transfer scenario)
                    if delayedCompletedFiles[newItem.uuid] == nil {
                        delayedCompletedFiles[newItem.uuid] = Date()
                        Logger.info("Fast transfer completed, will be removed after 2s delay: \(newItem.currentfileName ?? newItem.uuid)")
                        scheduleDelayedRemoval()
                    }
                }
            }
        }
        
        // Clean up records for files that no longer exist
        let currentUUIDs = Set(newData.map { $0.uuid })
        delayedCompletedFiles = delayedCompletedFiles.filter { currentUUIDs.contains($0.key) }
    }
    
    // Schedule delayed removal timer
    private func scheduleDelayedRemoval() {
        // Don't create duplicate timer if already running
        if delayRemovalTimer != nil {
            return
        }
        
        // Create timer that checks every 0.5 seconds
        delayRemovalTimer = Timer.scheduledTimer(withTimeInterval: 0.5, repeats: true) { [weak self] _ in
            self?.checkAndRemoveExpiredFiles()
        }
    }
    
    // Check and remove expired delayed files
    private func checkAndRemoveExpiredFiles() {
        let now = Date()
        var hasExpired = false
        
        // Find all files that have exceeded 2 seconds
        let expiredUUIDs = delayedCompletedFiles.filter { _, completedTime in
            now.timeIntervalSince(completedTime) >= 2.0
        }.map { $0.key }
        
        if !expiredUUIDs.isEmpty {
            hasExpired = true
            Logger.info("Removing delayed files count: \(expiredUUIDs.count)")
            
            // Remove from delayed list
            expiredUUIDs.forEach { delayedCompletedFiles.removeValue(forKey: $0) }
        }
        
        // Stop timer if no more files to process
        if delayedCompletedFiles.isEmpty {
            delayRemovalTimer?.invalidate()
            delayRemovalTimer = nil
        }
        
        // Refresh view if any files expired
        if hasExpired {
            DispatchQueue.main.async { [weak self] in
                self?.updateViewState()
                self?.reloadAllTableViews()
            }
        }
    }
    
    private lazy var segmentButtons: [UIButton] = {
        return FileTransferState.allCases.map { state in
            createSegmentButton(for: state)
        }
    }()
    
    private lazy var containerViews: [UIView] = {
        return FileTransferState.allCases.map { _ in
            createContainerView()
        }
    }()
    
    private lazy var tableViews: [UITableView] = {
        return FileTransferState.allCases.map { state in
            createTableView(for: state)
        }
    }()
    
    private lazy var emptyViews: [UIView] = {
        return FileTransferState.allCases.map { state in
            createEmptyView(for: state)
        }
    }()
    
    func switchToInProgress() {
        let inProgressIndex = FileTransferState.inProgress.rawValue
        if currentFilterState != .inProgress {
            scrollToPage(inProgressIndex)
            updateSegmentSelection(inProgressIndex)
            currentFilterState = .inProgress
            updateViewForCurrentState()
        }
    }
    
    func isCurrentlyOnInProgress() -> Bool {
        return currentFilterState == .inProgress
    }
    
    lazy var customSegmentView: UIView = {
        let view = UIView()
        view.backgroundColor = UIColor.systemBackground
        
        let stackView = UIStackView(arrangedSubviews: segmentButtons)
        stackView.axis = .horizontal
        stackView.distribution = .fillEqually
        stackView.spacing = 0
        
        view.addSubview(stackView)
        stackView.snp.makeConstraints { make in
            make.edges.equalToSuperview()
        }
        
        return view
    }()
    
    lazy var scrollView: UIScrollView = {
        let scrollView = UIScrollView()
        scrollView.isPagingEnabled = true
        scrollView.showsHorizontalScrollIndicator = false
        scrollView.delegate = self
        return scrollView
    }()

    override init(frame: CGRect) {
        super.init(frame: frame)
        setupUI()
    }
    
    required init?(coder: NSCoder) {
        super.init(coder: coder)
        setupUI()
    }
    
    deinit {
        delayRemovalTimer?.invalidate()
        delayRemovalTimer = nil
    }
    
    func setupUI() {
        addSubview(customSegmentView)
        addSubview(scrollView)
        
        containerViews.forEach { scrollView.addSubview($0) }
        
        setupContainerViews()
        setupConstraints()
        
        updateSegmentAppearance()
        updateViewState()
    }
    
    private func setupContainerViews() {
        for (index, containerView) in containerViews.enumerated() {
            containerView.addSubview(tableViews[index])
            containerView.addSubview(emptyViews[index])
        }
        setupTableViewConstraints()
    }
    
    private func setupConstraints() {
        customSegmentView.snp.makeConstraints { make in
            make.top.equalToSuperview().offset(10)
            make.left.right.equalToSuperview()
            make.height.equalTo(50)
        }
        
        scrollView.snp.makeConstraints { make in
            make.top.equalTo(customSegmentView.snp.bottom)
            make.left.right.bottom.equalToSuperview()
        }
        
        setupContainerConstraints()
    }
    
    private func setupContainerConstraints() {
        for (index, containerView) in containerViews.enumerated() {
            containerView.snp.makeConstraints { make in
                make.top.bottom.equalToSuperview()
                make.width.equalTo(self)
                make.height.equalTo(scrollView)
                
                if index == 0 {
                    make.left.equalToSuperview()
                } else {
                    make.left.equalTo(containerViews[index - 1].snp.right)
                }
                
                if index == containerViews.count - 1 {
                    make.right.equalToSuperview()
                }
            }
        }
    }
    
    private func setupTableViewConstraints() {
        for (index, tableView) in tableViews.enumerated() {
            tableView.snp.makeConstraints { make in
                make.edges.equalToSuperview().inset(UIEdgeInsets(top: 0, left: 0, bottom: 0, right: 0))
            }
            tableView.contentInset = UIEdgeInsets(top: 4, left: 0, bottom: 4, right: 0)
            
            emptyViews[index].snp.makeConstraints { make in
                make.center.equalToSuperview()
                make.width.equalTo(250)
                make.height.equalTo(200)
            }
        }
        
        setupEmptyViewsConstraints()
    }
    
    private func setupEmptyViewsConstraints() {
        for emptyView in emptyViews {
            guard let imageView = emptyView.subviews.first as? UIImageView,
                  let label = emptyView.subviews.last as? UILabel else { continue }
            
            imageView.snp.makeConstraints { make in
                make.centerX.equalToSuperview()
                make.centerY.equalToSuperview().offset(-30)
                make.width.height.equalTo(80)
            }
            
            label.snp.makeConstraints { make in
                make.centerX.equalToSuperview()
                make.top.equalTo(imageView.snp.bottom).offset(20)
                make.left.right.equalToSuperview().inset(20)
            }
        }
    }
    
    private func createSegmentButton(for state: FileTransferState) -> UIButton {
        let button = UIButton()
        button.setTitle(state.title, for: .normal)
        button.setTitleColor(state.color, for: .normal)
        button.titleLabel?.font = UIFont.systemFont(ofSize: 16, weight: .medium)
        button.tag = state.rawValue
        button.addTarget(self, action: #selector(segmentButtonTapped(_:)), for: .touchUpInside)
        
        let indicatorView = UIView()
        indicatorView.backgroundColor = state.color
        indicatorView.isHidden = state != .inProgress // 默认选中 recorded
        button.addSubview(indicatorView)
        
        indicatorView.snp.makeConstraints { make in
            make.bottom.equalToSuperview()
            make.left.right.equalToSuperview().inset(20)
            make.height.equalTo(3)
        }
        
        return button
    }
    
    private func createContainerView() -> UIView {
        let view = UIView()
        view.backgroundColor = .clear
        return view
    }
    
    private func createTableView(for state: FileTransferState) -> UITableView {
        let tableView = UITableView()
        tableView.backgroundColor = .init(hex: 0xF0F0F0)
        tableView.delegate = self
        tableView.dataSource = self
        tableView.showsVerticalScrollIndicator = false
        tableView.rowHeight = 130
        tableView.separatorStyle = .none
        tableView.tag = state.rawValue
        tableView.register(DownloadViewCell.self, forCellReuseIdentifier: NSStringFromClass(DownloadViewCell.self))
        return tableView
    }
    
    private func createEmptyView(for state: FileTransferState) -> UIView {
        let containerView = UIView()
        containerView.backgroundColor = .clear
        
        let imageView = UIImageView(image: UIImage(named: "empty_files"))
        imageView.contentMode = .scaleAspectFit
        
        let label = UILabel()
        label.text = state.emptyMessage
        label.textColor = UIColor.gray
        label.textAlignment = .center
        label.font = UIFont.systemFont(ofSize: 16)
        label.numberOfLines = 0
        
        containerView.addSubview(imageView)
        containerView.addSubview(label)
        
        return containerView
    }
    
    @objc private func segmentButtonTapped(_ sender: UIButton) {
        guard let state = FileTransferState(rawValue: sender.tag) else { return }
        scrollToPage(state.rawValue)
        updateSegmentSelection(state.rawValue)
    }
    
    private func scrollToPage(_ index: Int) {
        let pageWidth = scrollView.frame.width
        let offsetX = CGFloat(index) * pageWidth
        scrollView.setContentOffset(CGPoint(x: offsetX, y: 0), animated: true)
        
        if let state = FileTransferState(rawValue: index) {
            currentFilterState = state
        }
        updateViewForCurrentState()
    }
    
    private func updateSegmentSelection(_ index: Int) {
        for (buttonIndex, button) in segmentButtons.enumerated() {
            let isSelected = buttonIndex == index
            let state = FileTransferState.allCases[buttonIndex]
            
            button.subviews.last?.isHidden = !isSelected
            button.setTitleColor(isSelected ? state.color : UIColor.gray, for: .normal)
        }
    }
    
    private func updateSegmentAppearance() {
        updateSegmentSelection(currentFilterState.rawValue)
    }
    
    private func updateViewForCurrentState() {
        updateViewState()
    }
    
    func updateViewState() {
        if dataArray.isEmpty {
            showEmptyState()
            return
        }
        
        for state in FileTransferState.allCases {
            updateView(for: state)
        }
    }
    
    private func updateView(for state: FileTransferState) {
        let index = state.rawValue
        let data = dataSource(for: state)
        
        if data.isEmpty {
            tableViews[index].isHidden = true
            emptyViews[index].isHidden = false
        } else {
            emptyViews[index].isHidden = true
            tableViews[index].isHidden = false
        }
    }
    
    private func showEmptyState() {
        tableViews.forEach { $0.isHidden = true }
        emptyViews.forEach { $0.isHidden = false }
    }
    
    private func reloadAllTableViews() {
        DispatchQueue.main.async { [weak self] in
            guard let self = self else { return }
            self.tableViews.forEach { tableView in
                if tableView.window != nil {
                    tableView.reloadData()
                }
            }
        }
    }
    
    func switchToFailed() {
        let failedIndex = FileTransferState.failed.rawValue
        if currentFilterState != .failed {
            scrollToPage(failedIndex)
            updateSegmentSelection(failedIndex)
            currentFilterState = .failed
            updateViewForCurrentState()
            // Ensure table view data is refreshed when switching to failed page
            reloadAllTableViews()
        }
    }
}

// MARK: - UIScrollViewDelegate
extension DeviceTransportView: UIScrollViewDelegate {
    func scrollViewDidEndDecelerating(_ scrollView: UIScrollView) {
        guard scrollView == self.scrollView else { return }
        let pageIndex = Int(scrollView.contentOffset.x / scrollView.frame.width)
        updateSegmentSelection(pageIndex)
        
        if let state = FileTransferState(rawValue: pageIndex) {
            currentFilterState = state
        }
        updateViewForCurrentState()
    }
    
    func scrollViewDidEndDragging(_ scrollView: UIScrollView, willDecelerate decelerate: Bool) {
        guard scrollView == self.scrollView else { return }
        if !decelerate {
            let pageIndex = Int(scrollView.contentOffset.x / scrollView.frame.width)
            updateSegmentSelection(pageIndex)
            if let state = FileTransferState(rawValue: pageIndex) {
                currentFilterState = state
            }
            updateViewForCurrentState()
        }
    }
}

// MARK: - File Operations
extension DeviceTransportView {
    @objc private func clearAction() {
        DispatchQueue.main.async { [weak self] in
            guard let self = self else { return }
            switch self.currentFilterState {
            case .inProgress:
                self.dataArray.removeAll { $0.receiveSize ?? 0 < $0.totalSize ?? 0 }
            case .recorded:
                self.dataArray.removeAll { $0.receiveSize ?? 0 >= $0.totalSize ?? 0 && $0.error == nil }
            case .failed:
                self.dataArray.removeAll { $0.error != nil }
            }
            self.updateViewState()
        }
    }
    
    @objc private func deleteFile(at index: Int, from tableView: UITableView) {
        DispatchQueue.main.async { [weak self] in
            guard let self = self,
                  let state = FileTransferState(rawValue: tableView.tag) else { return }
            
            let filteredData = self.dataSource(for: state)
            
            if index >= 0 && index < filteredData.count {
                let itemToDelete = filteredData[index]
                if let originalIndex = self.dataArray.firstIndex(where: { $0.uuid == itemToDelete.uuid }) {
                    self.dataArray.remove(at: originalIndex)
                    self.updateViewState()
                }
            }
        }
    }

    @objc private func openFile(at index: Int, from tableView: UITableView) {
        DispatchQueue.main.async { [weak self] in
            guard let self = self,
                  let state = FileTransferState(rawValue: tableView.tag) else { return }
            
            let filteredData = self.dataSource(for: state)
            
            if index >= 0 && index < filteredData.count {
                let filesItem = filteredData[index]
                
                guard let fileCnt = filesItem.totalFileCnt else { return }

                if fileCnt > 1 {
                    MBProgressHUD.showTips(.error,"Only file can be opened", toView: self)
                    return
                }

                guard let filePath = filesItem.currentfileName else { return }

                var fileName = ""
                if filesItem.isMutip {
                    if let nameArray = filesItem.currentfileName?.components(separatedBy: "/") as? [String], nameArray.count > 1 {
                        fileName = nameArray.last ?? ""
                    }
                } else {
                    fileName = filePath
                }

                guard let viewController = self.parentViewController else { return }

                self.mFileOpener = FileOpener(presenter: viewController)
                if (self.mFileOpener?.openFile(fileName: fileName) == false) {
                    MBProgressHUD.showTips(.error,"Only file can be opened", toView: self)
                }
            }
        }
    }
    
    @objc private func cancelFile(at index: Int, from tableView: UITableView) {
        guard let currentWindow = UtilsHelper.shared.getTopWindow() else {
            Logger.error("No active window found")
            return
        }
        let popView = FileCancelPopView()
        popView.frame = currentWindow.bounds
        popView.alpha = 0
        self.fileCancelView = popView
        popView.onCancel = { [weak self] in
            guard let self = self else { return }
            self.dismissDevicePopView()
        }
        popView.onSure = { [weak self] in
            DispatchQueue.main.async { [weak self] in
                guard let self = self,
                      let state = FileTransferState(rawValue: tableView.tag) else { return }
                let filteredData = self.dataSource(for: state)
                if index >= 0 && index < filteredData.count {
                    let itemToCancel = filteredData[index]
                    if let originalIndex = self.dataArray.firstIndex(where: { $0.uuid == itemToCancel.uuid }) {
                        if state == .inProgress {
                            self.dataArray[originalIndex].error = "Transfer cancelled by user"
                            self.dataArray[originalIndex].finishTime = Date().timeIntervalSince1970
                            Logger.info("file has been cancel: \(itemToCancel.currentfileName ?? "Unknown")")
                            
                            // Manually trigger view update since modifying array element properties doesn't trigger didSet
                            self.updateViewState()
                            self.reloadAllTableViews()
                            
                            P2PManager.shared.setCancelFileTransfer(ipPort: itemToCancel.ip, clientID: itemToCancel.fileId, timeStamp: UInt64(itemToCancel.timestamp ?? Date().timeIntervalSince1970))
                            
                            // Ensure table view is refreshed before switching to failed page
                            DispatchQueue.main.asyncAfter(deadline: .now() + 0.1) {
                                self.switchToFailed()
                                // Refresh again to ensure latest data is displayed
                                self.reloadAllTableViews()
                            }
                        } else {
                            self.dataArray.remove(at: originalIndex)
                        }
                    }
                }
                self.dismissDevicePopView()
            }
        }
        currentWindow.addSubview(popView)
        UIView.animate(withDuration: 0.3) {
            popView.alpha = 1
            if let contentView = popView.subviews.first {
                contentView.transform = .identity
            }
        }
    }
    
    private func dismissDevicePopView() {
        guard let popView = self.fileCancelView else { return }
        UIView.animate(withDuration: 0.3, animations: {
            popView.alpha = 0
            if let contentView = popView.subviews.first {
                contentView.transform = CGAffineTransform(scaleX: 0.8, y: 0.8)
            }
        }) { _ in
            popView.removeFromSuperview()
            self.fileCancelView = nil
        }
    }
}

// MARK: - UITableViewDelegate, UITableViewDataSource
extension DeviceTransportView: UITableViewDelegate, UITableViewDataSource {
    func tableView(_ tableView: UITableView, numberOfRowsInSection section: Int) -> Int {
        guard let state = FileTransferState(rawValue: tableView.tag) else { return 0 }
        return dataSource(for: state).count
    }
    
    func tableView(_ tableView: UITableView, cellForRowAt indexPath: IndexPath) -> UITableViewCell {
        guard let cell = tableView.dequeueReusableCell(withIdentifier: NSStringFromClass(DownloadViewCell.self)) as? DownloadViewCell,
              let state = FileTransferState(rawValue: tableView.tag) else {
            return UITableViewCell()
        }
        
        let data = dataSource(for: state)
        guard indexPath.row < data.count else { return cell }
        
        let model = data[indexPath.row]
        
        cell.refreshUI(with: model)
        cell.deleteBlock = { [weak self] in
            guard let self = self else { return }
            self.deleteFile(at: indexPath.row, from: tableView)
        }
        cell.openBlock = { [weak self] in
            guard let self = self else { return }
            self.openFile(at: indexPath.row, from: tableView)
        }
        cell.cancelBlock = { [weak self] in
            guard let self = self else { return }
            self.cancelFile(at: indexPath.row, from: tableView)
        }
        return cell
    }
}
