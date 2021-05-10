package main

import (
	"math"
	"projects/engine/elog"
	"projects/engine/enet"
	"projects/engine/etimer"
	"projects/frame"
	"time"
)

type TcpServer struct {
	logger *elog.Logger
}

func (t *TcpServer) Init() bool {
	t.logger = elog.NewLogger("./log", 0)
	t.logger.Init()
	ELog.SetLogger(t.logger)

	ELog.Info("TcpServer Log System Init Success")

	GClientMsgHandler.Init()

	if frame.GCSSessionMgr.Listen("127.0.0.1:2000", GClientMsgHandler, nil, math.MaxUint16) == false {
		return false
	}

	ELog.Info("TcpServer Init Success")
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
