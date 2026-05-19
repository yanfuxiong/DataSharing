//
//  TransferListView.swift
//  CrossShare
//
//  Created on 2025/1/XX.
//

import Cocoa
import SnapKit

class TransferListView: NSView {
    
    // Subviews - these will be set from MainHomeViewController
    var bottomScrollView: NSScrollView!
    
    override init(frame frameRect: NSRect) {
        super.init(frame: frameRect)
    }
    
    required init?(coder: NSCoder) {
        fatalError("init(coder:) has not been implemented")
    }
    
    func setupUI() {
        guard bottomScrollView != nil else {
            return
        }
        
        addSubview(bottomScrollView)
        
        // bottomScrollView constraints - same as original layout but relative to TransferListView
        bottomScrollView.snp.makeConstraints { make in
            make.leading.equalToSuperview().offset(20)
            make.trailing.equalToSuperview().offset(-20)
            make.top.equalToSuperview()
            make.bottom.equalToSuperview().offset(-20)
        }
    }
}
