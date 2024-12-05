package utils

import (
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"net"
)

type addr struct{ id peer.ID }

var Network = "libp2p"

// Network returns the name of the network that this address belongs to
// (libp2p).
func (a *addr) Network() string { return Network }

// String returns the peer ID of this address in string form
// (B58-encoded).
func (a *addr) String() string { return a.id.String() }

// conn is an implementation of net.Conn which wraps
// libp2p streams.
type conn struct {
	network.Stream
}

// newConn creates a conn given a libp2p stream
func NewConnFromStream(s network.Stream) net.Conn {
	return &conn{s}
}
func (c *conn) LocalAddr() net.Addr {
	return &addr{c.Stream.Conn().LocalPeer()}
}

// RemoteAddr returns the remote network address.
func (c *conn) RemoteAddr() net.Addr {
	return &addr{c.Stream.Conn().RemotePeer()}
}
