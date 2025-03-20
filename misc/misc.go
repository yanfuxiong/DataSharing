package misc

import (
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
)

func GoSafe(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				LoggerWriteFile.Close()

				LoggerCrashWriteFile := lumberjack.Logger{
					Filename:   CrashLogPath,
					MaxSize:    256,
					MaxBackups: 3,
					MaxAge:     30,
					Compress:   true,
				}
				log.SetOutput(&LoggerCrashWriteFile)

				log.Printf("Recovered from panic: %v\n", r)
				log.Printf("Stack trace:\n%s", debug.Stack())

				LoggerCrashWriteFile.Close()
				os.Exit(1)
			}
		}()
		fn()
	}()
}

func GetFuncName() string {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return "UnknownFunction"
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "UnknownFunction"
	}
	fullName := fn.Name()
	if fullName == "" {
		return "UnknownFunction"
	}
	parts := strings.Split(fullName, ".")
	if len(parts) == 0 {
		return "UnknownFunction"
	}
	return parts[len(parts)-1]
}

func GetLine() int {
	_, _, line, ok := runtime.Caller(1)
	if !ok {
		return -1
	}
	return line
}

func GetFuncInfo() string {
	funcName := "UnknownFunc:"
	pc, _, line, ok := runtime.Caller(1)
	if !ok {
		return funcName + strconv.Itoa(line)
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return funcName + strconv.Itoa(line)
	}
	fullName := fn.Name()
	if fullName == "" {
		return funcName + strconv.Itoa(line)
	}
	parts := strings.Split(fullName, ".")
	if len(parts) == 0 {
		return funcName + strconv.Itoa(line)
	}
	funcName = parts[len(parts)-1] + ":"
	return funcName + strconv.Itoa(line)
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func ConcatIP(ip string, port string) string {
	publicIP := fmt.Sprintf("%s:%s", ip, port)
	return publicIP
}
