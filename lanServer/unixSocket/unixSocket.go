package unixSocket

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	rtkClientManager "rtk-cross-share/lanServer/clientManager"
	rtkdbManager "rtk-cross-share/lanServer/dbManager"
	rtkGlobal "rtk-cross-share/lanServer/global"
	rtkMisc "rtk-cross-share/misc"
	"sync"
)

type DragFileSrcInfo struct {
	index int
	id    string
}

var (
	mDragFileSrcInfo DragFileSrcInfo
)

type SocketNodeType int

const (
	UNIX_SOCKET_NODE_TYPE_DDCCI             SocketNodeType = 1
	UNIX_SOCKET_NODE_TYPE_JAVA_VIEWMANAGER  SocketNodeType = 2
	UNIX_SOCKET_NODE_TYPE_JAVA_SOURCEPLAYER SocketNodeType = 3
)

type SourcePort struct {
	source int
	port   int
}

type ConnMapKey struct {
	nodeType   SocketNodeType
	srcAndPort SourcePort
}

const (
	SOCKET_PATH_DDCCI             = "ddcci2go_%d_%d.sock"
	SOCKET_PATH_JAVA_VIEWMANAGER  = "go2java_viewmanager.sock"
	SOCKET_PATH_JAVA_SOURCEPLAYER = "go2java_sourceplayer_%d_%d.sock"
	SOCKET_DATA_BUFFER_SIZE       = 1024
)

type ConnMap struct {
	m sync.Map
}

var (
	mConnMap                      ConnMap
	COMM_MAP_KEY_JAVA_VIEWMANAGER = ConnMapKey{
		nodeType:   UNIX_SOCKET_NODE_TYPE_JAVA_VIEWMANAGER,
		srcAndPort: SourcePort{0, 0},
	}
)

func (t SocketNodeType) toString() string {
	switch t {
	case UNIX_SOCKET_NODE_TYPE_DDCCI:
		return "DDCCI"
	case UNIX_SOCKET_NODE_TYPE_JAVA_VIEWMANAGER:
		return "ViewManager"
	case UNIX_SOCKET_NODE_TYPE_JAVA_SOURCEPLAYER:
		return "SourcePlayer"
	default:
		return "Unknown"
	}
}

func (sp *SourcePort) toString() string {
	return fmt.Sprintf("%d,%d", sp.source, sp.port)
}

func (cm *ConnMap) get(key ConnMapKey) (net.Conn, bool, error) {
	val, ok := cm.m.Load(key)
	if !ok {
		return nil, false, nil
	}
	conn, ok := val.(net.Conn)
	if !ok {
		return nil, false, errors.New("[UnixSocket] Get conn failed: type failed")
	}

	return conn, true, nil
}

func (cm *ConnMap) add(key ConnMapKey, conn net.Conn) error {
	if conn == nil {
		return errors.New("[UnixSocket] Add conn failed: null connection")
	}

	preConn, ret, err := cm.get(key)
	if err != nil {
		return err
	}
	if ret && (preConn != nil) {
		log.Println("[UnixSocket] Add conn: conn existed. Disconnect previous conn")
		preConn.Close()
	}

	cm.m.Store(key, conn)
	return nil
}

func (cm *ConnMap) remove(key ConnMapKey) error {
	conn, ret, err := cm.get(key)
	if err != nil {
		return err
	}
	if ret && (conn != nil) {
		conn.Close()
		cm.m.Delete(key)
	}
	return nil
}

func writeData(conn net.Conn, data []byte) error {
	dataLen := len(data)
	if dataLen > SOCKET_DATA_BUFFER_SIZE {
		return fmt.Errorf("[UnixSocket] Error: data size reached limitation: %d", SOCKET_DATA_BUFFER_SIZE)
	}

	if dataLen < SOCKET_DATA_BUFFER_SIZE {
		padding := make([]byte, SOCKET_DATA_BUFFER_SIZE-len(data))
		data = append(data, padding...)
	}

	// DEBUG
	log.Printf("[UnixSocket] Write: % X", data)
	_, err := conn.Write(data)
	if err != nil {
		return fmt.Errorf("[UnixSocket] Error: write authDevice Resp failed: %s", err.Error())
	}

	return nil
}

