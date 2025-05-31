#!/bin/sh

mainFile="main/lanServer_main.go"
version="2.1.5"
buildDate=$(date "+%Y-%m-%dT%H:%M:%S")
serverName="cross_share_lan_serv"

ldflags="-X 'rtk-cross-share/lanServer/buildConfig.Version=$version' -X 'rtk-cross-share/lanServer/buildConfig.BuildDate=$buildDate' -X 'rtk-cross-share/lanServer/buildConfig.ServerName=$serverName'"

echo "Compile Start"
cd "lanServer"
export CC=armv7a-linux-androideabi34-clang
export CXX=armv7a-linux-androideabi34-clang++
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=arm
export GOARM=7
go build -ldflags "$ldflags" -o $serverName  $mainFile
cd ..
echo "Compile Done"