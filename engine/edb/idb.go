package edb

import "database/sql"

type IMysqlCommand interface {
	//协程执行mysql操作
	OnExecuteSql(IMysqlConn)
	//主协程处理返回结果
	OnExecuted()
}

type IMysqlRecordSet interface {
	GetRecordSet() []map[string]string
	GetAffectRows() int64
	GetInsertId() int64
}

type IMysqlConn interface {
	QueryWithResult(sql string) (IMysqlRecordSet, error)
	QueryWithoutResult(sql string) (IMysqlRecordSet, error)
	FindSqlDb() *sql.DB
	BeginTransact()
	CommitTransact()
	RollbackTransact()

	AddComand(command IMysqlCommand)
}
