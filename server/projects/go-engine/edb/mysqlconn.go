package edb

import (
	"database/sql"
	"projects/go-engine/elog"

	_ "github.com/go-sql-driver/mysql"
)

const DB_QUEUE_CHAN_SIZE = 1024 * 10 * 10

type MysqlConn struct {
	name string

	dsn string

	sqlDB *sql.DB

	cmdQueue chan IMysqlCommand

	exitChan chan struct{}

	sqlTx *sql.Tx
}

func NewMysqlConn(name string) *MysqlConn {
	conn := &MysqlConn{
		name:     name,
		cmdQueue: make(chan IMysqlCommand, DB_QUEUE_CHAN_SIZE),
		exitChan: make(chan struct{}),
		sqlTx:    nil,
		sqlDB:    nil,
	}

	return conn
}

func (m *MysqlConn) Connect(dsn string) error {
	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	err = sqlDB.Ping()
	if err != nil {
		return err
	}

	m.sqlDB = sqlDB
	m.dsn = dsn

	m.sqlDB.SetMaxOpenConns(1)
	m.sqlDB.SetConnMaxLifetime(0)
	m.run()
	return nil
}

func (m *MysqlConn) AddComand(command IMysqlCommand) {
	m.cmdQueue <- command
}

func (m *MysqlConn) EscapeString(sql string) string {
	return sql
}

func (m *MysqlConn) run() {
	go func() {
		for {
			select {
			case cmd, ok := <-m.cmdQueue:
				if !ok {
					return
				}

				cmd.OnExecuteSql(m)
				GDBModule.AddExecutedCommand(cmd)
			case <-m.exitChan:
				elog.InfoAf("Name %v MysqlConn Exit", m.name)
				return
			}
		}
	}()
}

func (m *MysqlConn) QueryWithResult(sql string) (IMysqlRecordSet, error) {
	rows, err := m.sqlDB.Query(sql)
	if err != nil {
		elog.ErrorAf("[Mysql] QueryWithResult Sql=%v, Error=%v", sql, err)
		return nil, err
	}

	elog.InfoAf("[Mysql] QueryWithResult Sql=%v Success", sql)
	return NewMysqlRecordSet(rows, DB_DEFAULT_AFFECTED_ROWS, DB_DEFAULT_INSERT_ID), nil
}

func (m *MysqlConn) QueryWithoutResult(sql string) (IMysqlRecordSet, error) {
	res, err := m.sqlDB.Exec(sql)
	if err != nil {
		elog.InfoAf("[Mysql] QueryWithoutResult Sql=%v, Error=%v", sql, err)
		return nil, err
	}

	affectedRows, err1 := res.RowsAffected()
	if err1 != nil {
		elog.InfoAf("[Mysql] QueryWithoutResult Sql=%v,RowsAffected Error=%v", sql, err1)
		return nil, err1
	}

	insertId, err2 := res.LastInsertId()
	if err2 != nil {
		elog.InfoAf("[Mysql] QueryWithoutResult Sql=%v,LastInsertId Error=%v", sql, err2)
		return nil, err2
	}

	elog.InfoAf("[Mysql] QueryWithoutResult Sql=%v Success", sql)

	return NewMysqlRecordSet(nil, affectedRows, insertId), nil
}

func (m *MysqlConn) BeginTransact() {
	if m.sqlTx != nil {
		elog.ErrorAf("[MysqlConn] Begin SqlTx Not Nil")
		m.sqlTx = nil
	}

	var err error
	m.sqlTx, err = m.sqlDB.Begin()
	if err != nil {
		elog.InfoAf("[MysqlConn] Begin Error=%v", err)
	}
}

func (m *MysqlConn) CommitTransact() {
	if m.sqlTx == nil {
		return
	}

	err := m.sqlTx.Commit()
	m.sqlTx = nil
	if err != nil {
		elog.InfoAf("[MysqlConn] Commit Error=%v", err)
	}
}

func (m *MysqlConn) RollbackTransact() {
	if m.sqlTx == nil {
		return
	}

	err := m.sqlTx.Rollback()
	m.sqlTx = nil
	if err != nil {
		elog.InfoAf("[MysqlConn] Rollback Error=%v", err)
	}
}
