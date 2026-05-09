#!/bin/sh

mainFile="main.go"
version="2.2.24"
buildDate=$(date "+%Y-%m-%dT%H:%M:%S")
targetDLL_release="client_windows_release.dll"
targetDLL_debug="client_windows_debug.dll"
buildFolder="./build"

isBuildRelease=false
isBuildDebug=true

ldflags_release="-X rtk-cross-share/client/buildConfig.Version=$version -X rtk-cross-share/client/buildConfig.BuildDate=$buildDate -X rtk-cross-share/client/buildConfig.Debug=0 -s -w"
ldflags_debug="-X rtk-cross-share/client/buildConfig.Version=$version -X rtk-cross-share/client/buildConfig.BuildDate=$buildDate -X rtk-cross-share/client/buildConfig.Debug=1 -w=false "

echo "Compile Start"
cd "client/platform/windows"

rm -rf $buildFolder
mkdir -p $buildFolder

if [[ "$isBuildRelease" == "true" ]]; then
    echo "build release Dll ..."
    go build -ldflags "$ldflags_release" -buildmode=c-shared  -o "$buildFolder/$targetDLL_release"  $mainFile
    echo "Compressing release Dll with UPX ..."
    upx --best --lzma "$buildFolder/$targetDLL_release"
fi

if [[ "$isBuildDebug" == "true" ]]; then
    echo "build debug Dll ..."
    go build -ldflags "$ldflags_debug" -gcflags="all=-N -l" -buildmode=c-shared  -o "$buildFolder/$targetDLL_debug"  $mainFile
    echo "Compressing debug Dll with UPX ..."
    upx --best --lzma "$buildFolder/$targetDLL_debug"
fi

cd -
echo "Compile Done"


#下载：https://github.com/upx/upx/releases
#解压到 C:\upx
#添加到系统 PATH：
#Win+R → sysdm.cpl → 高级 → 环境变量
#编辑 Path → 新建 → C:\upx
#命令行验证：upx --version，显示版本号即为安装成功。