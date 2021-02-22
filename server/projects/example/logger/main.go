package main

import (
	"projects/go-engine/elog"
	"projects/util"
	"time"
)

//log 12万 QPS
func main() {
	elog.Init("./log", 0, nil)

	start_tick := util.GetMillsecond()
	qps_count := 0
	loop_num := 1000000
	for i := 0; i < loop_num; i++ {
		elog.Debug("DebugA")

		qps_count++
		end_tick := util.GetMillsecond()
		if (end_tick - start_tick) >= 1000 {
			elog.InfoAf("Sync Qps=%v", qps_count)
			qps_count = 0
			start_tick = end_tick
		}
	}

	for i := 0; i < loop_num; i++ {
		elog.DebugA("DebugA")
	}

	for {
		time.Sleep(1 * time.Second)
	}

}
