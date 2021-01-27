package frame

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"unsafe"
)

// add `` to all table field.
// just usw bytes.Buffer to string to build sql statement.

func putString(buf *bytes.Buffer, field string, equal bool) {
	buf.WriteString("`")
	buf.WriteString(field)
	if equal {
		buf.WriteString("`=")
	} else {
		buf.WriteString("`")
	}
}

// 和上面不一样的引用方式哦。
func putValue(buf *bytes.Buffer, value string) {
	buf.WriteString("'")
	buf.WriteString(value)
	buf.WriteString("'")
}

func asSQLString(src interface{}) string {
	switch v := src.(type) {
	case string:
		var buf bytes.Buffer
		putValue(&buf, v)
		return buf.String()

	case []byte:
		var buf bytes.Buffer
		strValue := (*string)(unsafe.Pointer(&v)) // 这种效率更高
		putValue(&buf, *strValue)
		return buf.String()
	}

	rv := reflect.ValueOf(src)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10)
	case reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 64)
	case reflect.Float32:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 32)
	case reflect.Bool:
		return strconv.FormatBool(rv.Bool())
	}
	return fmt.Sprintf("%v", src)
}

type DBField struct {
	Fileds []string
}

func NewDBField() *DBField {
	return &DBField{
		Fileds: make([]string, 0, 0),
	}
}

func (d *DBField) Add(filed string) {
	d.Fileds = append(d.Fileds, filed)
}

type DBFieldPair struct {
	FieldMap map[string]string
}

func NewDBFieldPair() *DBFieldPair {
	return &DBFieldPair{
		FieldMap: make(map[string]string),
	}
}

func (d *DBFieldPair) Add(key string, value interface{}) {
	str := asSQLString(value)
	d.FieldMap[key] = str
}

func GetSelectSQL(tbl string, selects *DBField, wheres *DBFieldPair) string {
	var buf bytes.Buffer
	buf.WriteString("SELECT ")

	if selects == nil {
		buf.WriteString(" * ")
	} else {
		for index, name := range selects.Fileds {
			if index != 0 {
				buf.WriteString(" , ")
			}
			putString(&buf, name, false)
		}
	}

	buf.WriteString(" FROM ")
	buf.WriteString(tbl)

	if wheres != nil {
		buf.WriteString(" WHERE ")
		firstflag := true
		for k, v := range wheres.FieldMap {
			if !firstflag {
				buf.WriteString(" AND ")
			}
			firstflag = false
			putString(&buf, k, true)
			buf.WriteString(v)
		}
	}
	buf.WriteString(";")

	return buf.String()
}

func GetInsertSQL(tbl string, inserts *DBFieldPair) string {
	var buf bytes.Buffer

	buf.WriteString("INSERT INTO ")
	buf.WriteString(tbl)
	buf.WriteString("(")

	if inserts != nil {
		firstflag := true
		var values []string
		for k, v := range inserts.FieldMap {
			if !firstflag {
				buf.WriteString(" , ")
			}
			firstflag = false
			putString(&buf, k, false)
			values = append(values, v)
		}
		buf.WriteString(" ) VALUES(")

		firstflag = true
		for i := 0; i < len(values); i++ {
			if !firstflag {
				buf.WriteString(" , ")
			}
			firstflag = false

			buf.WriteString(values[i])
		}
	}
	buf.WriteString(");")

	return buf.String()
}

func GetUpdateSQL(tbl string, updates *DBFieldPair, wheres *DBFieldPair) string {
	var buf bytes.Buffer
	buf.WriteString("UPDATE ")
	buf.WriteString(tbl)
	buf.WriteString(" SET ")

	if updates != nil {
		firstflag := true
		for k, v := range updates.FieldMap {
			if !firstflag {
				buf.WriteString(" , ")
			}
			firstflag = false

			putString(&buf, k, true)
			buf.WriteString(v)
		}
	}

	if wheres != nil {
		buf.WriteString(" WHERE ")
		firstflag := true
		for k, v := range wheres.FieldMap {
			if !firstflag {
				buf.WriteString(" AND ")
			}
			firstflag = false

			putString(&buf, k, true)
			buf.WriteString(v)
		}
	}
	buf.WriteString(";")
	return buf.String()
}

