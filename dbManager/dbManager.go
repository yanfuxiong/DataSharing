package dbManager

import (
	"context"
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"log"
	rtkMisc "rtk-cross-share/misc"
	"time"
)

var (
	g_DBConnectionStr     = "file:cross_share.db?_busy_timeout=5000&cache=shared&mode=rwc&_jounal_mode=WAL" // sqlite connect string
	g_KeppDBAliveInterval = 1                                                                               // minute
	g_SqlInstance         *sql.DB                                                                           // DB Instance

	// SQL
	sql_Query_ClientInfo   = "SELECT PkIndex,ClientId FROM t_client_info WHERE ClientId=?;"
	sql_Insert_ClientInfo  = "INSERT INTO t_client_info (ClientId,Host, IPAddr,DeviceName, Platform) VALUES (?, ?, ?, ?, ?);"
	sql_Update_ClientInfo  = "UPDATE t_client_info SET Online=true,IPAddr=? , UpdateTime = (datetime('now','localtime')) WHERE PkIndex=?;"
	sql_Update_SourcePort  = "UPDATE t_client_info SET SourceAndPort = ?, UpdateTime = (datetime('now','localtime')) where PkIndex = ? ;"
	sql_Update_HeartBeat   = "UPDATE t_client_info SET UpdateTime = (datetime('now','localtime')) where PkIndex = ? ;"
	sql_Online_Client      = "UPDATE t_client_info SET Online=true,UpdateTime = (datetime('now','localtime')) where PkIndex = ? ;"
	sql_Offline_Client     = "UPDATE t_client_info SET Online = false, UpdateTime = (datetime('now','localtime')) where PkIndex = ? ;"
	sql_Query_OnlineClient = "SELECT  client.ClientId, client.Host, client.IPAddr,client.DeviceName, client.Platform, client.UpdateTime FROM t_client_info client,t_auth_info auth where client.PkIndex=auth.ClientIndex and client.Online = true and auth.AuthStatus=true; "
	sql_UpSert_AuthStatus  = "INSERT INTO t_auth_info (ClientIndex) VALUES (?) ON CONFLICT (ClientIndex) DO UPDATE SET UpdateTime = (datetime('now','localtime')),AuthStatus=? ;"
)

// sqlite table struct
type ClientInfoTb struct {
	index         int
	clientId      string
	host          string
	ipAddr        string
	sourceAndPort int
	deviceName    string
	platform      string
	online        bool
	updateTime    string
	createTime    string
}
type AuthInfoTb struct {
	index       int
	clientIndex int
	authStatus  bool
	updateTime  string
}

