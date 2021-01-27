package main

import (
	"projects/frame"
	"projects/go-engine/ehttp"
	"time"

	"projects/go-engine/elog"
	"projects/go-engine/enet"
	"projects/go-engine/etimer"
)

type RegistryServer struct {
	frame.Server
}

func (r *RegistryServer) Init() bool {
	if !r.Server.Init() {
		return false
	}

	if !enet.GNet.Init() {
		elog.Error("RegistryServer Net Init Error")
		return false
	}
	elog.Info("RegistryServer Net Init Success")

	service_registry_path := "./service_registry.xml"
	cfg, err := ReadRegistryCfg(service_registry_path)
	if err != nil {
		elog.Errorf("RegistryServer ReadRegistryCfg Error=%v", err)
		return false
	}
	GRegistryCfg = cfg
	elog.Info("RegistryServer ReadRegistryCfg  Success")

	GServiceDiscoveryServer.Init(service_registry_path)

	//frame.GSSClientSessionMgr.SSClientListen(frame.GServerCfg.SDServerAddr, NewSDServer(), nil)

	go frame.StartHttpServer(frame.GServerCfg.SDServerUrl, NewSDHttpServer())

	elog.Info("RegistryServer Init Success")
	return true
}

func (r *RegistryServer) Run() {
	busy := false
	net_module := enet.GNet
	http_net_module := ehttp.GHttpNet
	timer_module := etimer.GTimerMgr
	meter := frame.NewTimeMeter(frame.METER_LOOP_COUNT)

	for !r.Server.IsQuit() {
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