func GetDeleteSQL(tbl string, wheres *DBFieldPair) string {
	var buf bytes.Buffer
	buf.WriteString("DELETE FROM ")
	buf.WriteString(tbl)

	if wheres != nil {
		buf.WriteString(" WHERE ")
		firstflag := true
		for k, v := range wheres.FieldMap {
			if !firstflag {
				buf.WriteString(" AND ")
			}
			firstflag = false

			putString(&buf, k, true)
			buf.WriteString(v)
		}
	}
	buf.WriteString(";")
	return buf.String()
}

// find key from list
func find(key string, l []string) bool {
	for i := 0; i < len(l); i++ {
		if l[i] == key {
			return true
		}
	}
	return false
}
func GetInsertOrUpdateSQL(tbl string, updates *DBFieldPair, keys *DBField) string {
	var buf bytes.Buffer
	buf.WriteString("INSERT INTO ")
	buf.WriteString(tbl)
	buf.WriteString("( ")

	// upate list enum.
	if updates != nil {
		firstflag := true
		var values []string
		for k, v := range updates.FieldMap { // key
			if !firstflag {
				buf.WriteString(" , ")
			}
			firstflag = false

			putString(&buf, k, false)
			values = append(values, v)
		}

		buf.WriteString(" ) VALUES ( ") // value
		firstflag = true
		for i := 0; i < len(values); i++ {
			if !firstflag {
				buf.WriteString(", ")
			}
			firstflag = false
			buf.WriteString(values[i])
		}
		buf.WriteString("  ) ON DUPLICATE KEY UPDATE ")
	}

	if keys != nil { // exclude key value.
		firstflag := true
		for k, v := range updates.FieldMap { // key
			if find(k, keys.Fileds) {
				continue
			}

			if !firstflag {
				buf.WriteString(", ")
			}
			firstflag = false

			putString(&buf, k, true)
			buf.WriteString(v)
		}
	}
	return buf.String()
}

func BuildSelectSQL(tbl string, selects []string, wheres map[string]interface{}) string {
	var selectfields *DBField
	if selects != nil {
		selectfields = NewDBField()
		for _, name := range selects {
			selectfields.Add(name)
		}
	} else {
		selectfields = nil
	}

	var wheresmap *DBFieldPair
	if wheres != nil && len(wheres) > 0 {
		wheresmap = NewDBFieldPair()
		for k, v := range wheres {
			wheresmap.Add(k, v)
		}
	} else {
		wheresmap = nil
	}
	return GetSelectSQL(tbl, selectfields, wheresmap)
}

func BuildInsertSQL(tbl string, inserts map[string]interface{}) string {
	insertmap := NewDBFieldPair()
	for k, v := range inserts {
		insertmap.Add(k, v)
	}
	return GetInsertSQL(tbl, insertmap)
}

func BuildUpdateSQL(tbl string, updates map[string]interface{}, wheres map[string]interface{}) string {
	updatemap := NewDBFieldPair()
	for k, v := range updates {
		updatemap.Add(k, v)
	}

	var wheresmap *DBFieldPair
	if wheres != nil && len(wheres) > 0 {
		wheresmap = NewDBFieldPair()
		for k, v := range wheres {
			wheresmap.Add(k, v)
		}
	} else {
		wheresmap = nil
	}
	return GetUpdateSQL(tbl, updatemap, wheresmap)
}

func BuildDeleteSQL(tbl string, wheres map[string]interface{}) string {
	wheremap := NewDBFieldPair()
	if len(wheres) > 0 {
		for k, v := range wheres {
			wheremap.Add(k, v)
		}
	} else {
		wheremap = nil
	}
	return GetDeleteSQL(tbl, wheremap)
}

func BuildInsertOrUpdateSQL(tbl string, updates map[string]interface{}, keys []string) string {
	updatemap := NewDBFieldPair()
	for k, v := range updates {
		updatemap.Add(k, v)
	}

	fields := NewDBField()
	for _, name := range keys {
		fields.Add(name)
	}

	return GetInsertOrUpdateSQL(tbl, updatemap, fields)
}
