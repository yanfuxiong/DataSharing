//
//  HelperNotificationManager.swift
//  CrossShareHelper
//
//  Created by ts on 2025/1/30.
//  Notification manager singleton for Helper process
//

import Foundation
import AppKit
import UserNotifications

/// Notification permission request callback protocol
protocol HelperNotificationManagerDelegate: AnyObject {
    func notificationManagerRequestOpenNotiAlert()
}

/// Notification manager singleton for Helper process
class HelperNotificationManager {
    
    static let shared = HelperNotificationManager()
    
    private let logger = CSLogger.shared
    weak var delegate: HelperNotificationManagerDelegate?
    
    private init() {}
    
    // MARK: - Permission Check
    
    /// Check notification permission status
    /// - Returns: true if notification permission is authorized
    func checkNotificationStatus() -> Bool {
        let center = UNUserNotificationCenter.current()
        var isAuthorized = false
        let semaphore = DispatchSemaphore(value: 0)
        
        center.getNotificationSettings { settings in
            isAuthorized = (settings.authorizationStatus == .authorized)
            self.logger.log("Helper notification permission status: \(isAuthorized ? "authorized" : "not authorized")", level: .info)
            semaphore.signal()
        }
        
        // Wait for async result, timeout after 1 second
        _ = semaphore.wait(timeout: .now() + 1.0)
        
        return isAuthorized
    }
    
    /// Ensure notification permission is available for Helper process.
    /// If status is notDetermined, this will trigger system authorization request.
    /// - Parameter completion: Returns true when authorized/provisional, false otherwise
    func checkAndRequestNotificationAuthorization(completion: @escaping (Bool) -> Void) {
        ensureNotificationAuthorization(completion: completion)
    }
    
    // MARK: - Clipboard Notification
    
    /// Send clipboard copy notification
    /// - Parameters:
    ///   - text: Text content (TEXT takes priority)
    ///   - image: Image content
    ///   - html: HTML content
    ///   - rtf: RTF content
    func sendClipboardNotification(text: String?, image: NSImage?, html: String?, rtf: String? = nil) {
        let shareFeatAvailable = GetShareFeatAvailable()
        logger.log("GetShareFeatAvailable() returned: \(shareFeatAvailable)", level: .info)
        if shareFeatAvailable != 1 {
            logger.log("ShareFeat not available (value=\(shareFeatAvailable)), skipping clipboard notification popup", level: .info)
            return
        }
        let hasImage = image != nil
        let hasText = text != nil && !text!.isEmpty
        let hasHTML = html != nil && !html!.isEmpty
        let hasRTF = rtf != nil && !rtf!.isEmpty
        
        var title: String
        var body: String = ""
        var notificationImage: NSImage? = nil
        
        // Display priority: Image > TEXT > HTML = RTF
        if hasImage {
            title = "Image saved to clipboard"
            notificationImage = createThumbnail(from: image!, maxHeight: 80)
        }
        else if hasText {
            title = "Text saved to clipboard"
            body = formatTextForNotification(text!)
        }
        else if hasHTML {
            title = "Text saved to clipboard"
            let plainText = extractPlainTextFromHTML(html!)
            body = formatTextForNotification(plainText)
        }
        else if hasRTF {
            title = "Text saved to clipboard"
            let plainText = extractPlainTextFromRTF(rtf!)
            body = formatTextForNotification(plainText)
        }
        else {
            logger.log("No valid clipboard content for notification", level: .info)
            return
        }
        
        logger.log("Sending clipboard notification - title: \(title), hasImage: \(notificationImage != nil)", level: .info)
        
        ensureNotificationAuthorization { [weak self] authorized in
            guard let self = self else { return }
            guard authorized else {
                self.logger.warn("Clipboard notification skipped because notification permission is not granted")
                return
            }
            
            self.deliverNotification(
                title: title,
                body: body,
                identifier: "clipboard-\(UUID().uuidString)",
                image: notificationImage,
                category: "clipboard"
            )
        }
    }
    
    // MARK: - System Notification
    
