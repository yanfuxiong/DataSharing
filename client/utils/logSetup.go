package utils

import (
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"log"
	"os"
	rtkGlobal "rtk-cross-share/global"
)

func SetupLogFile() {
	log.Printf("set log written to the file:[%s] !\n", rtkGlobal.LogPath)

	rtkGlobal.LoggerWriteFile.Close()
	rtkGlobal.LoggerWriteFile = lumberjack.Logger{
		Filename:   rtkGlobal.LogPath,
		MaxSize:    256,
		MaxBackups: 3,
		MaxAge:     30,
		Compress:   true,
	}
	log.SetOutput(&rtkGlobal.LoggerWriteFile)
}

func SetupLogShut() {
	log.Println("set log shut down !\n")
	rtkGlobal.LoggerWriteFile.Close()
	log.SetOutput(io.Discard)
}

func SetupLogConsole() {
	log.Println("Set log printed to the console !\n")
	rtkGlobal.LoggerWriteFile.Close()
	log.SetOutput(os.Stdout)
}

func SetupLogConsoleFile() {
	log.Printf("Set log written to the file:[%s] and printed to the console!\n", rtkGlobal.LogPath)
	rtkGlobal.LoggerWriteFile.Close()
	rtkGlobal.LoggerWriteFile = lumberjack.Logger{
		Filename:   rtkGlobal.LogPath,
		MaxSize:    256,
		MaxBackups: 3,
		MaxAge:     30,
		Compress:   true,
	}

	log.SetOutput(io.MultiWriter(os.Stdout, &rtkGlobal.LoggerWriteFile))
}
