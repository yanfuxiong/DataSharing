//
//  PrivacyViewController.swift
//  CrossShare_iOS
//
//  Created by ts on 2025/6/16.
//

import UIKit

struct licenseModel {
    var tittle:String = ""
    var content:String = ""
    init(tittle: String, content: String) {
        self.tittle = tittle
        self.content = content
    }
}

class PrivacyViewController: BaseViewController {
    
    override func viewDidLoad() {
        super.viewDidLoad()
        setupUI()
        initialize()
    }
    
    func setupUI() {
        self.title = "Cross Share"
        self.view.addSubview(self.deviceView)
        
        self.deviceView.snp.makeConstraints { make in
            make.left.right.equalToSuperview()
            make.top.equalTo(self.view.safeAreaLayoutGuide.snp.top)
            make.bottom.equalToSuperview()
        }
        self.deviceView.dataArray = self.dataSource()
        self.deviceView.didSelectBlock = { [weak self] license in
            guard let self = self else { return }
            let vc = LicenseViewViewController()
            vc.licenseText = license.content
            vc.licenseTitle = license.tittle
            self.navigationController?.pushViewController(vc, animated: true)
        }
    }
    
    func initialize() {
        
    }
    
    private func dataSource() -> [licenseModel] {
        var licenses: [licenseModel] = []
        let alamofirePolicy = licenseModel(tittle: "Alamofire", content:
                                                  """
                                                  Copyright (c) 2014-2022 Alamofire Software Foundation (http://alamofire.org/)
                                                  
                                                  Permission is hereby granted, free of charge, to any person obtaining a copy
                                                  of this software and associated documentation files (the "Software"), to deal
                                                  in the Software without restriction, including without limitation the rights
                                                  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
                                                  copies of the Software, and to permit persons to whom the Software is
                                                  furnished to do so, subject to the following conditions:
                                                  
                                                  The above copyright notice and this permission notice shall be included in
                                                  all copies or substantial portions of the Software.
                                                  
                                                  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
                                                  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
                                                  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
                                                  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
                                                  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
                                                  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
                                                  THE SOFTWARE.
                                                  """
        )
        licenses.append(alamofirePolicy)
        
        let mBProgressHUDPolicy = licenseModel(tittle: "MBProgressHUD", content:
                                                  """
                                                  Copyright © 2009-2020 Matej Bukovinski
                                                  
                                                  Permission is hereby granted, free of charge, to any person obtaining a copy
                                                  of this software and associated documentation files (the "Software"), to deal
                                                  in the Software without restriction, including without limitation the rights
                                                  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
                                                  copies of the Software, and to permit persons to whom the Software is
                                                  furnished to do so, subject to the following conditions:
                                                  
                                                  The above copyright notice and this permission notice shall be included in
                                                  all copies or substantial portions of the Software.
                                                  
                                                  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
                                                  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
                                                  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
                                                  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
                                                  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
                                                  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
                                                  THE SOFTWARE.
                                                  """
        )
        licenses.append(mBProgressHUDPolicy)
        
        let swiftyJSONPolicy = licenseModel(tittle: "SwiftyJSON", content:
                                                  """
                                                  The MIT License (MIT)
                                                  
                                                  Copyright (c) 2017 Ruoyu Fu
                                                  
                                                  Permission is hereby granted, free of charge, to any person obtaining a copy
                                                  of this software and associated documentation files (the "Software"), to deal
                                                  in the Software without restriction, including without limitation the rights
                                                  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
                                                  copies of the Software, and to permit persons to whom the Software is
                                                  furnished to do so, subject to the following conditions:
                                                  
                                                  The above copyright notice and this permission notice shall be included in
                                                  all copies or substantial portions of the Software.
                                                  
                                                  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
                                                  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
                                                  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
                                                  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
                                                  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
                                                  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
                                                  THE SOFTWARE.
                                                  """
        )
        licenses.append(swiftyJSONPolicy)
        let snapKitPolicy = licenseModel(tittle: "SnapKit", content:
                                                  """
                                                  Copyright (c) 2011-Present SnapKit Team - https://github.com/SnapKit
                                                  
                                                  Permission is hereby granted, free of charge, to any person obtaining a copy
                                                  of this software and associated documentation files (the "Software"), to deal
                                                  in the Software without restriction, including without limitation the rights
                                                  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
                                                  copies of the Software, and to permit persons to whom the Software is
                                                  furnished to do so, subject to the following conditions:
                                                  
                                                  The above copyright notice and this permission notice shall be included in
                                                  all copies or substantial portions of the Software.
                                                  
                                                  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
                                                  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
                                                  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
                                                  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
                                                  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
                                                  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
                                                  THE SOFTWARE.
                                                  """
        )
        licenses.append(snapKitPolicy)
        let sWCompressionPolicy = licenseModel(tittle: "SWCompression", content:
                                                  """
                                                  MIT License

                                                  Copyright (c) 2024 Timofey Solomko

                                                  Permission is hereby granted, free of charge, to any person obtaining a copy
                                                  of this software and associated documentation files (the "Software"), to deal
                                                  in the Software without restriction, including without limitation the rights
                                                  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
                                                  copies of the Software, and to permit persons to whom the Software is
                                                  furnished to do so, subject to the following conditions:

                                                  The above copyright notice and this permission notice shall be included in all
                                                  copies or substantial portions of the Software.

                                                  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
                                                  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
                                                  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
                                                  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
                                                  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
                                                  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
                                                  SOFTWARE.
                                                  """
        )
        licenses.append(sWCompressionPolicy)
        let bitByteDataPolicy = licenseModel(tittle: "BitByteData", content:
                                                  """
                                                  MIT License

                                                  Copyright (c) 2024 Timofey Solomko

                                                  Permission is hereby granted, free of charge, to any person obtaining a copy
                                                  of this software and associated documentation files (the "Software"), to deal
                                                  in the Software without restriction, including without limitation the rights
                                                  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
                                                  copies of the Software, and to permit persons to whom the Software is
                                                  furnished to do so, subject to the following conditions:

                                                  The above copyright notice and this permission notice shall be included in all
                                                  copies or substantial portions of the Software.

                                                  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
                                                  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
                                                  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
                                                  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
                                                  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
                                                  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
                                                  SOFTWARE.
                                                  """
        )
        licenses.append(bitByteDataPolicy)
        return licenses
    }
    
    lazy var deviceView: LicenseView = {
        let view = LicenseView(frame: .zero)
        view.backgroundColor = UIColor.clear
        view.isUserInteractionEnabled = true
        return view
    }()
    
}
