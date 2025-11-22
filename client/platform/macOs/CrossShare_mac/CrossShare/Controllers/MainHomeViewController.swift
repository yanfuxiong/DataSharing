import Cocoa
import SnapKit

// Enum defining table types for clear distinction
enum TableType {
    case left
    case right
    case bottom
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
    // Clear button for bottom table
    private let clearButton = NSButton()

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

    // Right side path and breadcrumb
    private var currentRightPath: String = ""
    private let rightHeaderContainer = NSView()
    private let backButton = NSButton(title: "Back", target: nil, action: nil)
    private let breadcrumbStack = NSStackView()

    lazy var topView: HomeHeaderView = {
        let cview = HomeHeaderView(frame: .zero)
        cview.wantsLayer = true
        cview.layer?.backgroundColor = NSColor(hex: 0x377AF6).cgColor
        return cview
    }()
    
    // About view
    private var aboutView: CSAboutView?
    private var clickMonitor: Any?
    
    // License view
    private var licenseView: CSLicenseView?
    
    deinit {
        // Remove notification observers
        NotificationCenter.default.removeObserver(self)
        // Stop transfer monitoring
        DataTransmissionManager.shared.stopListening()
        // Remove click monitor
        removeClickMonitor()
    }

    override func viewDidLoad() {
        super.viewDidLoad()
        setupUI()
        RealmDataManager.shared.setupRealm()
        setupShowData()
        setupNotifications()
        setupDataBlock()
    }

