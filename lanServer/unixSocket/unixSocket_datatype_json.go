package unixSocket

import (
	"encoding/json"
	"errors"
)

type UnixSocketJson struct {
	Header  string          `json:"Header"`
	Type    int             `json:"Type"`
	Code    int             `json:"Code"`
	Content json.RawMessage `json:"Content"`
}

type UnixSocketUpdateDeviceNameNotiMsg struct {
	Source int    `json:"Source"`
	Port   int    `json:"Port"`
	Name   string `json:"Name"`
}

type UnixSocketDragFileStartNotiMsg struct {
	Source   int `json:"Source"`
	Port     int `json:"Port"`
	HorzSize int `json:"HorzSize"`
	VertSize int `json:"VertSize"`
	PosX     int `json:"PosX"`
	PosY     int `json:"PosY"`
}

type UnixSocketDragFileEndNotiMsg struct {
	Source int `json:"Source"`
	Port   int `json:"Port"`
}

type UnixSocketGetDiasIdRespMsg struct {
	ID string `json:"ID"`
}

func buildUnixSocketJson(msgType, msgCode int, content []byte) (UnixSocketJson, error) {
	var msg UnixSocketJson
	if msgType >= UNIX_SOCKET_TYPE_MAX {
		return msg, errors.New("Error: invalid UnixSocketJson msg type")
	}
	if msgCode >= UNIX_SOCKET_CODE_MAX {
		return msg, errors.New("Error: invalid UnixSocketJson msg type")
	}

	msg.Header = string(kHeader[:])
	msg.Type = msgType
	msg.Code = msgCode
	msg.Content = content
	return msg, nil
}

func buildUnixSocketUpdateDeviceNameNotiMsg(src, port int, name string) (UnixSocketJson, error) {
	var content UnixSocketUpdateDeviceNameNotiMsg
	content.Source = src
	content.Port = port
	content.Name = name

	var jsonData UnixSocketJson
	contentByte, err := json.Marshal(content)
	if err != nil {
		return jsonData, err
	}

	jsonData, errJson := buildUnixSocketJson(UNIX_SOCKET_TYPE_NOTIFY, UNIX_SOCKET_CODE_UPDATE_DEVICE_NAME, contentByte)
	if errJson != nil {
		return jsonData, errJson
	}

	return jsonData, nil
}

func buildUnixSocketDragFileStartNotiMsg(src, port int, horzSize, vertSize, posX, posY int) (UnixSocketJson, error) {
	var content UnixSocketDragFileStartNotiMsg
	content.Source = src
	content.Port = port
	content.HorzSize = horzSize
	content.VertSize = vertSize
	content.PosX = posX
	content.PosY = posY

	var jsonData UnixSocketJson
	contentByte, err := json.Marshal(content)
	if err != nil {
		return jsonData, err
	}

	jsonData, errJson := buildUnixSocketJson(UNIX_SOCKET_TYPE_NOTIFY, UNIX_SOCKET_CODE_DRAG_FILE_START, contentByte)
	if errJson != nil {
		return jsonData, errJson
	}

	return jsonData, nil
}

func buildUnixSocketGetDiasIdRespMsg(id string) (UnixSocketJson, error) {
	var content UnixSocketGetDiasIdRespMsg
	content.ID = id

	var jsonData UnixSocketJson
	contentByte, err := json.Marshal(content)
	if err != nil {
		return jsonData, err
	}

	jsonData, errJson := buildUnixSocketJson(UNIX_SOCKET_TYPE_RESPONSE, UNIX_SOCKET_CODE_GET_DIAS_ID, contentByte)
	if errJson != nil {
		return jsonData, errJson
	}

	return jsonData, nil
}
