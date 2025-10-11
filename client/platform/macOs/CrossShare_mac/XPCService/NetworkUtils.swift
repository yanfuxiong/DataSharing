//
//  NetworkUtils.swift
//  XPCService
//
//  Created by TS on 2025/9/8.
//  Network utility functions
//

import Foundation

class NetworkUtils {
    
    static let shared = NetworkUtils()
    
    private init() {}
    
    func getLocalIPAddress(completion: @escaping (String?) -> Void) {
        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            let result = self?.callGoFunction {
                return GetLocalIPAddress()
            }
            
            DispatchQueue.main.async {
                if let success = result?.success, success {
                    completion(result?.data)
                } else {
                    completion(nil)
                }
            }
        }
    }
    
    func checkPortAvailability(port: Int, completion: @escaping (Bool) -> Void) {
        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            let result = self?.callGoFunction {
                return CheckPortAvailability(Int32(port))
            }
            
            DispatchQueue.main.async {
                completion(result?.success ?? false)
            }
        }
    }
    
    private func callGoFunction(_ goCall: () -> UnsafeMutablePointer<GoResult>?) -> NetworkGoResult? {
        guard let cResult = goCall() else { return nil }
        let swiftResult = NetworkGoResult(cResult: UnsafePointer(cResult))
        if let data = cResult.pointee.data {
            free(data)
        }
        if let errorMessage = cResult.pointee.error_message {
            free(errorMessage)
        }
        free(cResult)
        return swiftResult
    }
}

private struct NetworkGoResult {
    let success: Bool
    let data: String?
    let errorMessage: String?
    
    init(cResult: UnsafePointer<GoResult>) {
        self.success = cResult.pointee.success != 0
        
        if let dataPtr = cResult.pointee.data {
            self.data = String(cString: dataPtr)
        } else {
            self.data = nil
        }
        
        if let errorPtr = cResult.pointee.error_message {
            self.errorMessage = String(cString: errorPtr)
        } else {
            self.errorMessage = nil
        }
    }
}
