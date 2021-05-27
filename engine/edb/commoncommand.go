package edb

type ExecSqlFunc func(conn IMysqlConn, attach []interface{}) (IMysqlRecordSet, int32, error)
type ExecSqlRecordFunc func(record_set IMysqlRecordSet, attach []interface{}, error_code int32, err error)

type CommonCommand struct {
	execSql    ExecSqlFunc
	execRec    ExecSqlRecordFunc
	attach     []interface{}
	record_set IMysqlRecordSet
	error_code int32
	err        error
}

func NewCommonCommand(execSql ExecSqlFunc, execRec ExecSqlRecordFunc, attach []interface{}) *CommonCommand {
	if execSql == nil {
		return nil
	}

	if execRec == nil {
		return nil
	}
	return &CommonCommand{
		execSql:    execSql,
		execRec:    execRec,
		attach:     attach,
		record_set: nil,
		err:        nil,
	}
}

func (c *CommonCommand) SetAttach(datas []interface{}) {
	c.attach = datas
}

func (c *CommonCommand) OnExecuteSql(conn IMysqlConn) {
	c.record_set, c.error_code, c.err = c.execSql(conn, c.attach)
}

func (c *CommonCommand) OnExecuted() {
	c.execRec(c.record_set, c.attach, c.error_code, c.err)
}
