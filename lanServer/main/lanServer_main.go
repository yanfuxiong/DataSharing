package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	rtkBuildConfig "rtk-cross-share/lanServer/buildConfig"
	rtkClientManager "rtk-cross-share/lanServer/clientManager"
	rtkdbManager "rtk-cross-share/lanServer/dbManager"
	rtkDebug "rtk-cross-share/lanServer/debug"
	rtkGlobal "rtk-cross-share/lanServer/global"

	rtkNetwork "rtk-cross-share/lanServer/network"
	rtkMisc "rtk-cross-share/misc"
	"time"

	"github.com/grandcat/zeroconf"
)

var cancelServer = make(chan struct{})
var acceptErrFlag = make(chan struct{})

var (
	supInterfaces = []string{"wlan0", "eth0"} // e.g., "en0", "wlan0", "lo0", "eth0.100"

	name          = flag.String("name", rtkMisc.LanServerName, "The name for the service.")
	service       = flag.String("service", rtkMisc.LanServiceType, "Set the service type of the new service.")
	domain        = flag.String("domain", rtkMisc.LanServerDomain, "Set the network domain. Default should be fine.")
	port          = flag.Int("port", rtkMisc.LanServerPort, "Set the port the service is listening to.")
	interfaceName = flag.String("interface", supInterfaces[0], "Set the network interface name. Default WLAN.")
)

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
	err = resolver.Lookup(ctx, rtkMisc.LanServerName, *service, *domain, getLanServerEntry)
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

func MainInit() {
	flag.Parse()

	logFile := fmt.Sprintf("%s%s.log", rtkGlobal.LOG_PATH, rtkBuildConfig.ServerName)
	crashLogFile := fmt.Sprintf("%s%sCrash.log", rtkGlobal.LOG_PATH, rtkBuildConfig.ServerName)
	rtkMisc.InitLog(logFile, crashLogFile, 32)
	rtkMisc.SetupLogConsoleFile()

	rtkMisc.CreateDir(rtkGlobal.SOCKET_PATH_ROOT, os.ModePerm)

	log.Println("=====================================================")
	log.Printf("%s LanServer Version: %s , Client Base Version: %s", rtkBuildConfig.ServerName, rtkGlobal.LanServerVersion, rtkGlobal.ClientBaseVersion)
	log.Printf("%s Build Date: %s", rtkBuildConfig.ServerName, rtkBuildConfig.BuildDate)
	log.Printf("=====================================================\n\n")

	for {
		if rtkMisc.IsNetworkConnected() {
			break
		}

		log.Printf("******** the network is unavailable! %s is not start! ******** ", rtkBuildConfig.ServerName)
		time.Sleep(5 * time.Second)
	}

	rtkGlobal.ServerIPAddr = ""

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	rtkdbManager.InitSqlite(ctx)
	rtkMisc.GoSafe(func() { rtkDebug.DebugCmdLine() })
	rtkMisc.GoSafe(func() { rtkNetwork.WatchNetworkInfo(ctx) })
	rtkMisc.GoSafe(func() { Run() })

	for {
		select {
		case <-ctx.Done():
			return
		case <-rtkNetwork.GetNetworkSwitchFlag():
			rtkdbManager.UpdateAllClientOffline()
			cancelServer <- struct{}{}
			time.Sleep(5 * time.Second)
			log.Printf("==============================================================================")
		case <-acceptErrFlag:
			time.Sleep(5 * time.Second)
		}

		log.Printf("registerMdns Server restart...\n\n")
		rtkMisc.GoSafe(func() { Run() })
	}
}

func getValidAddrs(iface *net.Interface) ([]net.Addr, error) {
	if iface == nil {
		return nil, fmt.Errorf("[%s] Err: null interface", rtkMisc.GetFuncInfo())
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}

	var retAddrs []net.Addr
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				retAddrs = append(retAddrs, addr)
			}
		}
	}

	if len(retAddrs) == 0 {
		return nil, fmt.Errorf("Err: Empty addrs (%s)", iface.Name)
	}

	return retAddrs, nil
}

