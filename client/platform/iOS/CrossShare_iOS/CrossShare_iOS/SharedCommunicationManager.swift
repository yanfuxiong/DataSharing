//
//  SharedCommunicationManager.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/7/21.
//

import UIKit
import Foundation

class SharedCommunicationManager: NSObject {
    static let shared = SharedCommunicationManager()
    
    private let groupId = UserDefaults.groupId
    private var sharedDefaults: UserDefaults?
    private let communicationKey = "ShareExtensionCommunication"
    
    struct CommunicationData: Codable {
        let id: String
        let timestamp: Double
        let type: CommunicationType
        let payload: [String: Any]
        
        enum CommunicationType: String, Codable {
            case transferPrepare = "transfer_prepare"
            case transferStart = "transfer_start"
            case transferProgress = "transfer_progress"
            case transferComplete = "transfer_complete"
            case transferError = "transfer_error"
            case transferCancel = "transfer_cancel"
        }
        
        init(type: CommunicationType, payload: [String: Any]) {
            self.id = UUID().uuidString
            self.timestamp = Date().timeIntervalSince1970
            self.type = type
            self.payload = payload
        }
        
        enum CodingKeys: String, CodingKey {
            case id, timestamp, type, payload
        }
        
        func encode(to encoder: Encoder) throws {
            var container = encoder.container(keyedBy: CodingKeys.self)
            try container.encode(id, forKey: .id)
            try container.encode(timestamp, forKey: .timestamp)
            try container.encode(type, forKey: .type)
            
            let payloadData = try JSONSerialization.data(withJSONObject: payload)
            try container.encode(payloadData, forKey: .payload)
        }
        
        init(from decoder: Decoder) throws {
            let container = try decoder.container(keyedBy: CodingKeys.self)
            id = try container.decode(String.self, forKey: .id)
            timestamp = try container.decode(Double.self, forKey: .timestamp)
            type = try container.decode(CommunicationType.self, forKey: .type)
            
            let payloadData = try container.decode(Data.self, forKey: .payload)
            payload = try JSONSerialization.jsonObject(with: payloadData) as? [String: Any] ?? [:]
        }
    }
    
    var onTransferPrepare: (([String: Any]) -> Void)?
    var onTransferStart: (([String: Any]) -> Void)?
    var onTransferProgress: (([String: Any]) -> Void)?
    var onTransferComplete: (([String: Any]) -> Void)?
    var onTransferError: (([String: Any]) -> Void)?
    var onTransferCancel: (([String: Any]) -> Void)?
    
    private override init() {
        super.init()
        setupSharedDefaults()
    }
    
    private func setupSharedDefaults() {
        sharedDefaults = UserDefaults(suiteName: groupId)
        guard sharedDefaults != nil else {
            Logger.error("SharedCommunicationManager: Cannot access shared UserDefaults")
            return
        }
        
        startObserving()
    }
    
    private func startObserving() {
        sharedDefaults?.addObserver(
            self,
            forKeyPath: communicationKey,
            options: [.new, .old],
            context: nil
        )
//        Logger.info("SharedCommunicationManager: Started observing UserDefaults changes")
    }
    
    func stopObserving() {
        sharedDefaults?.removeObserver(self, forKeyPath: communicationKey)
//        Logger.info("SharedCommunicationManager: Stopped observing UserDefaults changes")
    }
    
    override func observeValue(forKeyPath keyPath: String?, of object: Any?, change: [NSKeyValueChangeKey : Any]?, context: UnsafeMutableRawPointer?) {
        
        guard keyPath == communicationKey else { return }
        
        guard let newValue = change?[.newKey] as? String,
              !newValue.isEmpty,
              let data = newValue.data(using: .utf8) else {
            return
        }
        
        do {
            let communicationData = try JSONDecoder().decode(CommunicationData.self, from: data)
            
            Logger.info("SharedCommunicationManager: Received communication - Type: \(communicationData.type.rawValue), ID: \(communicationData.id)")
            
            DispatchQueue.main.async { [weak self] in
                self?.handleCommunicationData(communicationData)
            }
            
        } catch {
            Logger.error("SharedCommunicationManager: Failed to decode communication data: \(error)")
        }
    }
    
    private func handleCommunicationData(_ data: CommunicationData) {
        switch data.type {
        case .transferPrepare:
            onTransferPrepare?(data.payload)
        case .transferStart:
            onTransferStart?(data.payload)
        case .transferProgress:
            onTransferProgress?(data.payload)
        case .transferComplete:
            onTransferComplete?(data.payload)
        case .transferError:
            onTransferError?(data.payload)
        case .transferCancel:
            onTransferCancel?(data.payload)
        }
    }
    
    func sendCommunication(type: CommunicationData.CommunicationType, payload: [String: Any]) {
        let communicationData = CommunicationData(type: type, payload: payload)
        
        do {
            let data = try JSONEncoder().encode(communicationData)
            let jsonString = String(data: data, encoding: .utf8) ?? ""
            
            sharedDefaults?.set(jsonString, forKey: communicationKey)
            sharedDefaults?.synchronize()
            
            Logger.info("SharedCommunicationManager: Sent communication - Type: \(type.rawValue), ID: \(communicationData.id)")
            
        } catch {
            Logger.error("SharedCommunicationManager: Failed to encode communication data: \(error)")
        }
    }
    
    deinit {
        stopObserving()
    }
}
