import Cocoa
import SnapKit

// Enum defining table types for clear distinction
enum TableType {
    case left
    case right
    case bottom
}

private final class CSMenuItemWithSwitchView: NSView {
    private let titleLabel = NSTextField(labelWithString: "")
    private let iconImageView = NSImageView()
    private let toggleButton = NSButton()
    private var isOnState: Bool
    private var switchAction: ((Bool) -> Void)?
    
    init(title: String, iconName: String, isOn: Bool, switchAction: @escaping (Bool) -> Void) {
        self.isOnState = isOn
        super.init(frame: NSRect(x: 0, y: 0, width: 250, height: 24))
        self.switchAction = switchAction
        
        iconImageView.image = NSImage(named: iconName)
        iconImageView.imageScaling = .scaleProportionallyDown
        addSubview(iconImageView)
        
        titleLabel.stringValue = title
        titleLabel.font = NSFont.systemFont(ofSize: 13)
        titleLabel.textColor = .labelColor
        titleLabel.isBordered = false
        titleLabel.isEditable = false
        titleLabel.backgroundColor = .clear
        addSubview(titleLabel)
        
        toggleButton.setButtonType(.toggle)
        toggleButton.isBordered = false
        toggleButton.title = ""
        toggleButton.imagePosition = .imageOnly
        toggleButton.target = self
        toggleButton.action = #selector(switchValueChanged)
        addSubview(toggleButton)
        
        updateToggleImage()
        
        setupLayout()
    }
    
    required init?(coder: NSCoder) {
        fatalError("init(coder:) has not been implemented")
    }
    
    private func setupLayout() {
        iconImageView.frame = NSRect(x: 8, y: 4, width: 16, height: 16)
        titleLabel.frame = NSRect(x: 30, y: 4, width: 140, height: 16)
        toggleButton.frame = NSRect(x: 188, y: 2, width: 44, height: 20)
    }
    
    private func updateToggleImage() {
        let imageName = isOnState ? "setting_runInBackground_on" : "setting_runInBackground_off"
        if let image = NSImage(named: imageName) {
            image.size = NSSize(width: 38, height: 20)
            toggleButton.image = image
        } else {
            logger.warn("\(imageName) image not found")
        }
    }
    
    @objc private func switchValueChanged() {
        isOnState.toggle()
        updateToggleImage()
        switchAction?(isOnState)
    }
}

class MainHomeViewController: NSViewController {

    // Table views
    private let leftTableView = NSTableView()
    private let rightTableView = NSTableView()
    private let bottomTableView = NSTableView()

    // Scroll view containers
    private let leftScrollView = NSScrollView()
    private let rightScrollView = NSScrollView()
    private let bottomScrollView = NSScrollView()

    private let bottomTitleLabel = NSTextField()
    // Middle view containing bottomTitleLabel and clearButton
    private let middleView = NSView()
    
    // Drag view for resizing FileBrowserView and TransferListView
    private let middleDragView = NSView()
    
    // Drag tracking
    private var fileBrowserViewHeightConstraint: Constraint?
    private var transferListViewHeightConstraint: Constraint?
    private var isDragging = false
    private var dragStartY: CGFloat = 0
    private var dragStartFileBrowserHeight: CGFloat = 0
    
    // Track last view size for layout changes
    private var lastViewSize: CGSize?
    var hasSetInitialRatio = false  // Flag to track if initial ratio has been set (public for AppDelegate)
    
    // Debounce timer for window resize save
    private var windowResizeSaveTimer: Timer?
    
    // Clear button for bottom table - lazy loaded
    private lazy var clearButton: NSButton = {
        let button = NSButton()
        button.bezelStyle = .texturedRounded
        button.setButtonType(.momentaryPushIn)
        button.isBordered = false
        button.title = ""
        
        if let clearImage = NSImage(named: "clearIcon") {
            // Adjust image size to match back button
            clearImage.size = NSSize(width: 20, height: 20)
            button.image = clearImage
            button.imageScaling = .scaleProportionallyDown
        } else {
            logger.warn("clearIcon image not found")
        }
        
        // Set highlight image
        if let highlightImage = NSImage(named: "clearIcon_highlight") {
            highlightImage.size = NSSize(width: 20, height: 20)
            button.alternateImage = highlightImage
        } else {
            logger.warn("clearIcon_highlight image not found")
        }
        
        button.target = self
        button.action = #selector(handleClearButtonTap)
        
        return button
    }()
    
    // Container views
    private let fileBrowserView = FileBrowserView()
    private let transferListView = TransferListView()

    // Independent data sources for three tables (right changed to [FileInfo])
    private var leftTableData: [FileInfo] = []
    private var rightTableData: [FileInfo] = []  // Right data source: FileInfo struct array
    
    // Sorting state for right table
    private enum SortColumn {
        case name
        case size
        case type
        case date
    }
    
    private enum SortOrder {
        case ascending
        case descending
    }
    
    private var currentSortColumn: SortColumn = .name
    private var currentSortOrder: SortOrder = .ascending
    
    // Device list (for DataTransmissionManager access)
    var deviceList: [CrossShareDevice] = [] {
        didSet {
            topView.refreshDeviceList(deviceList)
        }
    }

    // Bottom table data (for DataTransmissionManager access)
    var bottomTableData: [CSFileInfo] = []
    
    // Drag capability controlled by Go capability API
    private var isSupportFileDrag: Bool = false

    // Right side path and breadcrumb
    private var currentRightPath: String = ""
    private let rightHeaderContainer = NSView()
    private let backButton = NSButton(title: "Back", target: nil, action: nil)
    private let breadcrumbStack = NSStackView()

    lazy var topView: HomeHeaderView = {
        let cview = HomeHeaderView(frame: .zero)
        cview.wantsLayer = true
        cview.layer?.backgroundColor = ThemeManager.shared.currentTheme.topViewBackgroundColor.cgColor
        return cview
    }()
    
    // About view
    private var aboutView: CSAboutView?
    private var clickMonitor: Any?
    
    // License view
    private var licenseView: CSLicenseView?
    
    // Main view border layer
    private var mainViewBorderLayer: CAShapeLayer?
    
    // Red separator line below topView
    private lazy var topViewRedLine: NSView = {
        let line = NSView()
        line.wantsLayer = true
        line.layer?.backgroundColor = NSColor.red.cgColor
        return line
    }()
    
    // MARK: - File System Monitor
    // FSEvents stream for monitoring download directory changes
    private var fileSystemEventStream: FSEventStreamRef?
    private var lastRefreshTime: Date = .distantPast
    private let refreshDebounceInterval: TimeInterval = 0.5 // Debounce interval in seconds
    
    deinit {
        // Remove notification observers
        NotificationCenter.default.removeObserver(self)
        // Stop transfer monitoring
        DataTransmissionManager.shared.stopListening()
        // Remove click monitor
        removeClickMonitor()
        // Stop file system monitoring
        stopFileSystemMonitoring()
    }

    override func viewDidLoad() {
        super.viewDidLoad()
        setupUI()
        RealmDataManager.shared.setupRealm()
        setupShowData()
        setupNotifications()
        setupDataBlock()
        setupFileSystemMonitoring()
    }
    
    private func loadFileDragCapabilityFromHelper() {
        HelperClient.shared.getIsSupportFileDrag { [weak self] isSupported in
            DispatchQueue.main.async {
                guard let self = self else { return }
                self.isSupportFileDrag = isSupported
                self.applyRightTableDragCapability()
                logger.info("GetIsSupportFileDrag result via Helper: isSupportFileDrag=\(isSupported)")
            }
        }
    }
    
    @objc private func handleDeviceDiasStatusForFileDragCapability(_ notification: Notification) {
        logger.info("DIAS status notification received — refreshing IsSupportFileDrag from Helper")
        loadFileDragCapabilityFromHelper()
    }
    
    override func viewDidLayout() {
        super.viewDidLayout()
        // Add border after view layout is complete, ensuring bounds are determined
        if ThemeManager.shared.shouldShowMainViewBorder {
            addMainViewBorder()
        }
        
        // When window is resized, reapply the saved height ratio
        let currentSize = view.bounds.size
        
        if lastViewSize != currentSize {
            logger.debug("viewDidLayout - size changed from \(String(describing: lastViewSize)) to \(currentSize), hasSetInitialRatio: \(hasSetInitialRatio)")
            lastViewSize = currentSize
            
            // Only apply ratio if initial ratio has been set (to avoid overwriting initial setup)
            if hasSetInitialRatio && !isDragging {
                logger.debug("Applying current height ratio after window resize")
                applyCurrentHeightRatio()
            } else if !hasSetInitialRatio {
                logger.debug("Initial ratio not yet set, skipping")
            } else if isDragging {
                logger.debug("Currently dragging, skipping ratio application")
            }
        }
    }

    private func setupNotifications() {
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(handleDeviceDiasStatusForFileDragCapability(_:)),
            name: .deviceDiasStatusNotification,
            object: nil
        )
        
        // Listen for device list updates
        NotificationCenter.default.addObserver(
            forName: .deviceDataReceived, object: nil, queue: .main
        ) { ntf in
            guard let userInfo = ntf.userInfo as? [String: Any],
                let deviceList = userInfo["deviceList"] as? [CrossShareDevice]
            else {
                return
            }
            DispatchQueue.main.async {
                self.deviceList = deviceList
            }
        }
        
