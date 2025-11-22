import Cocoa

// MARK: - Custom Menu Item View with Button

class CSMenuItemWithButtonView: NSView {
    private let titleLabel = NSTextField(labelWithString: "")
    private let iconImageView = NSImageView()
    private let button = NSButton()
    private var buttonAction: (() -> Void)?
    
    init(title: String, iconName: String, buttonTitle: String, buttonIconName: String, buttonAction: @escaping () -> Void) {
        super.init(frame: NSRect(x: 0, y: 0, width: 250, height: 24))
        
        self.buttonAction = buttonAction
        
        // Setup icon
        iconImageView.image = NSImage(named: iconName)
        iconImageView.imageScaling = .scaleProportionallyDown
        addSubview(iconImageView)
        
        // Setup title label
        titleLabel.stringValue = title
        titleLabel.font = NSFont.systemFont(ofSize: 13)
        titleLabel.textColor = .labelColor
        titleLabel.isBordered = false
        titleLabel.isEditable = false
        titleLabel.backgroundColor = .clear
        addSubview(titleLabel)
        
        // Setup button
        button.title = ""
        button.bezelStyle = .rounded
        button.isBordered = true
        button.target = self
        button.action = #selector(buttonClicked)
        
        // 防止按钮点击时文字闪烁的关键设置
        button.setButtonType(.momentaryPushIn)
        button.showsBorderOnlyWhileMouseInside = false
        
        // Create attributed string for button with icon and text
        let attachment = NSTextAttachment()
        if let buttonIcon = NSImage(named: buttonIconName) {
            let iconSize = NSSize(width: 14, height: 14)
            buttonIcon.size = iconSize
            attachment.image = buttonIcon
            
            // Adjust vertical alignment
            let imageHeight = iconSize.height
            let fontHeight = NSFont.systemFont(ofSize: 11).capHeight
            let yOffset = (fontHeight - imageHeight) / 2
            attachment.bounds = NSRect(x: 0, y: yOffset, width: iconSize.width, height: iconSize.height)
        }
        
        let attachmentString = NSAttributedString(attachment: attachment)
        let buttonTitleString = NSAttributedString(string: " " + buttonTitle, attributes: [
            .font: NSFont.systemFont(ofSize: 11),
            .foregroundColor: NSColor.controlTextColor
        ])
        
        let combinedString = NSMutableAttributedString()
        combinedString.append(attachmentString)
        combinedString.append(buttonTitleString)
        
        button.attributedTitle = combinedString
        button.contentTintColor = .controlTextColor
        
        // Style the button to look clickable
        button.wantsLayer = true
        button.layer?.backgroundColor = NSColor.controlBackgroundColor.cgColor
        button.layer?.cornerRadius = 4
        button.layer?.borderWidth = 0.5
        button.layer?.borderColor = NSColor.separatorColor.cgColor
        
        addSubview(button)
        
        setupLayout()
    }
    
    required init?(coder: NSCoder) {
        fatalError("init(coder:) has not been implemented")
    }
    
    private func setupLayout() {
        iconImageView.frame = NSRect(x: 8, y: 4, width: 16, height: 16)
        titleLabel.frame = NSRect(x: 30, y: 4, width: 120, height: 16)
        button.frame = NSRect(x: 160, y: 2, width: 80, height: 20)
    }
    
    @objc private func buttonClicked() {
        buttonAction?()
        
        // Close the menu after action
        if let menu = self.enclosingMenuItem?.menu {
            menu.cancelTracking()
        }
    }
}

