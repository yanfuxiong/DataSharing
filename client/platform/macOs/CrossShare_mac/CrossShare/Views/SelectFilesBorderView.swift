//
//  SelectFilesBorderView.swift
//  CrossShare
//
//  Created by user00 on 2025/3/6.
//

import Cocoa
import UniformTypeIdentifiers

class SelectFilesBorderView: NSView {
    
    let fileUploader = FileUploader()
    
    override func draw(_ dirtyRect: NSRect) {
        super.draw(dirtyRect)
        
    }
    
    override init(frame frameRect: NSRect) {
        super.init(frame: frameRect)
        setupUI()
    }
    
    required init?(coder: NSCoder) {
        super.init(coder: coder)
        setupUI()
    }
    
    func setupUI(){
        addSubview(dashedView)
        dashedView.addSubview(iconImgView)
        dashedView.addSubview(dropLabel)
        dashedView.addSubview(orLabel)
        dashedView.addSubview(leftLineView)
        dashedView.addSubview(rightLineView)
        dashedView.addSubview(selctFilesBtn)
        
        dashedView.snp.makeConstraints { make in
            make.width.equalTo(600)
            make.top.bottom.equalToSuperview()
            make.centerX.equalToSuperview()
        }
        
        iconImgView.snp.makeConstraints { make in
            make.centerX.equalToSuperview()
            make.top.equalToSuperview().offset(50)
            make.size.equalTo(CGSize(width: 216, height: 132))
        }
        
        dropLabel.snp.makeConstraints { make in
            make.centerX.equalToSuperview()
            make.top.equalTo(iconImgView.snp.bottom).offset(22)
        }
        
        leftLineView.snp.makeConstraints { make in
            make.height.equalTo(1)
            make.centerY.equalTo(orLabel.snp.centerY)
            make.width.equalTo(180)
            make.right.equalTo(orLabel.snp.left).offset(-60)
        }
        
        rightLineView.snp.makeConstraints { make in
            make.height.equalTo(1)
            make.centerY.equalTo(orLabel.snp.centerY)
            make.width.equalTo(180)
            make.left.equalTo(orLabel.snp.right).offset(60)
        }
        
        orLabel.snp.makeConstraints { make in
            make.centerX.equalToSuperview()
            make.top.equalTo(dropLabel.snp.bottom).offset(13)
        }
        
        selctFilesBtn.snp.makeConstraints { make in
            make.size.equalTo(CGSize(width: 180, height: 40))
            make.centerX.equalToSuperview()
            make.bottom.equalToSuperview().offset(-20)
        }
    }
    
    lazy var dashedView: DashedBorderView = {
        let cview = DashedBorderView(frame: CGRect(x: 0, y: 0, width: 480, height: 200))
        cview.wantsLayer = true
        cview.layer?.backgroundColor = NSColor.clear.cgColor
        return cview
    }()
    
    lazy var iconImgView: NSImageView = {
        let cview = NSImageView(frame: .zero)
        cview.wantsLayer = true
        cview.image = NSImage(named: "bk_1")
        cview.imageScaling = .scaleAxesIndependently
        cview.layer?.backgroundColor = NSColor.clear.cgColor
        return cview
    }()
    
    lazy var dropLabel: NSTextField = {
        let label = NSTextField(labelWithString: "Drag and drop files here")
        label.frame = CGRect()
        label.font = NSFont.systemFont(ofSize: 16)
        label.alignment = .center
        label.textColor = NSColor.white
        return label
    }()
    
    lazy var orLabel: NSTextField = {
        let label = NSTextField(labelWithString: "OR")
        label.frame = CGRect()
        label.font = NSFont.systemFont(ofSize: 16)
        label.alignment = .center
        label.textColor = NSColor.white
        return label
    }()
    
    lazy var leftLineView: NSView = {
        let cview = NSView(frame: .zero)
        cview.wantsLayer = true
        cview.layer?.backgroundColor = NSColor.init(hex: 0xF5F5F5).cgColor
        return cview
    }()
    
    lazy var rightLineView: NSView = {
        let cview = NSView(frame: .zero)
        cview.wantsLayer = true
        cview.layer?.backgroundColor = NSColor.init(hex: 0xF5F5F5).cgColor
        return cview
    }()
    