        // Start listening for transfer events
        DataTransmissionManager.shared.startListening(viewController: self) { [weak self] updatedData, isNewRecord, index in
            guard let self = self else { return }
            
            // Update data source
            self.bottomTableData = updatedData
            
            // Refresh UI
            DispatchQueue.main.async {
                if isNewRecord {
                    // New record: refresh entire table
                    self.bottomTableView.reloadData()
                } else if let index = index {
                    // Update record: only refresh corresponding row
                    let rowIndexSet = IndexSet(integer: index)
                    let columnIndexSet = IndexSet(integersIn: 0..<self.bottomTableView.numberOfColumns)
                    self.bottomTableView.reloadData(forRowIndexes: rowIndexSet, columnIndexes: columnIndexSet)
                }
                
                // Check if the updated item is completed and refresh FileBrowser if needed
                // Use the actual updated item (by index) instead of last item
                if let index = index, index < updatedData.count {
                    let updatedItem = updatedData[index]
                    if updatedItem.isCompleted {
                        logger.info("File transfer completed for item at index \(index)")
                        self.refreshFileBrowserIfNeeded(for: updatedItem)
                    }
                } else if isNewRecord, let lastItem = updatedData.last, lastItem.isCompleted {
                    // For new records, check the last item
                    logger.info("New file transfer record completed")
                    self.refreshFileBrowserIfNeeded(for: lastItem)
                }
            }
        }
    }
    
    // Update and refresh CSFileInfo data using callback pattern
    private func updateAndRefreshCSFileInfo(_ csFileInfo: CSFileInfo) {
        // Use RealmDataManager's callback method to handle data updates
        RealmDataManager.shared.updateCSFileInfo(csFileInfo, bottomTableData: self.bottomTableData) { result in
            switch result {
            case .success(let data):
                // Update data source
                self.bottomTableData = data.updatedData
                
                // Update UI on main thread
                DispatchQueue.main.async {
                    if data.isNewRecord {
                        // If it's a new record, refresh entire table
                        self.bottomTableView.reloadData()
                    } else if let index = data.index {
                        // If it's an update of existing record, only refresh related row
                        let rowIndexSet = IndexSet(integer: index)
                        let columnIndexSet = IndexSet(integersIn: 0..<self.bottomTableView.numberOfColumns)
                        self.bottomTableView.reloadData(forRowIndexes: rowIndexSet, columnIndexes: columnIndexSet)
                    }
                }
            case .failure(let error):
                logger.error("Failed to update CSFileInfo: \(error.localizedDescription)")
            }
        }
    }

    private func setupDataBlock() {
        HelperClient.shared.getDeviceListFromHelper { deviceDicList in
            DispatchQueue.main.async {
                self.deviceList = deviceDicList.compactMap {
                  CrossShareDevice(from: $0) }
                self.topView.refreshDeviceList(self.deviceList)
            }
        }
        
        // Setup button action blocks
        topView.tapMoreBtnBlock = { [weak self] in
            guard let self = self else { return }
            self.showOptionsMenu()
        }
        
        topView.tapDeviceListBtnBlock = { [weak self] in
            guard let self = self else { return }
            let btnFrameInView = self.topView.deviceBtn.convert(self.topView.deviceBtn.bounds, to: self.view)
            let buttonCenterX = btnFrameInView.midX
            let menuY = self.topView.frame.minY - 10
            self.showDeviceListMenu(buttonCenterX: buttonCenterX, menuY: menuY, in: self.view)
        }
    }
    
    // MARK: - Menu Actions
    private func showOptionsMenu() {
        // Create a simple pop-up menu
        let menu = NSMenu(title: "Options")
        menu.delegate = self
        
        // Add settings option with submenu
        let settingMenuItem = NSMenuItem(title: "setting", action: nil, keyEquivalent: "")
        if let image = NSImage(named: "setting") {
            image.size = NSSize(width: 16, height: 16)
            settingMenuItem.image = image
        }
        
        // Create settings submenu
        let settingSubmenu = NSMenu()
        settingSubmenu.delegate = self
        
        // Download path submenu item
        let downloadPathItem = NSMenuItem()
        let downloadPathView = CSMenuItemWithButtonView(
            title: "Download path",
            iconName: "settings_downloadLocation",
            buttonTitle: "Default",
            buttonIconName: "settings_folder"
        ) { [weak self] in
            self?.selectDownloadPath()
        }
        downloadPathItem.view = downloadPathView
        settingSubmenu.addItem(downloadPathItem)
        
        // Run in background submenu item
        let runInBackgroundItem = NSMenuItem()
        let runInBackgroundEnabled = HelperCommunication.shared.isRunInBackgroundEnabled
        logger.info("Run in background menu prepared, initial state: \(runInBackgroundEnabled)")
        let runInBackgroundView = CSMenuItemWithSwitchView(
            title: "Run in background",
            iconName: "setting_runInBackground",
            isOn: runInBackgroundEnabled
        ) { isOn in
            logger.info("Run in background toggled by user: \(isOn)")
            HelperCommunication.shared.setRunInBackgroundEnabled(isOn)
            if isOn {
                // ON: ensure login-item is registered immediately.
                HelperCommunication.shared.syncLoginItemRegistrationForRunInBackground(true) { success in
                    if success {
                        logger.info("RunInBackground ON: login-item registered")
                    } else {
                        logger.error("RunInBackground ON: failed to register login-item")
                    }
                }
            } else {
                // OFF: only persist flag here so helper stays alive until main app quits.
                // Login-item unregister runs in AppDelegate.applicationWillTerminate (OFF mode).
                logger.info("RunInBackground OFF: flag saved; login-item unregister deferred until main app terminates")
            }
        }
        runInBackgroundItem.view = runInBackgroundView
        
        // Bug report submenu item
        let bugReportItem = NSMenuItem()
        let bugReportView = CSMenuItemWithButtonView(
            title: "Bug report",
            iconName: "settings_bugReport",
            buttonTitle: "Export",
            buttonIconName: "settings_fileExport"
        ) { [weak self] in
            self?.exportBugReport()
        }
        bugReportItem.view = bugReportView
        settingSubmenu.addItem(bugReportItem)
        
        settingSubmenu.addItem(runInBackgroundItem)
        
        settingMenuItem.submenu = settingSubmenu
        menu.addItem(settingMenuItem)
        
        // Add separator
        menu.addItem(NSMenuItem.separator())
        
        // Add info option with submenu
        let infoMenuItem = NSMenuItem(title: "info", action: nil, keyEquivalent: "")
        if let image = NSImage(named: "CSinfo") {
            image.size = NSSize(width: 16, height: 16)
            infoMenuItem.image = image
        }
        
        // Create info submenu
        let infoSubmenu = NSMenu()
        infoSubmenu.delegate = self
        
        // About submenu item
        let aboutItem = NSMenuItem(title: "About", action: #selector(showAbout), keyEquivalent: "")
        aboutItem.target = self
        if let image = NSImage(named: "info_about") {
            image.size = NSSize(width: 16, height: 16)
            aboutItem.image = image
        }
        infoSubmenu.addItem(aboutItem)
        
        // Open Source license submenu item
        let licenseItem = NSMenuItem(title: "Open Source license", action: #selector(showLicense), keyEquivalent: "")
        licenseItem.target = self
        if let image = NSImage(named: "info_license") {
            image.size = NSSize(width: 16, height: 16)
            licenseItem.image = image
        }
        infoSubmenu.addItem(licenseItem)
        
        infoMenuItem.submenu = infoSubmenu
        menu.addItem(infoMenuItem)
        
        // Show menu at button position
        let menuPoint = NSPoint(x: self.view.frame.width - 130, y: self.topView.frame.minY - 10)
        menu.popUp(positioning: nil, at: menuPoint, in: self.view)
    }
    
    private func showDeviceListMenu(buttonCenterX: CGFloat, menuY: CGFloat, in view: NSView) {
        let menu = NSMenu(title: "Device list")
        menu.delegate = self
        
        if self.deviceList.isEmpty {
            let noDeviceItem = CustomMenuItem.createNoDeviceMenuItem(CSDeviceManager.shared.diasStatus)
            menu.addItem(noDeviceItem)
        } else {
            for (index, device) in self.deviceList.enumerated() {
                if index > 0 {
                    menu.addItem(NSMenuItem.separator())
                }
                let imageName = "Device_\(device.sourcePortType)"
                let deviceItem = CustomMenuItem.createDeviceMenuItem(title: device.deviceName, 
                                                                    imageName: imageName, 
                                                                    target: self, 
                                                                    tag: index + 1, 
                                                                    ipAddr: device.ipAddr)
                menu.addItem(deviceItem)
            }
        }
        
        var menuWidth: CGFloat = 0
        for item in menu.items {
            if let itemView = item.view {
                menuWidth = max(menuWidth, itemView.frame.width)
            }
        }
        
        let menuX = buttonCenterX - menuWidth / 2
        let targetPoint = NSPoint(x: menuX, y: menuY)
        menu.popUp(positioning: nil, at: targetPoint, in: view)
    }
    
    // MARK: - Settings Actions
    
    private func selectDownloadPath() {
        let openPanel = NSOpenPanel()
        openPanel.canChooseFiles = false
        openPanel.canChooseDirectories = true
        openPanel.canCreateDirectories = true
        openPanel.allowsMultipleSelection = false
        openPanel.prompt = "Select"
        
        // Set current directory to saved path or default
        let currentPath = CSUserPreferences.shared.getDownloadPathOrDefault()
        openPanel.directoryURL = URL(fileURLWithPath: currentPath)
        
        openPanel.begin { response in
            if response == .OK, let url = openPanel.url {
                let selectedPath = url.path
                logger.info("Selected download path: \(selectedPath)")
                
                // Update download path in Go service
                HelperClient.shared.requestUpdateDownloadPath(downloadPath: selectedPath) { success, error in
                    if success {
                        logger.info("Download path updated successfully in Go service: \(selectedPath)")
                    } else {
                        logger.error("Failed to update download path in Go service: \(error ?? "Unknown error")")
                    }
                }
                // Save the path to user preferences
                CSUserPreferences.shared.saveDownloadPath(selectedPath)
                
                // Restart file system monitoring for new download path
                self.stopFileSystemMonitoring()
                self.setupFileSystemMonitoring()
                
                // Show success message
                CSAlertManager.shared.showDownloadPathUpdated(path: selectedPath)
            }
        }
    }
    
    private func exportBugReport() {
        CSBugReportExporter.shared.exportBugReport { result in
            DispatchQueue.main.async {
                switch result {
                case .success(let zipFilePath):
                    CSAlertManager.shared.showBugReportExported(zipFilePath: zipFilePath)
                    
                case .failure(let error):
                    CSAlertManager.shared.showBugReportExportFailed(error: error)
                }
            }
        }
    }
    
    // MARK: - Info Actions
    
    @objc private func showAbout() {
        // If already displayed, hide it
        if aboutView != nil {
            hideAboutView()
            return
        }
        
        // Get AppDelegate's window
        guard let appDelegate = NSApplication.shared.delegate as? AppDelegate,
              let window = appDelegate.window,
              let contentView = window.contentView else {
            return
        }
        
        // Create CSAboutView
        let aboutViewSize = NSSize(width: 400, height: 350)
        let newAboutView = CSAboutView(frame: NSRect(origin: .zero, size: aboutViewSize))
        newAboutView.layer?.shadowColor = NSColor.black.cgColor
        newAboutView.layer?.shadowOpacity = 0.3
        newAboutView.layer?.shadowOffset = NSSize(width: 0, height: -2)
        newAboutView.layer?.shadowRadius = 8
        
        // Add to window's contentView
        contentView.addSubview(newAboutView)
        
        // Set constraints using SnapKit
        newAboutView.snp.makeConstraints { make in
            make.top.equalTo(topView.snp.bottom).offset(8)
            make.right.equalToSuperview().offset(-8)
            make.width.equalTo(aboutViewSize.width)
            make.height.equalTo(aboutViewSize.height)
        }
        
        // Save reference
        aboutView = newAboutView
        
        // Add fade-in animation
        newAboutView.alphaValue = 0
        newAboutView.layer?.transform = CATransform3DMakeScale(0.95, 0.95, 1.0)
        NSAnimationContext.runAnimationGroup { context in
            context.duration = 0.25
            context.timingFunction = CAMediaTimingFunction(name: .easeOut)
            newAboutView.animator().alphaValue = 1.0
            newAboutView.layer?.transform = CATransform3DIdentity
        }
        
        // Add click monitor to hide aboutView
        addClickMonitor()
    }
    
    private func hideAboutView() {
        guard let aboutView = aboutView else { return }
        
        // Remove click monitor
        removeClickMonitor()
        
        // Add fade-out animation
        NSAnimationContext.runAnimationGroup({ context in
            context.duration = 0.2
            context.timingFunction = CAMediaTimingFunction(name: .easeIn)
            aboutView.animator().alphaValue = 0
            aboutView.layer?.transform = CATransform3DMakeScale(0.95, 0.95, 1.0)
        }, completionHandler: {
            aboutView.removeFromSuperview()
            self.aboutView = nil
        })
    }
    
    private func addClickMonitor() {
        // Remove old monitor if exists
        removeClickMonitor()
        
        // Add local event monitor
        clickMonitor = NSEvent.addLocalMonitorForEvents(matching: [.leftMouseDown]) { [weak self] event in
            guard let self = self, let aboutView = self.aboutView else { return event }
            
            // Get click location
            let locationInWindow = event.locationInWindow
            let locationInAboutView = aboutView.convert(locationInWindow, from: nil)
            
            // Check if click is outside aboutView
            if !aboutView.bounds.contains(locationInAboutView) {
                self.hideAboutView()
            }
            
            return event
        }
    }
    
    private func removeClickMonitor() {
        if let monitor = clickMonitor {
            NSEvent.removeMonitor(monitor)
            clickMonitor = nil
        }
    }
    
    @objc private func showLicense() {
        // Create license menu
        let menu = NSMenu(title: "Open Source License")
        menu.delegate = self
        
        // First item: "open source license" with icon - using custom view
        let headerItem = NSMenuItem()
        let headerView = CSMenuHeaderView(title: "open source license", iconName: "info_license")
        headerItem.view = headerView
        menu.addItem(headerItem)
        
        // Dynamically add library items from CSLicenseManager
        let libraries = CSLicenseManager.shared.getAllLibraries()
        for (index, libraryName) in libraries.enumerated() {
            // Add separator before each item except the first
            if index > 0 {
                menu.addItem(NSMenuItem.separator())
            }
            
            // Create library menu item with arrow
            let libraryItem = NSMenuItem()
            let libraryView = CSMenuItemWithArrowView(title: libraryName, arrowIconName: "rightArrow") { [weak self] in
                self?.showLicenseDetail(libraryName: libraryName)
            }
            libraryItem.view = libraryView
            menu.addItem(libraryItem)
        }
        
        // Show menu at button position
        let menuPoint = NSPoint(x: self.view.frame.width - 198, y: self.topView.frame.minY - 10)
        menu.popUp(positioning: nil, at: menuPoint, in: self.view)
    }
    
    private func showLicenseDetail(libraryName: String) {
        // Hide about view if it's showing
        if aboutView != nil {
            hideAboutView()
        }
        
        // If already displayed, hide it
        if licenseView != nil {
            hideLicenseView()
            return
        }
        
        // Get AppDelegate's window
        guard let appDelegate = NSApplication.shared.delegate as? AppDelegate,
              let window = appDelegate.window,
              let contentView = window.contentView else {
            return
        }
        
        // Get license text for the library from CSLicenseManager
        let licenseText = CSLicenseManager.shared.getLicenseText(for: libraryName)
        
        // Create CSLicenseView
        let licenseViewSize = NSSize(width: 400, height: 350)
        let newLicenseView = CSLicenseView(
            frame: NSRect(origin: .zero, size: licenseViewSize),
            libraryName: libraryName,
            licenseText: licenseText
        )
        newLicenseView.layer?.shadowColor = NSColor.black.cgColor
        newLicenseView.layer?.shadowOpacity = 0.3
        newLicenseView.layer?.shadowOffset = NSSize(width: 0, height: -2)
        newLicenseView.layer?.shadowRadius = 8
        
        // Set back button action
        newLicenseView.onBackButtonTapped = { [weak self] in
            self?.hideLicenseView()
        }
        
        // Add to window's contentView
        contentView.addSubview(newLicenseView)
        
        // Set constraints using SnapKit
        newLicenseView.snp.makeConstraints { make in
            make.top.equalTo(topView.snp.bottom).offset(8)
            make.right.equalToSuperview().offset(-8)
            make.width.equalTo(licenseViewSize.width)
            make.height.equalTo(licenseViewSize.height)
        }
        
        // Save reference
        licenseView = newLicenseView
        
        // Add fade-in animation
        newLicenseView.alphaValue = 0
        newLicenseView.layer?.transform = CATransform3DMakeScale(0.95, 0.95, 1.0)
        NSAnimationContext.runAnimationGroup { context in
            context.duration = 0.25
            context.timingFunction = CAMediaTimingFunction(name: .easeOut)
            newLicenseView.animator().alphaValue = 1.0
            newLicenseView.layer?.transform = CATransform3DIdentity
        }
        
        // Add click monitor to hide licenseView
        addLicenseClickMonitor()
    }
    
    private func hideLicenseView() {
        guard let licenseView = licenseView else { return }
        
        // Remove click monitor
        removeLicenseClickMonitor()
        
        // Add fade-out animation
        NSAnimationContext.runAnimationGroup({ context in
            context.duration = 0.2
            context.timingFunction = CAMediaTimingFunction(name: .easeIn)
            licenseView.animator().alphaValue = 0
            licenseView.layer?.transform = CATransform3DMakeScale(0.95, 0.95, 1.0)
        }, completionHandler: {
            licenseView.removeFromSuperview()
            self.licenseView = nil
        })
    }
    
    private func addLicenseClickMonitor() {
        // Remove old monitor if exists
        removeLicenseClickMonitor()
        
        // Add local event monitor
        clickMonitor = NSEvent.addLocalMonitorForEvents(matching: [.leftMouseDown]) { [weak self] event in
            guard let self = self, let licenseView = self.licenseView else { return event }
            
            // Get click location
            let locationInWindow = event.locationInWindow
            let locationInLicenseView = licenseView.convert(locationInWindow, from: nil)
            
            // Check if click is outside licenseView
            if !licenseView.bounds.contains(locationInLicenseView) {
                self.hideLicenseView()
            }
            
            return event
        }
    }
    
    private func removeLicenseClickMonitor() {
        if let monitor = clickMonitor {
            NSEvent.removeMonitor(monitor)
            clickMonitor = nil
        }
    }
    
    @objc private func refreshDeviceList() {
        // Refresh device list
        HelperClient.shared.getDeviceListFromHelper { deviceDicList in
            DispatchQueue.main.async {
                self.deviceList = deviceDicList.compactMap { CrossShareDevice(from: $0) }
                self.topView.refreshDeviceList(self.deviceList)
            }
        }
    }
    
    // Initialize test data: call right data loading method
    private func setupShowData() {
        initLeftTableViewData()
        readRelamData()
        // Remove table separators
        bottomTableView.intercellSpacing = NSSize(width: 0, height: 0)
        bottomTableView.gridStyleMask = []
        // Refresh all tables
        leftTableView.reloadData()
        rightTableView.reloadData()
    }
    
    // Left table data loading (unchanged)
    private func initLeftTableViewData() {
        leftTableData.removeAll() // Clear old data
        
        guard let downloadsURL = FileManager.default.urls(for: .downloadsDirectory, in: .userDomainMask).first else {
            logger.error("Failed to retrieve downloads directory path")
            return
        }
        let downloadsPath = downloadsURL.path
        if let fileInfo = FileSystemInfoFetcher.getItemInfo(for: downloadsPath) {
            logger.debug("Left file info - Name: \(fileInfo.name), Path: \(fileInfo.path)")
            leftTableData.append(fileInfo)
        }

        // Dynamically get current user's Documents directory path
        guard let documentsURL = FileManager.default.urls(for: .documentDirectory, in: .userDomainMask).first else {
            logger.error("Failed to retrieve Documents directory path")
            return
        }
        let targetPath = documentsURL.path
        if let fileInfo = FileSystemInfoFetcher.getItemInfo(for: targetPath) {
            logger.debug("Left file info - Name: \(fileInfo.name), Path: \(fileInfo.path)")
            leftTableData.append(fileInfo)
        }
        
        // Dynamically get current user's Desktop directory path
        guard let desktopURL = FileManager.default.urls(for: .desktopDirectory, in: .userDomainMask).first else {
            logger.error("Failed to retrieve Desktop directory path")
            return
        }
        let DesktopPath = desktopURL.path
        if let fileInfo = FileSystemInfoFetcher.getItemInfo(for: DesktopPath) {
            logger.debug("Left file info - Name: \(fileInfo.name)")
            leftTableData.append(fileInfo)
        }
        
        // Initialize right table data (use first path in left table if available)
        guard let firstPath = leftTableData.first?.path else {
            logger.warn("No valid path in left table data to initialize right table")
            return
        }
        initRightTableViewData(firstPath)
    }
    
    // Right table data loading: directly add FileInfo to data source
    private func initRightTableViewData(_ targetPath: String) {
        let shouldResetMonitoring = currentRightPath != targetPath
        rightTableData.removeAll() // Clear old data
        currentRightPath = targetPath
        
        if let folderContents = FileSystemInfoFetcher.getFolderContentsInfo(in: targetPath) {
            logger.debug("Right folder contents loaded from: \(targetPath)")
            
            for item in folderContents {
                rightTableData.append(item)
            }
        } else {
            logger.error("Unable to get folder contents: \(targetPath)")
        }
        
        // Apply current sorting
        sortRightTableData()
        rightTableView.reloadData()
        updateBreadcrumbs(for: targetPath)
        
        if shouldResetMonitoring {
            restartFileSystemMonitoring(for: targetPath)
        }
    }
    
    // MARK: - Sorting Methods
    private func sortRightTableData() {
        rightTableData.sort { item1, item2 in
            let result: ComparisonResult
            
            switch currentSortColumn {
            case .name:
                result = item1.name.localizedCaseInsensitiveCompare(item2.name)
            case .size:
                let size1 = item1.fileSize ?? 0
                let size2 = item2.fileSize ?? 0
                result = size1 < size2 ? .orderedAscending : (size1 > size2 ? .orderedDescending : .orderedSame)
            case .type:
                let type1 = item1.fileType ?? ""
                let type2 = item2.fileType ?? ""
                result = type1.localizedCaseInsensitiveCompare(type2)
            case .date:
                let date1 = item1.modificationDate ?? ""
                let date2 = item2.modificationDate ?? ""
                result = date1.localizedCaseInsensitiveCompare(date2)
            }
            
            return currentSortOrder == .ascending ? result == .orderedAscending : result == .orderedDescending
        }
    }
    
    private func updateSortColumn(_ column: SortColumn) {
        if currentSortColumn == column {
            // Same column: toggle order
            currentSortOrder = currentSortOrder == .ascending ? .descending : .ascending
        } else {
            // Different column: set to ascending
            currentSortColumn = column
            currentSortOrder = .ascending
        }
        
        sortRightTableData()
        rightTableView.reloadData()
    }
    
    func readRelamData(){
        // Load historical data from Realm database
        bottomTableData = Array(RealmDataManager.shared.loadCSFileInfosFromRealm().reversed())
        bottomTableView.reloadData()
    }
        
    // MARK: - setupUI
    private func setupUI() {
        view.wantsLayer = true
        view.layer?.backgroundColor = ThemeManager.shared.currentTheme.backgroundColor.cgColor

        // Add views to main container
        view.addSubview(topView)
        view.addSubview(middleView)
        view.addSubview(middleDragView)
        view.addSubview(fileBrowserView)
        view.addSubview(transferListView)
        
        setupMiddleView()
        setupDragGesture()

        // Create overall blue dashed border view
        let combinedBorderView = createBorderView()

        // Configure table views
        setupTableView(leftTableView, scrollView: leftScrollView, type: .left)
        setupTableView(rightTableView, scrollView: rightScrollView, type: .right)
        setupTableView(bottomTableView, scrollView: bottomScrollView, type: .bottom)
        
        leftTableView.allowsMultipleSelection = false
        rightTableView.allowsMultipleSelection = true
        bottomTableView.allowsMultipleSelection = false
        // Ensure bottom table row height accommodates title + subtitle
        bottomTableView.rowHeight = 50
        
        // Setup FileBrowserView and TransferListView
        fileBrowserView.combinedBorderView = combinedBorderView
        fileBrowserView.leftScrollView = leftScrollView
        fileBrowserView.rightHeaderContainer = rightHeaderContainer
        fileBrowserView.rightScrollView = rightScrollView
        fileBrowserView.setupUI()
        
        transferListView.bottomScrollView = bottomScrollView
        transferListView.setupUI()
        
        // Add top view separator line if needed
        if ThemeManager.shared.shouldShowTopViewSeparator {
            view.addSubview(topViewRedLine)
        }
        
        // topView constraints - unchanged
        topView.snp.makeConstraints { make in
            if ThemeManager.shared.shouldShowMainViewBorder {
                make.left.equalTo(4)
                make.top.equalTo(4)
                make.right.equalTo(-4)
            } else {
                make.left.right.top.equalToSuperview()
            }
            make.height.equalTo(100)
        }
        
        // Top view separator line constraints
        if ThemeManager.shared.shouldShowTopViewSeparator {
            topViewRedLine.snp.makeConstraints { make in
                make.left.equalTo(2)
                make.right.equalTo(-2)
                make.top.equalTo(topView.snp.bottom)
                make.height.equalTo(1.5)
            }
        }
        
        // FileBrowserView constraints - will be updated by drag gesture
        fileBrowserView.snp.makeConstraints { make in
            make.leading.trailing.equalToSuperview()
            make.top.equalTo(topView.snp.bottom)
            // Height constraint will be updated during drag
            fileBrowserViewHeightConstraint = make.height.equalTo(400).constraint
        }
        
        // middleView constraints - positioned below fileBrowserView
        middleView.snp.makeConstraints { make in
            make.leading.equalToSuperview()
            make.width.equalTo(200)
            make.height.equalTo(36)
            make.top.equalTo(fileBrowserView.snp.bottom)
        }
        
        // middleDragView constraints - between middleView and self.view right edge
        middleDragView.snp.makeConstraints { make in
            make.leading.equalTo(middleView.snp.trailing)
            make.trailing.equalToSuperview()
            make.height.equalTo(middleView.snp.height)
            make.bottom.equalTo(middleView.snp.bottom)
        }
        
        // TransferListView constraints - will be updated by drag gesture
        transferListView.snp.makeConstraints { make in
            make.leading.trailing.equalToSuperview()
            make.top.equalTo(middleView.snp.bottom)
            // Height constraint will be updated during drag
            transferListViewHeightConstraint = make.height.equalTo(200).constraint
        }
        
        // Set initial ratio to 2:1 (FileBrowserView : TransferListView)
        DispatchQueue.main.async {
            self.setInitialRatio()
        }
    }
        
    func setupMiddleView(){
        // Setup middleView
        setupTitleLabel(bottomTitleLabel, text: "Files records")
        middleView.addSubview(bottomTitleLabel)
        middleView.addSubview(clearButton)

        // bottomTitleLabel constraints - relative to middleView
        bottomTitleLabel.snp.makeConstraints { make in
            make.leading.equalTo(middleView.snp.leading).offset(20)
            make.bottom.equalTo(middleView.snp.bottom).offset(-10)
        }
        
        // clearButton constraints - relative to middleView
        clearButton.snp.makeConstraints { make in
            make.leading.equalTo(bottomTitleLabel.snp.trailing).offset(10)
            make.bottom.equalTo(middleView.snp.bottom).offset(-10)
            make.width.equalTo(30)
            make.height.equalTo(20)
        }
    }
    
    // MARK: - Drag gesture for resizing FileBrowserView and TransferListView
    private func setupDragGesture() {
        let panGesture = NSPanGestureRecognizer(target: self, action: #selector(handlePanGesture(_:)))
        middleDragView.addGestureRecognizer(panGesture)
        
        // Setup tracking area for cursor change
        setupDragViewTrackingArea()
    }
    
    private func setupDragViewTrackingArea() {
        let trackingArea = NSTrackingArea(
            rect: middleDragView.bounds,
            options: [.mouseEnteredAndExited, .activeInKeyWindow, .inVisibleRect],
            owner: self,
            userInfo: nil
        )
        middleDragView.addTrackingArea(trackingArea)
    }
    
    // Handle mouse enter/exit events
    override func mouseEntered(with event: NSEvent) {
        // Check if the event is from middleDragView
        if let view = event.trackingArea?.owner as? MainHomeViewController,
           view == self {
            NSCursor.resizeUpDown.push()
        }
    }
    
    override func mouseExited(with event: NSEvent) {
        // Check if the event is from middleDragView
        if let view = event.trackingArea?.owner as? MainHomeViewController,
           view == self {
            NSCursor.pop()
        }
    }
    
    @objc private func handlePanGesture(_ gesture: NSPanGestureRecognizer) {
        let location = gesture.location(in: view)
        
        switch gesture.state {
        case .began:
            isDragging = true
            dragStartY = location.y
            dragStartFileBrowserHeight = fileBrowserView.bounds.height
            
        case .changed:
            guard isDragging else { return }
            
            // Calculate new fileBrowserView height
            let deltaY = dragStartY - location.y  // Up is positive (increases fileBrowserView)
            let newFileBrowserHeight = dragStartFileBrowserHeight + deltaY
            
            // Get available space (total height minus topView and middleView)
            let topViewHeight: CGFloat = 100
            let middleViewHeight: CGFloat = 36
            let availableHeight = view.bounds.height - topViewHeight - middleViewHeight
            
            // Calculate constraints: FileBrowserView : TransferListView
            // Min ratio: 1:2 (FileBrowserView gets 1/3)
            // Max ratio: 2:1 (FileBrowserView gets 2/3)
            let minFileBrowserHeight = availableHeight / 3.0
            let maxFileBrowserHeight = availableHeight * (2.0 / 3.0)
            
            // Clamp the height
            let clampedHeight = max(minFileBrowserHeight, min(maxFileBrowserHeight, newFileBrowserHeight))
            
            // Update constraints
            updateFileBrowserTransferListHeights(fileBrowserHeight: clampedHeight, availableHeight: availableHeight)
            
        case .ended, .cancelled:
            isDragging = false
            // Save the ratio when dragging ends
            saveCurrentFileBrowserHeightRatio()
            
        default:
            break
        }
    }
    
    private func updateFileBrowserTransferListHeights(fileBrowserHeight: CGFloat, availableHeight: CGFloat) {
        let transferListHeight = availableHeight - fileBrowserHeight
        
        logger.debug("updateFileBrowserTransferListHeights - fileBrowserHeight: \(fileBrowserHeight), transferListHeight: \(transferListHeight)")
        
        // Update existing constraints instead of deactivating and recreating
        fileBrowserViewHeightConstraint?.update(offset: fileBrowserHeight)
        transferListViewHeightConstraint?.update(offset: transferListHeight)
        
        // Update layout immediately
        NSAnimationContext.runAnimationGroup { context in
            context.duration = 0
            context.allowsImplicitAnimation = false
            view.layoutSubtreeIfNeeded()
        }
        
        logger.debug("After layout - fileBrowserView.bounds.height: \(fileBrowserView.bounds.height), transferListView.bounds.height: \(transferListView.bounds.height)")
    }
    
    private func setInitialRatio() {
        let topViewHeight: CGFloat = 100
        let middleViewHeight: CGFloat = 36
        let availableHeight = view.bounds.height - topViewHeight - middleViewHeight
        
        logger.debug("setInitialRatio called - view.bounds.height: \(view.bounds.height), availableHeight: \(availableHeight)")
        
        guard availableHeight > 0 else {
            // Retry after a short delay if height not available yet
            logger.debug("availableHeight <= 0, retrying after 0.1s")
            DispatchQueue.main.asyncAfter(deadline: .now() + 0.1) {
                self.setInitialRatio()
            }
            return
        }
        
        // Load saved ratio or use default 2:1 (FileBrowserView : TransferListView)
        let savedSettings = SharedDataManager.shared.getWindowSettings()
        let ratio = savedSettings?.fileBrowserHeightRatio ?? (2.0 / 3.0)
        
        logger.debug("Saved settings: width=\(savedSettings?.width ?? 0), height=\(savedSettings?.height ?? 0), ratio=\(savedSettings?.fileBrowserHeightRatio ?? 0)")
        logger.debug("Using ratio: \(ratio), default was: \(2.0 / 3.0)")
        
        let fileBrowserHeight = availableHeight * ratio
        logger.debug("Calculated fileBrowserHeight: \(fileBrowserHeight), transferListHeight: \(availableHeight - fileBrowserHeight)")
        
        updateFileBrowserTransferListHeights(fileBrowserHeight: fileBrowserHeight, availableHeight: availableHeight)
        
        // Mark that initial ratio has been set
        hasSetInitialRatio = true
        logger.info("Initial ratio set complete - now ready to save window settings")
    }
    
    // MARK: - Save/Restore Height Ratio
    
    /// Apply current saved height ratio
    private func applyCurrentHeightRatio() {
        let topViewHeight: CGFloat = 100
        let middleViewHeight: CGFloat = 36
        let availableHeight = view.bounds.height - topViewHeight - middleViewHeight
        
        guard availableHeight > 0 else {
            logger.debug("applyCurrentHeightRatio - availableHeight <= 0, skipping")
            return
        }
        
        // Get saved ratio
        let savedSettings = SharedDataManager.shared.getWindowSettings()
        let ratio = savedSettings?.fileBrowserHeightRatio ?? (2.0 / 3.0)
        
        logger.debug("applyCurrentHeightRatio - availableHeight: \(availableHeight), ratio: \(ratio)")
        
        let fileBrowserHeight = availableHeight * ratio
        updateFileBrowserTransferListHeights(fileBrowserHeight: fileBrowserHeight, availableHeight: availableHeight)
    }
    
    /// Get current fileBrowserView height ratio
    func getCurrentFileBrowserHeightRatio() -> Double {
        let topViewHeight: CGFloat = 100
        let middleViewHeight: CGFloat = 36
        let availableHeight = view.bounds.height - topViewHeight - middleViewHeight
        
        guard availableHeight > 0 else {
            logger.debug("getCurrentFileBrowserHeightRatio - availableHeight <= 0, returning default")
            return 2.0 / 3.0  // Default ratio
        }
        
        let currentHeight = fileBrowserView.bounds.height
        
        // If fileBrowserView hasn't been laid out yet, return default ratio
        if currentHeight <= 0 {
            logger.debug("getCurrentFileBrowserHeightRatio - fileBrowserView height is 0, returning default")
            return 2.0 / 3.0  // Default ratio
        }
        
        let ratio = Double(currentHeight / availableHeight)
        logger.debug("getCurrentFileBrowserHeightRatio - currentHeight: \(currentHeight), availableHeight: \(availableHeight), ratio: \(ratio)")
        return ratio
    }
    
    /// Save current fileBrowserView height ratio
    private func saveCurrentFileBrowserHeightRatio() {
        guard let window = view.window else {
            logger.debug("saveCurrentFileBrowserHeightRatio - window is nil")
            return
        }
        
        let ratio = getCurrentFileBrowserHeightRatio()
        let frame = window.frame
        
        logger.debug("Saving window settings - width: \(frame.width), height: \(frame.height), ratio: \(ratio)")
        
        SharedDataManager.shared.saveWindowSettings(
            width: Double(frame.width),
            height: Double(frame.height),
            fileBrowserHeightRatio: ratio
        )
        
        logger.debug("Window settings saved successfully")
    }
    
    // Helper method: Add blue dashed border to the view
    func addBlueDashedBorder(to view: NSView) {
        let borderLayer = CAShapeLayer()
        borderLayer.strokeColor = NSColor.blue.cgColor
        borderLayer.fillColor = nil
        borderLayer.lineWidth = 1.0
        borderLayer.lineDashPattern = [4, 2] // Dashed line style: 4 points solid, 2 points blank
        borderLayer.frame = view.bounds
        
        if #available(macOS 14.0, *) {
            borderLayer.path = NSBezierPath(roundedRect: view.bounds, xRadius: 4, yRadius: 4).cgPath
        } else {
            // Use extension method to convert to CGPath
            borderLayer.path = NSBezierPath(roundedRect: view.bounds, xRadius: 4, yRadius: 4).cgPath
        }
        
        view.layer?.addSublayer(borderLayer)
        
        // Listen for view frame changes to update border
        view.postsFrameChangedNotifications = true
        NotificationCenter.default.addObserver(forName: NSView.frameDidChangeNotification, object: view, queue: nil) {
            notification in
            borderLayer.frame = view.bounds
            if #available(macOS 14.0, *) {
                borderLayer.path = NSBezierPath(roundedRect: view.bounds, xRadius: 4, yRadius: 4).cgPath
            } else {
                borderLayer.path = NSBezierPath(roundedRect: view.bounds, xRadius: 4, yRadius: 4).cgPath
            }
        }
    }

    
    // Create border view with solid border color #52a5fe
    private func createBorderView() -> NSView {
        let combinedBorderView = NSView()
        combinedBorderView.wantsLayer = true
        combinedBorderView.layer?.backgroundColor = NSColor.clear.cgColor
        
        // Add solid border with color #52a5fe
        let borderLayer = CAShapeLayer()
        borderLayer.strokeColor = ThemeManager.shared.currentTheme.fileBrowserViewBorderColor?.cgColor
        borderLayer.fillColor = nil
        borderLayer.lineWidth = 1.0
        // No lineDashPattern for solid line
        borderLayer.frame = combinedBorderView.bounds
        
        if #available(macOS 14.0, *) {
            borderLayer.path = NSBezierPath(roundedRect: combinedBorderView.bounds, xRadius: 4, yRadius: 4).cgPath
        } else {
            borderLayer.path = NSBezierPath(roundedRect: combinedBorderView.bounds, xRadius: 4, yRadius: 4).cgPath
        }
        
        combinedBorderView.layer?.addSublayer(borderLayer)
        
        // Listen for view frame changes to update border
        combinedBorderView.postsFrameChangedNotifications = true
        NotificationCenter.default.addObserver(forName: NSView.frameDidChangeNotification, object: combinedBorderView, queue: nil) {
            notification in
            borderLayer.frame = combinedBorderView.bounds
            if #available(macOS 14.0, *) {
                borderLayer.path = NSBezierPath(roundedRect: combinedBorderView.bounds, xRadius: 4, yRadius: 4).cgPath
            } else {
                borderLayer.path = NSBezierPath(roundedRect: combinedBorderView.bounds, xRadius: 4, yRadius: 4).cgPath
            }
        }
        
        return combinedBorderView
    }

    
    // Add border to main view
    private func addMainViewBorder() {
        // Remove old border layer if exists
        mainViewBorderLayer?.removeFromSuperlayer()
        
        let theme = ThemeManager.shared.currentTheme
        let borderLayer = CAShapeLayer()
        borderLayer.strokeColor = theme.mainViewBorderColor.cgColor
        borderLayer.fillColor = nil
        borderLayer.lineWidth = theme.mainViewBorderWidth
        borderLayer.lineDashPattern = nil
        borderLayer.frame = view.bounds
        
        // Inset border for better visual effect
        let inset: CGFloat = 2.0
        let insetBounds = view.bounds.insetBy(dx: inset, dy: inset)
        
        // Use rounded rectangle path
        if #available(macOS 14.0, *) {
            borderLayer.path = NSBezierPath(roundedRect: insetBounds, xRadius: 8, yRadius: 8).cgPath
        } else {
            borderLayer.path = NSBezierPath(roundedRect: insetBounds, xRadius: 8, yRadius: 8).cgPath
        }
        
        view.layer?.addSublayer(borderLayer)
        mainViewBorderLayer = borderLayer
        
        // Listen for view size changes to update border
        view.postsFrameChangedNotifications = true
        NotificationCenter.default.addObserver(forName: NSView.frameDidChangeNotification, object: view, queue: nil) { [weak self] _ in
            guard let self = self, let borderLayer = self.mainViewBorderLayer else { return }
            borderLayer.frame = self.view.bounds
            let inset: CGFloat = 2.0
            let insetBounds = self.view.bounds.insetBy(dx: inset, dy: inset)
            if #available(macOS 14.0, *) {
                borderLayer.path = NSBezierPath(roundedRect: insetBounds, xRadius: 8, yRadius: 8).cgPath
            } else {
                borderLayer.path = NSBezierPath(roundedRect: insetBounds, xRadius: 8, yRadius: 8).cgPath
            }
        }
    }
    
    // Title label configuration
    private func setupTitleLabel(_ label: NSTextField, text: String) {
        label.stringValue = text
        label.isEditable = false
        label.isBordered = false
        label.backgroundColor = .clear
        label.textColor = ThemeManager.shared.currentTheme.filesRecordsTextColor
        label.font = NSFont.systemFont(ofSize: 14, weight: .medium)
    }
    
    // Table column configuration (unchanged)
    private func setupTableView(_ tableView: NSTableView, scrollView: NSScrollView, type: TableType) {
        scrollView.borderType = .bezelBorder
        scrollView.hasVerticalScroller = true
        scrollView.hasHorizontalScroller = true
        scrollView.autohidesScrollers = true
        
        if(type == .bottom){
            tableView.backgroundColor = NSColor(white: 0.95, alpha: 1.0)
        }else{
            tableView.backgroundColor = .white
        }
        tableView.headerView = type == .right ? NSTableHeaderView(frame: NSRect(x: 0, y: 0, width: 0, height: 24)) : nil
        tableView.dataSource = self
        tableView.delegate = self
        
        // Enable sorting for right table
        if type == .right {
            tableView.sortDescriptors = []
        }
        tableView.selectionHighlightStyle = .regular // Restore visible selection highlight, auto-cancel later
        // Right table supports double-click to enter directory and right-click menu
        if type == .right {
            tableView.target = self
            tableView.doubleAction = #selector(handleRightTableDoubleClick(_:))
            tableView.menu = createRightClickMenu()
        }
        
        // [New] Register drag types only for right table (support file URL and text, adapt to FileInfo data)
        if type == .right {
            applyRightTableDragCapability()
        }
        
        switch type {
        case .left, .bottom:
            let nameColumn = NSTableColumn(identifier: NSUserInterfaceItemIdentifier("name"))
            nameColumn.title = "Name"
            nameColumn.width = type == .left ? 180 : 300
            tableView.addTableColumn(nameColumn)
            
        case .right:
            // Name column
            let nameColumn = NSTableColumn(identifier: NSUserInterfaceItemIdentifier("name"))
            nameColumn.title = "Name"
            nameColumn.width = 200
            nameColumn.sortDescriptorPrototype = NSSortDescriptor(key: "name", ascending: true)
            tableView.addTableColumn(nameColumn)
            
            // Size column
            let sizeColumn = NSTableColumn(identifier: NSUserInterfaceItemIdentifier("size"))
            sizeColumn.title = "Size"
            sizeColumn.width = 80
            sizeColumn.sortDescriptorPrototype = NSSortDescriptor(key: "fileSize", ascending: true)
            tableView.addTableColumn(sizeColumn)
            
            // Type column
            let typeColumn = NSTableColumn(identifier: NSUserInterfaceItemIdentifier("type"))
            typeColumn.title = "Type"
            typeColumn.width = 120
            typeColumn.sortDescriptorPrototype = NSSortDescriptor(key: "fileType", ascending: true)
            tableView.addTableColumn(typeColumn)

            
            // Date column
            let dateColumn = NSTableColumn(identifier: NSUserInterfaceItemIdentifier("date"))
            dateColumn.title = "Date Modified"
            dateColumn.width = 120
            dateColumn.sortDescriptorPrototype = NSSortDescriptor(key: "modificationDate", ascending: true)
            tableView.addTableColumn(dateColumn)
        }
        
        scrollView.documentView = tableView
        tableView.frame = scrollView.bounds
        tableView.autoresizingMask = [.width, .height]
    }
    
    
    // Handle open file action
    @objc private func handleOpenFile(_ sender: Any?) {
        let clickedRow = rightTableView.clickedRow
        guard clickedRow >= 0, let item = rightTableData[safe: clickedRow], !item.isDirectory else { return }
        
        let url = URL(fileURLWithPath: item.path)
        NSWorkspace.shared.open(url)
    }
    
    // Handle send to device action
    @objc private func handleSendToDevice(_ sender: NSMenuItem) {
        let selectedRowIndexes = rightTableView.selectedRowIndexes

        // Filter out folders and invalid row indexes
        let selectedItems = selectedRowIndexes.compactMap { index -> FileInfo? in
            guard index >= 0, let item = rightTableData[safe: index] else { return nil }
            return item
        }

        // Ensure there are selected file items
        guard !selectedItems.isEmpty else {
            logger.warn("No valid files selected for sending")
            return
        }

        // Get device information
        guard sender.tag - 1 < self.deviceList.count, let device = self.deviceList[safe: sender.tag-1] else {
            logger.error("Device index out of range: \(sender.tag - 1)")
            return
        }

        // Get device ID and IP address
        let deviceId = device.id
        guard let deviceIp = device.ipAddress else {
            logger.error("Failed to get device IP address for device: \(device.deviceName)")
            return
        }
        
        // Build file path list (support multiple selection)
        let pathList = selectedItems.map { $0.path }
        
        // Build JSON data
        let fileDropData: [String: Any] = [
            "Id": deviceId,
            "Ip": deviceIp,
            "PathList": pathList
        ]

        // Convert to JSON string
        do {
            let jsonData = try JSONSerialization.data(withJSONObject: fileDropData, options: [])
            guard let jsonString = String(data: jsonData, encoding: .utf8) else {
                logger.error("Failed to convert JSON data to string")
                return
            }

            logger.info("Sending \(selectedItems.count) files to device: \(device.deviceName)")

            // Call send method
            HelperClient.shared.sendMultiFilesDropRequest(multiFilesData: jsonString) { success, statusString in
                if success {
                    logger.info("Successfully sent file request to \(device.deviceName), status: \(statusString ?? "")")
                } else {
                    logger.error("Failed to send file request to \(device.deviceName), status: \(statusString ?? "")")
                    let str = statusString ?? ""
                    if(str.count > 0){
                        CSToastManager.shared.showWarning(str)
                    }
                }
            }
        } catch {
            logger.error("Error creating JSON for file transfer: \(error)")
        }
    }
    
    // Handle clear button tap for bottom table
    @objc private func handleClearButtonTap() {
        // Clear all data in bottomTableData array
        bottomTableData.removeAll()
        // Refresh bottomTableView
        bottomTableView.reloadData()
        RealmDataManager.shared.deleteAllData()
    }
    
    // Create CSFileInfo model from userInfo dictionary
    private func createCSFileInfo(from userInfo: [String: Any]) -> CSFileInfo? {
        guard let sessionDict = userInfo["session"] as? [String: Any],
              let sessionId = userInfo["sessionId"] as? String,
              let senderID = userInfo["senderID"] as? String,
              let progress = userInfo["progress"] as? Double,
              let isCompletedInt = userInfo["isCompleted"] as? Int
        else {
            logger.error("The data required to create CSFileInfo is missing")
            return nil
        }
        
        // Extract errCode (optional)
        let errCode = userInfo["errCode"] as? Int
        
        // Create FileTransferSession instance
        if let fileTransferSession = createFileTransferSession(from: sessionDict) {
            return CSFileInfo(
                session: fileTransferSession,
                sessionId: sessionId,
                senderID: senderID,
                isCompleted: isCompletedInt != 0,
                progress: progress,
                errCode: errCode
            )
        }
        
        return nil
    }
    
    // Create FileTransferSession instance from dictionary
    private func createFileTransferSession(from dict: [String: Any]) -> FileTransfer? {
        // Try to use FileTransferSession's init?(from dict: [String: Any]) initialization method
        if let fileTransferSession = FileTransfer(from: dict) {
            // Call separate method to find matching device and update device name
            return findAndUpdateDeviceName(for: dict, with: fileTransferSession)
        }
        return nil
    }
    
    private func findAndUpdateDeviceName(for dict: [String: Any], with fileTransferSession: FileTransfer) -> FileTransfer? {
        // Find matching device and set deviceName
        if let senderID = dict["senderID"] as? String, let senderIP = dict["senderIP"] as? String {
            // Iterate through deviceList to find matching device
            for device in deviceList {
                // Check if senderID and ipAddr match
                if device.id == senderID && device.ipAddr == senderIP {
                    logger.debug("Found matching device: \(device.deviceName)")
                    
                    // Create a new FileTransfer instance but add deviceName key to newDict
                    var newDict = dict
                    newDict["deviceName"] = device.deviceName
                    
                    // First try to create a new FileTransfer instance using the new dictionary
                    if let newFileTransferSession = FileTransfer(from: newDict) {
                        return newFileTransferSession
                    }
                    
                    // If the above method fails, return the original instance
                    return fileTransferSession
                }
            }
            logger.warn("No matching device found for senderID: \(senderID), senderIP: \(senderIP)")
        }
        
        return fileTransferSession
    }
    
    
    // Table type determination (unchanged)
    private func getTableType(for tableView: NSTableView) -> TableType {
        if tableView === leftTableView {
            return .left
        } else if tableView === rightTableView {
            return .right
        } else {
            return .bottom
        }
    }
    
    // Handle table item tap method
    private func handleTableItemTap(_ tableView: NSTableView, at row: Int) {
        let tableType = getTableType(for: tableView)
        
        switch tableType {
        case .left:
            guard let item = leftTableData[safe: row] else { return }
            logger.debug("Tapped left table row \(row+1): \(item.name)")
            // Left table tap event handling logic...
            initRightTableViewData(item.path)
            
        case .right:
            guard let item = rightTableData[safe: row] else { return }
            logger.debug("Tapped right table row \(row+1): \(item.name) (\(item.isDirectory ? "Folder" : "File"))")
            // Single click no longer enters directory, wait for double-click handling
            
        case .bottom:
            guard let fileInfo = bottomTableData[safe: row] else { return }
            logger.debug("Tapped bottom table row \(row+1): \(fileInfo.session.currentFileName)")
            // Bottom table tap event handling logic...
        }
    }

}

