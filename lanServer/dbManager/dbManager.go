package dbManager

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os"
	rtkCommon "rtk-cross-share/lanServer/common"
	rtkGlobal "rtk-cross-share/lanServer/global"
	rtkMisc "rtk-cross-share/misc"
	"time"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

const (
	g_KeppDBAliveInterval = 1 // minute
	g_ReconnListInterval  = 2 // seconds
)

var (
	g_SqlInstance *sql.DB //	sqlite Instance
)

const (
	g_DBConnectionStr = "file:" + rtkGlobal.DB_PATH + rtkGlobal.DB_NAME + "?_busy_timeout=5000&cache=shared&mode=rwc&_jounal_mode=WAL" //sqlite connect string

	// SQL
	sql_Query_ClientInfo      = "SELECT PkIndex,ClientId FROM t_client_info WHERE ClientId=?;"
	sql_Insert_ClientInfo     = "INSERT INTO t_client_info (ClientId,Host, IPAddr,DeviceName, Platform) VALUES (?, ?, ?, ?, ?);"
	sql_Update_ClientInfo     = "UPDATE t_client_info SET Online=true,IPAddr=? , UpdateTime = (datetime('now','localtime')) WHERE PkIndex=?;"
	sql_Update_SourcePort     = "UPDATE t_client_info SET Source = ?, Port = ?, UpdateTime = (datetime('now','localtime')) where PkIndex = ? ;"
	sql_Update_HeartBeat      = "UPDATE t_client_info SET UpdateTime = (datetime('now','localtime')) where PkIndex = ? ;"
	sql_Online_Client         = "UPDATE t_client_info SET Online=true,UpdateTime = (datetime('now','localtime')) where PkIndex = ? ;"
	sql_Offline_Client        = "UPDATE t_client_info SET Online = false, UpdateTime = (datetime('now','localtime')) where PkIndex = ? ;"
	sql_Offline_All           = "UPDATE t_client_info SET Online = false, UpdateTime = (datetime('now','localtime')) where Online = true;"
	sql_Query_OnlineClient    = "SELECT  client.ClientId, client.Host, client.IPAddr,client.DeviceName, client.Platform,client.Source,client.Port, client.UpdateTime FROM t_client_info client,t_auth_info auth where client.PkIndex=auth.ClientIndex and client.Online = true and auth.AuthStatus=true; "
	sql_UpSert_AuthStatus     = "INSERT INTO t_auth_info (ClientIndex) VALUES (?) ON CONFLICT (ClientIndex) DO UPDATE SET UpdateTime = (datetime('now','localtime')),AuthStatus=? ;"
	sql_Query_DeviceName      = "SELECT DeviceName FROM t_client_info WHERE PkIndex=?;"
	sql_Query_ClientBySrcPort = "SELECT PkIndex, ClientId FROM t_client_info WHERE Source=? AND Port=? AND Online=1 ORDER BY UpdateTime DESC LIMIT 1"
	sql_Query_ReconnList      = "SELECT c.ClientId, c.IPAddr, c.Platform, c.DeviceName FROM t_client_info c, t_auth_info a WHERE c.Online=1 AND c.PkIndex=a.ClientIndex AND a.AuthStatus=true AND (strftime('%s', 'now') - strftime('%s', c.UpdateTime)) > ?;"
	sql_Query_SrcPort         = "SELECT Source,Port FROM t_client_info WHERE PkIndex=?;"
)

// sqlite table struct
type ClientInfoTb struct {
	index      int
	clientId   string
	host       string
	ipAddr     string
	source     int
	port       int
	deviceName string
	platform   string
	online     bool
	updateTime string
	createTime string
}
type AuthInfoTb struct {
	index       int
	clientIndex int
	authStatus  bool
	updateTime  string
}

