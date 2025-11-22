//
//  CSLicenseManager.swift
//  CrossShare
//
//  Created by Assistant on 2025/10/20.
//

import Foundation

class CSLicenseManager {
    
    // MARK: - Singleton
    
    static let shared = CSLicenseManager()
    
    private init() {}
    
    // MARK: - License Data Structure
    
    struct LibraryLicense {
        let name: String
        let licenseText: String
    }
    
    // MARK: - Available Libraries
    
    /// Get all supported third-party libraries list
    func getAllLibraries() -> [String] {
        return [
            "RealmSwift",
            "SnapKit",
            "Alamofire"
        ]
    }
    
    /// Get the License text for the specified library
    func getLicenseText(for libraryName: String) -> String {
        switch libraryName {
        case "RealmSwift":
            return loadLicenseFromFile(fileName: "realm-swift.txt")
        case "SnapKit":
            return loadLicenseFromFile(fileName: "snapkit.txt")
        case "Alamofire":
            return loadLicenseFromFile(fileName: "Alamofire.txt")
        default:
            return "License information not available."
        }
    }
    
    // MARK: - Private Helper Methods
    
    /// Read license file from Resources/license directory
    private func loadLicenseFromFile(fileName: String) -> String {
        // Try multiple path search methods
        var filePath: String?
        
        // Method 1: Try license subdirectory
        let fileNameWithoutExt = fileName.replacingOccurrences(of: ".txt", with: "")
        filePath = Bundle.main.path(forResource: "license/\(fileNameWithoutExt)", ofType: "txt")
        
        // Method 2: Try direct filename lookup
        if filePath == nil {
            filePath = Bundle.main.path(forResource: fileNameWithoutExt, ofType: "txt", inDirectory: "license")
        }
        
        // Method 3: Try searching in Resources directory
        if filePath == nil {
            filePath = Bundle.main.path(forResource: "Resources/license/\(fileNameWithoutExt)", ofType: "txt")
        }
        
        // Method 4: Find all txt files and match
        if filePath == nil {
            let paths = Bundle.main.paths(forResourcesOfType: "txt", inDirectory: nil)
            if !paths.isEmpty {
                filePath = paths.first { $0.contains(fileNameWithoutExt) }
                if let foundPath = filePath {
                    print("CSLicenseManager: Found file at: \(foundPath)")
                }
            }
        }
        
        guard let validPath = filePath else {
            print("CSLicenseManager: Cannot find license file: \(fileName)")
            print("CSLicenseManager: Searched in Bundle.main.resourcePath: \(Bundle.main.resourcePath ?? "nil")")
            
            // List all available txt files for debugging
            let txtFiles = Bundle.main.paths(forResourcesOfType: "txt", inDirectory: nil)
            if !txtFiles.isEmpty {
                print("CSLicenseManager: Available txt files in bundle:")
                txtFiles.forEach { print("  - \($0)") }
            } else {
                print("CSLicenseManager: No txt files found in bundle")
            }
            
            return "License file not found: \(fileName)\nPlease ensure the file is added to the Xcode project target."
        }
        
        do {
            let content = try String(contentsOfFile: validPath, encoding: .utf8)
            print("CSLicenseManager: Successfully loaded license from: \(validPath)")
            return content
        } catch {
            print("CSLicenseManager: Error reading license file \(fileName): \(error)")
            return "Error loading license: \(error.localizedDescription)"
        }
    }
}

// Reference for the source of the License file:
// - RealmSwift: https://github.com/realm/realm-swift/blob/master/LICENSE (Apache 2.0)
// - SnapKit: https://github.com/SnapKit/SnapKit/blob/develop/LICENSE (MIT)

