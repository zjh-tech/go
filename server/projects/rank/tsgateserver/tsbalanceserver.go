package main

import (
	"projects/frame"
	"projects/go-engine/elog"
)

type TsBalanceServerFunc func(datas []byte, t *TsBalanceServer) bool

type TsBalanceServer struct {
	frame.LogicServer
	dealer *frame.IDDealer

	ClientIp        string
	ClientPort      uint32
	ClientConnCount uint32
}

func NewTsBalanceServer() *TsBalanceServer {
	gateway := &TsBalanceServer{
		dealer: frame.NewIDDealer(),
	}
	gateway.Init()
	return gateway
}
func (t *TsBalanceServer) Init() bool {
	return true
}

func (t *TsBalanceServer) OnHandler(msgID uint32, attach_datas []byte, datas []byte, sess *frame.SSServerSession) {
	defer func() {
		if err := recover(); err != nil {
			elog.ErrorAf("TsBalanceServer onHandler MsgID = %v Error=%v", msgID, err)
		}
	}()

	dealer := t.dealer.FindHandler(msgID)
	if dealer == nil {
		elog.ErrorAf("TsBalanceServer MsgHandler Can Not Find MsgID = %v", msgID)
		return
	}
}

func (t *TsBalanceServer) OnEstablish(serversess *frame.SSServerSession) {
	elog.InfoAf("TsBalanceServer OnEstablish Remote [ID=%v,Type=%v,Ip=%v] ", serversess.GetRemoteServerID(), serversess.GetRemoteServerType(), serversess.GetRemoteOuter())
	BroadCastTs2TsGatewayInfoNtf()
}

func (t *TsBalanceServer) OnTerminate(serversess *frame.SSServerSession) {
	elog.InfoAf("TsBalanceServer OnTerminate Remote [ID=%v,Type=%v,Ip=%v] ", serversess.GetRemoteServerID(), serversess.GetRemoteServerType(), serversess.GetRemoteOuter())
}
