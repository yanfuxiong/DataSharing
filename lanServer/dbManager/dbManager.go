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
	"strings"
	"sync"
	"time"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

const (
	g_KeppDBAliveInterval = 1 // minute
	g_ReconnListInterval  = 2 // seconds
	g_ReCreateDbRetry     = 5
)

var (
	g_SqlInstance *sql.DB //	sqlite Instance
	dbMutex       sync.Mutex
)

const (
	// Not allow setup WAL by SELinux rules
	g_DBConnectionStr = "file:" + rtkGlobal.DB_PATH + rtkGlobal.DB_NAME + "?cache=shared&mode=rwc" //sqlite connect string
)

func init() {
	g_SqlInstance = nil
}

func reCreateDb() error {
	filePath := rtkGlobal.DB_PATH + rtkGlobal.DB_NAME
	if rtkMisc.FileExists(filePath) {
		err := os.Remove(filePath)
		if err != nil {
			log.Printf("Remove corrupt db:[%s] err:%+v", filePath, err)
			return errors.New("Remove db error!")
		}
	}

	return createDb()
}


func createDb() error {
	err := rtkMisc.CreateDir(rtkGlobal.DB_PATH, os.ModePerm)
	if err != nil {
		log.Printf("Create DB failed: %s", err.Error())
		return errors.New("buildDbPath error!")
	}

	db, err := sql.Open("sqlite3", g_DBConnectionStr)
	if err != nil {
		return err
	}
	db.SetConnMaxIdleTime(time.Duration(0)) //connections are not closed due to a connection's idle time
	db.SetMaxIdleConns(10)

	initDB := func() error {
		tx, err := db.Begin()
		if err != nil {
			log.Printf("[%s] Err: %s", rtkMisc.GetFuncInfo(), err.Error())
			return err
		}
		defer tx.Rollback()

		var tableCnt int
		if err = tx.QueryRow(SqlDataQueryTableExist.toString()).Scan(&tableCnt); err != nil {
			log.Printf("[%s] Scan Err: %s", rtkMisc.GetFuncInfo(), err.Error())
			return err
		}

		var version int
		if err = tx.QueryRow(SqlDataQueryDbVersion.toString()).Scan(&version); err != nil {
			log.Printf("[%s] Scan Err: %s", rtkMisc.GetFuncInfo(), err.Error())
			return err
		}

		//	If the database has not been created, create and update to the latest version
		if tableCnt == 0 && version == 0 {
			_, errUpdateVer := tx.Exec(getUpdateDbVersion(latestDBVersion))
			if errUpdateVer != nil {
				log.Printf("[%s] Exec Err: %s", rtkMisc.GetFuncInfo(), errUpdateVer.Error())
				return errUpdateVer
			}
			log.Printf("[%s] table t_client_info is not exist, Upgrade database version from default(0) to (%d)", rtkMisc.GetFuncInfo(), latestDBVersion)
		}

		_, err = tx.Exec(SqlDataCreateTable.toString())
		if err != nil {
			log.Printf("[%s] Exec Err: %s", rtkMisc.GetFuncInfo(), err.Error())
			return err
		}

		_, err = tx.Exec(SqlDataResetAuthInfo.toString())
		if err != nil {
			log.Printf("[%s] Exec Err: %s", rtkMisc.GetFuncInfo(), err.Error())
			return err
		}

		if err = tx.Commit(); err != nil {
			log.Printf("[%s] Commit Err: %s", rtkMisc.GetFuncInfo(), err.Error())
		}
 		return err
 	}

	initErr := initDB()
	if initErr != nil {
		db.Close()
		return initErr
 	}
 
	dbMutex.Lock()
	g_SqlInstance = db
	dbMutex.Unlock()

	return nil
}


func InitSqlite(ctx context.Context) {
	err := createDb()
	if err != nil {
		log.Printf("Create database failed. Err: %s", err.Error())

		retryRet := false
		for i := 1; i < g_ReCreateDbRetry; i++ {
			time.Sleep(time.Duration(i*50) * time.Millisecond)
			log.Printf("Create database failed. Retry(%d/%d)", i, g_ReCreateDbRetry)
			errRetry := reCreateDb()
			if errRetry == nil {
				retryRet = true
				break
			}
			log.Printf("Re-Create database filed. Err: %s", errRetry.Error())
		}

		if !retryRet {
			log.Println("[WARN] Cannot create database. CrossShare Server will be abnormal!")
			return
		}
	}

	upgradeDb() //If the database is an old version, update it to the latest version in sequence

	UpdateAllClientOffline()

	rtkMisc.GoSafe(func() {
		tkKeepAlive := time.NewTicker(time.Duration(g_KeppDBAliveInterval) * time.Minute)
		defer tkKeepAlive.Stop()

		for {
			select {
			case <-ctx.Done():
				dbMutex.Lock()
				g_SqlInstance.Close()
				g_SqlInstance = nil
				dbMutex.Unlock()
				return
			case <-tkKeepAlive.C:
				keppAlive()
			}
		}
	})
}

