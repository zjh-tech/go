package edb

import (
	"sync"
)

var GDBAsyncCommandPool = sync.Pool{
	New: func() interface{} {
		return &DBAsyncCommand{}
	},
}

var GMysqlRecordSetPool = sync.Pool{
	New: func() interface{} {
		return &MysqlRecordSet{}
	},
}
