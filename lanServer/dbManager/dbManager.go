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
)

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

	g_SqlInstance = db
	g_SqlInstance.SetConnMaxIdleTime(time.Duration(0)) //connections are not closed due to a connection's idle time
	g_SqlInstance.SetMaxIdleConns(10)

	_, err = g_SqlInstance.Exec(SqlDataCreateTable.toString())
	if err != nil {
		log.Fatal(err)
	}

	UpdateAllClientOffline()

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

func upsertClientInfo(pkIndex *int, clientId, host, ipAddr, deviceName, platform string) rtkMisc.CrossShareErr {
	if pkIndex == nil {
		log.Printf("[%s] pkIndex is null", rtkMisc.GetFuncInfo())
		return rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	tx, err := g_SqlInstance.Begin()
	if err != nil {
		log.Printf("[%s] Err: %s", rtkMisc.GetFuncInfo(), err.Error())
		return rtkMisc.ERR_DB_SQLITE_BEGIN
	}

	sqlData := SqlDataQueryeClientMaxIndex
	row := tx.QueryRow(sqlData.toString())
	if err != nil {
		log.Printf("[%s] Err: %s", rtkMisc.GetFuncInfo(), err.Error())
		return rtkMisc.ERR_DB_SQLITE_EXEC
	}

	maxIndex := int(0)
	if err = row.Scan(&maxIndex); err != nil && err != sql.ErrNoRows {
		log.Printf("[%s] Err: %s", rtkMisc.GetFuncInfo(), err.Error())
		return rtkMisc.ERR_DB_SQLITE_SCAN
	}

	if maxIndex < 255 {
		sqlData = SqlDataUpsertClientInfo
		param := []any{clientId, host, ipAddr, deviceName, platform}
		if sqlData.checkArgsCount(param) == false {
			tx.Rollback()
			return rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
		}

		row = tx.QueryRow(sqlData.toString(), param...)
		if err = row.Scan(pkIndex); err != nil {
			log.Printf("[%s] Err: %s", rtkMisc.GetFuncInfo(), err.Error())
			tx.Rollback()
			return rtkMisc.ERR_DB_SQLITE_SCAN
		}
	} else {
		sqlData = SqlDataQueryEarliestClient
		var OldIndex int
		var OldUpdate string
		rowIndex := tx.QueryRow(sqlData.toString())
		if err = rowIndex.Scan(&OldIndex, &OldUpdate); err != nil {
			tx.Rollback()
			log.Printf("[%s] Err: %s", rtkMisc.GetFuncInfo(), err.Error())
			return rtkMisc.ERR_DB_SQLITE_SCAN
		}

		log.Printf("[%s] get index:%d, is over range, and earliest client PkIndex:%d updatetime:%s, bein to replace it", rtkMisc.GetFuncInfo(), maxIndex+1, OldIndex, OldUpdate)

		args := []any{clientId, ipAddr, deviceName, platform, OldIndex}
		sqlData = SqlDataUpdateClientInfo.withCond_SET(SqlCondClientId, SqlCondIPAddr, SqlCondDeviceName, SqlCondPlatform, SqlCondOnline).withCond_WHERE(SqlCondPkIndex)
		if !sqlData.checkArgsCount(args) {
			tx.Rollback()
			return rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
		}
		rowIndex = tx.QueryRow(sqlData.toString(), args...)
		if err = rowIndex.Scan(pkIndex); err != nil {
			tx.Rollback()
			log.Printf("[%s] Err: %s", rtkMisc.GetFuncInfo(), err.Error())
			return rtkMisc.ERR_DB_SQLITE_SCAN
		}

		sqlData = SqlDataDeleteAuthInfo.withCond_WHERE(SqlCondClientIndex)
		_, err = tx.Exec(sqlData.toString(), OldIndex)
		if err != nil {
			tx.Rollback()
			log.Printf("[%s] Err: %s", rtkMisc.GetFuncInfo(), err.Error())
			return rtkMisc.ERR_DB_SQLITE_EXEC
		}

		log.Printf("[%s] the earliest client info PkIndex:%d updatetime:%s, is replace by clientId:[%s] success!", rtkMisc.GetFuncInfo(), OldIndex, OldUpdate, clientId)
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("[%s] Err: %s", rtkMisc.GetFuncInfo(), err.Error())
		return rtkMisc.ERR_DB_SQLITE_COMMIT
	}

	return rtkMisc.SUCCESS
}

func upsertAuthStatus(authPkIndex *int, clientPkIndex int, authStatus bool) rtkMisc.CrossShareErr {
	if authPkIndex == nil {
		log.Printf("[%s] pkIndex is null", rtkMisc.GetFuncInfo())
		return rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	sqlData := SqlDataUpsertAuthInfo
	param := []any{clientPkIndex, authStatus}
	if sqlData.checkArgsCount(param) == false {
		return rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	row := g_SqlInstance.QueryRow(sqlData.toString(), param...)
	if err := row.Scan(authPkIndex); err != nil {
		log.Printf("[%s] Err: %s", rtkMisc.GetFuncInfo(), err.Error())
		return rtkMisc.ERR_DB_SQLITE_SCAN
	}

	return rtkMisc.SUCCESS
}

func updateClientInfo(pkIndexList *[]int, setConds []SqlCond, whereConds []SqlCond, args ...any) rtkMisc.CrossShareErr {
	if pkIndexList == nil {
		log.Printf("[%s] pkIndexList is null", rtkMisc.GetFuncInfo())
		return rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	sqlData := SqlDataUpdateClientInfo.withCond_SET(setConds...).withCond_WHERE(whereConds...)
	if sqlData.checkArgsCount(args) == false {
		return rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	rows, err := g_SqlInstance.Query(sqlData.toString(), args...)
	if err != nil {
		log.Printf("[%s] Query error[%+v]", rtkMisc.GetFuncInfo(), err)
		log.Printf("[%s] Err: %s", rtkMisc.GetFuncInfo(), sqlData.toString())
		return rtkMisc.ERR_DB_SQLITE_QUERY
	}

	*pkIndexList = (*pkIndexList)[:0]
	for rows.Next() {
		var pkIndex int
		if err = rows.Scan(&pkIndex); err != nil {
			log.Printf("[%s] Err: %s", rtkMisc.GetFuncInfo(), err.Error())
			continue
		}
		*pkIndexList = append(*pkIndexList, pkIndex)
	}

	defer rows.Close()
	if err = rows.Err(); err != nil {
		log.Printf("[%s] rows err:%+v", rtkMisc.GetFuncInfo(), err)
		return rtkMisc.ERR_DB_SQLITE_EXEC
	}

	return rtkMisc.SUCCESS
}

func queryClientInfo(clientInfoList *[]rtkCommon.ClientInfoTb, conds []SqlCond, args ...any) rtkMisc.CrossShareErr {
	if clientInfoList == nil {
		log.Printf("[%s] ClientInfoList is null", rtkMisc.GetFuncInfo())
		return rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	sqlData := SqlDataQueryClientInfo.withCond_WHERE(conds...)
	if sqlData.checkArgsCount(args) == false {
		return rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	rows, err := g_SqlInstance.Query(sqlData.toString(), args...)
	if err != nil {
		log.Printf("[%s] Query error[%+v]", rtkMisc.GetFuncInfo(), err)
		log.Printf("[%s] Err: %s", rtkMisc.GetFuncInfo(), sqlData.toString())
		return rtkMisc.ERR_DB_SQLITE_QUERY
	}

	*clientInfoList = (*clientInfoList)[:0]
	for rows.Next() {
		var client rtkCommon.ClientInfoTb
		if err = rows.Scan(&client.Index, &client.ClientId, &client.Host, &client.IpAddr,
			&client.Source, &client.Port, &client.DeviceName, &client.Platform,
			&client.Online, &client.AuthStatus, &client.UpdateTime, &client.CreateTime); err != nil {
			log.Printf("[%s] Err: %s", rtkMisc.GetFuncInfo(), err.Error())
			continue
		}
		*clientInfoList = append(*clientInfoList, client)
	}

	defer rows.Close()
	if err = rows.Err(); err != nil {
		log.Printf("[%s] rows err:%+v", rtkMisc.GetFuncInfo(), err)
		return rtkMisc.ERR_DB_SQLITE_EXEC
	}

	return rtkMisc.SUCCESS
}

// =============================
// ClientInfo update event
// =============================
type NotifyUpdateClientInfoCallback func(rtkCommon.ClientInfoTb)

var notifyUpdateClientInfoCallback NotifyUpdateClientInfoCallback

func SetNotifyUpdateClientInfoCallback(cb NotifyUpdateClientInfoCallback) {
	notifyUpdateClientInfoCallback = cb
}

func notifyUpdateClientInfo(pkIndexList []int) {
	for _, pkIndex := range pkIndexList {
		clientInfo, err := QueryClientInfoByIndex(pkIndex)
		if err != rtkMisc.SUCCESS {
			continue
		}

		notifyUpdateClientInfoCallback(clientInfo)
	}
}

// =============================
// Query
// =============================
func QueryOnlineClientList(clientInfoList *[]rtkCommon.ClientInfoTb) rtkMisc.CrossShareErr {
	err := queryClientInfo(
		clientInfoList,
		[]SqlCond{SqlCondOnline, SqlCondAuthStatusIsTrue},
	)
	if err != rtkMisc.SUCCESS {
		return err
	}

	log.Printf("QueryOnlineClientList get len:%d", len(*clientInfoList))
	return rtkMisc.SUCCESS
}

func QueryClientInfoByIndex(pkIndex int) (rtkCommon.ClientInfoTb, rtkMisc.CrossShareErr) {
	if pkIndex <= 0 {
		log.Printf("pkIndex:[%d] Err, QueryDeviceName failed! Invalid Client Index", pkIndex)
		return rtkCommon.ClientInfoTb{}, rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	clientInfoList := make([]rtkCommon.ClientInfoTb, 0)
	err := queryClientInfo(&clientInfoList,
		[]SqlCond{SqlCondPkIndex},
		pkIndex,
	)
	if err != rtkMisc.SUCCESS {
		return rtkCommon.ClientInfoTb{}, err
	}

	if len(clientInfoList) == 0 {
		return rtkCommon.ClientInfoTb{}, rtkMisc.ERR_DB_SQLITE_EMPTY_RESULT
	}

	if len(clientInfoList) > 1 {
		log.Printf("[%s] WARNING! The result count from database more than 1", rtkMisc.GetFuncInfo())
	}

	return clientInfoList[0], rtkMisc.SUCCESS
}

func QueryClientInfoBySrcPort(source, port int) (rtkCommon.ClientInfoTb, rtkMisc.CrossShareErr) {
	if source < 0 || port < 0 {
		log.Printf("[%s] Invalid (source,port)=(%d,%d)", rtkMisc.GetFuncInfo(), source, port)
		return rtkCommon.ClientInfoTb{}, rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	clientInfoList := make([]rtkCommon.ClientInfoTb, 0)
	err := queryClientInfo(
		&clientInfoList,
		[]SqlCond{SqlCondSource, SqlCondPort},
		source, port,
	)
	if err != rtkMisc.SUCCESS {
		return rtkCommon.ClientInfoTb{}, err
	}

	if len(clientInfoList) == 0 {
		return rtkCommon.ClientInfoTb{Source: source, Port: port}, rtkMisc.ERR_DB_SQLITE_EMPTY_RESULT
	}

	if len(clientInfoList) > 1 {
		log.Printf("[%s] WARNING! The result count from database more than 1", rtkMisc.GetFuncInfo())
	}

	clientInfoList[0].Source = source
	clientInfoList[0].Port = port
	return clientInfoList[0], rtkMisc.SUCCESS
}

func QueryReconnList(clientInfoList *[]rtkCommon.ClientInfoTb) rtkMisc.CrossShareErr {
	err := queryClientInfo(
		clientInfoList,
		[]SqlCond{SqlCondOnline, SqlCondAuthStatusIsTrue, SqlCondLastUpdateTime},
		g_ReconnListInterval,
	)
	if err != rtkMisc.SUCCESS {
		return err
	}

	return rtkMisc.SUCCESS
}

// =============================
// Update/Insert
// =============================
func UpsertClientInfo(clientId, host, ipAddr, deviceName, platform string) (int, rtkMisc.CrossShareErr) {
	pkIndex := 0
	err := upsertClientInfo(&pkIndex, clientId, host, ipAddr, deviceName, platform)
	if err != rtkMisc.SUCCESS {
		return 0, err
	}

	notifyUpdateClientInfo([]int{pkIndex})
	return pkIndex, err
}

func UpdateClientOffline(pkIndex int) rtkMisc.CrossShareErr {
	if pkIndex <= 0 {
		log.Printf("pkIndex:[%d] Err, UpdateClientOffline skip!", pkIndex)
		return rtkMisc.ERR_BIZ_S2C_INVALID_INDEX
	}

	pkIndexList := make([]int, 0)
	err := updateClientInfo(
		&pkIndexList,
		[]SqlCond{SqlCondOffline},
		[]SqlCond{SqlCondPkIndex},
		pkIndex,
	)
	if err != rtkMisc.SUCCESS {
		return err
	}

	authPkIndex := int(0)
	errUpsertAuthStatus := upsertAuthStatus(&authPkIndex, pkIndex, false)
	if errUpsertAuthStatus != rtkMisc.SUCCESS {
		return errUpsertAuthStatus
	}

	notifyUpdateClientInfo(pkIndexList)
	return rtkMisc.SUCCESS
}

func UpdateClientOnline(pkIndex int) rtkMisc.CrossShareErr {
	if pkIndex <= 0 {
		log.Printf("pkIndex:[%d] Err, UpdateClientOnline skip!", pkIndex)
		return rtkMisc.ERR_BIZ_S2C_INVALID_INDEX
	}

	pkIndexList := make([]int, 0)
	err := updateClientInfo(
		&pkIndexList,
		[]SqlCond{SqlCondOnline},
		[]SqlCond{SqlCondPkIndex},
		pkIndex,
	)
	if err != rtkMisc.SUCCESS {
		return err
	}

	notifyUpdateClientInfo(pkIndexList)
	return rtkMisc.SUCCESS
}

func UpdateAllClientOffline() rtkMisc.CrossShareErr {
	pkIndexList := make([]int, 0)
	err := updateClientInfo(
		&pkIndexList,
		[]SqlCond{SqlCondOffline},
		[]SqlCond{SqlCondOnline},
	)
	if err != rtkMisc.SUCCESS {
		return err
	}

	notifyUpdateClientInfo(pkIndexList)
	return rtkMisc.SUCCESS
}

func UpdateAuthAndSrcPort(pkIndex int, status bool, source, port int) rtkMisc.CrossShareErr {
	if pkIndex <= 0 {
		log.Printf("pkIndex:[%d] Err, UpdateClientSourceAndPort skip!", pkIndex)
		return rtkMisc.ERR_BIZ_S2C_INVALID_INDEX
	}
	if source < 0 || port < 0 {
		log.Printf("[%s] Invalid (source,port)=(%d,%d)", rtkMisc.GetFuncInfo(), source, port)
		return rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	// Notify if source/port is changed
	preClientInfo, errQueryPre := QueryClientInfoByIndex(pkIndex)
	if errQueryPre == rtkMisc.SUCCESS {
		if preClientInfo.Source != source || preClientInfo.Port != port {
			emptyClientInfo := rtkCommon.ClientInfoTb{}
			emptyClientInfo.Source = preClientInfo.Source
			emptyClientInfo.Port = preClientInfo.Port
			notifyUpdateClientInfoCallback(emptyClientInfo)
		}
	}

	// Update AuthStatus
	authPkIndex := 0
	errUpsertAuthStatus := upsertAuthStatus(&authPkIndex, pkIndex, status)
	if errUpsertAuthStatus != rtkMisc.SUCCESS {
		return errUpsertAuthStatus
	}

	// Update Source, Port
	pkIndexList := make([]int, 0)
	errUpdateSrcPort := updateClientInfo(
		&pkIndexList,
		[]SqlCond{SqlCondSource, SqlCondPort},
		[]SqlCond{SqlCondPkIndex},
		source, port, pkIndex,
	)
	if errUpdateSrcPort != rtkMisc.SUCCESS {
		return errUpdateSrcPort
	}

	// Notify the lastest ClientInfoTb
	curClientInfo, errQueryCur := QueryClientInfoByIndex(pkIndex)
	if errQueryCur == rtkMisc.SUCCESS {
		notifyUpdateClientInfoCallback(curClientInfo)
	}

	return rtkMisc.SUCCESS
}

// =============================
// Misc
// =============================
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
