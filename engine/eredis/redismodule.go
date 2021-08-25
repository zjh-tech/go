package eredis

import (
	"errors"

	"github.com/go-redis/redis"
)

//https://studygolang.com/articles/25522?fr=sidebar

var (
	_ redis.Cmdable = (*redis.Client)(nil)
	_ redis.Cmdable = (*redis.ClusterClient)(nil)
)

var GRedisCmd redis.Cmdable

func ConnectRedisCluster(addrs []string, password string) (*redis.ClusterClient, error) {
	if (addrs == nil) || (len(addrs) == 0) {
		return nil, errors.New("RedisAddrs Is Empty")
	}

	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    addrs,    // use default Addr
		Password: password, // no password set
	})

	if _, err := client.Ping().Result(); err != nil {
		return nil, err
	}

	return client, nil
}

func ConnectRedis(addr string, password string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,     // use default Addr
		Password: password, // no password set
		DB:       0,        // use default DB
	})

	if _, err := client.Ping().Result(); err != nil {
		return nil, err
	}
	return client, nil
}