    /// Send system notification (based on notiCode)
    /// - Parameters:
    ///   - notiCode: Notification code
    ///   - notiParam: Notification parameters
    func sendSystemNotification(notiCode: UInt32, notiParam: [String]) {
        var title: String = ""
        var body: String = ""

        switch notiCode {
        case 1: // Connection status - online
            guard notiParam.count >= 2 else {
                logger.error("NotiCode 1 requires at least 2 parameters, got: \(notiParam.count)")
                return
            }
            title = "Connection status"
            body = "\(notiParam[0]) is now online, Devices online count: \(notiParam[1])"

        case 2: // Connection status - disconnected
            guard notiParam.count >= 2 else {
                logger.error("NotiCode 2 requires at least 2 parameters, got: \(notiParam.count)")
                return
            }
            title = "Connection status"
            body = "\(notiParam[0]) has been disconnected, Devices online count: \(notiParam[1])"

        case 3: // File transfer - sender complete
            guard notiParam.count >= 2 else {
                logger.error("NotiCode 3 requires at least 2 parameters, got: \(notiParam.count)")
                return
            }
            title = "File transfer"
            body = "\(notiParam[0]) transferred to \(notiParam[1]) is complete"

        case 4: // File transfer - receiver complete
            guard notiParam.count >= 2 else {
                logger.error("NotiCode 4 requires at least 2 parameters, got: \(notiParam.count)")
                return
            }
            title = "File transfer"
            body = "\(notiParam[0]) received from \(notiParam[1]) is complete"

        case 5: // File transfer - declined
            guard notiParam.count >= 2 else {
                logger.error("NotiCode 5 requires at least 2 parameters, got: \(notiParam.count)")
                return
            }
            title = "File transfer"
            body = "\(notiParam[0]) declined to receive \(notiParam[1])"

        default:
            logger.warn("Unknown notiCode: \(notiCode), params: \(notiParam)")
            return
        }

        ensureNotificationAuthorization { [weak self] authorized in
            guard let self = self else { return }
            guard authorized else { return }
            
            self.deliverNotification(
                title: title,
                body: body,
                identifier: "system-\(notiCode)-\(UUID().uuidString)",
                category: "system-\(notiCode)"
            )
        }
    }
    
    // MARK: - Core Notification Delivery
    
    /// Unified notification delivery method
    /// - Parameters:
    ///   - title: Notification title
    ///   - body: Notification body content
    ///   - identifier: Notification identifier (defaults to UUID)
    ///   - image: Optional image attachment
    ///   - category: Notification category (for logging)
    ///   - soundEnabled: Whether to play sound
    func deliverNotification(
        title: String,
        body: String,
        identifier: String = UUID().uuidString,
        image: NSImage? = nil,
        category: String = "general",
        soundEnabled: Bool = true
    ) {
        let content = UNMutableNotificationContent()
        content.title = title
        content.body = body
        content.sound = soundEnabled ? .default : nil
        
        // If there's an image, add as attachment
        if let image = image {
            if let attachment = createNotificationAttachment(from: image) {
                content.attachments = [attachment]
            }
        }

        let request = UNNotificationRequest(
            identifier: identifier,
            content: content,
            trigger: nil
        )

        UNUserNotificationCenter.current().add(request) { error in
            if let error = error {
                self.logger.error("Failed to send \(category) notification: \(error.localizedDescription)")
            } else {
                self.logger.info("Notification sent [\(category)] - Title: \(title), Body: \(body.prefix(50))...")
            }
        }
    }
    
    // MARK: - Helper Methods
    
    private func ensureNotificationAuthorization(completion: @escaping (Bool) -> Void) {
        let center = UNUserNotificationCenter.current()
        
        center.getNotificationSettings { settings in
            self.logger.info("Current notification settings - Authorization: \(settings.authorizationStatus.rawValue), Alert: \(settings.alertSetting.rawValue), Sound: \(settings.soundSetting.rawValue)")
            
            switch settings.authorizationStatus {
            case .authorized, .provisional:
                completion(true)
                
            case .notDetermined:
                self.logger.info("Requesting notification permission...")
                center.requestAuthorization(options: [.alert, .sound, .badge]) { granted, error in
                    if granted {
                        self.logger.info("Notification permission granted")
                    } else {
                        self.logger.warn("Notification permission denied: \(error?.localizedDescription ?? "Unknown error")")
                        DispatchQueue.main.async { [weak self] in
                            self?.delegate?.notificationManagerRequestOpenNotiAlert()
                        }
                    }
                    completion(granted)
                }
                
            case .denied:
                self.logger.warn("Notification permission denied by user. Please enable it in System Settings > Notifications > CrossShare Helper")
                DispatchQueue.main.async { [weak self] in
                    self?.delegate?.notificationManagerRequestOpenNotiAlert()
                }
                completion(false)
                
            @unknown default:
                self.logger.warn("Unknown notification authorization status: \(settings.authorizationStatus.rawValue)")
                completion(false)
            }
        }
    }
    
    /// Format text for notification display (max 2 lines, truncate with "...")
    /// macOS notification displays about 22-25 characters per line, ~45-50 for 2 lines
    private func formatTextForNotification(_ text: String) -> String {
        // Replace all line breaks with spaces, merge into single line
        var cleanText = text.replacingOccurrences(of: "\r\n", with: " ")
        cleanText = cleanText.replacingOccurrences(of: "\n", with: " ")
        cleanText = cleanText.replacingOccurrences(of: "\r", with: " ")
        
        // Clean up extra spaces
        cleanText = cleanText.replacingOccurrences(
            of: "\\s+",
            with: " ",
            options: .regularExpression
        )
        cleanText = cleanText.trimmingCharacters(in: .whitespacesAndNewlines)
        
        // Limit total characters (approximately 2 lines)
        let maxTotalChars = 45
        
        if cleanText.count <= maxTotalChars {
            return cleanText
        }
        
        // Truncate and add ellipsis
        return String(cleanText.prefix(maxTotalChars)) + "..."
    }
    
