package debug

import (
	"bufio"
	"fmt"
	"os"
	rtkPlatform "rtk-cross-share/platform"
	"strconv"
	"strings"
	"time"
)

type TestCase struct {
	FileName     string
	FileSizeHigh uint32
	FileSizeLow  uint32
}

func DebugCmdLine() {

	// test_case_5 := TestCase{
	// 	FileName:     "/Users/hp/myGolang/test.mp4",
	// 	FileSizeHigh: 0,
	// 	FileSizeLow:  8986659,
	// }

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter text to debug:")
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println("You entered:", line)
		if strings.Contains(line, "PIPE_UPDATE_CLIENT") {
			var status uint32 = 1
			ip := "192.168.30.1:12345"
			id := "QmQ7obXFx1XMFr6hCYXtovn9zREFqSXEtH5hdtpBDLjrAz"
			name := "jack_huang123"
			rtkPlatform.GoUpdateClientStatus(status, ip, id, name)
		} else if strings.Contains(line, "PIPE_SETUP_FILE_DROP") {
			ip := "192.168.30.1:12345"
			id := "QmQ7obXFx1XMFr6hCYXtovn9zREFqSXEtH5hdtpBDLjrAz"
			// fileName := "D:\\jack_huang\\Downloads\\ÐÂÔöÙYÁÏŠA\\œyÔ‡.mp4"
			fileName := "D:\\jack_huang\\Downloads\\newFolder\\test.mp4"
			var fileSize uint64 = 60727169
			var timestamp int64 = 1697049243123
			rtkPlatform.GoSetupFileDrop(ip, id, fileName, rtkPlatform.GetPlatform(), fileSize, timestamp)
		} else if strings.Contains(line, "UPDATE_PROGRESS") {
			ip := "192.168.30.1:12345"
			id := "QmQ7obXFx1XMFr6hCYXtovn9zREFqSXEtH5hdtpBDLjrAz"

			// fileName := "D:\\jack_huang\\Downloads\\newFolder\\test.mp4test.mp4"
			fileName := "D:\\jack_huang\\Downloads\\newFolder\\test.mp4"
			var fileSize uint64 = 60727169
			var timestamp int64 = int64(time.Now().UnixMilli())
			var sentSize uint64 = 0
			for scanner.Scan() {
				line2 := scanner.Text()
				fmt.Println("SliceSize:", line2)

				sliceSize, err := strconv.ParseUint(line2, 10, 64)
				if err == nil {
					if sliceSize == 0 {
						break
					}

					if (sentSize + sliceSize) >= fileSize {
						sliceSize = fileSize - sentSize
						sentSize = fileSize
					} else {
						sentSize += sliceSize
					}
					rtkPlatform.GoUpdateProgressBar(ip, id, fileSize, sentSize, timestamp, fileName)
				}
			}
		}
	}
}
