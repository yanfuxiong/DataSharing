package debug

import (
	"bufio"
	"fmt"
	"os"
	rtkGlobal "rtk-cross-share/global"
	rtkUtils "rtk-cross-share/utils"
	"strings"
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
		if strings.Contains(line, "GetClientList") {
			fmt.Println("ClientList:", rtkUtils.GetClientList())
		} else if strings.Contains(line, "WaitConnPeer") {
			fmt.Println("WaitConnPeerMap size:", len(rtkGlobal.WaitConnPeerMap))
			for k, _ := range rtkGlobal.WaitConnPeerMap {
				fmt.Printf("Id: %s", k)
			}
		}
	}
}
