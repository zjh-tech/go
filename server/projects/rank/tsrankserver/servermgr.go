package main

import (
	"projects/frame"
)

type LogicServerFactory struct {
}

func (l *LogicServerFactory) SetLogicServer(serversess *frame.SSServerSession) {
	if serversess == nil {
		return
	}

	serverType := serversess.GetRemoteServerType()
	if serverType == frame.TS_RANK_GATEWAY_SERVER_TYPE {
		serversess.SetLogicServer(NewTsGatewayServer())
	}
}

var GLogicServerFactory *LogicServerFactory

func init() {
	GLogicServerFactory = &LogicServerFactory{}
}
