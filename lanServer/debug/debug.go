package debug

import (
	"bufio"
	"fmt"
	"os"
	rtkdbManager "rtk-cross-share/lanServer/dbManager"
	rtkIfaceMgr "rtk-cross-share/lanServer/interfaceMgr"
	rtkMisc "rtk-cross-share/misc"
	"strconv"
)

type DebugCmd struct {
	cmdName string
	cmdFunc func()
}

var (
	scanner  = bufio.NewScanner(os.Stdin)
	cmdArray = []DebugCmd{
		{"UpdateDeviceName", func() { updateDeviceName(scanner) }},
		{"SendDragFileStart", func() { sendDragFileStart(scanner) }},
		{"QueryClientInfoBySrcPort(Database)", func() { queryClientInfoBySrcPort(scanner) }},
	}
)

func displayCmd() {
	for i, v := range cmdArray {
		fmt.Printf("Index: %d, %s\n", i, v.cmdName)
	}
}

func DebugCmdLine() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\nDEBUG: Enter index or Enter 'h' to show all of cmds")
		if !scanner.Scan() {
			break
		}

		cmd := scanner.Text()
		if cmd == "h" {
			displayCmd()
			continue
		} else if cmd == "" {
			continue
		}

		cmdIdx, err := strconv.Atoi(cmd)
		if err != nil {
			fmt.Println("Invalid cmd value. Err: ", err.Error())
			continue
		}

		if cmdIdx >= len(cmdArray) {
			fmt.Println("Unknown cmd index")
			continue
		}

		cmdArray[cmdIdx].cmdFunc()
	}
}

func readIntInput(prompt string, scanner *bufio.Scanner) (int, bool) {
	fmt.Printf(prompt)
	for scanner.Scan() {
		text := scanner.Text()
		val, err := strconv.Atoi(text)
		if err != nil {
			fmt.Println("Invalid int value. Err: ", err.Error())
			break
		}

		return val, true
	}
	return 0, false
}

func readTextInput(prompt string, scanner *bufio.Scanner) string {
	fmt.Printf(prompt)
	for scanner.Scan() {
		return scanner.Text()
	}
	return ""
}

func queryClientInfoBySrcPort(scanner *bufio.Scanner) {
	src, ret := readIntInput("Source(HDMI:8, DP/TypeC:13): ", scanner)
	if !ret {
		return
	}
	port, ret := readIntInput("Port(0~3): ", scanner)
	if !ret {
		return
	}
	clientInfoList, err := rtkdbManager.QueryClientInfoBySrcPort(src, port)
	if err != rtkMisc.SUCCESS {
		fmt.Println("Err: ", err)
	}

	for _, clientInfo := range clientInfoList {
		clientInfo.Dump()
	}
}

func updateDeviceName(scanner *bufio.Scanner) {
	src, ret := readIntInput("Source(HDMI:8, DP/TypeC:13): ", scanner)
	if !ret {
		return
	}
	port, ret := readIntInput("Port(0~3): ", scanner)
	if !ret {
		return
	}
	name := readTextInput("Name: ", scanner)
	rtkIfaceMgr.GetInterfaceMgr().TriggerUpdateDeviceName(src, port, name)
}

func sendDragFileStart(scanner *bufio.Scanner) {
	src, ret := readIntInput("Source(HDMI:8, DP/TypeC:13): ", scanner)
	if !ret {
		return
	}
	port, ret := readIntInput("Port(0~3): ", scanner)
	if !ret {
		return
	}
	horzSize, ret := readIntInput("Resolution horiziation size: ", scanner)
	if !ret {
		return
	}
	vertSize, ret := readIntInput("Resolution vertical size: ", scanner)
	if !ret {
		return
	}
	posX, ret := readIntInput("Position X: ", scanner)
	if !ret {
		return
	}
	posY, ret := readIntInput("Position Y: ", scanner)
	if !ret {
		return
	}
	rtkIfaceMgr.GetInterfaceMgr().TriggerDragFileStart(src, port, horzSize, vertSize, posX, posY)
}