func InitSqlite(ctx context.Context) {
	err := rtkMisc.CreateDir(rtkGlobal.DB_PATH, os.ModePerm)
	if err != nil {
		log.Printf("Create DB failed: %s", err.Error())
		log.Fatal(errors.New("buildDbPath error!"))
	}

	db, err := sql.Open("sqlite3", g_DBConnectionStr)
	if err != nil {
		log.Fatal(err)
	}

	createTableSql := `
		CREATE TABLE IF NOT EXISTS t_client_info (
		    PkIndex			INTEGER PRIMARY KEY AUTOINCREMENT,
		    ClientId		TEXT UNIQUE,
		    Host 			TEXT,
		    IPAddr   		TEXT NOT NULL ,
		    Source		 	INTEGER,
			Port			INTEGER,
		    Online 			BOOLEAN NOT NULL DEFAULT TRUE,
		    DeviceName 		TEXT,
		    Platform  		TEXT,
		    UpdateTime  	DATETIME NOT NULL DEFAULT (datetime('now','localtime')),
		    CreateTime 		DATETIME NOT NULL DEFAULT (datetime('now','localtime'))
        );
		CREATE TABLE IF NOT EXISTS t_auth_info (
		    PkIndex 		INTEGER PRIMARY KEY AUTOINCREMENT,
		    ClientIndex		INTEGER UNIQUE,
		    AuthStatus 		BOOLEAN NOT NULL DEFAULT TRUE,
		    UpdateTime  	DATETIME NOT NULL DEFAULT (datetime('now','localtime')),
		    CreateTime 		DATETIME NOT NULL DEFAULT (datetime('now','localtime'))
        );
	`
	g_SqlInstance = db
	g_SqlInstance.SetConnMaxIdleTime(time.Duration(0)) //connections are not closed due to a connection's idle time
	g_SqlInstance.SetMaxIdleConns(10)

	_, err = g_SqlInstance.Exec(createTableSql)
	if err != nil {
		log.Fatal(err)
	}

	OfflineClientList()

	rtkMisc.GoSafe(func() {
		tkKeepAlive := time.NewTicker(time.Duration(g_KeppDBAliveInterval) * time.Minute)
		defer tkKeepAlive.Stop()

		for {
			select {
			case <-ctx.Done():
				g_SqlInstance.Close()
				return
			case <-tkKeepAlive.C:
				keppAlive()
			}
		}
	})

}

func QueryOnlineClientList(onlineList *[]rtkMisc.ClientInfo) rtkMisc.CrossShareErr {
	rows, err := g_SqlInstance.Query(sql_Query_OnlineClient)
	if err != nil {
		log.Printf("QueryOnlineClientList Query error[%+v]", err)
		return rtkMisc.ERR_DB_SQLITE_QUERY
	}

	clientList := make([]ClientInfoTb, 0)
	for rows.Next() {
		var client ClientInfoTb
		if err = rows.Scan(&client.clientId, &client.host, &client.ipAddr, &client.deviceName, &client.platform, &client.source, &client.port, &client.updateTime); err != nil {
			log.Println(err)
			continue
		}
		clientList = append(clientList, client)
	}

	defer rows.Close()
	if err = rows.Err(); err != nil {
		log.Printf("QueryOnlineClientList rows err:%+v", err)
		return rtkMisc.ERR_DB_SQLITE_EXEC
	}

	for _, client := range clientList {
		*onlineList = append(*onlineList, rtkMisc.ClientInfo{
			ID:             client.clientId,
			IpAddr:         client.ipAddr,
			Platform:       client.platform,
			DeviceName:     client.deviceName,
			SourcePortType: rtkCommon.GetClientSourcePortType(client.source, client.port),
		})
	}

	log.Printf("QueryOnlineClientList get len:%d", len(*onlineList))
	return rtkMisc.SUCCESS
}

