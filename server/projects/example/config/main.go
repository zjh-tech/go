package main

import (
	"projects/config"
	"projects/go-engine/elog"
	"time"
)

func main() {
	elog.Init("./log", 0, nil)
	if err := config.GConfigMgr.LoadAllCfg("./binary"); err != nil {
		elog.Errorf("Error=%v", err)
	}

	tip := config.GConfigMgr.GetTipByID(1)
	if tip != nil {
		elog.InfoAf("Item=%+v", tip)
	}

	for {
		time.Sleep(1 * time.Millisecond)
	}
}
