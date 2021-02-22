package main

import (
	"math"
	"projects/frame"
	"projects/go-engine/elog"
	"projects/go-engine/enet"
	"projects/go-engine/etimer"
	"time"
)

type TcpServer struct {
}

func (t *TcpServer) Init() bool {
	elog.Init("./log", 1, nil)
	elog.Info("TcpServer Log System Init Success")

	GClientMsgHandler.Init()

	if frame.GCSClientSessionMgr.Listen("127.0.0.1:2000", GClientMsgHandler, nil, math.MaxUint16) == false {
		return false
	}

	elog.Info("TcpServer Init Success")
	return true
}

func (d *TcpServer) Run() {
	busy := false
	net_module := enet.GNet
	timer_module := etimer.GTimerMgr
	meter := frame.NewTimeMeter(frame.METER_LOOP_COUNT)

	for {
		busy = false
		meter.Clear()

		if net_module.Run(frame.NET_LOOP_COUNT) {
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

func main() {
	var server TcpServer
	if server.Init() {
		server.Run()
	}
}
