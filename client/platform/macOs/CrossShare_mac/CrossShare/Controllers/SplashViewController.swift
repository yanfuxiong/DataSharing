//
//  SplashViewController.swift
//  CrossShare
//
//  Created for splash screen functionality
//

import Cocoa
import QuartzCore
import ImageIO
import SnapKit

class SplashViewController: NSViewController {
    
    // MARK: - UI Components
    private lazy var backgroundImageView: NSImageView = {
        let imageView = DraggableImageView()
        imageView.image = NSImage(named: "CSstartbg")
        imageView.imageScaling = .scaleAxesIndependently
        imageView.wantsLayer = true
        return imageView
    }()
    
    private lazy var closeButton: NSButton = {
        let button = NSButton()
        button.image = NSImage(named: "CSclose")
        button.isBordered = false
        button.bezelStyle = .regularSquare
        button.imageScaling = .scaleProportionallyDown
        button.target = self
        button.action = #selector(closeButtonTapped)
        return button
    }()
    
    private lazy var gifImageView: NSImageView = {
        let imageView = DraggableImageView()
        imageView.imageScaling = .scaleProportionallyDown
        imageView.wantsLayer = true
        return imageView
    }()
    
    private lazy var nameSequenceImageView: NSImageView = {
        let imageView = DraggableImageView()
        imageView.image = NSImage(named: "CSName1")
        imageView.imageScaling = .scaleProportionallyDown
        imageView.wantsLayer = true
        return imageView
    }()
    
    private lazy var messageLabel: NSTextField = {
        let label = NSTextField()
        label.stringValue = "please plug in a CrossShare-compatible monitor."
        label.isEditable = false
        label.isBordered = false
        label.backgroundColor = .clear
        label.alignment = .center
        label.font = NSFont.systemFont(ofSize: 16)
        label.textColor = NSColor.gray
        label.lineBreakMode = .byWordWrapping
        label.maximumNumberOfLines = 2
        return label
    }()
    
    // MARK: - Properties
    private var nameSequenceTimer: Timer?
    private var currentNameIndex = 1
    private let maxNameIndex = 6
    private var splashTimer: Timer?
    private var gifTimer: Timer?
    private var gifFrames: [NSImage] = []
    private var currentGifFrameIndex = 0
    private var configCheckTimer: Timer?  // Timer to check config data
    
    // Completion handler to notify when splash should dismiss
    var onSplashComplete: (() -> Void)?
    
    // MARK: - Lifecycle
    override func loadView() {
        // Create view with 600x500 size
        let view = DraggableView(frame: NSRect(x: 0, y: 0, width: 600, height: 500))
        self.view = view
    }
    
    override func viewDidLoad() {
        super.viewDidLoad()
        
        setupUI()
        startAnimations()
        startSplashTimer()
        startConfigCheckTimer()
    }
    
    // MARK: - Config Check Timer
    
    /// Start timer to check config data every 0.5 seconds
    private func startConfigCheckTimer() {
        configCheckTimer = Timer.scheduledTimer(withTimeInterval: 0.5, repeats: true) { [weak self] _ in
            self?.checkConfigData()
        }
    }
    
    /// Check config data and transition to main page if isInited is true
    private func checkConfigData() {
        guard let config = SharedDataManager.shared.loadConfig() else {
            print("Config data not available yet")
            return
        }
        
        print("  - isInited: \(config.uiTheme.isInitedBool)")
        print("  - customerId: \(config.uiTheme.customerID)")
        
        // Check if isInited is true
        if config.uiTheme.customerID != "-1" {
            print("config customerID = \(config.uiTheme.customerID), transitioning to main page...")
            SharedDataManager.shared.updateIsInited(true)
            
            // Stop config check timer
            configCheckTimer?.invalidate()
            configCheckTimer = nil
            
            // Transition to main page
            completeSplash()
        } else {
            print("Config isInited = false, continue waiting...")
        }
    }
    
