package main

import (
	"fmt"
	"projects/frame"
	"projects/go-engine/ehttp"
	"projects/go-engine/elog"
	"projects/go-engine/enet"
	"projects/go-engine/etimer"
	"projects/util"
	"time"
)

type Robot struct {
}

//https://github.com/bigwhite/experiments/tree/master/gohttps
func (r *Robot) Init() bool {
	//Log
	elog.Init("./log", 1, nil, nil)
	elog.Info("Server Log System Init Success")

	if !enet.GNet.Init() {
		elog.Error("Robot Net Init Error")
		return false
	}
	elog.Info("Robot Net Init Success")

	if GLoginSys.Init("ca.crt", "localhost:3000") == false {
		elog.Info("Robot GLoginSys Error")
		return false
	}

	accountName := fmt.Sprintf("%s_%d", "test", util.GetMillsecond())
	go GLoginSys.SendCsAccountRegisterReq(accountName, "123456")

	elog.Info("Robot Init Success")
	return true
}

func (r *Robot) Run() {
	net_module := enet.GNet
	http_net_module := ehttp.GHttpNet
	timer_module := etimer.GTimerMgr
	busy := false
	for {
		busy = false

		if net_module.Run(frame.NET_LOOP_COUNT) {
			busy = true
		}

		if http_net_module.Run(frame.HTTP_LOOP_COUNT) {
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

func (r *Robot) UnInit() {

}
