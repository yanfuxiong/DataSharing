//
//  GlobalLogger.swift
//  CrossShare
//
//  Global logger instance that can be used directly in all files
//

import Foundation

/// Global logger instance that can be used anywhere directly
/// Usage: logger.info("..."), logger.error("..."), etc.
let logger = CSLogger.shared

