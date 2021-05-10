package frame

import (
	"github.com/zjh-tech/go-frame/engine/edb"
)

//同步
func SyncDoSqlOpt(exec_sql ExecSqlFunc, exec_rec ExecSqlRecordFunc, attach []interface{}, uid uint64) {
	command := NewCommonCommand(exec_sql, exec_rec, attach)
	if command == nil {
		ELog.ErrorAf("Mysql AsyncDoSqlOpt NewCommonCommand Error Uid=%v", uid)
		return
	}

	conn := edb.GDBModule.GetMysqlConn(uid)
	if conn == nil {
		ELog.ErrorAf("Mysql AsyncDoSqlOpt GetMysqlConn Error Uid=%v", uid)
		return
	}

	command.OnExecuteSql(conn)
	command.OnExecuted()
}

//异步
func AsyncDoSqlOpt(exec_sql ExecSqlFunc, exec_rec ExecSqlRecordFunc, attach []interface{}, uid uint64) {

	command := NewCommonCommand(exec_sql, exec_rec, attach)
	if command == nil {
		ELog.ErrorAf("Mysql SyncDoSqlOpt NewCommonCommand Error Uid=%v", uid)
		return
	}

	if !edb.GDBModule.AddCommand(uid, command) {
		ELog.ErrorAf("Mysql SyncDoSqlOpt AddCommand Error Uid=%v", uid)
		return
	}
}
