package misc

// TODO: client and lanServer ERR CODE  are Define here, TSTAS-82

type CrossShareErr int

const SUCCESS CrossShareErr = iota + 100

type Response struct {
	Code CrossShareErr
	Msg  string
}

func GetResponse(code CrossShareErr) Response {
	info, ok := errInfoMap[code]
	if !ok {
		return Response{
			Code: code,
			Msg:  "unknown err message!",
		}
	}
	return Response{Code: code, Msg: info}
}

/***************************************  network info err code, begin with 3000  ****************************************************/
const ( // peer to peer connect error code
	ERR_NETWORK_P2P_OTHER CrossShareErr = iota + 3000
	ERR_NETWORK_P2P_CONNECT
	ERR_NETWORK_P2P_CONNECT_CANCEL
	ERR_NETWORK_P2P_CONNECT_DEADLINE
	ERR_NETWORK_P2P_OPEN_STREAM
	ERR_NETWORK_P2P_OPEN_STREAM_CANCEL
	ERR_NETWORK_P2P_OPEN_STREAM_DEADLINE
	ERR_NETWORK_P2P_TIMEOUT
	ERR_NETWORK_P2P_RESET
	ERR_NETWORK_P2P_EOF
	ERR_NETWORK_P2P_READER
	ERR_NETWORK_P2P_READER_CANCELED
	ERR_NETWORK_P2P_READER_DEADLINE
	ERR_NETWORK_P2P_READER_RESET
	ERR_NETWORK_P2P_WRITER
	ERR_NETWORK_P2P_WRITER_CANCELED
	ERR_NETWORK_P2P_WRITER_DEADLINE
	ERR_NETWORK_P2P_WRITER_RESET
	ERR_NETWORK_P2P_FLUSH
	ERR_NETWORK_P2P_INVALID
)

// client to  lanserver connect error code
const (
	ERR_NETWORK_C2S_OTHER CrossShareErr = iota + 3200
	ERR_NETWORK_C2S_RESOLVER
	ERR_NETWORK_C2S_BROWSER
	ERR_NETWORK_C2S_LOOKUP
	ERR_NETWORK_C2S_LOOKUP_TIMEOUT
	ERR_NETWORK_C2S_LOOKUP_INVALID
	ERR_NETWORK_C2S_DIAL
	ERR_NETWORK_C2S_DIAL_TIMEOUT
	ERR_NETWORK_C2S_WRITE
	ERR_NETWORK_C2S_FLUSH
	ERR_NETWORK_C2S_READ
)

// lanserver to client connect error
const (
	ERR_NETWORK_S2C_OTHER CrossShareErr = iota + 3400
	ERR_NETWORK_S2C_ACCEPT
	ERR_NETWORK_S2C_READ
	ERR_NETWORK_S2C_FLUSH
	ERR_NETWORK_S2C_WRITE
)

/***************************************  DB info error code, begin with 4000  ****************************************************/
const ( //sqlite error code
	ERR_DB_SQLITE_OPEN CrossShareErr = iota + 4000
	ERR_DB_SQLITE_QUERY
	ERR_DB_SQLITE_SCAN
	ERR_DB_SQLITE_ROWS
	ERR_DB_SQLITE_EXEC
	ERR_DB_SQLITE_LAST_INSERTID
)

/***************************************  business  info error code, begin with 5000  ****************************************************/
const ( // public business error code
	ERR_BIZ__OTHER CrossShareErr = iota + 5000
	ERR_BIZ_GET_STREAM_EMPTY
	ERR_BIZ_GET_CLIENT_INFO_EMPTY
	ERR_BIZ_UNKNOWN_FMTTYPE
	ERR_BIZ_JSON_MARSHAL
	ERR_BIZ_JSON_UNMARSHAL
	ERR_BIZ_JSON_EXTDATA_UNMARSHAL
	ERR_BIZ_GET_STREAM_RESET
)