func getDb() (*sql.DB, bool) {
	if g_SqlInstance == nil {
		log.Println("[ERROR] Database instance is null!")
		return nil, false
	}
	return g_SqlInstance, true
}

func upgradeDbVer(updateVer int, sqlData SqlData) rtkMisc.CrossShareErr {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	db, ok := getDb()
	if !ok {
		return rtkMisc.ERR_DB_SQLITE_INSTANCE_NULL
	}
	var version int
	row := db.QueryRow(SqlDataQueryDbVersion.toString())
	if err := row.Scan(&version); err != nil {
		log.Printf("[%s] Err: %s", rtkMisc.GetFuncInfo(), err.Error())
		return rtkMisc.ERR_DB_SQLITE_EXEC
	}

	if version >= updateVer {
		return rtkMisc.SUCCESS
	}

	log.Printf("[%s] Upgrade database version from (%d) to (%d)", rtkMisc.GetFuncInfo(), version, updateVer)
	tx, err := db.Begin()
	if err != nil {
		log.Printf("[%s] Err: %s", rtkMisc.GetFuncInfo(), err.Error())
		return rtkMisc.ERR_DB_SQLITE_BEGIN
	}
	defer tx.Rollback()

	_, errUpgrade := tx.Exec(sqlData.toString())
	if errUpgrade != nil {
		log.Printf("[%s] Err: %s", rtkMisc.GetFuncInfo(), errUpgrade.Error())
		// Special case
		if (version == 0) && strings.Contains(errUpgrade.Error(), "LastAuthTime") {
			log.Printf("[%s] Skip the error: [duplicate column name: LastAuthTime] for compatibility", rtkMisc.GetFuncInfo())
		} else {
			return rtkMisc.ERR_DB_SQLITE_EXEC
		}
	}

	_, errUpdateVer := tx.Exec(getUpdateDbVersion(updateVer))
	if errUpdateVer != nil {
		log.Printf("[%s] Err: %s", rtkMisc.GetFuncInfo(), errUpdateVer.Error())
		return rtkMisc.ERR_DB_SQLITE_EXEC
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("[%s] Err: %s", rtkMisc.GetFuncInfo(), err.Error())
		return rtkMisc.ERR_DB_SQLITE_COMMIT
	}

	return rtkMisc.SUCCESS
}

func upgradeDb() {
	for _, data := range sqlDbVerData {
		err := upgradeDbVer(data.Ver, data.SQL)
		if err != rtkMisc.SUCCESS {
			log.Printf("[%s] Err: Upgrade database version:[%d] failed: %s", rtkMisc.GetFuncInfo(), data.Ver, rtkMisc.GetResponse(err).Msg)
			break
		}
	}
}

