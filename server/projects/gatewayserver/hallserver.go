package main

import (
	"projects/frame"
	"projects/go-engine/elog"
	"projects/pb"

	"github.com/golang/protobuf/proto"
)

type HallFunc func(datas []byte, h *HallServer) bool

type HallServer struct {
	frame.LogicServer
	dealer           *frame.IDDealer
	playerTotalCount uint64
}

func NewHallServer() *HallServer {
	hall := &HallServer{
		dealer: frame.NewIDDealer(),
	}
	hall.Init()
	return hall
}

func (h *HallServer) Init() bool {
	h.dealer.RegisterHandler(uint32(pb.S2SLogicMsgId_s2s_hgl_kick_player_ack_id), HallFunc(OnHandlerS2SHglKickPlayerAck))
	h.dealer.RegisterHandler(uint32(pb.S2SLogicMsgId_hg_select_hall_ack_id), HallFunc(OnHandlerHgSelectHallAck))
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

func (h *HallServer) GetPlayerTotalCount() uint64 {
	return h.playerTotalCount
}

func OnHandlerHgSelectHallAck(datas []byte, h *HallServer) bool {
	hallAck := pb.HgSelectHallAck{}
	err := proto.Unmarshal(datas, &hallAck)
	if err != nil {
		return false
	}

	player := GPlayerMgr.FindPlayer(hallAck.Playeruid)
	if player == nil {
		elog.WarnAf("[Login] HgSelectHallAck Find Player %v Error", hallAck.Playeruid)
		return false
	}
	player.SetAccountId(hallAck.Accountid)
	player.SetTranferFlag()
	return true
}

func OnHandlerS2SHglKickPlayerAck(datas []byte, h *HallServer) bool {
	loginAck := pb.S2SHglKickPlayerAck{}
	err := proto.Unmarshal(datas, &loginAck)
	if err != nil {
		return false
	}

	frame.GSSServerSessionMgr.SendProtoMsg(loginAck.Loginsrvid, uint32(pb.S2SLogicMsgId_s2s_hgl_kick_player_ack_id), &loginAck)
	elog.InfoAf("[Login] AccountId=%v Send2Login S2SHglKickPlayerAck ", loginAck.Accountid)
	return true
}
