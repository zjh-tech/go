package eredis

import (
	"errors"
	"projects/go-engine/elog"

	redis "github.com/alphazero/Go-Redis"
)

type RedisConn struct {
	redisClient redis.Client
	connSpec    *RedisConnSpec
}

func NewRedisConn() *RedisConn {
	return &RedisConn{}
}

func (r *RedisConn) Connect(connSpec *RedisConnSpec) error {
	if connSpec == nil {
		return errors.New("[Redis][RedisConn] RedisConnSpec is Nil")
	}

	spec := redis.DefaultSpec()
	spec.Host(connSpec.Host)
	spec.Port(connSpec.Port)
	spec.Password(connSpec.Password)
	client, connErr := redis.NewSynchClientWithSpec(spec)
	if connErr != nil {
		return connErr
	}

	r.connSpec = connSpec
	r.redisClient = client

	if pingErr := r.redisClient.Ping(); pingErr != nil {
		return pingErr
	}

	elog.InfoAf("[Redis][RedisConn] Connect Host=%v,Post=%v Success", connSpec.Host, connSpec.Port)
	return nil
}

func (r *RedisConn) GetRedisClient() redis.Client {
	return r.redisClient
}
