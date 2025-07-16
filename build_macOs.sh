#!/bin/sh
version="2.1.18"
buildDate=$(date "+%Y-%m-%dT%H:%M:%S")
ldflags="-X rtk-cross-share/client/buildConfig.Version=$version -X rtk-cross-share/client/buildConfig.BuildDate=$buildDate -X rtk-cross-share/client/buildConfig.Debug=1 -s -w -extldflags=-lresolv"
cd ./client/platform/macOs

getFileName() {
    if [[ -z $3 ]]; then
        echo "$1.$2"
    else
        echo "$1_$2.$3"
    fi
}

buildFolder="./build/libs"
arm="arm64"
x86_64="x86_64"
a="a"
h="h"
fileName="libcross_share"
mainGo="./main.go"

export GO111MODULE=on

# build arm for real device
export GOARCH=arm64
export GOOS=darwin
export CGO_ENABLED=1
export CGO_CFLAGS="-mmacosx-version-min=10.13"
export CGO_LDFLAGS="-mmacosx-version-min=10.13"

printf "macOS lib compiling...\n"
go clean -cache
go build -buildmode=c-archive -trimpath -o "$buildFolder/$(getFileName $fileName $arm $a)" -ldflags "$ldflags" "$mainGo"

# bulid x86_64 for simulator
export GOARCH=amd64
export GOOS=darwin
export CGO_ENABLED=1
export CGO_CFLAGS="-mmacosx-version-min=10.13"
export CGO_LDFLAGS="-mmacosx-version-min=10.13"

printf "Simulator lib compiling...\n"
go clean -cache
go build -buildmode=c-archive -trimpath -o "$buildFolder/$(getFileName $fileName $x86_64 $a)" -ldflags "$ldflags" "$mainGo"

# build common binary
printf "Common lib compiling...\n"
cd $buildFolder
lipo -create "$(getFileName $fileName $arm $a)" $(getFileName $fileName $x86_64 $a) -output $(getFileName $fileName $a)
# copy header as common
cp $(getFileName $fileName $arm $h) $(getFileName $fileName $h)
# remove compiled files
rm $(getFileName $fileName $arm $a)
rm $(getFileName $fileName $arm $h)
rm $(getFileName $fileName $x86_64 $a)
rm $(getFileName $fileName $x86_64 $h)

cd -
