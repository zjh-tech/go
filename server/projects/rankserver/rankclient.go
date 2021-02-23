package main

import (
	"projects/frame"
	"projects/go-engine/elog"
	"projects/pb"

	"github.com/golang/protobuf/proto"
)

type RankClientLogicFunc func(tid uint32, datas []byte, sess *frame.SSClientSession, cbId uint64) bool

type RankClient struct {
	logicDealer *frame.IDDealer
}

func NewRankClient() *RankClient {
	gateway := &RankClient{
		logicDealer: frame.NewIDDealer(),
	}
	return gateway
}

func (r *RankClient) Init() bool {
	r.logicDealer.RegisterHandler(uint32(pb.CRLogicMsgId_c2r_update_rank_req_id), RankClientLogicFunc(OnHandlerC2RUpdateRankReq))
	r.logicDealer.RegisterHandler(uint32(pb.CRLogicMsgId_c2r_query_rank_req_id), RankClientLogicFunc(OnHandlerC2RQueryRankReq))
	r.logicDealer.RegisterHandler(uint32(pb.CRLogicMsgId_c2r_clear_all_req_id), RankClientLogicFunc(OnHandlerC2RClearAllReq))
	r.logicDealer.RegisterHandler(uint32(pb.CRLogicMsgId_c2r_clear_player_req_id), RankClientLogicFunc(OnHandlerC2RClearPlayerReq))
	return true
}

func (r *RankClient) OnHandler(msgID uint32, datas []byte, sess *frame.SSClientSession) {
	defer func() {
		if err := recover(); err != nil {
			elog.ErrorAf("OnHandler MsgID = %v Error=%v", msgID, err)
		}
	}()

	if msgID == uint32(pb.CRLogicMsgId_c2r_tranfer_msg_id) {
		OnHandlerC2RTranferMsg(datas, sess, r)
		return
	} else if msgID == uint32(pb.CRLogicMsgId_c2r_rank_verify_req_id) {
		OnHandlerC2RRankVerifyReq(datas, sess)
		return
	}
}

func (t *RankClient) OnConnect(sess *frame.SSClientSession) {

}

func (t *RankClient) OnDisconnect(sess *frame.SSClientSession) {

}

func (t *RankClient) OnBeatHeartError(sess *frame.SSClientSession) {

}

func SendRankClientProtobufMsg(sess *frame.SSClientSession, msgID uint32, msg proto.Message, tid uint32, cbId uint64) bool {
	datas, err := proto.Marshal(msg)
	if err != nil {
		elog.ErrorAf("RankClient SendRankClientProtobufMsg Msg=%v Marshal Err %v ", msgID, err)
		return false
	}

	rg_tranfer := pb.R2CTranferMsg{}
	rg_tranfer.Msgid = msgID
	rg_tranfer.Datas = datas
	rg_tranfer.Tid = tid
	rg_tranfer.Cbid = cbId
	sess.SendProtoMsg(uint32(pb.CRLogicMsgId_r2c_tranfer_msg_id), &rg_tranfer, nil)
	return true
}

func OnHandlerC2RRankVerifyReq(datas []byte, sess *frame.SSClientSession) bool {
	req := pb.C2RRankVerifyReq{}
	unmarshalErr := proto.Unmarshal(datas, &req)
	if unmarshalErr != nil {
		return false
	}

	if req.GatewayToken != frame.GServer.GetLocalToken() {
		elog.ErrorAf("[RankClient] SessionID=%v Verfiy Error", sess.GetSessID())
		sess.Terminate()
		return false
	}

	ack := pb.R2CRankVerifyAck{}
	sess.AsyncSendProtoMsg(uint32(pb.CRLogicMsgId_r2c_rank_verify_ack_id), &ack, nil)

	return true
}

func OnHandlerC2RTranferMsg(datas []byte, sess *frame.SSClientSession, r *RankClient) bool {
	cg_tranfer := pb.C2RTranferMsg{}
	unmarshalErr := proto.Unmarshal(datas, &cg_tranfer)
	if unmarshalErr != nil {
		return false
	}

	clientDealer := r.logicDealer.FindHandler(cg_tranfer.Msgid)
	if clientDealer == nil {
		elog.ErrorAf("Client Logic MsgHandler Can Not Find MsgID = %v", cg_tranfer.Msgid)
		return false
	}
	clientDealer.(RankClientLogicFunc)(cg_tranfer.Tid, cg_tranfer.Datas, sess, cg_tranfer.Cbid)
	return true
}

