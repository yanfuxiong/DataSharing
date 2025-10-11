//
//  GoStringHelper.swift
//  XPCService
//
//  Created by TS on 2025/9/8.
//  Go string conversion helper for XPCService
//

import Foundation

extension String {
    func toGoStringXPC() -> _GoString_ {
        let cString = strdup(self)
        return _GoString_(p: cString, n: Int(Int64(self.utf8.count)))
    }
    
    func toCStringXPC() -> UnsafePointer<CChar>? {
        guard let cString = strdup(self) else { return nil }
        return UnsafePointer(cString)
    }
}
