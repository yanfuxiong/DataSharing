package dbManager

import (
	"fmt"
	"log"
	"strings"
)

type SqlData string
type SqlCond string

const (
	SqlDataCreateTable SqlData = `
		CREATE TABLE IF NOT EXISTS t_client_info (
			PkIndex			INTEGER PRIMARY KEY,
			ClientId		TEXT UNIQUE,
			Host			TEXT,
			IPAddr			TEXT NOT NULL ,
			Source			INTEGER NOT NULL DEFAULT 0,
			Port			INTEGER NOT NULL DEFAULT 0,
			Online			BOOLEAN NOT NULL DEFAULT TRUE,
			DeviceName		TEXT,
			Platform		TEXT,
			Version			TEXT NOT NULL,
			UpdateTime		DATETIME NOT NULL DEFAULT (datetime('now','localtime')),
			CreateTime		DATETIME NOT NULL DEFAULT (datetime('now','localtime'))
		);
		CREATE TABLE IF NOT EXISTS t_auth_info (
			PkIndex			INTEGER PRIMARY KEY,
			ClientIndex		INTEGER UNIQUE,
			AuthStatus		BOOLEAN NOT NULL DEFAULT TRUE,
			UpdateTime		DATETIME NOT NULL DEFAULT (datetime('now','localtime')),
			CreateTime		DATETIME NOT NULL DEFAULT (datetime('now','localtime')),
			LastAuthTime	DATETIME DEFAULT NULL
		);
		CREATE TABLE IF NOT EXISTS t_timing_info (
			Source			INTEGER NOT NULL,
			Port			INTEGER NOT NULL,
			Width			INTEGER,
			Height			INTEGER,
			Framerate		INTEGER,
			UpdateTime		DATETIME NOT NULL DEFAULT (datetime('now','localtime')),
			CreateTime		DATETIME NOT NULL DEFAULT (datetime('now','localtime')),
			PRIMARY KEY (source, port)
		);
		CREATE TABLE IF NOT EXISTS t_link_info (
			ClientIndex		INTEGER PRIMARY KEY,
			Link			TEXT,
			UpdateTime		DATETIME NOT NULL DEFAULT (datetime('now','localtime')),
			CreateTime		DATETIME NOT NULL DEFAULT (datetime('now','localtime'))
		);`

	SqlDataQueryTableExist SqlData = `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name = 't_client_info';`

	SqlDataUpsertClientInfo SqlData = `
		INSERT INTO t_client_info (ClientId, Host, IPAddr, DeviceName, Platform, Version, UpdateTime)
		VALUES (?, ?, ?, ?, ?, ?, (datetime('now','localtime')))
		ON CONFLICT (ClientId)
		DO UPDATE SET
			Host		= excluded.Host,
			IPAddr		= excluded.IPAddr,
			DeviceName	= excluded.DeviceName,
			Platform	= excluded.Platform,
			Version		= excluded.Version,
			UpdateTime	= excluded.UpdateTime,
			Online		= 1
		RETURNING t_client_info.PkIndex;`

	SqlDataUpsertAuthInfo SqlData = `
		INSERT INTO t_auth_info (ClientIndex, AuthStatus, UpdateTime, LastAuthTime)
		VALUES (?, ?, (datetime('now','localtime')), (datetime('now','localtime')))
		ON CONFLICT (ClientIndex)
		DO UPDATE SET AuthStatus=excluded.AuthStatus, UpdateTime=excluded.UpdateTime, LastAuthTime=excluded.LastAuthTime
		RETURNING t_auth_info.PkIndex;`

	SqlDataUpsertUnauthInfo SqlData = `
		INSERT INTO t_auth_info (ClientIndex, AuthStatus, UpdateTime)
		VALUES (?, ?, (datetime('now','localtime')))
		ON CONFLICT (ClientIndex)
		DO UPDATE SET AuthStatus=excluded.AuthStatus, UpdateTime=excluded.UpdateTime
		RETURNING t_auth_info.PkIndex;`

	SqlDataUpdateClientInfo SqlData = `
		UPDATE t_client_info
		SET %s, UpdateTime=(datetime('now','localtime'))
		WHERE %s
		RETURNING t_client_info.PkIndex;`

	SqlDataResetAuthInfo SqlData = `
		UPDATE t_auth_info
		SET AuthStatus=0, UpdateTime=(datetime('now','localtime'))
		WHERE AuthStatus=1;`

	SqlDataQueryClientInfo SqlData = `
		SELECT t_client_info.PkIndex, ClientId, Host, IPAddr,
		Source, Port, DeviceName, Platform, Version,
		Online, COALESCE(t_auth_info.AuthStatus, 0) AS AuthStatus, t_client_info.UpdateTime, t_client_info.CreateTime,
		COALESCE(t_auth_info.LastAuthTime, '') AS LastAuthTime
		FROM t_client_info
		LEFT JOIN t_auth_info ON t_auth_info.ClientIndex=t_client_info.PkIndex
		WHERE %s
		ORDER BY t_client_info.UpdateTime DESC;`

	SqlDataUpsertTimingInfo SqlData = `
		INSERT INTO t_timing_info (Source, Port, Width, Height, Framerate, UpdateTime)
		VALUES (?, ?, ?, ?, ?, (datetime('now','localtime')))
		ON CONFLICT(Source, Port)
		DO UPDATE SET
			Width		= excluded.Width,
			Height		= excluded.Height,
			Framerate	= excluded.Framerate,
			UpdateTime	= excluded.UpdateTime;`

	SqlDataUpsertLinkInfo SqlData = `
		INSERT INTO t_link_info (ClientIndex, Link, UpdateTime)
		VALUES (?, ?, (datetime('now','localtime')))
		ON CONFLICT(ClientIndex)
		DO UPDATE SET
			Link		= excluded.Link,
			UpdateTime	= excluded.UpdateTime;`

	SqlDataQueryLinkInfo SqlData = `
		SELECT Link
		FROM t_link_info
		LEFT JOIN t_client_info ON t_client_info.pkIndex=t_link_info.ClientIndex
		WHERE %s
		ORDER BY t_link_info.UpdateTime DESC;`

	SqlDataQueryeClientMaxIndex SqlData = `SELECT PkIndex FROM t_client_info ORDER BY PkIndex DESC limit 1;`
	SqlDataQueryEarliestClient  SqlData = `SELECT PkIndex,UpdateTime FROM t_client_info WHERE Online=0 ORDER BY  UpdateTime ASC LIMIT 1;`
	SqlDataDeleteAuthInfo       SqlData = `DELETE FROM t_auth_info WHERE %s;`

	SqlCondOnline           SqlCond = "Online=1"
	SqlCondOffline          SqlCond = "Online=0"
	SqlCondAuthStatusIsTrue SqlCond = "AuthStatus=1"
	SqlCondPkIndex          SqlCond = "t_client_info.PkIndex=?"
	SqlCondClientId         SqlCond = "ClientId=?"
	SqlCondClientIndex      SqlCond = "ClientIndex=?"
	SqlCondSource           SqlCond = "Source=?"
	SqlCondPort             SqlCond = "Port=?"
	SqlCondIPAddr           SqlCond = "IPAddr=?"
	SqlCondDeviceName       SqlCond = "DeviceName=?"
	SqlCondPlatform         SqlCond = "Platform=?"
	SqlCondVersion          SqlCond = "Version=?"
	SqlCondLinkNotEmpty     SqlCond = "Link!=''"
	SqlCondLastUpdateTime   SqlCond = "(strftime('%s', 'now') - strftime('%s', t_client_info.UpdateTime)) > ?"
)

