//
//  PIPStatusView.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/7/2.
//

import UIKit
import AVKit
import AVFoundation

enum PIPContentType {
    case idle           // Idle, no content.
    case textReceived   // Received a text message.
    case imageReceived  // Received an image.
}

class PIPStatusView: NSObject {
    
    private var timer: Timer!
    var bufferDisplayLayer = AVSampleBufferDisplayLayer()
    
    var contentType: PIPContentType = .idle
    var receivedText: String?
    var receivedImage: UIImage?
    
    private let containerView: UIView = {
        let view = UIView()
        view.backgroundColor = UIColor.init(hex: 0xD1E8F9, alpha: 1.0)
        view.layer.cornerRadius = 8
        view.frame = CGRect(x: 0, y: 0, width: 320, height: 60)
        return view
    }()
    
    private let contentView: UIView = {
        let view = UIView()
        view.backgroundColor = UIColor.clear
        return view
    }()
    
    private let emptyTextLabel: UILabel = {
        let label = UILabel()
        label.textColor = UIColor.init(hex: 0x333333)
        label.font = UIFont.systemFont(ofSize: 14)
        label.textAlignment = .center
        label.numberOfLines = 2
        label.lineBreakMode = .byTruncatingTail
        label.backgroundColor = UIColor.clear
        return label
    }()
    
    private let emptyImageView: UIImageView = {
        let imageView = UIImageView()
        imageView.contentMode = .scaleAspectFit
        imageView.clipsToBounds = true
        imageView.backgroundColor = UIColor.clear
        return imageView
    }()
    
    private let textLabel: UILabel = {
        let label = UILabel()
        label.textColor = UIColor.init(hex: 0x333333)
        label.font = UIFont.systemFont(ofSize: 14)
        label.textAlignment = .center
        label.numberOfLines = 2
        label.lineBreakMode = .byTruncatingTail
        label.backgroundColor = UIColor.clear
        return label
    }()
    
    private let imageView: UIImageView = {
        let imageView = UIImageView()
        imageView.contentMode = .scaleAspectFit
        imageView.clipsToBounds = true
        imageView.layer.cornerRadius = 5
        imageView.backgroundColor = UIColor.clear
        return imageView
    }()
    
    private let imageDescriptionLabel: UILabel = {
        let label = UILabel()
        label.textColor = UIColor.init(hex: 0x007AFF)
        label.font = UIFont.systemFont(ofSize: 13, weight: .medium)
        label.textAlignment = .left
        label.numberOfLines = 1
        label.text = "Image saved to clipboard"
        label.backgroundColor = UIColor.clear
        return label
    }()

    override init() {
        super.init()
        setupUI()
    }
    
    private func setupUI() {
        containerView.addSubview(contentView)
        containerView.addSubview(emptyTextLabel)
        containerView.addSubview(emptyImageView)
        contentView.addSubview(textLabel)
        contentView.addSubview(imageView)
        contentView.addSubview(imageDescriptionLabel)
        
        contentView.frame = CGRect(x: 10, y: 5, width: 288, height: 50)
        
        textLabel.frame = contentView.bounds
        
        emptyTextLabel.isHidden = true
        emptyImageView.isHidden = true
        textLabel.isHidden = true
        imageView.isHidden = true
        imageDescriptionLabel.isHidden = true
        
        updateUI()
    }
    
    func updateStatus(contentType: PIPContentType, text: String? = nil, image: UIImage? = nil) {
        Logger.info("[PIPStatusView] updateStatus called with contentType: \(contentType)")
        self.contentType = contentType
        self.receivedText = text
        self.receivedImage = image
        
        DispatchQueue.main.async { [weak self] in
            Logger.info("[PIPStatusView] About to call updateUI on main thread")
            self?.updateUI()
        }
    }
    
