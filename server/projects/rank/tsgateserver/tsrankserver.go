package main

import (
	"projects/frame"
	"projects/go-engine/elog"
	"projects/pb"

	"github.com/golang/protobuf/proto"
)

type TsRankServerFunc func(datas []byte, t *TsRankServer) bool

type TsRankServer struct {
	frame.LogicServer
	dealer *frame.IDDealer
}

func NewTsRankServer() *TsRankServer {
	rank := &TsRankServer{
		dealer: frame.NewIDDealer(),
	}
	rank.Init()
	return rank
}
func (t *TsRankServer) Init() bool {
	t.dealer.RegisterHandler(uint32(pb.TS2TSLogicMsgId_ts2ts_rg_tranfer_msg_id), TsRankServerFunc(OnHandlerTs2TsRgTranferMsg))
	return true
}

func (t *TsRankServer) OnHandler(msgID uint32, attach_datas []byte, datas []byte, sess *frame.SSServerSession) {
	defer func() {
		if err := recover(); err != nil {
			elog.ErrorAf("TsRankServer onHandler MsgID = %v Error=%v", msgID, err)
		}
	}()

	dealer := t.dealer.FindHandler(msgID)
	if dealer == nil {
		elog.ErrorAf("TsRankServer MsgHandler Can Not Find MsgID = %v", msgID)
		return
	}
	dealer.(TsRankServerFunc)(datas, t)
}

func (t *TsRankServer) OnEstablish(serversess *frame.SSServerSession) {
	elog.InfoAf("TsRankServer OnEstablish Remote [ID=%v,Type=%v,Ip=%v] ", serversess.GetRemoteServerID(), serversess.GetRemoteServerType(), serversess.GetRemoteOuter())
}

func (t *TsRankServer) OnTerminate(serversess *frame.SSServerSession) {
	elog.InfoAf("TsRankServer OnTerminate Remote [ID=%v,Type=%v,Ip=%v] ", serversess.GetRemoteServerID(), serversess.GetRemoteServerType(), serversess.GetRemoteOuter())
}

func OnHandlerTs2TsRgTranferMsg(datas []byte, t *TsRankServer) bool {
	rg_tranfer := pb.Ts2TsRgTranferMsg{}
	unmarshalErr := proto.Unmarshal(datas, &rg_tranfer)
	if unmarshalErr != nil {
		return false
	}

	gc_tranfer := pb.Ts2CGcTranferMsg{}
	gc_tranfer.Tid = rg_tranfer.Tid
	gc_tranfer.Msgid = rg_tranfer.Msgid
	gc_tranfer.Datas = rg_tranfer.Datas
	gc_tranfer.Cbid = rg_tranfer.Cbid
	frame.GSSClientSessionMgr.AsyncSendProtoMsgBySessionID(rg_tranfer.RankclientSessid, uint32(pb.C2TSLogicMsgId_ts2c_gc_tranfer_msg_id), &gc_tranfer)
	return true
}
