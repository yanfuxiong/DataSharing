//
//  P2PConfig.swift
//  CrossShare
//
//  Created by Assistant on 2025/9/11.
//  Shared P2P service configuration
//

import Foundation

struct P2PConfig {
    let deviceName: String
    let rootPath: String
    let downloadPath: String
    let serverId: String
    let serverIpInfo: String
    let listenHost: String
    let listenPort: Int32
    
    func toXPCDict() -> [String: Any] {
        return [
            "deviceName": deviceName,
            "rootPath": rootPath,
            "downloadPath":downloadPath,
            "serverId": serverId,
            "serverIpInfo": serverIpInfo,
            "listenHost": listenHost,
            "listenPort": listenPort
        ]
    }
    
    static func defaultConfig() -> P2PConfig {
        // Get saved path from user preferences, or use default
        let rootPath = getRootPath()
        let downloadPath = CSUserPreferences.shared.getDownloadPathOrDefault()
        let fileManager = FileManager.default
        if !fileManager.fileExists(atPath: rootPath) {
            do {
                try fileManager.createDirectory(atPath: rootPath, withIntermediateDirectories: true, attributes: nil)
                print("P2PConfig: Created rootPath directory: \(rootPath)")
            } catch {
                print("P2PConfig: Failed to create rootPath directory: \(error)")
            }
        }
        
        return P2PConfig(
            deviceName: Host.current().localizedName ?? "Mac",
            rootPath: rootPath,
            downloadPath: downloadPath,
            serverId: getLocalIPAddress() ?? "192.168.1.1",
            serverIpInfo: UUID().uuidString,
            listenHost: getLocalIPAddress() ?? "192.168.1.1",
            listenPort: getAvailablePort() ?? Int32(8080)
        )
    }
    
    private static func getLocalIPAddress() -> String? {
        var address: String?
        var ifaddr: UnsafeMutablePointer<ifaddrs>?
        
        if getifaddrs(&ifaddr) == 0 {
            var ptr = ifaddr
            while ptr != nil {
                defer { ptr = ptr?.pointee.ifa_next }
                
                guard let interface = ptr?.pointee else { continue }
                let addrFamily = interface.ifa_addr.pointee.sa_family
                if addrFamily == UInt8(AF_INET) {
                    let name = String(cString: interface.ifa_name)
                    if name == "en0" || name == "en1" {
                        var hostname = [CChar](repeating: 0, count: Int(NI_MAXHOST))
                        getnameinfo(interface.ifa_addr, socklen_t(interface.ifa_addr.pointee.sa_len),
                                    &hostname, socklen_t(hostname.count), nil, socklen_t(0), NI_NUMERICHOST)
                        address = String(cString: hostname)
                        break
                    }
                }
            }
            freeifaddrs(ifaddr)
        }
        return address
    }
    
    private static func getAvailablePort() -> Int32? {
        var attempts = 0
        while attempts < 100 {
            let port = Int32.random(in: 1024...65535)
            if isPortAvailable(port: port) {
                return port
            }
            attempts += 1
        }
        return nil
    }
    
    private static func isPortAvailable(port: Int32) -> Bool {
        let socketFD = socket(AF_INET, SOCK_STREAM, 0)
        if socketFD == -1 {
            return false
        }
        
        defer {
            close(socketFD)
        }
        
        var addr = sockaddr_in()
        addr.sin_family = sa_family_t(AF_INET)
        addr.sin_port = UInt16(port).bigEndian
        addr.sin_addr.s_addr = INADDR_ANY
        
        let bindResult = withUnsafePointer(to: &addr) {
            $0.withMemoryRebound(to: sockaddr.self, capacity: 1) { pointer in
                Darwin.bind(socketFD, pointer, socklen_t(MemoryLayout<sockaddr_in>.size))
            }
        }
        
        return bindResult == 0
    }
}
