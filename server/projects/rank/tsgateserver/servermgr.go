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
	if serverType == frame.TS_RANK_BALANCE_SERVER_TYPE {
		serversess.SetLogicServer(NewTsBalanceServer())
	} else if serverType == frame.TS_RANK_RANK_SERVER_TYPE {
		serversess.SetLogicServer(NewTsRankServer())
	}
}

var GLogicServerFactory *LogicServerFactory

func init() {
	GLogicServerFactory = &LogicServerFactory{}
}
