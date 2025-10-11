//
//  DashedBorderView.swift
//  CrossShare
//
//  Created by user00 on 2025/3/6.
//

import Cocoa

class DashedBorderView: NSView {
    private var dragStartPoint: NSPoint? // 记录拖拽开始的位置
    private let uploadThreshold: CGFloat = 20 // 超过多少像素触发上传
    
    override func draw(_ dirtyRect: NSRect) {
        super.draw(dirtyRect)
        guard (NSGraphicsContext.current?.cgContext) != nil else { return }
        let path = NSBezierPath(rect: bounds.insetBy(dx: 5, dy: 5))
        let dashPattern: [CGFloat] = [6, 4]
        path.setLineDash(dashPattern, count: dashPattern.count, phase: 0)
        path.lineWidth = 2
        NSColor(hex: 0x377AF6).setStroke()
        path.stroke()
    }
    
    override init(frame frameRect: NSRect) {
        super.init(frame: frameRect)
        registerForDraggedTypes([.fileURL])
    }
    
    required init?(coder: NSCoder) {
        super.init(coder: coder)
        registerForDraggedTypes([.fileURL])
    }
    
    //1. 当拖拽进入View时
    override func draggingEntered(_ sender: NSDraggingInfo) -> NSDragOperation {
        dragStartPoint = sender.draggingLocation // 记录拖拽的起点
        return .copy
    }
    
    //2. 当拖拽在View内移动时，检测是否超过60px
    override func draggingUpdated(_ sender: NSDraggingInfo) -> NSDragOperation {
        guard let startPoint = dragStartPoint else { return .copy }
        
        let currentPoint = sender.draggingLocation
        let distanceDragged = hypot(currentPoint.x - startPoint.x, currentPoint.y - startPoint.y)
        
        if distanceDragged > uploadThreshold {
            print("🚀 拖拽超过 60px，开始上传文件")
            if let files = getDraggedFileURLs(from: sender) {
                for file in files {
                    uploadFileToServer(file)
                }
            }
            dragStartPoint = nil // 避免重复触发
        }
        return .copy
    }
    
    //3. 拖拽结束
    override func draggingExited(_ sender: NSDraggingInfo?) {
        dragStartPoint = nil // 清空起点
    }
    
    override func performDragOperation(_ sender: NSDraggingInfo) -> Bool {
        return true
    }
    
    // 4 解析拖拽进来的文件路径
    private func getDraggedFileURLs(from sender: NSDraggingInfo) -> [URL]? {
        let pasteboard = sender.draggingPasteboard
        return pasteboard.readObjects(forClasses: [NSURL.self], options: nil) as? [URL]
    }
    
}

extension DashedBorderView {
    @objc private func uploadFileToServer(_ fileUrl:URL) {
        print("🚀 准备上传的文件：\(fileUrl.absoluteString) \(fileUrl.lastPathComponent)")
    }
}
