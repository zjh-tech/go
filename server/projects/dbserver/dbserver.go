package main

import (
	"projects/frame"
	"projects/go-engine/edb"
	"projects/go-engine/elog"
	"projects/go-engine/enet"
	"projects/go-engine/etimer"
	"time"
)

type DbServer struct {
	frame.Server
}

func (d *DbServer) Init() bool {
	frame.GSSServerSessionMgr.SetLogicServerFactory(GLogicServerFactory)

	if !d.Server.Init() {
		return false
	}

	if !enet.GNet.Init() {
		elog.Error("DbServer Net Init Error")
		return false
	}
	elog.Info("DbServer Net Init Success")

	if err := frame.GDatabaseCfgMgr.Load("./db_cfg.xml"); err != nil {
		elog.Error(err)
		return false
	}

	if err := edb.GDBModule.Init(frame.GDatabaseCfgMgr.DBConnMaxCount, frame.GDatabaseCfgMgr.DBTableMaxCount, frame.GDatabaseCfgMgr.DBConnSpecs); err != nil {
		elog.Error(err)
		return false
	}

	//if frame.GServiceDiscoveryClient.Init(frame.GServerCfg.SDClientAddr, d.GetLocalServerID(), d.GetLocalToken(), nil) == false {
	//	elog.Error("DbServer SDClient Error")
	//	return false
	//}

	if frame.GServiceDiscoveryHttpClient.Init(frame.GServerCfg.SDClientUrl, d.GetLocalServerID(), d.GetLocalToken(), nil) == false {
		elog.Error("DbServer SDClient Http Error")
		return false
	}

	elog.Info("DbServer Init Success")
	return true
}

func (d *DbServer) Run() {
	busy := false
	net_module := enet.GNet
	timer_module := etimer.GTimerMgr
	meter := frame.NewTimeMeter(frame.METER_LOOP_COUNT)

	for !d.Server.IsQuit() {
		busy = false
		meter.Clear()

		if net_module.Run(frame.NET_LOOP_COUNT) {
			busy = true
		}
		meter.Stamp()

		if timer_module.Update(frame.TIMER_LOOP_COUNT) {
			busy = true
		}
		meter.CheckOut()

		if !busy {
			time.Sleep(1 * time.Millisecond)
		}
	}
}
