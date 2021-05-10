package main

import (
	"time"

	"github.com/zjh-tech/go-frame/engine/enet"
	"github.com/zjh-tech/go-frame/engine/etimer"
	"github.com/zjh-tech/go-frame/frame"
)

type SGServer struct {
	frame.Server
}

func (s *SGServer) Init() bool {
	if !s.Server.Init() {
		return false
	}

	ELog.SetLogger(s.Server.GetLogger())

	if !enet.GNet.Init() {
		ELog.Error("RegistryServer Net Init Error")
		return false
	}
	ELog.Info("RegistryServer Net Init Success")

	service_registry_path := "./service_registry.xml"
	cfg, err := ReadRegistryCfg(service_registry_path)
	if err != nil {
		ELog.Errorf("RegistryServer ReadRegistryCfg Error=%v", err)
		return false
	}
	GRegistryCfg = cfg
	ELog.Info("RegistryServer ReadRegistryCfg  Success")

	GServiceDiscoveryServer.Init(service_registry_path)

	//frame.GSDKSessionMgr.SSClientListen(frame.GServerCfg.SDServerAddr, NewSDServer(), nil)

	go frame.StartHttpServer(frame.GServerCfg.SDServerUrl, NewSDHttpServer())

	ELog.Info("RegistryServer Init Success")
	return true
}

func (s *SGServer) Run() {
	busy := false
	net_module := enet.GNet
	timer_module := etimer.GTimerMgr
	async_module := GAsyncModule
	meter := frame.NewTimeMeter(frame.METER_LOOP_COUNT)

	for !s.Server.IsQuit() {
		busy = false
		meter.Clear()

		if net_module.Run(frame.NET_LOOP_COUNT) {
			busy = true
		}
		meter.Stamp()

		if async_module.Run(frame.ASYNC_LOOP_COUNT) {
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
