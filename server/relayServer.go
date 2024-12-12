// relayserver.go
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"
	rtkUtils "rtk-cross-share/utils"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	ma "github.com/multiformats/go-multiaddr"
)

const ProtocolID = "host_register"

type RegMessage struct {
	HOST  string
	GUEST string
}

type RegResponseMessage struct {
	GUEST_LIST            []string
	GUEST_PUBLIC_TCP_IP   string
	GUEST_PUBLIC_TCP_PORT string
}

type Host struct {
	GuestList []string
}

type Handler struct {
	HostPool map[string]*Host
}

func main() {
	relayNode := setupNode()
	if relayNode == nil {
		panic("Fail to create node!")
	}
	defer relayNode.Close()
	setupRegisterFunc(relayNode)
	select {}
}

func listen_addrs(port int) []string {
	addrs := []string{
		"/ip4/0.0.0.0/tcp/%d",
		"/ip4/0.0.0.0/udp/%d/quic",
		"/ip6/::/tcp/%d",
		"/ip6/::/udp/%d/quic",
	}

	for i, a := range addrs {
		addrs[i] = fmt.Sprintf(a, port)
	}

	return addrs
}

func dump(handler *Handler) {
	for k, v := range handler.HostPool {
		log.Println("Host ID ", k)
		log.Println("=============================")
		for _, guest := range v.GuestList {
			log.Println("Register Guest ID ", guest)
		}
		log.Println("=============================")
	}
}

func isInTheList(target string, list []string) bool {
	for _, item := range list {
		if strings.EqualFold(item, target) {
			return true
		}
	}
	return false
}

func (h *Handler) updateHostPool(regMsg RegMessage) {
	if h.HostPool[regMsg.HOST] == nil {
		h.HostPool[regMsg.HOST] = &Host{}
	}
	if !isInTheList(regMsg.GUEST, h.HostPool[regMsg.HOST].GuestList) {
		h.HostPool[regMsg.HOST].GuestList = append(h.HostPool[regMsg.HOST].GuestList, regMsg.GUEST)
	}
}

func extractTCPIPandPort(maddr ma.Multiaddr) (string, string) {
	ip, err := maddr.ValueForProtocol(ma.P_IP4)
	if err != nil {
		log.Printf("Failed to get IP: %v", err)
	}

	port, err := maddr.ValueForProtocol(ma.P_TCP)
	if err != nil {
		log.Printf("Failed to get port: %v", err)
	}
	return ip, port
}

func handleStream(s network.Stream, handler *Handler, node host.Host) {
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	rtkUtils.GoSafe(func() {
		defer s.Close()
		log.Println("Wait for Node ...")
		for {
			var regMsg RegMessage
			err := json.NewDecoder(rw).Decode(&regMsg)
			if err != nil {
				if err == context.Canceled || err == context.DeadlineExceeded {
					log.Println("Stream context canceled or deadline exceeded:", err)
					return
				}
				if err.Error() == "stream reset" {
					log.Println("Stream reset by peer:", err)
					return
				}
				continue
			}

			handler.updateHostPool(regMsg)
			dump(handler)

			ip, port := extractTCPIPandPort(s.Conn().RemoteMultiaddr())
			var regResonseMsg RegResponseMessage = RegResponseMessage{GUEST_LIST: handler.HostPool[regMsg.HOST].GuestList, GUEST_PUBLIC_TCP_IP: ip, GUEST_PUBLIC_TCP_PORT: port}
			if err := json.NewEncoder(rw).Encode(&regResonseMsg); err != nil {
				log.Println("failed to read register response message: %w", err)
				return
			}
			if err := rw.Flush(); err != nil {
				log.Println("Error flushing write buffer: %w", err)
				return
			}

			time.Sleep(time.Second)
		}
	})
}

func marshalPrivateKeyToPEM(key crypto.PrivKey) ([]byte, error) {
	encoded, err := crypto.MarshalPrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %v", err)
	}
	pemEncoded := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: encoded,
	})
	return pemEncoded, nil
}

func unmarshalPrivateKeyFromPEM(pemData []byte) (crypto.PrivKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing private key")
	}
	return crypto.UnmarshalPrivateKey(block.Bytes)
}

func genKey() crypto.PrivKey {
	privKeyFile := ".priv.pem"
	var priv crypto.PrivKey
	var err error
	var content []byte
	content, err = ioutil.ReadFile(privKeyFile)
	if err != nil {
		priv, _, err = crypto.GenerateKeyPair(crypto.RSA, 2048)
		if err != nil {
			log.Fatal(err)
		}

		jsonData, err := marshalPrivateKeyToPEM(priv)
		err = ioutil.WriteFile(privKeyFile, jsonData, 0644)
		if err != nil {
			log.Fatal(err)
		}
		return priv
	}

	priv, err = unmarshalPrivateKeyFromPEM(content)

	return priv
}

func setupNode() host.Host {
	priv := genKey()
	relayNode, err := libp2p.New(
		libp2p.ListenAddrStrings(listen_addrs(7999)...),
		libp2p.Identity(priv),
		libp2p.EnableHolePunching(),
	)
	if err != nil {
		log.Printf("Failed to create relayNode: %v", err)
		return nil
	}

	relayOptions := relay.WithResources(relay.Resources{
		Limit: &relay.RelayLimit{
			Data:     1 << 20, // 1 MB
			Duration: time.Hour,
		},
		ReservationTTL:         time.Hour,
		MaxReservations:        128,
		MaxCircuits:            128,
		BufferSize:             65535,
		MaxReservationsPerPeer: 128,
		MaxReservationsPerIP:   128,
		MaxReservationsPerASN:  128,
	})

	for _, p := range relayNode.Peerstore().Peers() {
		relayNode.Peerstore().ClearAddrs(p)
	}

	_, err = relay.New(relayNode, relayOptions)
	if err != nil {
		log.Printf("Failed to instantiate the relay: %v", err)
		return nil
	}
	log.Printf("relayNodeInfo ID: %v Addrs: %v", relayNode.ID(), relayNode.Addrs())
	return relayNode
}

func setupRegisterFunc(relayNode host.Host) {
	handler := &Handler{HostPool: make(map[string]*Host)}
	relayNode.SetStreamHandler(ProtocolID, func(s network.Stream) {
		rtkUtils.GoSafe(func() {handleStream(s, handler, relayNode)})
	})
}
