#!/bin/sh

mainFile="main\lanServer_main.go"
version="2.0.0"
buildDate=$(date "+%Y-%m-%dT%H:%M:%S")
serverName="lanServer"

isHiddenWin=0
if [ $isHiddenWin -eq 0 ]; then
  ldflags="-X 'rtk-cross-share/lanServer/buildConfig.Version=$version' -X 'rtk-cross-share/lanServer/buildConfig.BuildDate=$buildDate' -X 'rtk-cross-share/lanServer/buildConfig.ServerName=$serverName' "
else
  ldflags="-X 'rtk-cross-share/lanServer/buildConfig.Version=$version' -X 'rtk-cross-share/lanServer/buildConfig.BuildDate=$buildDate' -X 'rtk-cross-share/lanServer/buildConfig.ServerName=$serverName' -H=windowsgui"
fi

echo "Compile Start"
cd "lanServer"
go build  -ldflags "$ldflags"  -o $serverName".exe"  $mainFile
cd ..
echo "Compile Done"
