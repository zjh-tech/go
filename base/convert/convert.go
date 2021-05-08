package convert

import "strconv"

//convert
func Str2Uint64(str string) (uint64, bool) {
	n, err := strconv.ParseUint(str, 10, 64)
	if err == nil {
		return n, true
	}

	return 0, false
}

func Str2Int64(str string) (int64, bool) {
	n, err := strconv.ParseInt(str, 10, 64)
	if err == nil {
		return n, true
	}

	return 0, false
}

func Str2Uint32(str string) (uint32, bool) {
	n, err := strconv.ParseUint(str, 10, 32)
	if err == nil {
		return uint32(n), true
	}

	return 0, false
}

func Str2Int32(str string) (int32, bool) {
	n, err := strconv.ParseInt(str, 10, 32)
	if err == nil {
		return int32(n), true
	}

	return 0, false
}

func Int642Str(n int64) string {
	return strconv.FormatInt(n, 10)
}

func Uint642Str(n uint64) string {
	return strconv.FormatUint(n, 10)
}

func Str2Int(str string) (int, bool) {
	n, err := strconv.Atoi(str)
	if err == nil {
		return n, true
	}

	return 0, false
}

func Int2Str(n int) string {
	return strconv.Itoa(n)
}

func Str2Bool(str string) bool {
	value, _ := Str2Uint32(str)
	if value == 0 {
		return false
	}

	return true
}

func Str2Float64(str string) (float64, bool) {
	n, err := strconv.ParseFloat(str, 64)
	if err == nil {
		return n, true
	}

	return 0, false
}
