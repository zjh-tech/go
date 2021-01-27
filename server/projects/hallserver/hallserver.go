package main

import (
	"projects/config"
	"projects/frame"
	"projects/go-engine/edb"
	"projects/go-engine/ehttp"
	"projects/go-engine/elog"
	"projects/go-engine/enet"
	"projects/go-engine/etimer"
	"time"
)

type HallServer struct {
	frame.Server
}

func (h *HallServer) Init() bool {
	frame.GSSServerSessionMgr.SetLogicServerFactory(GLogicServerFactory)

	if !h.Server.Init() {
		return false
	}

	if !enet.GNet.Init() {
		elog.Error("HallServer Net Init Error")
		return false
	}
	elog.Info("HallServer Net Init Success")

	if err := config.GConfigMgr.LoadAllCfg("../config/binary"); err != nil {
		elog.Error("HallServer LoadAllCfg Error=%v", err)
		return false
	}

	//if frame.GServiceDiscoveryClient.Init(frame.GServerCfg.SDClientAddr, h.GetLocalServerID(), h.GetLocalToken(), nil) == false {
	//	elog.Error("HallServer SDClient Error")
	//	return false
	//}

	if frame.GServiceDiscoveryHttpClient.Init(frame.GServerCfg.SDClientUrl, h.GetLocalServerID(), h.GetLocalToken(), nil) == false {
		elog.Error("HallServer SDClient Http Error")
		return false
	}

	elog.Info("HallServer Init Success")
	return true
}

func (h *HallServer) Run() {
	busy := false
	net_module := enet.GNet
	http_net_module := ehttp.GHttpNet
	db_module := edb.GDBModule
	timer_module := etimer.GTimerMgr
	meter := frame.NewTimeMeter(frame.METER_LOOP_COUNT)

	for !h.Server.IsQuit() {
		busy = false
		meter.Clear()

		if net_module.Run(frame.NET_LOOP_COUNT) {
			busy = true
		}
		meter.Stamp()

		if http_net_module.Run(frame.HTTP_LOOP_COUNT) {
			busy = true
		}
		meter.Stamp()

		if db_module.Run(frame.DB_LOOP_COUNT) {
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