// client to lan server business error code
const (
	ERR_BIZ_C2S_OTHER CrossShareErr = iota + 5100
	ERR_BIZ_C2S_GET_NO_SERVER_NAME
	ERR_BIZ_C2S_UNKNOWN_MSG_TYPE
	ERR_BIZ_C2S_READ_EMPTY_DATA
	ERR_BIZ_C2S_GET_EMPTY_CONNECT
)

// lan server to client business error code
const (
	ERR_BIZ_S2C_OTHER CrossShareErr = iota + 5200
	ERR_BIZ_S2C_UNKNOWN_MSG_TYPE
	ERR_BIZ_S2C_GET_EMPTY_CONNECT
	ERR_BIZ_S2C_READ_EMPTY_DATA
	ERR_BIZ_S2C_INVALID_INDEX
)

// peer to peer business error code
const (
	ERR_BIZ_P2P_OTHER CrossShareErr = iota + 5300
	ERR_BIZ_P2P_WRITE_EMPTY_DATA
	ERR_BIZ_P2P_GET_EMPTY_STREAM
)

// clipboard business error code
const (
	ERR_BIZ_CB_OTHER CrossShareErr = iota + 5400
	ERR_BIZ_CB_GET_STREAM_EMPTY
	ERR_BIZ_CB_GET_DATA_TYPE_ERR
	ERR_BIZ_CB_NO_DATA
	ERR_BIZ_CB_INVALID_DATA
	ERR_BIZ_CB_SRC_COPY_IMAGE
	ERR_BIZ_CB_SRC_COPY_IMAGE_TIMEOUT

	ERR_BIZ_CB_DST_COPY_IMAGE
	ERR_BIZ_CB_DST_COPY_IMAGE_TIMEOUT
	ERR_BIZ_CB_DST_COPY_IMAGE_LOSS
)

// file  business error code
const (
	ERR_BIZ_FD_OTHER CrossShareErr = iota + 5500
	ERR_BIZ_FD_DIR_NOT_EXISTS
	ERR_BIZ_FD_FILE_NOT_EXISTS
	ERR_BIZ_FD_FILE_SIZE_ERROR
	ERR_BIZ_FD_GET_STREAM_EMPTY
	ERR_BIZ_FD_UNKNOWN_CMD
	ERR_BIZ_FD_UNKNOWN_TYPE
	ERR_BIZ_FD_DATA_EMPTY
	ERR_BIZ_FD_DATA_INVALID
	ERR_BIZ_FD_SRC_OPEN_FILE
	ERR_BIZ_FD_SRC_FILE_SEEK
	ERR_BIZ_FD_SRC_COPY_FILE
	ERR_BIZ_FD_SRC_COPY_FILE_EMPTY
	ERR_BIZ_FD_SRC_COPY_FILE_TIMEOUT

	ERR_BIZ_FD_DST_OPEN_FILE
	ERR_BIZ_FD_DST_COPY_FILE
	ERR_BIZ_FD_DST_COPY_FILE_LOSS
	ERR_BIZ_FD_DST_COPY_FILE_TIMEOUT

	ERR_BIZ_DF_DATA_EMPTY
	ERR_BIZ_DF_INVALID_TIMESTAMP
	ERR_BIZ_DF_FILE_NOT_EXISTS
)

var errInfoMap = map[CrossShareErr]string{
	SUCCESS:                     "success!",
	ERR_DB_SQLITE_OPEN:          "open sqlite error!",
	ERR_DB_SQLITE_QUERY:         "query sqlite error!",
	ERR_DB_SQLITE_SCAN:          "scan sqlite error!",
	ERR_DB_SQLITE_EXEC:          "exec sqlite error!",
	ERR_DB_SQLITE_ROWS:          "get sqlite rows error!",
	ERR_DB_SQLITE_LAST_INSERTID: "get  sqlite last insert id error!",

	ERR_BIZ_JSON_MARSHAL:           "json marshal failed!",
	ERR_BIZ_JSON_UNMARSHAL:         "json unmarshal failed!",
	ERR_BIZ_JSON_EXTDATA_UNMARSHAL: "json ext data unmarshal failed!",
	ERR_BIZ_S2C_INVALID_INDEX:      "client index is invalid",
}
