import Cocoa
import SnapKit

// MARK: - CSCustomProgressView
// 自定义进度条视图，支持自定义高度和颜色
class CSCustomProgressView: NSView {
    
    // 进度值 (0.0 ~ 1.0)
    var progress: Double = 0.0 {
        didSet {
            progress = min(max(progress, 0.0), 1.0)
            needsDisplay = true
            updatePercentageLabel()
        }
    }
    
    // 未完成部分的颜色（深色背景）
    var trackColor: NSColor = NSColor(white: 0.65, alpha: 0.4)
    
    // 已完成部分的颜色（进度条颜色）
    var progressColor: NSColor = NSColor.systemBlue
    
    // 圆角半径
    private let cornerRadius: CGFloat = 10.0
    
    // 百分比标签
    private lazy var percentageLabel: NSTextField = {
        let label = NSTextField()
        label.isEditable = false
        label.isBordered = false
        label.backgroundColor = .clear
        label.font = NSFont.systemFont(ofSize: 11, weight: .medium)
        label.textColor = .white
        label.alignment = .center
        label.stringValue = "0%"
        return label
    }()
    
    override init(frame frameRect: NSRect) {
        super.init(frame: frameRect)
        setupView()
    }
    
    required init?(coder: NSCoder) {
        super.init(coder: coder)
        setupView()
    }
    
    private func setupView() {
        wantsLayer = true
        addSubview(percentageLabel)
        
        percentageLabel.snp.makeConstraints { make in
            make.center.equalToSuperview()
        }
    }
    
    override func draw(_ dirtyRect: NSRect) {
        super.draw(dirtyRect)
        
        let bounds = self.bounds
        
        // 绘制背景（未完成部分 - 深色）
        let backgroundPath = NSBezierPath(roundedRect: bounds, xRadius: cornerRadius, yRadius: cornerRadius)
        trackColor.setFill()
        backgroundPath.fill()
        
        // 绘制进度（已完成部分 - 系统蓝色）
        if progress > 0 {
            let progressWidth = bounds.width * CGFloat(progress)
            let progressRect = NSRect(x: 0, y: 0, width: progressWidth, height: bounds.height)
            let progressPath = NSBezierPath(roundedRect: progressRect, xRadius: cornerRadius, yRadius: cornerRadius)
            progressColor.setFill()
            progressPath.fill()
        }
    }
    
    private func updatePercentageLabel() {
        let percentage = Int(progress * 100)
        percentageLabel.stringValue = "\(percentage)%"
    }
}