// MARK: - 4. Core adjustment: NSTableViewDataSource (right data mapping)
extension MainHomeViewController: NSTableViewDataSource {
    func numberOfRows(in tableView: NSTableView) -> Int {
        switch getTableType(for: tableView) {
        case .left: return leftTableData.count
        case .right: return rightTableData.count // Right row count = FileInfo array count
        case .bottom: return bottomTableData.count
        }
    }
    
    func tableView(_ tableView: NSTableView, objectValueFor tableColumn: NSTableColumn?, row: Int) -> Any? {
        guard let columnId = tableColumn?.identifier.rawValue else { return nil }
        let tableType = getTableType(for: tableView)
        
        switch tableType {
        // Left table: logic unchanged
        case .left:
            guard let fileInfo = leftTableData[safe: row] else { return "" }
            return fileInfo.name
            
        case .bottom:
            guard let fileInfo = bottomTableData[safe: row] else { return "" }
            return fileInfo.senderID

        // Right table: extract corresponding column data from FileInfo
        case .right:
            guard let fileInfo = rightTableData[safe: row] else { return "" }
            
            // Return formatted property based on column identifier
            switch columnId {
            case "name": // Name column: directly return FileInfo.name
                return fileInfo.name
                
            case "size": // Size column: show "---" for folders, formatted size for files
                return fileInfo.isDirectory ? "" : UtilsHelper.formatFileSize(fileInfo.fileSize ?? 0)
                
            case "type": // Type column: show Folder for folders; fileType for files
                return fileInfo.fileType ?? (fileInfo.isDirectory ? "Folder" : "File")
            case "date": // Date column: format Date to string, show "Unknown" if no date
                return (fileInfo.modificationDate != nil) ? fileInfo.modificationDate  : "Unknown date"
            default:
                return ""
            }
        }
    }
}

