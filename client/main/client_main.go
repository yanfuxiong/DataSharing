package main

import (
	rtkBuildConfig "rtk-cross-share/client/buildConfig"
	rtkCmd "rtk-cross-share/client/cmd"
)

func main() {
	rtkBuildConfig.CmdDebug = "1"
	rtkCmd.Run()
}
