package edb

import (
	"database/sql"
	"reflect"
)

type MysqlRecordSet struct {
	recordSet    []map[string]string
	affectedRows int64
	insertId     int64
}

func NewMysqlRecordSet(sqlRows *sql.Rows, affectedRows int64, insertId int64) *MysqlRecordSet {
	result := GMysqlRecordSetPool.Get().(*MysqlRecordSet)
	result.affectedRows = affectedRows
	result.insertId = insertId
	result.recordSet = make([]map[string]string, 0)

	result.build(sqlRows)
	return result
}

func (m *MysqlRecordSet) build(sqlRows *sql.Rows) {
	if sqlRows == nil {
		return
	}

	defer func() {
		sqlRows.Close()
		sqlRows = nil
	}()

	columns, _ := sqlRows.Columns()
	cache := make([]interface{}, len(columns))
	values := make([]sql.RawBytes, len(columns))
	for index, _ := range cache {
		cache[index] = &values[index]
	}

	for sqlRows.Next() {
		_ = sqlRows.Scan(cache...)
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