func UpSertClientInfo(reqMsg *rtkMisc.InitClientMessageReq) (uint32, rtkMisc.CrossShareErr) {
	err := g_SqlInstance.Ping()
	if err != nil {
		log.Printf("sqlite3 Ping error:%+v,  so reconnect it!", err)
		if errCode := reOpenDBInstance(); errCode != rtkMisc.SUCCESS {
			return 0, errCode
		}
	}

	rows, err := g_SqlInstance.Query(sql_Query_ClientInfo, reqMsg.ClientID)
	if err != nil {
		log.Printf("UpSertClientInfo Query error[%+v]", err)
		return 0, rtkMisc.ERR_DB_SQLITE_QUERY
	}
	defer rows.Close()

	clientIndex := uint32(0)
	var clientID string
	for rows.Next() {
		if err = rows.Scan(&clientIndex, &clientID); err != nil {
			log.Println(err)
			return 0, rtkMisc.ERR_DB_SQLITE_SCAN
		}
	}
	if err = rows.Err(); err != nil {
		log.Printf("UpSertClientInfo rows err:%+v", err)
		return 0, rtkMisc.ERR_DB_SQLITE_ROWS
	}

	var nAffectCount int64
	if clientIndex == 0 {
		result, ExecErr := g_SqlInstance.Exec(sql_Insert_ClientInfo, reqMsg.ClientID, reqMsg.HOST, reqMsg.IPAddr, reqMsg.DeviceName, reqMsg.Platform)
		if ExecErr != nil {
			log.Printf("ID:[%s] InsertClientInfo Exec error:%+v", reqMsg.ClientID, ExecErr)
			return 0, rtkMisc.ERR_DB_SQLITE_EXEC
		}
		pkIndex, LastErr := result.LastInsertId()
		if LastErr != nil {
			log.Printf("ID:[%s] InsertClientInfo LastInsertId error:%+v", reqMsg.ClientID, LastErr)
			return 0, rtkMisc.ERR_DB_SQLITE_LAST_INSERTID
		}
		clientIndex = uint32(pkIndex)
		nAffectCount, _ = result.RowsAffected()
	} else {
		result, ExecErr := g_SqlInstance.Exec(sql_Update_ClientInfo, reqMsg.IPAddr, clientIndex)
		if ExecErr != nil {
			log.Printf("ID:[%s] UpdateClientInfo Exec error:%+v", reqMsg.ClientID, ExecErr)
			return 0, rtkMisc.ERR_DB_SQLITE_EXEC
		}
		nAffectCount, _ = result.RowsAffected()
	}

	log.Printf("ID:[%s] UpSertClientInfo success , get ClientIndex:[%d], RowsAffected:[%d]", reqMsg.ClientID, clientIndex, nAffectCount)
	return clientIndex, rtkMisc.SUCCESS
}

func OfflineClient(pkIndex uint32) error {
	if pkIndex <= 0 {
		log.Printf("pkIndex:[%d] Err, OfflineClient skip!", pkIndex)
		return errors.New("Invalid Client Index")
	}
	result, ExecErr := g_SqlInstance.Exec(sql_Offline_Client, pkIndex)
	if ExecErr != nil {
		log.Printf("pkIndex:[%d] OfflineClient Exec error:%+v", pkIndex, ExecErr)
		return ExecErr
	}
	nAffectCount, _ := result.RowsAffected()
	log.Printf("pkIndex:[%d] OfflineClient  success! RowsAffected:[%d]", pkIndex, nAffectCount)
	return nil
}

func OnlineClient(pkIndex uint32) rtkMisc.CrossShareErr {
	if pkIndex <= 0 {
		log.Printf("pkIndex:[%d] Err, OnlineClient skip!", pkIndex)
		return rtkMisc.ERR_BIZ_S2C_INVALID_INDEX
	}

	result, ExecErr := g_SqlInstance.Exec(sql_Online_Client, pkIndex)
	if ExecErr != nil {
		log.Printf("pkIndex:[%d] OnlineClient Exec error:%+v", pkIndex, ExecErr)
		return rtkMisc.ERR_DB_SQLITE_EXEC
	}
	nAffectCount, _ := result.RowsAffected()
	log.Printf("pkIndex:[%d] OnlineClient  success! RowsAffected:[%d]", pkIndex, nAffectCount)
	return rtkMisc.SUCCESS
}

func OfflineClientList() error {
	result, ExecErr := g_SqlInstance.Exec(sql_Offline_All)
	if ExecErr != nil {
		log.Printf("OfflineClientList Exec error:%+v", ExecErr)
		return ExecErr
	}
	nAffectCount, _ := result.RowsAffected()
	log.Printf("Offline All Client List success! RowsAffected:[%d]", nAffectCount)
	return nil
}

func UpdateSourceAndPort(pkIndex, source int, port int) rtkMisc.CrossShareErr {
	if pkIndex <= 0 {
		log.Printf("pkIndex:[%d] Err, UpdateSourceAndPort skip!", pkIndex)
		return rtkMisc.ERR_BIZ_S2C_INVALID_INDEX
	}
	result, ExecErr := g_SqlInstance.Exec(sql_Update_SourcePort, source, port, pkIndex)
	if ExecErr != nil {
		log.Printf("pkIndex:[%d] UpdateSourceAndPort Exec error:%+v", pkIndex, ExecErr)
		return rtkMisc.ERR_DB_SQLITE_EXEC
	}
	nAffectCount, _ := result.RowsAffected()
	log.Printf("pkIndex:[%d] UpdateSourceAndPort  success! RowsAffected:[%d]", pkIndex, nAffectCount)
	return rtkMisc.SUCCESS
}