    override func viewDidAppear() {
        super.viewDidAppear()
        
        // Ensure corner radius takes effect after view appears
        view.layer?.cornerRadius = 10
        view.layer?.masksToBounds = true
        
        // Also set window corner radius
        if let window = view.window {
            window.contentView?.layer?.cornerRadius = 10
            window.contentView?.layer?.masksToBounds = true
        }
    }
    
    // MARK: - UI Setup
    private func setupUI() {
        // Set view background color with corner radius (matching macOS standard window)
        view.wantsLayer = true
        view.layer?.backgroundColor = NSColor.white.cgColor
        view.layer?.cornerRadius = 10  // macOS standard window corner radius
        view.layer?.masksToBounds = true
        
        // Add subviews
        view.addSubview(backgroundImageView)
        view.addSubview(closeButton)
        view.addSubview(gifImageView)
        view.addSubview(nameSequenceImageView)
        view.addSubview(messageLabel)
        
        // Setup constraints
        setupConstraints()
        
        // Load GIF frames
        loadGifFrames()
    }
    
    private func setupConstraints() {
        // Background image
        backgroundImageView.snp.makeConstraints { make in
            make.edges.equalToSuperview()
        }
        
        // Close button (top right corner)
        closeButton.snp.makeConstraints { make in
            make.top.equalToSuperview().offset(10)
            make.trailing.equalToSuperview().offset(-10)
            make.width.height.equalTo(40)
        }
        
        // GIF loading animation (centered with offset)
        gifImageView.snp.makeConstraints { make in
            make.centerX.equalToSuperview()
            make.top.equalToSuperview().offset(100)
            make.width.height.equalTo(100)
        }
        
        // Name sequence image view (below GIF)
        nameSequenceImageView.snp.makeConstraints { make in
            make.centerX.equalToSuperview()
            make.top.equalTo(gifImageView.snp.bottom).offset(15)
            make.width.equalTo(300)
            make.height.equalTo(85)
        }
        
        // Message label (below name sequence with larger spacing)
        messageLabel.snp.makeConstraints { make in
            make.centerX.equalToSuperview()
            make.top.equalTo(nameSequenceImageView.snp.bottom).offset(30)  // Increased spacing from 15 to 30
            make.leading.equalToSuperview().offset(30)
            make.trailing.equalToSuperview().offset(-30)
            make.height.equalTo(40)
        }
    }
    
    // MARK: - GIF Loading
    private func loadGifFrames() {
        // Method 1: Search in gif subdirectory
        if let gifPath = Bundle.main.path(forResource: "CSloading-5", ofType: "gif", inDirectory: "gif") {
            print("Found GIF file (method 1): \(gifPath)")
            loadGifFromPath(gifPath)
            return
        }
        
        // Method 2: Search directly in Resources directory
        if let gifPath = Bundle.main.path(forResource: "CSloading-5", ofType: "gif") {
            print("Found GIF file (method 2): \(gifPath)")
            loadGifFromPath(gifPath)
            return
        }
        
        // Method 3: List all gif files
        if let resourcePath = Bundle.main.resourcePath {
            print("Resource path: \(resourcePath)")
            let fileManager = FileManager.default
            do {
                let contents = try fileManager.contentsOfDirectory(atPath: resourcePath)
                let gifFiles = contents.filter { $0.contains(".gif") }
                print("Found gif files: \(gifFiles)")
                
                // Try to find gif subdirectory
                let gifDirPath = (resourcePath as NSString).appendingPathComponent("gif")
                if fileManager.fileExists(atPath: gifDirPath) {
                    let gifDirContents = try fileManager.contentsOfDirectory(atPath: gifDirPath)
                    print("Gif directory contents: \(gifDirContents)")
                }
            } catch {
                print("Failed to read directory: \(error)")
            }
        }
        print("GIF file not found")
    }
    
