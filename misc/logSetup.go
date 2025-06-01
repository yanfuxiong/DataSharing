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

	maxSize    int
	maxBackups int
	maxAge     int
)

func init() {
	maxSize = 256
	maxBackups = 3
	maxAge = 30
}

func InitLog(logPath, crashLogPath string, maxsize int) {
	LogPath = logPath
	CrashLogPath = crashLogPath

	if maxsize > 0 {
		maxSize = maxsize
	}

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Printf("InitLog logPath:[%s] maxSize:[%d] maxBackups:[%d] maxAge:[%d]!", LogPath, maxSize, maxBackups, maxAge)
	log.Printf("InitLog crashLogPath:[%s]", CrashLogPath)
}

func SetupLogFile() {
	log.Printf("set log written to the file:[%s] !\n", LogPath)

	LoggerWriteFile.Close()
	LoggerWriteFile = lumberjack.Logger{
		Filename:   LogPath,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
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
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   true,
	}

	log.SetOutput(io.MultiWriter(os.Stdout, &LoggerWriteFile))
}

func SetLogRotate() {
	log.Printf("Set log file Rotate!\n")
	LoggerWriteFile.Rotate()
}
