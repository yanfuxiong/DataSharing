#!/bin/sh
version="2.1.24"
buildDate=$(date "+%Y-%m-%dT%H:%M:%S")

ldflags="-X 'rtk-cross-share/client/buildConfig.Version=$version' -X 'rtk-cross-share/client/buildConfig.BuildDate=$buildDate' -X 'rtk-cross-share/client/buildConfig.Debug=1' -checklinkname=0"

cd client/platform/libp2p_clipboard
gomobile bind -target=android -androidapi 21 -ldflags "$ldflags"
cd -
echo "Compile Done"
