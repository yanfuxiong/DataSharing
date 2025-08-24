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
fi

if [[ "$isBuildDebug" == "true" ]]; then
    echo "build debug Dll ..."
    go build -ldflags "$ldflags_debug" -gcflags="all=-N -l" -buildmode=c-shared  -o "$buildFolder/$targetDLL_debug"  $mainFile
fi

cd -
echo "Compile Done"
