//
//  CSToastManager.swift
//  CrossShare
//
//  Toast manager for displaying non-blocking toast messages on macOS
//

import Cocoa

enum ToastPosition {
    case top
    case center
    case bottom
}

enum ToastType {
    case info
    case success
    case warning
    case error
}

class CSToastManager {
    
    static let shared = CSToastManager()
    
    private var activeToasts: [NSWindow] = []
    private let toastQueue = DispatchQueue(label: "com.crossshare.toast", qos: .userInitiated)
    
    private init() {}
    
    // MARK: - Public Methods
    
    /// Show a toast message
    /// - Parameters:
    ///   - message: The message to display
    ///   - duration: Display duration in seconds (default: 2.0)
    ///   - position: Toast position (default: .center)
    ///   - type: Toast type (default: .info)
    ///   - inView: Optional view to show toast in (default: nil, shows in key window)
    func showToast(
        message: String,
        duration: TimeInterval = 2.0,
        position: ToastPosition = .center,
        type: ToastType = .info,
        inView: NSView? = nil
    ) {
        DispatchQueue.main.async { [weak self] in
            self?.displayToast(message: message, duration: duration, position: position, type: type, inView: inView)
        }
    }
    
    /// Show a success toast (centered by default)
    func showSuccess(_ message: String, duration: TimeInterval = 2.0, inView: NSView? = nil) {
        showToast(message: message, duration: duration, position: .center, type: .success, inView: inView)
    }
    
    /// Show an error toast (centered by default)
    func showError(_ message: String, duration: TimeInterval = 2.5, inView: NSView? = nil) {
        showToast(message: message, duration: duration, position: .center, type: .error, inView: inView)
    }
    
    /// Show a warning toast (centered by default)
    func showWarning(_ message: String, duration: TimeInterval = 2.0, inView: NSView? = nil) {
        showToast(message: message, duration: duration, position: .center, type: .warning, inView: inView)
    }
    
    /// Show an info toast (centered by default)
    func showInfo(_ message: String, duration: TimeInterval = 2.0, inView: NSView? = nil) {
        showToast(message: message, duration: duration, position: .center, type: .info, inView: inView)
    }
    
    // MARK: - Private Methods
    
    private func displayToast(
        message: String,
        duration: TimeInterval,
        position: ToastPosition,
        type: ToastType,
        inView: NSView?
    ) {
        guard let targetView = inView ?? getKeyWindowContentView() else {
            print("CSToastManager: No view available to display toast")
            return
        }
        
        guard let targetWindow = targetView.window else {
            print("CSToastManager: Target view has no window")
            return
        }
        
        let toastWindow = createToastWindow(message: message, type: type)
        positionToastWindow(toastWindow, in: targetWindow, position: position)
        
        // Add to active toasts
        activeToasts.append(toastWindow)
        
        // Show toast with animation
        toastWindow.alphaValue = 0.0
        toastWindow.makeKeyAndOrderFront(nil)
        
        NSAnimationContext.runAnimationGroup({ context in
            context.duration = 0.2
            context.timingFunction = CAMediaTimingFunction(name: .easeOut)
            toastWindow.animator().alphaValue = 1.0
        })
        
        // Auto dismiss after duration
        DispatchQueue.main.asyncAfter(deadline: .now() + duration) { [weak self] in
            guard let self = self else { return }
            self.dismissToast(toastWindow)
        }
    }
    
