import Cocoa

// MARK: - CSScrollTextField
// Custom scrollable text field that automatically scrolls when text content exceeds control width
class CSScrollTextField: NSTextField {
    
    // Scroll animation timer
    private var scrollTimer: Timer?
    
    // Current scroll offset
    private var scrollOffset: CGFloat = 0
    
    // Whether scrolling is in progress
    private var isScrolling: Bool = false
    
    // Whether mouse is inside the control
    private var isMouseInside: Bool = false
    
    // Mouse tracking area
    private var trackingArea: NSTrackingArea?
    
    // Scroll speed (pixels per second)
    private let scrollSpeed: CGFloat = 30.0
    
    // Pause duration after scrolling to end (seconds)
    private let pauseDuration: TimeInterval = 1.0
    
    // Delay time before triggering scroll after mouse hover (seconds)
    private let hoverDelay: TimeInterval = 0.5
    
    // Delayed scroll work item
    private var delayedScrollWorkItem: DispatchWorkItem?
    
    // Text width
    private var textWidth: CGFloat = 0
    
    // Control width
    private var controlWidth: CGFloat = 0
    
    // Whether scrolling is needed (cached calculation result)
    private var _needsScroll: Bool = false
    
    // Calculate whether scrolling is needed
    private func updateScrollNeeds() {
        let text = self.stringValue
        if text.isEmpty {
            _needsScroll = false
            textWidth = 0
            controlWidth = 0
            return
        }
        
        // Calculate text width
        let attributes: [NSAttributedString.Key: Any] = [
            .font: self.font ?? NSFont.systemFont(ofSize: 14)
        ]
        let attributedString = NSAttributedString(string: text, attributes: attributes)
        textWidth = attributedString.size().width
        
        // Get actual control width
        controlWidth = max(self.bounds.width, 1) // Avoid division by zero
        
        _needsScroll = textWidth > controlWidth
    }
    
    override init(frame frameRect: NSRect) {
        super.init(frame: frameRect)
        setupTextField()
    }
    
    required init?(coder: NSCoder) {
        super.init(coder: coder)
        setupTextField()
    }
    
    private func setupTextField() {
        // Setup text field properties
        self.isEditable = false
        self.isBordered = false
        self.backgroundColor = .clear
        self.cell?.wraps = false
        self.cell?.isScrollable = true
        
        // Listen to text and size changes
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(handleTextDidChange(_:)),
            name: NSControl.textDidChangeNotification,
            object: self
        )
        
