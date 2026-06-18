package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	rtkBuildConfig "rtk-cross-share/lanServer/buildConfig"
	rtkClientManager "rtk-cross-share/lanServer/clientManager"
	rtkCommon "rtk-cross-share/lanServer/common"
	rtkdbManager "rtk-cross-share/lanServer/dbManager"
	rtkDebug "rtk-cross-share/lanServer/debug"
	rtkGlobal "rtk-cross-share/lanServer/global"
	rtkIfaceMgr "rtk-cross-share/lanServer/interfaceMgr"
	"strconv"
	"syscall"

	rtkNetwork "rtk-cross-share/lanServer/network"
	rtkMisc "rtk-cross-share/misc"
	"time"

	"github.com/grandcat/zeroconf"
)

var acceptErrFlag = make(chan struct{}, 1)

var (
	service          = flag.String("service", rtkMisc.LanServiceType, "Set the service type of the new service.")
	domain           = flag.String("domain", rtkMisc.LanServerDomain, "Set the network domain. Default should be fine.")
	port             = flag.Int("port", rtkMisc.LanServerPort, "Set the port the service is listening to.")
	serviceForServer = flag.String("serviceForServer", rtkMisc.LanServiceTypeForServer, "Set the service type of the new service.")

	g_foundOtherServer bool     = false
	lockFd             *os.File = nil
	lockFilePath       string
)

func init() {
	flag.Parse()

	logFile := fmt.Sprintf("%s%s.log", rtkGlobal.LOG_PATH, rtkBuildConfig.ServerName)
	crashLogFile := fmt.Sprintf("%s%sCrash.log", rtkGlobal.LOG_PATH, rtkBuildConfig.ServerName)
	rtkMisc.InitLog(logFile, crashLogFile, 32)
	rtkMisc.SetupLogConsoleFile()

	rtkMisc.CreateDir(rtkGlobal.SOCKET_PATH_ROOT, os.ModePerm)

	lockFilePath = filepath.Join(rtkGlobal.LOG_PATH, "singleton.lock")
	rtkGlobal.Scenario = rtkMisc.ScenarioType_ViewManager
}

func checkLanServerExists() (string, bool) {
	startTime := time.Now().UnixMilli()
	resolver, err := zeroconf.NewResolver(nil, nil)
	if err != nil {
		log.Println("Failed to initialize resolver:", err.Error())
		return "", false
	}

	getLanServerEntry := make(chan *zeroconf.ServiceEntry)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(5))
	defer cancel()
	err = resolver.Lookup(ctx, rtkMisc.LanServerName, *service, *domain, getLanServerEntry, false)
	if err != nil {
		log.Println("Failed to Lookup:", err.Error())
		return "", false
	}

	select {
	case <-ctx.Done():
		return "", false
	case entry := <-getLanServerEntry:
		log.Printf("Found a Service is running, use [%d] ms", time.Now().UnixMilli()-startTime)
		ipAddr := ""
		if len(entry.AddrIPv4[0]) > 0 {
			ipAddr = fmt.Sprintf("%s:%d", entry.AddrIPv4[0].String(), entry.Port)
		}
		return ipAddr, true
	}
}

func browseOtherServer(ctx context.Context) {
	startBrowse := func(subCtx context.Context) {
		g_foundOtherServer = false
		resolver, err := zeroconf.NewResolver(nil, nil)
		if err != nil {
			log.Println("Failed to initialize resolver:", err.Error())
			return
		}

		entries := make(chan *zeroconf.ServiceEntry)
		rtkMisc.GoSafe(func() {
			for entry := range entries {
				if (len(entry.AddrIPv4) == 0) || (entry.Instance == rtkGlobal.ServerMdnsId) {
					continue
				}

				ip := fmt.Sprintf("%s:%d", entry.AddrIPv4[0].String(), entry.Port)
				log.Printf("[%s] Found other server: %s, IP:%s", rtkMisc.GetFuncInfo(), entry.Instance, ip)
				g_foundOtherServer = true
			}
		})

		err = resolver.Browse(subCtx, *serviceForServer, *domain, entries)
		if err != nil {
			log.Printf("[%s] Failed to browse:%+v", rtkMisc.GetFuncInfo(), err.Error())
		}
		// log.Printf("Start Browse other server...")
	}

	ticker := time.NewTicker(time.Duration(5 * time.Minute))
	defer ticker.Stop()
	var subCancel context.CancelFunc

	{
		// first time browse
		var subCtx context.Context
		subCtx, subCancel = context.WithCancel(ctx)
		startBrowse(subCtx)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if subCancel != nil {
				subCancel()
				time.Sleep(1 * time.Second)
			}
			var subCtx context.Context
			subCtx, subCancel = context.WithCancel(ctx)
			startBrowse(subCtx)
		}
	}
}

