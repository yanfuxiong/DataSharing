//
//  FileSystemInfoFetcher.swift
//  CrossShare
//
//  Created by mac on 2025/9/16.
//

import Foundation
import AppKit
import UniformTypeIdentifiers

struct FileInfo {
    let name: String          // Name (filename/foldername)
    let path: String          // Full path
    let isDirectory: Bool     // Whether it's a folder
    let fileSize: Int64?      // File size (only valid for files, nil for folders)
    let modificationDate: String? // Modification time (formatted string)
    let fileType: String?     // File type (human-readable, e.g., "PNG image", "Folder" for folders)
}

class FileSystemInfoFetcher {
    /// Check if the given path exists and whether it's a file or folder
    static func checkPathType(_ path: String) -> (exists: Bool, isDirectory: Bool) {
        let fileManager = FileManager.default
        var isDirectory: ObjCBool = false
        
        let exists = fileManager.fileExists(atPath: path, isDirectory: &isDirectory)
        return (exists, isDirectory.boolValue)
    }
    
    /// Get information for all items in the specified folder
    static func getFolderContentsInfo(in folderPath: String) -> [FileInfo]? {
        let fileManager = FileManager.default
        let url = URL(fileURLWithPath: folderPath)
        
        let pathCheck = checkPathType(folderPath)
        guard pathCheck.exists else {
            print("Path does not exist: \(folderPath)")
            return nil
        }
        
        guard pathCheck.isDirectory else {
            print("Specified path is not a folder: \(folderPath)")
            return nil
        }
        
        do {
            let contents = try fileManager.contentsOfDirectory(
                at: url,
                includingPropertiesForKeys: [.isDirectoryKey, .fileSizeKey, .contentModificationDateKey, .typeIdentifierKey],
                options: [.skipsHiddenFiles]
            )
            
            var itemsInfo: [FileInfo] = []
            
            for itemURL in contents {
                let resourceValues = try itemURL.resourceValues(forKeys: [
                    .nameKey,
                    .pathKey,
                    .isDirectoryKey,
                    .fileSizeKey,
                    .contentModificationDateKey,
                    .typeIdentifierKey
                ])
                
                let name = resourceValues.name ?? itemURL.lastPathComponent
                let path = resourceValues.path ?? itemURL.path
                let isDirectory = resourceValues.isDirectory ?? false
                // Fix: Convert Int? to Int64?
                let fileSize = isDirectory ? nil : resourceValues.fileSize.flatMap { Int64($0) }
                // Core modification: Convert Date to formatted string
                // Use flatMap to handle possible nil (return nil if modification time cannot be obtained)
                let modificationDate = resourceValues.contentModificationDate.flatMap { formatDate($0) }
                // New: File type (fixed as "Folder" for folders; use UTI localized description for files)
                let uti = resourceValues.typeIdentifier
                let localizedType = uti.flatMap { UTType($0)?.localizedDescription }
                let fileType = isDirectory ? "Folder" : (localizedType ?? (itemURL.pathExtension.isEmpty ? "File" : itemURL.pathExtension.uppercased() + " File"))

                itemsInfo.append(FileInfo(
                    name: name,
                    path: path,
                    isDirectory: isDirectory,
                    fileSize: fileSize,
                    modificationDate: modificationDate,
                    fileType: fileType
                ))
            }
            
            return itemsInfo
        } catch {
            print("Failed to retrieve folder contents: \(error.localizedDescription)")
            return nil
        }
    }
    
    /// Helper method: Format Date into string (yyyy-MM-dd HH:mm)
    private static func formatDate(_ date: Date) -> String {
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy-MM-dd HH:mm"  // Year-Month-Day Hour:Minute
        formatter.timeZone = TimeZone.current      // Use current time zone
        return formatter.string(from: date)
    }
    
    /// Get detailed information for a single file/folder by path
    static func getItemInfo(for path: String) -> FileInfo? {
        _ = FileManager.default
        let url = URL(fileURLWithPath: path)
        
        let pathCheck = checkPathType(path)
        guard pathCheck.exists else {
            print("Error: Path does not exist -> \(path)")
            return nil
        }
        
        do {
            let resourceValues = try url.resourceValues(forKeys: [
                .nameKey,
                .pathKey,
                .isDirectoryKey,
                .fileSizeKey,
                .contentModificationDateKey,
                .typeIdentifierKey
            ])
            
            let name = resourceValues.name ?? url.lastPathComponent
            let standardPath = resourceValues.path ?? path
            let isDirectory = resourceValues.isDirectory ?? false
            // Fix: Convert Int? to Int64?
            let fileSize = isDirectory ? nil : resourceValues.fileSize.flatMap { Int64($0) }
            // Core modification: Convert Date to formatted string
            // Use flatMap to handle possible nil (return nil if modification time cannot be obtained)
            let modificationDate = resourceValues.contentModificationDate.flatMap { formatDate($0) }
            // New: File type
            let uti = resourceValues.typeIdentifier
            let localizedType = uti.flatMap { UTType($0)?.localizedDescription }
            let fileType = isDirectory ? "Folder" : (localizedType ?? (url.pathExtension.isEmpty ? "File" : url.pathExtension.uppercased() + " File"))

            return FileInfo(
                name: name,
                path: standardPath,
                isDirectory: isDirectory,
                fileSize: fileSize,
                modificationDate: modificationDate,
                fileType: fileType
            )
        } catch {
            print("Failed to get item info for \(path): \(error.localizedDescription)")
            return nil
        }
    }
}
