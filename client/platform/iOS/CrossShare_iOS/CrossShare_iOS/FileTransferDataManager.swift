//
//  FileTransferDataManager.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/7/16.
//

import UIKit

protocol FileTransferDataObserver: AnyObject {
    func dataDidUpdate(_ data: [DownloadItem])
}

private class WeakObserver {
    weak var observer: FileTransferDataObserver?
    
    init(observer: FileTransferDataObserver) {
        self.observer = observer
    }
}

class FileTransferDataManager {
    static let shared = FileTransferDataManager()
    
    private var dataArray: [DownloadItem] = []
    private var observers: [WeakObserver] = []
    
    public var isNewDataTransfering: Bool  = false
    private init() {
        addNotifications()
    }
    
    private func addNotifications() {
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(receiveNewFiles(_:)),
            name: ReceiveFuleSuccessNotification,
            object: nil
        )
    }
    
    @objc private func receiveNewFiles(_ ntf: Notification) {
        DispatchQueue.main.async { [weak self] in
            guard let self = self else { return }

            if let userInfo = ntf.userInfo as? [String: Any],
               let newItem = userInfo["download"] as? DownloadItem {
                
                self.updateDataArray(with: newItem)
                self.notifyObservers()
            }
        }
    }
    
    private func updateDataArray(with newItem: DownloadItem) {
        if let index = dataArray.firstIndex(where: { $0.uuid == newItem.uuid }) {
            isNewDataTransfering = false
            dataArray[index].receiveSize = newItem.receiveSize
            dataArray[index].totalSize = newItem.totalSize
            dataArray[index].finishTime = newItem.finishTime
            dataArray[index].timestamp = newItem.timestamp
            if let recvFileCnt = newItem.recvFileCnt, let currentfileName = newItem.currentfileName {
                dataArray[index].recvFileCnt = recvFileCnt
                dataArray[index].currentfileName = currentfileName
            }
        } else {
            isNewDataTransfering = true
            dataArray.append(newItem)
            let fileCnt = (newItem.totalFileCnt ?? 0)
            let fileMsg = fileCnt > 1 ? "\(fileCnt) files" : "\(fileCnt) file"
            let deviceName = newItem.deviceName ?? ""
            let params: [String] = [fileMsg, deviceName]
            PushNotiManager.shared.sendLocalNotification(code: .receiveStart, with: params)
        }
        dataArray.sort { ($0.timestamp ?? 0) > ($1.timestamp ?? 0) }
    }
    
    func addObserver(_ observer: FileTransferDataObserver) {
        observers.append(WeakObserver(observer: observer))
        cleanupObservers()
    }
    
    func removeObserver(_ observer: FileTransferDataObserver) {
        observers.removeAll { $0.observer === observer }
    }
    
    private func notifyObservers() {
        cleanupObservers()
        observers.forEach { $0.observer?.dataDidUpdate(dataArray) }
    }
    
    private func cleanupObservers() {
        observers.removeAll { $0.observer == nil }
    }
    
    func getCurrentData() -> [DownloadItem] {
        return dataArray
    }
}