// MARK: - NSTableViewDelegate and NSDraggingSource implementation (support multi-select drag)
extension MainHomeViewController: NSTableViewDelegate, NSDraggingSource {
    // Double-click handling: folders enter directory, files open with system default app
    @objc private func handleRightTableDoubleClick(_ sender: Any?) {
        guard let tableView = sender as? NSTableView ?? (sender == nil ? rightTableView : nil) else { return }
        let row = tableView.clickedRow
        guard row >= 0, let item = rightTableData[safe: row] else { return }
        
        if item.isDirectory {
            // Folder: enter directory
            initRightTableViewData(item.path)
        } else {
            // File: open with system default application
            let url = URL(fileURLWithPath: item.path)
            NSWorkspace.shared.open(url)
        }
    }
    // Configure right header (Back button + breadcrumb)
    private func setupRightHeaderIfNeeded() {
        guard rightHeaderContainer.subviews.isEmpty else { return }
        
        // Add background if theme requires it
        if ThemeManager.shared.shouldShowRightHeaderBackground,
           let bgColor = ThemeManager.shared.currentTheme.rightHeaderBackgroundColor {
            rightHeaderContainer.wantsLayer = true
            rightHeaderContainer.layer?.backgroundColor = bgColor.cgColor
            rightHeaderContainer.layer?.cornerRadius = 4
        }
        
        // Back button
        backButton.target = self
        backButton.action = #selector(backButtonTapped)
        // Set back button image, remove text
        backButton.title = ""
        if let backImage = NSImage(named: "backArrows") {
            // Adjust image size
            backImage.size = NSSize(width: 12, height: 12) // Set smaller image size
            backButton.image = backImage
            backButton.imageScaling = .scaleProportionallyDown // Set image scaling mode
        } else {
            logger.warn("backArrows image not found")
        }
        backButton.bezelStyle = .rounded
        rightHeaderContainer.addSubview(backButton)
        
        // Breadcrumb stack
        breadcrumbStack.orientation = .horizontal
        breadcrumbStack.alignment = .centerY
        breadcrumbStack.distribution = .fillProportionally
        breadcrumbStack.spacing = 6
        rightHeaderContainer.addSubview(breadcrumbStack)
        
        // Set constraints based on theme padding
        let padding = ThemeManager.shared.currentTheme.rightHeaderPadding
        
        backButton.snp.makeConstraints { make in
            make.leading.equalTo(rightHeaderContainer.snp.leading).offset(padding)
            make.centerY.equalTo(rightHeaderContainer.snp.centerY)
        }
        
        breadcrumbStack.snp.makeConstraints { make in
            make.leading.equalTo(backButton.snp.trailing).offset(8)
            make.trailing.lessThanOrEqualTo(rightHeaderContainer.snp.trailing).offset(padding > 0 ? -padding : 0)
            make.centerY.equalTo(rightHeaderContainer.snp.centerY)
        }
    }
    
