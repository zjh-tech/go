package eredis

import (
	"errors"
	"projects/go-engine/elog"
	"projects/util"
	"strings"

	redis "github.com/alphazero/Go-Redis"
)

type RedisConnSpec struct {
	Name     string
	Host     string
	Port     int
	Password string
}

type RedisModule struct {
	connMaxCount uint64
	conns        map[uint64]*RedisConn
}

func (r *RedisModule) Init(connMaxCount uint64, connSpecs []*RedisConnSpec) error {
	r.connMaxCount = connMaxCount

	if r.connMaxCount == 0 {
		return errors.New("[Redis][RedisModule] No Connect")
	}

	if r.connMaxCount != uint64(len(connSpecs)) {
		return errors.New("[Redis][RedisModule] No Match")
	}

	for i := uint64(0); i < r.connMaxCount; i++ {
		redisNameSlices := strings.Split(connSpecs[i].Name, "_")
		if len(redisNameSlices) != 2 {
			return errors.New("[Redis][RedisModule] Index Error")
		}

		redisIndex, _ := util.Str2Uint64(redisNameSlices[1])

		if redisIndex >= connMaxCount {
			return errors.New("[Redis][RedisModule] Redis Index Error")
		}

		if _, ok := r.conns[redisIndex]; ok {
			return errors.New("[Redis][RedisModule] Redis Index Repeat Error")
		}

		redisConn := NewRedisConn()
		if err := redisConn.Connect(connSpecs[i]); err != nil {
			elog.ErrorAf("[Redis][RedisModule] Connect Error=%v", err)
			return err
		}
		r.conns[redisIndex] = redisConn
	}

	return nil
}

func (r *RedisModule) GetRedisClient(uid uint64) redis.Client {
	index := uid % r.connMaxCount
	//elog.InfoAf("[Redis][RedisModule] UId=%v Index=%v RedisName=redis_%v", uid, index, index)
	if conn, ok := r.conns[index]; ok {
		return conn.GetRedisClient()
	}
	return nil
}

var GRedisModule *RedisModule

func init() {
	GRedisModule = &RedisModule{
		connMaxCount: 0,
		conns:        make(map[uint64]*RedisConn),
	}
}
