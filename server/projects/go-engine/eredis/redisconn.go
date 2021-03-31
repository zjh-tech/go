package eredis

import (
	"errors"
	"projects/go-engine/elog"

	redis "github.com/alphazero/Go-Redis"
)

type RedisConn struct {
	redis_client redis.Client
	conn_spec    *RedisConnSpec
}

func new_redis_conn() *RedisConn {
	return &RedisConn{}
}

func (r *RedisConn) connect(conn_spec *RedisConnSpec) error {
	if conn_spec == nil {
		return errors.New("[Redis][RedisConn] RedisConnSpec is Nil")
	}

	spec := redis.DefaultSpec()
	spec.Host(conn_spec.Host)
	spec.Port(conn_spec.Port)
	spec.Password(conn_spec.Password)
	client, connErr := redis.NewSynchClientWithSpec(spec)
	if connErr != nil {
		return connErr
	}

	r.conn_spec = conn_spec
	r.redis_client = client

	if pingErr := r.redis_client.Ping(); pingErr != nil {
		return pingErr
	}

	elog.InfoAf("[Redis][RedisConn] Connect Host=%v,Post=%v Success", conn_spec.Host, conn_spec.Port)
	return nil
}

func (r *RedisConn) get_redis_client() redis.Client {
	return r.redis_client
}
