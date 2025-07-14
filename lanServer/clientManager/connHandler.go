package clientManager

import (
	"bufio"
	"bytes"
	"log"
	"net"
	rtkMisc "rtk-cross-share/misc"
	"sync"
)

type clientConnInfo struct {
	conn      net.Conn
	timeStamp int64
}

var (
	clientConnMap   = make(map[string](clientConnInfo)) //KEY: ID
	clientConnMutex sync.RWMutex
)

func updateConn(id string, timestamp int64, conn net.Conn) {
	clientConnMutex.Lock()
	if oldClient, exsit := clientConnMap[id]; exsit {
		log.Printf("ID:[%s] old conn is exsit! close it first!", id)
		oldClient.conn.Close()
		delete(clientConnMap, id)
	}
	clientConnMap[id] = clientConnInfo{
		conn:      conn,
		timeStamp: timestamp,
	}
	clientConnMutex.Unlock()
	log.Printf("[%s] ID:[%s]  IPAddr:[%s] connection is update!", rtkMisc.GetFuncInfo(), id, conn.RemoteAddr())
}

func closeConn(id string, timestamp int64) bool {
	clientConnMutex.Lock()
	defer clientConnMutex.Unlock()

	if client, exsit := clientConnMap[id]; exsit {
		if client.timeStamp == timestamp {
			client.conn.Close()
			delete(clientConnMap, id)
			log.Printf("[%s] ID:[%s] connection is closed!", rtkMisc.GetFuncInfo(), id)
			return true
		}
	}
	return false
}

func getConn(id string) (net.Conn, bool) {
	clientConnMutex.RLock()
	defer clientConnMutex.RUnlock()
	if clientInfo, exsit := clientConnMap[id]; exsit {
		return clientInfo.conn, exsit
	}
	return nil, false
}

func write(b []byte, id string, timestamp int64) rtkMisc.CrossShareErr {
	clientConnMutex.RLock()
	defer clientConnMutex.RUnlock()
	clientInfo, exsit := clientConnMap[id]
	if !exsit {
		log.Printf("[%s] ID:[%s] get no connection!", rtkMisc.GetFuncInfo(), id)
		return rtkMisc.ERR_BIZ_S2C_GET_EMPTY_CONNECT
	}

	if timestamp > 0 && timestamp != clientInfo.timeStamp {
		log.Printf("[%s] ID:[%s] get connection is reset!", rtkMisc.GetFuncInfo(), id)
		return rtkMisc.ERR_BIZ_S2C_GET_CONNECT_RESET
	}

	encodedData := bytes.Trim(b, "\x00")
	_, err := clientInfo.conn.Write(append(encodedData, '\n'))
	if err != nil {
		log.Printf("[%s] ID:[%s] write Error:%+v ", rtkMisc.GetFuncInfo(), id, err.Error())
		return rtkMisc.ERR_NETWORK_S2C_WRITE
	}

	err = bufio.NewWriter(clientInfo.conn).Flush()
	if err != nil {
		log.Printf("[%s] ID:[%s] Flush Error:%+v ", rtkMisc.GetFuncInfo(), id, err.Error())
		return rtkMisc.ERR_NETWORK_S2C_FLUSH
	}

	return rtkMisc.SUCCESS
}
