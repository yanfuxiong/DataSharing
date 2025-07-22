//
//  BonjourService.swift
//  CrossShare_iOS
//
//  Created by huang on 5/5/25.
//

import Foundation

class BonjourService: NSObject, NetServiceBrowserDelegate, NetServiceDelegate {
    static let shared = BonjourService()
    private var mBrowser: NetServiceBrowser?
    private var mDiscoveredService: [NetService] = []
    private var mIsSearching: Bool = false
    private var mFilterInstance: String? = nil
    private let syncQueue = DispatchQueue(label: "BonjourService")

    override init() {
        super.init()
        mBrowser = NetServiceBrowser()
        mBrowser?.delegate = self
    }

    private func reset() {
        mIsSearching = false
        mFilterInstance = nil
        mDiscoveredService.removeAll()
    }

    func start(instanceName: String, serviceType: String) {
        syncQueue.async { [self] in
            Logger.info("[Bonjour] Start search type: \(serviceType)")
            if mIsSearching {
                stop()
            } else {
                mIsSearching = true
            }
            mFilterInstance = instanceName
            DispatchQueue.global(qos: .background).async { [self] in
                mBrowser?.searchForServices(ofType: serviceType, inDomain: "local.")
            }
        }
    }

    func stop() {
        syncQueue.async { [self] in
            Logger.info("[Bonjour] Stop search")
            if !mIsSearching {
                return
            }
            mBrowser?.stop()
            reset()
        }
    }

    func foundServices(instanceName: String, ip: String, port: Int) {
        if mFilterInstance == nil {
            return
        }

        if mFilterInstance != "" && mFilterInstance != instanceName {
            Logger.info("[Bonjour] Found: \(instanceName) and skip by \(mFilterInstance ?? "null")")
            return
        }

        syncQueue.async { [self] in
            SetBrowseMdnsResult(instanceName.toGoString(), ip.toGoString(), GoInt(port))
            if mFilterInstance != "" {
                stop()
            }
        }
    }

    func netServiceBrowser(_ browser: NetServiceBrowser, didFind service: NetService, moreComing: Bool) {
        service.delegate = self

        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy-MM-dd HH:mm:ss.SSS"
        let timestamp = formatter.string(from: Date())
        Logger.info("[Bonjour][\(timestamp)] Found service: [Instance: \(service.name), Type: \(service.type), Domain: \(service.domain)]")
        mDiscoveredService.append(service)
        service.resolve(withTimeout: 5)
    }

    func netServiceDidResolveAddress(_ sender: NetService) {
        guard let addrAry = sender.addresses, let txtData = sender.txtRecordData() else {
            return
        }

        guard let ipData = NetService.dictionary(fromTXTRecord: txtData)["ip"],
              let ipFromTextData = String(data: ipData, encoding: .utf8) else {
            Logger.info("[Bonjour][Err] Empty IP by textData in \(sender.name)")
            return
        }

        for addrData in addrAry {
            addrData.withUnsafeBytes { (buffer: UnsafeRawBufferPointer) in
                guard let addr = buffer.bindMemory(to: sockaddr.self).baseAddress else {
                    return
                }

                if addr.pointee.sa_family == sa_family_t(AF_INET) {
                    var ip = [CChar](repeating: 0, count: Int(INET_ADDRSTRLEN))
                    var addr_in = unsafeBitCast(addr, to: UnsafePointer<sockaddr_in>.self).pointee
                    inet_ntop(AF_INET, &addr_in.sin_addr, &ip, socklen_t(INET_ADDRSTRLEN))
                    let ipStr = String(cString: ip)

                    if !(ipFromTextData.isEmpty) && ipStr != ipFromTextData {
                        Logger.info("[Bonjour][Err] Skip IP:(\(ipStr)) by textRecord:(\(ipFromTextData))")
                        return
                    }
                    let formatter = DateFormatter()
                    formatter.dateFormat = "yyyy-MM-dd HH:mm:ss.SSS"
                    let timestamp = formatter.string(from: Date())
                    Logger.info("[Bonjour][\(timestamp)] Resolve instance: \(sender.name) address: \(ipStr), port: \(sender.port)")
                    foundServices(instanceName: sender.name, ip: ipStr, port: sender.port)
                } else {
                    // DEBUG: for IPv6
//                            Logger.info("[Bonjour][Error]: Unavailable IP address type. Only support IPv4 now")
                }
            }
        }
    }

    func netService(_ sender: NetService, didNotResolve errorDict: [String : NSNumber]) {
        Logger.info("[Bonjour][Error]: Service not resolve: \(errorDict)")
    }
}
