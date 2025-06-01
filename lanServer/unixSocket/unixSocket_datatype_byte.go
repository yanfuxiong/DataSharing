package unixSocket

import (
	"bytes"
	"encoding/binary"
	"errors"
)

type UnixSocketHeaderByte struct {
	csHeader  [5]byte
	msgType   uint8
	msgCode   uint8
	msgLength uint32
}

type UnixSocketAuthDeviceReqMsg struct {
	header    UnixSocketHeaderByte
	clientIdx uint8
	source    uint8
	port      uint8
}

type UnixSocketAuthDeviceRespMsg struct {
	header     UnixSocketHeaderByte
	authStatus uint8
}

type UnixSocketUpdateMousePosNotiMsg struct {
	header   UnixSocketHeaderByte
	horzSize uint16
	vertSize uint16
	posX     int16
	posY     int16
}

func (h UnixSocketHeaderByte) toByte() []byte {
	var data []byte
	data = append(data, h.csHeader[:]...)
	data = append(data, h.msgType)
	data = append(data, h.msgCode)

	dataMsgLength := make([]byte, 4)
	binary.BigEndian.PutUint32(dataMsgLength, h.msgLength)
	data = append(data, dataMsgLength...)
	return data
}

func toUnixSocketHeader(data []byte) (UnixSocketHeaderByte, error) {
	var header UnixSocketHeaderByte
	if len(data) < UNIX_SOCKET_HEADER_LEN_POS {
		return header, errors.New("Error: invalid UnixSocketHeaderByte data")
	}

	if !bytes.Equal(data[:len(kHeader)], kHeader[:]) {
		return header, errors.New("Error: invalid UnixSocketHeaderByte header")
	}

	var pos = 0
	header.csHeader = kHeader
	pos += len(kHeader)
	header.msgType = data[pos]
	pos += 1
	header.msgCode = data[pos]
	pos += 1

	bufLen := bytes.NewReader(data[pos : pos+UNIX_SOCKET_HEADER_LEN_SIZE])
	err := binary.Read(bufLen, binary.BigEndian, &(header.msgLength))
	if err != nil {
		return header, err
	}

	return header, nil
}

func buildUnixSocketHeader(msgType, msgCode uint8) (UnixSocketHeaderByte, error) {
	var header UnixSocketHeaderByte
	if msgType >= UNIX_SOCKET_TYPE_MAX {
		return header, errors.New("Error: invalid UnixSocketHeaderByte msg type")
	}
	if msgCode >= UNIX_SOCKET_CODE_MAX {
		return header, errors.New("Error: invalid UnixSocketHeaderByte msg code")
	}

	header.csHeader = kHeader
	header.msgType = msgType
	header.msgCode = msgCode
	header.msgLength = 0

	return header, nil
}

func toUnixSocketAuthDeviceReqMsg(header UnixSocketHeaderByte, data []byte) (UnixSocketAuthDeviceReqMsg, error) {
	var msg UnixSocketAuthDeviceReqMsg

	if len(data) < (UNIX_SOCKET_HEADER_LEN_POS + UNIX_SOCKET_CONTENT_LEN_AUTH_DEVICE_REQ) {
		return msg, errors.New("Error: invalid UnixSocketAuthDeviceReqMsg data")
	}

	msg.header = header
	var pos = UNIX_SOCKET_HEADER_LEN_POS
	msg.clientIdx = data[pos]
	pos += 1
	msg.source = data[pos]
	pos += 1
	msg.port = data[pos]

	return msg, nil
}

func buildUnixSocketAuthDeviceRespMsg(authStatus bool) (UnixSocketAuthDeviceRespMsg, error) {
	var msg UnixSocketAuthDeviceRespMsg

	header, err := buildUnixSocketHeader(UNIX_SOCKET_TYPE_RESPONSE, UNIX_SOCKET_CODE_AUTH_DEVICE)
	if err != nil {
		return msg, err
	}

	msg.header = header
	msg.header.msgLength = UNIX_SOCKET_CONTENT_LEN_AUTH_DEVICE_RESP
	var status uint8 = 0
	if authStatus {
		status = 1
	}
	msg.authStatus = status
	return msg, nil
}

func (msg UnixSocketAuthDeviceReqMsg) toByte() []byte {
	var data []byte
	data = append(data, msg.header.toByte()...)
	data = append(data, msg.clientIdx)
	data = append(data, msg.source)
	data = append(data, msg.port)
	return data
}

func (msg UnixSocketAuthDeviceRespMsg) toByte() []byte {
	var data []byte
	data = append(data, msg.header.toByte()...)
	data = append(data, msg.authStatus)
	return data
}

func toUnixSocketUpdateMousePosNotiMsg(header UnixSocketHeaderByte, data []byte) (UnixSocketUpdateMousePosNotiMsg, error) {
	var msg UnixSocketUpdateMousePosNotiMsg

	if len(data) < (UNIX_SOCKET_HEADER_LEN_POS + UNIX_SOCKET_CONTENT_LEN_UPDATE_MOUSE_POS_REQ) {
		return msg, errors.New("Error: invalid UnixSocketUpdateMousePosNotiMsg data")
	}

	msg.header = header
	var pos = UNIX_SOCKET_HEADER_LEN_POS
	msg.horzSize = uint16(data[pos+1])<<8 | uint16(data[pos])
	pos += 2
	msg.vertSize = uint16(data[pos+1])<<8 | uint16(data[pos])
	pos += 2
	msg.posX = int16(data[pos+1])<<8 | int16(data[pos])
	pos += 2
	msg.posY = int16(data[pos+1])<<8 | int16(data[pos])

	return msg, nil
}
