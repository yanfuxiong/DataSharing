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
                    // Future migration logic can be added here
                }
            )
            Realm.Configuration.defaultConfiguration = config
            
            // Initialize Realm instance
            realm = try Realm()
            print("Realm database path: \(realm?.configuration.fileURL?.absoluteString ?? "Unknown")")
        } catch {
            print("Failed to initialize Realm: \(error.localizedDescription)")
        }
    }

    // Save CSFileInfo to Realm database
    func saveCSFileInfoToRealm(_ csFileInfo: CSFileInfo) {
        do {
            guard let realm = realm else {
                print("Realm is not initialized")
                return
            }
            
            try realm.write {
                // Check if a record with the same sessionId already exists
                if let existingInfo = realm.object(ofType: RealmCSFileInfo.self, forPrimaryKey: csFileInfo.sessionId) {
                    // Update existing record
                    existingInfo.senderID = csFileInfo.senderID
                    existingInfo.isCompleted = csFileInfo.isCompleted
                    existingInfo.progress = csFileInfo.progress
                    existingInfo.session = RealmFileTransfer(from: csFileInfo.session)
                    existingInfo.updatedAt = Date()
                } else {
                    // Create new record
                    let realmInfo = RealmCSFileInfo(from: csFileInfo)
                    realm.add(realmInfo)
                }
            }
//            print("CSFileInfo saved to Realm: \(csFileInfo.sessionId)")
        } catch {
            print("Failed to save CSFileInfo to Realm: \(error.localizedDescription)")
        }
    }
    
    // Load all CSFileInfo records from Realm database
    func loadCSFileInfosFromRealm() -> [CSFileInfo] {
        guard let realm = realm else {
            print("Realm is not initialized")
            return []
        }
        
        let realmInfos = realm.objects(RealmCSFileInfo.self)
        return realmInfos.map { $0.toCSFileInfo() }
    }
    
    // Delete specific transfer record from database and table data source
    func deleteFileTransferRecord(with sessionId: String, from bottomTableData: [CSFileInfo]) -> [CSFileInfo] {
        print("Deleting file transfer record with sessionId: \(sessionId)")
        
        // Delete record from Realm database
        do {
            guard let realm = realm else {
                print("Realm is not initialized")
                return bottomTableData
            }
            
            try realm.write {
                if let recordToDelete = realm.object(ofType: RealmCSFileInfo.self, forPrimaryKey: sessionId) {
                    realm.delete(recordToDelete)
                    print("CSFileInfo record deleted from Realm: \(sessionId)")
                } else {
                    print("CSFileInfo record not found in Realm: \(sessionId)")
                }
            }
        } catch {
            print("Failed to delete CSFileInfo record from Realm: \(error.localizedDescription)")
        }
        
        // Delete corresponding record from bottomTableData
        if let index = bottomTableData.firstIndex(where: { $0.sessionId == sessionId }) {
            var updatedData = bottomTableData
            updatedData.remove(at: index)
            print("CSFileInfo record deleted from bottomTableData: \(sessionId)")
            return updatedData
        } else {
            print("CSFileInfo record not found in bottomTableData: \(sessionId)")
            return bottomTableData
        }
    }
    
    
    func deleteAllData(){
        // Delete all records from Realm database
        do {
            guard let realm = realm else {
                print("Realm is not initialized")
                return
            }
            
            try realm.write {
                let allCSFileInfos = realm.objects(RealmCSFileInfo.self)
                realm.delete(allCSFileInfos)
            }
            print("All CSFileInfo records deleted from Realm")
        } catch {
            print("Failed to delete CSFileInfo records from Realm: \(error.localizedDescription)")
        }

    }

}