func OnHandlerC2RUpdateRankReq(tid uint32, datas []byte, sess *frame.SSClientSession, cbId uint64) bool {
	req := pb.C2RUpdateRankReq{}
	unmarshalErr := proto.Unmarshal(datas, &req)
	if unmarshalErr != nil {
		return false
	}

	var AckFunc = func(errorcode uint32) {
		ack := pb.R2CUpdateRankAck{}
		ack.Tid = tid
		ack.Playerid = req.RankInfo.PlayerId
		ack.Errorcode = errorcode
		SendRankClientProtobufMsg(sess, uint32(pb.CRLogicMsgId_r2c_update_rank_ack_id), &ack, tid, cbId)
	}

	var err error
	var rank *Rank
	rank, err = GRankMgr.FindGlobalRank(tid)
	if err != nil {
		elog.ErrorAf("[TsGateServer] C2TsUpdateRankReq FindGlobalRank Tid=%v Error=%v", tid, err)
		AckFunc(frame.MSG_FAIL)
		return false
	}

	if rank == nil {
		elog.ErrorAf("[TsGateServer] C2TsUpdateRankReq Tid=%v Rank=Nil Error", tid)
		AckFunc(frame.MSG_FAIL)
		return false
	}

	rank.Update(req.RankInfo, false)
	AckFunc(frame.MSG_SUCCESS)
	return true
}

func OnHandlerC2RQueryRankReq(tid uint32, datas []byte, sess *frame.SSClientSession, cbId uint64) bool {
	req := pb.C2RQueryRankReq{}
	unmarshalErr := proto.Unmarshal(datas, &req)
	if unmarshalErr != nil {
		return false
	}

	ack := &pb.R2CQueryRankAck{}
	var AckFunc = func(errorcode uint32) {
		ack.Tid = tid
		ack.PlayerId = req.PlayerId
		ack.Errorcode = errorcode
		SendRankClientProtobufMsg(sess, uint32(pb.CRLogicMsgId_r2c_query_rank_ack_id), ack, tid, cbId)
	}

	rank, err := GRankMgr.FindGlobalRank(tid)
	if err != nil {
		elog.ErrorAf("[TsGateServer] C2TsQueryRankReq Tid=%v Error=%v", tid, err)
		AckFunc(frame.MSG_FAIL)
		return false
	}

	rank.FillRankItems(req.Topn, ack)
	AckFunc(frame.MSG_SUCCESS)
	return true
}

func OnHandlerC2RClearAllReq(tid uint32, datas []byte, sess *frame.SSClientSession, cbId uint64) bool {
	req := pb.C2RClearAllReq{}
	unmarshalErr := proto.Unmarshal(datas, &req)
	if unmarshalErr != nil {
		return false
	}

	var AckFunc = func(errorcode uint32) {
		ack := pb.R2CClearAllAck{}
		ack.Tid = tid
		ack.Errorcode = errorcode
		SendRankClientProtobufMsg(sess, uint32(pb.CRLogicMsgId_r2c_clear_all_ack_id), &ack, tid, cbId)
	}

	rank, err := GRankMgr.FindGlobalRank(tid)
	if err != nil {
		elog.ErrorAf("[TsGateServer] C2TsClearAllReq Tid=%v Error=%v", tid, err)
		AckFunc(frame.MSG_FAIL)
		return false
	}

	rank.ClearAll()
	AckFunc(frame.MSG_SUCCESS)
	return true
}

func OnHandlerC2RClearPlayerReq(tid uint32, datas []byte, sess *frame.SSClientSession, cbId uint64) bool {
	req := pb.C2RClearPlayerReq{}
	unmarshalErr := proto.Unmarshal(datas, &req)
	if unmarshalErr != nil {
		return false
	}

	var AckFunc = func(errorcode uint32) {
		ack := pb.R2CClearPlayerAck{}
		ack.Tid = tid
		ack.PlayerId = req.PlayerId
		ack.Errorcode = errorcode
		SendRankClientProtobufMsg(sess, uint32(pb.CRLogicMsgId_r2c_clear_player_ack_id), &ack, tid, cbId)
	}

	rank, err := GRankMgr.FindGlobalRank(tid)
	if err != nil {
		elog.ErrorAf("C2TsClearPlayerReq Tid=%v PlayerId=%v,Error=%v", tid, req.PlayerId, err)
		AckFunc(frame.MSG_FAIL)
		return false
	}

	rank.Remove(req.PlayerId)
	AckFunc(frame.MSG_SUCCESS)

	return true
}
