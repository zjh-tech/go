package edb

import (
	"errors"
	"fmt"
	"strings"

	"github.com/zjh-tech/go-frame/base/convert"
)

type DBModule struct {
	db_table_max_count uint64
	db_conn_max_count  uint64
	db_conn_specs      []*DBConnSpec
	executed_queus     chan IMysqlCommand
	conns              map[uint64]IMysqlConn
}

func (d *DBModule) Init(db_conn_max_count uint64, db_table_max_count uint64, db_conn_specs []*DBConnSpec) error {
	d.db_conn_max_count = db_conn_max_count
	d.db_table_max_count = db_table_max_count
	d.db_conn_specs = db_conn_specs

	if d.db_conn_max_count == 0 {
		return errors.New("[DBModule] No Mysql Connect")
	}

	if d.db_conn_max_count != uint64(len(d.db_conn_specs)) {
		return errors.New("[DBModule] Mysql No Match")
	}

	for i := uint64(0); i < d.db_conn_max_count; i++ {
		dbNameSlices := strings.Split(d.db_conn_specs[i].Name, "_")
		if len(dbNameSlices) != 2 {
			return errors.New("[DBModule] Mysql Index Error")
		}

		dbIndex, _ := convert.Str2Uint64(dbNameSlices[1])
		connErr := d.connect(dbIndex,
			d.db_conn_specs[i].Name, d.db_conn_specs[i].Ip,
			d.db_conn_specs[i].Port, d.db_conn_specs[i].User,
			d.db_conn_specs[i].Password, d.db_conn_specs[i].Charset)

		if connErr != nil {
			ELog.Errorf("[DBModule] Connect Mysql DBName=%v DBIp=%v,DBPort=%v, Error = %v",
				d.db_conn_specs[i].Name,
				d.db_conn_specs[i].Ip,
				d.db_conn_specs[i].Port,
				connErr)
			return connErr
		}
	}

	return nil
}

func (d *DBModule) UnInit() {
	ELog.InfoA("[DB] Stop")
}

func (d *DBModule) connect(db_index uint64, db_name string, ip string, port uint32, user string, password string, charset string) error {
	if _, ok := d.conns[db_index]; ok {
		errStr := fmt.Sprintln("[Mysql] DBIndx =%v DBName=%v Ip=%v Port=%v Exist", db_index, db_name, ip, port)
		return errors.New(errStr)
	}

	dsn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s?charset=%s", user, password, "tcp", ip, port, db_name, charset)
	name := db_name
	mysqlConn := new_mysql_conn(name)
	if err := mysqlConn.connect(dsn); err != nil {
		return err
	}

	ELog.Infof("[Mysql] DbIndex=%v DBName=%v Connect Ip=%v Port=%v  Success", db_index, db_name, ip, port)
	d.conns[db_index] = mysqlConn
	return nil
}

// bilibili must set 10 database 10 tables
// database_0  table_00  table_10 table_90
// database_1  table_01  table_11 table_91
// ...
func (d *DBModule) HashDBIndex(uid uint64) uint64 {
	ELog.DebugAf("[DBModule] UID=%v Hash DBIndex=%v", uid, uid%d.db_conn_max_count)
	return uid % d.db_conn_max_count
}

func (d *DBModule) HashTableIndex(uid uint64) uint64 {
	db_index := d.HashDBIndex(uid)
	db_table_index := uid % d.db_table_max_count
	ELog.DebugAf("[DBModule] UID=%v Hash TableIndex=%v", uid, db_table_index*10+db_index)
	return db_table_index*10 + db_index
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

	db_index := d.HashDBIndex(uid)
	conn, ok := d.conns[db_index]
	if !ok {
		ELog.ErrorAf("[DBModule] Mysql UId=%v DBIndex=%v Group Is Not Exist", uid, db_index)
		return false
	}

	conn.AddComand(command)
	return true
}

func (d *DBModule) GetMysqlConn(uid uint64) IMysqlConn {
	db_index := d.HashDBIndex(uid)
	conn, ok := d.conns[db_index]
	if !ok {
		ELog.ErrorAf("Mysql  UId=%v DBIndex=%v  Is Not Exist", uid, db_index)
		return nil
	}

	return conn
}

func (d *DBModule) AddExecutedCommand(command IMysqlCommand) {
	d.executed_queus <- command
}

func (d *DBModule) Run(loop_count int) bool {
	for i := 0; i < loop_count; i++ {
		select {
		case cmd, ok := <-d.executed_queus:
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
		conns:          make(map[uint64]IMysqlConn),
		executed_queus: make(chan IMysqlCommand, DB_EXECUTED_QUEUE_CHAN_SIZE),
	}
}
