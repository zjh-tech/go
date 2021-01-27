package frame

import "fmt"

func GetRedisKey(key string, uid uint64, subKey string) string {
	return fmt.Sprintf("%v:[%v]:%v", key, uid, subKey)
}