    private func createToastWindow(message: String, type: ToastType) -> NSWindow {
        // Calculate size - 更大的尺寸
        let maxWidth: CGFloat = 500
        let padding: CGFloat = 30
        let minHeight: CGFloat = 80
        
        let font = NSFont.systemFont(ofSize: 18, weight: .medium)
        let attributes: [NSAttributedString.Key: Any] = [.font: font]
        let size = (message as NSString).boundingRect(
            with: NSSize(width: maxWidth - padding * 2, height: .greatestFiniteMagnitude),
            options: [.usesLineFragmentOrigin, .usesFontLeading],
            attributes: attributes
        )
        
        let textHeight = ceil(size.height)
        let windowHeight = max(minHeight, textHeight + padding * 2)
        let windowWidth = max(300, min(maxWidth, size.width + padding * 2))
        
        // Create window
        let window = NSWindow(
            contentRect: NSRect(x: 0, y: 0, width: windowWidth, height: windowHeight),
            styleMask: [.borderless],
            backing: .buffered,
            defer: false
        )
        
        window.backgroundColor = .clear
        window.isOpaque = false
        window.level = .floating
        window.ignoresMouseEvents = true
        window.collectionBehavior = [.canJoinAllSpaces, .stationary]
        window.isReleasedWhenClosed = false  // 手动管理窗口生命周期
        window.hidesOnDeactivate = false    // 不随应用失活而隐藏
        
        // Create content view
        guard let contentViewContainer = window.contentView else {
            fatalError("Window contentView is nil")
        }
        
        let contentView = NSView(frame: contentViewContainer.bounds)
        contentView.wantsLayer = true
        contentView.layer?.cornerRadius = 12
        contentView.layer?.masksToBounds = true
        
        // Set background color based on type
        let backgroundColor: NSColor
        switch type {
        case .success:
            backgroundColor = NSColor(red: 0.2, green: 0.7, blue: 0.3, alpha: 0.95)
        case .error:
            backgroundColor = NSColor(red: 0.9, green: 0.3, blue: 0.3, alpha: 0.95)
        case .warning:
            backgroundColor = NSColor(red: 1.0, green: 0.7, blue: 0.2, alpha: 0.95)
        case .info:
            backgroundColor = NSColor(red: 0.3, green: 0.5, blue: 0.9, alpha: 0.95)
        }
        
        contentView.layer?.backgroundColor = backgroundColor.cgColor
        
        // Add shadow
        contentView.shadow = NSShadow()
        contentView.shadow?.shadowColor = NSColor.black.withAlphaComponent(0.3)
        contentView.shadow?.shadowOffset = NSSize(width: 0, height: -2)
        contentView.shadow?.shadowBlurRadius = 8
        
        // Create text label
        let textField = NSTextField(labelWithString: message)
        textField.font = font
        textField.textColor = .white
        textField.alignment = .center
        textField.lineBreakMode = .byWordWrapping
        textField.maximumNumberOfLines = 0
        textField.isEditable = false
        textField.isBezeled = false
        textField.drawsBackground = false
        
        let textFrame = NSRect(
            x: padding,
            y: padding,
            width: windowWidth - padding * 2,
            height: textHeight
        )
        textField.frame = textFrame
        
        contentView.addSubview(textField)
        window.contentView = contentView
        
        return window
    }
    
    private func positionToastWindow(_ window: NSWindow, in targetWindow: NSWindow, position: ToastPosition) {
        let windowFrame = window.frame
        let targetWindowFrame = targetWindow.frame
        
        var x: CGFloat
        var y: CGFloat
        
        // Horizontal: center in app window
        x = targetWindowFrame.midX - windowFrame.width / 2
        
        // Vertical: based on position (relative to app window)
        // macOS coordinate system: origin is at bottom-left
        switch position {
        case .top:
            y = targetWindowFrame.maxY - windowFrame.height - 50
        case .center:
            y = targetWindowFrame.midY - windowFrame.height / 2
        case .bottom:
            y = targetWindowFrame.minY + 50
        }
        
        // For center position, don't adjust for overlapping (usually only one center toast)
        // For other positions, adjust if needed
        if position == .center {
            // Center toasts typically don't overlap, but if they do, stack them
            let adjustedY = adjustYPositionForCenter(y, windowHeight: windowFrame.height)
            window.setFrameOrigin(NSPoint(x: x, y: adjustedY))
        } else {
            let adjustedY = adjustYPosition(y, windowHeight: windowFrame.height)
            window.setFrameOrigin(NSPoint(x: x, y: adjustedY))
        }
    }
    