    private func loadGifFromPath(_ path: String) {
        let gifURL = URL(fileURLWithPath: path)
        guard let imageSource = CGImageSourceCreateWithURL(gifURL as CFURL, nil) else {
            print("Failed to create image source")
            return
        }
        
        let frameCount = CGImageSourceGetCount(imageSource)
        print("GIF frame count: \(frameCount)")
        
        for i in 0..<frameCount {
            if let cgImage = CGImageSourceCreateImageAtIndex(imageSource, i, nil) {
                let nsImage = NSImage(cgImage: cgImage, size: NSSize(width: cgImage.width, height: cgImage.height))
                gifFrames.append(nsImage)
                print("Loaded frame \(i+1)")
            }
        }
        
        // Set first frame
        if !gifFrames.isEmpty {
            gifImageView.image = gifFrames[0]
        }
    }
    
    // MARK: - Animations
    private func startAnimations() {
        // Start GIF animation
        if !gifFrames.isEmpty {
            print("Starting GIF animation with \(gifFrames.count) frames")
            gifTimer = Timer.scheduledTimer(withTimeInterval: 0.1, repeats: true) { [weak self] _ in
                self?.updateGifFrame()
            }
        } else {
            print("No GIF frames to animate")
        }
        
        // Start name sequence animation (loop through CSName1 to CSName6)
        // Switching speed set to 0.8 seconds (slower than the original 0.5 seconds)
        nameSequenceTimer = Timer.scheduledTimer(withTimeInterval: 0.8, repeats: true) { [weak self] _ in
            self?.updateNameSequence()
        }
        print("Starting name sequence animation")
    }
    
    private func updateGifFrame() {
        currentGifFrameIndex += 1
        if currentGifFrameIndex >= gifFrames.count {
            currentGifFrameIndex = 0
        }
        gifImageView.image = gifFrames[currentGifFrameIndex]
    }
    
    private func updateNameSequence() {
        currentNameIndex += 1
        if currentNameIndex > maxNameIndex {
            currentNameIndex = 1
        }
        
        let imageName = "CSName\(currentNameIndex)"
        if let image = NSImage(named: imageName) {
            nameSequenceImageView.image = image
            print("Switched to: \(imageName), image size: \(image.size)")
            
            // Add fade transition
            let transition = CATransition()
            transition.duration = 0.3
            transition.type = .fade
            nameSequenceImageView.layer?.add(transition, forKey: "fadeTransition")
        } else {
            print("Image not found: \(imageName)")
        }
    }
    
    // MARK: - Timer
    private func startSplashTimer() {
        // Show splash for 5 seconds then transition to main window
//        splashTimer = Timer.scheduledTimer(withTimeInterval: 5.0, repeats: false) { [weak self] _ in
//            self?.completeSplash()
//        }
    }
    
    private func completeSplash() {
        DispatchQueue.main.asyncAfter(deadline: .now() + 5.0) { [weak self] in
            self?.cleanup()
            self?.onSplashComplete?()
        }
    }
    
    // MARK: - Actions
    @objc private func closeButtonTapped() {
        cleanup()
        NSApplication.shared.terminate(nil)
    }
    
    // MARK: - Cleanup
    private func cleanup() {
        gifTimer?.invalidate()
        gifTimer = nil
        nameSequenceTimer?.invalidate()
        nameSequenceTimer = nil
        splashTimer?.invalidate()
        splashTimer = nil
        configCheckTimer?.invalidate()
        configCheckTimer = nil
    }
    
    deinit {
        cleanup()
    }
}

// MARK: - DraggableView
/// Custom view that allows dragging the window by clicking anywhere on it
class DraggableView: NSView {
    override func mouseDown(with event: NSEvent) {
        // Allow dragging the window from anywhere in the view
        window?.performDrag(with: event)
    }
}

// MARK: - DraggableImageView
/// Custom image view that allows dragging the window by clicking on it
class DraggableImageView: NSImageView {
    override func mouseDown(with event: NSEvent) {
        // Allow dragging the window from the image view
        window?.performDrag(with: event)
    }
}

