//
//  XPCProtocol.swift
//  CrossShare
//
//  Created by TS on 2025/8/27.
//  XPC 通信协议定义
//

import Foundation

// MARK: - XPC Service Protocol

/// CrossShare XPC 服务协议
/// 定义主应用与 Helper 之间的通信接口
@objc protocol CrossShareXPCProtocol {
    
    // MARK: - 服务生命周期
    
    /// 启动 Go 服务
    /// - Parameters:
    ///   - config: 服务配置参数
    ///   - completion: 启动结果回调
    func startGoService(config: [String: Any], completion: @escaping (Bool, String?) -> Void)
    
    /// 停止 Go 服务
    /// - Parameter completion: 停止结果回调
    func stopGoService(completion: @escaping (Bool) -> Void)
    
    /// 获取服务状态
    /// - Parameter completion: 状态回调 (isRunning, serviceInfo)
    func getServiceStatus(completion: @escaping (Bool, [String: Any]?) -> Void)
    
    // MARK: - 设备发现与管理
    
    /// 开始设备发现
    /// - Parameter completion: 发现结果回调
    func startDeviceDiscovery(completion: @escaping (Bool) -> Void)
    
    /// 停止设备发现
    /// - Parameter completion: 停止结果回调
    func stopDeviceDiscovery(completion: @escaping (Bool) -> Void)
    
    /// 获取已发现的设备列表
    /// - Parameter completion: 设备列表回调
    func getDiscoveredDevices(completion: @escaping ([[String: Any]]) -> Void)
    
    /// 连接到指定设备
    /// - Parameters:
    ///   - deviceId: 设备ID
    ///   - completion: 连接结果回调
    func connectToDevice(deviceId: String, completion: @escaping (Bool, String?) -> Void)
    
    /// 断开设备连接
    /// - Parameters:
    ///   - deviceId: 设备ID
    ///   - completion: 断开结果回调
    func disconnectFromDevice(deviceId: String, completion: @escaping (Bool) -> Void)
    
    // MARK: - 文件传输
    
    /// 发送文件到设备
    /// - Parameters:
    ///   - filePath: 文件路径
    ///   - deviceId: 目标设备ID
    ///   - completion: 传输结果回调
    func sendFile(filePath: String, toDevice deviceId: String, completion: @escaping (Bool, String?) -> Void)
    
    /// 发送文件夹到设备
    /// - Parameters:
    ///   - folderPath: 文件夹路径
    ///   - deviceId: 目标设备ID
    ///   - completion: 传输结果回调
    func sendFolder(folderPath: String, toDevice deviceId: String, completion: @escaping (Bool, String?) -> Void)
    
    /// 取消文件传输
    /// - Parameters:
    ///   - transferId: 传输ID
    ///   - completion: 取消结果回调
    func cancelTransfer(transferId: String, completion: @escaping (Bool) -> Void)
    
    /// 获取传输进度
    /// - Parameter completion: 进度信息回调
    func getTransferProgress(completion: @escaping ([[String: Any]]) -> Void)
    
    // MARK: - 剪贴板同步
    
    /// 启动剪贴板监控
    /// - Parameter completion: 启动结果回调
    func startClipboardMonitoring(completion: @escaping (Bool) -> Void)
    
    /// 停止剪贴板监控
    /// - Parameter completion: 停止结果回调
    func stopClipboardMonitoring(completion: @escaping (Bool) -> Void)
    
    /// 同步剪贴板内容到设备
    /// - Parameters:
    ///   - deviceId: 目标设备ID
    ///   - completion: 同步结果回调
    func syncClipboard(toDevice deviceId: String, completion: @escaping (Bool) -> Void)
    
    // MARK: - 网络状态
    
    /// 获取网络配置信息
    /// - Parameter completion: 网络信息回调
    func getNetworkInfo(completion: @escaping ([String: Any]) -> Void)
    
    /// 获取本地IP地址
    /// - Parameter completion: IP地址回调
    func getLocalIPAddress(completion: @escaping (String?) -> Void)
    
    /// 检查端口可用性
    /// - Parameters:
    ///   - port: 端口号
    ///   - completion: 可用性回调
    func checkPortAvailability(port: Int, completion: @escaping (Bool) -> Void)
    
    // MARK: - 配置管理
    
    /// 更新服务配置
    /// - Parameters:
    ///   - config: 新的配置参数
    ///   - completion: 更新结果回调
    func updateConfiguration(config: [String: Any], completion: @escaping (Bool, String?) -> Void)
    
    /// 获取当前配置
    /// - Parameter completion: 配置信息回调
    func getCurrentConfiguration(completion: @escaping ([String: Any]) -> Void)
    
    // MARK: - 日志与调试
    
    /// 获取服务日志
    /// - Parameters:
    ///   - lines: 获取的行数
    ///   - completion: 日志内容回调
    func getServiceLogs(lines: Int, completion: @escaping ([String]) -> Void)
    
    /// 设置日志级别
    /// - Parameters:
    ///   - level: 日志级别 (0-4: Debug, Info, Warn, Error, Fatal)
    ///   - completion: 设置结果回调
    func setLogLevel(level: Int, completion: @escaping (Bool) -> Void)
    
    /// 健康检查
    /// - Parameter completion: 健康状态回调
    func healthCheck(completion: @escaping (Bool, [String: Any]) -> Void)
}

// MARK: - XPC Service Delegate Protocol

/// XPC 服务委托协议
/// 用于 Helper 向主应用推送事件和状态更新
@objc protocol CrossShareXPCDelegate {
    
    // MARK: - 设备事件
    
    /// 发现新设备
    /// - Parameter device: 设备信息
    func didDiscoverDevice(device: [String: Any])
    
    /// 设备离线
    /// - Parameter deviceId: 设备ID
    func didLoseDevice(deviceId: String)
    
    /// 设备连接状态变化
    /// - Parameters:
    ///   - deviceId: 设备ID
    ///   - connected: 连接状态
    func didChangeDeviceConnection(deviceId: String, connected: Bool)
    
    // MARK: - 传输事件
    
    /// 开始接收文件
    /// - Parameter transferInfo: 传输信息
    func didStartReceivingFile(transferInfo: [String: Any])
    
    /// 文件传输进度更新
    /// - Parameter progress: 进度信息
    func didUpdateTransferProgress(progress: [String: Any])
    
    /// 文件传输完成
    /// - Parameter result: 传输结果
    func didCompleteTransfer(result: [String: Any])
    
    /// 文件传输失败
    /// - Parameter error: 错误信息
    func didFailTransfer(error: [String: Any])
    
    // MARK: - 剪贴板事件
    
    /// 接收到剪贴板内容
    /// - Parameter clipboardData: 剪贴板数据
    func didReceiveClipboardData(clipboardData: [String: Any])
    
    // MARK: - 服务状态事件
    
    /// 服务状态变化
    /// - Parameters:
    ///   - isRunning: 运行状态
    ///   - info: 状态信息
    func didChangeServiceStatus(isRunning: Bool, info: [String: Any]?)
    
    /// 网络状态变化
    /// - Parameter networkInfo: 网络信息
    func didChangeNetworkStatus(networkInfo: [String: Any])
    
    /// 发生错误
    /// - Parameter error: 错误信息
    func didEncounterError(error: [String: Any])
    
    /// 服务日志输出
    /// - Parameter logEntry: 日志条目
    func didOutputLog(logEntry: [String: Any])
}

// 注意：数据模型定义已移至 SharedModels.swift
