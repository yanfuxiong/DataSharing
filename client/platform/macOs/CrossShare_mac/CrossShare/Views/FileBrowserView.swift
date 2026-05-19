//
//  FileBrowserView.swift
//  CrossShare
//
//  Created on 2025/1/XX.
//

import Cocoa
import SnapKit

class FileBrowserView: NSView {
    
    // Subviews - these will be set from MainHomeViewController
    var combinedBorderView: NSView!
    var leftScrollView: NSScrollView!
    var rightHeaderContainer: NSView!
    var rightScrollView: NSScrollView!
    
    override init(frame frameRect: NSRect) {
        super.init(frame: frameRect)
    }
    
    required init?(coder: NSCoder) {
        fatalError("init(coder:) has not been implemented")
    }
    
    func setupUI() {
        guard combinedBorderView != nil && leftScrollView != nil && rightHeaderContainer != nil && rightScrollView != nil else {
            return
        }
        
        addSubview(combinedBorderView)
        addSubview(leftScrollView)
        addSubview(rightHeaderContainer)
        addSubview(rightScrollView)
        
        // combinedBorderView constraints
        combinedBorderView.snp.makeConstraints { make in
            make.top.equalToSuperview().offset(16)
            make.leading.equalToSuperview().offset(16)
            make.trailing.equalToSuperview().offset(-16)
            make.bottom.equalToSuperview().offset(-4)
        }
        
        // leftScrollView constraints
        leftScrollView.snp.makeConstraints { make in
            make.top.equalToSuperview().offset(20)
            make.leading.equalToSuperview().offset(20)
            make.width.equalTo(200)
            make.bottom.equalToSuperview().offset(-8)
        }
        
        // rightHeaderContainer constraints
        rightHeaderContainer.snp.makeConstraints { make in
            make.top.equalToSuperview().offset(20)
            make.leading.equalTo(leftScrollView.snp.trailing).offset(20)
            make.trailing.equalToSuperview().offset(-20)
            make.height.equalTo(28)
        }
        
        // rightScrollView constraints
        rightScrollView.snp.makeConstraints { make in
            make.top.equalTo(rightHeaderContainer.snp.bottom).offset(10)
            make.leading.equalTo(leftScrollView.snp.trailing).offset(20)
            make.trailing.equalToSuperview().offset(-20)
            make.bottom.equalTo(leftScrollView.snp.bottom)
        }
    }
}
