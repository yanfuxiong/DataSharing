package main

import (
	"log"
	rtkBuildConfig "rtk-cross-share/buildConfig"
	rtkCmd "rtk-cross-share/cmd"
)

func main() {
	log.Println("========================")
	log.Println("Version: ", rtkBuildConfig.Version)
	log.Println("Build Date: ", rtkBuildConfig.BuildDate)
	log.Printf("========================\n\n")

	rtkCmd.Run()
}
