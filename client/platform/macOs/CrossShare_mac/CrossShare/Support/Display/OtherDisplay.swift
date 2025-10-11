//  Copyright © MonitorControl. @JoniVR, @theOneyouseek, @waydabber and others

import Cocoa
import IOKit
import os.log

class OtherDisplay: Display {
    var ddc: IntelDDC?
    var arm64ddc: Bool = false
    var arm64avService: IOAVService?
    var isDiscouraged: Bool = false
    let writeDDCQueue = DispatchQueue(label: "Local write DDC queue")
    var writeDDCNextValue: [Command: UInt16] = [:]
    var writeDDCLastSavedValue: [Command: UInt16] = [:]
    
    override init(_ identifier: CGDirectDisplayID, name: String, vendorNumber: UInt32?, modelNumber: UInt32?, serialNumber: UInt32?, isVirtual: Bool = false, isDummy: Bool = false) {
        super.init(identifier, name: name, vendorNumber: vendorNumber, modelNumber: modelNumber, serialNumber: serialNumber, isVirtual: isVirtual, isDummy: isDummy)
        if !isVirtual, !Arm64DDC.isArm64 {
            self.ddc = IntelDDC(for: identifier)
        }
        
    }
    
    let swAfterOsdAnimationSemaphore = DispatchSemaphore(value: 1)
    var lastAnimationStartedTime: CFTimeInterval = CACurrentMediaTime()
    
    public func writeDDCValues(command: Command, value: UInt16) {
        self.writeDDCQueue.async(flags: .barrier) {
            self.writeDDCNextValue[command] = value
        }
        DisplayManager.shared.globalDDCQueue.async(flags: .barrier) {
            self.asyncPerformWriteDDCValues(command: command)
        }
    }
    
    func asyncPerformWriteDDCValues(command: Command) {
        var value = UInt16.max
        var lastValue = UInt16.max
        self.writeDDCQueue.sync {
            value = self.writeDDCNextValue[command] ?? UInt16.max
            lastValue = self.writeDDCLastSavedValue[command] ?? UInt16.max
        }
        guard value != UInt16.max, value != lastValue else {
            return
        }
        self.writeDDCQueue.async(flags: .barrier) {
            self.writeDDCLastSavedValue[command] = value
        }
        var controlCodes = [UInt8]()
        controlCodes.append(command.rawValue)
        for controlCode in controlCodes {
            if Arm64DDC.isArm64 {
                if self.arm64ddc {
                    _ = Arm64DDC.write(service: self.arm64avService, command: controlCode, value: value)
                }
            } else {
                _ = self.ddc?.write(command: controlCode, value: value, errorRecoveryWaitTime: 2000) ?? false
            }
        }
    }
    
    func readDDCValues(for command: Command, tries: UInt, minReplyDelay delay: UInt64?) -> (current: UInt16, max: UInt16)? {
        var values: (UInt16, UInt16)?
        let controlCode = command.rawValue
        if Arm64DDC.isArm64 {
            guard self.arm64ddc else {
                return nil
            }
            DisplayManager.shared.globalDDCQueue.sync {
                if let unwrappedDelay = delay {
                    values = Arm64DDC.read(service: self.arm64avService, command: controlCode, readSleepTime: UInt32(unwrappedDelay / 1000), numOfRetryAttemps: UInt8(min(tries, 255)))
                } else {
                    values = Arm64DDC.read(service: self.arm64avService, command: controlCode, numOfRetryAttemps: UInt8(min(tries, 255)))
                }
            }
        } else {
            DisplayManager.shared.globalDDCQueue.sync {
                values = self.ddc?.read(command: controlCode, tries: tries, minReplyDelay: delay)
            }
        }
        return values
    }
}