func InitSqlite(ctx context.Context) {
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
		    SourceAndPort 	INTEGER,
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

func QueryOnlineClientList(onlineList *[]rtkMisc.ClientInfo) error {
	rows, err := g_SqlInstance.Query(sql_Query_OnlineClient)
	if err != nil {
		log.Printf("QueryOnlineClientList Query error[%+v]", err)
		return err
	}

	clientList := make([]ClientInfoTb, 0)
	for rows.Next() {
		var client ClientInfoTb
		if err = rows.Scan(&client.clientId, &client.host, &client.ipAddr, &client.deviceName, &client.platform, &client.updateTime); err != nil {
			log.Println(err)
			continue
		}
		clientList = append(clientList, client)
	}

	defer rows.Close()
	if err = rows.Err(); err != nil {
		log.Printf("QueryOnlineClientList rows err:%+v", err)
		return err
	}

	for _, client := range clientList {
		*onlineList = append(*onlineList, rtkMisc.ClientInfo{
			ID:         client.clientId,
			IpAddr:     client.ipAddr,
			Platform:   client.platform,
			DeviceName: client.deviceName,
		})
	}

	log.Printf("QueryOnlineClientList get len:%d", len(*onlineList))
	return nil
}

func UpSertClientInfo(reqMsg *rtkMisc.InitClientMessageReq) (int, error) {
	err := g_SqlInstance.Ping()
	if err != nil {
		log.Printf("sqlite3 Ping error:%+v,  so reconnect it!", err)
		if err = reOpenDBInstance(); err != nil {
			return 0, err
		}
	}

	rows, err := g_SqlInstance.Query(sql_Query_ClientInfo, reqMsg.ClientID)
	if err != nil {
		log.Printf("UpSertClientInfo Query error[%+v]", err)
		return 0, err
	}
	defer rows.Close()

	clientIndex := 0
	var clientID string
	for rows.Next() {
		if err = rows.Scan(&clientIndex, &clientID); err != nil {
			log.Println(err)
			return 0, err
		}
	}
	if err = rows.Err(); err != nil {
		log.Printf("UpSertClientInfo rows err:%+v", err)
		return 0, err
	}

	var nAffectCount int64
	if clientIndex == 0 {
		result, ExecErr := g_SqlInstance.Exec(sql_Insert_ClientInfo, reqMsg.ClientID, reqMsg.HOST, reqMsg.IPAddr, reqMsg.DeviceName, reqMsg.Platform)
		if ExecErr != nil {
			log.Printf("ID:[%s] InsertClientInfo Exec error:%+v", reqMsg.ClientID, ExecErr)
			return 0, ExecErr
		}
		pkIndex, LastErr := result.LastInsertId()
		if LastErr != nil {
			log.Printf("ID:[%s] InsertClientInfo LastInsertId error:%+v", reqMsg.ClientID, LastErr)
			return 0, err
		}
		clientIndex = int(pkIndex)
		nAffectCount, _ = result.RowsAffected()
	} else {
		result, ExecErr := g_SqlInstance.Exec(sql_Update_ClientInfo, reqMsg.IPAddr, clientIndex)
		if ExecErr != nil {
			log.Printf("ID:[%s] UpdateClientInfo Exec error:%+v", reqMsg.ClientID, ExecErr)
			return 0, ExecErr
		}
		nAffectCount, _ = result.RowsAffected()
	}

	log.Printf("ID:[%s] UpSertClientInfo success , get ClientIndex:[%d], RowsAffected:[%d]", reqMsg.ClientID, clientIndex, nAffectCount)
	return clientIndex, nil
}

func OfflineClient(pkIndex int) error {
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

func OnlineClient(pkIndex int) error {
	if pkIndex <= 0 {
		log.Printf("pkIndex:[%d] Err, OnlineClient skip!", pkIndex)
		return errors.New("Invalid Client Index")
	}

	result, ExecErr := g_SqlInstance.Exec(sql_Online_Client, pkIndex)
	if ExecErr != nil {
		log.Printf("pkIndex:[%d] OnlineClient Exec error:%+v", pkIndex, ExecErr)
		return ExecErr
	}
	nAffectCount, _ := result.RowsAffected()
	log.Printf("pkIndex:[%d] OnlineClient  success! RowsAffected:[%d]", pkIndex, nAffectCount)
	return nil
}

func UpdateSourceAndPort(pkIndex, sourcePort int) error {
	if pkIndex <= 0 {
		log.Printf("pkIndex:[%d] Err, UpdateSourceAndPort skip!", pkIndex)
		return errors.New("Invalid Client Index")
	}
	result, ExecErr := g_SqlInstance.Exec(sql_Update_SourcePort, sourcePort, pkIndex)
	if ExecErr != nil {
		log.Printf("pkIndex:[%d] UpdateSourceAndPort Exec error:%+v", pkIndex, ExecErr)
		return ExecErr
	}
	nAffectCount, _ := result.RowsAffected()
	log.Printf("pkIndex:[%d] UpdateSourceAndPort  success! RowsAffected:[%d]", pkIndex, nAffectCount)
	return nil
}

func UpdateHeartBeat(pkIndex int) error {
	if pkIndex <= 0 {
		log.Printf("pkIndex:[%d] Err, UpdateHeartBeat skip!", pkIndex)
		return errors.New("Invalid Client Index")
	}

	result, ExecErr := g_SqlInstance.Exec(sql_Update_HeartBeat, pkIndex)
	if ExecErr != nil {
		log.Printf("pkIndex:[%d] UpdateHeartBeat Exec error:%+v", pkIndex, ExecErr)
		return reOpenDBInstance()
	}
	nAffectCount, _ := result.RowsAffected()
	log.Printf("pkIndex:[%d] UpdateHeartBeat  success! RowsAffected:[%d]", pkIndex, nAffectCount)
	return nil
}

func UpdateAuthStatus(pkIndex int, status bool) error {
	if pkIndex <= 0 {
		log.Printf("pkIndex:[%d] Err, UpdateAuthStatus skip!", pkIndex)
		return errors.New("Invalid Client Index")
	}
	result, ExecErr := g_SqlInstance.Exec(sql_UpSert_AuthStatus, pkIndex, status)
	if ExecErr != nil {
		log.Printf("pkIndex:[%d] UpdateAuthStatus Exec error:%+v", pkIndex, ExecErr)
		return ExecErr
	}
	nAffectCount, _ := result.RowsAffected()
	log.Printf("pkIndex:[%d] UpdateAuthStatus [%+v] success! RowsAffected:[%d]", pkIndex, status, nAffectCount)
	return nil
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

	log.Printf("[%s] sqlite3 is alive!", rtkMisc.GetFuncInfo())
}

func reOpenDBInstance() error {
	g_SqlInstance.Close()
	db, err := sql.Open("sqlite3", g_DBConnectionStr)
	if err != nil {
		log.Printf("reOpen sqlite3 [%s] err:%+v", g_DBConnectionStr, err)
		return err
	}
	g_SqlInstance = db
	g_SqlInstance.SetConnMaxIdleTime(time.Duration(0)) //connections are not closed due to a connection's idle time
	g_SqlInstance.SetMaxIdleConns(10)

	log.Println("reOpenDBInstance success!")
	return nil
}
