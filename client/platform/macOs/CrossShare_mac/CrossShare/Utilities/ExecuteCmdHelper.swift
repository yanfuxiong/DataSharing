//
//  ExecuteCmdHelper.swift
//  CrossShare
//
//  Created by user00 on 2025/3/6.
//

import Cocoa

class ExecuteCmdHelper: NSObject {
    
    func executeCmdAndGetResult(_ cmd: String) -> String {
        // 初始化 Process
        let process = Process()
        process.launchPath = "/bin/bash"
        process.arguments = ["-c", cmd]
        
        // 设置输出管道
        let pipe = Pipe()
        process.standardOutput = pipe
        
        // 启动任务
        do {
            try process.run()
        } catch {
            logger.info("executeCmdAndGetResult: Task failed to launch.")
            return ""
        }
        
        // 读取输出数据
        let fileHandle = pipe.fileHandleForReading
        let data = fileHandle.readDataToEndOfFile()
        process.waitUntilExit() // 等待任务完成
        
        let status = process.terminationStatus
        if status == 0 {
            logger.info("executeCmdAndGetResult: Task succeeded.")
        } else {
            logger.info("executeCmdAndGetResult: Task failed.")
        }
        
        // 关闭文件
        fileHandle.closeFile()
        
        // 将 Data 转换为 String
        return String(data: data, encoding: .utf8) ?? ""
    }
    
    
}
