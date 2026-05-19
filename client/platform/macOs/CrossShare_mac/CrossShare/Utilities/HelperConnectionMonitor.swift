//
//  HelperConnectionMonitor.swift
//  CrossShare
//
//  Created by TS on 2025/12/19.
//  Singleton to periodically check Helper connection status
//

import Foundation
import Cocoa

/// Helper Connection Status Monitor
/// Periodically checks Helper process and connection status, updates UI in time
class HelperConnectionMonitor {
    static let shared = HelperConnectionMonitor()
    
    // MARK: - Properties
    
    /// Check interval in seconds
    private let checkInterval: TimeInterval = 2.0
    
    /// Timer for periodic checks
    private var checkTimer: Timer?
    
    /// Whether monitoring is active
    private var isMonitoring: Bool = false
    
    /// Current monitoring state
    private enum MonitorState {
        case checkingProcess      // Checking if process exists
        case connecting           // Attempting to connect
        case connected            // Connected
    }
    
    private var currentState: MonitorState = .checkingProcess

    private struct PrintLogStatus {
        var isXpcRunning: Bool? = nil
        var isFoundInWorkspace: Bool? = nil
        var isHelperRunning: Bool? = nil
        var isConnectBusy: Bool? = nil
    }
    private var lastLogStatus = PrintLogStatus()
    
    // MARK: - Initialization
    
    private init() {
        logger.info("HelperConnectionMonitor initialized")
    }
    
    deinit {
        stopMonitoring()
    }
    
    // MARK: - Public Methods
    
    /// Start monitoring Helper connection status
    func startMonitoring() {
        guard !isMonitoring else {
            logger.debug("HelperConnectionMonitor is already monitoring")
            return
        }
        
        isMonitoring = true
        currentState = .checkingProcess
        logger.info("Starting Helper connection monitoring (interval: \(checkInterval)s)")
        
        // Create timer
        checkTimer = Timer.scheduledTimer(withTimeInterval: checkInterval, repeats: true) { [weak self] _ in
            self?.checkAndUpdate()
        }
        
        // Execute check immediately
        checkAndUpdate()
    }
    
    /// Stop monitoring
    func stopMonitoring() {
        guard isMonitoring else {
            return
        }
        
        isMonitoring = false
        checkTimer?.invalidate()
        checkTimer = nil
        currentState = .checkingProcess
        logger.info("Stopped Helper connection monitoring")
    }
    
    // MARK: - Private Methods
    
    /// Periodically check and update status
    private func checkAndUpdate() {
        switch currentState {
        case .checkingProcess:
            // State 1: Check if Helper process exists
            checkHelperProcess()
            
        case .connecting:
            // State 2: Check if connection succeeded
            checkConnectionStatus()
            
        case .connected:
            // State 3: Verify connection is still valid
            verifyConnection()
        }
    }
    
    /// Check if Helper process exists
    private func checkHelperProcess() {
        // Use actual process check method instead of SMAppService status
        let helperRunning = isHelperProcessActuallyRunning()
        
        let isPrintLog = (helperRunning != lastLogStatus.isHelperRunning)
        lastLogStatus.isHelperRunning = helperRunning

        if helperRunning {
            if isPrintLog { logger.info("Helper process found, transitioning to connecting state") }
            currentState = .connecting
            // Once process exists, attempt connection immediately
            connectToHelper()
        } else {
            if isPrintLog { logger.debug("Helper process not found, will continue checking...") }
            // 兜底：登录项已开启但进程缺失时，尝试用 launchctl kickstart 拉起一次。
            // 内部自带限频，避免高频触发。
            HelperCommunication.shared.tryReviveHelperByLaunchctlIfNeeded(reason: "monitor_checking_process") { [weak self] revived in
                guard let self = self else { return }
                // 优化：兜底拉起成功后，立即切到连接态，避免完全等待下一轮定时器。
                guard revived else { return }
                logger.info("Helper revived by launchctl, trying immediate connection")
                self.currentState = .connecting
                self.connectToHelper()
            }
            // Process doesn't exist, set status = 1
            setDisconnectedStatus()
        }
    }
    
    /// Actually check if Helper process is running (using pgrep)
    private func isHelperProcessActuallyRunning() -> Bool {
        let helperBundleIdentifier = BundleIdentifiers.helper
        
        // Method 1: Try NSWorkspace.runningApplications
        let runningApps = NSWorkspace.shared.runningApplications
        let foundInWorkspace = runningApps.contains { $0.bundleIdentifier == helperBundleIdentifier }
        
        let isPrintLog = (foundInWorkspace != lastLogStatus.isFoundInWorkspace)
        lastLogStatus.isFoundInWorkspace = foundInWorkspace

        if foundInWorkspace {
            if isPrintLog { logger.debug("Helper detected via NSWorkspace.runningApplications") }
            return true
        }
        
        // Method 2: Use pgrep with bundle identifier (most reliable)
        let task = Process()
        task.launchPath = "/usr/bin/pgrep"
        task.arguments = ["-f", helperBundleIdentifier]
        
        let pipe = Pipe()
        task.standardOutput = pipe
        task.standardError = pipe
        
        do {
            try task.run()
            task.waitUntilExit()
            
            // pgrep returns 0 if process found, 1 if not found
            if task.terminationStatus == 0 {
                let data = pipe.fileHandleForReading.readDataToEndOfFile()
                if let output = String(data: data, encoding: .utf8), !output.isEmpty {
                    let pids = output.trimmingCharacters(in: .whitespacesAndNewlines).components(separatedBy: "\n")
                    logger.debug("Helper detected via pgrep, PIDs: \(pids.joined(separator: ", "))")
                    return true
                }
            }
        } catch {
            logger.debug("Failed to run pgrep: \(error)")
        }
        
        if isPrintLog { logger.debug("Helper process not detected") }
        return false
    }
    
