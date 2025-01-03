#!/bin/sh
version="2.0.11"
buildDate=$(date "+%Y-%m-%dT%H:%M:%S")
platform="android"
ldflags="-X 'rtk-cross-share/buildConfig.Version=$version' -X 'rtk-cross-share/buildConfig.BuildDate=$buildDate' -X 'rtk-cross-share/buildConfig.Platform=$platform' -checklinkname=0"

cd client/platform/libp2p_clipboard
gomobile bind -target=android -androidapi 21 -ldflags "$ldflags"
cd -
