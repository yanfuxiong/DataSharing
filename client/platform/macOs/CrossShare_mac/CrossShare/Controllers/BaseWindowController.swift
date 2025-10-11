//
//  BaseWindowController.swift
//  CrossShare
//
//  Created by user00 on 2025/3/5.
//

import Cocoa

class BaseWindowController: NSWindowController {
    convenience init() {
        let windowSize = NSRect(x: 0, y: 0, width: 600, height: 400)
        let window = NSWindow(contentRect: windowSize,
                              styleMask: [.titled, .closable, .resizable],
                              backing: .buffered,
                              defer: false)
        window.title = "My macOS App"

        self.init(window: window)
        self.window?.center()

        // Load NSViewController
//        let viewController = HomeViewController()
        let viewController = MainHomeViewController()

        self.contentViewController = viewController
    }
}

