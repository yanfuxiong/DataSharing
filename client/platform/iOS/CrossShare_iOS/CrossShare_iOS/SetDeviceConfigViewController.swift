//
//  SetDeviceConfigViewController.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/6/17.
//

import UIKit
import MBProgressHUD

class SetDeviceConfigViewController: BaseViewController {
    
    override func viewDidLoad() {
        super.viewDidLoad()
        
        setupUI()
    }
    
    func setupUI() {
        self.view.addSubview(self.configContenView)
        
        self.configContenView.snp.makeConstraints { make in
            make.left.right.equalToSuperview()
            make.top.equalTo(self.view.safeAreaLayoutGuide.snp.top)
            make.bottom.equalToSuperview()
        }
        
        self.configContenView.refreshUI()
        self.configContenView.submitConfigBlock = {[weak self] (ddcciText,deviceSourceText,devicePortText) in
            guard let self = self else { return  }
            guard ddcciText.isEmpty == false else {
                MBProgressHUD.showTips(.error,"Please enter device ddcci id", toView: self.view)
                return
            }
            guard deviceSourceText.isEmpty == false else {
                MBProgressHUD.showTips(.error,"Please enter device source", toView: self.view)
                return
            }
            guard devicePortText.isEmpty == false else {
                MBProgressHUD.showTips(.error,"Please enter device port", toView: self.view)
                return
            }
            if let intSource = Int(deviceSourceText), let intPort = Int(devicePortText) {
                UserDefaults.set(forKey: .DEVICECONFIG_DIAS_ID, value: ddcciText)
                UserDefaults.setInt(forKey: .DEVICECONFIG_SRC, value: intSource)
                UserDefaults.setInt(forKey: .DEVICECONFIG_PORT, value: intPort)
                P2PManager.shared.setupDeviceConfig(ddcciText, intSource, intPort)
                self.configContenView.refreshUI()
            }
        }
        
        self.view.setNeedsLayout()
        self.view.layoutIfNeeded()
    }
    
    lazy var configContenView: DeviceConfigView = {
        let view = DeviceConfigView(frame: .zero)
        view.backgroundColor = UIColor.clear
        view.isUserInteractionEnabled = true
        return view
    }()
    
    
    /*
     // MARK: - Navigation
     
     // In a storyboard-based application, you will often want to do a little preparation before navigation
     override func prepare(for segue: UIStoryboardSegue, sender: Any?) {
     // Get the new view controller using segue.destination.
     // Pass the selected object to the new view controller.
     }
     */
    
}
