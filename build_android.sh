#!/bin/sh
version="2.1.32"
buildDate=$(date "+%Y-%m-%dT%H:%M:%S")

ldflags="-X 'rtk-cross-share/client/buildConfig.Version=$version' -X 'rtk-cross-share/client/buildConfig.BuildDate=$buildDate' -X 'rtk-cross-share/client/buildConfig.Debug=1' -checklinkname=0"
buildFolder="./build"

export CGO_LDFLAGS="-O2 -g -s -w -Wl,-z,max-page-size=16384"

cd client/platform/android
rm -rf $buildFolder
mkdir -p $buildFolder

gomobile bind -target=android -androidapi 21 -ldflags "$ldflags"
mv *.aar *.jar $buildFolder

cd -
echo "Compile Done"
