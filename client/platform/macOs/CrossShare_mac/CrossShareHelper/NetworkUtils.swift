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
    
    // MARK: - UDP Port Management
    
    /// Check if a UDP port is available by attempting to bind
    func checkUDPPortAvailability(port: UInt16) -> Bool {
        let sock = Darwin.socket(AF_INET, SOCK_DGRAM, 0)
        guard sock != -1 else { return false }
        defer { Darwin.close(sock) }
        
        var addr = sockaddr_in()
        addr.sin_len = UInt8(MemoryLayout<sockaddr_in>.size)
        addr.sin_family = sa_family_t(AF_INET)
        addr.sin_addr.s_addr = INADDR_ANY.bigEndian
        addr.sin_port = port.bigEndian
        
        let result = withUnsafePointer(to: &addr) {
            $0.withMemoryRebound(to: sockaddr.self, capacity: 1) {
                bind(sock, $0, socklen_t(MemoryLayout<sockaddr_in>.size))
            }
        }
        
        return result == 0
    }
    
    /// Find an available UDP port starting from the given port
    func findAvailableUDPPort(startPort: UInt16 = 10000, endPort: UInt16 = 65000) -> UInt16? {
        for port in startPort...endPort {
            if checkUDPPortAvailability(port: port) {
                return port
            }
        }
        return nil
    }
    
    /// Find a pair of available UDP ports (for mouse and keyboard)
    /// Returns two consecutive available ports
    func findAvailableUDPPortPair(startPort: UInt16 = 10000) -> (mousePort: UInt16, keyboardPort: UInt16)? {
        var port = startPort
        while port < 65000 {
            if checkUDPPortAvailability(port: port) && checkUDPPortAvailability(port: port + 1) {
                return (mousePort: port, keyboardPort: port + 1)
            }
            port += 2
        }
        return nil
    }
}
