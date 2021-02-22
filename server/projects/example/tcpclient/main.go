package main

import (
	"projects/frame"
	"projects/go-engine/elog"
	"projects/go-engine/enet"
	"projects/go-engine/etimer"
	"time"
)

type TcpClient struct {
}

func (t *TcpClient) Init() bool {
	elog.Init("./log", 1, nil)
	elog.Info("TcpClient Log System Init Success")

	if !enet.GNet.Init() {
		elog.Error("TcpClient Net Init Error")
		return false
	}
	elog.Info("TcpClient Net Init Success")

	for i := 0; i < 5; i++ {
		frame.GCSClientSessionMgr.Connect("127.0.0.1:2000", GClientMsgHandler, nil)
	}

	elog.Info("TcpClient Init Success")
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
	elog.UnInit(nil)
}

func main() {
	var client TcpClient
	if client.Init() {
		client.Run()
	}

	client.UnInit()
}
