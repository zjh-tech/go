package edb

type IMysqlRecordSet interface {
	GetRecordSet() []map[string]string
	GetAffectRows() int64
	GetInsertId() int64
}
