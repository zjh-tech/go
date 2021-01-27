package main

import (
	"projects/frame"
	"projects/go-engine/elog"
	"projects/go-engine/eredis"
	"time"
)

func main() {
	elog.InitLog("./log", 0, nil, nil)

	if err := frame.GRedisCfgMgr.Load("../../../bin/redis_cfg.xml"); err != nil {
		elog.Errorf("RedidCfgMgr Load Error=%v", err)
		return
	}

	if err := eredis.GRedisModule.Init(frame.GRedisCfgMgr.ConnMaxCount, frame.GRedisCfgMgr.RedisConnSpecs); err != nil {
		elog.Errorf("RedisModule Init Error=%v", err)
		return
	}

	eredis.GRedisModule.GetRedisClient(1).Set("id", []byte(string(1)))
	eredis.GRedisModule.GetRedisClient(2).Set("id", []byte(string(2)))
	eredis.GRedisModule.GetRedisClient(3).Set("id", []byte(string(3)))
	eredis.GRedisModule.GetRedisClient(4).Set("id", []byte(string(4)))
	eredis.GRedisModule.GetRedisClient(5).Set("id", []byte(string(5)))
	eredis.GRedisModule.GetRedisClient(6).Set("id", []byte(string(6)))
	eredis.GRedisModule.GetRedisClient(7).Set("id", []byte(string(7)))
	eredis.GRedisModule.GetRedisClient(8).Set("id", []byte(string(8)))
	eredis.GRedisModule.GetRedisClient(9).Set("id", []byte(string(9)))
	eredis.GRedisModule.GetRedisClient(10).Set("id", []byte(string(10)))

	for i := 0; i < 10; i++ {
		value, _ := eredis.GRedisModule.GetRedisClient(uint64(i)).Get("id")
		elog.InfoAf("[id] uid=%v, value=%v", i, value)
	}

	for {
		time.Sleep(1 * time.Millisecond)
	}
}
