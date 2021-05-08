package edb

import "fmt"

type DBConnSpec struct {
	Ip       string
	Port     uint32
	User     string
	Password string
	Name     string
	Charset  string
}

const (
	DB_EXEC_SUCCESS int32 = 10000
	DB_EXEC_FAIL    int32 = 10001
)

const DB_DEFAULT_INSERT_ID int64 = 0
const DB_DEFAULT_AFFECTED_ROWS int64 = 0

const DB_WAIT_QUEUE_CHAN_SIZE = 1024 * 10 * 10
const DB_EXECUTED_QUEUE_CHAN_SIZE = 1024 * 10 * 10

const DBMajorVersion = 1
const DBMinorVersion = 1

type DBVersion struct {
}

func (d *DBVersion) GetVersion() string {
	return fmt.Sprintf("DB Version: %v.%v", DBMajorVersion, DBMinorVersion)
}

var GDBVersion *DBVersion

func init() {
	GDBVersion = &DBVersion{}
}
