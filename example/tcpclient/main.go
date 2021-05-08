package main

import (
	"projects/frame"
	"projects/engine/elog"
	"projects/engine/enet"
	"projects/engine/etimer"
	"time"
)

type TcpClient struct {
	logger *elog.Logger
}

func (t *TcpClient) Init() bool {
	t.logger = elog.NewLogger("./log", 0)
	t.logger.Init()
	ELog.SetLogger(t.logger)

	ELog.Info("TcpClient Log System Init Success")

	if !enet.GNet.Init() {
		ELog.Error("TcpClient Net Init Error")
		return false
	}
	ELog.Info("TcpClient Net Init Success")

	for i := 0; i < 5; i++ {
		frame.GCSSessionMgr.Connect("127.0.0.1:2000", GClientMsgHandler, nil)
	}

	ELog.Info("TcpClient Init Success")
	return true
}

func (t *TcpClient) Run() {
	net_module := enet.GNet
	timer_module := etimer.GTimerMgr
	busy := false
	for {
		busy = false

		if net_module.Run(frame.NET_LOOP_COUNT) {
			busy = true
		}

		if timer_module.Update(frame.TIMER_LOOP_COUNT) {
			busy = true
		}

		if !busy {
			time.Sleep(1 * time.Millisecond)
		}
	}
}

func (t *TcpClient) UnInit() {
	t.logger.UnInit()
}

func main() {
	var client TcpClient
	if client.Init() {
		client.Run()
	}

	client.UnInit()
}
