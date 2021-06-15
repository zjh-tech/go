package edb

import (
	"database/sql"
	"reflect"
)

type MysqlRecordSet struct {
	sqlRows      *sql.Rows
	recordSet    []map[string]string
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
		m.recordSet = append(m.recordSet, item)
	}
}

func (m *MysqlRecordSet) GetRecordSet() []map[string]string {
	return m.recordSet
}

func (m *MysqlRecordSet) GetAffectRows() int64 {
	return m.affectedRows
}

func (m *MysqlRecordSet) GetInsertId() int64 {
	return m.insertId
}