func MainInit(shutdownCtx context.Context) {
	log.Println("=====================================================")
	log.Printf("%s LanServer Version: %s ", rtkBuildConfig.ServerName, rtkGlobal.LanServerVersion)
	log.Printf("%s Build Date: %s", rtkBuildConfig.ServerName, rtkBuildConfig.BuildDate)
	log.Printf("=====================================================\n\n")

	err := lockFile()
	if err != nil {
		log.Printf("Another instance is already running, return!\n\n")
		return
	}
	defer unLockFile()

	runCtx, runCancel := context.WithCancel(shutdownCtx)
	defer runCancel()
	rtkdbManager.InitSqlite(runCtx)
	initDpSrcType()

	var printErrNetwork = true
	for {
		if rtkMisc.IsNetworkConnected() {
			printErrNetwork = true
			break
		}

		if printErrNetwork {
			log.Printf("******** the network is unavailable! %s is not start! ******** ", rtkBuildConfig.ServerName)
			printErrNetwork = false
		}
		select {
		case <-runCtx.Done():
			return
		case <-time.After(5 * time.Second):
		}
	}

	rtkGlobal.ServerIPAddr = ""

	rtkIfaceMgr.GetInterfaceMgr().TriggerGetCapability()
	rtkMisc.GoSafe(func() { rtkDebug.DebugCmdLine() })
	rtkMisc.GoSafe(func() { Run(runCtx) })

	for {
		select {
		case <-runCtx.Done():
			return
		case <-acceptErrFlag:
			rtkdbManager.UpdateAllClientOffline()
			select {
			case <-runCtx.Done():
				return
			case <-time.After(5 * time.Second):
			}
		}

		log.Printf("registerMdns Server restart...\n\n")
		rtkMisc.GoSafe(func() { Run(runCtx) })
	}
}

func initDpSrcType() {
	usbcCnt := 0
	dpCnt := 0
	for port := range rtkCommon.MAX_PORT_DP {
		dpSrcType := rtkIfaceMgr.GetInterfaceMgr().TriggerGetDpSrcTypeCb(rtkGlobal.Src_DP, port)
		if dpSrcType == rtkGlobal.DP_SRC_TYPE_USBC {
			if usbcCnt == 0 {
				rtkCommon.DpSrcTypeAry[port] = rtkCommon.SrcPortType_USBC_1
			} else if usbcCnt == 1 {
				rtkCommon.DpSrcTypeAry[port] = rtkCommon.SrcPortType_USBC_2
			}
			usbcCnt++
		} else {
			if dpCnt == 0 {
				rtkCommon.DpSrcTypeAry[port] = rtkCommon.SrcPortType_DP_1
			} else if dpCnt == 1 {
				rtkCommon.DpSrcTypeAry[port] = rtkCommon.SrcPortType_DP_2
			}
			dpCnt++
		}
	}
}

func getValidHarwdareAddr(iface *net.Interface) (string, error) {
	if iface == nil {
		return "", fmt.Errorf("[%s] Err: null interface", rtkMisc.GetFuncInfo())
	}

	var hardwareAddr = ""
	for _, b := range iface.HardwareAddr {
		hardwareAddr += fmt.Sprintf("%02X", b)
	}

	if hardwareAddr == "" {
		return "", fmt.Errorf("[%s] Err: empty MAC address", rtkMisc.GetFuncInfo())
	}

	return hardwareAddr, nil
}

