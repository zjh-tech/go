package edb

type ExecSqlFunc func(conn IMysqlConn, attach []interface{}) (IDBResult, int32, error)
type ExecSqlRecordFunc func(recordSet IDBResult, attach []interface{}, errorCode int32, err error)

type CommonCommand struct {
	execSqlFunc ExecSqlFunc
	execRecFunc ExecSqlRecordFunc
	attach      []interface{}
	recordSet   IDBResult
	errorCode   int32
	err         error
}

func NewCommonCommand(execSqlFunc ExecSqlFunc, execRecFunc ExecSqlRecordFunc, attach []interface{}) *CommonCommand {
	if execSqlFunc == nil {
		return nil
	}

	if execRecFunc == nil {
		return nil
	}
	return &CommonCommand{
		execSqlFunc: execSqlFunc,
		execRecFunc: execRecFunc,
		attach:      attach,
		recordSet:   nil,
		err:         nil,
	}
}

func (c *CommonCommand) SetAttach(datas []interface{}) {
	c.attach = datas
}

func (c *CommonCommand) OnExecuteSql(conn IMysqlConn) {
	c.recordSet, c.errorCode, c.err = c.execSqlFunc(conn, c.attach)
}

func (c *CommonCommand) OnExecuted() {
	c.execRecFunc(c.recordSet, c.attach, c.errorCode, c.err)
}

//---------------------------------------------------------------------------------------------------------
type SyncCommonCommand struct {
	sql       string
	queryFlag bool
	recordSet IDBResult
	err       error
}

func NewSyncCommonCommand(sql string, queryFlag bool) *SyncCommonCommand {
	return &SyncCommonCommand{
		sql:       sql,
		queryFlag: queryFlag,
	}
}

func (c *SyncCommonCommand) OnExecuteSql(conn IMysqlConn) {
	if c.queryFlag {
		c.recordSet, c.err = conn.QuerySqlOpt(c.sql)
	} else {
		c.recordSet, c.err = conn.NonQuerySqlOpt(c.sql)
	}
}

func (c *SyncCommonCommand) GetExecuteSqlResult() (IDBResult, error) {
	return c.recordSet, c.err
}