func writeFailedData(key ConnMapKey) {
	conn, ret, err := mConnMap.get(key)
	if err == nil && ret && conn != nil {
		writeData(conn, []byte{0})
	}
}

func readData(key ConnMapKey, conn net.Conn) {
	buffer := make([]byte, SOCKET_DATA_BUFFER_SIZE)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			log.Printf("[UnixSocket] read failed: %s", err.Error())
			return
		}

		// DEBUG
		log.Printf("[UnixSocket] (type,src,port):(%s,%s) Read: % X", key.nodeType.toString(), key.srcAndPort.toString(), buffer[:n])
		if key.nodeType == UNIX_SOCKET_NODE_TYPE_DDCCI {
			readDdcciData(key, buffer[:n])
		} else if key.nodeType == UNIX_SOCKET_NODE_TYPE_JAVA_VIEWMANAGER || key.nodeType == UNIX_SOCKET_NODE_TYPE_JAVA_SOURCEPLAYER {
			readJavaData(key, buffer[:n])
		}
	}
}

func getSocketPath(key ConnMapKey) string {
	switch key.nodeType {
	case UNIX_SOCKET_NODE_TYPE_DDCCI:
		return rtkGlobal.SOCKET_PATH_ROOT + fmt.Sprintf(SOCKET_PATH_DDCCI, key.srcAndPort.source, key.srcAndPort.port)
	case UNIX_SOCKET_NODE_TYPE_JAVA_VIEWMANAGER:
		return rtkGlobal.SOCKET_PATH_ROOT + SOCKET_PATH_JAVA_VIEWMANAGER
	case UNIX_SOCKET_NODE_TYPE_JAVA_SOURCEPLAYER:
		return rtkGlobal.SOCKET_PATH_ROOT + fmt.Sprintf(SOCKET_PATH_JAVA_SOURCEPLAYER, key.srcAndPort.source, key.srcAndPort.port)
	default:
		return ""
	}
}

func buildListener(sockPath string) (net.Listener, error) {
	os.Remove(sockPath)

	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		log.Printf("Build DDCCI to GO socket failed: %s", err.Error())
		return nil, err
	}

	if listener == nil {
		return nil, errors.New("Socket listener is null")
	}
	return listener, nil
}

// ========================
// # DDCCI Handler
// ========================

func BuildDdcciListener() {
	for port := 0; port < rtkGlobal.Port_max; port++ {
		rtkMisc.GoSafe(func() { buildDdcciListenerInternal(SourcePort{rtkGlobal.Src_HDMI, port}) })
		rtkMisc.GoSafe(func() { buildDdcciListenerInternal(SourcePort{rtkGlobal.Src_DP, port}) })
	}
}

func buildDdcciListenerInternal(srcPort SourcePort) {
	var key ConnMapKey
	key.nodeType = UNIX_SOCKET_NODE_TYPE_DDCCI
	key.srcAndPort = srcPort
	socketPath := getSocketPath(key)
	if socketPath == "" {
		log.Printf("[UnixSocket] get empty socket path. Type: %s", key.nodeType.toString())
		return
	}

	listener, err := buildListener(socketPath)
	if err != nil {
		log.Printf("Build DDCCI to GO socket failed: %s", err.Error())
		return
	}

	defer listener.Close()

	for {
		log.Printf("[UnixSocket] (DDCCI to GO) (%s) listening...", srcPort.toString())
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("[UnixSocket] conn failed: %s", err.Error())
			continue
		}
		log.Printf("[UnixSocket] (DDCCI to GO) (%s) connected!", srcPort.toString())

		mConnMap.add(key, conn)
		readData(key, conn)
		log.Printf("[UnixSocket] (DDCCI to GO)(source,port)=(%s) disconnected", srcPort.toString())
		mConnMap.remove(key)
	}
}

