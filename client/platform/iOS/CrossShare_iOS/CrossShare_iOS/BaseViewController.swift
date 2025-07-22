//
//  BaseViewController.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/6/16.
//

import UIKit

class BaseViewController: UIViewController {

    override func viewDidLoad() {
        super.viewDidLoad()

        self.view.backgroundColor = UIColor.init(hex: 0xFFFFFF)
    }
    
    override func present(_ viewControllerToPresent: UIViewController, animated flag: Bool, completion: (() -> Void)? = nil) {
            viewControllerToPresent.modalPresentationStyle = .fullScreen
            super.present(viewControllerToPresent, animated: flag, completion: completion)
        }

    deinit {
        Logger.info(String(describing: object_getClass(self)))
    }
    
}
