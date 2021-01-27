package main

import "projects/frame"

type LogicServerFactory struct {
}

func (l *LogicServerFactory) SetLogicServer(serversess *frame.SSServerSession) {
	if serversess == nil {
		return
	}

	serverType := serversess.GetRemoteServerType()
	if serverType == frame.GATEWAY_SERVER_TYPE {
		serversess.SetLogicServer(NewGatewayServer())
	} else if serverType == frame.MATCH_SERVER_TYPE {
		serversess.SetLogicServer(NewMatchServer())
	} else if serverType == frame.CENTER_SERVER_TYPE {
		serversess.SetLogicServer(NewCenterServer())
	} else if serverType == frame.DB_SERVER_TYPE {
		serversess.SetLogicServer(NewDBServer())
	}
}

var GLogicServerFactory *LogicServerFactory

func init() {
	GLogicServerFactory = &LogicServerFactory{}
}
