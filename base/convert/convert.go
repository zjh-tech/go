package convert

import (
	"fmt"
	"reflect"
	"strconv"
)

func StrToUint64(str string) (uint64, error) {
	if n, err := strconv.ParseUint(str, 10, 64); err != nil {
		return 0, err
	} else {
		return n, nil
	}
}

func StrToInt64(str string) (int64, error) {
	if n, err := strconv.ParseInt(str, 10, 64); err != nil {
		return 0, err
	} else {
		return n, nil
	}
}

func StrToUint32(str string) (uint32, error) {
	if n, err := strconv.ParseUint(str, 10, 32); err != nil {
		return 0, err
	} else {
		return uint32(n), nil
	}
}

func StrToInt32(str string) (int32, error) {
	if n, err := strconv.ParseInt(str, 10, 32); err != nil {
		return 0, err
	} else {
		return int32(n), nil
	}
}

func StrToInt(str string) (int, error) {
	if n, err := strconv.Atoi(str); err != nil {
		return 0, err
	} else {
		return n, nil
	}
}

func StrToUint16(str string) (uint16, error) {
	if n, err := strconv.ParseUint(str, 10, 16); err != nil {
		return 0, err
	} else {
		return uint16(n), nil
	}
}

func StrToInt16(str string) (int16, error) {
	if n, err := strconv.ParseInt(str, 10, 16); err != nil {
		return 0, err
	} else {
		return int16(n), nil
	}
}

func StrToUint8(str string) (uint8, error) {
	if n, err := strconv.ParseUint(str, 10, 8); err != nil {
		return 0, err
	} else {
		return uint8(n), nil
	}
}

func StrToInt8(str string) (int8, error) {
	if n, err := strconv.ParseInt(str, 10, 8); err != nil {
		return 0, err
	} else {
		return int8(n), nil
	}
}

func Str2Bool(str string) bool {
	value, _ := StrToUint32(str)
	return value == 0
}

func StrToFloat64(str string) (float64, error) {
	if n, err := strconv.ParseFloat(str, 64); err != nil {
		return 0, err
	} else {
		return n, nil
	}
}

//---------------------------------
func IntToStr(n int) string {
	return strconv.Itoa(n)
}

func Int8ToStr(n int8) string {
	return strconv.FormatInt(int64(n), 10)
}

func UInt8ToStr(n uint8) string {
	return strconv.FormatUint(uint64(n), 10)
}

func Int16ToStr(n int16) string {
	return strconv.FormatInt(int64(n), 10)
}

func UInt16ToStr(n uint16) string {
	return strconv.FormatUint(uint64(n), 10)
}

func Int32ToStr(n int32) string {
	return strconv.FormatInt(int64(n), 10)
}

func Uint32ToStr(n uint32) string {
	return strconv.FormatUint(uint64(n), 10)
}

func Int64ToStr(n int64) string {
	return strconv.FormatInt(n, 10)
}

func Uint64ToStr(n uint64) string {
	return strconv.FormatUint(n, 10)
}

func ToString(n interface{}) string {
	switch v := n.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case int:
		return strconv.Itoa(v)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	}

	//type MyType int64
	rv := reflect.ValueOf(n)
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
	return fmt.Sprintf("%v", n)
}
