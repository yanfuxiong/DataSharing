import Cocoa
import SnapKit
import UniformTypeIdentifiers

// Custom cell class for bottom table view
class CSBottomTableViewCell: NSTableCellView {
    
    // Data structure for bottom table items
    struct Item {
        let name: String
        let date: String
        let isDirectory: Bool
    }
    
    // Stores the currently configured CSFileInfo object
    private var currentFileInfo: CSFileInfo?
    
    // Deletion closure for notifying controller to perform deletion
    var onDelete: ((String) -> Void)?
    
    // Cancellation closure for notifying controller to cancel transmission
    var onCancel: ((String) -> Void)?
    
    // UI elements
    private lazy var contentView: NSView = {
        let view = NSView()
        view.wantsLayer = true
        view.layer?.backgroundColor = NSColor.white.cgColor
        view.layer?.cornerRadius = 10.0
        return view
    }()
    
    // Transmission status label (currently hidden)
    private lazy var statusLabel: NSTextField = {
        let textField = NSTextField()
        // textField.stringValue = "Transmission completed"
        textField.isEditable = false
        textField.isBordered = false
        textField.backgroundColor = .clear
        textField.font = NSFont.systemFont(ofSize: 12)
        textField.textColor = NSColor.gray
        textField.isHidden = true
        return textField
    }()
    
    // 自定义进度条视图
    private lazy var customProgressView: CSCustomProgressView = {
        let view = CSCustomProgressView()
        view.isHidden = true
        return view
    }()
    
    // Receive button (currently hidden)
    private lazy var receiveBtn: NSButton = {
        let button = NSButton()
        button.bezelStyle = .texturedRounded
        button.setButtonType(.momentaryPushIn)
        button.isBordered = false
        if let image = NSImage(named: "receiveTickSucess") {
            image.size = NSSize(width: 16, height: 16)
            button.image = image
        }
        button.isHidden = true
        return button
    }()
    
    // Open button
    private lazy var openBtn: NSButton = {
        let button = NSButton()
        button.bezelStyle = .texturedRounded
        button.setButtonType(.momentaryPushIn)
        button.isBordered = false
        if let image = NSImage(named: "open") {
            image.size = NSSize(width: 16, height: 16)
            button.image = image
        }
        button.target = self
        button.action = #selector(handleOpenButtonClick)
        return button
    }()
    
    // Delete button
    private lazy var deleteBtn: NSButton = {
        let button = NSButton()
        button.bezelStyle = .texturedRounded
        button.setButtonType(.momentaryPushIn)
        button.isBordered = false
        if let image = NSImage(named: "delete") {
            image.size = NSSize(width: 16, height: 16)
            button.image = image
        }
        button.target = self
        button.action = #selector(handleDeleteButtonClick)
        return button
    }()
    
    private lazy var customImageView: NSImageView = {
        let imageView = NSImageView.init()
        imageView.imageScaling = .scaleProportionallyUpOrDown
        imageView.imageAlignment = .alignCenter
        return imageView
    }()
    
    // 右侧内容容器
    private lazy var rightContentView: NSView = {
        let view = NSView()
        return view
    }()
    
    private lazy var mainTextField: NSTextField = {
        let textField = NSTextField()
        textField.isEditable = false
        textField.isBordered = false
        textField.backgroundColor = .clear
        textField.font = NSFont.systemFont(ofSize: 14, weight: .medium)
        return textField
    }()
    
    // 文件名文本框（用于多文件传输）
    private lazy var fileNameTextField: NSTextField = {
        let textField = NSTextField()
        textField.isEditable = false
        textField.isBordered = false
        textField.backgroundColor = .clear
        textField.font = NSFont.systemFont(ofSize: 12)
        textField.textColor = NSColor.darkGray
        textField.lineBreakMode = .byTruncatingTail
        textField.setContentCompressionResistancePriority(.defaultLow, for: .horizontal)
        return textField
    }()
    
