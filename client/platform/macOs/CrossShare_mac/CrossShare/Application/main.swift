//  Copyright © MonitorControl. @JoniVR, @theOneyouseek, @waydabber and others

import Cocoa
import Foundation

let DEBUG_SW = false
let DEBUG_VIRTUAL = false
let DEBUG_MACOS10 = false
let DEBUG_GAMMA_ENFORCER = false
let DDC_MAX_DETECT_LIMIT: Int = 100

let MIN_PREVIOUS_BUILD_NUMBER = 6262

var app: AppDelegate!

let prefs = UserDefaults.standard

private let storyboard = NSStoryboard(name: "Main", bundle: Bundle.main)

autoreleasepool { () in
  let mc = NSApplication.shared
  let mcDelegate = AppDelegate()
  app = mcDelegate
  mc.delegate = mcDelegate
  mc.run()
}
