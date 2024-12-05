package utils

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	rtkCommon "rtk-cross-share/common"
	rtkGlobal "rtk-cross-share/global"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/libp2p/go-libp2p/core/network"

	"github.com/libp2p/go-libp2p/core/crypto"
	ma "github.com/multiformats/go-multiaddr"
	mh "github.com/multiformats/go-multihash"
	"gopkg.in/natefinch/lumberjack.v2"
)

func GoSafe(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered from panic: %v\n", r)
				log.Printf("Stack trace:\n%s", debug.Stack())
				log.SetOutput(&lumberjack.Logger{
					Filename:   "crash.log",
					MaxSize:    256,
					MaxBackups: 3,
					MaxAge:     30,
					Compress:   true,
				})

				os.Exit(1)
			}
		}()
		fn()
	}()
}

func GetFuncName() string {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return "UnknownFunction"
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "UnknownFunction"
	}
	fullName := fn.Name()
	if fullName == "" {
		return "UnknownFunction"
	}
	parts := strings.Split(fullName, ".")
	if len(parts) == 0 {
		return "UnknownFunction"
	}
	return parts[len(parts)-1]
}

func GetLine() int {
	_, _, line, ok := runtime.Caller(1)
	if !ok {
		return -1
	}
	return line
}

func ContentEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func ConcatIP(ip string, port string) string {
	publicIP := fmt.Sprintf("%s:%s", ip, port)
	return publicIP
}

