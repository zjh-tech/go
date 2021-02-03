package tsrankclient

import (
	"projects/frame"
	"projects/go-engine/elog"
	"projects/go-engine/etimer"
	"projects/pb"

	"github.com/golang/protobuf/proto"
)

type TsRankBalanceFunc func(datas []byte, sess *frame.SSClientSession) bool

type RankBalanceMsgHandler struct {
	dealer        *frame.IDDealer
	timerRegister etimer.ITimerRegister
}

func (r *RankBalanceMsgHandler) Init() bool {
	r.dealer.RegisterHandler(uint32(pb.C2TSLogicMsgId_ts2c_select_tsgate_ack_id), TsRankBalanceFunc(OnHandlerTs2CSelectTsgateAck))
	return true
}

func (r *RankBalanceMsgHandler) OnHandler(msgID uint32, datas []byte, sess *frame.SSClientSession) {
	defer func() {
		if err := recover(); err != nil {
			elog.ErrorAf("RankBalanceMsgHandler onHandler MsgID = %v Error=%v", msgID, err)
		}
	}()

	dealer := r.dealer.FindHandler(msgID)
	if dealer == nil {
		elog.ErrorAf("RankBalanceMsgHandler MsgHandler Can Not Find MsgID = %v", msgID)
		return
	}

	dealer.(TsRankBalanceFunc)(datas, sess)
}

func (r *RankBalanceMsgHandler) OnConnect(sess *frame.SSClientSession) {
	elog.InfoAf("[RankBalanceMsgHandler] SessId=%v OnConnect", sess.GetSessID())
	SendC2TsSelectTsgateReq(sess)
}

func (r *RankBalanceMsgHandler) OnDisconnect(sess *frame.SSClientSession) {
	elog.InfoAf("[RankBalanceMsgHandler] SessId=%v OnDisconnect", sess.GetSessID())
	r.timerRegister.KillAllTimer()
	GTsRankClient.tsBalanceSessID = 0
}

func (r *RankBalanceMsgHandler) OnBeatHeartError(sess *frame.SSClientSession) {

}

var GRankBalanceMsgHandler *RankBalanceMsgHandler

func init() {
	GRankBalanceMsgHandler = &RankBalanceMsgHandler{
		dealer:        frame.NewIDDealer(),
		timerRegister: etimer.NewTimerRegister(),
	}
	GRankBalanceMsgHandler.Init()
}

func OnHandlerTs2CSelectTsgateAck(datas []byte, sess *frame.SSClientSession) bool {
	ack := pb.Ts2CSelectTsgateAck{}
	unmarshalErr := proto.Unmarshal(datas, &ack)
	if unmarshalErr != nil {
		return false
	}

	sess.Terminate()
	if ack.RemoteAddr == "" || ack.GatewayToken == "" {
		elog.ErrorA("[RankBalanceMsgHandler] Select TsGateway Error")
		return false
	}

	GTsRankClient.tsGatewayToken = ack.GatewayToken
	elog.InfoAf("[RankBalanceMsgHandler] Connect TsGateway RemoteAddr=%v", ack.RemoteAddr)
	GTsRankClient.tsGatewaySessID = frame.GSSClientSessionMgr.SSClientConnect(ack.RemoteAddr, GRankGatewayMsgHandler, nil)
	return true
}

func SendC2TsSelectTsgateReq(sess *frame.SSClientSession) {
	req := pb.C2TsSelectTsgateReq{}
	req.BalanceToken = GTsRankClient.tsBalanceToken
	sess.SendProtoMsg(uint32(pb.C2TSLogicMsgId_c2ts_select_tsgate_req_id), &req, nil)
}