func readDdcciData(key ConnMapKey, data []byte) {
	header, err := toUnixSocketHeader(data)
	if err != nil {
		log.Printf("[UnixSocket] Error: invalid DDCCI data: %s", err.Error())
		return
	}

	switch header.msgCode {
	case UNIX_SOCKET_CODE_AUTH_DEVICE:
		if header.msgType == UNIX_SOCKET_TYPE_REQUEST {
			if ddcciDataHandlerAuthDevice(key, header, data) == false {
				writeFailedData(key)
			}
		}
		break
	case UNIX_SOCKET_CODE_UPDATE_MOUSE_POS:
		if header.msgType == UNIX_SOCKET_TYPE_NOTIFY {
			ddcciDataHandlerUpdateMousePos(key, header, data)
		}
	default:
		log.Printf("[UnixSocket][%s] Error: Unknown msg code: %d", rtkMisc.GetFuncInfo(), header.msgCode)
		break
	}
}

func ddcciDataHandlerAuthDevice(key ConnMapKey, header UnixSocketHeaderByte, data []byte) bool {
	conn, ret, err := mConnMap.get(key)
	if err != nil {
		log.Printf("[UnixSocket][%s] Error: get conn failed: %s", rtkMisc.GetFuncInfo(), err.Error())
		return false
	}
	if !ret {
		log.Printf("[UnixSocket][%s] Error: get conn failed: Empty conn", rtkMisc.GetFuncInfo())
		return false
	}

	authDeviceMsg, errReqMsg := toUnixSocketAuthDeviceReqMsg(header, data)
	if errReqMsg != nil {
		log.Printf("[UnixSocket][%s] Error: invalid AuthDevice Req data: %s", rtkMisc.GetFuncInfo(), errReqMsg.Error())
		return false
	}

	pkIdx := int(authDeviceMsg.clientIdx)
	authStatus := true
	errAuthAndSrcPort := rtkdbManager.UpdateAuthAndSrcPort(pkIdx, authStatus, int(authDeviceMsg.source), int(authDeviceMsg.port))
	if errAuthAndSrcPort != rtkMisc.SUCCESS {
		log.Printf("[UnixSocket][%s] Error: update auth status and source port failed: %s", rtkMisc.GetFuncInfo(), rtkMisc.GetResponse(errAuthAndSrcPort).Msg)
		return false
	}

	respMsg, errRespMsg := buildUnixSocketAuthDeviceRespMsg(authStatus)
	if errRespMsg != nil {
		log.Printf("[UnixSocket][%s] Error: build AuthDevice Resp data failed: %s", rtkMisc.GetFuncInfo(), errRespMsg.Error())
		return false
	}

	errWrite := writeData(conn, respMsg.toByte())
	if errWrite != nil {
		log.Println(errWrite.Error())
		return false
	}

	deviceName, errDeviceName := rtkdbManager.QueryDeviceName(pkIdx)
	if errDeviceName != nil {
		log.Printf("[UnixSocket][%s] Error: query device name failed: %s", rtkMisc.GetFuncInfo(), errDeviceName.Error())
		return false
	}
	SendUpdateDeviceName(key.srcAndPort.source, key.srcAndPort.port, deviceName)
	return true
}

func ddcciDataHandlerUpdateMousePos(key ConnMapKey, header UnixSocketHeaderByte, data []byte) bool {
	updateMousePos, errUpdateMousePosMsg := toUnixSocketUpdateMousePosNotiMsg(header, data)
	if errUpdateMousePosMsg != nil {
		log.Printf("[UnixSocket][%s] Error: invalid Noti data: %s", rtkMisc.GetFuncInfo(), errUpdateMousePosMsg.Error())
		return false
	}

	clientIdx, clientId, err := rtkdbManager.QueryClientBySrcPort(key.srcAndPort.source, key.srcAndPort.port)
	if err != nil {
		log.Printf("[UnixSocket][%s] Error: get client by (source,port):(%d,%d) failed: %s", rtkMisc.GetFuncInfo(), key.srcAndPort.source, key.srcAndPort.port, err.Error())
		return false
	}

	if SendDragFileStart(key.srcAndPort.source, key.srcAndPort.port, int(updateMousePos.horzSize), int(updateMousePos.vertSize), int(updateMousePos.posX), int(updateMousePos.posY)) {
		mDragFileSrcInfo = DragFileSrcInfo{clientIdx, clientId}
	}
	return true
}

