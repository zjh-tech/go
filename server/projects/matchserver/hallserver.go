package main

import (
	"projects/frame"
	"projects/go-engine/elog"
)

type HallFunc func(datas []byte, h *HallServer) bool

type HallServer struct {
	frame.LogicServer
	dealer *frame.IDDealer
}

func NewHallServer() *HallServer {
	hall := &HallServer{
		dealer: frame.NewIDDealer(),
	}
	hall.Init()
	return hall
}

func (h *HallServer) Init() bool {
	return true
}

func (h *HallServer) OnHandler(msgID uint32, attach_datas []byte, datas []byte, sess *frame.SSServerSession) {
	defer func() {
		if err := recover(); err != nil {
			elog.ErrorAf("HallServer onHandler MsgID = %v Error=%v", msgID, err)
		}
	}()

	dealer := h.dealer.FindHandler(msgID)
	if dealer == nil {
		elog.ErrorAf("HallServer MsgHandler Can Not Find MsgID = %v", msgID)
		return
	}

	dealer.(HallFunc)(datas, h)
}

func (h *HallServer) OnEstablish(serversess *frame.SSServerSession) {
	elog.InfoAf("HallServer OnEstablish Remote [ID=%v,Type=%v,Ip=%v] ", serversess.GetRemoteServerID(), serversess.GetRemoteServerType(), serversess.GetRemoteOuter())
}

func (h *HallServer) OnTerminate(serversess *frame.SSServerSession) {
	elog.InfoAf("HallServer OnTerminate Remote [ID=%v,Type=%v,Ip=%v] ", serversess.GetRemoteServerID(), serversess.GetRemoteServerType(), serversess.GetRemoteOuter())
}
