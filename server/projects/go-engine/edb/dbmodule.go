package edb

import (
	"errors"
	"fmt"
	"projects/go-engine/elog"
	"projects/util"
	"strings"
)

const DB_EXECUTED_QUEUE_CHAN_SIZE = 1024 * 10 * 10

type DBConnSpec struct {
	Ip       string
	Port     uint32
	User     string
	Password string
	Name     string
	Charset  string
}

type DBModule struct {
	dbTableMaxCount uint64
	dbConnMaxCount  uint64
	DBConnSpecs     []*DBConnSpec
	conns           map[uint64]IMysqlConn //DBIndex -IMysqlConn
	executedQueus   chan IMysqlCommand
}

func (d *DBModule) Init(dbConnMaxCount uint64, tableMaxCount uint64, connSpecs []*DBConnSpec) error {
	d.dbConnMaxCount = dbConnMaxCount
	d.dbTableMaxCount = tableMaxCount
	d.DBConnSpecs = connSpecs

	if d.dbConnMaxCount == 0 {
		return errors.New("[DBModule] No Mysql Connect")
	}

	if d.dbConnMaxCount != uint64(len(d.DBConnSpecs)) {
		return errors.New("[DBModule] Mysql No Match")
	}

	for i := uint64(0); i < d.dbConnMaxCount; i++ {
		dbNameSlices := strings.Split(d.DBConnSpecs[i].Name, "_")
		if len(dbNameSlices) != 2 {
			return errors.New("[DBModule] Mysql Index Error")
		}

		dbIndex, _ := util.Str2Uint64(dbNameSlices[1])
		connErr := d.Connect(dbIndex,
			d.DBConnSpecs[i].Name, d.DBConnSpecs[i].Ip,
			d.DBConnSpecs[i].Port, d.DBConnSpecs[i].User,
			d.DBConnSpecs[i].Password, d.DBConnSpecs[i].Charset)

		if connErr != nil {
			elog.Errorf("[DBModule] Connect Mysql DBName=%v DBIp=%v,DBPort=%v, Error = %v",
				d.DBConnSpecs[i].Name,
				d.DBConnSpecs[i].Ip,
				d.DBConnSpecs[i].Port,
				connErr)
			return connErr
		}
	}

	return nil
}

func (d *DBModule) UnInit() {
	elog.InfoA("[DB] Stop")
}

func (d *DBModule) Connect(dbIndex uint64, dbName string, ip string, port uint32, user string, password string, charset string) error {
	if _, ok := d.conns[dbIndex]; ok {
		errStr := fmt.Sprintln("[Mysql] DBIndx =%v DBName=%v Ip=%v Port=%v Exist", dbIndex, dbName, ip, port)
		return errors.New(errStr)
	}

	dsn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s?charset=%s", user, password, "tcp", ip, port, dbName, charset)
	name := dbName
	mysqlConn := NewMysqlConn(name)
	if err := mysqlConn.Connect(dsn); err != nil {
		return err
	}

	elog.Infof("[Mysql] DbIndex=%v DBName=%v Connect Ip=%v Port=%v  Success", dbIndex, dbName, ip, port)
	d.conns[dbIndex] = mysqlConn
	return nil
}

// bilibili must set 10 database 10 tables
// database_0  table_00  table_10 table_90
// database_1  table_01  table_11 table_91
// ...
func (d *DBModule) HashDBIndex(uid uint64) uint64 {
	elog.InfoAf("[DBModule] UID=%v Hash DBIndex=%v", uid, uid%d.dbConnMaxCount)
	return uid % d.dbConnMaxCount
}

func (d *DBModule) HashTableIndex(uid uint64) uint64 {
	dbIndex := d.HashDBIndex(uid)
	dbTableIndex := uid % d.dbTableMaxCount
	elog.InfoAf("[DBModule] UID=%v Hash TableIndex=%v", uid, dbTableIndex*10+dbIndex)
	return dbTableIndex*10 + dbIndex
}

func (d *DBModule) GetTableNameByUID(tableName string, uid uint64) string {
	tableIndex := d.HashTableIndex(uid)
	return fmt.Sprintf("%v_%02d", tableName, tableIndex)
}

//async function
func (d *DBModule) AddCommand(uid uint64, command IMysqlCommand) bool {
	if command == nil {
		elog.ErrorAf("[DBModule] Mysql UId=%v AddCommand Command Is Nil", uid)
		return false
	}

	dbIndex := d.HashDBIndex(uid)
	conn, ok := d.conns[dbIndex]
	if !ok {
		elog.ErrorAf("[DBModule] Mysql UId=%v DBIndex=%v Group Is Not Exist", uid, dbIndex)
		return false
	}

	conn.AddComand(command)
	return true
}

func (d *DBModule) GetMysqlConn(uid uint64) IMysqlConn {
	dbIndex := d.HashDBIndex(uid)
	conn, ok := d.conns[dbIndex]
	if !ok {
		elog.ErrorAf("Mysql  UId=%v DBIndex=%v  Is Not Exist", uid, dbIndex)
		return nil
	}

	return conn
}

func (d *DBModule) AddExecutedCommand(command IMysqlCommand) {
	d.executedQueus <- command
}

func (d *DBModule) Run(loop_count int) bool {
	for i := 0; i < loop_count; i++ {
		select {
		case cmd, ok := <-d.executedQueus:
			if !ok {
				return false
			}

			cmd.OnExecuted()
			return true
		default:
			return false
		}
	}
	elog.ErrorA("[DBModule] Run Error")
	return false
}

var GDBModule *DBModule

func init() {
	GDBModule = &DBModule{
		conns:         make(map[uint64]IMysqlConn),
		executedQueus: make(chan IMysqlCommand, DB_EXECUTED_QUEUE_CHAN_SIZE),
	}
}
