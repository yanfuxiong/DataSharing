//
//  DownloadItem.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/5/7.
//

import UIKit
import SwiftyJSON

/*
 char* ip,char* id, char* deviceName, char* currentfileName,unsigned int recvFileCnt, unsigned int totalFileCnt,unsigned long long currentFileSize,unsigned long long totalSize,unsigned long long recvSize,unsigned long long timestamp
 */

/*
 多文件进度条API
 invokeCallbackUpdateMultipleProgressBar(
 char* ip,    //发送端IP地址 包含port
 char* id,   //发送端ID
 char* deviceName,  //发送端设备名称
 char* currentfileName, //当前正在接收的文件名
 unsigned int recvFileCnt, //当前已经接收完毕的文件个数
 unsigned int totalFileCnt, //总的文件数量
 unsigned long long currentFileSize, // 当前正在接收的文件size
 unsigned long long totalSize,   //当前一笔多档的所有文档的总的size
 unsigned long long recvSize,    //当前已经接收的总size
 unsigned long long timestamp)   //当前一笔多档传输的时间戳，用于标识一笔传输（后续有多笔排队传输的场景），可以当做类似ID标识使用
 */

class DownloadItem: NSObject {
    
    var ip: String = ""
    
    var deviceName:String?
    
    var fileId: String = ""
   
    var currentfileName:String?
    
    var uuid:String = UUID().uuidString
    
    var currentFileSize:UInt64?
    
    var receiveSize:UInt64?
    
    var totalSize:UInt64?
    
    var progress:Float?
    
    var finishTime:TimeInterval?
    
    var recvFileCnt:UInt32?
    
    var totalFileCnt:UInt32?
    
    var timestamp:TimeInterval?
    
    var isMutip:Bool = false
    
    var error:String?
}