// ========================
// # JAVA Handler
// ========================

func BuildJavaListener() {
	rtkMisc.GoSafe(func() { buildJavaListenerInternal(COMM_MAP_KEY_JAVA_VIEWMANAGER) })

	var key ConnMapKey
	key.nodeType = UNIX_SOCKET_NODE_TYPE_JAVA_SOURCEPLAYER
	for port := 0; port < rtkGlobal.Port_max; port++ {
		rtkMisc.GoSafe(func() {
			key.srcAndPort = SourcePort{rtkGlobal.Src_HDMI, port}
			buildJavaListenerInternal(key)
		})
		rtkMisc.GoSafe(func() {
			key.srcAndPort = SourcePort{rtkGlobal.Src_DP, port}
			buildJavaListenerInternal(key)
		})
	}
	rtkMisc.GoSafe(func() {
		key.srcAndPort = SourcePort{rtkGlobal.Src_STREAM, rtkGlobal.Port_subType_Miracast}
		buildJavaListenerInternal(key)
	})
}

func buildJavaListenerInternal(key ConnMapKey) {
	socketPath := getSocketPath(key)
	if socketPath == "" {
		log.Printf("[UnixSocket] get empty socket path. Type: %s", key.nodeType.toString())
		return
	}

	listener, err := buildListener(socketPath)
	if err != nil {
		log.Printf("Build JAVA to GO socket failed: %s", err.Error())
		return
	}

	defer listener.Close()

	for {
		log.Printf("[UnixSocket] (JAVA to GO) (%s)(%s) listening...", key.nodeType.toString(), key.srcAndPort.toString())
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("[UnixSocket] conn failed %s", err.Error())
			continue
		}

		log.Printf("[UnixSocket] (JAVA to GO) (%s)(%s) connected!", key.nodeType.toString(), key.srcAndPort.toString())
		mConnMap.add(key, conn)
		readData(key, conn)
		log.Printf("[UnixSocket] (JAVA to GO) (%s)(%s) disconnected", key.nodeType.toString(), key.srcAndPort.toString())
		mConnMap.remove(key)
	}
}

func readJavaData(key ConnMapKey, data []byte) {
	data = bytes.Trim(data, "\x00")
	// DEBUG
	log.Printf("[UnixSocket][%s] read from Java: %s", key.nodeType.toString(), string(data))

	var socketJson UnixSocketJson
	if err := json.Unmarshal(data, &socketJson); err != nil {
		log.Printf("[UnixSocket][%s] Error: invalid Java json data: %s", rtkMisc.GetFuncInfo(), err.Error())
		return
	}

	if socketJson.Header != string(kHeader[:]) {
		log.Printf("[UnixSocket][%s] Error: invalid json header", rtkMisc.GetFuncInfo())
		return
	}

	switch socketJson.Code {
	case UNIX_SOCKET_CODE_DRAG_FILE_END:
		if socketJson.Type == UNIX_SOCKET_TYPE_NOTIFY {
			javaDataHandlerDragFileEnd(socketJson.Content)
		}
		break
	case UNIX_SOCKET_CODE_GET_DIAS_ID:
		if socketJson.Type == UNIX_SOCKET_TYPE_REQUEST {
			javaDataHandlerGetDiasId(key)
		}
		break
	default:
		log.Printf("[UnixSocket][%s] Error: Unknown json code: %d", rtkMisc.GetFuncInfo(), socketJson.Code)
		break
	}
}