    private func updateUI() {
        Logger.info("[PIPStatusView] updateUI called with contentType: \(contentType)")
        
        emptyImageView.isHidden = true
        emptyTextLabel.isHidden = true
        textLabel.isHidden = true
        imageView.isHidden = true
        imageDescriptionLabel.isHidden = true
        
        switch contentType {
        case .idle:
            Logger.info("[PIPStatusView] Processing .idle state")
            emptyImageView.image = UIImage(named: "noClipboard")
            emptyImageView.isHidden = false
            emptyImageView.frame = CGRect(x: 20, y: (containerView.bounds.height - 36) / 2, width: 36, height: 36)
            emptyTextLabel.text = "Clipboard is currently empty"
            emptyTextLabel.isHidden = false
            emptyTextLabel.frame = CGRect(x: emptyImageView.frame.maxX, y: (contentView.bounds.size.height - 20) / 2,
                                          width: contentView.bounds.width - emptyImageView.frame.maxX - 8, height: 20)
            emptyTextLabel.center.y = emptyImageView.center.y
            break
            
        case .textReceived:
            Logger.info("[PIPStatusView] Processing .textReceived state")
            if let text = receivedText, !text.isEmpty {
                textLabel.text = text
                textLabel.isHidden = false
                setupTextLabelForContent(text)
            }
            
        case .imageReceived:
            Logger.info("[PIPStatusView] Processing .imageReceived state")
            if let image = receivedImage {
                imageView.image = image
                imageView.isHidden = false
                imageDescriptionLabel.isHidden = false
                setupImageViewForContent(image)
            }
        }
        Logger.info("[PIPStatusView] updateUI completed")
    }
    
    private func setupTextLabelForContent(_ text: String) {
        let maxWidth = contentView.bounds.width
        let font = textLabel.font ?? UIFont.systemFont(ofSize: 12)
        
        let singleLineSize = text.size(withAttributes: [.font: font])
        
        if singleLineSize.width <= maxWidth {
            textLabel.numberOfLines = 1
        } else {
            textLabel.numberOfLines = 2
        }
        textLabel.textAlignment = .left
        textLabel.frame = contentView.bounds
    }
    
    private func setupImageViewForContent(_ image: UIImage) {
        let containerSize = contentView.bounds.size
        
        let descriptionWidth = containerSize.width * 0.6 // 60% width for text
        let imageWidth = containerSize.width * 0.4 - 8   // 40% width for image, minus spacing
        
        imageDescriptionLabel.frame = CGRect(
            x: 0,
            y: (containerSize.height - 20) / 2, // Center align
            width: descriptionWidth,
            height: 20
        )
        
        let imageHeight = containerSize.height
        let imageSize = image.size
        let aspectRatio = imageSize.width / imageSize.height
        
        var newImageSize: CGSize
        newImageSize = CGSize(width: imageHeight * aspectRatio, height: imageHeight)
        if newImageSize.width > imageWidth {
            newImageSize = CGSize(width: imageWidth, height: imageWidth / aspectRatio)
        }
        
        let imageX = containerSize.width - newImageSize.width
        let imageY = (containerSize.height - newImageSize.height) / 2
        
        imageView.frame = CGRect(
            x: imageX,
            y: imageY,
            width: newImageSize.width,
            height: newImageSize.height
        )
        imageView.contentMode = .scaleAspectFit
    }
    
    func nextBuffer() -> UIImage {
        return containerView.uiImage
    }
    
    func start() {
        bufferDisplayLayer.backgroundColor = UIColor.clear.cgColor
        bufferDisplayLayer.videoGravity = .resizeAspectFill
        
        let timerBlock: ((Timer) -> Void) = { [weak self] timer in
            guard let self = self else { return }
            
            if self.bufferDisplayLayer.status == .failed {
                Logger.info("Buffer display layer failed, flushing...")
                self.bufferDisplayLayer.flush()
                return
            }
            
            guard let buffer = self.nextBuffer().cmSampleBuffer else {
                Logger.info("Failed to create sample buffer from image")
                return
            }
            
            self.bufferDisplayLayer.enqueue(buffer)
        }
        
        timer = Timer(timeInterval: 1.0/15.0, repeats: true, block: timerBlock) // 15 FPS
        RunLoop.main.add(timer, forMode: .default)
    }
    
    func stop() {
        if timer != nil {
            timer.invalidate()
            timer = nil
        }
    }
    
    func isRunning() -> Bool {
        return timer != nil
    }
    
    deinit {
        stop()
    }
}
