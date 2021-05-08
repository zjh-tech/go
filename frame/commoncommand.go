package frame

import (
	"projects/engine/edb"
)

type ExecSqlFunc func(conn edb.IMysqlConn, attach []interface{}) (edb.IMysqlRecordSet, int32, error)
type ExecSqlRecordFunc func(record_set edb.IMysqlRecordSet, attach []interface{}, error_code int32, err error)

type CommonCommand struct {
	exec_sql   ExecSqlFunc
	exec_rec   ExecSqlRecordFunc
	attach     []interface{}
	record_set edb.IMysqlRecordSet
	error_code int32
	err        error
}

func NewCommonCommand(exec_sql ExecSqlFunc, exec_rec ExecSqlRecordFunc, attach []interface{}) *CommonCommand {
	if exec_sql == nil {
		return nil
	}

	if exec_rec == nil {
		return nil
	}
	return &CommonCommand{
		exec_sql:   exec_sql,
		exec_rec:   exec_rec,
		attach:     attach,
		record_set: nil,
		err:        nil,
	}
}

func (c *CommonCommand) SetAttach(datas []interface{}) {
	c.attach = datas
}

func (c *CommonCommand) OnExecuteSql(conn edb.IMysqlConn) {
	c.record_set, c.error_code, c.err = c.exec_sql(conn, c.attach)
}

func (c *CommonCommand) OnExecuted() {
	c.exec_rec(c.record_set, c.attach, c.error_code, c.err)
}