func SplitIP(ip string) (string, string) {
	parts := strings.Split(ip, ":")
	return parts[0], parts[1]
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func WriteNodeID(ID string, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	_, err = file.Write([]byte(ID))
	if err != nil {
		log.Println(err)
	}
}

func WriteMdnsPort(port string, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	_, err = file.Write([]byte(port))
	if err != nil {
		log.Println(err)
	}
}

func ReadMdnsPort(filename string) string {
	var err error
	var content []byte
	content, err = os.ReadFile(filename)
	if err != nil {
		return string(content[:])
	}
	return ""
}

func WriteErrJson(name string, strContent []byte) {
	fileName := fmt.Sprintf("/storage/emulated/0/Android/data/com.rtk.myapplication/files/%s.log", name)
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	_, err = file.Write(strContent)
	if err != nil {
		log.Println(err)
	}
}

func CreateMD5Hash(data []byte) (mh.Multihash, error) {
	hash := md5.Sum(data)

	multihash, err := mh.Encode(hash[:], mh.MD5)
	if err != nil {
		return nil, err
	}

	return multihash, nil
}

func MarshalPrivateKeyToPEM(key crypto.PrivKey) ([]byte, error) {
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

func UnmarshalPrivateKeyFromPEM(pemData []byte) (crypto.PrivKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing private key")
	}
	return crypto.UnmarshalPrivateKey(block.Bytes)
}

func GenKey(privKeyFile string) crypto.PrivKey {
	var priv crypto.PrivKey
	var err error
	var content []byte
	content, err = os.ReadFile(privKeyFile)
	if err != nil {
		priv, _, err = crypto.GenerateKeyPair(crypto.RSA, 2048)
		if err != nil {
			log.Fatal(err)
		}

		jsonData, err := MarshalPrivateKeyToPEM(priv)
		err = os.WriteFile(privKeyFile, jsonData, 0644)
		if err != nil {
			log.Fatal(err)
		}
		return priv
	}

	priv, err = UnmarshalPrivateKeyFromPEM(content)

	return priv
}

func GetLocalPort(Addrs []ma.Multiaddr) string {
	var localPort string
	for _, maddr := range Addrs {
		protocols := maddr.Protocols()
		hasTCP := false
		hasIP4 := false
		for _, protocol := range protocols {
			if protocol.Code == ma.P_TCP {
				hasTCP = true
			}
			if protocol.Code == ma.P_IP4 {
				hasIP4 = true
			}
		}
		if hasTCP && hasIP4 {
			port, err := maddr.ValueForProtocol(ma.P_TCP)
			if err != nil {
				return ""
			}
			localPort = port
			break
		}
	}

	//log.Println("Local port: " + localPort)
	return localPort
}

func ExtractTCPIPandPort(maddr ma.Multiaddr) (string, string) {
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

func GetRemoteAddrFromStream(stream network.Stream) string {
	ip, port := ExtractTCPIPandPort(stream.Conn().RemoteMultiaddr())
	return ConcatIP(ip, port)
}

func RemoveMySelfID(slice []string, s string) []string {
	i := 0
	for _, v := range slice {
		if v != s {
			slice[i] = v
			i++
		}
	}
	return slice[:i]
}

func ClearCachePeerInfo(peerId string) {
	if val, ok := rtkGlobal.WaitConnPeerMap[peerId]; ok {
		close(val.ChExitTimer)
		delete(rtkGlobal.WaitConnPeerMap, peerId)
		log.Println("peer:", peerId, " is connected, so clear cache Info")
	}
}

func IsInMdnsClientList(peerID string) bool {
	rtkGlobal.MdnsListRWMutex.Lock()
	defer rtkGlobal.MdnsListRWMutex.Unlock()
	for _, val := range rtkGlobal.MdnsClientList {
		if strings.EqualFold(peerID, val.ID) {
			return true
		}
	}
	return false
}

func InsertMdnsClientList(Id, IpAddr, platform string) {
	rtkGlobal.MdnsListRWMutex.Lock()
	defer rtkGlobal.MdnsListRWMutex.Unlock()
	isExist := false
	for _, val := range rtkGlobal.MdnsClientList {
		if strings.EqualFold(Id, val.ID) && strings.EqualFold(IpAddr, val.IpAddr) {
			isExist = true
		}
	}
	if !isExist {
		rtkGlobal.MdnsClientList = append(rtkGlobal.MdnsClientList, rtkCommon.ClientInfo{ID: Id, IpAddr: IpAddr, Platform: platform})
	}
}

func LostMdnsClientList(peerID string) {
	rtkGlobal.MdnsListRWMutex.Lock()
	defer rtkGlobal.MdnsListRWMutex.Unlock()
	var temList []rtkCommon.ClientInfo
	for _, item := range rtkGlobal.MdnsClientList {
		if !strings.EqualFold(item.ID, peerID) {
			temList = append(temList, item)
		}
	}
	rtkGlobal.MdnsClientList = temList
}

func RemoveMdnsClientFromGuest() {
	rtkGlobal.MdnsListRWMutex.RLock()
	defer rtkGlobal.MdnsListRWMutex.RUnlock()

	for _, val := range rtkGlobal.MdnsClientList {
		rtkGlobal.GuestList = RemoveMySelfID(rtkGlobal.GuestList, val.ID)
	}
}

func GetClientList() string {
	rtkGlobal.MdnsListRWMutex.RLock()
	defer rtkGlobal.MdnsListRWMutex.RUnlock()

	var clientList string
	for _, val := range rtkGlobal.MdnsClientList {
		clientList += val.IpAddr + "-" + val.Platform + "#"
	}
	return strings.Trim(clientList, "#")
}

/*func GetNodeCBData(ipAddr string) (rtkCommon.ClipBoardData, bool) {
	val, ok := rtkGlobal.CBData.Load(ipAddr)
	if !ok {
		log.Printf("Key:[%s] is not found", ipAddr)
		return rtkCommon.ClipBoardData{}, ok
	}

	if cbData, ok := val.(rtkCommon.ClipBoardData); ok {
		return cbData, ok
	}

	log.Printf("Key:[%+v] is not rtkCommon.ClipBoardData", val)
	return rtkCommon.ClipBoardData{}, false
}*/

func Base64Decode(src string) []byte {
	bytes, err := base64.StdEncoding.DecodeString(src)
	if err != nil {
		log.Printf("Base64Decode error:[%+v] [%s]", err, src)
		return nil
	}

	return bytes
}

func Base64Encode(src []byte) string {
	return base64.StdEncoding.EncodeToString(src)
}