func UpdateHeartBeat(pkIndex int) rtkMisc.CrossShareErr {
	if pkIndex <= 0 {
		log.Printf("pkIndex:[%d] Err, UpdateHeartBeat skip!", pkIndex)
		return rtkMisc.ERR_BIZ_S2C_INVALID_INDEX
	}

	result, ExecErr := g_SqlInstance.Exec(sql_Update_HeartBeat, pkIndex)
	if ExecErr != nil {
		log.Printf("pkIndex:[%d] UpdateHeartBeat Exec error:%+v", pkIndex, ExecErr)
		return reOpenDBInstance()
	}
	nAffectCount, _ := result.RowsAffected()
	log.Printf("pkIndex:[%d] UpdateHeartBeat  success! RowsAffected:[%d]", pkIndex, nAffectCount)
	return rtkMisc.SUCCESS
}

func UpdateAuthStatus(pkIndex int, status bool) rtkMisc.CrossShareErr {
	if pkIndex <= 0 {
		log.Printf("pkIndex:[%d] Err, UpdateAuthStatus skip!", pkIndex)
		return rtkMisc.ERR_BIZ_S2C_INVALID_INDEX
	}
	result, ExecErr := g_SqlInstance.Exec(sql_UpSert_AuthStatus, pkIndex, status)
	if ExecErr != nil {
		log.Printf("pkIndex:[%d] UpdateAuthStatus Exec error:%+v", pkIndex, ExecErr)
		return rtkMisc.ERR_DB_SQLITE_EXEC
	}
	nAffectCount, _ := result.RowsAffected()
	log.Printf("pkIndex:[%d] UpdateAuthStatus [%+v] success! RowsAffected:[%d]", pkIndex, status, nAffectCount)
	return rtkMisc.SUCCESS
}

func UpdateAuthAndSrcPort(pkIndex int, status bool, source int, port int) rtkMisc.CrossShareErr {
	errAuthStatus := UpdateAuthStatus(pkIndex, status)
	if errAuthStatus != rtkMisc.SUCCESS {
		return errAuthStatus
	}

	errSrcPort := UpdateSourceAndPort(pkIndex, source, port)
	if errSrcPort != rtkMisc.SUCCESS {
		return errSrcPort
	}

	return rtkMisc.SUCCESS
}

func QueryDeviceName(pkIndex int) (string, error) {
	if pkIndex <= 0 {
		log.Printf("pkIndex:[%d] Err, QueryDeviceName failed!", pkIndex)
		return "", errors.New("Invalid Client Index")
	}

	rows, err := g_SqlInstance.Query(sql_Query_DeviceName, pkIndex)
	if err != nil {
		log.Printf("QueryDeviceName Query error[%+v]", err)
		return "", err
	}
	defer rows.Close()

	var deviceName string = ""
	var errQuery error
	for rows.Next() {
		if errQuery = rows.Scan(&deviceName); errQuery != nil {
			return "", errQuery
		}
	}
	if err = rows.Err(); err != nil {
		log.Printf("QueryDeviceName rows err:%+v", err)
		return "", err
	}

	return deviceName, nil
}

func QuerySrcPort(pkIndex uint32, src, port *int) rtkMisc.CrossShareErr {
	if pkIndex <= 0 {
		log.Printf("pkIndex:[%d] Err, QuerySrcPort failed!", pkIndex)
		return rtkMisc.ERR_BIZ_S2C_INVALID_INDEX
	}

	rows, err := g_SqlInstance.Query(sql_Query_SrcPort, pkIndex)
	if err != nil {
		log.Printf("QuerySrcPort by pkIndex:[%d] Query error[%+v]", pkIndex, err)
		return rtkMisc.ERR_DB_SQLITE_QUERY
	}
	defer rows.Close()

	var errQuery error
	for rows.Next() {
		if errQuery = rows.Scan(src, port); errQuery != nil {
			return rtkMisc.ERR_DB_SQLITE_SCAN
		}
	}
	if err = rows.Err(); err != nil {
		log.Printf("QuerySrcPort rows err:%+v", err)
		return rtkMisc.ERR_DB_SQLITE_ROWS
	}
	log.Printf("[%s] get src:%d  port:%d by pkIndex:%d", rtkMisc.GetFuncInfo(), *src, *port, pkIndex)
	return rtkMisc.SUCCESS
}

