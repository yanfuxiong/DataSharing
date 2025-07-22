//
//  LicenseViewViewController.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/6/20.
//

import UIKit

class LicenseViewViewController: BaseViewController {
    
    var licenseText: String = "" {
        didSet {
            DispatchQueue.main.async { [weak self] in
                guard let self = self else { return }
                if self.isViewLoaded {
                    self.updateTextContent()
                }
            }
        }
    }
    
    var licenseTitle: String = "" {
        didSet {
            self.title = licenseTitle
        }
    }
    
    lazy var textView: UITextView = {
        let view = UITextView(frame: .zero)
        view.backgroundColor = UIColor.clear
        view.isEditable = false
        view.textColor = UIColor.black
        view.font = UIFont.systemFont(ofSize: 14)
        view.textAlignment = .left
        view.textContainerInset = UIEdgeInsets(top: 20, left: 20, bottom: 20, right: 20)
        view.showsVerticalScrollIndicator = true
        view.alwaysBounceVertical = true
        return view
    }()
    
    override func viewDidLoad() {
        super.viewDidLoad()
        
        self.view.backgroundColor = UIColor.white
        self.view.addSubview(textView)
        
        self.textView.snp.makeConstraints { make in
            make.edges.equalTo(view.safeAreaLayoutGuide)
        }
        
        updateTextContent()
    }
    
    private func updateTextContent() {
        guard !licenseText.isEmpty else { return }
        
        let processedText = processLicenseText(licenseText)
        
        if processedText.contains("<") && processedText.contains(">") {
            setHTMLText(processedText)
        } else {
            textView.text = processedText
        }
    }
    
    private func processLicenseText(_ text: String) -> String {
        var processedText = text
        processedText = processedText.replacingOccurrences(of: "\\n", with: "\n")
        processedText = processedText.replacingOccurrences(of: "\\t", with: "\t")
        processedText = processedText.replacingOccurrences(of: "\\\"", with: "\"")
        return processedText
    }
    
    private func setHTMLText(_ htmlString: String) {
        guard let data = htmlString.data(using: .utf8) else {
            textView.text = htmlString
            return
        }
        
        do {
            let attributedString = try NSAttributedString(
                data: data,
                options: [
                    .documentType: NSAttributedString.DocumentType.html,
                    .characterEncoding: String.Encoding.utf8.rawValue
                ],
                documentAttributes: nil
            )
            textView.attributedText = attributedString
        } catch {
            Logger.info("Failed to parse HTML: \(error)")
            textView.text = htmlString
        }
    }
}
