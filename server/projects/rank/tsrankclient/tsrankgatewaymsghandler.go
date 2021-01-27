package tsrankclient

import (
	"projects/frame"
	"projects/go-engine/elog"
	"projects/go-engine/etimer"
	"projects/pb"

	"github.com/golang/protobuf/proto"
)

type TsRankGatewayFunc func(datas []byte, sess *frame.SSClientSession) bool

type RankGatewayMsgHandler struct {
	dealer        *frame.IDDealer
	timerRegister etimer.ITimerRegister
}

func (r *RankGatewayMsgHandler) Init() bool {
	r.dealer.RegisterHandler(uint32(pb.C2TSLogicMsgId_ts2c_tsgate_verify_ack_id), TsRankGatewayFunc(OnHandlerTs2CTsgateVerifyAck))
	r.dealer.RegisterHandler(uint32(pb.C2TSLogicMsgId_ts2c_gc_tranfer_msg_id), TsRankGatewayFunc(OnHandlerTs2CGcTranferMsg))
	return true
}

func (r *RankGatewayMsgHandler) OnHandler(msgID uint32, datas []byte, sess *frame.SSClientSession) {
	defer func() {
		if err := recover(); err != nil {
			elog.ErrorAf("RankGatewayMsgHandler onHandler MsgID = %v Error=%v", msgID, err)
		}
	}()

	dealer := r.dealer.FindHandler(msgID)
	if dealer == nil {
		elog.ErrorAf("RankGatewayMsgHandler MsgHandler Can Not Find MsgID = %v", msgID)
		return
	}

	dealer.(TsRankGatewayFunc)(datas, sess)
}

func (r *RankGatewayMsgHandler) OnConnect(sess *frame.SSClientSession) {
	elog.InfoAf("[RankGatewayMsgHandler] SessId=%v OnConnect", sess.GetSessID())
	GRankGatewaySession = sess

	req := pb.C2TsTsgateVerifyReq{}
	req.GatewayToken = GTsRankClient.tsGatewayToken
	sess.SendProtoMsg(uint32(pb.C2TSLogicMsgId_c2ts_tsgate_verify_req_id), &req, nil)
}

func (r *RankGatewayMsgHandler) OnDisconnect(sess *frame.SSClientSession) {
	elog.InfoAf("[RankGatewayMsgHandler] SessId=%v OnDisconnect", sess.GetSessID())
	GRankGatewaySession = nil
	r.timerRegister.KillAllTimer()
}

func OnHandlerTs2CTsgateVerifyAck(datas []byte, sess *frame.SSClientSession) bool {
	GTsRankClient.tsGatewayVerifyOk = true
	elog.InfoA("[RankGatewayMsgHandler] Verify Ok")
	return true
}

func OnHandlerTs2CGcTranferMsg(datas []byte, sess *frame.SSClientSession) bool {
	gc_tranfer := pb.Ts2CGcTranferMsg{}
	unmarshalErr := proto.Unmarshal(datas, &gc_tranfer)
	if unmarshalErr != nil {
		return false
	}

	cbMap := GTsRankClient.GetCbMap()
	cbItem, ok := cbMap[gc_tranfer.Cbid]
	if !ok {
		elog.ErrorAf("[RankGatewayMsgHandler] Not Find CbId=%v", gc_tranfer.Cbid)
		return false
	}

	if cbItem.Func != nil {
		cbItem.Func(gc_tranfer.Tid, gc_tranfer.Datas, cbItem.Agrs)
	}

	return true
}

var GRankGatewayMsgHandler *RankGatewayMsgHandler
var GRankGatewaySession *frame.SSClientSession

func init() {
	GRankGatewayMsgHandler = &RankGatewayMsgHandler{
		dealer:        frame.NewIDDealer(),
		timerRegister: etimer.NewTimerRegister(),
	}
	GRankGatewayMsgHandler.Init()
}
