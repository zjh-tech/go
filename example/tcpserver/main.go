package main

import (
	"fmt"
	"time"

	"github.com/zjh-tech/go-frame/engine/elog"
	"github.com/zjh-tech/go-frame/engine/enet"
)

type TcpServer struct {
	logger *elog.Logger
}

func (t *TcpServer) Init() bool {
	cfgPath := "./cfg.yaml"
	cfg, err := ReadCfg(cfgPath)
	if err != nil {
		fmt.Printf("ReadCfg Error=%v", err)
		return false
	}

	GCfg = cfg

	t.logger = elog.NewLogger(GCfg.LogInfo.Path, GCfg.LogInfo.Level)
	t.logger.Init()
	ELog = t.logger
	enet.ELog = t.logger

	enet.GCSSessionMgr = enet.NewCSSessionMgr()
	if !enet.GCSSessionMgr.Listen(GCfg.TcpInfo.Addr, GClientMsgHandler, nil, 3000, true) {
		return false
	}

	PrintQps()

	ELog.Info("TcpServer Init Success")
	return true
}

func (d *TcpServer) Run() {
	busy := false

	for {
		busy = false

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