        // Setup mouse tracking
        setupTrackingArea()
    }
    
    private func setupTrackingArea() {
        // Remove old tracking area
        if let oldTrackingArea = trackingArea {
            removeTrackingArea(oldTrackingArea)
        }
        
        // Create new tracking area
        let options: NSTrackingArea.Options = [
            .activeInKeyWindow,
            .mouseEnteredAndExited,
            .inVisibleRect
        ]
        
        trackingArea = NSTrackingArea(
            rect: bounds,
            options: options,
            owner: self,
            userInfo: nil
        )
        
        if let newTrackingArea = trackingArea {
            addTrackingArea(newTrackingArea)
        }
    }
    
    override var stringValue: String {
        didSet {
            stopScrolling()
            // If mouse is inside control, check and start scrolling after delay
            if isMouseInside {
                delayedScrollWorkItem?.cancel()
                let workItem = DispatchWorkItem { [weak self] in
                    guard let self = self, self.isMouseInside else { return }
                    self.checkAndStartScrolling()
                }
                delayedScrollWorkItem = workItem
                DispatchQueue.main.asyncAfter(deadline: .now() + hoverDelay, execute: workItem)
            }
        }
    }
    
    override func viewDidMoveToWindow() {
        super.viewDidMoveToWindow()
        if window != nil {
            setupTrackingArea()
            checkAndStartScrolling()
        } else {
            stopScrolling()
        }
    }
    
    override func updateTrackingAreas() {
        super.updateTrackingAreas()
        setupTrackingArea()
    }
    
    override func resizeSubviews(withOldSize oldSize: NSSize) {
        super.resizeSubviews(withOldSize: oldSize)
        setupTrackingArea()
        stopScrolling()
        checkAndStartScrolling()
    }
    
    override func mouseEntered(with event: NSEvent) {
        super.mouseEntered(with: event)
        isMouseInside = true
        
        // Cancel previous delayed task (if exists)
        delayedScrollWorkItem?.cancel()
        
        // Create delayed task to trigger scrolling after delay
        let workItem = DispatchWorkItem { [weak self] in
            guard let self = self, self.isMouseInside else { return }
            self.checkAndStartScrolling()
        }
        delayedScrollWorkItem = workItem
        
        // Execute after delay
        DispatchQueue.main.asyncAfter(deadline: .now() + hoverDelay, execute: workItem)
    }
    
    override func mouseExited(with event: NSEvent) {
        super.mouseExited(with: event)
        isMouseInside = false
        
        // Cancel delayed task
        delayedScrollWorkItem?.cancel()
        delayedScrollWorkItem = nil
        
        // Stop scrolling
        stopScrolling()
        // Reset scroll position to initial state
        scrollOffset = 0
        needsDisplay = true
    }
    
    override func draw(_ dirtyRect: NSRect) {
        // Update scroll needs
        updateScrollNeeds()
        
        // If scrolling is not needed, use default drawing
        if !_needsScroll {
            scrollOffset = 0
            super.draw(dirtyRect)
            return
        }
        
        // Custom drawing for scrolling text
        guard let cell = self.cell,
              let context = NSGraphicsContext.current?.cgContext else {
            super.draw(dirtyRect)
            return
        }
        
        // Save current graphics context state
        context.saveGState()
        
        // Set clipping region
        let clipRect = self.bounds
        context.clip(to: clipRect)
        
        // Calculate text drawing position (consider vertical alignment)
        let cellHeight = cell.cellSize.height
        let verticalOffset = (self.bounds.height - cellHeight) / 2
        
        let textRect = NSRect(
            x: -scrollOffset,
            y: verticalOffset,
            width: max(textWidth, controlWidth),
            height: cellHeight
        )
        
        // Draw text
        cell.draw(withFrame: textRect, in: self)
        
        // Restore graphics context state
        context.restoreGState()
    }
    
    @objc private func handleTextDidChange(_ notification: Notification) {
        stopScrolling()
        checkAndStartScrolling()
    }
    
    private func checkAndStartScrolling() {
        // Delay check to ensure layout is completed
        DispatchQueue.main.async { [weak self] in
            guard let self = self else { return }
            
            self.updateScrollNeeds()
            
            // Only start if mouse is inside control and scrolling is needed
            if self._needsScroll && !self.isScrolling && self.isMouseInside {
                self.startScrolling()
            } else if !self._needsScroll {
                self.scrollOffset = 0
                self.needsDisplay = true
            }
        }
    }
    
    private func startScrolling() {
        updateScrollNeeds()
        // Only start scrolling if mouse is inside control
        guard _needsScroll && !isScrolling && isMouseInside else { return }
        
        isScrolling = true
        scrollOffset = 0
        
        // Create timer to update scroll position every frame
        scrollTimer = Timer.scheduledTimer(withTimeInterval: 1.0/60.0, repeats: true) { [weak self] _ in
            self?.updateScrollPosition()
        }
        
        // Add timer to run loop's common modes to ensure it works during scrolling
        if let timer = scrollTimer {
            RunLoop.current.add(timer, forMode: .common)
        }
    }
    
    private func updateScrollPosition() {
        // Update scroll needs (size may change during scrolling)
        updateScrollNeeds()
        
        // If mouse left or scrolling is not needed, stop scrolling
        guard _needsScroll && isMouseInside else {
            stopScrolling()
            if !isMouseInside {
                scrollOffset = 0
            }
            needsDisplay = true
            return
        }
        
        // Calculate maximum scroll offset
        let maxScrollOffset = textWidth - controlWidth
        
        if scrollOffset < maxScrollOffset {
            // Continue scrolling to the right
            scrollOffset += scrollSpeed / 60.0 // Distance moved per frame
            scrollOffset = min(scrollOffset, maxScrollOffset) // Ensure it doesn't exceed maximum
            needsDisplay = true
        } else {
            // Scrolled to the end, pause then reset
            stopScrolling()
            
            // Restart scrolling after delay (only if mouse is still inside control)
            DispatchQueue.main.asyncAfter(deadline: .now() + pauseDuration) { [weak self] in
                guard let self = self else { return }
                // Check again if mouse is still inside control
                if self.isMouseInside {
                    self.scrollOffset = 0
                    self.needsDisplay = true
                    self.startScrolling()
                }
            }
        }
    }
    
    private func stopScrolling() {
        scrollTimer?.invalidate()
        scrollTimer = nil
        isScrolling = false
        
        // Cancel delayed task
        delayedScrollWorkItem?.cancel()
        delayedScrollWorkItem = nil
    }
    
    deinit {
        stopScrolling()
        if let trackingArea = trackingArea {
            removeTrackingArea(trackingArea)
        }
        NotificationCenter.default.removeObserver(self)
    }
}

