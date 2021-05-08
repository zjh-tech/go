package edb

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type MysqlConn struct {
	name string

	dsn string

	sql_db *sql.DB

	cmd_queue chan IMysqlCommand

	exit_chan chan struct{}

	sql_tx *sql.Tx
}

func new_mysql_conn(name string) *MysqlConn {
	conn := &MysqlConn{
		name:      name,
		cmd_queue: make(chan IMysqlCommand, DB_WAIT_QUEUE_CHAN_SIZE),
		exit_chan: make(chan struct{}),
		sql_tx:    nil,
		sql_db:    nil,
	}

	return conn
}

func (m *MysqlConn) connect(dsn string) error {
	sql_db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	err = sql_db.Ping()
	if err != nil {
		return err
	}

	m.sql_db = sql_db
	m.dsn = dsn

	m.sql_db.SetMaxOpenConns(1)
	m.sql_db.SetConnMaxLifetime(0)
	m.run()

	return nil
}

func (m *MysqlConn) AddComand(command IMysqlCommand) {
	m.cmd_queue <- command
}

func (m *MysqlConn) run() {
	go func() {
		for {
			select {
			case cmd, ok := <-m.cmd_queue:
				if !ok {
					return
				}

				cmd.OnExecuteSql(m)
				GDBModule.AddExecutedCommand(cmd)
			case <-m.exit_chan:
				ELog.InfoAf("Name %v MysqlConn Exit", m.name)
				return
			}
		}
	}()
}

func (m *MysqlConn) FindSqlDb() *sql.DB {
	return m.sql_db
}

func (m *MysqlConn) QueryWithResult(sql string) (IMysqlRecordSet, error) {
	rows, err := m.sql_db.Query(sql)
	if err != nil {
		ELog.ErrorAf("[Mysql] QueryWithResult Sql=%v, Error=%v", sql, err)
		return nil, err
	}

	ELog.InfoAf("[Mysql] QueryWithResult Sql=%v Success", sql)
	return NewMysqlRecordSet(rows, DB_DEFAULT_AFFECTED_ROWS, DB_DEFAULT_INSERT_ID), nil
}

func (m *MysqlConn) QueryWithoutResult(sql string) (IMysqlRecordSet, error) {
	res, err := m.sql_db.Exec(sql)
	if err != nil {
		ELog.InfoAf("[Mysql] QueryWithoutResult Sql=%v, Error=%v", sql, err)
		return nil, err
	}

	affectedRows, err1 := res.RowsAffected()
	if err1 != nil {
		ELog.InfoAf("[Mysql] QueryWithoutResult Sql=%v,RowsAffected Error=%v", sql, err1)
		return nil, err1
	}

	insertId, err2 := res.LastInsertId()
	if err2 != nil {
		ELog.InfoAf("[Mysql] QueryWithoutResult Sql=%v,LastInsertId Error=%v", sql, err2)
		return nil, err2
	}

	ELog.InfoAf("[Mysql] QueryWithoutResult Sql=%v Success", sql)

	return NewMysqlRecordSet(nil, affectedRows, insertId), nil
}

func (m *MysqlConn) BeginTransact() {
	if m.sql_tx != nil {
		ELog.ErrorAf("[MysqlConn] Begin SqlTx Not Nil")
		m.sql_tx = nil
	}

	var err error
	m.sql_tx, err = m.sql_db.Begin()
	if err != nil {
		ELog.InfoAf("[MysqlConn] Begin Error=%v", err)
	}
}

func (m *MysqlConn) CommitTransact() {
	if m.sql_tx == nil {
		return
	}

	err := m.sql_tx.Commit()
	m.sql_tx = nil
	if err != nil {
		ELog.InfoAf("[MysqlConn] Commit Error=%v", err)
	}
}

func (m *MysqlConn) RollbackTransact() {
	if m.sql_tx == nil {
		return
	}

	err := m.sql_tx.Rollback()
	m.sql_tx = nil
	if err != nil {
		ELog.InfoAf("[MysqlConn] Rollback Error=%v", err)
	}
}
