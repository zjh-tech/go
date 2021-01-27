package main

import (
	"projects/frame"
	"projects/go-engine/elog"
	"projects/pb"
	"projects/rank/tscommon"

	"github.com/golang/protobuf/proto"
)

type TsRankClientFunc func(datas []byte, sess *frame.SSClientSession) bool

type TsRankClient struct {
	frame.LogicServer
	dealer *frame.IDDealer
}

func NewTsRankClient() *TsRankClient {
	client := &TsRankClient{
		dealer: frame.NewIDDealer(),
	}
	return client
}

func (t *TsRankClient) Init() bool {
	t.dealer.RegisterHandler(uint32(pb.C2TSLogicMsgId_c2ts_tsgate_verify_req_id), TsRankClientFunc(OnHandlerC2TsTsgateVerifyReq))
	t.dealer.RegisterHandler(uint32(pb.C2TSLogicMsgId_c2ts_cg_tranfer_msg_id), TsRankClientFunc(OnHandlerC2TsCgTranferMsg))
	return true
}

func (t *TsRankClient) OnHandler(msgID uint32, datas []byte, sess *frame.SSClientSession) {
	defer func() {
		if err := recover(); err != nil {
			elog.ErrorAf("TsRankClient onHandler MsgID = %v Error=%v", msgID, err)
		}
	}()

	dealer := t.dealer.FindHandler(msgID)
	if dealer == nil {
		elog.ErrorAf("TsRankClient MsgHandler Can Not Find MsgID = %v", msgID)
		return
	}
	dealer.(TsRankClientFunc)(datas, sess)
}

func (c *TsRankClient) OnConnect(sess *frame.SSClientSession) {
	elog.InfoAf("[TsRankClient] SessId=%v OnConnect", sess.GetSessID())
}

func (c *TsRankClient) OnDisconnect(sess *frame.SSClientSession) {
	elog.InfoAf("[TsRankClient] SessId=%v OnDisconnect", sess.GetSessID())
}

func OnHandlerC2TsTsgateVerifyReq(datas []byte, sess *frame.SSClientSession) bool {
	req := pb.C2TsTsgateVerifyReq{}
	unmarshalErr := proto.Unmarshal(datas, &req)
	if unmarshalErr != nil {
		return false
	}

	if req.GatewayToken != frame.GServer.GetLocalToken() {
		elog.ErrorAf("[TsRankClient] SessionID=%v Verfiy Error", sess.GetSessID())
		sess.Terminate()
		return false
	}

	ack := pb.Ts2CTsgateVerifyAck{}
	sess.AsyncSendProtoMsg(uint32(pb.C2TSLogicMsgId_ts2c_tsgate_verify_ack_id), &ack, nil)

	return true
}

func OnHandlerC2TsCgTranferMsg(datas []byte, sess *frame.SSClientSession) bool {
	cg_tranfer := pb.C2TsCgTranferMsg{}
	unmarshalErr := proto.Unmarshal(datas, &cg_tranfer)
	if unmarshalErr != nil {
		return false
	}

	cfg, ok := tscommon.GRankCfg.RankAtrrMap[cg_tranfer.Tid]
	if !ok {
		elog.WarnAf("[TsRankClient] Not Find Tid=%v Cfg", cg_tranfer.Tid)
		return false
	}

	gr_tranfer := pb.Ts2TsGrTranferMsg{}
	gr_tranfer.Msgid = cg_tranfer.Msgid
	gr_tranfer.Datas = cg_tranfer.Datas
	gr_tranfer.Tid = cg_tranfer.Tid
	gr_tranfer.RankclientSessid = sess.GetSessID()
	gr_tranfer.Cbid = cg_tranfer.Cbid
	if frame.GSSServerSessionMgr.SendProtoMsg(cfg.RankServerId, uint32(pb.TS2TSLogicMsgId_ts2ts_gr_tranfer_msg_id), &gr_tranfer) == false {
		elog.WarnAf("[TsRankClient] RankServerId=%v Error", cfg.RankServerId)
		return false
	}
	return true
}

var GTsRankClient *TsRankClient

func init() {
	GTsRankClient = NewTsRankClient()
	GTsRankClient.Init()
}
