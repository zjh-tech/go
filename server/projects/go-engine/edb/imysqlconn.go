package edb

import "database/sql"

const (
	DB_EXEC_SUCCESS int32 = 10000
	DB_EXEC_FAIL    int32 = 10001
)

type IMysqlConn interface {
	QueryWithResult(sql string) (IMysqlRecordSet, error)
	QueryWithoutResult(sql string) (IMysqlRecordSet, error)
	FindSqlDb() *sql.DB
	BeginTransact()
	CommitTransact()
	RollbackTransact()

	AddComand(command IMysqlCommand)
}
