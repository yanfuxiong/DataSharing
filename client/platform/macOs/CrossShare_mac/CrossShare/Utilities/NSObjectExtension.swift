//
//  NSObjectExtension.swift
//  CrossShare
//
//  Created by TS on 2025/10/9.
//

import Foundation
import Cocoa

extension NSBezierPath {
    var cgPath: CGPath {
        let path = CGMutablePath()
        var points = [NSPoint](repeating: .zero, count: 3)  // Mutable array to store up to 3 points
        
        for i in 0..<elementCount {
            let type = element(at: i, associatedPoints: &points)
            switch type {
            case .moveTo:
                path.move(to: points[0])
            case .lineTo:
                path.addLine(to: points[0])
            case .curveTo:
                path.addCurve(to: points[2], control1: points[0], control2: points[1])
            case .closePath:
                path.closeSubpath()
            case .cubicCurveTo:
                break
            case .quadraticCurveTo:
                break
            @unknown default:
                break
            }
        }
        
        return path
    }
}

// Safe array access (unchanged)
extension Collection {
    subscript(safe index: Index) -> Element? {
        return indices.contains(index) ? self[index] : nil
    }
}
