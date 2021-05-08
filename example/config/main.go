package main

import (
	"projects/config"
	"projects/go-engine/elog"
	"time"
)

func main() {
	logger := elog.NewLogger("./log", 0)
	logger.Init()
	config.ELog.SetLogger(logger)
	ELog.SetLogger(logger)
	if err := config.GConfigMgr.LoadAllCfg("./binary"); err != nil {
		ELog.Errorf("Error=%v", err)
	}

	tip := config.GConfigMgr.GetTipByID(1)
	if tip != nil {
		ELog.InfoAf("Item=%+v", tip)
	}

	ELog.InfoAf("LoadAllCfg Ok")

	for {
		time.Sleep(1 * time.Millisecond)
	}
	logger.UnInit()
}