    private lazy var subTextField: NSTextField = {
        let textField = NSTextField()
        textField.isEditable = false
        textField.isBordered = false
        textField.backgroundColor = .clear
        textField.font = NSFont.systemFont(ofSize: 12)
        textField.textColor = NSColor.gray
        textField.lineBreakMode = .byTruncatingTail
        textField.setContentCompressionResistancePriority(.defaultLow, for: .horizontal)
        return textField
    }()
    
    // Identifier for reuse
    static let identifier = NSUserInterfaceItemIdentifier("BottomTableCell")
    
    override init(frame frameRect: NSRect) {
        super.init(frame: frameRect)
        setupUI()
    }
    
    required init?(coder: NSCoder) {
        super.init(coder: coder)
        setupUI()
    }
    
    private func setupUI() {
        // Configure cell background
        wantsLayer = true
        layer?.backgroundColor = NSColor(white: 0.95, alpha: 1.0).cgColor
        
        // Add content view to cell
        addSubview(contentView)
        
        // Configure image view reference and add to content view
        self.imageView = customImageView
        contentView.addSubview(customImageView)
        
        // Add rightContentView to content view
        contentView.addSubview(rightContentView)
        
        // Add text fields to rightContentView
        rightContentView.addSubview(mainTextField)
        rightContentView.addSubview(fileNameTextField)
        rightContentView.addSubview(subTextField)
        
        // Add new UI elements to content view
        contentView.addSubview(statusLabel)
        contentView.addSubview(receiveBtn)
        contentView.addSubview(customProgressView) // Add custom progress view
        contentView.addSubview(openBtn)
        contentView.addSubview(deleteBtn)
        
        // Set constraints
        setupConstraints()
    }
    
    private func setupConstraints() {
        // Content view constraints - height 50, centered in cell
        contentView.snp.makeConstraints { make in
            make.top.equalTo(5)
            make.height.equalTo(70)
            make.leading.equalTo(0)
            make.trailing.equalTo(0)
        }
        
        // Image view constraints
        customImageView.snp.makeConstraints { make in
            make.centerY.equalToSuperview()
            make.leading.equalToSuperview().offset(5)
            make.width.height.equalTo(25)
        }
        
        // rightContentView 约束 - 上下居中
        rightContentView.snp.makeConstraints { make in
            make.centerY.equalToSuperview()
            make.leading.equalTo(customImageView.snp.trailing).offset(5)
            make.trailing.equalTo(statusLabel.snp.leading).offset(-10)
        }
        
        // mainTextField 约束 - 在 rightContentView 内
        mainTextField.snp.makeConstraints { make in
            make.top.equalToSuperview()
            make.leading.trailing.equalToSuperview()
        }
        
        // fileNameTextField 约束 - 在 rightContentView 内，初始时隐藏
        fileNameTextField.snp.makeConstraints { make in
            make.top.equalTo(mainTextField.snp.bottom).offset(5)
            make.leading.trailing.equalToSuperview()
        }
        
        // subTextField 约束 - 在 rightContentView 内
        subTextField.snp.makeConstraints { make in
            make.top.equalTo(fileNameTextField.snp.bottom).offset(5)
            make.leading.trailing.equalToSuperview()
            make.bottom.equalToSuperview()
        }
        
        // Right side elements constraints
        deleteBtn.snp.makeConstraints { make in
            make.centerY.equalToSuperview()
            make.trailing.equalToSuperview().offset(-5)
            make.width.height.equalTo(24)
        }
        
        openBtn.snp.makeConstraints { make in
            make.centerY.equalToSuperview()
            make.trailing.equalTo(deleteBtn.snp.leading).offset(-5)
            make.width.height.equalTo(24)
        }
        
        receiveBtn.snp.makeConstraints { make in
            make.centerY.equalToSuperview()
            make.trailing.equalTo(openBtn.snp.leading).offset(-5)
            make.width.height.equalTo(24)
        }
        
        // 自定义进度条约束
        customProgressView.snp.makeConstraints { make in
            make.centerY.equalToSuperview()
            make.trailing.equalTo(receiveBtn.snp.leading).offset(-5)
            make.width.equalTo(160)
            make.height.equalTo(18) // 进度条高度
        }
        
        // Keep statusLabel constraints with fixed width
        statusLabel.snp.makeConstraints { make in
            make.centerY.equalToSuperview()
            make.trailing.equalTo(receiveBtn.snp.leading).offset(-5)
            make.width.equalTo(160) // Fixed width
        }
    }
    
