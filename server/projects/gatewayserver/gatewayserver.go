package main

import (
	"projects/frame"
	"projects/go-engine/ehttp"
	"projects/go-engine/elog"
	"projects/go-engine/enet"
	"projects/go-engine/etimer"
	"time"
)

type GatewayServer struct {
	frame.Server
}

func (g *GatewayServer) Init() bool {
	frame.GSSServerSessionMgr.SetLogicServerFactory(GLogicServerFactory)

	if !g.Server.Init() {
		return false
	}

	if !enet.GNet.Init() {
		elog.Error("GatewayServer Net Init Error")
		return false
	}
	elog.Info("GatewayServer Net Init Success")

	//flag := frame.GServiceDiscoveryClient.Init(frame.GServerCfg.SDClientAddr, g.GetLocalServerID(), g.GetLocalToken(), func(...interface{}) {
	//	if frame.GServerCfg.C2SInterListen == "" || frame.GServerCfg.C2SOuterListen == "" {
	//		elog.Error("GatewayServer C2SInterListen C2SOuterListen Is Empty")
	//		frame.GServer.Quit()
	//		return
	//	}
	//	if frame.GCSClientSessionMgr.Listen(frame.GServerCfg.C2SInterListen, GClientMsgHandler, nil) == false {
	//		elog.Error("GatewayServer C2SInterListen Error")
	//		frame.GServer.Quit()
	//		return
	//	}
	//})

	flag := frame.GServiceDiscoveryHttpClient.Init(frame.GServerCfg.SDClientUrl, g.GetLocalServerID(), g.GetLocalToken(), func(...interface{}) {
		if frame.GServerCfg.C2SInterListen == "" || frame.GServerCfg.C2SOuterListen == "" {
			elog.Error("GatewayServer C2SInterListen C2SOuterListen Is Empty")
			frame.GServer.Quit()
			return
		}
		if frame.GCSClientSessionMgr.Listen(frame.GServerCfg.C2SInterListen, GClientMsgHandler, nil) == false {
			elog.Error("GatewayServer C2SInterListen Error")
			frame.GServer.Quit()
			return
		}
	})

	if !flag {
		elog.Error("GatewayServer SDClient Error")
		return false
	}

	GPlayerMgr.Init()
	elog.Info("GatewayServer Init Success")
	return true
}

func (g *GatewayServer) Run() {
	busy := false
	net_module := enet.GNet
	http_net_module := ehttp.GHttpNet
	timer_module := etimer.GTimerMgr
	meter := frame.NewTimeMeter(frame.METER_LOOP_COUNT)

	for !g.Server.IsQuit() {
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
