//
//  RealmDataManager.swift
//  CrossShare
//
//  Created by TS on 2025/10/9.
//

import Foundation
import RealmSwift

class RealmDataManager {
    static let shared = RealmDataManager()
    private var realm: Realm?

    private init() {
        
    }
    
    func setupRealm() {
        do {
            // Configure Realm
            let config = Realm.Configuration(
                schemaVersion: RealmSchemaVersion.currentVersion,
                migrationBlock: {
                    migration, oldSchemaVersion in
                    // Upgrade from version 1 to version 2: Add errCode field
                    if oldSchemaVersion < 2 {
                        // New field errCode will be added automatically with default value nil
                        logger.info("Migrating Realm from version \(oldSchemaVersion) to 2")
                    }
                }
            )
            Realm.Configuration.defaultConfiguration = config
            
            // Initialize Realm instance
            realm = try Realm()
            logger.info("Realm database path: \(realm?.configuration.fileURL?.absoluteString ?? "Unknown")")
        } catch {
            logger.info("Failed to initialize Realm: \(error.localizedDescription)")
        }
    }

    // Save CSFileInfo to Realm database
    func saveCSFileInfoToRealm(_ csFileInfo: CSFileInfo) {
        do {
            guard let realm = realm else {
                logger.info("Realm is not initialized")
                return
            }
            
            try realm.write {
                // Check if a record with the same sessionId already exists
                if let existingInfo = realm.object(ofType: RealmCSFileInfo.self, forPrimaryKey: csFileInfo.sessionId) {
                    // Update existing record
                    existingInfo.senderID = csFileInfo.senderID
                    existingInfo.isCompleted = csFileInfo.isCompleted
                    existingInfo.progress = csFileInfo.progress
                    existingInfo.errCode = csFileInfo.errCode
                    existingInfo.session = RealmFileTransfer(from: csFileInfo.session)
                    existingInfo.updatedAt = Date()
                } else {
                    // Create new record
                    let realmInfo = RealmCSFileInfo(from: csFileInfo)
                    realm.add(realmInfo)
                }
            }
//            logger.info("CSFileInfo saved to Realm: \(csFileInfo.sessionId)")
        } catch {
            logger.info("Failed to save CSFileInfo to Realm: \(error.localizedDescription)")
        }
    }
    
    // Load all CSFileInfo records from Realm database
    func loadCSFileInfosFromRealm() -> [CSFileInfo] {
        guard let realm = realm else {
            logger.info("Realm is not initialized")
            return []
        }
        
        let realmInfos = realm.objects(RealmCSFileInfo.self)
        return realmInfos.map { $0.toCSFileInfo() }
    }
    
    // Delete specific transfer record from database and table data source
    func deleteFileTransferRecord(with sessionId: String, from bottomTableData: [CSFileInfo]) -> [CSFileInfo] {
        logger.info("Deleting file transfer record with sessionId: \(sessionId)")
        
        // Delete record from Realm database
        do {
            guard let realm = realm else {
                logger.info("Realm is not initialized")
                return bottomTableData
            }
            
            try realm.write {
                if let recordToDelete = realm.object(ofType: RealmCSFileInfo.self, forPrimaryKey: sessionId) {
                    realm.delete(recordToDelete)
                    logger.info("CSFileInfo record deleted from Realm: \(sessionId)")
                } else {
                    logger.info("CSFileInfo record not found in Realm: \(sessionId)")
                }
            }
        } catch {
            logger.info("Failed to delete CSFileInfo record from Realm: \(error.localizedDescription)")
        }
        
        // Delete corresponding record from bottomTableData
        if let index = bottomTableData.firstIndex(where: { $0.sessionId == sessionId }) {
            var updatedData = bottomTableData
            updatedData.remove(at: index)
            logger.info("CSFileInfo record deleted from bottomTableData: \(sessionId)")
            return updatedData
        } else {
            logger.info("CSFileInfo record not found in bottomTableData: \(sessionId)")
            return bottomTableData
        }
    }
    
    
    func deleteAllData(){
        // Delete all records from Realm database
        do {
            guard let realm = realm else {
                logger.info("Realm is not initialized")
                return
            }
            
            try realm.write {
                let allCSFileInfos = realm.objects(RealmCSFileInfo.self)
                realm.delete(allCSFileInfos)
            }
            logger.info("All CSFileInfo records deleted from Realm")
        } catch {
            logger.info("Failed to delete CSFileInfo records from Realm: \(error.localizedDescription)")
        }

    }
    
    // Update CSFileInfo and return UI operations via callback
    func updateCSFileInfo(_ csFileInfo: CSFileInfo, bottomTableData: [CSFileInfo], completion: @escaping (Result<(updatedData: [CSFileInfo], isNewRecord: Bool, index: Int?), Error>) -> Void) {
        do {
            // Save data to Realm
            guard let realm = realm else {
                completion(.failure(NSError(domain: "RealmDataManager", code: 0, userInfo: [NSLocalizedDescriptionKey: "Realm is not initialized"])))
                return
            }
            
            try realm.write {
                // Check if a record with the same sessionId already exists
                if let existingInfo = realm.object(ofType: RealmCSFileInfo.self, forPrimaryKey: csFileInfo.sessionId) {
                    // Update existing record
                    existingInfo.senderID = csFileInfo.senderID
                    existingInfo.isCompleted = csFileInfo.isCompleted
                    existingInfo.progress = csFileInfo.progress
                    existingInfo.errCode = csFileInfo.errCode
                    existingInfo.session = RealmFileTransfer(from: csFileInfo.session)
                    existingInfo.updatedAt = Date()
                } else {
                    // Create new record
                    let realmInfo = RealmCSFileInfo(from: csFileInfo)
                    realm.add(realmInfo)
                }
            }
            
            // Handle bottomTableData
            if let index = bottomTableData.firstIndex(where: { $0.sessionId == csFileInfo.sessionId }) {
                // Update existing record
                var updatedData = bottomTableData
                updatedData[index] = csFileInfo
                completion(.success((updatedData, false, index)))
            } else {
                // Add new record to the beginning of array (newest first)
                var updatedData = bottomTableData
                updatedData.insert(csFileInfo, at: 0)
                completion(.success((updatedData, true, 0)))
            }
        } catch {
            completion(.failure(error))
        }
    }

}
