package edb

import (
	"fmt"
	"strconv"
)

func Str2Uint64(str string) (uint64, bool) {
	n, err := strconv.ParseUint(str, 10, 64)
	if err == nil {
		return n, true
	}

	return 0, false
}

type DBConnSpec struct {
	Name     string
	Host     string
	Port     uint32
	User     string
	Password string
	Charset  string
}

const (
	DbExecSuccess int32 = 10000
	DbExecFail    int32 = 10001
)

const DbDefaultInsertId int64 = 0
const DbDefaultAffectedRows int64 = 0

const DbWaitChanSize = 1024 * 10 * 10
const DbExecutedChanSize = 1024 * 10 * 10

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
