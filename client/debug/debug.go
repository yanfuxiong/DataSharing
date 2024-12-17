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
	// test_case_1 := TestCase{
	// 	FileName:     "D:\\share\\PC1.txt",
	// 	FileSizeHigh: 0,
	// 	FileSizeLow:  13902,
	// }

	// test_case_2 := TestCase{
	// 	FileName:     "D:\\share\\library.zip",
	// 	FileSizeHigh: 0,
	// 	FileSizeLow:  6452991,
	// }

	// test_case_3 := TestCase{
	// 	FileName:     "E:\\CODE\\png\\3.png",
	// 	FileSizeHigh: 0,
	// 	FileSizeLow:  1508,
	// }

	// test_case_4 := TestCase{
	// 	FileName:     "/Users/hp/myGolang/test.png",
	// 	FileSizeHigh: 0,
	// 	FileSizeLow:  109939,
	// }

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
			// fileName := "D:\\jack_huang\\Downloads\\新增資料夾\\測試.mp4"
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
		// } else if strings.Contains(line, "PASTE_FILE") {
		// 	rtkPlatform.GoClipboardPasteFileCallback("123")
		// } else if strings.Contains(line, "FILE_DROP_TEST_1") {
		// 	rtkGlobal.Handler.CopyFilePath.Store(test_case_4.FileName)
		// 	rtkGlobal.Handler.CopyDataSize.SizeHigh = test_case_4.FileSizeHigh
		// 	rtkGlobal.Handler.CopyDataSize.SizeLow = test_case_4.FileSizeLow
		// 	var fileInfo = rtkCommon.FileInfo{
		// 		FileSize_: rtkCommon.FileSize{
		// 			SizeHigh: test_case_4.FileSizeHigh,
		// 			SizeLow:  test_case_4.FileSizeLow,
		// 		},
		// 		FilePath: test_case_4.FileName,
		// 	}
		// 	rtkFileDrop.SendFileDropCmd(rtkCommon.FILE_DROP_REQUEST, fileInfo)
		// } else if strings.Contains(line, "FILE_DROP_TEST_2") {
		// 	rtkGlobal.Handler.CopyFilePath.Store(test_case_5.FileName)
		// 	rtkGlobal.Handler.CopyDataSize.SizeHigh = test_case_5.FileSizeHigh
		// 	rtkGlobal.Handler.CopyDataSize.SizeLow = test_case_5.FileSizeLow
		// 	var fileInfo = rtkCommon.FileInfo{
		// 		FileSize_: rtkCommon.FileSize{
		// 			SizeHigh: test_case_5.FileSizeHigh,
		// 			SizeLow:  test_case_5.FileSizeLow,
		// 		},
		// 		FilePath: test_case_5.FileName,
		// 	}
		// 	rtkFileDrop.SendFileDropCmd(rtkCommon.FILE_DROP_REQUEST, fileInfo)
		// }
	}
}
