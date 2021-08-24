package main

import (
	"fmt"
	"time"

	"github.com/zjh-tech/go-frame/engine/elog"
	"github.com/zjh-tech/go-frame/engine/enet"
)

type TcpClient struct {
	logger *elog.Logger
}

func (t *TcpClient) Init() bool {
	cfgPath := "./cfg.yaml"
	cfg, err := ReadCfg(cfgPath)
	if err != nil {
		fmt.Printf("ReadCfg Error=%v", err)
		return false
	}

	GCfg = cfg

	t.logger = elog.NewLogger(GCfg.LogInfo.Path, GCfg.LogInfo.Level)
	t.logger.Init()
	ELog.SetLogger(t.logger)
	enet.ELog.SetLogger(t.logger)

	if !enet.GNet.Init() {
		ELog.Error("TcpClient Net Init Error")
		return false
	}

	enet.GCSSessionMgr = enet.NewCSSessionMgr()
	for i := 0; i < GCfg.ClientCount; i++ {
		enet.GCSSessionMgr.Connect(GCfg.TcpInfo.Addr, GClientMsgHandler, nil, true)
	}

	ELog.Info("TcpClient Init Success")
	return true
}

func (t *TcpClient) Run() {
	busy := false
	for {
		busy = false

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
