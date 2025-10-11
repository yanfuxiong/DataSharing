//  Copyright © MonitorControl. @JoniVR, @theOneyouseek, @waydabber and others

import Foundation
import os.log

class AppleDisplay: Display {
    private var displayQueue: DispatchQueue
    
    override init(_ identifier: CGDirectDisplayID, name: String, vendorNumber: UInt32?, modelNumber: UInt32?, serialNumber: UInt32?, isVirtual: Bool = false, isDummy: Bool = false) {
        self.displayQueue = DispatchQueue(label: String("displayQueue-\(identifier)"))
        super.init(identifier, name: name, vendorNumber: vendorNumber, modelNumber: modelNumber, serialNumber: serialNumber, isVirtual: isVirtual, isDummy: isDummy)
    }
    
    public func getAppleBrightness() -> Float {
        guard !self.isDummy else {
            return 1
        }
        var brightness: Float = 0
        DisplayServicesGetBrightness(self.identifier, &brightness)
        return brightness
    }
    
    public func setAppleBrightness(value: Float) {
        guard !self.isDummy else {
            return
        }
        _ = self.displayQueue.sync {
            DisplayServicesSetBrightness(self.identifier, value)
        }
    }
    
    func setDirectBrightness(_ to: Float, transient: Bool = false) -> Bool {
        guard !self.isDummy else {
            return false
        }
        let value = max(min(to, 1), 0)
        self.setAppleBrightness(value: value)
        if !transient {
            self.brightnessSyncSourceValue = value
            self.smoothBrightnessTransient = value
        }
        return true
    }
}
