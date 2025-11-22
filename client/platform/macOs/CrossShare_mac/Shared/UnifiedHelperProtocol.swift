//
//  UnifiedHelperProtocol.swift
//  Shared between CrossShare and CrossShareHelper
//
//  Created by TS on 2025/9/11.
//  Unified Helper XPC Protocol Definition
//

import Foundation

@objc public protocol CrossShareHelperXPCProtocol {
    
    func initializeGoService(config: [String: Any], completion: @escaping (Bool, String?) -> Void)
    func getServiceStatus(completion: @escaping (Bool, [String: Any]?) -> Void)
    func updateCount(_ count: Int, completion: @escaping (Int) -> Void)
    func setDIASID(_ diasID: String, completion: @escaping (Bool, String?) -> Void)
    func setExtractDIAS(completion: @escaping (Bool, String?) -> Void)
    func updateDisplayMapping(mac: String, displayID: UInt32, completion: @escaping (Bool) -> Void)
    func rescanDisplays(completion: @escaping (Bool) -> Void)
    func sendTextToRemote(_ text: String, completion: @escaping (Bool, String?) -> Void)
    func sendImageToRemote(_ imageData: Data, completion: @escaping (Bool, String?) -> Void)
    func startClipboardMonitoring(completion: @escaping (Bool) -> Void)
    func stopClipboardMonitoring(completion: @escaping (Bool) -> Void)
    func getDeviceList(completion: @escaping ([[String: Any]]) -> Void)
    func sendMultiFilesDropRequest(multiFilesData: String, completion: @escaping (Bool, String?) -> Void)
    func setCancelFileTransfer(ipPort: String, clientID: String, timeStamp: UInt64, completion: @escaping (Bool, String?) -> Void)
    func setDragFileListRequest(multiFilesData: String, timestamp: UInt64, width: UInt16, height: UInt16, posX: Int16, posY: Int16, completion: @escaping (Bool, String?) -> Void)
    func requestUpdateDownloadPath(downloadPath: String, completion: @escaping (Bool, String?) -> Void)
}

@objc public protocol CrossShareHelperXPCDelegate: AnyObject {
    @objc optional func didUpdateCount(_ newCount: Int)
    @objc optional func didDetectPluginEvent(eventType: String, eventData: [String: Any]?)
    @objc optional func didSetDiasIDEvent(success: Bool, message: String?)
    @objc optional func didSetExtractDIAS(success: Bool, message: String?)
    @objc optional func didReceiveAuthRequest(index: UInt32)
    @objc optional func didReceiveDeviceData(deviceData: [String: Any])
    @objc optional func didReceiveRemoteClipboard(text: String?, imageData: Data?, html: String?)
    @objc optional func didDetectLocalClipboardChange(content: [String: Any])
    @objc optional func didDetectScreenCountChange(change: String, currentCount: Int, previousCount: Int)
    @objc optional func didReceiveFileTransferUpdate(_ sessionInfo: [String: Any])
    //
    @objc optional func didReceiveFilesData(_ userInfo: [String: Any])
    @objc optional func didReceiveTransferFilesDataUpdate(_ userInfo: [String: Any])
    @objc optional func didReceiveDIASStatus(_ status: Int)
    @objc optional func didReceiveErrorEvent(_ errorInfo: [String: Any])
    @objc optional func didReceiveSystemInfoUpdate(_ systemInfo: [String: Any])
}
