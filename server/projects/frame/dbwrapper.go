package frame

import (
	"projects/go-engine/edb"
	"projects/go-engine/elog"
)

//同步
func SyncDoSqlOpt(execSql ExecSqlFunc, execRec ExecSqlRecordFunc, attach []interface{}, uid uint64) {
	command := NewCommonCommand(execSql, execRec, attach)
	if command == nil {
		elog.ErrorAf("Mysql AsyncDoSqlOpt NewCommonCommand Error Uid=%v", uid)
		return
	}

	conn := edb.GDBModule.GetMysqlConn(uid)
	if conn == nil {
		elog.ErrorAf("Mysql AsyncDoSqlOpt GetMysqlConn Error Uid=%v", uid)
		return
	}

	command.OnExecuteSql(conn)
	command.OnExecuted()
}

//异步
func AsyncDoSqlOpt(execSql ExecSqlFunc, execRec ExecSqlRecordFunc, attach []interface{}, uid uint64) {

	command := NewCommonCommand(execSql, execRec, attach)
	if command == nil {
		elog.ErrorAf("Mysql SyncDoSqlOpt NewCommonCommand Error Uid=%v", uid)
		return
	}

	if !edb.GDBModule.AddCommand(uid, command) {
		elog.ErrorAf("Mysql SyncDoSqlOpt AddCommand Error Uid=%v", uid)
		return
	}
}
