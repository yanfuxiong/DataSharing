
import Cocoa

class MouseMonitor {
    static let shared = MouseMonitor()
    
    private var lastClickTime: Date?
    private var lastClickLocation: NSPoint?
    private var mouseMonitor: Any?
    
    private let doubleClickInterval: TimeInterval = 0.3
    private let maxDistance: CGFloat = 20
    
    struct DoubleClickEvent {
        let location: NSPoint
        let timestamp: Date
        let interval: TimeInterval
    }
    
    // Add method to get current mouse position
    func getCurrentMousePosition() -> NSPoint {
        return NSEvent.mouseLocation
    }

    // Add method to print current mouse position
    func printCurrentMousePosition() {
        let position = getCurrentMousePosition()
        // The coordinate system of macOS has its origin at the bottom-left corner of the screen; convert it here to use the more intuitive top-left corner as the origin
        let screenHeight = NSScreen.main?.frame.height ?? 0
        let adjustedY = screenHeight - position.y
        
        print(String(format: "The current position of the mouse: X=%.0f, Y=%.0f (The top left corner of the screen is the origin)", position.x, adjustedY))
    }
    
    func startMonitoring(callback: @escaping (DoubleClickEvent) -> Void) {
        mouseMonitor = NSEvent.addGlobalMonitorForEvents(
            matching: [.leftMouseDown],
            handler: { [weak self] event in
                guard let self = self else { return }
                
                let currentTime = Date()
                let currentLocation = event.locationInWindow
                
                if let lastTime = self.lastClickTime,
                   let lastLocation = self.lastClickLocation {
                    
                    let timeSinceLastClick = currentTime.timeIntervalSince(lastTime)
                    let distance = self.distance(currentLocation, lastLocation)
                    
                    if timeSinceLastClick <= self.doubleClickInterval && distance <= self.maxDistance {
                        callback(DoubleClickEvent(
                            location: currentLocation,
                            timestamp: currentTime,
                            interval: timeSinceLastClick
                        ))
                        
                        self.lastClickTime = nil
                        self.lastClickLocation = nil
                        return
                    }
                }
                
                self.lastClickTime = currentTime
                self.lastClickLocation = currentLocation
            }
        )
    }
    
    private func distance(_ p1: NSPoint, _ p2: NSPoint) -> CGFloat {
        let dx = p1.x - p2.x
        let dy = p1.y - p2.y
        return sqrt(dx * dx + dy * dy)
    }
    
    func stopMonitoring() {
        if let monitor = mouseMonitor {
            NSEvent.removeMonitor(monitor)
            mouseMonitor = nil
        }
        lastClickTime = nil
        lastClickLocation = nil
    }
}
