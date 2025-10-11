//
//  NetworkUtils.swift
//  CrossShareHelper
//
//  Created by TS on 2025/9/11.
//  Network utility functions for CrossShareHelper
//

import Foundation

class NetworkUtils {
    
    static let shared = NetworkUtils()
    
    private init() {}
    
    func getLocalIPAddress(completion: @escaping (String?) -> Void) {
        var address: String?
        var ifaddr: UnsafeMutablePointer<ifaddrs>?
        
        guard getifaddrs(&ifaddr) == 0 else {
            completion(nil)
            return
        }
        guard let firstAddr = ifaddr else {
            completion(nil)
            return
        }
        
        for ifptr in sequence(first: firstAddr, next: { $0.pointee.ifa_next }) {
            let interface = ifptr.pointee
            let addrFamily = interface.ifa_addr.pointee.sa_family
            if addrFamily == UInt8(AF_INET) || addrFamily == UInt8(AF_INET6) {
                let name = String(cString: interface.ifa_name)
                if name == "en0" || name == "en1" {
                    var hostname = [CChar](repeating: 0, count: Int(NI_MAXHOST))
                    getnameinfo(interface.ifa_addr, socklen_t(interface.ifa_addr.pointee.sa_len),
                                &hostname, socklen_t(hostname.count),
                                nil, socklen_t(0), NI_NUMERICHOST)
                    address = String(cString: hostname)
                    if addrFamily == UInt8(AF_INET) {
                        break
                    }
                }
            }
        }
        
        freeifaddrs(ifaddr)
        completion(address)
    }
    
    func checkPortAvailability(port: Int, completion: @escaping (Bool) -> Void) {
        let socket = Darwin.socket(AF_INET, SOCK_STREAM, 0)
        guard socket != -1 else {
            completion(false)
            return
        }
        
        defer { close(socket) }
        
        var addr = sockaddr_in()
        addr.sin_family = sa_family_t(AF_INET)
        addr.sin_addr.s_addr = inet_addr("127.0.0.1")
        addr.sin_port = in_port_t(port).bigEndian
        
        let result = withUnsafePointer(to: &addr) {
            $0.withMemoryRebound(to: sockaddr.self, capacity: 1) {
                bind(socket, $0, socklen_t(MemoryLayout<sockaddr_in>.size))
            }
        }
        
        completion(result == 0)
    }
}
