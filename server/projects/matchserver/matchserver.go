package main

import (
	"projects/frame"
	"projects/go-engine/ehttp"
	"projects/go-engine/elog"
	"projects/go-engine/enet"
	"projects/go-engine/etimer"
	"time"
)

type MatchServer struct {
	frame.Server
}

func (m *MatchServer) Init() bool {
	frame.GSSServerSessionMgr.SetLogicServerFactory(GLogicServerFactory)

	if !m.Server.Init() {
		return false
	}

	if !enet.GNet.Init() {
		elog.Error("MatchServer Net Init Error")
		return false
	}
	elog.Info("MatchServer Net Init Success")

	//if frame.GServiceDiscoveryClient.Init(frame.GServerCfg.SDClientAddr, m.GetLocalServerID(), m.GetLocalToken(), nil) == false {
	//	elog.Error("MatchServer SDClient Error")
	//	return false
	//}

	if frame.GServiceDiscoveryHttpClient.Init(frame.GServerCfg.SDClientUrl, m.GetLocalServerID(), m.GetLocalToken(), nil) == false {
		elog.Error("MatchServer SDClient Http Error")
		return false
	}

	elog.Info("MatchServer Init Success")
	return true
}

func (h *MatchServer) Run() {
	busy := false
	net_module := enet.GNet
	http_net_module := ehttp.GHttpNet
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

		if timer_module.Update(frame.TIMER_LOOP_COUNT) {
			busy = true
		}
		meter.CheckOut()

		if !busy {
			time.Sleep(1 * time.Millisecond)
		}
	}
}
