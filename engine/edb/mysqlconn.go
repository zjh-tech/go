package edb

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type MysqlConn struct {
	name string

	dsn string

	sqlDb *sql.DB

	cmdQueue chan IMysqlCommand

	exitChan chan struct{}

	sqlTx *sql.Tx

	dbmodule *DBModule
}

func newMysqlConn(name string, dbmodule *DBModule) *MysqlConn {
	conn := &MysqlConn{
		name:     name,
		cmdQueue: make(chan IMysqlCommand, DbWaitChanSize),
		exitChan: make(chan struct{}),
		sqlTx:    nil,
		sqlDb:    nil,
		dbmodule: dbmodule,
	}

	return conn
}

func (m *MysqlConn) connect(dsn string) error {
	m.dsn = dsn

	if sqlDb, err := sql.Open("mysql", dsn); err != nil {
		return err
	} else {
		m.sqlDb = sqlDb
	}

	if err := m.sqlDb.Ping(); err != nil {
		return err
	}

	m.sqlDb.SetMaxOpenConns(1)
	m.sqlDb.SetConnMaxLifetime(0)
	m.run()

	return nil
}

func (m *MysqlConn) AddCommand(command IMysqlCommand) {
	m.cmdQueue <- command
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
				m.dbmodule.AddExecutedCommand(cmd)
			case <-m.exitChan:
				ELog.InfoAf("Name %v MysqlConn Exit", m.name)
				return
			}
		}
	}()
}

func (m *MysqlConn) FindSqlDb() *sql.DB {
	return m.sqlDb
}

func (m *MysqlConn) QuerySqlOpt(sql string) (IDBResult, error) {
	rows, err := m.sqlDb.Query(sql)
	if err != nil {
		ELog.ErrorAf("[Mysql] QuerySqlOpt Sql=%v, Error=%v", sql, err)
		return nil, err
	}

	ELog.InfoAf("[Mysql] QuerySqlOpt Sql=%v Success", sql)
	return NewMysqlRecordSet(rows, DbDefaultAffectedRows, DbDefaultInsertId), nil
}

func (m *MysqlConn) NonQuerySqlOpt(sql string) (IDBResult, error) {
	res, err := m.sqlDb.Exec(sql)
	if err != nil {
		ELog.InfoAf("[Mysql] NonQuerySqlOpt Sql=%v, Error=%v", sql, err)
		return nil, err
	}

	affectedRows, err1 := res.RowsAffected()
	if err1 != nil {
		ELog.InfoAf("[Mysql] NonQuerySqlOpt Sql=%v,RowsAffected Error=%v", sql, err1)
		return nil, err1
	}

	insertId, err2 := res.LastInsertId()
	if err2 != nil {
		ELog.InfoAf("[Mysql] NonQuerySqlOpt Sql=%v,LastInsertId Error=%v", sql, err2)
		return nil, err2
	}

	ELog.InfoAf("[Mysql] NonQuerySqlOpt Sql=%v Success", sql)

	return NewMysqlRecordSet(nil, affectedRows, insertId), nil
}

func (m *MysqlConn) BeginTransact() {
	if m.sqlTx != nil {
		ELog.ErrorAf("[MysqlConn] Begin SqlTx Not Nil")
		m.sqlTx = nil
	}

	var err error
	m.sqlTx, err = m.sqlDb.Begin()
	if err != nil {
		ELog.InfoAf("[MysqlConn] Begin Error=%v", err)
	}
}

func (m *MysqlConn) CommitTransact() {
	if m.sqlTx == nil {
		return
	}

	err := m.sqlTx.Commit()
	m.sqlTx = nil
	if err != nil {
		ELog.InfoAf("[MysqlConn] Commit Error=%v", err)
	}
}

func (m *MysqlConn) RollbackTransact() {
	if m.sqlTx == nil {
		return
	}

	err := m.sqlTx.Rollback()
	m.sqlTx = nil
	if err != nil {
		ELog.InfoAf("[MysqlConn] Rollback Error=%v", err)
	}
}