    private func setupNotifications() {
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
            let menuPoint = NSPoint(x: self.view.frame.width - 225, y: self.topView.frame.minY - 10)
            self.showDeviceListMenu(at: menuPoint, in: self.view)
        }
    }
    
    // MARK: - Menu Actions
    private func showOptionsMenu() {
        // Create a simple pop-up menu
        let menu = NSMenu(title: "Options")
        
        // Add settings option with submenu
        let settingMenuItem = NSMenuItem(title: "setting", action: nil, keyEquivalent: "")
        if let image = NSImage(named: "setting") {
            image.size = NSSize(width: 16, height: 16)
            settingMenuItem.image = image
        }
        
        // Create settings submenu
        let settingSubmenu = NSMenu()
        
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
    
    private func showDeviceListMenu(at point: NSPoint? = nil, in view: NSView? = nil) {
        // Create device list menu
        let menu = NSMenu(title: "Device list")
        
        // Check if there are online devices
        if self.deviceList.isEmpty {
            // Create custom "No online devices" menu item
            let noDeviceItem = CustomMenuItem.createNoDeviceMenuItem(CSDeviceManager.shared.diasStatus)
            menu.addItem(noDeviceItem)
        } else {
            // Iterate through device list and add device menu items
            for (index, device) in self.deviceList.enumerated() {
                // Add separator before non-first device items
                if index > 0 {
                    menu.addItem(NSMenuItem.separator())
                }
                
                // Create custom device menu item (icon on left, text on right)
                let imageName = "Device_\(device.sourcePortType)"
                let deviceItem = CustomMenuItem.createDeviceMenuItem(title: device.deviceName, 
                                                                    imageName: imageName, 
                                                                    target: self, 
                                                                    tag: index + 1, 
                                                                    ipAddr: device.ipAddr)
                menu.addItem(deviceItem)
            }
        }
        
        // Show menu at specified position or default button position
        let targetView = view ?? self.topView.deviceBtn
        let targetPoint = point ?? NSPoint(x: 0, y: self.topView.deviceBtn.frame.height)
        menu.popUp(positioning: nil, at: targetPoint, in: targetView)
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
        newAboutView.wantsLayer = true
        newAboutView.layer?.backgroundColor = NSColor.controlBackgroundColor.cgColor
        newAboutView.layer?.cornerRadius = 10
        newAboutView.layer?.shadowColor = NSColor.black.cgColor
        newAboutView.layer?.shadowOpacity = 0.3
        newAboutView.layer?.shadowOffset = NSSize(width: 0, height: -2)
        newAboutView.layer?.shadowRadius = 8
        newAboutView.layer?.borderWidth = 1
        newAboutView.layer?.borderColor = NSColor.separatorColor.cgColor
        
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
        newLicenseView.wantsLayer = true
        newLicenseView.layer?.backgroundColor = NSColor.controlBackgroundColor.cgColor
        newLicenseView.layer?.cornerRadius = 10
        newLicenseView.layer?.shadowColor = NSColor.black.cgColor
        newLicenseView.layer?.shadowOpacity = 0.3
        newLicenseView.layer?.shadowOffset = NSSize(width: 0, height: -2)
        newLicenseView.layer?.shadowRadius = 8
        newLicenseView.layer?.borderWidth = 1
        newLicenseView.layer?.borderColor = NSColor.separatorColor.cgColor
        
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
        view.layer?.backgroundColor = NSColor(white: 0.95, alpha: 1.0).cgColor
        
        // Create overall blue dashed border view
        let combinedBorderView = NSView()
        combinedBorderView.wantsLayer = true
        combinedBorderView.layer?.backgroundColor = NSColor.clear.cgColor
        
        
        // Add blue dashed border to the view
        addBlueDashedBorder(to: combinedBorderView)
        
        // Configure title labels
        // Remove left table title
        // Right side uses custom header container (Back + breadcrumb)
        setupTitleLabel(bottomTitleLabel, text: "Files records")
        
        // Setup clear button for bottom table
        clearButton.bezelStyle = .texturedRounded
        clearButton.setButtonType(.momentaryPushIn)
        clearButton.isBordered = false
        clearButton.title = ""
    
        if let clearImage = NSImage(named: "clearIcon") {
            // Adjust image size to match back button
            clearImage.size = NSSize(width: 20, height: 20)
            clearButton.image = clearImage
            clearButton.imageScaling = .scaleProportionallyDown
        } else {
            logger.warn("clearIcon image not found")
        }
        clearButton.target = self
        clearButton.action = #selector(handleClearButtonTap)
        
        // Configure table views
        setupTableView(leftTableView, scrollView: leftScrollView, type: .left)
        setupTableView(rightTableView, scrollView: rightScrollView, type: .right)
        setupTableView(bottomTableView, scrollView: bottomScrollView, type: .bottom)
        
        leftTableView.allowsMultipleSelection = false
        rightTableView.allowsMultipleSelection = true
        bottomTableView.allowsMultipleSelection = false
        // Ensure bottom table row height accommodates title + subtitle
        bottomTableView.rowHeight = 50
        
        // Add views to main container
        view.addSubview(topView)
        view.addSubview(combinedBorderView) // Add overall blue dashed border
        // Remove left title label from subviews
        view.addSubview(rightHeaderContainer)
        view.addSubview(bottomTitleLabel)
        view.addSubview(clearButton)
        view.addSubview(leftScrollView)
        view.addSubview(rightScrollView)
        view.addSubview(bottomScrollView)
        
        topView.snp.makeConstraints { make in
            make.left.right.top.equalToSuperview()
            make.height.equalTo(100)
        }
        
        // Set constraints for overall dashed border
        combinedBorderView.snp.makeConstraints { make in
            make.top.equalTo(topView.snp.bottom).offset(16)
            make.leading.equalTo(view.snp.leading).offset(16)
            make.trailing.equalTo(view.snp.trailing).offset(-16)
            make.bottom.equalTo(bottomScrollView.snp.top).offset(-36)
        }
        
        leftScrollView.snp.makeConstraints { make in
            // Align left scroll view top with right header container
            make.top.equalTo(topView.snp.bottom).offset(20)
            make.leading.equalTo(view.snp.leading).offset(20)
            make.width.equalTo(200)
            make.bottom.equalTo(bottomScrollView.snp.top).offset(-40)
        }
        
        rightHeaderContainer.snp.makeConstraints { make in
            make.top.equalTo(topView.snp.bottom).offset(20)
            make.leading.equalTo(leftScrollView.snp.trailing).offset(20)
            make.trailing.equalTo(view.snp.trailing).offset(-20)
            make.height.equalTo(28)
        }
        
        rightScrollView.snp.makeConstraints { make in
            make.top.equalTo(rightHeaderContainer.snp.bottom).offset(10)
            make.leading.equalTo(leftScrollView.snp.trailing).offset(20)
            make.trailing.equalTo(view.snp.trailing).offset(-20)
            make.bottom.equalTo(bottomScrollView.snp.top).offset(-40)
        }
        
        bottomTitleLabel.snp.makeConstraints { make in
            make.leading.equalTo(view.snp.leading).offset(20)
            make.bottom.equalTo(bottomScrollView.snp.top).offset(-10)
        }
        
        clearButton.snp.makeConstraints { make in
            make.leading.equalTo(bottomTitleLabel.snp.trailing).offset(10)
            make.bottom.equalTo(bottomScrollView.snp.top).offset(-10)
            make.width.equalTo(30)
            make.height.equalTo(20)
        }
        
        bottomScrollView.snp.makeConstraints { make in
            make.leading.equalTo(view.snp.leading).offset(20)
            make.trailing.equalTo(view.snp.trailing).offset(-20)
            make.bottom.equalTo(view.snp.bottom).offset(-20)
            make.height.equalTo(200)
        }
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

    
    // Title label configuration (unchanged)
    private func setupTitleLabel(_ label: NSTextField, text: String) {
        label.stringValue = text
        label.isEditable = false
        label.isBordered = false
        label.backgroundColor = .clear
        label.textColor = .darkGray
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
            let dragTypes: [NSPasteboard.PasteboardType] = [.fileURL, .string]
            tableView.registerForDraggedTypes(dragTypes)
            tableView.setDraggingSourceOperationMask(.copy, forLocal: true)
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
    
    // Create right-click context menu for files
    private func createRightClickMenu() -> NSMenu {
        let menu = NSMenu()
        // Dynamic menu: Update items before popup through delegate
        menu.delegate = self
        return menu
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
            HelperClient.shared.sendMultiFilesDropRequest(multiFilesData: jsonString) { success, statusCode in
                if success {
                    logger.info("Successfully sent file request to \(device.deviceName), status: \(String(describing: statusCode))")
                } else {
                    logger.error("Failed to send file request to \(device.deviceName), status: \(String(describing: statusCode))")
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
        
        backButton.snp.makeConstraints { make in
            make.leading.equalTo(rightHeaderContainer.snp.leading)
            make.centerY.equalTo(rightHeaderContainer.snp.centerY)
        }
        breadcrumbStack.snp.makeConstraints { make in
            make.leading.equalTo(backButton.snp.trailing).offset(8)
            make.trailing.lessThanOrEqualTo(rightHeaderContainer.snp.trailing)
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
    // Provide independent pasteboard writer for each row, ensuring drag items match pasteboard items one-to-one
    func tableView(_ tableView: NSTableView, pasteboardWriterForRow row: Int) -> NSPasteboardWriting? {
        guard getTableType(for: tableView) == .right else { return nil }
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
        // Only allow drag to start when user drags from "selected rows"
        // Otherwise allow user to multi-select by holding mouse and moving
        let currentSelection = tableView.selectedRowIndexes
        let canBeginDrag = !currentSelection.isEmpty && rowIndexes.isSubset(of: currentSelection)
        return canBeginDrag
    }

    // When using modern API, logs should be placed in drag session start callback
    func tableView(_ tableView: NSTableView, draggingSession session: NSDraggingSession, willBeginAt screenPoint: NSPoint, forRowIndexes rowIndexes: IndexSet) {
        guard getTableType(for: tableView) == .right else { return }
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
        return .copy // Drag operation is copy
    }
}

// MARK: - NSMenuDelegate for dynamic right-click menu updates
extension MainHomeViewController: NSMenuDelegate {
    func menuNeedsUpdate(_ menu: NSMenu) {
        guard menu === rightTableView.menu else { return }
        menu.removeAllItems()

        // Calculate current selection and clicked row
        let selectedIndexes = rightTableView.selectedRowIndexes
        let clickedRow = rightTableView.clickedRow
        let selectionCount = selectedIndexes.count

        // Build Send to submenu
        let sendToItem = NSMenuItem(title: "Send to", action: nil, keyEquivalent: "")
        let sendToSubmenu = NSMenu()
        
        // Create device names array
//        let deviceNames = ["Device 1", "Device 2", "Device 3"]
        
        // Iterate through device names using index to set tag
        for (index, device) in self.deviceList.enumerated() {
            // Add separator before non-first device items
            if index > 0 {
                sendToSubmenu.addItem(NSMenuItem.separator())
            }
            
            let deviceItem = NSMenuItem(title: device.deviceName, action: #selector(handleSendToDevice(_:)), keyEquivalent: "")
            deviceItem.target = self
            deviceItem.tag = index + 1
            
            // Add DeviceComputer icon
            let imageName = "Device_\(device.sourcePortType)"
            if let deviceIcon = NSImage(named: imageName) {
                deviceIcon.size = NSSize(width: 16, height: 16) // Set icon size
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
                menu.addItem(openItem)
            }
            menu.addItem(sendToItem)
        } else {
            // Multiple selection: Only show Send to
            menu.addItem(sendToItem)
        }
    }
}