func upsertClientInfo(pkIndex *int, clientId, host, ipAddr, deviceName, platform, version string) rtkMisc.CrossShareErr {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if pkIndex == nil {
		log.Printf("[%s] pkIndex is null", rtkMisc.GetFuncInfo())
		return rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	db, ok := getDb()
	if !ok {
		return rtkMisc.ERR_DB_SQLITE_INSTANCE_NULL
	}
	tx, err := db.Begin()
	if err != nil {
		log.Printf("[%s] Err: %s", rtkMisc.GetFuncInfo(), err.Error())
		return rtkMisc.ERR_DB_SQLITE_BEGIN
	}
	defer tx.Rollback()

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
		param := []any{clientId, host, ipAddr, deviceName, platform, version}
		if sqlData.checkArgsCount(param) == false {
			return rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
		}

		row = tx.QueryRow(sqlData.toString(), param...)
		if err = row.Scan(pkIndex); err != nil {
			log.Printf("[%s] Err: %s", rtkMisc.GetFuncInfo(), err.Error())
			return rtkMisc.ERR_DB_SQLITE_SCAN
		}
	} else {
		sqlData = SqlDataQueryEarliestClient
		var OldIndex int
		var OldUpdate string
		rowIndex := tx.QueryRow(sqlData.toString())
		if err = rowIndex.Scan(&OldIndex, &OldUpdate); err != nil {
			log.Printf("[%s] Err: %s", rtkMisc.GetFuncInfo(), err.Error())
			return rtkMisc.ERR_DB_SQLITE_SCAN
		}

		log.Printf("[%s] get index:%d, is over range, and earliest client PkIndex:%d updatetime:%s, bein to replace it", rtkMisc.GetFuncInfo(), maxIndex+1, OldIndex, OldUpdate)

		args := []any{clientId, ipAddr, deviceName, platform, version, OldIndex}
		sqlData = SqlDataUpdateClientInfo.withCond_SET(SqlCondClientId, SqlCondIPAddr, SqlCondDeviceName, SqlCondPlatform, SqlCondVersion, SqlCondOnline).withCond_WHERE(SqlCondPkIndex)
		if !sqlData.checkArgsCount(args) {
			return rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
		}
		rowIndex = tx.QueryRow(sqlData.toString(), args...)
		if err = rowIndex.Scan(pkIndex); err != nil {
			log.Printf("[%s] Err: %s", rtkMisc.GetFuncInfo(), err.Error())
			return rtkMisc.ERR_DB_SQLITE_SCAN
		}

		sqlData = SqlDataDeleteAuthInfo.withCond_WHERE(SqlCondClientIndex)
		_, err = tx.Exec(sqlData.toString(), OldIndex)
		if err != nil {
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
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if authPkIndex == nil {
		log.Printf("[%s] pkIndex is null", rtkMisc.GetFuncInfo())
		return rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	var sqlData SqlData
	if authStatus {
		sqlData = SqlDataUpsertAuthInfo
	} else {
		sqlData = SqlDataUpsertUnauthInfo
	}
	param := []any{clientPkIndex, authStatus}
	if sqlData.checkArgsCount(param) == false {
		return rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	db, ok := getDb()
	if !ok {
		return rtkMisc.ERR_DB_SQLITE_INSTANCE_NULL
	}
	row := db.QueryRow(sqlData.toString(), param...)
	if err := row.Scan(authPkIndex); err != nil {
		log.Printf("[%s] Err: %s", rtkMisc.GetFuncInfo(), err.Error())
		return rtkMisc.ERR_DB_SQLITE_SCAN
	}

	return rtkMisc.SUCCESS
}

func updateClientInfo(pkIndexList *[]int, setConds []SqlCond, whereConds []SqlCond, args ...any) rtkMisc.CrossShareErr {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if pkIndexList == nil {
		log.Printf("[%s] pkIndexList is null", rtkMisc.GetFuncInfo())
		return rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	sqlData := SqlDataUpdateClientInfo.withCond_SET(setConds...).withCond_WHERE(whereConds...)
	if sqlData.checkArgsCount(args) == false {
		return rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	db, ok := getDb()
	if !ok {
		return rtkMisc.ERR_DB_SQLITE_INSTANCE_NULL
	}
	rows, err := db.Query(sqlData.toString(), args...)
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

func queryClientInfoInternal(clientInfoList *[]rtkCommon.ClientInfoTb, conds []SqlCond, args ...any) rtkMisc.CrossShareErr {
	if clientInfoList == nil {
		log.Printf("[%s] ClientInfoList is null", rtkMisc.GetFuncInfo())
		return rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	sqlData := SqlDataQueryClientInfo.withCond_WHERE(conds...)
	if sqlData.checkArgsCount(args) == false {
		return rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	db, ok := getDb()
	if !ok {
		return rtkMisc.ERR_DB_SQLITE_INSTANCE_NULL
	}
	rows, err := db.Query(sqlData.toString(), args...)
	if err != nil {
		log.Printf("[%s] Query error[%+v]", rtkMisc.GetFuncInfo(), err)
		log.Printf("[%s] Err: %s", rtkMisc.GetFuncInfo(), sqlData.toString())
		return rtkMisc.ERR_DB_SQLITE_QUERY
	}

	*clientInfoList = (*clientInfoList)[:0]
	for rows.Next() {
		var client rtkCommon.ClientInfoTb
		if err = rows.Scan(&client.Index, &client.ClientId, &client.Host, &client.IpAddr,
			&client.Source, &client.Port, &client.DeviceName, &client.Platform, &client.Version,
			&client.Online, &client.AuthStatus, &client.UpdateTime, &client.CreateTime, &client.LastAuthTime); err != nil {
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

func queryClientInfo(clientInfoList *[]rtkCommon.ClientInfoTb, conds []SqlCond, args ...any) rtkMisc.CrossShareErr {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	return queryClientInfoInternal(clientInfoList, conds, args...)
}

func clientQueryClientList(clientPkIndex int, clientInfoList *[]rtkCommon.ClientInfoTb) rtkMisc.CrossShareErr {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	if clientPkIndex <= 0 {
		log.Printf("[%s] clientPkIndex:[%d] is invalid!", rtkMisc.GetFuncInfo(), clientPkIndex)
		return rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	if clientInfoList == nil {
		log.Printf("[%s] ClientInfoList is null", rtkMisc.GetFuncInfo())
		return rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	sqlData := SqlDataUpdateClientInfo.withCond_SET(SqlCondGetClientListTrue).withCond_WHERE(SqlCondPkIndex)
	var pkIndex int
	db, ok := getDb()
	if !ok {
		return rtkMisc.ERR_DB_SQLITE_INSTANCE_NULL
	}
	rowIndex := db.QueryRow(sqlData.toString(), clientPkIndex)
	if err := rowIndex.Scan(&pkIndex); err != nil {
		log.Printf("[%s] Err: %s", rtkMisc.GetFuncInfo(), err.Error())
		return rtkMisc.ERR_DB_SQLITE_SCAN
	}

	clientQueryCond := []SqlCond{SqlCondOnline, SqlCondAuthStatusIsTrue, SqlCondGetClientListTrue}
	return queryClientInfoInternal(clientInfoList, clientQueryCond)
}

func upsertTimingInfo(source, port, width, height, framerate int) rtkMisc.CrossShareErr {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	sqlData := SqlDataUpsertTimingInfo
	param := []any{source, port, width, height, framerate}
	if sqlData.checkArgsCount(param) == false {
		return rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	db, ok := getDb()
	if !ok {
		return rtkMisc.ERR_DB_SQLITE_INSTANCE_NULL
	}
	_, err := db.Exec(sqlData.toString(), param...)
	if err != nil {
		log.Printf("[%s] Query error[%+v]", rtkMisc.GetFuncInfo(), err)
		log.Printf("[%s] Err: %s", rtkMisc.GetFuncInfo(), sqlData.toString())
		return rtkMisc.ERR_DB_SQLITE_QUERY
	}

	return rtkMisc.SUCCESS
}

func upsertLinkInfo(pkIndex int, link string) rtkMisc.CrossShareErr {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	sqlData := SqlDataUpsertLinkInfo
	param := []any{pkIndex, link}
	if sqlData.checkArgsCount(param) == false {
		return rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	db, ok := getDb()
	if !ok {
		return rtkMisc.ERR_DB_SQLITE_INSTANCE_NULL
	}
	_, err := db.Exec(sqlData.toString(), param...)
	if err != nil {
		log.Printf("[%s] Query error[%+v]", rtkMisc.GetFuncInfo(), err)
		log.Printf("[%s] Err: %s", rtkMisc.GetFuncInfo(), sqlData.toString())
		return rtkMisc.ERR_DB_SQLITE_QUERY
	}

	return rtkMisc.SUCCESS
}

func queryLinkInfo(conds []SqlCond, args ...any) (string, rtkMisc.CrossShareErr) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	sqlData := SqlDataQueryLinkInfo.withCond_WHERE(conds...)
	if sqlData.checkArgsCount(args) == false {
		return "", rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	var link string = ""
	db, ok := getDb()
	if !ok {
		return "", rtkMisc.ERR_DB_SQLITE_INSTANCE_NULL
	}
	row := db.QueryRow(sqlData.toString(), args...)
	if err := row.Scan(&link); err != nil {
		log.Printf("[%s] Err: %s", rtkMisc.GetFuncInfo(), err.Error())
		return "", rtkMisc.ERR_DB_SQLITE_SCAN
	}

	return link, rtkMisc.SUCCESS
}

func upsertSrcPortInfo(source, port, clientPkIndex, udpMousePort, udpKeyboardPort int) rtkMisc.CrossShareErr {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if source <= 0 || port < 0 || clientPkIndex <= 0 || udpMousePort <= 0 || udpKeyboardPort <= 0 {
		log.Printf("[%s] %d %d %d %d %d, args is invalid!", rtkMisc.GetFuncInfo(), clientPkIndex, source, port, udpMousePort, udpKeyboardPort)
		return rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	sqlData := SqlDataUpsertSrcPortInfo
	param := []any{source, port, clientPkIndex, udpMousePort, udpKeyboardPort}
	if sqlData.checkArgsCount(param) == false {
		return rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	db, ok := getDb()
	if !ok {
		return rtkMisc.ERR_DB_SQLITE_INSTANCE_NULL
	}

	_, err := db.Exec(sqlData.toString(), param...)
	if err != nil {
		log.Printf("[%s] Exec error[%+v]", rtkMisc.GetFuncInfo(), err)
		log.Printf("[%s] Err: %s", rtkMisc.GetFuncInfo(), sqlData.toString())
		return rtkMisc.ERR_DB_SQLITE_QUERY
	}

	return rtkMisc.SUCCESS
}

func querySrcPortInfo(clientPkIndex int, srcPortInfoList *[]rtkMisc.SourcePortInfo) rtkMisc.CrossShareErr {
	if clientPkIndex <= 0 {
		log.Printf("[%s] clientPkIndex:%d, args is invalid!", rtkMisc.GetFuncInfo(), clientPkIndex)
		return rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}
	dbMutex.Lock()
	defer dbMutex.Unlock()

	sqlData := SqlDataQuerySrcPortInfo.withCond_WHERE(SqlCondClientIndex)
	param := []any{clientPkIndex}
	if sqlData.checkArgsCount(param) == false {
		return rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	db, ok := getDb()
	if !ok {
		return rtkMisc.ERR_DB_SQLITE_INSTANCE_NULL
	}
	rows, err := db.Query(sqlData.toString(), param...)
	if err != nil {
		log.Printf("[%s] Query error[%+v]", rtkMisc.GetFuncInfo(), err)
		log.Printf("[%s] Err sql: %s", rtkMisc.GetFuncInfo(), sqlData.toString())
		return rtkMisc.ERR_DB_SQLITE_QUERY
	}

	*srcPortInfoList = (*srcPortInfoList)[:0]
	for rows.Next() {
		var srcPortInfo rtkMisc.SourcePortInfo
		if err = rows.Scan(&srcPortInfo.Source, &srcPortInfo.Port, &srcPortInfo.UdpMousePort, &srcPortInfo.UdpKeyboardPort); err != nil {
			log.Printf("[%s] Err: %s", rtkMisc.GetFuncInfo(), err.Error())
			continue
		}
		*srcPortInfoList = append(*srcPortInfoList, srcPortInfo)
	}

	defer rows.Close()
	if err = rows.Err(); err != nil {
		log.Printf("[%s] rows err:%+v", rtkMisc.GetFuncInfo(), err)
		return rtkMisc.ERR_DB_SQLITE_EXEC
	}

	return rtkMisc.SUCCESS
}

func resetSrcPortInfo(source, port int) rtkMisc.CrossShareErr {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if source <= 0 || port < 0 {
		log.Printf("[%s] source:%d port:%d , args is invalid!", rtkMisc.GetFuncInfo(), source, port)
		return rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	db, ok := getDb()
	if !ok {
		return rtkMisc.ERR_DB_SQLITE_INSTANCE_NULL
	}

	sqlData := SqlDataResetSrcPortInfo
	sqlData.withCond_WHERE(SqlCondSource, SqlCondPort)
	param := []any{source, port}
	if sqlData.checkArgsCount(param) == false {
		return rtkMisc.ERR_DB_SQLITE_INVALID_ARGS
	}

	_, err := db.Exec(sqlData.toString(), param...)
	if err != nil {
		log.Printf("[%s] Exec error[%+v]", rtkMisc.GetFuncInfo(), err)
		log.Printf("[%s] Err sql: %s", rtkMisc.GetFuncInfo(), sqlData.toString())
		return rtkMisc.ERR_DB_SQLITE_QUERY
	}

	return rtkMisc.SUCCESS
}

// =============================
// Misc
// =============================
func keppAlive() {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	db, ok := getDb()
	if !ok {
		return
	}

	err := db.Ping()
	if err != nil {
		db.Close()
		g_SqlInstance = nil
		log.Printf("sqlite3 Ping error:%+v,  so reconnect it!", err)
		newDb, err := sql.Open("sqlite3", g_DBConnectionStr)
		if err != nil {
			log.Printf("Open sqlite3 [%s] err:%+v", g_DBConnectionStr, err)
			return
		}

		newDb.SetConnMaxIdleTime(time.Duration(0)) //connections are not closed due to a connection's idle time
		newDb.SetMaxIdleConns(10)
		g_SqlInstance = newDb
	}

	// DEBUG log
	// log.Printf("[%s] sqlite3 is alive!", rtkMisc.GetFuncInfo())
}
