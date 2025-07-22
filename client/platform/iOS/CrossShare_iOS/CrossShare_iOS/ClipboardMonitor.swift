//
//  ClipboardMonitor.swift
//  CrossShare
//
//  Created by user00 on 2025/3/7.
//

import UIKit

class ClipboardMonitor {
    private var lastChangeCount = UIPasteboard.general.changeCount
    private var clipboardQueue: DispatchQueue?
    private var clipboardTimer: DispatchSourceTimer?
    private let intervalTime = 1.0
    private static var skipChecking: Bool = false

    private static let shared = ClipboardMonitor()

    private init() {}

    public static func shareInstance() -> ClipboardMonitor {
        return shared
    }

    func startMonitoring() {
        clipboardQueue = DispatchQueue(label: "com.crossShare.clipboardMonitor", qos: .background)
        clipboardTimer = DispatchSource.makeTimerSource(queue: clipboardQueue)
        clipboardTimer?.schedule(deadline: .now(), repeating: intervalTime)
        clipboardTimer?.setEventHandler { [weak self] in
            self?.checkClipboard()
        }
        clipboardTimer?.resume()
    }

    func skipLocalChecking() {
        Logger.info("[Clipboard] suspend monitor")
        ClipboardMonitor.skipChecking = true
        let delayTime = DispatchTime.now() + intervalTime
        DispatchQueue.global().asyncAfter(deadline: delayTime) {
            Logger.info("[Clipboard] resume monitor")
            ClipboardMonitor.skipChecking = false
        }
    }

    private func checkClipboard() {
        let pasteboard = UIPasteboard.general
        if pasteboard.changeCount != lastChangeCount {
            lastChangeCount = pasteboard.changeCount
            if ClipboardMonitor.skipChecking {
                Logger.info("[Clipboard] skip local copy event")
                return
            }

            DispatchQueue.main.async { [weak self] in
                self?.handleClipboardChange()
            }
        }
    }

    private func handleClipboardChange() {
        let pasteboard = UIPasteboard.general
        if pasteboard.types.count == 0 {
            return
        }

        Logger.info("[Clipboard] Local copy event type: \(pasteboard.types)")
        if pasteboard.hasImages {
            if let copiedImage = pasteboard.image,let imageBase64 = copiedImage.imageToBase64() {
                Logger.info("[Clipboard] Local copy image event")
                SendImage(imageBase64.toGoString())
            } else {
                Logger.info("[Clipboard][Err] Unknown image type")
            }
        }
        else if pasteboard.hasStrings {
            if let copiedText = pasteboard.string {
                Logger.info("[Clipboard] Local copy text event: \(copiedText)")
                SendText(copiedText.toGoString())
            } else {
                Logger.info("[Clipboard][Err] Unknown text type")
            }
        }
    }

    func stopMonitoring() {
        clipboardTimer?.cancel()
        clipboardTimer = nil
        clipboardQueue = nil
    }
}
