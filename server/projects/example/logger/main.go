package main

import (
	"projects/go-engine/elog"
	"time"
)

func main() {
	elog.InitLog("./log", 0, nil, nil)
	for i := 0; i < 10000000; i++ {
		elog.DebugA("DebugA")
		elog.InfoA("InfoA")
	}

	elog.Debug("a")
	elog.Info("b")

	for {
		time.Sleep(1 * time.Second)
	}

}