    lazy var selctFilesBtn: NSButton = {
        let button = NSButton(title: "Select Files", target: self, action: #selector(selctFilesAction(_:)))
        button.wantsLayer = true
        button.layer?.backgroundColor = NSColor(hex: 0x377AF6).cgColor
        button.bezelStyle = .regularSquare
        button.isBordered = false
        button.layer?.cornerRadius = 5
        let attributes: [NSAttributedString.Key: Any] = [
            .foregroundColor: NSColor.white,
            .backgroundColor: NSColor.clear,
            .font: NSFont.systemFont(ofSize: 13)
        ]
        let attributedTitle = NSAttributedString(string: button.title, attributes: attributes)
        button.attributedTitle = attributedTitle
        return button
    }()
}

extension SelectFilesBorderView {
    @objc func selctFilesAction(_ sender:NSButton) {
        FileSelector.shared.showFileSelector { urls in
            logger.info("用户选择的文件路径：\(urls)")
            if urls.count > 0 {
                if let attributes = try? FileManager.default.attributesOfItem(atPath: urls.first!.path),
                   let fileSize = attributes[.size] as? UInt64 {
                    logger.info("File size: \(fileSize) bytes")
                }
            }
        }
    }
}

class FileSelector {
    static let shared = FileSelector()
    private var openPanel: NSOpenPanel?
    
    func showFileSelector(completion: @escaping ([URL]) -> Void) {
        let openPanel = NSOpenPanel()
        openPanel.title = "请选择你需要传输的文件"
        openPanel.allowedContentTypes = [
            UTType.mp3,
            UTType.quickTimeMovie,
            UTType.avi,
            UTType.audio,
            UTType.video,
            UTType.wav,
            UTType.movie,
            UTType.zip,
            UTType.rtf,
            UTType.image,
            UTType.svg
        ]
        openPanel.allowsMultipleSelection = false
        openPanel.canChooseFiles = true
        openPanel.canChooseDirectories = false    
        self.openPanel = openPanel

        MouseMonitor.shared.startMonitoring { [weak self] event in
            guard let self = self,
                 let panel = self.openPanel else { return }
           guard let selectedUrl = panel.urls.first else { return }
           var isDirectory: ObjCBool = false
           if FileManager.default.fileExists(atPath: selectedUrl.path, isDirectory: &isDirectory) {
               if isDirectory.boolValue {
                   panel.directoryURL = selectedUrl
                   panel.validateVisibleColumns()
               }
           }
        }
        openPanel.begin { [weak self] response in
            MouseMonitor.shared.stopMonitoring()
            if response == .OK {
                completion(openPanel.urls)
            }
            self?.openPanel = nil
        }
    }
    
    private func handleDoubleClick(in view: NSView, at point: NSPoint, panel: NSOpenPanel) {
        guard let selectedUrl = panel.urls.first else { return }
        var isDirectory: ObjCBool = false
        if FileManager.default.fileExists(atPath: selectedUrl.path, isDirectory: &isDirectory) {
            if isDirectory.boolValue {
                // 是目录，进入下一层
                panel.directoryURL = selectedUrl
                // 刷新面板显示
                panel.validateVisibleColumns()
            }
        }
    }
}

class FileUploader {
    func showFileUploadDialog() {
        FileSelector.shared.showFileSelector { urls in
            for url in urls {
                logger.info("Selected file: \(url.path)")
            }
        }
    }
}

extension FileSelector {
    private func customizeFileBrowser(_ panel: NSOpenPanel) {
        panel.directoryURL = FileManager.default.homeDirectoryForCurrentUser
        panel.allowedFileTypes = ["jpg", "png", "pdf"]
        panel.title = "选择要上传的文件"
        panel.prompt = "上传"
        panel.message = "双击文件夹进入，连续双击两次选择文件"
    }
    
    private func handleFileSelection(_ urls: [URL]) {
        for url in urls {
            if url.hasDirectoryPath {
                logger.info("Selected directory: \(url.path)")
            } else {
                logger.info("Selected file: \(url.path)")
                // 检查文件大小
                if let attributes = try? FileManager.default.attributesOfItem(atPath: url.path),
                   let fileSize = attributes[.size] as? UInt64 {
                    logger.info("File size: \(fileSize) bytes")
                }
            }
        }
    }
}
