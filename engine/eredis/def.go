package eredis

import "fmt"

type RedisConnSpec struct {
	Name     string
	Host     string
	Port     int
	Password string
}

const RedisMajorVersion = 1
const RedisMinorVersion = 1

type RedisVersion struct {
}

func (r *RedisVersion) GetVersion() string {
	return fmt.Sprintf("Redis Version: %v.%v", RedisMajorVersion, RedisMinorVersion)
}

var GRedisVersion *RedisVersion

func init() {
	GRedisVersion = &RedisVersion{}
}
