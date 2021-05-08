package eredis

import (
	"errors"
	"projects/base/convert"
	"strings"

	redis "github.com/alphazero/Go-Redis"
)

type RedisModule struct {
	conn_max_count uint64
	conns          map[uint64]*RedisConn
}

func new_redis_module() *RedisModule {
	redis_module := &RedisModule{
		conn_max_count: 0,
		conns:          make(map[uint64]*RedisConn),
	}
	return redis_module
}

func (r *RedisModule) Init(conn_max_count uint64, conn_specs []*RedisConnSpec) error {
	r.conn_max_count = conn_max_count

	if r.conn_max_count == 0 {
		return errors.New("[Redis][RedisModule] No Connect")
	}

	if r.conn_max_count != uint64(len(conn_specs)) {
		return errors.New("[Redis][RedisModule] No Match")
	}

	for i := uint64(0); i < r.conn_max_count; i++ {
		redis_name_slices := strings.Split(conn_specs[i].Name, "_")
		if len(redis_name_slices) != 2 {
			return errors.New("[Redis][RedisModule] Index Error")
		}

		redis_index, _ := convert.Str2Uint64(redis_name_slices[1])

		if redis_index >= conn_max_count {
			return errors.New("[Redis][RedisModule] Redis Index Error")
		}

		if _, ok := r.conns[redis_index]; ok {
			return errors.New("[Redis][RedisModule] Redis Index Repeat Error")
		}

		redis_conn := new_redis_conn()
		if err := redis_conn.connect(conn_specs[i]); err != nil {
			ELog.ErrorAf("[Redis][RedisModule] Connect Error=%v", err)
			return err
		}
		r.conns[redis_index] = redis_conn
	}

	return nil
}

func (r *RedisModule) GetRedisClient(uid uint64) redis.Client {
	index := uid % r.conn_max_count
	ELog.InfoAf("[Redis][RedisModule] UId=%v Index=%v RedisName=redis_%v", uid, index, index)
	if conn, ok := r.conns[index]; ok {
		return conn.get_redis_client()
	}
	return nil
}

var GRedisModule *RedisModule

func init() {
	GRedisModule = new_redis_module()
}
