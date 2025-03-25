# DataSharing
数据共享项目

#on Android
#Step1: sh ./build_android.sh

#on macOs
#Step1: sh ./build_macOs.sh
#Step2: ./client_mac

#on Windows
#Step1: ./build_windows.ps1
#Step2: ./client_windows



windows PowerShell  build:  &".\build_windows.ps1"                run on windows
windows GitBash     build:  ./build_android.sh                    run on android
windows GitBash     build:  ./build_lanServer_windows.sh          run on windows
windows GitBash     build:  ./build_lanServer_android.sh          run on android

