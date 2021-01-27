package main

import (
	"projects/frame"
	"projects/go-engine/elog"
)

type BattleFunc func(datas []byte, b *BattleServer) bool

type BattleServer struct {
	frame.LogicServer
	dealer *frame.IDDealer
}

func NewBattleServer() *BattleServer {
	battle := &BattleServer{
		dealer: frame.NewIDDealer(),
	}
	battle.Init()
	return battle
}

func (b *BattleServer) Init() bool {
	return true
}

func (b *BattleServer) OnHandler(msgID uint32, attach_datas []byte, datas []byte, sess *frame.SSServerSession) {
	defer func() {
		if err := recover(); err != nil {
			elog.ErrorAf("BattleServer onHandler MsgID = %v Error=%v", msgID, err)
		}
	}()

	dealer := b.dealer.FindHandler(msgID)
	if dealer == nil {
		elog.ErrorAf("BattleServer MsgHandler Can Not Find MsgID = %v", msgID)
		return
	}

	dealer.(BattleFunc)(datas, b)
}

func (b *BattleServer) OnEstablish(serversess *frame.SSServerSession) {
	elog.InfoAf("BattleServer OnEstablish Remote [ID=%v,Type=%v,Ip=%v] ", serversess.GetRemoteServerID(), serversess.GetRemoteServerType(), serversess.GetRemoteOuter())
}

func (b *BattleServer) OnTerminate(serversess *frame.SSServerSession) {
	elog.InfoAf("BattleServer OnTerminate Remote [ID=%v,Type=%v,Ip=%v] ", serversess.GetRemoteServerID(), serversess.GetRemoteServerType(), serversess.GetRemoteOuter())
}
