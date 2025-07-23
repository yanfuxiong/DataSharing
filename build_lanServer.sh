#!/bin/sh

mainFile="./main"
version="2.1.8"
buildDate=$(date "+%Y-%m-%dT%H:%M:%S")
serverName="cross_share_lan_serv"
libName="librtk_cross_share_lan_serv.so"

ldflags="-X rtk-cross-share/lanServer/buildConfig.Version=$version -X rtk-cross-share/lanServer/buildConfig.BuildDate=$buildDate -X rtk-cross-share/lanServer/buildConfig.ServerName=$serverName -s -w -extldflags=-Wl,-soname,$libName"

echo "Compile Start"
cd "lanServer"
export CC=armv7a-linux-androideabi34-clang
export CXX=armv7a-linux-androideabi34-clang++
export CGO_ENABLED=1
export GOOS=android
export GOARCH=arm
export GOARM=7
go build -buildmode=c-shared -trimpath -ldflags "$ldflags" -o $libName $mainFile
cd ..
echo "Compile Done"