package edb

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type DBModule struct {
	dbTableMaxCount uint64
	connMaxCount    uint64
	connSpecs       []*DBConnSpec
	executedQueue   chan IMysqlCommand
	conns           map[uint64]IMysqlConn
}

func (d *DBModule) Init(connMaxCount uint64, dbTableMaxCount uint64, connSpecs []*DBConnSpec) error {
	d.connMaxCount = connMaxCount
	d.dbTableMaxCount = dbTableMaxCount
	d.connSpecs = connSpecs

	if d.connMaxCount == 0 {
		return errors.New("[DBModule] Mysql ConnMaxCount = 0")
	}

	if d.connMaxCount != uint64(len(d.connSpecs)) {
		return errors.New("[DBModule] Mysql ConnMaxCount And DBConnSpec No Match")
	}

	for i := uint64(0); i < d.connMaxCount; i++ {
		dbNameSlices := strings.Split(d.connSpecs[i].Name, "_")
		if len(dbNameSlices) != 2 {
			return errors.New("[DBModule] Mysql Database must is _ split as logindb_0 Error")
		}

		//string to uint64
		dbIndex, err := strconv.ParseUint(dbNameSlices[1], 10, 64)
		if err != nil {
			return errors.New("[DBModule] Mysql Database after _ must is number as logindb_0 Error")
		}

		if connErr := d.connect(dbIndex, d.connSpecs[i].Name, d.connSpecs[i].Host, d.connSpecs[i].Port, d.connSpecs[i].User, d.connSpecs[i].Password, d.connSpecs[i].Charset); connErr != nil {
			return connErr
		}
	}

	return nil
}

func (d *DBModule) UnInit() {
	ELog.InfoA("[DB] Stop")
}

func (d *DBModule) connect(dbIndex uint64, dbName string, host string, port uint32, user string, password string, charset string) error {
	if _, ok := d.conns[dbIndex]; ok {
		errStr := fmt.Sprintf("[Mysql] DBIndex =%v DBName=%v Host=%s Port=%v Exist", dbIndex, dbName, host, port)
		return errors.New(errStr)
	}

	dsn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s?charset=%s", user, password, "tcp", host, port, dbName, charset)
	name := dbName
	mysqlConn := newMysqlConn(name, d)
	if err := mysqlConn.connect(dsn); err != nil {
		return err
	}

	ELog.Infof("[Mysql] DbIndex=%v DBName=%v Connect Host=%v Port=%v  Success", dbIndex, dbName, host, port)
	d.conns[dbIndex] = mysqlConn
	return nil
}

// bilibili must set 10 database 10 tables
// database_0  table_00  table_10 table_90
// database_1  table_01  table_11 table_91
// ...
func (d *DBModule) HashDBIndex(uid uint64) uint64 {
	ELog.DebugAf("[DBModule] UID=%v Hash DBIndex=%v", uid, uid%d.connMaxCount)
	return uid % d.connMaxCount
}

func (d *DBModule) HashTableIndex(uid uint64) uint64 {
	dbIndex := d.HashDBIndex(uid)
	dbTableIndex := uid % d.dbTableMaxCount
	ELog.DebugAf("[DBModule] UID=%v Hash TableIndex=%v", uid, dbTableIndex*10+dbIndex)
	return dbTableIndex*10 + dbIndex
}

func (d *DBModule) GetTableNameByUID(tableName string, uid uint64) string {
	tableIndex := d.HashTableIndex(uid)
	return fmt.Sprintf("%v_%02d", tableName, tableIndex)
}

func (d *DBModule) SyncQuerySqlOpt(sql string, uid uint64) (IDBResult, error) {
	return d.syncSqlOpt(sql, uid, true)
}

func (d *DBModule) SyncNonQuerySqlOpt(sql string, uid uint64) (IDBResult, error) {
	return d.syncSqlOpt(sql, uid, false)
}

func (d *DBModule) syncSqlOpt(sql string, uid uint64, queryFlag bool) (IDBResult, error) {
	dbIndex := d.HashDBIndex(uid)
	conn, ok := d.conns[dbIndex]
	if !ok {
		message := fmt.Sprintf("Mysql SyncNonQuerySql GetMysqlConn Error Uid=%v", uid)
		ELog.ErrorAf(message)
		return nil, errors.New(message)
	}

	if queryFlag {
		return conn.QuerySqlOpt(sql)
	} else {
		return conn.NonQuerySqlOpt(sql)
	}
}

func (d *DBModule) AsyncDoSqlOpt(execSql ExecSqlFunc, execRec ExecSqlRecordFunc, attach []interface{}, uid uint64) {
	command := NewDBAsyncCommand(execSql, execRec, attach)
	if command == nil {
		ELog.ErrorAf("Mysql SyncDoSqlOpt NewCommonCommand Error Uid=%v", uid)
		return
	}

	dbIndex := d.HashDBIndex(uid)
	conn, ok := d.conns[dbIndex]
	if !ok {
		ELog.ErrorAf("[DBModule] Mysql UId=%v DBIndex=%v Group Is Not Exist", uid, dbIndex)
		return
	}

	conn.AddCommand(command)
}

func (d *DBModule) AddExecutedCommand(command IMysqlCommand) {
	d.executedQueue <- command
}

func (d *DBModule) Run(loopCount int) bool {
	i := 0
	for ; i < loopCount; i++ {
		select {
		case cmd, ok := <-d.executedQueue:
			if !ok {
				return false
			}

			cmd.OnExecuted()
			return true
		default:
			return false
		}
	}
	ELog.ErrorA("[DBModule] Run Error")
	return false
}

var GDBModule *DBModule

func init() {
	GDBModule = &DBModule{
		conns:         make(map[uint64]IMysqlConn),
		executedQueue: make(chan IMysqlCommand, DbExecutedChanSize),
	}
}