func getValidInterface(ifaceName string) (*net.Interface, error) {
	iface, err := net.InterfaceByName(ifaceName)
	if err == nil && iface != nil {
		return iface, nil
	}
	return nil, err
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

func setupMdnsId(iface *net.Interface) {
	if iface == nil {
		log.Printf("[%s] Err: null interface", rtkMisc.GetFuncInfo())
		return
	}

	rtkGlobal.ServerMdnsId = ""
	for _, b := range iface.HardwareAddr {
		rtkGlobal.ServerMdnsId += fmt.Sprintf("%02X", b)
	}
	*name = rtkGlobal.ServerMdnsId
}

func registerMdns(server *zeroconf.Server) []net.Addr {
	var mdnsId = ""
	var printErrIface = true
	var printErrMac = true
	var printErrIp = true
	var printErrMdns = true
	for {
		mdnsId = ""
		for _, ifaceName := range supInterfaces {
			iface, err := getValidInterface(ifaceName)
			if err != nil {
				if printErrIface {
					log.Printf("Err: Get network interface(%s) failed: %s", ifaceName, err.Error())
					printErrIface = false
				}
				continue
			}

			// Use the perferred interface MAC address as mDNS ID, even the interface has no IP
			if mdnsId == "" {
				mdnsId, err = getValidHarwdareAddr(iface)
				if err != nil {
					if printErrMac {
						log.Printf("Err: Get server MAC address failed: %s", err.Error())
						printErrMac = false
					}
					continue
				}

				*name = mdnsId
				rtkGlobal.ServerMdnsId = mdnsId
			}
			printErrMac = true

			addrs, err := getValidAddrs(iface)
			if err != nil {
				if printErrIp {
					log.Printf("Err: Get server IP address failed: %s", err.Error())
					printErrIp = false
				}
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
			rtkGlobal.ServerPort, _ = getAvailablePort()
			if rtkGlobal.ServerPort == 0 {
				rtkGlobal.ServerPort = *port
			}

			// It's necessary be a contentText.
			// If the contentText is null or empty, that iOS cannot discover service
			// iOS use the IP in textRecord to skip the different IP from bonjour service
			textRecordIp := rtkMisc.TextRecordKeyIp + "=" + ipAddr
			// TODO: ProductName from DIAS
			server, err = zeroconf.Register(*name, *service, *domain, rtkGlobal.ServerPort, []string{textRecordIp}, []net.Interface{*iface})

			if err != nil {
				if printErrMdns {
					log.Printf("Err: mDNS register failed: %s", err.Error())
					printErrMdns = false
				}
				continue
			}
			printErrMdns = true
			printErrIface = true
			return addrs
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func Run() {
	/*ipAddress, bExisted := checkLanServerExists()
	if bExisted {
		log.Printf("an other %s IPAddr:[%s] is already running! so exit!", rtkBuildConfig.ServerName, ipAddress)
		log.Fatal(fmt.Sprintf("%s is already exist!", rtkBuildConfig.ServerName))
	}*/

	getServerListening := func() (net.Listener, error) {
		startTime := time.Now().UnixMilli()
		var server *zeroconf.Server
		addrs := registerMdns(server)
		defer func() {
			if server != nil {
				server.Shutdown()
			}
		}()

		log.Printf("Register use [%d]ms, Published service info:", time.Now().UnixMilli()-startTime)
		log.Println("- Name:", *name)
		log.Println("- Type:", *service)
		log.Println("- Domain:", *domain)
		log.Println("- Port:", rtkGlobal.ServerPort)

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
		serverAddr = fmt.Sprintf("%s:%d", serverAddr, rtkGlobal.ServerPort)
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
		time.Sleep(5 * time.Second)
	}

	defer listener.Close()
	log.Printf("%s is listening on %s:%d success ! \n", rtkBuildConfig.ServerName, rtkGlobal.ServerIPAddr, rtkGlobal.ServerPort)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rtkMisc.GoSafe(func() { rtkClientManager.ReconnClientListHandler(ctx) })

	for {
		select {
		case <-ctx.Done():
			return
		case <-cancelServer:
			log.Printf("%s  %s:%d listening is cancel!", rtkBuildConfig.ServerName, rtkGlobal.ServerIPAddr, rtkGlobal.ServerPort)
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Server %s:%d accepting connection Error:%+v", rtkGlobal.ServerIPAddr, rtkGlobal.ServerPort, err.Error())
				acceptErrFlag <- struct{}{}
				log.Printf("%s  %s:%d listening is cancel!\n\n", rtkBuildConfig.ServerName, rtkGlobal.ServerIPAddr, rtkGlobal.ServerPort)
				return
			}
			timestamp := time.Now().UnixMicro()
			log.Printf("%s Accept a connect, RemoteAddr: %s \n", rtkBuildConfig.ServerName, conn.RemoteAddr().String())
			rtkMisc.GoSafe(func() { rtkClientManager.HandleClient(ctx, conn, timestamp) })
		}
	}
}

func getAvailablePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		log.Printf("[%s] error:%+v ", rtkMisc.GetFuncInfo(), err)
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Printf("[%s]  error:%+v", rtkMisc.GetFuncInfo(), err)
		return 0, err
	}
	defer l.Close()

	nPort := l.Addr().(*net.TCPAddr).Port
	log.Printf("[%s] get port:%d\n", rtkMisc.GetFuncInfo(), nPort)
	return nPort, nil
}
