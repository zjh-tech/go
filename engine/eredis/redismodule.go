package eredis

import (
	"errors"

	"github.com/go-redis/redis"
)

//https://studygolang.com/articles/25522?fr=sidebar

var GRedisClient *redis.ClusterClient

func ConnectRedis(addrs []string, password string) (*redis.ClusterClient, error) {
	if (addrs == nil) || (len(addrs) == 0) {
		return nil, errors.New("RedisAddrs Is Empty")
	}

	redisClient := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    addrs,    // use default Addr
		Password: password, // no password set
	})

	return redisClient, nil
}
