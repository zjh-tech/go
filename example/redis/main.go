package main

import (
	"fmt"
	"projects/base/convert"
	"projects/base/util"
	"projects/frame"
	"projects/engine/elog"
	"projects/engine/eredis"
	"time"
)

func main() {
	logger := elog.NewLogger("./log", 0)
	logger.Init()
	ELog.SetLogger(logger)

	if err := frame.GRedisCfgMgr.Load("./serverconfig/redis_cfg.xml"); err != nil {
		ELog.Errorf("RedidCfgMgr Load Error=%v", err)
		return
	}

	if err := eredis.GRedisModule.Init(frame.GRedisCfgMgr.ConnMaxCount, frame.GRedisCfgMgr.RedisConnSpecs); err != nil {
		ELog.Errorf("RedisModule Init Error=%v", err)
		return
	}

	start_tick := util.GetMillsecond()
	qps_count := 0
	loop_num := 500000
	for i := 0; i < loop_num; i++ {
		key := fmt.Sprintf("id%v", i)
		eredis.GRedisModule.GetRedisClient(uint64(i)).Set(key, []byte(convert.Int2Str(i)))

		qps_count++
		end_tick := util.GetMillsecond()
		if (end_tick - start_tick) >= 1000 {
			ELog.InfoAf("Get Qps=%v", qps_count)
			qps_count = 0
			start_tick = end_tick
		}
	}

	qps_count = 0
	for i := 0; i < loop_num; i++ {
		key := fmt.Sprintf("id%v", i)
		id_value, _ := eredis.GRedisModule.GetRedisClient(uint64(i)).Get(key)
		value, _ := convert.Str2Int(string(id_value))
		if value != i {
			ELog.InfoAf("key=%v, value=%v", key, value)
		}

		qps_count++
		end_tick := util.GetMillsecond()
		if (end_tick - start_tick) >= 1000 {
			ELog.InfoAf("Set Qps=%v", qps_count)
			qps_count = 0
			start_tick = end_tick
		}
	}

	for {
		time.Sleep(1 * time.Millisecond)
	}
}
