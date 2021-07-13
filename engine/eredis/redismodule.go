package eredis

import (
	"github.com/go-redis/redis"
)

//https://studygolang.com/articles/25522?fr=sidebar

var GRedisClient *redis.ClusterClient

func ConnectRedis(addrs []string, password string) (*redis.ClusterClient, error) {
	redisClient := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    addrs,    // use default Addr
		Password: password, // no password set
	})

	return redisClient, nil
}
