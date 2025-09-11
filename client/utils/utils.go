package utils

import (
	"bufio"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	rtkCommon "rtk-cross-share/client/common"
	rtkGlobal "rtk-cross-share/client/global"
	rtkMisc "rtk-cross-share/misc"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/network"

	"github.com/libp2p/go-libp2p/core/crypto"
	ma "github.com/multiformats/go-multiaddr"
	mh "github.com/multiformats/go-multihash"
)

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

func SplitIPAddr(ipAddr string) (string, string) {
	parts := strings.Split(ipAddr, ":")
	if len(parts) >= 2 {
		return parts[0], parts[1]
	}
	return parts[0], ""
}

func WriteNodeID(ID string, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("[%s] err:%+v", rtkMisc.GetFuncInfo(), err)
		panic(err)
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
	fileName := fmt.Sprintf("/storage/emulated/0/Android/data/com.realtek.crossshare/files/%s.log", name)
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
			log.Printf("[%s] err:%+v", rtkMisc.GetFuncInfo(), err)
			panic(err)
		}

		jsonData, err := MarshalPrivateKeyToPEM(priv)
		err = os.WriteFile(privKeyFile, jsonData, 0644)
		if err != nil {
			log.Printf("[%s] err:%+v", rtkMisc.GetFuncInfo(), err)
			panic(err)
		}
		return priv
	}

	priv, err = UnmarshalPrivateKeyFromPEM(content)
	if err != nil {
		log.Printf("[%s] err:%+v", rtkMisc.GetFuncInfo(), err)
		panic(err)
	}

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
	return rtkMisc.ConcatIP(ip, port)
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

func GetClientInfo(id string) (rtkMisc.ClientInfo, error) {
	rtkGlobal.ClientListRWMutex.RLock()
	defer rtkGlobal.ClientListRWMutex.RUnlock()

	if val, ok := rtkGlobal.ClientInfoMap[id]; ok {
		return val, nil
	}
	return rtkMisc.ClientInfo{}, errors.New(fmt.Sprintf("not found ClientInfo by id:%s", id))
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

func InsertClientInfoMap(id, ipAddr, platform, name, srcPortType, ver string) {
	rtkGlobal.ClientListRWMutex.Lock()
	defer rtkGlobal.ClientListRWMutex.Unlock()

	if _, ok := rtkGlobal.ClientInfoMap[id]; !ok {
		rtkGlobal.ClientInfoMap[id] = rtkMisc.ClientInfo{ID: id, IpAddr: ipAddr, Platform: platform, DeviceName: name, SourcePortType: srcPortType, Version: ver}
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
		clientList += val.DeviceName + "#"
		clientList += val.SourcePortType + ","
	}
	return strings.Trim(clientList, ",")
}

func GetClientListEx() string {
	clientLstInfo := rtkCommon.ClientListInfo{
		TimeStamp:  time.Now().UnixMilli(),
		ID:         rtkGlobal.NodeInfo.ID,
		IpAddr:     rtkMisc.ConcatIP(rtkGlobal.NodeInfo.IPAddr.PublicIP, rtkGlobal.NodeInfo.IPAddr.PublicPort),
		ClientList: make([]rtkMisc.ClientInfo, 0),
	}
	rtkGlobal.ClientListRWMutex.RLock()
	for _, val := range rtkGlobal.ClientInfoMap {
		clientLstInfo.ClientList = append(clientLstInfo.ClientList, rtkMisc.ClientInfo{
			ID:             val.ID,
			IpAddr:         val.IpAddr,
			Platform:       val.Platform,
			DeviceName:     val.DeviceName,
			SourcePortType: val.SourcePortType,
			Version:        val.Version,
		})
	}
	rtkGlobal.ClientListRWMutex.RUnlock()

	encodedData, err := json.Marshal(clientLstInfo)
	if err != nil {
		log.Println("Failed to Marshal ClientListInfo data, err:", err)
		return ""
	}
	return string(encodedData)
}

func GetClientCount() int {
	rtkGlobal.ClientListRWMutex.RLock()
	defer rtkGlobal.ClientListRWMutex.RUnlock()

	return len(rtkGlobal.ClientInfoMap)
}

func GetClientMap() map[string]rtkMisc.ClientInfo {
	rtkGlobal.ClientListRWMutex.RLock()
	defer rtkGlobal.ClientListRWMutex.RUnlock()

	return rtkGlobal.ClientInfoMap
}

func WalkPath(dirPath string, pathList *[]string, fileInfoList *[]rtkCommon.FileInfo, totalSize *uint64) error {
	rootPath := filepath.Dir(dirPath)

	// TODO: Need to be compatible with incompatible system separators
	dirPath = strings.ReplaceAll(dirPath, "/", "\\")

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			dstPath, bExsit := strings.CutPrefix(path, rootPath)
			if bExsit {
				*pathList = append(*pathList, dstPath)
			} else {
				log.Printf("full path:[%s] CutPrefix:[%s] error\n", dstPath, rootPath)
			}
		} else {
			fileSize := info.Size()
			dstFile, bOk := strings.CutPrefix(path, rootPath)
			if bOk {
				file := rtkCommon.FileInfo{
					FileSize_: rtkCommon.FileSize{
						SizeHigh: uint32(fileSize >> 32),
						SizeLow:  uint32(fileSize & 0xFFFFFFFF),
					},
					FilePath: path,
					FileName: dstFile,
				}
				*totalSize += uint64(fileSize)
				*fileInfoList = append(*fileInfoList, file)
			} else {
				log.Printf("file full path:[%s] CutPrefix:[%s] error\n", dstFile, rootPath)
			}

		}
		return nil
	})

	if err != nil {
		log.Printf("[%s] Walk path:[%s] Error:%+v\n", rtkMisc.GetFuncInfo(), dirPath, err)
	}

	return err
}

