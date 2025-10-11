//
//  WifiManager.swift
//  CrossShareHelper
//
//  Created by TS on 2025/9/12.
//  Network interface manager for Helper
//

import Cocoa
import SystemConfiguration.CaptiveNetwork
import SystemConfiguration
import Darwin

private let SIOCGIFMTU = 0xc0206933 as UInt

private struct ifreq {
    var ifr_name: (CChar, CChar, CChar, CChar, CChar, CChar, CChar, CChar, CChar, CChar, CChar, CChar, CChar, CChar, CChar, CChar) = (0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
    var ifr_ifru_mtu: Int32 = 0
}

class WifiManager: NSObject {
    
    static private let shared = WifiManager();
    
    private let networkInterface = "en0"
    
    public static func shareInstance() -> WifiManager {
        return shared
    }
    
    func getNetInfoFromLocalIp(completion: @escaping (_ netname: String?, _ mac: String?, _ mtu: Int32?, _ index: UInt32?, _ flags: UInt32?) -> Void) {
        DispatchQueue.global(qos: .background).async { [self] in
            var ifaddr: UnsafeMutablePointer<ifaddrs>? = nil
            var resultName: String?
            var resultMAC: String?
            var resultMTU: Int32?
            var resultIndex: UInt32?
            var resultFlags: UInt32?
            
            if getifaddrs(&ifaddr) == 0 {
                var ptr = ifaddr
                while ptr != nil {
                    guard let interface = ptr?.pointee else {
                        ptr = ptr?.pointee.ifa_next
                        continue
                    }
                    let name = String(cString: interface.ifa_name)
                    if name.hasPrefix(self.networkInterface) {
                        let index = if_nametoindex(interface.ifa_name)
                        resultName = name
                        resultIndex = index
                        
                        resultFlags = UInt32(interface.ifa_flags)
                        
                        let sock = socket(AF_INET, SOCK_DGRAM, 0)
                        if sock >= 0 {
                            var ifr = ifreq()
                            withUnsafeMutableBytes(of: &ifr.ifr_name) { ptr in
                                name.withCString { cStr in
                                    _ = memcpy(ptr.baseAddress!, cStr, min(name.count, 15))
                                }
                            }
                        
                            if ioctl(sock, SIOCGIFMTU, &ifr) == 0 {
                                resultMTU = ifr.ifr_ifru_mtu
                            }
                            
                            close(sock)
                        }
                        
                        var macPtr = ifaddr
                        while macPtr != nil {
                            let macInterface = macPtr!.pointee
                            let macName = String(cString: macInterface.ifa_name)
                            if macName == name && macInterface.ifa_addr?.pointee.sa_family == UInt8(AF_LINK) {
                                let sdl = macInterface.ifa_addr.withMemoryRebound(to: sockaddr_dl.self, capacity: 1) { $0 }
                                let sdlData = sdl.pointee
                                let macStartIndex = Int(sdlData.sdl_nlen)
                                let macLength = Int(sdlData.sdl_alen)
                                
                                if macLength == 6 {
                                    withUnsafeBytes(of: sdlData.sdl_data) { bytes in
                                        let macBytes = bytes.bindMemory(to: UInt8.self)
                                        var macArray: [String] = []
                                        for i in macStartIndex..<(macStartIndex + macLength) {
                                            if i < macBytes.count {
                                                macArray.append(String(format: "%02x", macBytes[i]))
                                            }
                                        }
                                        if macArray.count == 6 {
                                            resultMAC = macArray.joined(separator: ":")
                                        }
                                    }
                                }
                                break
                            }
                            macPtr = macPtr!.pointee.ifa_next
                        }
                        break
                    }
                    ptr = interface.ifa_next
                }
                freeifaddrs(ifaddr)
            }
            DispatchQueue.main.async {
                completion(resultName, resultMAC, resultMTU, resultIndex, resultFlags)
            }
        }
    }
}