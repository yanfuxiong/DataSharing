//
//  SharedDataManager.swift
//  CrossShare
//
//  Created for shared data management between processes
//

import Foundation

// MARK: - Configuration Models
struct CSConfig: Codable {
    var uiTheme: UITheme
    var windowSettings: WindowSettings?
    
    enum CodingKeys: String, CodingKey {
        case uiTheme = "UITheme"
        case windowSettings = "WindowSettings"
    }
}

struct UITheme: Codable {
    var isInited: String
    var customerID: String
    
    enum CodingKeys: String, CodingKey {
        case isInited
        case customerID
    }
    
    var isInitedBool: Bool {
        return isInited.lowercased() == "true"
    }
}

struct WindowSettings: Codable {
    var width: Double
    var height: Double
    var fileBrowserHeightRatio: Double  // FileBrowserView height ratio (0.0 to 1.0)
    
    enum CodingKeys: String, CodingKey {
        case width
        case height
        case fileBrowserHeightRatio
    }
}

// MARK: - SharedDataManager
class SharedDataManager {
    static let shared = SharedDataManager()
    
    // Config file name
    private let configFileName = "csconfig.json"
    
    private init() {
        // Ensure config directory exists and copy default config if needed
        setupConfigFile()
    }
    
    // MARK: - Config File Path
    
    /// Get the Application Support directory for CrossShare
    private func getAppSupportDirectory() -> URL? {
        guard let appSupport = FileManager.default.urls(for: .applicationSupportDirectory, in: .userDomainMask).first else {
            return nil
        }
        
        let crossShareDir = appSupport.appendingPathComponent("CrossShare", isDirectory: true)
        
        // Create directory if it doesn't exist
        if !FileManager.default.fileExists(atPath: crossShareDir.path) {
            try? FileManager.default.createDirectory(at: crossShareDir, withIntermediateDirectories: true)
            print("Created config directory: \(crossShareDir.path)")
        }
        
        return crossShareDir
    }
    
    /// Get the config file URL in Application Support
    private func getConfigFileURL() -> URL? {
        return getAppSupportDirectory()?.appendingPathComponent(configFileName)
    }
    
    /// Setup config file (create default config in Application Support if not exists)
    private func setupConfigFile() {
        guard let configURL = getConfigFileURL() else {
            print("Failed to get config file URL")
            return
        }
        
        // Check if config file exists in Application Support
        if !FileManager.default.fileExists(atPath: configURL.path) {
            // Create default config
            print("First run: Creating default config in Application Support")
            let defaultConfig = createDefaultConfig()
            saveConfig(defaultConfig)
            print("Default config created at: \(configURL.path)")
        } else {
            print("Using existing config at: \(configURL.path)")
        }
    }
    
    // MARK: - Config Management
    
    /// Load config from file
    func loadConfig() -> CSConfig? {
        guard let fileURL = getConfigFileURL() else {
            print("Failed to get config file URL")
            return nil
        }
        
        do {
            let data = try Data(contentsOf: fileURL)
            let decoder = JSONDecoder()
            let config = try decoder.decode(CSConfig.self, from: data)
            // print("Config loaded successfully")
            // print("Path: \(fileURL.path)")
            // print("isInited: \(config.uiTheme.isInited)")
            // print("customerID: \(config.uiTheme.customerID)")
            return config
        } catch {
            print("Failed to load config: \(error)")
            return nil
        }
    }
    
    /// Save config to file
    func saveConfig(_ config: CSConfig) {
        guard let fileURL = getConfigFileURL() else {
            print("Failed to get config file URL")
            return
        }
        
        do {
            let encoder = JSONEncoder()
            encoder.outputFormatting = [.prettyPrinted, .sortedKeys]
            let data = try encoder.encode(config)
            try data.write(to: fileURL, options: .atomic)
            print("Config saved successfully")
            print("Path: \(fileURL.path)")
            print("isInited: \(config.uiTheme.isInited)")
            print("customerID: \(config.uiTheme.customerID)")
        } catch {
            print("Failed to save config: \(error)")
        }
    }
    
    /// Create default config
    func createDefaultConfig() -> CSConfig {
        return CSConfig(
            uiTheme: UITheme(isInited: "false", customerID: "-1"),
            windowSettings: nil
        )
    }
    
    // MARK: - Window Settings Management
    
    /// Save window settings
    func saveWindowSettings(width: Double, height: Double, fileBrowserHeightRatio: Double) {
        var config = loadConfig() ?? createDefaultConfig()
        config.windowSettings = WindowSettings(
            width: width,
            height: height,
            fileBrowserHeightRatio: fileBrowserHeightRatio
        )
        saveConfig(config)
        logger.debug("Window settings saved - Width: \(width), Height: \(height), FileBrowserRatio: \(fileBrowserHeightRatio)")
    }
    
    /// Get window settings
    func getWindowSettings() -> WindowSettings? {
        let settings = loadConfig()?.windowSettings
        if let settings = settings {
            logger.debug("Window settings loaded - Width: \(settings.width), Height: \(settings.height), FileBrowserRatio: \(settings.fileBrowserHeightRatio)")
        } else {
            logger.debug("No window settings found, will use defaults")
        }
        return settings
    }
    
    /// Update isInited flag
    func updateIsInited(_ value: Bool) {
        var config = loadConfig() ?? createDefaultConfig()
        config.uiTheme.isInited = value ? "true" : "false"
        saveConfig(config)
        print("Updated isInited to: \(value)")
    }
    
    /// Update customerID
    func updateCustomerID(_ id: String) {
        var config = loadConfig() ?? createDefaultConfig()
        config.uiTheme.customerID = id
        saveConfig(config)
        print("Updated customerID to: \(id)")
    }
    
    /// Get config file path (for debugging)
    func getConfigPath() -> String? {
        return getConfigFileURL()?.path
    }
    
    /// Reset config (delete runtime config and force copy from bundle)
    func resetConfig() {
        guard let configURL = getConfigFileURL() else {
            print("Failed to get config file URL")
            return
        }
        
        // Delete existing config
        if FileManager.default.fileExists(atPath: configURL.path) {
            do {
                try FileManager.default.removeItem(at: configURL)
                print("Deleted existing config at: \(configURL.path)")
            } catch {
                print("Failed to delete config: \(error)")
                return
            }
        }
        
        // Re-setup (will copy from bundle)
        setupConfigFile()
        print("Config reset complete")
    }
    
    /// 根据 customerID 获取定制化的图片名称
    /// 当 customerID 为 "44" 时，在图片名称后添加 "_44" 后缀
    /// - Parameter imageName: 原始图片名称
    /// - Returns: 处理后的图片名称
    func getCustomizedImageName(_ imageName: String) -> String {
        guard let config = loadConfig() else {
            return imageName
        }
        
        let customerID = config.uiTheme.customerID
        
        // 如果 customerID 是 "44"，添加后缀
        if customerID == "44" {
            return "\(imageName)_44"
        }
        
        return imageName
    }
    
    func currentCustomerID() -> String{
        return SharedDataManager.shared.loadConfig()?.uiTheme.customerID ?? ""
    }
    
    func currentThemeIsRedAndBlack() -> Bool{
        if(SharedDataManager.shared.loadConfig()?.uiTheme.customerID == "44"){
            return true
        }
        return false
    }

    
}


