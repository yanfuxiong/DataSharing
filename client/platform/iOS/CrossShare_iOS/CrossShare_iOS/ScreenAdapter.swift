import UIKit

/// 屏幕适配工具，基于设计图尺寸和基准屏幕尺寸进行等比例缩放
class ScreenAdapter {
    
    static let shared = ScreenAdapter()
    
    let designWidth: CGFloat = 402.0
    
    let designHeight: CGFloat = 874.0
    
    let designScale: CGFloat = 1.0
    
    let baseScreenWidth: CGFloat = 375.0
    
    let baseScreenHeight: CGFloat = 812.0
    
    var screenWidth: CGFloat {
        return UIScreen.main.bounds.width
    }
    
    var screenHeight: CGFloat {
        return UIScreen.main.bounds.height
    }
    
    var scaleX: CGFloat {
        return screenWidth / baseScreenWidth
    }
    
    var scaleY: CGFloat {
        return screenHeight / baseScreenHeight
    }
    
    var designToBaseScaleX: CGFloat {
        return baseScreenWidth / (designWidth / designScale)
    }
    
    var designToBaseScaleY: CGFloat {
        return baseScreenHeight / (designHeight / designScale)
    }
    
    func adaptWidth(_ value: CGFloat) -> CGFloat {
        return (value * designToBaseScaleX * scaleX).rounded(.down)
    }
    
    func adaptHeight(_ value: CGFloat) -> CGFloat {
        return (value * designToBaseScaleY * scaleY).rounded(.down)
    }
    
    func adapt(_ value: CGFloat) -> CGFloat {
        return adaptWidth(value)
    }
    
    func adaptPoint(_ point: CGPoint) -> CGPoint {
        return CGPoint(
            x: adaptWidth(point.x),
            y: adaptHeight(point.y)
        )
    }
    
    func adaptSize(_ size: CGSize) -> CGSize {
        return CGSize(
            width: adaptWidth(size.width),
            height: adaptHeight(size.height)
        )
    }
    
    func adaptRect(_ rect: CGRect) -> CGRect {
        return CGRect(
            x: adaptWidth(rect.origin.x),
            y: adaptHeight(rect.origin.y),
            width: adaptWidth(rect.size.width),
            height: adaptHeight(rect.size.height)
        )
    }
    
    func adaptInsets(_ insets: UIEdgeInsets) -> UIEdgeInsets {
        return UIEdgeInsets(
            top: adaptHeight(insets.top),
            left: adaptWidth(insets.left),
            bottom: adaptHeight(insets.bottom),
            right: adaptWidth(insets.right)
        )
    }
}

extension CGFloat {
    var adaptW: CGFloat {
        return ScreenAdapter.shared.adaptWidth(self)
    }
    
    var adaptH: CGFloat {
        return ScreenAdapter.shared.adaptHeight(self)
    }
    
    var adapt: CGFloat {
        return ScreenAdapter.shared.adapt(self)
    }
}

extension Int {
    var adaptW: CGFloat {
        return CGFloat(self).adaptW
    }
    
    var adaptH: CGFloat {
        return CGFloat(self).adaptH
    }
    
    var adapt: CGFloat {
        return CGFloat(self).adapt
    }
}

extension CGPoint {
    var adapt: CGPoint {
        return ScreenAdapter.shared.adaptPoint(self)
    }
}

extension CGSize {
    var adapt: CGSize {
        return ScreenAdapter.shared.adaptSize(self)
    }
}

extension CGRect {
    var adapt: CGRect {
        return ScreenAdapter.shared.adaptRect(self)
    }
}

extension UIEdgeInsets {
    var adapt: UIEdgeInsets {
        return ScreenAdapter.shared.adaptInsets(self)
    }
}
