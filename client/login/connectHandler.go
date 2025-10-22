package login

import (
	"bufio"
	"errors"
	"log"
	"net"
	rtkMisc "rtk-cross-share/misc"
	"sync"
	"time"
)

type safeConnect struct {
	connectMutex     sync.RWMutex
	connectLanServer net.Conn
	isAlive          bool
}

func (s *safeConnect) Reset(conn net.Conn) {
	s.connectMutex.Lock()
	defer s.connectMutex.Unlock()
	if conn == nil {
		log.Printf("[%s] new connectLanServer is null, Reset failed! ", rtkMisc.GetFuncInfo())
		return
	}

	if !s.isAlive {
		if s.connectLanServer != nil {
			s.connectLanServer.Close()
		}

		s.connectLanServer = conn
		s.isAlive = true
		log.Printf("[%s] connectLanServer Reset success! ", rtkMisc.GetFuncInfo())
	} else {
		if s.connectLanServer != nil {
			s.connectLanServer.Close()
		}
		s.connectLanServer = conn
		log.Printf("[%s] the old connectLanServer is alive, replace it!", rtkMisc.GetFuncInfo())
	}
}

func (s *safeConnect) GetConnect() net.Conn {
	s.connectMutex.RLock()
	defer s.connectMutex.RUnlock()

	if s.isAlive && s.connectLanServer != nil {
		return s.connectLanServer
	}

	return nil
}

func (s *safeConnect) Write(b []byte) rtkMisc.CrossShareErr {
	s.connectMutex.Lock()
	defer s.connectMutex.Unlock()
	if s.isAlive && s.connectLanServer != nil {
		writeDone := make(chan rtkMisc.CrossShareErr, 1)
		go func() {
			
			s.connectLanServer.SetWriteDeadline(time.Now().Add(5 * time.Second))
			_, err := s.connectLanServer.Write(append(b, '\n'))
			if err != nil {
				log.Printf("[%s] LanServer Write err:%+v", rtkMisc.GetFuncInfo(), err)
				writeDone <-  rtkMisc.ERR_NETWORK_C2S_WRITE
				return
			}

			err = bufio.NewWriter(s.connectLanServer).Flush()
			if err != nil {
				log.Printf("[%s] LanServer Flush Error:%+v ", rtkMisc.GetFuncInfo(), err.Error())
				writeDone <- rtkMisc.ERR_NETWORK_C2S_FLUSH
				return 
			}
			writeDone <- rtkMisc.SUCCESS
    		}()

		
		
		select {
		case err := <-writeDone:
			return err
		case <-time.After(5 * time.Second):
			return rtkMisc.ERR_NETWORK_C2S_WRITE_TIME_OUT
		}	
		
	}

	log.Printf("[%s] connectLanServer is not alive! Write failed!", rtkMisc.GetFuncInfo())
	return rtkMisc.ERR_BIZ_C2S_GET_EMPTY_CONNECT
}

func (s *safeConnect) Read(b *[]byte) (int, error) {
	s.connectMutex.Lock()
	defer s.connectMutex.Unlock()
	if s.isAlive {
		buf := bufio.NewReader(s.connectLanServer)
		readStrLine, err := buf.ReadString('\n')
		if err != nil {
			return 0, err
		}

		log.Printf("ReadString len[%d]", len(readStrLine))
		*b = []byte(readStrLine)
		return len(*b), nil
	}

	log.Printf("[%s] connectLanServer is not alive! Read failed!", rtkMisc.GetFuncInfo())
	return 0, errors.New("connectLanServer is not alive")
}

func (s *safeConnect) Close() error {
	s.connectMutex.Lock()
	defer s.connectMutex.Unlock()
	if s.isAlive {
		s.isAlive = false
		if s.connectLanServer != nil {
			if err := s.connectLanServer.Close(); err != nil {
				time.Sleep(100 * time.Millisecond)
				return s.connectLanServer.Close()
			} else {
				s.connectLanServer = nil
				return nil
			}
		}
	}

	return nil
}

func (s *safeConnect) IsAlive() bool {
	s.connectMutex.RLock()
	defer s.connectMutex.RUnlock()
	if s.isAlive && s.connectLanServer != nil {
		return true
	}

	return false
}

func (s *safeConnect) ConnectIPAddr() string {
	s.connectMutex.RLock()
	defer s.connectMutex.RUnlock()
	if s.connectLanServer != nil {
		return s.connectLanServer.RemoteAddr().String()
	}

	return ""
}