func javaDataHandlerDragFileEnd(data json.RawMessage) {
	var socketJson UnixSocketDragFileEndNotiMsg
	if err := json.Unmarshal(data, &socketJson); err != nil {
		log.Printf("[UnixSocket][%s] Error: invalid Java json data: %s", rtkMisc.GetFuncInfo(), err.Error())
		return
	}

	_, dstClientId, err := rtkdbManager.QueryClientBySrcPort(socketJson.Source, socketJson.Port)
	if err != nil {
		log.Printf("[UnixSocket][%s] Error: get client by (source,port):(%d,%d) failed: %s", rtkMisc.GetFuncInfo(), socketJson.Source, socketJson.Port, err.Error())
		return
	}

	rtkClientManager.SendDragFileEvent(mDragFileSrcInfo.id, dstClientId, uint32(mDragFileSrcInfo.index))
}

func javaDataHandlerGetDiasId(key ConnMapKey) {
	conn, ret, err := mConnMap.get(key)
	if err != nil {
		log.Printf("[UnixSocket][%s] Error: get conn failed: %s", rtkMisc.GetFuncInfo(), err.Error())
		return
	}
	if !ret {
		log.Printf("[UnixSocket][%s] Error: get conn failed: Empty conn", rtkMisc.GetFuncInfo())
		return
	}

	if rtkGlobal.ServerMdnsId == "" {
		log.Printf("[UnixSocket][%s] Error: empty mDNS ID", rtkMisc.GetFuncInfo())
		return
	}

	jsonData, errBuild := buildUnixSocketGetDiasIdRespMsg(rtkGlobal.ServerMdnsId)
	if errBuild != nil {
		log.Printf("[UnixSocket][%s] Error: build msg failed: %s", rtkMisc.GetFuncInfo(), errBuild.Error())
		return
	}

	data, errJson := json.Marshal(jsonData)
	if errJson != nil {
		log.Printf("[UnixSocket][%s] Error: Json failed: %s", rtkMisc.GetFuncInfo(), errJson.Error())
		return
	}

	// DEBUG
	log.Printf("[UnixSocket] write to Java((%s,%s) : %s", key.nodeType.toString(), key.srcAndPort.toString(), string(data))
	errWrite := writeData(conn, data)
	if errWrite != nil {
		log.Printf("[UnixSocket][%s] Error: write failed: %s", rtkMisc.GetFuncInfo(), errWrite.Error())
	}
}

func SendUpdateDeviceName(targetSrc, targetPort int, name string) bool {
	conn, ret, err := mConnMap.get(COMM_MAP_KEY_JAVA_VIEWMANAGER)
	if err != nil {
		log.Printf("[UnixSocket][%s] Error: get conn failed: %s", rtkMisc.GetFuncInfo(), err.Error())
		return false
	}
	if !ret {
		log.Printf("[UnixSocket][%s] Error: get conn failed: Empty conn", rtkMisc.GetFuncInfo())
		return false
	}

	jsonData, errBuild := buildUnixSocketUpdateDeviceNameNotiMsg(targetSrc, targetPort, name)
	if errBuild != nil {
		log.Printf("[UnixSocket][%s] Error: build msg failed: %s", rtkMisc.GetFuncInfo(), errBuild.Error())
		return false
	}

	data, errJson := json.Marshal(jsonData)
	if errJson != nil {
		log.Printf("[UnixSocket][%s] Error: Json failed: %s", rtkMisc.GetFuncInfo(), errJson.Error())
		return false
	}

	// DEBUG
	log.Printf("[UnixSocket] write to Java((%s,%s) : %s", COMM_MAP_KEY_JAVA_VIEWMANAGER.nodeType.toString(), COMM_MAP_KEY_JAVA_VIEWMANAGER.srcAndPort.toString(), string(data))
	errWrite := writeData(conn, data)
	if errWrite != nil {
		log.Printf("[UnixSocket][%s] Error: write failed: %s", rtkMisc.GetFuncInfo(), errWrite.Error())
		return false
	}

	return true
}

