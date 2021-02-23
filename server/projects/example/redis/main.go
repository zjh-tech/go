package main

import (
	"fmt"
	"projects/frame"
	"projects/go-engine/elog"
	"projects/go-engine/eredis"
	"projects/util"
	"time"
)

func main() {
	elog.Init("./log", 0, nil)

	if err := frame.GRedisCfgMgr.Load("./serverconfig/redis_cfg.xml"); err != nil {
		elog.Errorf("RedidCfgMgr Load Error=%v", err)
		return
	}

	if err := eredis.GRedisModule.Init(frame.GRedisCfgMgr.ConnMaxCount, frame.GRedisCfgMgr.RedisConnSpecs); err != nil {
		elog.Errorf("RedisModule Init Error=%v", err)
		return
	}

	start_tick := util.GetMillsecond()
	qps_count := 0
	loop_num := 500000
	for i := 0; i < loop_num; i++ {
		key := fmt.Sprintf("id%v", i)
		eredis.GRedisModule.GetRedisClient(uint64(i)).Set(key, []byte(util.Int2Str(i)))

		qps_count++
		end_tick := util.GetMillsecond()
		if (end_tick - start_tick) >= 1000 {
			elog.InfoAf("Get Qps=%v", qps_count)
			qps_count = 0
			start_tick = end_tick
		}
	}

	qps_count = 0
	for i := 0; i < loop_num; i++ {
		key := fmt.Sprintf("id%v", i)
		id_value, _ := eredis.GRedisModule.GetRedisClient(uint64(i)).Get(key)
		value, _ := util.Str2Int(string(id_value))
		if value != i {
			elog.InfoAf("key=%v, value=%v", key, value)
		}

		qps_count++
		end_tick := util.GetMillsecond()
		if (end_tick - start_tick) >= 1000 {
			elog.InfoAf("Set Qps=%v", qps_count)
			qps_count = 0
			start_tick = end_tick
		}
	}

	for {
		time.Sleep(1 * time.Millisecond)
	}
}
