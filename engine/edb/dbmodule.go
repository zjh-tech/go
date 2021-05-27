package edb

import (
	"errors"
	"fmt"
	"strings"

	"github.com/zjh-tech/go-frame/base/convert"
)

type DBModule struct {
	dbTableMaxCount uint64
	connMaxCount    uint64
	connSpecs       []*DBConnSpec
	executedQueue   chan IMysqlCommand
	conns           map[uint64]IMysqlConn
}

func (d *DBModule) Init(connMaxCount uint64, dbTableMaxCount uint64, connSpecs []*DBConnSpec) error {
	d.connMaxCount = connMaxCount
	d.dbTableMaxCount = dbTableMaxCount
	d.connSpecs = connSpecs

	if d.connMaxCount == 0 {
		return errors.New("[DBModule] No Mysql Connect")
	}

	if d.connMaxCount != uint64(len(d.connSpecs)) {
		return errors.New("[DBModule] Mysql No Match")
	}

	for i := uint64(0); i < d.connMaxCount; i++ {
		dbNameSlices := strings.Split(d.connSpecs[i].Name, "_")
		if len(dbNameSlices) != 2 {
			return errors.New("[DBModule] Mysql Index Error")
		}

		dbIndex, _ := convert.Str2Uint64(dbNameSlices[1])
		connErr := d.connect(dbIndex,
			d.connSpecs[i].Name, d.connSpecs[i].Ip,
			d.connSpecs[i].Port, d.connSpecs[i].User,
			d.connSpecs[i].Password, d.connSpecs[i].Charset)

		if connErr != nil {
			ELog.Errorf("[DBModule] Connect Mysql DBName=%v DBIp=%v,DBPort=%v, Error = %v",
				d.connSpecs[i].Name,
				d.connSpecs[i].Ip,
				d.connSpecs[i].Port,
				connErr)
			return connErr
		}
	}

	return nil
}

func (d *DBModule) UnInit() {
	ELog.InfoA("[DB] Stop")
}

func (d *DBModule) connect(dbIndex uint64, db_name string, ip string, port uint32, user string, password string, charset string) error {
	if _, ok := d.conns[dbIndex]; ok {
		errStr := fmt.Sprintln("[Mysql] DBIndex =%v DBName=%v Ip=%s Port=%v Exist", dbIndex, db_name, ip, port)
		return errors.New(errStr)
	}

	dsn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s?charset=%s", user, password, "tcp", ip, port, db_name, charset)
	name := db_name
	mysqlConn := newMysqlConn(name)
	if err := mysqlConn.connect(dsn); err != nil {
		return err
	}

	ELog.Infof("[Mysql] DbIndex=%v DBName=%v Connect Ip=%v Port=%v  Success", dbIndex, db_name, ip, port)
	d.conns[dbIndex] = mysqlConn
	return nil
}

// bilibili must set 10 database 10 tables
// database_0  table_00  table_10 table_90
// database_1  table_01  table_11 table_91
// ...
func (d *DBModule) HashDBIndex(uid uint64) uint64 {
	ELog.DebugAf("[DBModule] UID=%v Hash DBIndex=%v", uid, uid%d.connMaxCount)
	return uid % d.connMaxCount
}

func (d *DBModule) HashTableIndex(uid uint64) uint64 {
	dbIndex := d.HashDBIndex(uid)
	db_table_index := uid % d.dbTableMaxCount
	ELog.DebugAf("[DBModule] UID=%v Hash TableIndex=%v", uid, db_table_index*10+dbIndex)
	return db_table_index*10 + dbIndex
}

func (d *DBModule) GetTableNameByUID(table_name string, uid uint64) string {
	table_index := d.HashTableIndex(uid)
	return fmt.Sprintf("%v_%02d", table_name, table_index)
}

//async function
func (d *DBModule) AddCommand(uid uint64, command IMysqlCommand) bool {
	if command == nil {
		ELog.ErrorAf("[DBModule] Mysql UId=%v AddCommand Command Is Nil", uid)
		return false
	}

	dbIndex := d.HashDBIndex(uid)
	conn, ok := d.conns[dbIndex]
	if !ok {
		ELog.ErrorAf("[DBModule] Mysql UId=%v DBIndex=%v Group Is Not Exist", uid, dbIndex)
		return false
	}

	conn.AddComand(command)
	return true
}

func (d *DBModule) GetMysqlConn(uid uint64) IMysqlConn {
	dbIndex := d.HashDBIndex(uid)
	conn, ok := d.conns[dbIndex]
	if !ok {
		ELog.ErrorAf("Mysql  UId=%v DBIndex=%v  Is Not Exist", uid, dbIndex)
		return nil
	}

	return conn
}

func (d *DBModule) AddExecutedCommand(command IMysqlCommand) {
	d.executedQueue <- command
}

func (d *DBModule) Run(loopCount int) bool {
	for i := 0; i < loopCount; i++ {
		select {
		case cmd, ok := <-d.executedQueue:
			if !ok {
				return false
			}

			cmd.OnExecuted()
			return true
		default:
			return false
		}
	}
	ELog.ErrorA("[DBModule] Run Error")
	return false
}

var GDBModule *DBModule

func init() {
	GDBModule = &DBModule{
		conns:         make(map[uint64]IMysqlConn),
		executedQueue: make(chan IMysqlCommand, DbExecutedChanSize),
	}
}