func ClearSrcFileListFullPath(srcFileList *[]rtkCommon.FileInfo) []rtkCommon.FileInfo {
	dstSrcList := make([]rtkCommon.FileInfo, 0)
	for _, fileInfo := range *srcFileList {
		dstSrcList = append(dstSrcList, rtkCommon.FileInfo{
			FileSize_: rtkCommon.FileSize{
				SizeHigh: fileInfo.FileSize_.SizeHigh,
				SizeLow:  fileInfo.FileSize_.SizeLow,
			},
			FilePath: "",
			FileName: fileInfo.FileName,
		})
	}
	return dstSrcList
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

func addSuffixBeforeExt(path, suffix string) string {
	ext := filepath.Ext(path)
	name := strings.TrimSuffix(path, ext)
	return fmt.Sprintf("%s%s%s", name, suffix, ext)
}

func GetTargetDstPathName(dstFullPath, dstFileName string) (string, string) {
	index := uint(0)
	var dstPath string

	for {
		if index == 0 {
			dstPath = dstFullPath
		} else {
			dstPath = addSuffixBeforeExt(dstFullPath, fmt.Sprintf(" (%d)", index))
		}
		if !rtkMisc.FileExists(dstPath) {
			if index == 0 {
				return dstPath, dstFileName
			} else {
				return dstPath, addSuffixBeforeExt(dstFileName, fmt.Sprintf(" (%d)", index))
			}

		}
		index++
	}
}

func WithCancelSource(parent context.Context) (context.Context, func(rtkCommon.CancelBusinessSource)) {
	ctx, cancel := context.WithCancel(parent)
	cc := &rtkCommon.CustomContext{
		Context: ctx,
		Mutex:   &sync.Mutex{},
		Source:  0,
	}

	return cc, func(source rtkCommon.CancelBusinessSource) {
		cc.Mutex.Lock()
		defer cc.Mutex.Unlock()
		cc.Source = source
		cancel()
	}
}

func GetCancelSource(ctx context.Context) (rtkCommon.CancelBusinessSource, bool) {
	if cc, ok := ctx.(*rtkCommon.CustomContext); ok {
		cc.Mutex.Lock()
		defer cc.Mutex.Unlock()
		return cc.Source, cc.Source != 0
	}
	return 0, false
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
	DeviceSrcAndPort  rtkMisc.SourcePort
)

func InitDeviceInfo(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Printf("[%s %d] Err: Not found device table: %s", rtkMisc.GetFuncName(), rtkMisc.GetLine(), filename)
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
				log.Printf("[%s %d] Err: Invalid param count(port), please check .DeviceInfo file", rtkMisc.GetFuncName(), rtkMisc.GetLine())
				return
			}

			DeviceStaticPort = string(port[1])

			if DeviceStaticPort == "" {
				log.Printf("[%s %d] Err: Empty staticPort, please check .DeviceInfo file", rtkMisc.GetFuncName(), rtkMisc.GetLine())
				return
			}
		} else { // Get device info
			parts := strings.SplitN(line, ":", 3)
			if len(parts) < 3 {
				log.Printf("[%s %d] Err: Invalid param count, please check .DeviceInfo file", rtkMisc.GetFuncName(), rtkMisc.GetLine())
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

func InitDeviceSrcAndPort(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Printf("[%s] Err: Not found device src and port: %s", rtkMisc.GetFuncInfo(), filename)
		return
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		srcAndPort := strings.SplitN(line, ",", 2)
		if len(srcAndPort) < 2 {
			log.Printf("[%s] Err: Invalid param count(srcAndPort), please check .DeviceSrcPort file", rtkMisc.GetFuncInfo())
			return
		}

		if srcAndPort[0] == "" || srcAndPort[1] == "" {
			log.Printf("[%s] Err: Empty src port, please check .DeviceSrcPort file", rtkMisc.GetFuncInfo())
			return
		}

		src, err := strconv.Atoi(srcAndPort[0])
		if err != nil {
			log.Printf("[%s] Err: Invalid source format: %s", rtkMisc.GetFuncInfo(), srcAndPort[0])
			return
		}
		port, err := strconv.Atoi(srcAndPort[1])
		if err != nil {
			log.Printf("[%s] Err: Invalid port format: %s", rtkMisc.GetFuncInfo(), srcAndPort[1])
			return
		}

		DeviceSrcAndPort.Source = src
		DeviceSrcAndPort.Port = port
		log.Printf("[%s] Parse \"%s\" successfully. (source,port)=(%d,%d)", rtkMisc.GetFuncInfo(), filename, src, port)
		break
	}
}

func GetDeviceSrcPort() rtkMisc.SourcePort {
	return DeviceSrcAndPort
}

// Deprecated with InitDeviceInfo
func InitDeviceTable(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Printf("[%s %d] Err: Not found device table: %s", rtkMisc.GetFuncName(), rtkMisc.GetLine(), filename)
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
			log.Printf("[%s %d] Err: Invalid param count", rtkMisc.GetFuncName(), rtkMisc.GetLine())
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

func ReadDiasID(filename string) string {
	var err error
	var content []byte
	content, err = os.ReadFile(filename)
	if err != nil {
		log.Printf("[%s] diasID_Hack_file [%s]", rtkMisc.GetFuncInfo(), err)
		return ""
	}
	log.Printf("[%s]", string(content[:]))
	return string(content[:])
}
