package main

import (
	"projects/frame"
	"projects/go-engine/elog"
)

type MatchFunc func(datas []byte, h *MatchServer) bool

type MatchServer struct {
	frame.LogicServer
	dealer *frame.IDDealer
}

func NewMatchServer() *MatchServer {
	match := &MatchServer{
		dealer: frame.NewIDDealer(),
	}
	match.Init()
	return match
}

func (m *MatchServer) Init() bool {
	return true
}

func (m *MatchServer) OnHandler(msgID uint32, attach_datas []byte, datas []byte, sess *frame.SSServerSession) {
	defer func() {
		if err := recover(); err != nil {
			elog.ErrorAf("MatchServer onHandler MsgID = %v Error=%v", msgID, err)
		}
	}()

	dealer := m.dealer.FindHandler(msgID)
	if dealer == nil {
		elog.ErrorAf("MatchServer MsgHandler Can Not Find MsgID = %v", msgID)
		return
	}

	dealer.(MatchFunc)(datas, m)
}

func (m *MatchServer) OnEstablish(serversess *frame.SSServerSession) {
	elog.InfoAf("MatchServer OnEstablish Remote [ID=%v,Type=%v,Ip=%v] ", serversess.GetRemoteServerID(), serversess.GetRemoteServerType(), serversess.GetRemoteOuter())
}

func (m *MatchServer) OnTerminate(serversess *frame.SSServerSession) {
	elog.InfoAf("MatchServer OnTerminate Remote [ID=%v,Type=%v,Ip=%v] ", serversess.GetRemoteServerID(), serversess.GetRemoteServerType(), serversess.GetRemoteOuter())
}