func registerMdns(server **zeroconf.Server, serverForSearch **zeroconf.Server) []net.Addr {
	var printErrIface = true
	var printErrMac = true
	var printErrIp = true
	var printErrMdns = true
	index := 0
	for {
		time.Sleep(100 * time.Millisecond)
		interfaceList, err := rtkNetwork.GetValidInterfaceList()
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		if index >= len(interfaceList) {
			index = 0
		}

		iface, err := rtkMisc.GetValidInterface(interfaceList[index])
		if err != nil {
			if printErrIface {
				log.Printf("Err: Get network interface(%s) failed: %s", interfaceList[index], err.Error())
				printErrIface = false
			}
			index++
			continue
		}

		// Use the perferred interface MAC address as mDNS ID, even the interface has no IP
		if rtkGlobal.ServerMdnsId == "" {
			mdnsId, err := getValidHarwdareAddr(iface)
			if err != nil {
				if printErrMac {
					log.Printf("Err: Get server MAC address failed: %s", err.Error())
					printErrMac = false
				}
				index++
				continue
			}

			rtkGlobal.ServerMdnsId = mdnsId
		}
		printErrMac = true

		addrs, err := rtkMisc.GetValidAddrs(iface)
		if err != nil {
			if printErrIp {
				log.Printf("Err: Get server IP address failed: %s", err.Error())
				printErrIp = false
			}
			index++
			continue
		}
		printErrIp = true

		var ipAddr = ""
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil {
					ipAddr = ipNet.IP.String()
					break
				}
			}
		}
		// It's necessary be a contentText.
		// If the contentText is null or empty, that iOS cannot discover service
		// iOS use the IP in textRecord to skip the different IP from bonjour service
		getTextRecord := func(key, value string) string {
			return key + "=" + value
		}
		textRecordIp := getTextRecord(rtkMisc.TextRecordKeyIp, ipAddr)
		textRecordMonitorName := getTextRecord(rtkMisc.TextRecordKeyMonitorName, rtkGlobal.ServerMonitorName)
		textRecordProductName := getTextRecord(rtkMisc.TextRecordKeyProductName, rtkGlobal.ServerProductName)
		textRecordTimestamp := getTextRecord(rtkMisc.TextRecordKeyTimestamp, strconv.FormatInt(time.Now().UnixMilli(), 10))
		textRecordVersion := getTextRecord(rtkMisc.TextRecordKeyVersion, rtkGlobal.LanServerVersion)
		textRecords := []string{textRecordIp, textRecordMonitorName, textRecordProductName, textRecordTimestamp, textRecordVersion}
		*server, err = zeroconf.Register(rtkGlobal.ServerMdnsId, *service, *domain, *port, textRecords, []net.Interface{*iface})
		*serverForSearch, _ = zeroconf.Register(rtkGlobal.ServerMdnsId, *serviceForServer, *domain, *port, []string{}, []net.Interface{*iface})
		(*server).TTL(60)
		(*serverForSearch).TTL(60)

		if err != nil {
			if printErrMdns {
				log.Printf("Err: mDNS register failed: %s", err.Error())
				printErrMdns = false
			}
			index++
			continue
		}
		printErrMdns = true
		printErrIface = true
		return addrs
	}
}

