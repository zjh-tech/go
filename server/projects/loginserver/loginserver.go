package main

import (
	"projects/config"
	"projects/frame"
	"projects/go-engine/edb"
	"projects/go-engine/ehttp"
	"projects/go-engine/elog"
	"projects/go-engine/enet"
	"projects/go-engine/eredis"
	"projects/go-engine/etimer"
	"time"
)

type LoginServer struct {
	frame.Server
}

func (l *LoginServer) Init() bool {
	frame.GSSServerSessionMgr.SetLogicServerFactory(GLogicServerFactory)

	if !l.Server.Init() {
		return false
	}

	if !enet.GNet.Init() {
		elog.Error("LoginServer Net Init Error")
		return false
	}
	elog.Info("LoginServer Net Init Success")

	if err := frame.GDatabaseCfgMgr.Load("./db_cfg.xml"); err != nil {
		elog.Error(err)
		return false
	}

	if err := edb.GDBModule.Init(frame.GDatabaseCfgMgr.DBConnMaxCount, frame.GDatabaseCfgMgr.DBTableMaxCount, frame.GDatabaseCfgMgr.DBConnSpecs); err != nil {
		elog.Error(err)
		return false
	}

	if err := frame.GRedisCfgMgr.Load("../config/serverconfig/redis_cfg.xml"); err != nil {
		elog.Error(err)
		return false
	}

	if err := eredis.GRedisModule.Init(frame.GRedisCfgMgr.ConnMaxCount, frame.GRedisCfgMgr.RedisConnSpecs); err != nil {
		elog.Error(err)
		return false
	}

	//flag := frame.GServiceDiscoveryClient.Init(frame.GServerCfg.SDClientAddr, l.GetLocalServerID(), l.GetLocalToken(), func(...interface{}) {
	//	if frame.GServerCfg.C2SHttpsUrl == "" {
	//		elog.Error("LoginServer C2SHttpsUrl Is Empty")
	//		frame.GServer.Quit()
	//		return
	//	}
	//
	//	if frame.GServerCfg.C2SHttpsCert != "" && frame.GServerCfg.C2SHttpsKey != "" {
	//		go frame.StartHttpsServer(frame.GServerCfg.C2SHttpsUrl, frame.GServerCfg.C2SHttpsCert, frame.GServerCfg.C2SHttpsKey, &ClientMsgHandler{})
	//	} else {
	//		go frame.StartHttpServer(frame.GServerCfg.C2SHttpsUrl, &ClientMsgHandler{})
	//	}
	//})

	flag := frame.GServiceDiscoveryHttpClient.Init(frame.GServerCfg.SDClientUrl, l.GetLocalServerID(), l.GetLocalToken(), func(...interface{}) {
		if frame.GServerCfg.C2SHttpsUrl == "" {
			elog.Error("LoginServer C2SHttpsUrl Is Empty")
			frame.GServer.Quit()
			return
		}

		if frame.GServerCfg.C2SHttpsCert != "" && frame.GServerCfg.C2SHttpsKey != "" {
			go frame.StartHttpsServer(frame.GServerCfg.C2SHttpsUrl, frame.GServerCfg.C2SHttpsCert, frame.GServerCfg.C2SHttpsKey, &ClientMsgHandler{})
		} else {
			go frame.StartHttpServer(frame.GServerCfg.C2SHttpsUrl, &ClientMsgHandler{})
		}
	})

	if !flag {
		elog.Error("LoginServer SDClient Error")
		return false
	}

	if err := config.GConfigMgr.LoadAllCfg("../config/binary"); err != nil {
		elog.Error("LoginServer LoadAllCfg Error=%v", err)
		return false
	}

	GTokenMgr.Init()

	elog.Info("LoginServer Init Success")
	return true
}

func (l *LoginServer) Run() {
	busy := false
	net_module := enet.GNet
	http_net_module := ehttp.GHttpNet
	db_module := edb.GDBModule
	timer_module := etimer.GTimerMgr
	meter := frame.NewTimeMeter(frame.METER_LOOP_COUNT)

	for !l.Server.IsQuit() {
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