func QueryClientBySrcPort(source, port int) (int, string, error) {
	if source < 0 || port < 0 {
		log.Printf("[%s] Invalid (source,port)=(%d,%d)", rtkMisc.GetFuncInfo(), source, port)
		return 0, "", errors.New("Invalid source or port")
	}

	rows, err := g_SqlInstance.Query(sql_Query_ClientBySrcPort, source, port)
	if err != nil {
		log.Printf("[%s] Query error[%+v]", rtkMisc.GetFuncInfo(), err)
		return 0, "", err
	}
	defer rows.Close()

	var clientIdx int = 0
	var clientId string = ""
	var errQuery error
	for rows.Next() {
		if errQuery = rows.Scan(&clientIdx, &clientId); errQuery != nil {
			return 0, "", errQuery
		}
	}
	if err = rows.Err(); err != nil {
		log.Printf("[%s] rows err:%+v", rtkMisc.GetFuncInfo(), err)
		return 0, "", err
	}

	if clientId == "" {
		log.Printf("[%s] Empty client ID", rtkMisc.GetFuncInfo())
		return 0, "", errors.New("Empty clientID")
	}
	return clientIdx, clientId, nil
}

func QueryReconnList(reconnList *[]rtkMisc.ClientInfo) rtkMisc.CrossShareErr {
	rows, err := g_SqlInstance.Query(sql_Query_ReconnList, g_ReconnListInterval)
	if err != nil {
		log.Printf("QueryReconnList Query error[%+v]", err)
		return rtkMisc.ERR_DB_SQLITE_QUERY
	}

	clientList := make([]ClientInfoTb, 0)
	for rows.Next() {
		var client ClientInfoTb
		if err = rows.Scan(&client.clientId, &client.ipAddr, &client.platform, &client.deviceName); err != nil {
			log.Println(err)
			continue
		}
		clientList = append(clientList, client)
	}

	defer rows.Close()
	if err = rows.Err(); err != nil {
		log.Printf("QueryReconnList rows err:%+v", err)
		return rtkMisc.ERR_DB_SQLITE_EXEC
	}

	for _, client := range clientList {
		*reconnList = append(*reconnList, rtkMisc.ClientInfo{
			ID:         client.clientId,
			IpAddr:     client.ipAddr,
			Platform:   client.platform,
			DeviceName: client.deviceName,
		})
	}

	// DEBUG log
	// log.Printf("QueryReconnList get len:%d", len(*reconnList))
	return rtkMisc.SUCCESS
}

func keppAlive() {
	err := g_SqlInstance.Ping()
	if err != nil {
		g_SqlInstance.Close()
		log.Printf("sqlite3 Ping error:%+v,  so reconnect it!", err)
		db, err := sql.Open("sqlite3", g_DBConnectionStr)
		if err != nil {
			log.Printf("Open sqlite3 [%s] err:%+v", g_DBConnectionStr, err)
			return
		}
		g_SqlInstance = db
		g_SqlInstance.SetConnMaxIdleTime(time.Duration(0)) //connections are not closed due to a connection's idle time
		g_SqlInstance.SetMaxIdleConns(10)
	}

	// DEBUG log
	// log.Printf("[%s] sqlite3 is alive!", rtkMisc.GetFuncInfo())
}

func reOpenDBInstance() rtkMisc.CrossShareErr {
	g_SqlInstance.Close()
	db, err := sql.Open("sqlite3", g_DBConnectionStr)
	if err != nil {
		log.Printf("reOpen sqlite3 [%s] err:%+v", g_DBConnectionStr, err)
		return rtkMisc.ERR_DB_SQLITE_OPEN
	}
	g_SqlInstance = db
	g_SqlInstance.SetConnMaxIdleTime(time.Duration(0)) //connections are not closed due to a connection's idle time
	g_SqlInstance.SetMaxIdleConns(10)

	log.Println("reOpenDBInstance success!")
	return rtkMisc.SUCCESS
}
