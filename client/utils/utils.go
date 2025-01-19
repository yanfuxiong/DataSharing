package utils

import (
	"bufio"
	"crypto/md5"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"os"
	rtkCommon "rtk-cross-share/common"
	rtkGlobal "rtk-cross-share/global"
	"runtime"
	"runtime/debug"
	"strconv"
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
				log.SetOutput(&lumberjack.Logger{
					Filename:   "crash.log",
					MaxSize:    256,
					MaxBackups: 3,
					MaxAge:     30,
					Compress:   true,
				})

				log.Printf("Recovered from panic: %v\n", r)
				log.Printf("Stack trace:\n%s", debug.Stack())

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

func GetFuncInfo() string {
	funcName := "UnknownFunc:"
	pc, _, line, ok := runtime.Caller(1)
	if !ok {
		return funcName + strconv.Itoa(line)
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return funcName + strconv.Itoa(line)
	}
	fullName := fn.Name()
	if fullName == "" {
		return funcName + strconv.Itoa(line)
	}
	parts := strings.Split(fullName, ".")
	if len(parts) == 0 {
		return funcName + strconv.Itoa(line)
	}
	funcName = parts[len(parts)-1] + ":"
	return funcName + strconv.Itoa(line)
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

func GetClientInfo(id string) (rtkCommon.ClientInfo, error) {
	rtkGlobal.ClientListRWMutex.RLock()
	defer rtkGlobal.ClientListRWMutex.RUnlock()

	if val, ok := rtkGlobal.ClientInfoMap[id]; ok {
		return val, nil
	}
	return rtkCommon.ClientInfo{}, errors.New(fmt.Sprintf("not found ClientInfo by id:%s", id))
}

func GetClientIp(id string) (string, bool) {
	rtkGlobal.ClientListRWMutex.RLock()
	defer rtkGlobal.ClientListRWMutex.RUnlock()

	if val, ok := rtkGlobal.ClientInfoMap[id]; ok {
		return val.IpAddr, ok
	}
	log.Printf("not found ClientInfo by id:%s", id)
	return "", false
}

func InsertClientInfoMap(id, ipAddr, platform, name string) {
	rtkGlobal.ClientListRWMutex.Lock()
	defer rtkGlobal.ClientListRWMutex.Unlock()

	if _, ok := rtkGlobal.ClientInfoMap[id]; !ok {
		rtkGlobal.ClientInfoMap[id] = rtkCommon.ClientInfo{ID: id, IpAddr: ipAddr, Platform: platform, DeviceName: name}
	}

}

func LostClientInfoMap(id string) {
	rtkGlobal.ClientListRWMutex.Lock()
	defer rtkGlobal.ClientListRWMutex.Unlock()

	delete(rtkGlobal.ClientInfoMap, id)
}

func RemoveMdnsClientFromGuest() {
	rtkGlobal.ClientListRWMutex.RLock()
	defer rtkGlobal.ClientListRWMutex.RUnlock()

	for _, val := range rtkGlobal.ClientInfoMap {
		rtkGlobal.GuestList = RemoveMySelfID(rtkGlobal.GuestList, val.ID)
	}
}

func GetClientList() string {
	rtkGlobal.ClientListRWMutex.RLock()
	defer rtkGlobal.ClientListRWMutex.RUnlock()

	var clientList string
	for _, val := range rtkGlobal.ClientInfoMap {
		clientList += val.IpAddr + "#"
		clientList += val.ID + "#"
		clientList += val.DeviceName + ","
	}
	return strings.Trim(clientList, ",")
}

func GetClientCount() int {
	rtkGlobal.ClientListRWMutex.RLock()
	defer rtkGlobal.ClientListRWMutex.RUnlock()

	return len(rtkGlobal.ClientInfoMap)
}

func GetClientMap() map[string]rtkCommon.ClientInfo {
	rtkGlobal.ClientListRWMutex.RLock()
	defer rtkGlobal.ClientListRWMutex.RUnlock()

	return rtkGlobal.ClientInfoMap
}

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

// FIXME: hack code
type DeviceInfo struct {
	IP   string
	Name string
}

var (
	HackDeviceNameMap map[string]string     = make(map[string]string)
	deviceInfoMap     map[string]DeviceInfo = make(map[string]DeviceInfo)
	DeviceStaticPort  string                = ""
)

func InitDeviceInfo(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Printf("[%s %d] Err: Not found device table: %s", GetFuncName(), GetLine(), filename)
		return
	}

	scanner := bufio.NewScanner(file)
	idx := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Get static port
		if idx == 0 {
			port := strings.SplitN(line, ":", 2)
			if len(port) < 2 {
				log.Printf("[%s %d] Err: Invalid param count(port), please check .DeviceInfo file", GetFuncName(), GetLine())
				return
			}

			DeviceStaticPort = string(port[1])

			if DeviceStaticPort == "" {
				log.Printf("[%s %d] Err: Empty staticPort, please check .DeviceInfo file", GetFuncName(), GetLine())
				return
			}
		} else { // Get device info
			parts := strings.SplitN(line, ":", 3)
			if len(parts) < 3 {
				log.Printf("[%s %d] Err: Invalid param count, please check .DeviceInfo file", GetFuncName(), GetLine())
				return
			}

			ip := string(parts[0])
			id := string(parts[1])
			name := string(parts[2])

			deviceInfoMap[id] = DeviceInfo{
				IP:   ip,
				Name: name,
			}
		}

		idx++
	}
}

func GetDeviceInfoMap() map[string]DeviceInfo {
	return deviceInfoMap
}

func GetDeviceInfo(id string) (DeviceInfo, error) {
	if deviceInfo, ok := deviceInfoMap[id]; ok {
		return deviceInfo, nil
	} else {
		return DeviceInfo{}, errors.New("not found deviceInfo")
	}
}

func GetDeviceIp(id string) (string, error) {
	deviceInfo, err := GetDeviceInfo(id)
	if err != nil {
		return "", err
	} else {
		ipAddr := deviceInfo.IP
		ipAddr += (":" + DeviceStaticPort)
		return ipAddr, nil
	}
}

// Deprecated with InitDeviceInfo
func InitDeviceTable(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Printf("[%s %d] Err: Not found device table: %s", GetFuncName(), GetLine(), filename)
		return
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// skip space and mark symbol
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			log.Printf("[%s %d] Err: Invalid param count", GetFuncName(), GetLine())
			return
		}

		id := string(parts[0])
		name := string(parts[1])

		HackDeviceNameMap[id] = name
	}
}

// Deprecated: replace with GetDeviceInfo
func QueryDeviceName(id string) string {
	if name, ok := HackDeviceNameMap[id]; ok {
		return name
	} else {
		return ""
	}
}
