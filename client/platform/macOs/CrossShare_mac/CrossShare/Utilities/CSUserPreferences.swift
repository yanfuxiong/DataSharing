//
//  CSUserPreferences.swift
//  CrossShare
//
//  User preferences manager for storing app settings
//

import Foundation

class CSUserPreferences {
    static let shared = CSUserPreferences()
    
    private let defaults = UserDefaults.standard
    
    // Keys for UserDefaults
    private enum Keys {
        static let downloadPath = "com.realtek.crossshare.downloadPath"
    }
    
    private init() {}
    
    // MARK: - Download Path
    
    /// Save the download path to user preferences
    func saveDownloadPath(_ path: String) {
        defaults.set(path, forKey: Keys.downloadPath)
        defaults.synchronize()
        print("CSUserPreferences: Saved download path: \(path)")
    }
    
    /// Get the saved download path, returns nil if not set
    func getDownloadPath() -> String? {
        let path = defaults.string(forKey: Keys.downloadPath)
        if let path = path {
            print("CSUserPreferences: Retrieved download path: \(path)")
        } else {
            print("CSUserPreferences: No download path saved")
        }
        return path
    }
    
    /// Clear the saved download path
    func clearDownloadPath() {
        defaults.removeObject(forKey: Keys.downloadPath)
        defaults.synchronize()
        print("CSUserPreferences: Cleared download path")
    }
    
    /// Get the download path or default path if not set
    func getDownloadPathOrDefault() -> String {
        if let savedPath = getDownloadPath() {
            // Verify the saved path still exists
            if FileManager.default.fileExists(atPath: savedPath) {
                return savedPath
            } else {
                print("CSUserPreferences: Saved path no longer exists, using default")
                clearDownloadPath()
            }
        }

        return getDefDownloadPath()
    }
}

