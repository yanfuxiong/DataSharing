package misc

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

func GoSafeWithParam(fn func(args ...any), args ...any) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("crash info write into file:[%s]", CrashLogPath)
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
				os.Chmod(CrashLogPath, 0644)
				LoggerCrashWriteFile.Close()
				os.Exit(1)
			}
		}()
		fn(args...)
	}()
}

func GoSafe(fn func()) {
	GoSafeWithParam(func(args ...any) {
		fn()
	})
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

func FolderExists(folder string) bool {
	info, err := os.Stat(folder)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func FileSize(filePath string) (uint64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("File %s does not exist\n", filePath)
		} else {
			log.Printf("Getting file:[%s] info error: %+v\n", filePath, err)
		}
		return 0, err
	}

	if fileInfo.IsDir() {
		log.Printf("File %s is a directory!", filePath)
		return 0, errors.New("this is a directory")
	}

	return uint64(fileInfo.Size()), nil
}

func FileSizeDesc(size uint64) string {
	const (
		B = 1 << (10 * iota)
		KB
		MB
		GB
		TB
	)

	switch {
	case size >= TB:
		return fmt.Sprintf("%.2f TB", float64(size)/float64(TB))
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d B", size)
	}
}

func AdaptationPath(srcPath string) string {
	currentOsSeparator := string(filepath.Separator)
	replacements := make(map[string]string, 0)

	// TODO: Do different platforms need to handle it separately
	//if runtime.GOOS == "windows" {
	replacements[`"`] = ""
	replacements[`*`] = "_"
	replacements[`?`] = "_"
	replacements[`<`] = "_"
	replacements[`>`] = "_"
	replacements[`|`] = "_"
	replacements[`:`] = "_"
	replacements["/"] = currentOsSeparator
	replacements["\\"] = currentOsSeparator
	/*} else if runtime.GOOS == "android" {
	} else if runtime.GOOS == "darwin" {
		cleanedPath = strings.ReplaceAll(srcPath, ":", "_")
	} else if runtime.GOOS == "ios" {
	}*/

	for k, v := range replacements {
		srcPath = strings.ReplaceAll(srcPath, k, v)
	}

	return collapseSeparators(srcPath, currentOsSeparator)
}

func collapseSeparators(path, separator string) string {
	doubleSeparator := separator + separator
	for strings.Contains(path, doubleSeparator) {
		path = strings.ReplaceAll(path, doubleSeparator, separator)
	}
	return path
}

func ConcatIP(ip string, port string) string {
	publicIP := fmt.Sprintf("%s:%s", ip, port)
	return publicIP
}

func IsInTheList(target string, list []string) bool {
	for _, item := range list {
		if strings.EqualFold(item, target) {
			return true
		}
	}
	return false
}

func RemoveStringFromSlice(slice []string, s string) []string {
	i := 0
	for _, v := range slice {
		if v != s {
			slice[i] = v
			i++
		}
	}
	return slice[:i]
}

func CreateDir(dir string, dirModeOpt ...os.FileMode) error {
	dirMode := os.ModePerm
	if len(dirModeOpt) > 0 {
		dirMode = dirModeOpt[0]
	}

	_, err := os.Stat(dir)

	if os.IsNotExist(err) {
		err = os.MkdirAll(dir, dirMode)
		if err != nil {
			return fmt.Errorf("failed to create directory:[%s] error: %+v", dir, err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check directory[%s] error: %+v", dir, err)
	}
	return nil
}
