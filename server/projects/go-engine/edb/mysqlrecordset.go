package edb

import (
	"database/sql"
	"reflect"
)

const DB_DEFAULT_AFFECTED_ROWS int64 = 0
const DB_DEFAULT_INSERT_ID int64 = 0

type MysqlRecordSet struct {
	sql_rows      *sql.Rows
	recordset     []map[string]string
	affected_rows int64
	insert_id     int64
}

func NewMysqlRecordSet(sql_rows *sql.Rows, affected_rows int64, insert_id int64) *MysqlRecordSet {
	recordset := &MysqlRecordSet{
		sql_rows:      sql_rows,
		affected_rows: affected_rows,
		insert_id:     insert_id,
	}
	recordset.build()
	return recordset
}

func (m *MysqlRecordSet) build() {
	if m.sql_rows == nil {
		return
	}

	defer func() {
		m.sql_rows.Close()
		m.sql_rows = nil
	}()

	columns, _ := m.sql_rows.Columns()
	cache := make([]interface{}, len(columns))
	values := make([]sql.RawBytes, len(columns))
	for index, _ := range cache {
		cache[index] = &values[index]
	}

	for m.sql_rows.Next() {
		_ = m.sql_rows.Scan(cache...)
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
	return m.affected_rows
}

func (m *MysqlRecordSet) GetInsertId() int64 {
	return m.insert_id
}