    private func updateBreadcrumbs(for path: String) {
        setupRightHeaderIfNeeded()
        // Clear old breadcrumbs
        for view in breadcrumbStack.arrangedSubviews {
            breadcrumbStack.removeArrangedSubview(view)
            view.removeFromSuperview()
        }
        
        // Parse path into components (ignore root "/")
        let components = (path as NSString).pathComponents.filter { $0 != "/" }
        guard !components.isEmpty else { return }
        
        // Gradually build subpaths
        for (index, name) in components.enumerated() {
            let subComponents = Array(components.prefix(index + 1))
            let subPath = NSString.path(withComponents: ["/"] + subComponents)
            let button = NSButton(title: name, target: self, action: #selector(breadcrumbTapped(_:)))
            button.bezelStyle = .inline
            button.toolTip = subPath // Temporarily store path in toolTip
            breadcrumbStack.addArrangedSubview(button)
        }
    }
    
    @objc private func backButtonTapped() {
        guard !currentRightPath.isEmpty else { return }
        // If already at the root directory, do not perform the back operation
        if currentRightPath == "/" {
            return
        }
        let parent = URL(fileURLWithPath: currentRightPath).deletingLastPathComponent().path
        // Ensure the path is "/" when returning to the root directory
        let targetPath = parent.isEmpty ? "/" : parent
        initRightTableViewData(targetPath)
    }
    
    @objc private func breadcrumbTapped(_ sender: NSButton) {
        guard let path = sender.toolTip, !path.isEmpty else { return }
        initRightTableViewData(path)
    }
    
    private func applyRightTableDragCapability() {
        let dragTypes: [NSPasteboard.PasteboardType] = [.fileURL, .string]
        if isSupportFileDrag {
            rightTableView.registerForDraggedTypes(dragTypes)
            rightTableView.setDraggingSourceOperationMask(.copy, forLocal: true)
        } else {
            rightTableView.unregisterDraggedTypes()
            rightTableView.setDraggingSourceOperationMask([], forLocal: true)
        }
    }
    // Provide independent pasteboard writer for each row, ensuring drag items match pasteboard items one-to-one
    func tableView(_ tableView: NSTableView, pasteboardWriterForRow row: Int) -> NSPasteboardWriting? {
        guard getTableType(for: tableView) == .right else { return nil }
        guard isSupportFileDrag else { return nil }
        guard let item = rightTableData[safe: row] else { return nil }
        // Use NSURL as pasteboard writer, system automatically handles as draggable file URL
        return NSURL(fileURLWithPath: item.path)
    }

    func tableView(_ tableView: NSTableView, viewFor tableColumn: NSTableColumn?, row: Int) -> NSView? {
        guard let columnId = tableColumn?.identifier.rawValue else { return nil }
        let tableType = getTableType(for: tableView)
        
        // Special handling for bottom table with custom cell
        if tableType == .bottom && columnId == "name" {
            // Set row height for bottom table
            tableView.rowHeight = 80
            
            // Try to reuse existing custom cell
            let cellIdentifier = CSBottomTableViewCell.identifier
            if let cell = tableView.makeView(withIdentifier: cellIdentifier, owner: nil) as? CSBottomTableViewCell {
                // Configure cell with data
                if let fileInfo = bottomTableData[safe: row] {
                    // Configure cell directly using CSFileInfo
            cell.configure(with: fileInfo)
            
            // Set up delete closure
            cell.onDelete = { sessionId in
                        self.deleteFileTransferRecord(with: sessionId, from: self.bottomTableData)
                    }
                    
                    // Set up cancel closure
                    cell.onCancel = { sessionId in
                        self.cancelFileTransfer(with: sessionId)
                    }
                }
                return cell
            }
            
            // Create new custom cell if no reusable cell is found
            let cell = CSBottomTableViewCell()
            cell.identifier = cellIdentifier
            
            // Configure cell with data
            if let fileInfo = bottomTableData[safe: row] {
                // Configure cell directly using CSFileInfo
            cell.configure(with: fileInfo)
                // Set up delete closure
                cell.onDelete = { sessionId in
                    self.deleteFileTransferRecord(with: sessionId, from: self.bottomTableData)
                }
                
                // Set up cancel closure
                cell.onCancel = { sessionId in
                    self.cancelFileTransfer(with: sessionId)
                }
            }
            
            return cell
        }
        
        // Regular handling for other tables
        let cellIdentifier = NSUserInterfaceItemIdentifier(columnId)
        
        // Reuse existing cell
        if let cell = tableView.makeView(withIdentifier: cellIdentifier, owner: nil) as? NSTableCellView {
            cell.textField?.stringValue = self.tableView(tableView, objectValueFor: tableColumn, row: row) as? String ?? ""
            
            // Icon setting for left and right tables
            if (tableType == .left || tableType == .right) && columnId == "name" {
                if tableType == .left {
                    if let item = leftTableData[safe: row] {
                        cell.imageView?.image = NSWorkspace.shared.icon(forFile: item.path)
                    }
                } else if tableType == .right {
                    if let item = rightTableData[safe: row] {
                        cell.imageView?.image = NSWorkspace.shared.icon(forFile: item.path)
                    }
                }
            }
            return cell
        }
        
        // Create new cell for other tables
        let cell = NSTableCellView()
        cell.identifier = cellIdentifier
        
        // Add icon view for left and right tables
        if tableType == .left || tableType == .right {
            tableView.rowHeight = 30
            if columnId == "name" {
                let imageView = NSImageView(frame: NSRect(x: 0, y: 0, width: 16, height: 16))
                
                // Determine icon based on table type and data
                if tableType == .left {
                    if let item = leftTableData[safe: row] {
                        imageView.image = NSWorkspace.shared.icon(forFile: item.path)
                    }
                } else if tableType == .right {
                    if let item = rightTableData[safe: row] {
                        imageView.image = NSWorkspace.shared.icon(forFile: item.path)
                    }
                }

                cell.imageView = imageView
                cell.addSubview(imageView)
                
                // Set constraints for image view
                imageView.snp.makeConstraints { make in
                    make.centerY.equalToSuperview()
                    make.leading.equalToSuperview().offset(5)
                    make.width.height.equalTo(16)
                }
            }
        }
        
        // Regular text field for other tables
        let textField = NSTextField()
        textField.isEditable = false
        textField.isBordered = false
        textField.backgroundColor = .clear
        textField.stringValue = self.tableView(tableView, objectValueFor: tableColumn, row: row) as? String ?? ""
        textField.maximumNumberOfLines = 1
        textField.cell?.wraps = false
        cell.textField = textField
        cell.addSubview(textField)
        
        // Cell padding (avoid text sticking to edge)
        textField.snp.makeConstraints { make in
            if tableType == .left || tableType == .right {
                // Add left padding for text field to make space for the icon
                make.leading.equalTo(cell.imageView?.snp.trailing ?? 0).offset(5)
                make.trailing.equalToSuperview().inset(5)
                make.centerY.equalToSuperview()
            } else {
                make.edges.equalToSuperview().inset(5)
            }
        }
        
        return cell
    }
    
    func deleteFileTransferRecord(with sessionId: String, from bottomTableData: [CSFileInfo]) {
        self.bottomTableData = RealmDataManager.shared.deleteFileTransferRecord(with: sessionId, from: self.bottomTableData)
        self.bottomTableView.reloadData()
    }
    
    // Cancel file transfer
    func cancelFileTransfer(with sessionId: String) {
        // Find corresponding file transfer information
        guard let fileInfo = bottomTableData.first(where: { $0.sessionId == sessionId }) else {
            logger.warn("No transfer record found for sessionId: \(sessionId)")
            return
        }
        
        let ipPort = fileInfo.session.senderIP
        let clientID = fileInfo.session.senderID
        
        // Parse timestamp from sessionId (format: "senderID-timestamp")
        guard let timeStamp = extractTimestampFromSessionId(sessionId) else {
            logger.error("Unable to extract timestamp from sessionId: \(sessionId)")
            return
        }
        
        logger.info("Canceling file transfer - IP:\(ipPort), ClientID:\(clientID), TimeStamp:\(timeStamp)")
        
        // Call XPC service to cancel transfer via HelperClient
        HelperClient.shared.setCancelFileTransfer(ipPort: ipPort, clientID: clientID, timeStamp: timeStamp) { success, error in
            if success {
                logger.info("File transfer canceled successfully for session: \(sessionId)")
            } else {
                logger.error("Failed to cancel file transfer: \(error ?? "Unknown error")")
            }
        }
    }
    
    // Extract timestamp from sessionId
    // sessionId format: "senderID-timestamp"
    private func extractTimestampFromSessionId(_ sessionId: String) -> UInt64? {
        let components = sessionId.components(separatedBy: "-")
        guard components.count >= 2,
              let timestamp = UInt64(components.last ?? "") else {
            return nil
        }
        return timestamp
    }
    
    // Handle error event
    private func handleErrorEvent(_ errorInfo: [String: Any]) {
        logger.warn("GUI received error event")
        logger.debug("Error info: \(errorInfo)")
        
        // Extract id and timestamp (both string types)
        guard let id = errorInfo["id"] as? String else {
            logger.error("Failed to extract id from error event")
            return
        }
        
        guard let timestamp = errorInfo["timestamp"] as? String else {
            logger.error("Failed to extract timestamp from error event")
            return
        }
        
        // Extract errCode
        let errCode = errorInfo["errCode"] as? Int
        
        // Assemble into "id-timestamp" format
        let assembledSessionId = "\(id)-\(timestamp)"
        logger.debug("Processing error for SessionId: \(assembledSessionId), errCode: \(errCode ?? -1)")
        
        // Find matching sessionId in bottomTableData
        if var matchedItem = bottomTableData.first(where: { $0.sessionId == assembledSessionId }) {
            logger.info("Found matching item for error event, device: \(matchedItem.session.deviceName)")
            
            // Update only errCode
            matchedItem.errCode = errCode
            
            // Save to database and refresh corresponding cell
            updateAndRefreshCSFileInfo(matchedItem)
        } else {
            logger.warn("No matching SessionId found for error event: \(assembledSessionId)")
        }
    }
    
    func tableViewSelectionDidChange(_ notification: Notification) {
        guard let tableView = notification.object as? NSTableView else { return }
        let selectedRow = tableView.selectedRow
        guard selectedRow != -1 else { return } // Ignore deselection cases
        handleTableItemTap(tableView, at: selectedRow)
        // Only briefly highlight left and bottom tables; right table keeps selection for multi-select and drag
        let tableType = getTableType(for: tableView)
        if tableType == .left || tableType == .bottom {
            DispatchQueue.main.asyncAfter(deadline: .now() + 0.2) { [weak tableView] in
                tableView?.deselectAll(nil)
            }
        }
    }
    
    // Create right-click context menu for files
    private func createRightClickMenu() -> NSMenu {
        let menu = NSMenu()
        // Dynamic menu: Update items before popup through delegate
        menu.delegate = self
        return menu
    }
    
    // Update menu items before showing (similar to showOptionsMenu style)
    func menuNeedsUpdate(_ menu: NSMenu) {
        guard menu === rightTableView.menu else { return }
        menu.removeAllItems()

        // Calculate current selection and clicked row
        let selectedIndexes = rightTableView.selectedRowIndexes
        let clickedRow = rightTableView.clickedRow
        let selectionCount = selectedIndexes.count

        // Build Send to submenu (similar to showOptionsMenu)
        let sendToItem = NSMenuItem(title: "Send to", action: nil, keyEquivalent: "")
        let sendToSubmenu = NSMenu()
        sendToSubmenu.delegate = self
        
        // Iterate through device names using index to set tag
        for (index, device) in self.deviceList.enumerated() {
            // Add separator before non-first device items
            if index > 0 {
                sendToSubmenu.addItem(NSMenuItem.separator())
            }
            
            let deviceItem = NSMenuItem(title: device.deviceName, action: #selector(handleSendToDevice(_:)), keyEquivalent: "")
            deviceItem.target = self
            deviceItem.tag = index + 1
            
            // Set attributed title to ensure text is visible when highlighted
            let deviceAttributedTitle = NSAttributedString(
                string: device.deviceName,
                attributes: [.foregroundColor: NSColor.controlTextColor]
            )
            deviceItem.attributedTitle = deviceAttributedTitle
            
            // Add DeviceComputer icon
            let imageName = "Device_\(device.sourcePortType)"
            if let deviceIcon = NSImage(named: imageName) {
                deviceIcon.size = NSSize(width: 16, height: 16)
                deviceItem.image = deviceIcon
            }
            
            sendToSubmenu.addItem(deviceItem)
        }
        
        
        sendToItem.submenu = sendToSubmenu

        if selectionCount <= 1 {
            // Single selection: Show Open when clicking on a file
            if clickedRow >= 0, let item = rightTableData[safe: clickedRow], !item.isDirectory {
                let openItem = NSMenuItem(title: "Open", action: #selector(handleOpenFile(_:)), keyEquivalent: "")
                openItem.target = self
                // Set attributed title to ensure text is visible when highlighted
                let attributedTitle = NSAttributedString(
                    string: "Open",
                    attributes: [.foregroundColor: NSColor.controlTextColor]
                )
                openItem.attributedTitle = attributedTitle
                menu.addItem(openItem)
            }
            menu.addItem(sendToItem)
        } else {
            // Multiple selection: Only show Send to
            menu.addItem(sendToItem)
        }
    }
    
    // MARK: - Table Sorting Delegate
    func tableView(_ tableView: NSTableView, sortDescriptorsDidChange oldDescriptors: [NSSortDescriptor]) {
        guard getTableType(for: tableView) == .right else { return }
        
        guard let sortDescriptor = tableView.sortDescriptors.first else { return }
        
        let columnId = sortDescriptor.key ?? ""
        var sortColumn: SortColumn?
        
        switch columnId {
        case "name": sortColumn = .name
        case "fileSize": sortColumn = .size
        case "fileType": sortColumn = .type
        case "modificationDate": sortColumn = .date
        default: return
        }
        
        guard let column = sortColumn else { return }
        
        // Update sort state
        if currentSortColumn == column {
            currentSortOrder = sortDescriptor.ascending ? .ascending : .descending
        } else {
            currentSortColumn = column
            currentSortOrder = sortDescriptor.ascending ? .ascending : .descending
        }
        
        // Apply sorting
        sortRightTableData()
        tableView.reloadData()
    }
    
    // Right table drag handling logic (avoid manual pasteboard writing causing count mismatch)
    func tableView(_ tableView: NSTableView, writeRowsWith rowIndexes: IndexSet, to pboard: NSPasteboard) -> Bool {
        // Only handle right table drag
        guard getTableType(for: tableView) == .right else {
            return false
        }
        guard isSupportFileDrag else { return false }
        // Only allow drag to start when user drags from "selected rows"
        // Otherwise allow user to multi-select by holding mouse and moving
        let currentSelection = tableView.selectedRowIndexes
        let canBeginDrag = !currentSelection.isEmpty && rowIndexes.isSubset(of: currentSelection)
        return canBeginDrag
    }

    // When using modern API, logs should be placed in drag session start callback
    func tableView(_ tableView: NSTableView, draggingSession session: NSDraggingSession, willBeginAt screenPoint: NSPoint, forRowIndexes rowIndexes: IndexSet) {
        guard getTableType(for: tableView) == .right else { return }
        guard isSupportFileDrag else { return }
        logger.info("Right table drag started with \(rowIndexes.count) items")
        
        for row in rowIndexes.sorted() {
            guard let draggedItem = rightTableData[safe: row] else {
                logger.warn("Right table row \(row) data missing during drag")
                continue
            }
            logger.debug("Dragging row \(row + 1): \(draggedItem.name) (\(draggedItem.isDirectory ? "Folder" : "File"))")
        }

        let selectedRowIndexes = rowIndexes

        let selectedItems = selectedRowIndexes.compactMap { index -> FileInfo? in
            guard index >= 0, let item = rightTableData[safe: index] else { return nil }
            return item
        }

        let pathList = selectedItems.map { $0.path }

        // Build JSON data
        let fileDragData: [String: Any] = [
            "PathList": pathList
        ]

        // Convert to JSON string
        do {
            let jsonData = try JSONSerialization.data(withJSONObject: fileDragData, options: [])
            guard let jsonString = String(data: jsonData, encoding: .utf8) else {
                logger.error("Failed to convert JSON data to string for drag operation")
                return
            }

            guard let screen = NSScreen.main else {
                logger.error("Failed to get NSScreen main for drag operation")
                return
            }

            let width = UInt16(truncatingIfNeeded: Int(screen.frame.width.rounded()))
            let height = UInt16(truncatingIfNeeded: Int(screen.frame.height.rounded()))
            let posX = Int16(truncatingIfNeeded: Int(screenPoint.x.rounded()))
            let posY = Int16(truncatingIfNeeded: height) - Int16(truncatingIfNeeded: Int(screenPoint.y.rounded())) // start with left-top(0,0)
            let timestamp = UInt64(Date().timeIntervalSince1970 * 1000)
            
            logger.debug("Drag file request - position: (\(posX), \(posY)), screen: \(width)x\(height)")
            
            HelperClient.shared.setDragFileListRequest(multiFilesData: jsonString,
                                                        timestamp: timestamp,
                                                        width: width,
                                                        height: height,
                                                        posX: posX,
                                                        posY: posY) { success, statusCode in
                if success {
                    logger.info("Drag file list request sent successfully, status: \(String(describing: statusCode))")
                } else {
                    logger.error("Failed to send drag file list request, status: \(String(describing: statusCode))")
                }
            }
        } catch {
            logger.error("Error creating JSON for drag operation: \(error)")
        }
    }

    // Drag operation type setting
    func draggingSession(_ session: NSDraggingSession, sourceOperationMaskFor context: NSDraggingContext) -> NSDragOperation {
        guard isSupportFileDrag else { return [] }
        return .copy // Drag operation is copy
    }
}

// MARK: - NSMenuDelegate for dynamic right-click menu updates
extension MainHomeViewController: NSMenuDelegate {
    
    func menuWillOpen(_ menu: NSMenu) {
        if(ThemeManager.shared.shouldShowMenuBorder){
            applyMenuWindowBorder()
        }else{
            applyMenuWindowBackground()
        }
    }
    
    // MARK: - Menu Window Styling
    
    /// Applies border styling to the menu window
    private func applyMenuWindowBorder() {
        let theme = ThemeManager.shared.currentTheme
        // Wait for menu window to be created
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.05) {
            // Find the menu window
            for window in NSApp.windows {
                // Check if this is a menu window (class name contains "NSCarbonMenuWindow" or similar)
                let className = NSStringFromClass(type(of: window))
                if className.contains("MenuWindow") {
                    // Add border to window's content view
                    if let contentView = window.contentView {
                        contentView.wantsLayer = true
                        contentView.layer?.borderColor = theme.menuBorderColor.cgColor
                        contentView.layer?.borderWidth = theme.menuBorderWidth
                        contentView.layer?.cornerRadius = 8.0
                        
                        // Ensure border is visible by setting masksToBounds
                        contentView.layer?.masksToBounds = false
                        
                        // Force redraw
                        contentView.needsDisplay = true
                    }
                }
            }
        }
    }
    
    /// Applies background color to the menu window
    private func applyMenuWindowBackground() {
        // Set background color multiple times with delays to ensure all views are covered
        for delay in [0.01, 0.05, 0.1] {
            DispatchQueue.main.asyncAfter(deadline: .now() + delay) {
                // Find the menu window
                for window in NSApp.windows {
                    let className = NSStringFromClass(type(of: window))
                    if className.contains("MenuWindow") {
                        // Set window background color to white
                        window.backgroundColor = NSColor.white
                        window.isOpaque = true
                        
                        // Only set background for top-level views, not deep into menu items
                        // This allows system to handle menu item highlighting properly
                        if let contentView = window.contentView {
                            contentView.wantsLayer = true
                            // Only set if not already a highlight color
                            if !self.isHighlightColor(contentView.layer?.backgroundColor) {
                                contentView.layer?.backgroundColor = NSColor.white.cgColor
                            }
                            
                            // Set background for direct children, but skip menu item cells
                            for subview in contentView.subviews {
                                let subviewClassName = NSStringFromClass(type(of: subview))
                                // Skip menu item cells and highlight views completely
                                if !subviewClassName.contains("MenuItemCell") &&
                                   !subviewClassName.contains("Highlight") &&
                                   !subviewClassName.contains("Selection") {
                                    subview.wantsLayer = true
                                    if !self.isHighlightColor(subview.layer?.backgroundColor) {
                                        subview.layer?.backgroundColor = NSColor.white.cgColor
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    }
    
    // Helper method to check if a color is a highlight color (blue-ish)
    private func isHighlightColor(_ cgColor: CGColor?) -> Bool {
        guard let cgColor = cgColor else { return false }
        guard let color = NSColor(cgColor: cgColor) else { return false }
        guard let rgb = color.usingColorSpace(.deviceRGB) else { return false }
        
        // Check if it's a blue-ish color (typical highlight color)
        // Highlight colors usually have high blue component and lower red/green
        let red = rgb.redComponent
        let green = rgb.greenComponent
        let blue = rgb.blueComponent
        
        // Typical highlight blue: blue > 0.4 and blue > red and blue > green
        if blue > 0.4 && blue > red && blue > green {
            return true
        }
        
        // Also check for system accent color variations
        if blue > 0.3 && (red + green) < blue {
            return true
        }
        
        return false
    }

}

// MARK: - File System Monitoring
extension MainHomeViewController {
    
    /// Setup file system monitoring for the specified directory.
    /// If `path` is nil, monitor current displayed directory first, then fallback to download path.
    private func setupFileSystemMonitoring(for path: String? = nil) {
        let fallbackDownloadPath = CSUserPreferences.shared.getDownloadPathOrDefault()
        let pathToWatch = path ?? (currentRightPath.isEmpty ? fallbackDownloadPath : currentRightPath)
        
        guard FileManager.default.fileExists(atPath: pathToWatch) else {
            logger.warn("File system monitoring path does not exist: \(pathToWatch)")
            return
        }
        
        logger.info("Setting up file system monitoring for: \(pathToWatch)")
        
        // Create FSEvents callback context
        var context = FSEventStreamContext(
            version: 0,
            info: Unmanaged.passUnretained(self).toOpaque(),
            retain: nil,
            release: nil,
            copyDescription: nil
        )
        
        // Paths to watch
        let pathsToWatch = [pathToWatch] as CFArray
        
        // FSEvents callback function
        let callback: FSEventStreamCallback = { (
            streamRef: ConstFSEventStreamRef,
            clientCallBackInfo: UnsafeMutableRawPointer?,
            numEvents: Int,
            eventPaths: UnsafeMutableRawPointer,
            eventFlags: UnsafePointer<FSEventStreamEventFlags>,
            eventIds: UnsafePointer<FSEventStreamEventId>
        ) in
            // Get MainHomeViewController instance
            guard let clientCallBackInfo = clientCallBackInfo else { return }
            let viewController = Unmanaged<MainHomeViewController>.fromOpaque(clientCallBackInfo).takeUnretainedValue()
            
            // Convert paths to Swift array
            let paths = Unmanaged<CFArray>.fromOpaque(eventPaths).takeUnretainedValue() as! [String]
            
            // Handle each event
            for i in 0..<numEvents {
                let path = paths[i]
                let flags = eventFlags[i]
                
                // Check if it's a file creation, modification, rename or delete event
                let isCreated = (flags & UInt32(kFSEventStreamEventFlagItemCreated)) != 0
                let isModified = (flags & UInt32(kFSEventStreamEventFlagItemModified)) != 0
                let isRenamed = (flags & UInt32(kFSEventStreamEventFlagItemRenamed)) != 0
                let isRemoved = (flags & UInt32(kFSEventStreamEventFlagItemRemoved)) != 0
                
                if isCreated || isModified || isRenamed || isRemoved {
                    viewController.handleFileSystemChange(at: path)
                }
            }
        }
        
        // Create FSEventStream
        fileSystemEventStream = FSEventStreamCreate(
            kCFAllocatorDefault,
            callback,
            &context,
            pathsToWatch,
            FSEventStreamEventId(kFSEventStreamEventIdSinceNow),
            0.5, // Delay time in seconds
            UInt32(kFSEventStreamCreateFlagFileEvents | kFSEventStreamCreateFlagUseCFTypes)
        )
        
        guard let stream = fileSystemEventStream else {
            logger.error("Failed to create FSEventStream")
            return
        }
        
        // Add stream to run loop
        FSEventStreamScheduleWithRunLoop(stream, CFRunLoopGetMain(), CFRunLoopMode.defaultMode.rawValue)
        
        // Start stream
        if FSEventStreamStart(stream) {
            logger.info("File system monitoring started")
        } else {
            logger.error("Failed to start FSEventStream")
        }
    }
    
    /// Restart file system monitoring for a new directory.
    private func restartFileSystemMonitoring(for path: String) {
        stopFileSystemMonitoring()
        setupFileSystemMonitoring(for: path)
    }
    
    /// Stop file system monitoring
    private func stopFileSystemMonitoring() {
        guard let stream = fileSystemEventStream else { return }
        
        FSEventStreamStop(stream)
        FSEventStreamInvalidate(stream)
        FSEventStreamRelease(stream)
        fileSystemEventStream = nil
        
        logger.info("File system monitoring stopped")
    }
    
    /// Handle file system change events
    private func handleFileSystemChange(at path: String) {
        // Debounce: ignore if too soon after last refresh
        let now = Date()
        let timeSinceLastRefresh = now.timeIntervalSince(lastRefreshTime)
        if timeSinceLastRefresh < refreshDebounceInterval {
            logger.debug("FSEvents: Skipping refresh due to debounce (last refresh \(String(format: "%.2f", timeSinceLastRefresh))s ago)")
            return
        }
        
        // Check if changed path is in current directory
        guard isPathInCurrentDirectory(path) else {
            return
        }
        
        logger.info("FSEvents: File system change detected, refreshing FileBrowser")
        
        // Refresh FileBrowser on main thread
        DispatchQueue.main.async { [weak self] in
            guard let self = self else { return }
            
            // Remember current selection
            let selectedRows = self.rightTableView.selectedRowIndexes
            let selectedPaths = selectedRows.compactMap { index -> String? in
                guard let item = self.rightTableData[safe: index] else { return nil }
                return item.path
            }
            
            // Refresh current directory
            self.refreshCurrentDirectory()
            
            // Restore selection if files still exist
            self.restoreSelection(paths: selectedPaths)
            
            // Update last refresh time
            self.lastRefreshTime = now
        }
    }
    
    /// Check if given path is directly in current directory (not in subdirectory)
    private func isPathInCurrentDirectory(_ changedPath: String) -> Bool {
        // Don't refresh if current path is empty
        guard !currentRightPath.isEmpty else { return false }
        
        // Get parent directory of changed path
        let changedParentPath = (changedPath as NSString).deletingLastPathComponent
        
        // Only refresh if file is directly in current directory (strict equality, not subdirectory)
        return changedParentPath == currentRightPath
    }
    
    /// Refresh current directory
    private func refreshCurrentDirectory() {
        guard !currentRightPath.isEmpty else { return }
        
        rightTableData.removeAll()
        
        if let folderContents = FileSystemInfoFetcher.getFolderContentsInfo(in: currentRightPath) {
            for item in folderContents {
                rightTableData.append(item)
            }
        } else {
            logger.error("Unable to refresh folder contents: \(currentRightPath)")
        }
        
        // Apply current sorting
        sortRightTableData()
        rightTableView.reloadData()
    }
    
    /// Restore previous selection
    private func restoreSelection(paths: [String]) {
        guard !paths.isEmpty else { return }
        
        var indexesToSelect = IndexSet()
        
        for (index, item) in rightTableData.enumerated() {
            if paths.contains(item.path) {
                indexesToSelect.insert(index)
            }
        }
        
        if !indexesToSelect.isEmpty {
            rightTableView.selectRowIndexes(indexesToSelect, byExtendingSelection: false)
        }
    }
    
    /// Check if FileBrowser needs refresh after file transfer completion
    private func refreshFileBrowserIfNeeded(for fileInfo: CSFileInfo) {
        // Get download path
        let downloadPath = CSUserPreferences.shared.getDownloadPathOrDefault()
        
        // currentFileName might be:
        // 1. Just filename: "file.txt"
        // 2. Relative path: "subfolder/file.txt" or "Users/ts/Desktop/396/88/file.txt"
        // 3. Absolute path: "/Users/ts/Desktop/396/88/file.txt"
        let currentFileName = fileInfo.session.currentFileName
        
        // Build full path of received file
        let receivedFilePath: String
        if currentFileName.hasPrefix("/") {
            // Already an absolute path
            receivedFilePath = currentFileName
        } else if currentFileName.hasPrefix("Users/") || currentFileName.hasPrefix("home/") {
            // Relative path from root (without leading slash)
            receivedFilePath = "/" + currentFileName
        } else {
            // Just a filename or relative path, append to download path
            receivedFilePath = (downloadPath as NSString).appendingPathComponent(currentFileName)
        }
        
        // Get directory of received file
        let receivedFileDirectory = (receivedFilePath as NSString).deletingLastPathComponent
        
        // Debug logging
        /*
        logger.debug("Transfer completed - checking refresh need")
        logger.debug("  Current file name: \(currentFileName)")
        logger.debug("  Constructed file path: \(receivedFilePath)")
        logger.debug("  Received file directory: \(receivedFileDirectory)")
        logger.debug("  Current displayed directory: \(currentRightPath)")
        logger.debug("  match: \(receivedFileDirectory == currentRightPath)")
         */

        // Key check: received file directory must strictly equal current displayed directory
        // - If in subdirectory: don't refresh (will reload when user clicks into subdirectory)
        // - If in parent directory: don't refresh (will reload when user returns to parent)
        // - Only in current directory: refresh
        guard receivedFileDirectory == currentRightPath else {
            logger.debug("Directory mismatch, skipping refresh")
            return
        }
        
        // Debounce check
        let now = Date()
        let timeSinceLastRefresh = now.timeIntervalSince(lastRefreshTime)
        if timeSinceLastRefresh < refreshDebounceInterval {
            logger.debug("Skipping refresh due to debounce (last refresh \(String(format: "%.2f", timeSinceLastRefresh))s ago)")
            return
        }
        
        logger.info("Refreshing FileBrowser after transfer completion")
        
        // Remember current selection
        let selectedRows = rightTableView.selectedRowIndexes
        let selectedPaths = selectedRows.compactMap { index -> String? in
            guard let item = rightTableData[safe: index] else { return nil }
            return item.path
        }
        
        // Refresh current directory
        refreshCurrentDirectory()
        
        // Restore selection
        restoreSelection(paths: selectedPaths)
        
        // Update last refresh time
        lastRefreshTime = now
    }
}