    /// Extract plain text from HTML (users can't read HTML tags)
    private func extractPlainTextFromHTML(_ html: String) -> String {
        var plainText = html
        
        // Remove script and style tags with their content
        plainText = plainText.replacingOccurrences(
            of: "<script[^>]*>.*?</script>",
            with: "",
            options: [.regularExpression, .caseInsensitive]
        )
        plainText = plainText.replacingOccurrences(
            of: "<style[^>]*>.*?</style>",
            with: "",
            options: [.regularExpression, .caseInsensitive]
        )
        
        // Remove all HTML tags
        plainText = plainText.replacingOccurrences(
            of: "<[^>]+>",
            with: "",
            options: .regularExpression
        )
        
        // Decode HTML entities
        plainText = plainText.replacingOccurrences(of: "&nbsp;", with: " ")
        plainText = plainText.replacingOccurrences(of: "&amp;", with: "&")
        plainText = plainText.replacingOccurrences(of: "&lt;", with: "<")
        plainText = plainText.replacingOccurrences(of: "&gt;", with: ">")
        plainText = plainText.replacingOccurrences(of: "&quot;", with: "\"")
        plainText = plainText.replacingOccurrences(of: "&#39;", with: "'")
        
        // Clean up extra whitespace
        plainText = plainText.trimmingCharacters(in: .whitespacesAndNewlines)
        plainText = plainText.replacingOccurrences(
            of: "\\s+",
            with: " ",
            options: .regularExpression
        )
        
        return plainText
    }
    
    /// Extract plain text from RTF string
    private func extractPlainTextFromRTF(_ rtf: String) -> String {
        guard let rtfData = rtf.data(using: .utf8) else {
            return rtf
        }
        
        if let attributedString = try? NSAttributedString(
            data: rtfData,
            options: [.documentType: NSAttributedString.DocumentType.rtf],
            documentAttributes: nil
        ) {
            let plainText = attributedString.string.trimmingCharacters(in: .whitespacesAndNewlines)
            if !plainText.isEmpty {
                return plainText
            }
        }
        
        // Fallback: strip RTF control words with regex
        var stripped = rtf
        stripped = stripped.replacingOccurrences(
            of: "\\{\\\\[^{}]*\\}",
            with: "",
            options: .regularExpression
        )
        stripped = stripped.replacingOccurrences(
            of: "\\\\[a-zA-Z]+\\d*\\s?",
            with: "",
            options: .regularExpression
        )
        stripped = stripped.replacingOccurrences(of: "{", with: "")
        stripped = stripped.replacingOccurrences(of: "}", with: "")
        stripped = stripped.trimmingCharacters(in: .whitespacesAndNewlines)
        
        return stripped.isEmpty ? rtf : stripped
    }
    
    /// Create image thumbnail (limit max height, scale proportionally, compress quality)
    private func createThumbnail(from image: NSImage, maxHeight: CGFloat) -> NSImage? {
        let originalSize = image.size
        guard originalSize.width > 0 && originalSize.height > 0 else {
            return nil
        }
        
        // Calculate scaled size
        var newSize = originalSize
        if originalSize.height > maxHeight {
            let scale = maxHeight / originalSize.height
            newSize = NSSize(
                width: originalSize.width * scale,
                height: maxHeight
            )
        }
        
        // Create thumbnail
        let thumbnail = NSImage(size: newSize)
        thumbnail.lockFocus()
        NSGraphicsContext.current?.imageInterpolation = .high
        image.draw(
            in: NSRect(origin: .zero, size: newSize),
            from: NSRect(origin: .zero, size: originalSize),
            operation: .copy,
            fraction: 1.0
        )
        thumbnail.unlockFocus()
        
        // Compress image (reduce quality to minimize size)
        guard let tiffData = thumbnail.tiffRepresentation,
              let bitmapRep = NSBitmapImageRep(data: tiffData),
              let compressedData = bitmapRep.representation(
                using: .jpeg,
                properties: [.compressionFactor: 0.5]
              ) else {
            return thumbnail
        }
        
        return NSImage(data: compressedData)
    }
    
    /// Create notification image attachment
    private func createNotificationAttachment(from image: NSImage) -> UNNotificationAttachment? {
        guard let tiffData = image.tiffRepresentation,
              let bitmapRep = NSBitmapImageRep(data: tiffData),
              let pngData = bitmapRep.representation(using: .png, properties: [:]) else {
            logger.error("Failed to convert image to PNG data for notification attachment")
            return nil
        }
        
        // Write to temporary file
        let tempDir = FileManager.default.temporaryDirectory
        let fileName = "notification_image_\(UUID().uuidString).png"
        let fileURL = tempDir.appendingPathComponent(fileName)
        
        do {
            try pngData.write(to: fileURL)
            let attachment = try UNNotificationAttachment(
                identifier: fileName,
                url: fileURL,
                options: [UNNotificationAttachmentOptionsThumbnailHiddenKey: false]
            )
            return attachment
        } catch {
            logger.error("Failed to create notification attachment: \(error.localizedDescription)")
            return nil
        }
    }
}
