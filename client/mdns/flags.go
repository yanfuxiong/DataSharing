package mdns

import (
	"flag"
)

type Config struct {
	RendezvousString string
	ProtocolID       string
	ListenHost       string
	ListenPort       int
	LogSwitch        int
}

var MdnsCfg *Config

func ParseFlags() *Config {
	c := &Config{}

	flag.StringVar(&c.RendezvousString, "rendezvous", "meetme", "Unique string to identify group of nodes. Share this with your friends to let them connect with you")
	flag.StringVar(&c.ListenHost, "host", "0.0.0.0", "The bootstrap node host listen address\n")
	flag.StringVar(&c.ProtocolID, "pid", "/chat/1.1.0", "Sets a protocol id for stream headers")
	flag.IntVar(&c.ListenPort, "port", 0, "node listen port (0 pick a random unused port)")
	flag.IntVar(&c.LogSwitch, "log", 0, "log file switch (0: write log to p2p.log; 1: write log to console)")

	flag.Parse()
	return c
}
