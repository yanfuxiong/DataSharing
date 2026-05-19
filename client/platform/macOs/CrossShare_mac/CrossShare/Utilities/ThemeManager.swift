//
//  ThemeManager.swift
//  CrossShare
//
//  Created by TS on 2025/11/4.
//

import Cocoa

// MARK: - Theme Configuration

/// Theme color configuration
struct ThemeColors {
    // MARK: - Main colors
    
    /// 主色调 - 当前主题的主要颜色，可用于按钮、强调文本等需要突出显示的元素
    let primaryColor: NSColor
    
    /// 背景色 - 整个应用的主背景颜色
    /// 应用于：MainHomeViewController 的 view.layer?.backgroundColor
    let backgroundColor: NSColor
    
    /// 文字颜色 - 主要文本的颜色
    let textColor: NSColor
    
    // MARK: - Border colors
    
    /// 通用边框颜色 - 一般性边框的颜色（预留字段）
    let borderColor: NSColor
    
    /// 菜单边框颜色 - 所有 NSMenu 弹出菜单的边框颜色
    /// 应用于：Options 菜单、Device List 菜单、License 菜单、右键菜单等
    /// 实现位置：menuWillOpen(_ menu:) 中的 contentView.layer?.borderColor
    let menuBorderColor: NSColor
    
    /// 主视图边框颜色 - 整个主窗口外围的边框颜色
    /// 应用于：整个 MainHomeViewController 的 view 外边框
    /// 实现位置：addMainViewBorder() 方法
    let mainViewBorderColor: NSColor
    
    /// 虚线边框颜色 - 中间内容区域的边框颜色
    /// 应用于：包裹左右两个文件列表的虚线/实线边框
    /// 红黑主题下是红色实线，默认主题下是蓝色虚线
    /// 实现位置：addBorderView(to:) 方法
    let dashedBorderColor: NSColor
    
    // MARK: - Top view
    
    /// 顶部视图背景色 - HomeHeaderView 的背景颜色
    /// 应用于：顶部显示连接状态和设备按钮的区域
    /// 默认主题是蓝色，红黑主题是黑色
    let topViewBackgroundColor: NSColor
    
    /// 顶部分隔线颜色 - 顶部视图下方的横线颜色
    /// 应用于：topViewRedLine 视图的背景色
    /// 如果是 .clear（透明）则不显示分隔线
    /// 红黑主题显示红色横线，默认主题不显示
    let topViewSeparatorColor: NSColor
    
    // MARK: - Right header
    
    /// 右侧头部背景色 - 包含 "Back" 按钮和面包屑导航的容器背景色
    /// 应用于：rightHeaderContainer 的背景
    /// 如果是 nil 则不显示背景
    /// 红黑主题显示白色背景（形成对比），默认主题是透明的
    let rightHeaderBackgroundColor: NSColor?
    
    /// 右侧头部内边距 - Back 按钮和面包屑导航的内边距（单位：像素）
    /// 应用于：rightHeaderContainer 内部元素的约束
    /// 红黑主题是 8px（有白色背景需要内边距），默认主题是 0
    let rightHeaderPadding: CGFloat
    
    // MARK: - Specific elements
    
    /// "Files records" 文字颜色 - 底部表格标题的文字颜色
    /// 应用于：bottomTitleLabel 的 textColor
    /// 红黑主题是红色，默认主题是深灰色
    let filesRecordsTextColor: NSColor
    
    // MARK: - Border styles
    
    /// 主视图边框宽度 - 整个主窗口外边框的线条宽度（单位：像素）
    /// 应用于：addMainViewBorder() 中的 borderLayer.lineWidth
    /// 红黑主题是 1.5px，默认主题是 0（不显示）
    let mainViewBorderWidth: CGFloat
    
    /// 菜单边框宽度 - 所有弹出菜单的边框宽度（单位：像素）
    /// 应用于：menuWillOpen(_ menu:) 中的 contentView.layer?.borderWidth
    /// 红黑主题是 1.5px，默认主题是 0（不显示）
    let menuBorderWidth: CGFloat
    
    /// 是否使用虚线边框 - 控制中间内容区域的边框是虚线还是实线
    /// true = 虚线 [4, 2]（4 像素实线，2 像素空白）
    /// false = 实线 nil
    /// 应用于：addBorderView(to:) 中的 borderLayer.lineDashPattern
    /// 红黑主题是实线，默认主题是虚线
    let isDashedBorder: Bool
    
    
    let fileBrowserViewBorderColor:NSColor?
}

// MARK: - Theme Manager

class ThemeManager {
    static let shared = ThemeManager()
    
    private init() {}
    
    // MARK: - Current Theme
    
    var currentTheme: ThemeColors {
        if SharedDataManager.shared.currentThemeIsRedAndBlack() {
            return redBlackTheme
        } else {
            return defaultTheme
        }
    }
    
    // MARK: - Theme Definitions
    
    /// Default blue theme
    private let defaultTheme = ThemeColors(
        primaryColor: NSColor(hex: 0x377AF6),
        backgroundColor: NSColor(white: 0.95, alpha: 1.0),
        textColor: .darkGray,
        borderColor: .blue,
        menuBorderColor: .clear,
        mainViewBorderColor: .clear,
        dashedBorderColor: .blue,
        topViewBackgroundColor: NSColor(hex: 0x377AF6),
        topViewSeparatorColor: .clear,
        rightHeaderBackgroundColor: nil,
        rightHeaderPadding: 0,
        filesRecordsTextColor: .darkGray,
        mainViewBorderWidth: 0,
        menuBorderWidth: 0,
        isDashedBorder: true,
        fileBrowserViewBorderColor: NSColor(hex: 0x52a5fe)
    )
    
    /// Red-black theme
    private let redBlackTheme = ThemeColors(
        primaryColor: .red,
        backgroundColor: .black,
        textColor: .red,
        borderColor: .red,
        menuBorderColor: .red,
        mainViewBorderColor: .red,
        dashedBorderColor: .red,
        topViewBackgroundColor: .black,
        topViewSeparatorColor: .red,
        rightHeaderBackgroundColor: .white,
        rightHeaderPadding: 8,
        filesRecordsTextColor: .red,
        mainViewBorderWidth: 1.5,
        menuBorderWidth: 1.5,
        isDashedBorder: false,
        fileBrowserViewBorderColor: NSColor.red
    )
    
    // MARK: - Convenience Properties
    
    /// Whether the current theme is red-black
    var isRedBlackTheme: Bool {
        return SharedDataManager.shared.currentThemeIsRedAndBlack()
    }
    
    /// Whether to show main view border
    var shouldShowMainViewBorder: Bool {
        return currentTheme.mainViewBorderWidth > 0
    }
    
    /// Whether to show top view separator
    var shouldShowTopViewSeparator: Bool {
        return currentTheme.topViewSeparatorColor != .clear
    }
    
    /// Whether to show right header background
    var shouldShowRightHeaderBackground: Bool {
        return currentTheme.rightHeaderBackgroundColor != nil
    }
    
    /// Whether to show menu border
    var shouldShowMenuBorder: Bool {
        return currentTheme.menuBorderWidth > 0
    }
}

