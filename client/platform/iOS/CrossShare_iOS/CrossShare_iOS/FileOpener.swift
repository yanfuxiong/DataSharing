//
//  FileOpener.swift
//  CrossShare_iOS
//
//  Created by jack_huang on 5/23/25.
//

import UIKit
import QuickLook

class FileOpener: NSObject, QLPreviewControllerDataSource, QLPreviewControllerDelegate {

    private var mFileURL: URL? = nil
    private weak var presenter: UIViewController?

    init(presenter: UIViewController) {
        self.presenter = presenter
    }

    private func getFilePathAndURL(_ fileName: String) -> (path: String, url: URL)? {
        let fileManager = FileManager.default
        let kDownloadFolder = "Download"
        if let documentsURL = fileManager.urls(for: .documentDirectory, in: .userDomainMask).first {
            let fileURL = documentsURL.appendingPathComponent(kDownloadFolder).appendingPathComponent(fileName)

            if fileManager.fileExists(atPath: fileURL.path) {
                return (path: fileURL.path, url: fileURL)
            } else {
                print("[FileOpener][Err] Invalid filepath: \(fileURL.path)")
                return nil
            }
        }

        return nil
    }

    func openFile(fileName: String) -> Bool {
        guard let (filePath, fileUrl) = getFilePathAndURL(fileName) else {
            print("[FileOpener][Err] Invalid file")
            return false
        }
        print("Open file: [\(filePath)]")
        self.mFileURL = fileUrl

        guard FileManager.default.fileExists(atPath: fileUrl.path),
              FileManager.default.isReadableFile(atPath: fileUrl.path) else {
            print("[FileOpener][Err] File not existed or unreadable: [\(fileUrl.path)]")
            return false
        }

        // QLPreviewController
        if QLPreviewController.canPreview(fileUrl as NSURL) {
            let ql = QLPreviewController()
            ql.dataSource = self
            ql.delegate = self

            let doneButton = UIBarButtonItem(title: "Done", style: .plain, target: self, action: #selector(dismissPresented))
            doneButton.tintColor = UIColor { trait in
                switch trait.userInterfaceStyle {
                case .dark:
                    return .white
                default:
                    return .systemBlue
                }
            }
            ql.navigationItem.rightBarButtonItem = doneButton
            let nav = UINavigationController(rootViewController: ql)
            self.presenter?.present(nav, animated: true)
        }
        // Not support in QLPreview, display Share APP
        else {
            let activityVC = UIActivityViewController(activityItems: [fileUrl], applicationActivities: nil)
            self.presenter?.present(activityVC, animated: true)
        }
        return true
    }

    @objc func dismissPresented() {
        self.presenter?.dismiss(animated: true, completion: nil)
    }

    func numberOfPreviewItems(in controller: QLPreviewController) -> Int {
        return 1
    }

    func previewController(_ controller: QLPreviewController, previewItemAt index: Int) -> QLPreviewItem {
        return mFileURL! as NSURL
    }
}
