package frame

import (
	"projects/go-engine/edb"
)

type ExecSqlFunc func(conn edb.IMysqlConn, attach []interface{}) (edb.IMysqlRecordSet, int32, error)
type ExecSqlRecordFunc func(recordSet edb.IMysqlRecordSet, attach []interface{}, errorCode int32, err error)

type CommonCommand struct {
	execSql   ExecSqlFunc
	execRec   ExecSqlRecordFunc
	attach    []interface{}
	recordSet edb.IMysqlRecordSet
	errorCode int32
	err       error
}

func NewCommonCommand(execSql ExecSqlFunc, execRec ExecSqlRecordFunc, attach []interface{}) *CommonCommand {
	if execSql == nil {
		return nil
	}

	if execRec == nil {
		return nil
	}

	return &CommonCommand{
		execSql:   execSql,
		execRec:   execRec,
		attach:    attach,
		recordSet: nil,
		err:       nil,
	}
}

func (c *CommonCommand) SetAttach(datas []interface{}) {
	c.attach = datas
}

func (c *CommonCommand) OnExecuteSql(conn edb.IMysqlConn) {
	c.recordSet, c.errorCode, c.err = c.execSql(conn, c.attach)
}

func (c *CommonCommand) OnExecuted() {
	c.execRec(c.recordSet, c.attach, c.errorCode, c.err)
}
