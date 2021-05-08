package frame

import "fmt"

func GetRedisKey(key string, uid uint64, sub_key string) string {
	return fmt.Sprintf("%v:[%v]:%v", key, uid, sub_key)
}