func Run(runCtx context.Context) {
	/*ipAddress, bExisted := checkLanServerExists()
	if bExisted {
		log.Printf("an other %s IPAddr:[%s] is already running! so exit!", rtkBuildConfig.ServerName, ipAddress)
		log.Fatal(fmt.Sprintf("%s is already exist!", rtkBuildConfig.ServerName))
	}*/
	var server *zeroconf.Server = nil
	var serverForSearch *zeroconf.Server = nil
	defer func() {
		if server != nil {
			server.Shutdown()
			log.Printf("[%s] Rebuilt mDNS server, shutdown the last one", rtkMisc.GetFuncInfo())
		}
		if serverForSearch != nil {
			serverForSearch.Shutdown()
		}
	}()
	getServerListening := func() (net.Listener, error) {
		startTime := time.Now().UnixMilli()

		if server != nil {
			server.Shutdown()
			log.Printf("[%s] Rebuilt mDNS server, shutdown the last one", rtkMisc.GetFuncInfo())
		}
		if serverForSearch != nil {
			serverForSearch.Shutdown()
		}
		addrs := registerMdns(&server, &serverForSearch)

		var lastTrigger int64
		server.ListenQuery(func() {
			if g_foundOtherServer {
				now := time.Now().UnixMilli()
				if now-lastTrigger > 1000 { // debounce in 1 sec
					lastTrigger = now
					// DEBUG
					// log.Printf("[%s] Found other server and detect client query", rtkMisc.GetFuncInfo())
					rtkIfaceMgr.GetInterfaceMgr().TriggerDisplayMonitorName()
				}
			}
		})
		log.Printf("Register use [%d]ms, Published service info:", time.Now().UnixMilli()-startTime)
		log.Println("- Name:", rtkGlobal.ServerMdnsId)
		log.Println("- Type:", *service)
		log.Println("- Domain:", *domain)
		log.Println("- Port:", *port)

		serverAddr := ""
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil {
					serverAddr = ipNet.IP.String()
				}
			}
		}
		if serverAddr == "" {
			log.Printf("get serverAddr is  null !\n")
			return nil, errors.New("get serverAddr is null!")
		}

		rtkGlobal.ServerIPAddr = serverAddr
		rtkGlobal.ServerPort = *port

		serverAddr = fmt.Sprintf("%s:%d", serverAddr, *port)
		listener, err := net.Listen("tcp", serverAddr)
		if err != nil {
			log.Printf("Error listening:%+v", err)
		}

		return listener, err
	}

	var listener net.Listener
	var err error
	for {
		listener, err = getServerListening()
		if err == nil {
			break
		}
		select {
		case <-runCtx.Done():
			return
		case <-time.After(5 * time.Second):
		}
	}

	log.Printf("%s is listening on %s:%d success ! \n", rtkBuildConfig.ServerName, rtkGlobal.ServerIPAddr, rtkGlobal.ServerPort)
	sessCtx, sessCancel := context.WithCancel(runCtx)
	defer sessCancel()
	rtkMisc.GoSafe(func() {
		<-sessCtx.Done()
		listener.Close()
	})

	rtkMisc.GoSafe(func() { browseOtherServer(sessCtx) })
	rtkMisc.GoSafe(func() { rtkClientManager.PeriodicNotifyHandler(sessCtx) })
	rtkMisc.GoSafe(func() { rtkClientManager.HandleClientSignalChecking(sessCtx) })

	defer listener.Close()
	for {
		conn, accErr := listener.Accept()
		if accErr != nil {
			log.Printf("Server %s:%d accepting connection Error:%+v", rtkGlobal.ServerIPAddr, rtkGlobal.ServerPort, accErr.Error())
			if opErr, ok := accErr.(*net.OpError); ok {
				log.Printf("[%s] Accept OpError: %v, Op: %s, Net: %s, Err: %v", rtkMisc.GetFuncInfo(), opErr, opErr.Op, opErr.Net, opErr.Err)
				if opErr.Temporary() {
					log.Println("Temporary error, continuing...")
					continue
				}
			}
			sessCancel()
			var errno syscall.Errno
			if errors.As(accErr, &errno) {
				log.Printf("Accept errno: %v", errno)
				if errno == syscall.EINVAL {
					log.Printf("EINVAL")
				} else if errno == syscall.ECONNRESET {
					log.Printf("ECONNRESET")
				}
			}

			acceptErrFlag <- struct{}{}
			log.Printf("%s  %s:%d listening is cancel!", rtkBuildConfig.ServerName, rtkGlobal.ServerIPAddr, rtkGlobal.ServerPort)
			break
		}
		timestamp := time.Now().UnixMicro()
		log.Printf("%s Accept a connect, RemoteAddr: %s \n", rtkBuildConfig.ServerName, conn.RemoteAddr().String())

		tcpConn, _ := conn.(*net.TCPConn)
		if tcpConn != nil {
			tcpConn.SetKeepAlive(true)
			tcpConn.SetKeepAlivePeriod(15 * time.Second)
		}

		rtkMisc.GoSafe(func() { rtkClientManager.HandleClient(sessCtx, conn, timestamp) })
	}
	log.Printf("Cancel server. Close TCP and mDNS server in %s:%d \n\n", rtkGlobal.ServerIPAddr, rtkGlobal.ServerPort)
}

func lockFile() error {
	var err error
	lockFd, err = os.OpenFile(lockFilePath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Printf("Failed to open or create lock file:[%s] err:%+v", lockFilePath, err)
		return err
	}

	err = syscall.Flock(int(lockFd.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		log.Printf("Failed to lock file[%s] err:%+v", lockFilePath, err)
	}

	return err
}

func unLockFile() error {
	err := syscall.Flock(int(lockFd.Fd()), syscall.LOCK_UN|syscall.LOCK_NB)
	if err != nil {
		log.Printf("Failed to unlock file[%s] err:%+v", lockFilePath, err)
	}
	lockFd.Close()
	return err
}
