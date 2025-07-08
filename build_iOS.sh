#!/bin/bash

version="2.1.22"
buildDate=$(date "+%Y-%m-%dT%H:%M:%S")
ldflags="-X rtk-cross-share/client/buildConfig.Version=$version -X rtk-cross-share/client/buildConfig.BuildDate=$buildDate -X rtk-cross-share/client/buildConfig.Debug=1 -s -w -extldflags=-lresolv"
cd ./client/platform/iOS

getFileName() {
    if [[ -z $3 ]]; then
        echo "$1.$2"
    else
        echo "$1_$2.$3"
    fi
}

buildFolder="./build/libs"
arm="arm64"
a="a"
h="h"
fileName="libcross_share"
mainGo="./main.go"

rm -rf $buildFolder
mkdir -p $buildFolder

echo "Building for device (arm64)..."
export GOARCH=arm64
export GOOS=ios
export CGO_ENABLED=1
export CGO_CFLAGS="-arch arm64 -miphoneos-version-min=9.0 -isysroot "$(xcrun -sdk iphoneos --show-sdk-path)
export CGO_LDFLAGS="-arch arm64 -miphoneos-version-min=9.0 -isysroot "$(xcrun -sdk iphoneos --show-sdk-path)
export CC="$(xcrun -sdk iphoneos -find clang)"
export CXX="$(xcrun -sdk iphoneos -find clang++)"

go clean -cache
go build -tags ios -buildmode=c-archive -trimpath -o "$buildFolder/$(getFileName $fileName $arm $a)" -ldflags "$ldflags" "$mainGo"

cd $buildFolder

cp "$(getFileName $fileName $arm $a)" "$(getFileName $fileName $a)"
cp "$(getFileName $fileName $arm $h)" "$(getFileName $fileName $h)"

rm "$(getFileName $fileName $arm $a)"
rm "$(getFileName $fileName $arm $h)"

echo "Build completed successfully!"
echo "Static library: $buildFolder/$(getFileName $fileName $a)"
echo "Header file: $buildFolder/$(getFileName $fileName $h)"
