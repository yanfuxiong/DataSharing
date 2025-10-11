//
//  main.swift
//  CrossShareHelper
//
//  Created by ts on 2025/8/15.
//  Updated by TS on 2025/8/27 for XPC support
//

import Foundation
import Cocoa


// MARK: - 程序入口

// 创建 XPC Service 委托
class ServiceDelegate: NSObject, NSXPCListenerDelegate {
    
    private let logger = XPCLogger.shared
    private let xpcService = CrossShareXPCService()
    
    func listener(_ listener: NSXPCListener, shouldAcceptNewConnection newConnection: NSXPCConnection) -> Bool {
        logger.info("New XPC connection request from: \(newConnection)")
        
        // 配置连接
        newConnection.exportedInterface = NSXPCInterface(with: CrossShareXPCProtocol.self)
        newConnection.exportedObject = xpcService
        
        // 设置客户端接口（用于回调）
        newConnection.remoteObjectInterface = NSXPCInterface(with: CrossShareXPCDelegate.self)
        
        // 设置连接处理程序
        newConnection.invalidationHandler = { [weak self] in
            self?.logger.info("XPC connection invalidated")
        }
        
        newConnection.interruptionHandler = { [weak self] in
            self?.logger.info("XPC connection interrupted")
        }
        
        // 启动连接
        newConnection.resume()
        
        logger.info("XPC connection established successfully")
        return true
    }
}

// XPC Service 主入口
let delegate = ServiceDelegate()
let listener = NSXPCListener.service()
listener.delegate = delegate
listener.resume()

