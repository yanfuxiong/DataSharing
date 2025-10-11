//
//  UniversalDDC.swift
//  CrossShare
//
//  Created by mac on 2025/9/18.
//

import Foundation

class UniversalDDC{
    private var intelddc: IntelDDC?
    var arm64avService: IOAVService?
    
    func initIntelIdentifier(identifier: CGDirectDisplayID){
       if !Arm64DDC.isArm64 {
         self.intelddc = IntelDDC(for: identifier)
       }
    }
    
    func initArm64Service(arm64avService:IOAVService?){
        if(Arm64DDC.isArm64 && arm64avService != nil){
            self.arm64avService = arm64avService
        }
    }

    func readDDCValues(command: UInt8) -> (current: UInt16, max: UInt16)? {
        var result: (current: UInt16, max: UInt16)?
        if(Arm64DDC.isArm64){
            if(self.arm64avService != nil){
                //numOfRetryAttemps: The maximum number of retries when an operation fails.
                //retrySleepTime: The interval time between two retries. Avoid frequent retries to prevent excessive device load.
                DisplayManager.shared.globalDDCQueue.sync {
                    result = Arm64DDC.read(
                        service: self.arm64avService,
                        command: command,
                        numOfRetryAttemps: 5,
                        retrySleepTime: 50000)
                }
                return result
            }else{
                print("you need init arm64avService first")
                return result
            }
        }else{
            if(self.intelddc != nil){
                //tries: Maximum retry attempts
                //errorRecoveryWaitTime: Error recovery delay after read failure, used to alleviate hardware stress
                DisplayManager.shared.globalDDCQueue.sync {
                    result = self.intelddc?.read(command: command, tries: 5, errorRecoveryWaitTime: 50000)
                }
                return result
            }else{
                print("you need init intelddc first")
                return result
            }
        }
    }
    
    func writeDDCValues(command: UInt8, value: UInt16) -> Bool {
        var result = false
        if(Arm64DDC.isArm64){
            if(self.arm64avService != nil){
                result =  Arm64DDC.write(
                    service: self.arm64avService,
                    command: command,
                    value: value,
                    numOfRetryAttemps: 5,
                    retrySleepTime: 50000
                )
                return result
            }else{
                print("you need init arm64avService first")
                return result
            }
        }else{
            if(self.intelddc != nil){
                result = self.intelddc?.write(command: command, value: value, errorRecoveryWaitTime: 50000,numofWriteCycles: 5) ?? false
                return result
            }else{
                print("you need init intelddc first")
                return result
            }
        }

    }

    
    
    

    
}
