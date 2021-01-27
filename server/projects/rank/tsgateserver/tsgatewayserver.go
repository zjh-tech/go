package main

import (
	"projects/frame"
	"projects/go-engine/ehttp"
	"projects/go-engine/elog"
	"projects/go-engine/enet"
	"projects/go-engine/etimer"
	"projects/pb"
	"projects/rank/tscommon"
	"time"
)

const (
	NTF_TSGATEWAY_INFO_TIMER_ID uint32 = 1
)

const (
	NTF_TSGATEWAY_INFO_TIMER_DELAY uint64 = 1000 * 60
)

type TsGatewayServer struct {
	frame.Server
	timerRegister etimer.ITimerRegister
}

func (t *TsGatewayServer) Init() bool {
	t.timerRegister = etimer.NewTimerRegister()

	frame.GSSServerSessionMgr.SetLogicServerFactory(GLogicServerFactory)

	if !t.Server.Init() {
		return false
	}

	if !enet.GNet.Init() {
		elog.Error("TsGatewayServer Net Init Error")
		return false
	}
	elog.Info("TsGatewayServer Net Init Success")

	flag := frame.GServiceDiscoveryHttpClient.Init(frame.GServerCfg.SDClientUrl, t.GetLocalServerID(), t.GetLocalToken(), func(...interface{}) {
		if frame.GServerCfg.C2SInterListen == "" || frame.GServerCfg.C2SOuterListen == "" {
			elog.Error("TsGatewayServer C2SInterListen C2SOuterListen Is Empty")
			return
		}
		if frame.GSSClientSessionMgr.SSClientListen(frame.GServerCfg.C2SInterListen, GTsRankClient, nil) == false {
			elog.Error("TsGatewayServer C2SInterListen Error")
			frame.GServer.Quit()
			return
		}
	})

	if !flag {
		elog.Error("TsGatewayServer SDClient Error")
		return false
	}

	if rankCfg, err := tscommon.ReadRankCfg("../config/rank_config.xml"); err != nil {
		elog.Errorf("TsGatewayServer Load RankConfig xml Error=%v", err)
		return false
	} else {
		tscommon.GRankCfg = rankCfg
	}

	t.addTsGatewayNtfTimer()

	elog.Info("TsGatewayServer Init Success")
	return true
}

func (t *TsGatewayServer) Run() {
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

func (t *TsGatewayServer) addTsGatewayNtfTimer() {
	t.timerRegister.AddRepeatTimer(NTF_TSGATEWAY_INFO_TIMER_ID, NTF_TSGATEWAY_INFO_TIMER_DELAY, "TsGatewayServer-NtfGatewayInfo", func(v ...interface{}) {
		BroadCastTs2TsGatewayInfoNtf()
	}, []interface{}{}, true)
}

func BroadCastTs2TsGatewayInfoNtf() {
	ntf := &pb.Ts2TsGatewayInfoNtf{}
	ntf.RemoteAddr = frame.GServerCfg.C2SOuterListen
	ntf.Token = frame.GServer.GetLocalToken()
	ntf.ClientConnCount = uint32(frame.GSSClientSessionMgr.Count())
	frame.GSSServerSessionMgr.BroadProtoMsg(frame.TS_RANK_BALANCE_SERVER_TYPE, uint32(pb.TS2TSLogicMsgId_ts2ts_gateway_info_ntf_id), ntf)
	elog.InfoA("TsGatewayServer TsGatewayInfo Ntf")
}
