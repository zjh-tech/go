package eredis

import (
	"fmt"

	"github.com/go-redis/redis"
)

//https://studygolang.com/articles/25522?fr=sidebar

var GRedisClient *redis.Client

func ConnectRedis(host string, port int, password string) (*redis.Client, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	redisClient := redis.NewClient(&redis.Options{
		Addr:     addr,     // use default Addr
		Password: password, // no password set
		DB:       0,        // use default DB
	})

	if _, err := redisClient.Ping().Result(); err != nil {
		return nil, err
	}
	return redisClient, nil
}