    /// Connect to Helper
    private func connectToHelper() {
        // If already connected, update status directly
        if HelperClient.shared.connectionStatus {
            logger.info("Already connected to Helper, updating status")
            currentState = .connected
            updateDiasStatus()
            return
        }
        
        // 若主流程/其他路径已经在连接，监控侧不再重复触发，避免日志噪声和无效调用
        let connectBusy = HelperClient.shared.isConnectionBusy
        let busyLogChanged = (connectBusy != lastLogStatus.isConnectBusy)
        lastLogStatus.isConnectBusy = connectBusy
        if connectBusy {
            if busyLogChanged { logger.debug("Helper connection is busy, skip duplicate monitor connect trigger") }
            return
        }
        
        logger.info("Attempting to connect to Helper...")
        HelperClient.shared.connect { [weak self] success, error in
            guard let self = self else { return }
            
            if success {
                logger.info("Successfully connected to Helper")
                self.currentState = .connected
                // After successful connection, update status
                self.updateDiasStatus()
            } else {
                // 已由上层做去重，正常情况下不应再频繁出现 "Already connecting"
                // 即便出现，也按 debug 级别记录，避免污染主流程日志。
                logger.debug("Failed to connect to Helper: \(error ?? "Unknown error")")
                // Connection failed, stay in connecting state, will retry on next check
                // If process disappears, will return to checkingProcess state on next check
            }
        }
    }
    
    /// Check connection status (in connecting state)
    private func checkConnectionStatus() {
        // If already connected successfully, update status
        if HelperClient.shared.connectionStatus {
            logger.info("Connection established, updating status")
            currentState = .connected
            updateDiasStatus()
        } else {
            // Check if process still exists
            let helperRunning = isHelperProcessActuallyRunning()
            if !helperRunning {
                logger.warn("Helper process disappeared during connection attempt")
                currentState = .checkingProcess
                setDisconnectedStatus()
            } else {
                // Process exists but not connected, continue attempting connection
                connectToHelper()
            }
        }
    }
    
    /// Verify connection is still valid (in connected state)
    private func verifyConnection() {
        // First check if process is actually running
        let helperRunning = isHelperProcessActuallyRunning()
        
        if !helperRunning {
            logger.warn("Helper process not running, connection lost")
            currentState = .checkingProcess
            setDisconnectedStatus()
            return
        }
        
        // Check HelperClient connection status
        let clientConnected = HelperClient.shared.connectionStatus
        
        if !clientConnected {
            logger.warn("HelperClient not connected, connection lost")
            currentState = .checkingProcess
            setDisconnectedStatus()
            return
        }
        
        // Process exists and client is connected, but need to verify XPC connection is actually valid
        // Verify connection by attempting to call a simple method
        verifyXPCConnection()
    }
    
    /// Verify XPC connection is actually valid
    private func verifyXPCConnection() {
        // Attempt to call getServiceStatus to verify connection is actually valid
        HelperClient.shared.getServiceStatus { [weak self] isRunning, info in
            guard let self = self else { return }

            let isPrintLog = (isRunning != lastLogStatus.isXpcRunning)
            lastLogStatus.isXpcRunning = isRunning

            if isRunning {
                // Connection is valid
                if isPrintLog { logger.debug("Connection verified: OK (XPC call succeeded)") }
            } else {
                // XPC call failed, connection may be lost
                if isPrintLog { logger.warn("XPC connection verification failed, connection lost") }
                DispatchQueue.main.async {
                    self.currentState = .checkingProcess
                    self.setDisconnectedStatus()
                }
            }
        }
    }
    
    /// Set disconnected status (status = 1)
    private func setDisconnectedStatus() {
        // logger.warn("Setting disconnected status (status = 1)")
        
        DispatchQueue.main.async {
            NotificationCenter.default.post(
                name: .deviceDiasStatusNotification,
                object: nil,
                userInfo: [
                    "diasStatus": 1,
                ]
            )
        }
    }
    
    /// Update DIAS status (after successful connection)
    private func updateDiasStatus() {
        HelperClient.shared.updateDiasStatus { status in
            logger.info("DIAS status updated after connection: \(status)")
        }
    }
}

