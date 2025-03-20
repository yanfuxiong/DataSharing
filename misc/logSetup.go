package misc

import (
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"log"
	"os"
)

var (
	LogPath         string
	CrashLogPath    string
	LoggerWriteFile lumberjack.Logger
)

func InitLog(logPath, crashLogPath string) {
	LogPath = logPath
	CrashLogPath = crashLogPath

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Printf("InitLog logPath:[%s] crashLogPath:[%s] !", LogPath, CrashLogPath)
}

func SetupLogFile() {
	log.Printf("set log written to the file:[%s] !\n", LogPath)

	LoggerWriteFile.Close()
	LoggerWriteFile = lumberjack.Logger{
		Filename:   LogPath,
		MaxSize:    256,
		MaxBackups: 3,
		MaxAge:     30,
		Compress:   true,
	}
	log.SetOutput(&LoggerWriteFile)
}

func SetupLogShut() {
	log.Println("set log shut down !\n")
	LoggerWriteFile.Close()
	log.SetOutput(io.Discard)
}

func SetupLogConsole() {
	log.Println("Set log printed to the console !\n")
	LoggerWriteFile.Close()
	log.SetOutput(os.Stdout)
}

func SetupLogConsoleFile() {
	log.Printf("Set log written to the file:[%s] and printed to the console!\n", LogPath)
	LoggerWriteFile.Close()
	LoggerWriteFile = lumberjack.Logger{
		Filename:   LogPath,
		MaxSize:    256,
		MaxBackups: 3,
		MaxAge:     30,
		Compress:   true,
	}

	log.SetOutput(io.MultiWriter(os.Stdout, &LoggerWriteFile))
}
