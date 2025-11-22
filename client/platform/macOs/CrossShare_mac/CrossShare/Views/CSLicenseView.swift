//
//  CSLicenseView.swift
//  CrossShare
//
//  Created by Assistant on 2025/10/20.
//

import Cocoa
import SnapKit

class CSLicenseView: NSView {
    
    // MARK: - Properties
    
    private var libraryName: String
    private var licenseText: String
    var onBackButtonTapped: (() -> Void)?
    
    // MARK: - UI Components
    
    // Top navigation bar
    private lazy var navigationBar: NSView = {
        let view = NSView()
        view.wantsLayer = true
        view.layer?.backgroundColor = NSColor.controlBackgroundColor.cgColor
        return view
    }()
    
    // Back button
    private lazy var backButton: NSButton = {
        let button = NSButton()
        button.isBordered = false
        button.bezelStyle = .regularSquare
        if let image = NSImage(named: "backArrows") {
            image.size = NSSize(width: 20, height: 20)
            button.image = image
        }
        button.target = self
        button.action = #selector(backButtonTapped)
        return button
    }()
    
    // Title label
    private lazy var titleLabel: NSTextField = {
        let label = NSTextField(labelWithString: libraryName)
        label.font = NSFont.boldSystemFont(ofSize: 16)
        label.textColor = NSColor.labelColor
        label.alignment = .center
        label.isEditable = false
        label.isBordered = false
        label.backgroundColor = .clear
        return label
    }()
    
    // Separator line
    private lazy var separatorLine: NSBox = {
        let box = NSBox()
        box.boxType = .separator
        return box
    }()
    
    // Scroll view for license content
    private var scrollView: NSScrollView!
    
    // Text view for license content
    private var textView: NSTextView!
    
    // MARK: - Initialization
    
    init(frame frameRect: NSRect, libraryName: String, licenseText: String) {
        self.libraryName = libraryName
        self.licenseText = licenseText
        super.init(frame: frameRect)
        setupUI()
    }
    
    required init?(coder: NSCoder) {
        self.libraryName = ""
        self.licenseText = ""
        super.init(coder: coder)
        setupUI()
    }
    
    // MARK: - UI Setup
    
    private func setupUI() {
        wantsLayer = true
        layer?.backgroundColor = NSColor.controlBackgroundColor.cgColor
        
        // Create scroll view
        scrollView = NSScrollView()
        scrollView.hasVerticalScroller = true
        scrollView.hasHorizontalScroller = false
        scrollView.autohidesScrollers = false
        scrollView.borderType = .noBorder
        scrollView.drawsBackground = false
        
        // Create text view with proper frame
        // 初始化 TextView，初始大小会在后面重新设置
        textView = NSTextView()
        
        // Text view configuration
        textView.isEditable = false
        textView.isSelectable = true
        textView.backgroundColor = .clear
        textView.textColor = NSColor.labelColor
        textView.font = NSFont.systemFont(ofSize: 12)
        textView.textContainerInset = NSSize(width: 10, height: 10)
        
        // Enable text wrapping - 关键配置
        textView.isVerticallyResizable = true
        textView.isHorizontallyResizable = false
        textView.autoresizingMask = []
        
        // Set size constraints for proper wrapping
        textView.minSize = NSSize(width: 0, height: 0)
        textView.maxSize = NSSize(width: CGFloat.greatestFiniteMagnitude, height: CGFloat.greatestFiniteMagnitude)
        
        // Configure text container for wrapping
        if let textContainer = textView.textContainer, let layoutManager = textView.layoutManager {
            // 禁用水平滚动，强制文本换行
            textContainer.widthTracksTextView = false
            textContainer.heightTracksTextView = false
            
            // 设置容器大小 - 高度无限，宽度会在 layout 中设置
            textContainer.containerSize = NSSize(width: 349, height: CGFloat.greatestFiniteMagnitude)
            
            // 使用 byCharWrapping 来处理长 URL，确保不会被截断
            textContainer.lineBreakMode = .byCharWrapping
            
            // 确保 layout manager 会自动换行
            layoutManager.allowsNonContiguousLayout = false
        }
        
        // Set the text after all configurations
        textView.string = licenseText
        
        // Set text view as document view
        scrollView.documentView = textView
        
        // Add subviews
        addSubview(navigationBar)
        navigationBar.addSubview(backButton)
        navigationBar.addSubview(titleLabel)
        addSubview(separatorLine)
        addSubview(scrollView)
        
        setupConstraints()
    }
    
    private func setupConstraints() {
        // Navigation bar
        navigationBar.snp.makeConstraints { make in
            make.top.left.right.equalToSuperview()
            make.height.equalTo(44)
        }
        
        // Back button
        backButton.snp.makeConstraints { make in
            make.left.equalToSuperview().offset(8)
            make.centerY.equalToSuperview()
            make.width.height.equalTo(24)
        }
        
        // Title label
        titleLabel.snp.makeConstraints { make in
            make.center.equalToSuperview()
            make.left.greaterThanOrEqualTo(backButton.snp.right).offset(8)
            make.right.lessThanOrEqualToSuperview().offset(-8)
        }
        
        // Separator line
        separatorLine.snp.makeConstraints { make in
            make.top.equalTo(navigationBar.snp.bottom)
            make.left.equalToSuperview().offset(8)
            make.right.equalToSuperview().offset(-8)
            make.height.equalTo(1)
        }
        
        // Scroll view
        scrollView.snp.makeConstraints { make in
            make.top.equalTo(separatorLine.snp.bottom).offset(8)
            make.left.equalToSuperview().offset(8)
            make.right.equalToSuperview().offset(-8)
            make.bottom.equalToSuperview()
        }
    }
    
    // MARK: - Layout
    
    override func layout() {
        super.layout()
        
        // Update text container width when view size changes
        guard let scrollView = scrollView, 
              let textView = textView,
              let textContainer = textView.textContainer else {
            return
        }
        
        // 计算实际可用宽度
        let scrollViewWidth = scrollView.bounds.width
        let scrollerWidth: CGFloat = 15    // 滚动条宽度
        let textInset: CGFloat = 20        // TextView 的 textContainerInset (左右各10)
        
        // 文本容器的实际宽度
        let containerWidth = max(100, scrollViewWidth - scrollerWidth - textInset)
        
        // 设置 TextView 的 frame 宽度
        var frame = textView.frame
        frame.size.width = containerWidth
        textView.frame = frame
        
        // 更新 textContainer 的宽度
        textContainer.containerSize = NSSize(width: containerWidth, height: CGFloat.greatestFiniteMagnitude)
        
        // 强制重新布局文本
        if let layoutManager = textView.layoutManager {
            layoutManager.ensureLayout(for: textContainer)
        }
    }
    
    // MARK: - Actions
    
    @objc private func backButtonTapped() {
        onBackButtonTapped?()
    }
}

