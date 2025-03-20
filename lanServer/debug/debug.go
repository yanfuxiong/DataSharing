package debug

import (
	"bufio"
	"fmt"
	"os"
	rtkdbManager "rtk-cross-share/lanServer/dbManager"
	"strconv"
	"strings"
)

func DebugCmdLine() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter text to debug:")
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println("You entered:", line)
		if strings.Contains(line, "UpdateAuthStatus") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				index, _ := strconv.Atoi(parts[1])
				rtkdbManager.UpdateAuthStatus(index, true)
			}
		} else if strings.Contains(line, "UpdateAuthStatus") {
			//rtkdbManager.QueryOnlineClientList()
		}
	}
}