    private func adjustYPosition(_ baseY: CGFloat, windowHeight: CGFloat) -> CGFloat {
        let spacing: CGFloat = 10
        var adjustedY = baseY
        
        // Check for overlapping toasts (for top/bottom positions)
        for existingToast in activeToasts {
            let existingFrame = existingToast.frame
            // Check if y positions overlap
            if abs(existingFrame.midY - baseY) < (existingFrame.height / 2 + windowHeight / 2 + spacing) {
                if existingFrame.midY > baseY {
                    adjustedY = existingFrame.minY - windowHeight - spacing
                } else {
                    adjustedY = existingFrame.maxY + spacing
                }
            }
        }
        
        return adjustedY
    }
    
    private func adjustYPositionForCenter(_ baseY: CGFloat, windowHeight: CGFloat) -> CGFloat {
        let spacing: CGFloat = 20
        var adjustedY = baseY
        var offset: CGFloat = 0
        
        // Count how many center toasts already exist
        for existingToast in activeToasts {
            let existingFrame = existingToast.frame
            // If there's a center toast, stack them vertically
            if abs(existingFrame.midY - baseY) < 100 { // Within 100 points, consider it center
                offset += existingFrame.height + spacing
            }
        }
        
        // Stack new toast above existing ones
        adjustedY = baseY + offset / 2
        
        return adjustedY
    }
    
    private func dismissToast(_ window: NSWindow) {
        // 确保在主线程执行
        guard Thread.isMainThread else {
            DispatchQueue.main.async { [weak self] in
                self?.dismissToast(window)
            }
            return
        }
        
        // 先从数组中移除引用，避免在动画过程中被访问
        if let index = activeToasts.firstIndex(of: window) {
            activeToasts.remove(at: index)
        }
        
        // 执行淡出动画
        NSAnimationContext.runAnimationGroup({ context in
            context.duration = 0.2
            context.timingFunction = CAMediaTimingFunction(name: .easeIn)
            context.allowsImplicitAnimation = true
            window.animator().alphaValue = 0.0
        }, completionHandler: { [weak window] in
            // 使用 weak 引用，避免强引用导致窗口无法释放
            guard let window = window else { return }
            
            // 在主线程上执行清理
            DispatchQueue.main.async {
                // 先隐藏窗口，避免在关闭时仍然可见
                window.orderOut(nil)
                
                // 清空 contentView 的子视图，避免访问已释放的视图
                window.contentView?.subviews.forEach { $0.removeFromSuperview() }
                
                // 清空 contentView，避免在关闭时访问
                window.contentView = nil
                
                // 延迟关闭，确保所有清理操作完成
                // 使用更短的延迟，避免窗口长时间占用资源
                DispatchQueue.main.asyncAfter(deadline: .now() + 0.1) { [weak window] in
                    guard let window = window else { return }
                    // 检查窗口是否仍然有效
                    if !window.isReleasedWhenClosed {
                        // 设置自动释放，然后关闭
                        window.isReleasedWhenClosed = true
                        window.close()
                    }
                }
            }
        })
    }
    
    private func getKeyWindowContentView() -> NSView? {
        return NSApplication.shared.keyWindow?.contentView
    }
}

// MARK: - NSView Extension

extension NSView {
    /// Show a toast message on this view
    func showToast(_ message: String, duration: TimeInterval = 2.0, type: ToastType = .info) {
        CSToastManager.shared.showToast(message: message, duration: duration, type: type, inView: self)
    }
    
    /// Show a success toast
    func showSuccessToast(_ message: String, duration: TimeInterval = 2.0) {
        CSToastManager.shared.showSuccess(message, duration: duration, inView: self)
    }
    
    /// Show an error toast
    func showErrorToast(_ message: String, duration: TimeInterval = 2.5) {
        CSToastManager.shared.showError(message, duration: duration, inView: self)
    }
    
    /// Show a warning toast
    func showWarningToast(_ message: String, duration: TimeInterval = 2.0) {
        CSToastManager.shared.showWarning(message, duration: duration, inView: self)
    }
    
    /// Show an info toast
    func showInfoToast(_ message: String, duration: TimeInterval = 2.0) {
        CSToastManager.shared.showInfo(message, duration: duration, inView: self)
    }
}

