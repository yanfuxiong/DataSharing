package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/grandcat/zeroconf"
	"log"
	"net"
	rtkBuildConfig "rtk-cross-share/lanServer/buildConfig"
	rtkClientManager "rtk-cross-share/lanServer/clientManager"
	rtkdbManager "rtk-cross-share/lanServer/dbManager"
	rtkDebug "rtk-cross-share/lanServer/debug"
	rtkMisc "rtk-cross-share/misc"

	"time"
)

const (
	InterfaceName = "WLAN" // e.g., "en0", "lo0", "eth0.100"
)

var (
	name          = flag.String("name", rtkMisc.LanServerName, "The name for the service.")
	service       = flag.String("service", rtkMisc.LanServiceType, "Set the service type of the new service.")
	domain        = flag.String("domain", rtkMisc.LanServerDomain, "Set the network domain. Default should be fine.")
	port          = flag.Int("port", rtkMisc.LanServerPort, "Set the port the service is listening to.")
	interfaceName = flag.String("interface", InterfaceName, "Set the network interface name. Default WLAN.")
)

func checkLanServerExists() bool {
	startTime := time.Now().UnixMilli()
	resolver, err := zeroconf.NewResolver(nil, nil)
	if err != nil {
		log.Fatalln("Failed to initialize resolver:", err.Error())
	}

	getLanServerChan := make(chan struct{})
	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			log.Printf("Found a Service :%s ,use [%d] ms", entry.Instance, time.Now().UnixMilli()-startTime)
			if entry.Instance == rtkMisc.LanServerName {
				getLanServerChan <- struct{}{}
			}
		}
	}(entries)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(5))
	defer cancel()
	startTime = time.Now().UnixMilli()
	err = resolver.Browse(ctx, *service, *domain, entries)
	if err != nil {
		log.Fatalln("Failed to browse:", err.Error())
	}

	select {
	case <-ctx.Done():
		return false
	case <-getLanServerChan:
		return true
	}
}

func main() {
	flag.Parse()

	logPath := fmt.Sprintf("%s.log", rtkBuildConfig.ServerName)
	crashLogPath := fmt.Sprintf("%sCrash.log", rtkBuildConfig.ServerName)
	rtkMisc.InitLog(logPath, crashLogPath)
	rtkMisc.SetupLogConsole()

	log.Println("==========================================")
	log.Printf("%s Version: %s", rtkBuildConfig.ServerName, rtkBuildConfig.Version)
	log.Printf("%s Build Date: %s", rtkBuildConfig.ServerName, rtkBuildConfig.BuildDate)
	log.Println("========================================\n\n")

	if checkLanServerExists() {
		log.Printf("an other %s is already running!", rtkBuildConfig.ServerName)
		return
	}

	iface, err := net.InterfaceByName(*interfaceName) // TODO: We use specific network interface now. Survey how to auto fit network interface
	if err != nil {
		log.Fatal(err)
	}
	addrs, err := iface.Addrs()
	if err != nil {
		log.Printf("Err: Failed to get addresses for interface %s: %v", iface.Name, err)
		log.Fatal(err)
	}

	startTime := time.Now().UnixMilli()
	//server, err := zeroconf.Register(*name, *service, *domain, *port, []string{"txtv=0", "lo=1", "la=2", "path=/"}, nil)
	server, err := zeroconf.Register(*name, *service, *domain, *port, nil, []net.Interface{*iface})
	if err != nil {
		panic(err)
	}
	defer server.Shutdown()

	log.Printf("Register use [%d]ms, Published service info:", time.Now().UnixMilli()-startTime)
	log.Println("- Name:", *name)
	log.Println("- Type:", *service)
	log.Println("- Domain:", *domain)
	log.Println("- Port:", *port)

	var serverAddr string
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				serverAddr = ipNet.IP.String()
			}
		}
	}
	serverAddr += fmt.Sprintf(":%d", *port)
	listener, err := net.Listen("tcp", serverAddr)
	if err != nil {
		log.Fatal("Error listening:", err.Error())
	}
	defer listener.Close()
	log.Printf("LanServer is listening on %s success! \n", serverAddr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rtkdbManager.InitSqlite(ctx)
	rtkMisc.GoSafe(func() { rtkDebug.DebugCmdLine() })

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err.Error())
			continue
		}

		log.Printf("LanServer Accept a connect, RemoteAddr: %s \n", conn.RemoteAddr().String())
		rtkMisc.GoSafe(func() { rtkClientManager.HandleClient(conn) })
	}
}