    // Handle open button click
    @objc private func handleOpenButtonClick() {
//        logger.info("filePath:\(self.currentFileInfo?.session.currentFileName)")
        // Ensure currentFileInfo and file path are not empty
        guard let filePath = self.currentFileInfo?.session.currentFileName, !filePath.isEmpty else {
            CSAlertManager.shared.showInvalidFilePath()
            return
        }
        
        let url = URL(fileURLWithPath: filePath)
        
        // Check if file exists
        if FileManager.default.fileExists(atPath: url.path) {
            NSWorkspace.shared.open(url)
        } else {
            CSAlertManager.shared.showFileNotFound(filePath: filePath)
        }
    }
    
    // Handle delete/cancel button click
    @objc private func handleDeleteButtonClick() {
        guard let fileInfo = currentFileInfo else { return }
        
        // 判断当前传输状态，决定是取消还是删除
        if fileInfo.isCompleted {
            // 传输完成：执行删除操作
            onDelete?(fileInfo.sessionId)
        } else {
            // 传输中或传输开始：执行取消操作
            onCancel?(fileInfo.sessionId)
        }
    }
    
    
    // Configure cell with CSFileInfo
    func configure(with fileInfo: CSFileInfo) {
        // Save reference to fileInfo
        currentFileInfo = fileInfo
        
        // Set fixed icon, use system default file icon if not found
        if let csFileImage = NSImage(named: "CSFile") {
            customImageView.image = csFileImage
        } else {
            // If CSFile icon not found, use default document icon
            customImageView.image = NSWorkspace.shared.icon(for: UTType.data)
            logger.info("Warning: CSFile icon not found, using default document icon")
        }
        
        let dateFormatter = DateFormatter()
        dateFormatter.dateFormat = "yyyy-MM-dd HH:mm:ss"
        let date = dateFormatter.string(from: fileInfo.session.startTime)
        
        // Extract filename from path
        let fileName = URL(string: fileInfo.session.currentFileName)?.lastPathComponent ?? fileInfo.session.currentFileName
        
        // 根据文件数量决定布局
        if fileInfo.session.totalFileCount == 1 {
            // 单文件：只显示 mainTextField 和 subTextField
            fileNameTextField.isHidden = true
            
            // 更新 subTextField 的约束，使其直接跟在 mainTextField 后面
            subTextField.snp.remakeConstraints { make in
                make.top.equalTo(mainTextField.snp.bottom).offset(5)
                make.leading.trailing.equalToSuperview()
                make.bottom.equalToSuperview()
            }
            
            mainTextField.stringValue = fileName + " (" + "\(fileInfo.session.formattedTotalSize)" + ")"
            
        } else {
            // 多文件：显示三个文本框
            fileNameTextField.isHidden = false
            
            // 恢复 subTextField 的约束，使其跟在 fileNameTextField 后面
            subTextField.snp.remakeConstraints { make in
                make.top.equalTo(fileNameTextField.snp.bottom).offset(5)
                make.leading.trailing.equalToSuperview()
                make.bottom.equalToSuperview()
            }
            
            mainTextField.stringValue = "total: \(fileInfo.session.receivedFileCount)/\(fileInfo.session.totalFileCount) files" + " (" + "\(fileInfo.session.formattedReceivedSize)" + " / "  + "\(fileInfo.session.formattedTotalSize)" + ")"
            
            // fileNameTextField 根据传输状态显示不同内容
            if fileInfo.isCompleted {
                fileNameTextField.stringValue = "All done"
            } else {
                fileNameTextField.stringValue = fileName + " (" + "\(fileInfo.session.formattedTotalSize)" + ")"
            }
        }
        
        // Display different information based on transfer status
        if fileInfo.isCompleted {
            // 传输完成：显示删除按钮
            subTextField.stringValue = date + " from: " + fileInfo.session.deviceName 
            customProgressView.isHidden = true
            statusLabel.stringValue = "Transmission completed"
            statusLabel.isHidden = false
            openBtn.isHidden = false
            openBtn.isEnabled = true  // 允许点击
            receiveBtn.isHidden = false
            statusLabel.textColor = NSColor.gray

            if let image = NSImage(named: "receiveTickSucess") {
                image.size = NSSize(width: 16, height: 16)
                receiveBtn.image = image
            }

            
            // 恢复 openBtn 原来的图标
            if let openImage = NSImage(named: "open") {
                openImage.size = NSSize(width: 16, height: 16)
                openBtn.image = openImage
            }
            
            // 设置为删除图标
            if let deleteImage = NSImage(named: "delete") {
                deleteImage.size = NSSize(width: 16, height: 16)
                deleteBtn.image = deleteImage
            }
        } else if fileInfo.progress > 0 && fileInfo.progress < 1 {
            // 传输中：显示取消按钮
            subTextField.stringValue = date + " from: " + fileInfo.session.deviceName
            // 打印 errCode
            if let errCode = fileInfo.errCode {
                logger.info("CSBottomTableViewCell - errCode: \(errCode)")
                customProgressView.isHidden = true
                customProgressView.progress = fileInfo.progress
                
                // 设置为删除图标
                if let deleteImage = NSImage(named: "delete") {
                    deleteImage.size = NSSize(width: 16, height: 16)
                    deleteBtn.image = deleteImage
                }
                
                statusLabel.isHidden = false
                statusLabel.stringValue = useErrorCodeBackVauleTex(errCode)
                statusLabel.textColor = NSColor.red
                
                openBtn.isHidden = true
                receiveBtn.isHidden = false
                
                if let receiveTickFailImage = NSImage(named: "receiveTickFail") {
                    receiveTickFailImage.size = NSSize(width: 16, height: 16)
                    receiveBtn.image = receiveTickFailImage
                }

            } else {
                logger.info("CSBottomTableViewCell - errCode: nil")
                customProgressView.isHidden = false
                customProgressView.progress = fileInfo.progress
                
                statusLabel.isHidden = true
                openBtn.isHidden = true
                receiveBtn.isHidden = true
                
                // 设置为关闭图标
                if let closeImage = NSImage(named: "close") {
                    closeImage.size = NSSize(width: 16, height: 16)
                    deleteBtn.image = closeImage
                }
            }

        } else {
            // 传输开始：显示取消按钮
            customProgressView.isHidden = true
            statusLabel.stringValue = "Transmission start"
            statusLabel.isHidden = false
            openBtn.isHidden = true
            receiveBtn.isHidden = true
            
            // 设置为关闭图标
            if let closeImage = NSImage(named: "close") {
                closeImage.size = NSSize(width: 16, height: 16)
                deleteBtn.image = closeImage
            }
        }
        
        // More operations based on fileInfo can be added here
        // For example, displaying transfer direction, device name, etc.
    }
    
    private func useErrorCodeBackVauleTex(_ errorCode:Int) -> String{
        switch errorCode {
        case 5520:
            return "Transmission cancel"
        case 5502, 550,5509,5510,5513,5514,5511,5515,5518,5519,5516:
            return "Transmission Fail"
        default:
            return "Unknown error"
        }
    }
    
}
