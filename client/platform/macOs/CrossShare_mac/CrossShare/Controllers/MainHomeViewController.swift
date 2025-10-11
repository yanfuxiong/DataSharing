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
    private var deviceList: [CrossShareDevice] = [] {
        didSet {
            topView.refreshDeviceList(deviceList)
        }
    }

    // Bottom table data structure is now CSFileInfo
    private var bottomTableData: [CSFileInfo] = []

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
    
    deinit {
        // Remove notification observers
        NotificationCenter.default.removeObserver(self)
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
        
        // Add observer for file transfer session start notification
        NotificationCenter.default.addObserver(
            forName: .fileTransferSessionStarted, object: nil, queue: .main
        ) { ntf in
            guard let userInfo = ntf.userInfo as? [String: Any] else {
                print("Failed to parse the file transfer session data.")
                return
            }
            print("File transfer start:---\(userInfo)----")
            // Try to convert userInfo to CSFileInfo model
            if var csFileInfo = self.createCSFileInfo(from: userInfo) {
                // Directly find matching device and update device name through createFileTransferSession method
                if let sessionDict = userInfo["session"] as? [String: Any], let newFileTransfer = self.createFileTransferSession(from: sessionDict) {
                    csFileInfo = CSFileInfo(
                        session: newFileTransfer,
                        sessionId: csFileInfo.sessionId,
                        senderID: csFileInfo.senderID,
                        isCompleted: csFileInfo.isCompleted,
                        progress: csFileInfo.progress
                    )
                }
                
                // Save to Realm database
                RealmDataManager.shared.saveCSFileInfoToRealm(csFileInfo)
                self.readRelamData()
            }
        }
        
        
        NotificationCenter.default.addObserver(
            forName: .fileTransferSessionUpdated, object: nil, queue: .main
        ) { ntf in
            guard let userInfo = ntf.userInfo as? [String: Any] else {
                print("Failed to parse the file transfer session data.")
                return
            }
            // Try to convert userInfo to CSFileInfo model
            if let csFileInfo = self.createCSFileInfo(from: userInfo) {
                print("File transfer update-sessionId: \(csFileInfo.sessionId) currentFileName: \(csFileInfo.session.currentFileName) isCompleted: \(csFileInfo.isCompleted) progress: \(csFileInfo.progress)")

                self.updateAndRefreshCSFileInfo(csFileInfo)
            }
        }
        
        NotificationCenter.default.addObserver(
            forName: .fileTransferSessionCompleted, object: nil, queue: .main
        ) { ntf in
            guard let userInfo = ntf.userInfo as? [String: Any] else {
                print("Failed to parse the file transfer session data.")
                return
            }
            // Try to convert userInfo to CSFileInfo model
            if let csFileInfo = self.createCSFileInfo(from: userInfo) {
                print("File transfer completed-sessionId: \(csFileInfo.sessionId) currentFileName: \(csFileInfo.session.currentFileName) isCompleted: \(csFileInfo.isCompleted) progress: \(csFileInfo.progress)")
                self.updateAndRefreshCSFileInfo(csFileInfo)
            }
        }
    }
    
    
    // Update and refresh CSFileInfo data
    private func updateAndRefreshCSFileInfo(_ csFileInfo: CSFileInfo) {
        // Find and update the corresponding CSFileInfo based on sessionId
        if let index = self.bottomTableData.firstIndex(where: { $0.sessionId == csFileInfo.sessionId }) {
            // Update existing record
            self.bottomTableData[index] = csFileInfo
            
            // Only refresh relevant cells
            DispatchQueue.main.async {
                // Refresh specific row
                let rowIndexSet = IndexSet(integer: index)
                // Refresh all columns
                let columnIndexSet = IndexSet(integersIn: 0..<self.bottomTableView.numberOfColumns)
                self.bottomTableView.reloadData(forRowIndexes: rowIndexSet, columnIndexes: columnIndexSet)
            }
        } else {
            // Add if it's a new record
            self.bottomTableData.append(csFileInfo)
            
            // Need to refresh the entire table when adding a new record
            DispatchQueue.main.async {
                self.bottomTableView.reloadData()
            }
        }
        
        // Save to Realm database
        RealmDataManager.shared.saveCSFileInfoToRealm(csFileInfo)
        
    }

    private func setupDataBlock() {
        HelperClient.shared.getDeviceListFromHelper { deviceDicList in
            DispatchQueue.main.async {
                self.deviceList = deviceDicList.compactMap {
                  CrossShareDevice(from: $0) }
                self.topView.refreshDeviceList(self.deviceList)
            }
        }
        
        topView.tapMoreBtnBlock = { [weak self] in
            guard let self = self else { return }
            
            // Create a simple pop-up menu
            let menu = NSMenu(title: "Options")
            
            // Add settings option
            let languageMenuItem = NSMenuItem(title: "setting", action: #selector(self.gotoSetting), keyEquivalent: "")
            languageMenuItem.target = self
            menu.addItem(languageMenuItem)
            
            // Add separator
            menu.addItem(NSMenuItem.separator())
            
            // Add refresh device list option
            let refreshDevicesMenuItem = NSMenuItem(title: "refresh list", action: #selector(self.refreshDeviceList), keyEquivalent: "")
            refreshDevicesMenuItem.target = self
            menu.addItem(refreshDevicesMenuItem)
            
            // Show menu at button position
            menu.popUp(positioning: nil, at: NSPoint(x: 0, y: self.topView.moreBtn.frame.height), in: self.topView.moreBtn)
        }
        
        // Implement device list button click event
        topView.tapDeviceListBtnBlock = { [weak self] in
            guard let self = self else { return }
            
            // Create device list menu
            let menu = NSMenu(title: "Device list")
            
            // Check if there are online devices
            if self.deviceList.isEmpty {
                // Create custom "No online devices" menu item
                let noDeviceItem = CustomMenuItem.createNoDeviceMenuItem("No online devices")
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
            
            // Show menu at button position
            menu.popUp(positioning: nil, at: NSPoint(x: 0, y: self.topView.deviceBtn.frame.height), in: self.topView.deviceBtn)
        }
    }
    
    // MARK: - Menu Actions
    @objc private func gotoSetting() {
        // Language switching logic can be implemented here
        print("go to setting")
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
            print("Failed to retrieve downloads directory path")
            return
        }
        let downloadsPath = downloadsURL.path
        if let fileInfo = FileSystemInfoFetcher.getItemInfo(for: downloadsPath) {
            print("Left file info:")
            print("Name: \(fileInfo.name) | Path: \(fileInfo.path)")
            leftTableData.append(fileInfo)
        }

        // Dynamically get current user's Documents directory path
        guard let documentsURL = FileManager.default.urls(for: .documentDirectory, in: .userDomainMask).first else {
            print("Failed to retrieve Documents directory path")
            return
        }
        let targetPath = documentsURL.path
        if let fileInfo = FileSystemInfoFetcher.getItemInfo(for: targetPath) {
            print("Left file info:")
            print("Name: \(fileInfo.name) | Path: \(fileInfo.path)")
            leftTableData.append(fileInfo)
        }
        
        // Dynamically get current user's Desktop directory path
        guard let desktopURL = FileManager.default.urls(for: .desktopDirectory, in: .userDomainMask).first else {
            print("Failed to retrieve Desktop directory path")
            return
        }
        let DesktopPath = desktopURL.path
        if let fileInfo = FileSystemInfoFetcher.getItemInfo(for: DesktopPath) {
            print("Left file info:")
//            print("Name: \(fileInfo.name) | Path: \(fileInfo.path)")
            leftTableData.append(fileInfo)
        }
        
        // Initialize right table data (use first path in left table if available)
        guard let firstPath = leftTableData.first?.path else {
            print("No valid path in left table data to initialize right table")
            return
        }
        initRightTableViewData(firstPath)
    }
    
    // Right table data loading: directly add FileInfo to data source
    private func initRightTableViewData(_ targetPath: String) {
        rightTableData.removeAll() // Clear old data
        currentRightPath = targetPath
        
        if let folderContents = FileSystemInfoFetcher.getFolderContentsInfo(in: targetPath) {
            print("\nRight folder contents:")
            print("----------------------------------------")
            
            for item in folderContents {
                rightTableData.append(item)
            }
        } else {
            print("Unable to get folder contents: \(targetPath)")
        }
        rightTableView.reloadData()
        updateBreadcrumbs(for: targetPath)
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
            print("Warning: clearIcon image not found")
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
            make.height.equalTo(80)
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
            tableView.addTableColumn(nameColumn)
            
            // Size column
            let sizeColumn = NSTableColumn(identifier: NSUserInterfaceItemIdentifier("size"))
            sizeColumn.title = "Size"
            sizeColumn.width = 80
            tableView.addTableColumn(sizeColumn)
            
            // Type column
            let typeColumn = NSTableColumn(identifier: NSUserInterfaceItemIdentifier("type"))
            typeColumn.title = "Type"
            typeColumn.width = 120
            tableView.addTableColumn(typeColumn)

            
            // Date column
            let dateColumn = NSTableColumn(identifier: NSUserInterfaceItemIdentifier("date"))
            dateColumn.title = "Date Modified"
            dateColumn.width = 120
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
            print("No valid files selected")
            return
        }

        // Get device information
        guard sender.tag - 1 < self.deviceList.count, let device = self.deviceList[safe: sender.tag-1] else {
            print("Device index out of range")
            return
        }

        // Get device ID and IP address
        let deviceId = device.id
        guard let deviceIp = device.ipAddress else {
            print("Failed to get device IP address")
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
                print("Failed to convert JSON data to string")
                return
            }

            // Call send method
            HelperClient.shared.sendMultiFilesDropRequest(multiFilesData: jsonString) { success, statusCode in
                if success {
                    print("Successfully sent file request, status code: \(String(describing: statusCode))")
                } else {
                    print("Failed to send file request, status code: \(String(describing: statusCode))")
                }
            }
        } catch {
            print("Error creating JSON: \(error)")
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
            print("The data required to create CSFileInfo is missing.")
            return nil
        }
        
        // Create FileTransferSession instance
        if let fileTransferSession = createFileTransferSession(from: sessionDict) {
            return CSFileInfo(
                session: fileTransferSession,
                sessionId: sessionId,
                senderID: senderID,
                isCompleted: isCompletedInt != 0,
                progress: progress
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
                // Print senderID and ipAddr for debugging
//                print("Checking device - ID: \(device.id), IP: \(device.ipAddr) senderID:\(senderID) senderIP:\(senderIP)")
                // Check if senderID and ipAddr match
                if device.id == senderID && device.ipAddr == senderIP {
                    print("Found matching device: \(device.deviceName)")
                    
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
            print("No matching device found for senderID: \(senderID), senderIP: \(senderIP)")
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
            print("Tapped left table row \(row+1): \(item.name) (Path: \(item.path))")
            // Left table tap event handling logic...
            initRightTableViewData(item.path)
            
        case .right:
            guard let item = rightTableData[safe: row] else { return }
            print("Tapped right table row \(row+1): \(item.name) (Type: \(item.isDirectory ? "Folder" : "File"))")
            // Single click no longer enters directory, wait for double-click handling
            
        case .bottom:
            guard let fileInfo = bottomTableData[safe: row] else { return }
            print("Tapped bottom table row \(row+1): \(fileInfo.session.currentFileName) (SenderID: \(fileInfo.senderID), Progress: \(fileInfo.progress))")
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
            print("Warning: backArrows image not found")
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
        cell.textField = textField
        cell.addSubview(textField)
        
        // Cell padding (avoid text sticking to edge)
        textField.snp.makeConstraints { make in
            if tableType == .left || tableType == .right {
                // Add left padding for text field to make space for the icon
                make.leading.equalTo(cell.imageView?.snp.trailing ?? 0).offset(5)
                make.top.bottom.trailing.equalToSuperview().inset(5)
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
            print("No transfer record found for sessionId: \(sessionId)")
            return
        }
        
        let ipPort = fileInfo.session.senderIP
        let clientID = fileInfo.session.senderID
        
        // Parse timestamp from sessionId (format: "senderID-timestamp")
        guard let timeStamp = extractTimestampFromSessionId(sessionId) else {
            print("Unable to extract timestamp from sessionId: \(sessionId)")
            return
        }
        
        print("Canceling file transfer - IP:\(ipPort), ClientID:\(clientID), TimeStamp:\(timeStamp)")
        
        // Call XPC service to cancel transfer via HelperClient
        HelperClient.shared.setCancelFileTransfer(ipPort: ipPort, clientID: clientID, timeStamp: timeStamp) { [weak self] success, error in
            if success {
                print("File transfer canceled successfully")
                // Remove canceled transfer from list
                DispatchQueue.main.asyncAfter(deadline: .now() + 0.3) {
                    self?.bottomTableData.removeAll(where: { $0.sessionId == sessionId })
                    self?.bottomTableView.reloadData()
                }
            } else {
                print("Failed to cancel file transfer: \(error ?? "Unknown error")")
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
        print("\n📤 Right table drag started (\(rowIndexes.count) items):")
        for row in rowIndexes.sorted() {
            guard let draggedItem = rightTableData[safe: row] else {
                print("⚠️ Right table row \(row) data missing")
                continue
            }
            print("--- Row \(row + 1) ---")
            print("   Name: \(draggedItem.name)")
            print("   Type: \(draggedItem.isDirectory ? "Folder" : "File")")
            print("   Path: \(draggedItem.path)")
            print("   Size: \(draggedItem.isDirectory ? "---" : UtilsHelper.formatFileSize(draggedItem.fileSize ?? 0))")
            print("   Modified: \(draggedItem.modificationDate ?? "Unknown")")
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


