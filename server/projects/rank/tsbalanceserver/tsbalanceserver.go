package main

import (
	"projects/frame"
	"projects/go-engine/ehttp"
	"projects/go-engine/elog"
	"projects/go-engine/enet"
	"projects/go-engine/etimer"
	"time"
)

type TsBalanceServer struct {
	frame.Server
}

func (t *TsBalanceServer) Init() bool {
	frame.GSSServerSessionMgr.SetLogicServerFactory(GLogicServerFactory)

	if !t.Server.Init() {
		return false
	}

	if !enet.GNet.Init() {
		elog.Error("TsBalanceServer Net Init Error")
		return false
	}
	elog.Info("TsBalanceServer Net Init Success")

	flag := frame.GServiceDiscoveryHttpClient.Init(frame.GServerCfg.SDClientUrl, t.GetLocalServerID(), t.GetLocalToken(), func(...interface{}) {
		if frame.GServerCfg.C2SInterListen == "" || frame.GServerCfg.C2SOuterListen == "" {
			elog.Error("TsBalanceServer C2SInterListen C2SOuterListen Is Empty")
			frame.GServer.Quit()
			return
		}
		if frame.GSSClientSessionMgr.SSClientListen(frame.GServerCfg.C2SInterListen, GTsRankClient, nil) == false {
			elog.Error("TsBalanceServer C2SInterListen Error")
			frame.GServer.Quit()
			return
		}
	})

	if !flag {
		elog.Error("TsBalanceServer SDClient Error")
		return false
	}

	elog.Info("TsBalanceServer Init Success")
	return true
}

func (t *TsBalanceServer) Run() {
	busy := false
	net_module := enet.GNet
	http_net_module := ehttp.GHttpNet
	timer_module := etimer.GTimerMgr
	meter := frame.NewTimeMeter(frame.METER_LOOP_COUNT)

	for !t.Server.IsQuit() {
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