func SendDragFileStart(src, port, horzSize, vertSize, posX, posY int) bool {
	key := ConnMapKey{
		nodeType:   UNIX_SOCKET_NODE_TYPE_JAVA_SOURCEPLAYER,
		srcAndPort: SourcePort{src, port},
	}
	conn, ret, err := mConnMap.get(key)
	if err != nil {
		log.Printf("[UnixSocket][%s] Error: get conn failed: %s", rtkMisc.GetFuncInfo(), err.Error())
		return false
	}
	if !ret {
		log.Printf("[UnixSocket][%s] Error: get conn failed: Empty conn", rtkMisc.GetFuncInfo())
		return false
	}

	jsonData, errBuild := buildUnixSocketDragFileStartNotiMsg(src, port, horzSize, vertSize, posX, posY)
	if errBuild != nil {
		log.Printf("[UnixSocket][%s] Error: build msg failed: %s", rtkMisc.GetFuncInfo(), errBuild.Error())
		return false
	}

	data, errJson := json.Marshal(jsonData)
	if errJson != nil {
		log.Printf("[UnixSocket][%s] Error: Json failed: %s", rtkMisc.GetFuncInfo(), errJson.Error())
		return false
	}

	// DEBUG
	log.Printf("[UnixSocket] write to Java((%s,%s) : %s", key.nodeType.toString(), key.srcAndPort.toString(), string(data))
	errWrite := writeData(conn, data)
	if errWrite != nil {
		log.Printf("[UnixSocket][%s] Error: write failed: %s", rtkMisc.GetFuncInfo(), errWrite.Error())
		return false
	}

	return true
}

// ==================================
// # Test Function
// ==================================
func TestGetDiasId() {
	if rtkGlobal.ServerMdnsId == "" {
		log.Printf("[UnixSocket][%s] Error: empty mDNS ID", rtkMisc.GetFuncInfo())
		return
	}

	jsonData, errBuild := buildUnixSocketGetDiasIdRespMsg(rtkGlobal.ServerMdnsId)
	if errBuild != nil {
		log.Printf("[UnixSocket][%s] Error: build msg failed: %s", rtkMisc.GetFuncInfo(), errBuild.Error())
		return
	}

	data, errJson := json.Marshal(jsonData)
	if errJson != nil {
		log.Printf("[UnixSocket][%s] Error: Json failed: %s", rtkMisc.GetFuncInfo(), errJson.Error())
		return
	}

	// DEBUG
	log.Printf("[UnixSocket] write to Java: %s", string(data))
}

func TestSendUpdateDeviceName(targetSrc, targetPort int, name string) bool {
	jsonData, errBuild := buildUnixSocketUpdateDeviceNameNotiMsg(targetSrc, targetPort, name)
	if errBuild != nil {
		log.Printf("[UnixSocket][%s] Error: build msg failed: %s", rtkMisc.GetFuncInfo(), errBuild.Error())
		return false
	}

	data, errJson := json.Marshal(jsonData)
	if errJson != nil {
		log.Printf("[UnixSocket][%s] Error: Json failed: %s", rtkMisc.GetFuncInfo(), errJson.Error())
		return false
	}

	// DEBUG
	log.Printf("[UnixSocket] write to Java: %s", string(data))
	return true
}

func TestSendDragFileStart(src, port, horzSize, vertSize, posX, posY int) bool {
	jsonData, errBuild := buildUnixSocketDragFileStartNotiMsg(src, port, horzSize, vertSize, posX, posY)
	if errBuild != nil {
		log.Printf("[UnixSocket][%s] Error: build msg failed: %s", rtkMisc.GetFuncInfo(), errBuild.Error())
		return false
	}

	data, errJson := json.Marshal(jsonData)
	if errJson != nil {
		log.Printf("[UnixSocket][%s] Error: Json failed: %s", rtkMisc.GetFuncInfo(), errJson.Error())
		return false
	}

	// DEBUG
	log.Printf("[UnixSocket] write to Java: %s", string(data))
	return true
}
