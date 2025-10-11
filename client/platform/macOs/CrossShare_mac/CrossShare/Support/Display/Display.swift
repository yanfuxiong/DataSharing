//  Copyright © MonitorControl. @JoniVR, @theOneyouseek, @waydabber and others

import Cocoa
import Foundation
import os.log

class Display: Equatable {
    let identifier: CGDirectDisplayID
    let prefsId: String
    var name: String
    var vendorNumber: UInt32?
    var modelNumber: UInt32?
    var serialNumber: UInt32?
    var smoothBrightnessTransient: Float = 1
    var smoothBrightnessRunning: Bool = false
    var smoothBrightnessSlow: Bool = false
    let swBrightnessSemaphore = DispatchSemaphore(value: 1)
    
    static func == (lhs: Display, rhs: Display) -> Bool {
        lhs.identifier == rhs.identifier
    }
    
    var brightnessSyncSourceValue: Float = 1
    var isVirtual: Bool = false
    var isDummy: Bool = false
    
    var defaultGammaTableRed = [CGGammaValue](repeating: 0, count: 256)
    var defaultGammaTableGreen = [CGGammaValue](repeating: 0, count: 256)
    var defaultGammaTableBlue = [CGGammaValue](repeating: 0, count: 256)
    var defaultGammaTableSampleCount: UInt32 = 0
    var defaultGammaTablePeak: Float = 1
    
    init(_ identifier: CGDirectDisplayID, name: String, vendorNumber: UInt32?, modelNumber: UInt32?, serialNumber: UInt32?, isVirtual: Bool = false, isDummy: Bool = false) {
        self.identifier = identifier
        self.name = name
        self.vendorNumber = vendorNumber
        self.modelNumber = modelNumber
        self.serialNumber = serialNumber
        self.isVirtual = isVirtual
        self.isDummy = isDummy
        self.prefsId = "(\(name.filter { !$0.isWhitespace })\(vendorNumber ?? 0)\(modelNumber ?? 0)@\(self.isVirtual ? (self.serialNumber ?? 9999) : identifier))"
        if self.isVirtual , !self.isDummy {
            os_log("Creating or updating shade for display %{public}@", type: .info, String(self.identifier))
            _ = DisplayManager.shared.updateShade(displayID: self.identifier)
        } else {
            os_log("Destroying shade (if exists) for display %{public}@", type: .info, String(self.identifier))
            _ = DisplayManager.shared.destroyShade(displayID: self.identifier)
        }
    }
    
    func isBuiltIn() -> Bool {
        if CGDisplayIsBuiltin(self.identifier) != 0 {
            return true
        } else {
            return false
        }
    }
}
