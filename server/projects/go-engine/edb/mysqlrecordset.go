package edb

import (
	"database/sql"
	"reflect"
)

const DB_DEFAULT_AFFECTED_ROWS int64 = 0
const DB_DEFAULT_INSERT_ID int64 = 0

type MysqlRecordSet struct {
	sqlRows      *sql.Rows
	recordset    []map[string]string
	affectedRows int64
	insertId     int64
}

func NewMysqlRecordSet(sqlRows *sql.Rows, affectedRows int64, insertId int64) *MysqlRecordSet {
	recordSet := &MysqlRecordSet{
		sqlRows:      sqlRows,
		affectedRows: affectedRows,
		insertId:     insertId,
	}
	recordSet.build()
	return recordSet
}

func (m *MysqlRecordSet) build() {
	if m.sqlRows == nil {
		return
	}

	defer func() {
		m.sqlRows.Close()
		m.sqlRows = nil
	}()

	columns, _ := m.sqlRows.Columns()
	cache := make([]interface{}, len(columns))
	values := make([]sql.RawBytes, len(columns))
	for index, _ := range cache {
		cache[index] = &values[index]
	}

	for m.sqlRows.Next() {
		_ = m.sqlRows.Scan(cache...)
		item := make(map[string]string)
		for k, v := range cache {
			content := reflect.ValueOf(v).Interface().(*sql.RawBytes)
			item[columns[k]] = string(*content)
		}
		m.recordset = append(m.recordset, item)
	}
}

func (m *MysqlRecordSet) GetRecordSet() []map[string]string {
	return m.recordset
}

func (m *MysqlRecordSet) GetAffectRows() int64 {
	return m.affectedRows
}

func (m *MysqlRecordSet) GetInsertId() int64 {
	return m.insertId
}
