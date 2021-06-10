package edb

import (
	"errors"
	"fmt"
)

//同步
func SyncQuerySqlOpt(sql string, uid uint64) (IMysqlRecordSet, error) {
	command := NewSyncCommonCommand(sql, true)
	if command == nil {
		return nil, errors.New("NewSyncCommonCommand Error")
	}
	conn := GDBModule.GetMysqlConn(uid)
	if conn == nil {
		return nil, errors.New(fmt.Sprintf("Mysql SyncQuerySql GetMysqlConn Error Uid=%v", uid))
	}

	command.OnExecuteSql(conn)
	return command.GetExecuteSqlResult()
}

func SyncInsertSqlOpt(sql string, uid uint64) (IMysqlRecordSet, error) {
	return SyncNonQuerySqlOpt(sql, uid)
}

func SyncUpdateSqlOpt(sql string, uid uint64) (IMysqlRecordSet, error) {
	return SyncNonQuerySqlOpt(sql, uid)
}

func SyncDeleteSqlOpt(sql string, uid uint64) (IMysqlRecordSet, error) {
	return SyncNonQuerySqlOpt(sql, uid)
}

func SyncInsertOrUpdateSqlOpt(sql string, uid uint64) (IMysqlRecordSet, error) {
	return SyncNonQuerySqlOpt(sql, uid)
}

func SyncNonQuerySqlOpt(sql string, uid uint64) (IMysqlRecordSet, error) {
	command := NewSyncCommonCommand(sql, false)
	if command == nil {
		return nil, errors.New("NewSyncCommonCommand Error")
	}
	conn := GDBModule.GetMysqlConn(uid)
	if conn == nil {
		return nil, errors.New(fmt.Sprintf("Mysql SyncNonQuerySql GetMysqlConn Error Uid=%v", uid))
	}

	command.OnExecuteSql(conn)
	return command.GetExecuteSqlResult()
}

//异步
func AsyncDoSqlOpt(execSql ExecSqlFunc, execRec ExecSqlRecordFunc, attach []interface{}, uid uint64) {
	command := NewCommonCommand(execSql, execRec, attach)
	if command == nil {
		ELog.ErrorAf("Mysql SyncDoSqlOpt NewCommonCommand Error Uid=%v", uid)
		return
	}

	if !GDBModule.AddCommand(uid, command) {
		ELog.ErrorAf("Mysql SyncDoSqlOpt AddCommand Error Uid=%v", uid)
		return
	}
}
