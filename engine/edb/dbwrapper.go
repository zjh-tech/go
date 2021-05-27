package edb

//同步
func SyncDoSqlOpt(execSql ExecSqlFunc, execRec ExecSqlRecordFunc, attach []interface{}, uid uint64) {
	command := NewCommonCommand(execSql, execRec, attach)
	if command == nil {
		ELog.ErrorAf("Mysql AsyncDoSqlOpt NewCommonCommand Error Uid=%v", uid)
		return
	}

	conn := GDBModule.GetMysqlConn(uid)
	if conn == nil {
		ELog.ErrorAf("Mysql AsyncDoSqlOpt GetMysqlConn Error Uid=%v", uid)
		return
	}

	command.OnExecuteSql(conn)
	command.OnExecuted()
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
