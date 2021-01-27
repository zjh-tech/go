package main

import (
	"projects/config"
	"projects/go-engine/elog"
	"time"
)

func main() {
	elog.InitLog("./log", 0, nil, nil)
	if err := config.GConfigMgr.LoadAllCfg("../../../bin/binary"); err != nil {
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
