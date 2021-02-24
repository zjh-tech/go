package main

import (
	"projects/go-engine/elog"
	"projects/util"
	"time"
)

func main() {
	elog.Init("./log", 0, nil)

	loop_num := int64(1000000)

	start_tick := util.GetMillsecond()
	for i := int64(0); i < loop_num; i++ {
		elog.Debugf("This message is 116 characters long including the info that comes before it. %v", i)
	}
	end_tick := util.GetMillsecond()
	elog.InfoAf("Sync Qps=%v", loop_num*1000/(end_tick-start_tick))

	//sync 和async 差不多

	for {
		time.Sleep(1 * time.Second)
	}

}