func (s SqlData) withCond_SET(conds ...SqlCond) SqlData {
	aryConds := make([]string, 0, len(conds))
	for _, cond := range conds {
		aryConds = append(aryConds, string(cond))
	}

	const key = "SET %s"
	const repKey = "SET "
	if len(aryConds) == 0 {
		return SqlData(strings.Replace(s.toString(), key, repKey, 1))
	}

	joined := strings.Join(aryConds, ", ")
	return SqlData(strings.Replace(s.toString(), key, repKey+joined, 1))
}

func (s SqlData) withCond_WHERE(conds ...SqlCond) SqlData {
	aryConds := make([]string, 0, len(conds))
	for _, cond := range conds {
		aryConds = append(aryConds, string(cond))
	}

	const key = "WHERE %s"
	const repKey = "WHERE "
	if len(aryConds) == 0 {
		return SqlData(strings.Replace(s.toString(), key, "", 1))
	}

	joined := strings.Join(aryConds, " AND ")
	return SqlData(strings.Replace(s.toString(), key, repKey+joined, 1))
}

func (s SqlData) checkArgsCount(args []any) bool {
	count := strings.Count(s.toString(), "?")
	if count != len(args) {
		log.Printf("ERROR: Count of arguments(%d) not match with SQL, expected (%d)", len(args), count)
		log.Printf("ERROR SQL content:\n%s\n", s.toString())
		return false
	}

	return true
}

func (s SqlData) toString() string {
	return string(s)
}

func (s SqlData) dump() string {
	lines := strings.Split(s.toString(), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimLeft(line, " \t")
	}
	return strings.Join(lines, "\n")
}

// ==================================
// Upgrade database version
// ==================================
const (
	latestDBVersion = 1
	
	SqlDataQueryDbVersion SqlData = `
		PRAGMA user_version;`
	SqlDataUpgradeDbVersion1 SqlData = `
		ALTER TABLE t_auth_info
		ADD COLUMN LastAuthTime DATE DEFAULT NULL;`
)

type SqlDbVerData struct {
	Ver int
	SQL SqlData
}

// If the database needs to be upgraded, it must be added in sequence
var sqlDbVerData = []SqlDbVerData{
	{Ver: 1, SQL: SqlDataUpgradeDbVersion1}, // Add column LastAuthTime in t_auth_info
	{Ver: latestDBVersion, SQL: SqlDataUpgradeDbVersion1}, // Add column LastAuthTime in t_auth_info
}

func getUpdateDbVersion(ver int) string {
	return fmt.Sprintf("PRAGMA user_version = %d;", ver)
}
