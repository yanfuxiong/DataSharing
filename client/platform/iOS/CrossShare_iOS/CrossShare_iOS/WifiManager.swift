//
//  WifiManager.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/4/15.
//

import UIKit
import SystemConfiguration.CaptiveNetwork
import SystemConfiguration
import Network

class WifiManager: NSObject {
    
    static private let shared = WifiManager();
    
    public static func shareInstance() -> WifiManager {
        return shared
    }
    
    func getWiFiIPAddress() -> String? {
        var address: String?
        var ifaddr: UnsafeMutablePointer<ifaddrs>?
        if getifaddrs(&ifaddr) == 0 {
            var ptr = ifaddr
            while ptr != nil {
                if let interface = ptr?.pointee {
                    let addrFamily = interface.ifa_addr.pointee.sa_family
                    if addrFamily == UInt8(AF_INET) {
                        if let name = interface.ifa_name, String(cString: name) == networkInterface {
                            let sockaddr_inPointer = interface.ifa_addr.withMemoryRebound(to: sockaddr_in.self, capacity: 1) { pointer in
                                return pointer.pointee
                            }
                            let ip = String(cString: inet_ntoa(sockaddr_inPointer.sin_addr))
                            address = ip
                            break
                        }
                    }
                }
                ptr = ptr?.pointee.ifa_next
            }
            freeifaddrs(ifaddr)
        }
        return address
    }
    

    func getNetInfoFromLocalIp(completion: @escaping (_ netname: String?, _ index: UInt32?) -> Void) {
        DispatchQueue.global(qos: .background).async {
            var ifaddr: UnsafeMutablePointer<ifaddrs>? = nil
            var resultName: String?
            var resultIndex: UInt32?
            if getifaddrs(&ifaddr) == 0 {
                var ptr = ifaddr
                while ptr != nil {
                    guard let interface = ptr?.pointee else {
                        ptr = ptr?.pointee.ifa_next
                        continue
                    }
                    let name = String(cString: interface.ifa_name)
                    // 检查是否为 Wi-Fi 接口，一般是 en0
                    if name.hasPrefix(networkInterface) {
                        let index = if_nametoindex(interface.ifa_name)
                        resultName = name
                        resultIndex = index
                        break
                    }
                    ptr = interface.ifa_next
                }
                freeifaddrs(ifaddr)
            }
            DispatchQueue.main.async {
                completion(resultName, resultIndex)
            }
        }
    }
    
    func getAvailablePort() -> Int32? {
        while true {
            let port = Int32.random(in: 1024...65535)
            if isPortAvailable(port: port) {
                return port
            }
        }
    }
    
    func isPortAvailable(port: Int32) -> Bool {
        do {
            let listener = try NWListener(using: .tcp, on: NWEndpoint.Port(rawValue: UInt16(port))!)
            listener.cancel()
            return true
        } catch {
            return false
        }
    }
}
