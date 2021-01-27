package main

import (
	"projects/frame"
	"projects/go-engine/edb"
	"projects/go-engine/ehttp"
	"projects/go-engine/elog"
	"projects/go-engine/enet"
	"projects/go-engine/etimer"
	"time"
)

type BattleServer struct {
	frame.Server
}

func (b *BattleServer) Init() bool {
	frame.GSSServerSessionMgr.SetLogicServerFactory(GLogicServerFactory)

	if !b.Server.Init() {
		return false
	}

	if !enet.GNet.Init() {
		elog.Error("BattleServer Net Init Error")
		return false
	}
	elog.Info("BattleServer Net Init Success")

	//if frame.GServiceDiscoveryClient.Init(frame.GServerCfg.SDClientAddr, b.GetLocalServerID(), b.GetLocalToken(), nil) == false {
	//	elog.Error("BattleServer SDClient Error")
	//	return false
	//}

	if frame.GServiceDiscoveryHttpClient.Init(frame.GServerCfg.SDClientUrl, b.GetLocalServerID(), b.GetLocalToken(), nil) == false {
		elog.Error("BattleServer SDClient Http Error")
		return false
	}

	elog.Info("BattleServer Init Success")
	return true
}

func (h *BattleServer) Run() {
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
