package mdns

/*type Config struct {
	RendezvousString string
	ProtocolID       string
	ListenHost       string
	ListenPort       int
}

var MdnsCfg *Config

func ParseFlags() *Config {
	c := &Config{}

	flag.StringVar(&c.RendezvousString, "rendezvous", "meetme", "Unique string to identify group of nodes. Share this with your friends to let them connect with you")
	flag.StringVar(&c.ListenHost, "host", rtkGlobal.DefaultIp, "The bootstrap node host listen address\n")
	flag.StringVar(&c.ProtocolID, "pid", "/chat/1.1.0", "Sets a protocol id for stream headers")
	flag.IntVar(&c.ListenPort, "port", 0, "node listen port (0 pick a random unused port)")

	flag.Parse()
	return c
}

func ListenMultAddr() []ma.Multiaddr {
	addrs := []string{
		"/ip4/%s/tcp/%d",
		"/ip4/%s/udp/%d/quic",
	}

	for i, a := range addrs {
		addrs[i] = fmt.Sprintf(a, MdnsCfg.ListenHost, MdnsCfg.ListenPort)
	}

	multAddr := make([]ma.Multiaddr, 0)
	for _, addrstr := range addrs {
		a, err := ma.NewMultiaddr(addrstr)
		if err != nil {
			continue
		}
		multAddr = append(multAddr, a)
	}

	return multAddr
}
*/
