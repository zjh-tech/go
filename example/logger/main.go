package main

import (
	"projects/go-engine/elog"
	"time"
)

func main() {
	logger := elog.NewLogger("./testlog", 0)
	logger.Init()
	ELog.SetLogger(logger)
	loop_num := int64(1000000)

	start_tick := time.Now().UnixNano() / 1e6
	for i := int64(0); i < loop_num; i++ {
		ELog.Debugf("This message is 116 characters long including the info that comes before it. %v", i)
	}
	end_tick := time.Now().UnixNano() / 1e6

	ELog.InfoAf("Sync Qps=%v", loop_num*1000/(end_tick-start_tick))

	//sync 和async 差不多

	for {
		time.Sleep(1 * time.Second)
	}
}
