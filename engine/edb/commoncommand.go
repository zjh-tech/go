package edb

type ExecSqlFunc func(conn IMysqlConn, attach []interface{}) (IDBResult, error)
type ExecSqlRecordFunc func(recordSet IDBResult, attach []interface{}, err error)

type DBAsyncCommand struct {
	execSqlFunc ExecSqlFunc
	execRecFunc ExecSqlRecordFunc
	attach      []interface{}
	recordSet   IDBResult
	err         error
}

func NewDBAsyncCommand(execSqlFunc ExecSqlFunc, execRecFunc ExecSqlRecordFunc, attach []interface{}) *DBAsyncCommand {
	if execSqlFunc == nil {
		return nil
	}

	if execRecFunc == nil {
		return nil
	}

	cmd := GDBAsyncCommandPool.Get().(*DBAsyncCommand)
	cmd.execSqlFunc = execSqlFunc
	cmd.execRecFunc = execRecFunc
	cmd.attach = attach
	cmd.recordSet = nil
	cmd.err = nil
	return cmd
}

func (d *DBAsyncCommand) SetAttach(datas []interface{}) {
	d.attach = datas
}

func (d *DBAsyncCommand) OnExecuteSql(conn IMysqlConn) {
	d.recordSet, d.err = d.execSqlFunc(conn, d.attach)
}

func (d *DBAsyncCommand) OnExecuted() {
	d.execRecFunc(d.recordSet, d.attach, d.err)
}
