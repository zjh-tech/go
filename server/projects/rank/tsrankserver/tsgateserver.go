package main

import (
	"projects/frame"
	"projects/go-engine/elog"
	"projects/pb"

	"github.com/golang/protobuf/proto"
)

type TsGatewayFunc func(datas []byte, t *TsGatewayServer) bool

type TsRankClientFunc func(tid uint32, datas []byte, t *TsGatewayServer, rankClientSessId uint64, cbId uint64) bool

type TsGatewayServer struct {
	frame.LogicServer
	gatewayDealer *frame.IDDealer
	clientDealer  *frame.IDDealer
}

func NewTsGatewayServer() *TsGatewayServer {
	gateway := &TsGatewayServer{
		gatewayDealer: frame.NewIDDealer(),
		clientDealer:  frame.NewIDDealer(),
	}
	gateway.Init()
	return gateway
}

func (t *TsGatewayServer) Init() bool {
	//ts gatewayDealer
	t.gatewayDealer.RegisterHandler(uint32(pb.TS2TSLogicMsgId_ts2ts_gr_tranfer_msg_id), TsGatewayFunc(OnHandlerTs2TsGrTranferMsg))

	//client gatewayDealer
	t.clientDealer.RegisterHandler(uint32(pb.C2TSLogicMsgId_c2ts_update_rank_req_id), TsRankClientFunc(OnHandlerC2TsUpdateRankReq))
	t.clientDealer.RegisterHandler(uint32(pb.C2TSLogicMsgId_c2ts_query_rank_req_id), TsRankClientFunc(OnHandlerC2TsQueryRankReq))
	t.clientDealer.RegisterHandler(uint32(pb.C2TSLogicMsgId_c2ts_clear_all_req_id), TsRankClientFunc(OnHandlerC2TsClearAllReq))
	t.clientDealer.RegisterHandler(uint32(pb.C2TSLogicMsgId_c2ts_clear_player_req_id), TsRankClientFunc(OnHandlerC2TsClearPlayerReq))
	return true
}

func (t *TsGatewayServer) OnHandler(msgID uint32, attach_datas []byte, datas []byte, sess *frame.SSServerSession) {
	defer func() {
		if err := recover(); err != nil {
			elog.ErrorAf("TsGateServer onHandler MsgID = %v Error=%v", msgID, err)
		}
	}()

	gatewayDealer := t.gatewayDealer.FindHandler(msgID)
	if gatewayDealer == nil {
		elog.ErrorAf("TsGateServer Gateway MsgHandler Can Not Find MsgID = %v", msgID)
		return
	}
	gatewayDealer.(TsGatewayFunc)(datas, t)
}

func (t *TsGatewayServer) OnEstablish(serversess *frame.SSServerSession) {
	elog.InfoAf("TsGateServer OnEstablish Remote [ID=%v,Type=%v,Ip=%v] ", serversess.GetRemoteServerID(), serversess.GetRemoteServerType(), serversess.GetRemoteOuter())
}

func (t *TsGatewayServer) OnTerminate(serversess *frame.SSServerSession) {
	elog.InfoAf("TsGateServer OnTerminate Remote [ID=%v,Type=%v,Ip=%v] ", serversess.GetRemoteServerID(), serversess.GetRemoteServerType(), serversess.GetRemoteOuter())
}

func (t *TsGatewayServer) SendRankClientProtobufMsg(msgID uint32, msg proto.Message, rankClientSessId uint64, tid uint32, cbId uint64) bool {
	datas, err := proto.Marshal(msg)
	if err != nil {
		elog.ErrorAf("TsGatewayServer SendRankClientProtobufMsg Msg=%v Marshal Err %v ", msgID, err)
		return false
	}

	rg_tranfer := pb.Ts2TsRgTranferMsg{}
	rg_tranfer.Msgid = msgID
	rg_tranfer.Datas = datas
	rg_tranfer.RankclientSessid = rankClientSessId
	rg_tranfer.Tid = tid
	rg_tranfer.Cbid = cbId
	t.GetServerSession().SendProtoMsg(uint32(pb.TS2TSLogicMsgId_ts2ts_rg_tranfer_msg_id), &rg_tranfer, nil)
	return true
}

func OnHandlerTs2TsGrTranferMsg(datas []byte, t *TsGatewayServer) bool {
	gr_tranfer := pb.Ts2TsGrTranferMsg{}
	unmarshalErr := proto.Unmarshal(datas, &gr_tranfer)
	if unmarshalErr != nil {
		return false
	}

	clientDealer := t.clientDealer.FindHandler(gr_tranfer.Msgid)
	if clientDealer == nil {
		elog.ErrorAf("TsGateServer Client MsgHandler Can Not Find MsgID = %v", gr_tranfer.Msgid)
		return false
	}
	clientDealer.(TsRankClientFunc)(gr_tranfer.Tid, gr_tranfer.Datas, t, gr_tranfer.RankclientSessid, gr_tranfer.Cbid)
	return true
}

func OnHandlerC2TsUpdateRankReq(tid uint32, datas []byte, t *TsGatewayServer, rankClientSessId uint64, cbId uint64) bool {
	req := pb.C2TsUpdateRankReq{}
	unmarshalErr := proto.Unmarshal(datas, &req)
	if unmarshalErr != nil {
		return false
	}

	var AckFunc = func(errorcode uint32) {
		ack := pb.Ts2CUpdateRankAck{}
		ack.Tid = tid
		ack.Playerid = req.RankInfo.PlayerId
		ack.Errorcode = errorcode
		t.SendRankClientProtobufMsg(uint32(pb.C2TSLogicMsgId_ts2c_update_rank_ack_id), &ack, rankClientSessId, tid, cbId)
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

func OnHandlerC2TsQueryRankReq(tid uint32, datas []byte, t *TsGatewayServer, rankClientSessId uint64, cbId uint64) bool {
	req := pb.C2TsQueryRankReq{}
	unmarshalErr := proto.Unmarshal(datas, &req)
	if unmarshalErr != nil {
		return false
	}

	ack := &pb.Ts2CQueryRankAck{}
	var AckFunc = func(errorcode uint32) {
		ack.Tid = tid
		ack.PlayerId = req.PlayerId
		ack.Errorcode = errorcode
		t.SendRankClientProtobufMsg(uint32(pb.C2TSLogicMsgId_ts2c_query_rank_ack_id), ack, rankClientSessId, tid, cbId)
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

func OnHandlerC2TsClearAllReq(tid uint32, datas []byte, t *TsGatewayServer, rankClientSessId uint64, cbId uint64) bool {
	req := pb.C2TsClearAllReq{}
	unmarshalErr := proto.Unmarshal(datas, &req)
	if unmarshalErr != nil {
		return false
	}

	var AckFunc = func(errorcode uint32) {
		ack := pb.Ts2CClearAllAck{}
		ack.Tid = tid
		ack.Errorcode = errorcode
		t.SendRankClientProtobufMsg(uint32(pb.C2TSLogicMsgId_ts2c_clear_all_ack_id), &ack, rankClientSessId, tid, cbId)
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

func OnHandlerC2TsClearPlayerReq(tid uint32, datas []byte, t *TsGatewayServer, rankClientSessId uint64, cbId uint64) bool {
	req := pb.C2TsClearPlayerReq{}
	unmarshalErr := proto.Unmarshal(datas, &req)
	if unmarshalErr != nil {
		return false
	}

	var AckFunc = func(errorcode uint32) {
		ack := pb.Ts2CClearPlayerAck{}
		ack.Tid = tid
		ack.PlayerId = req.PlayerId
		ack.Errorcode = errorcode
		t.SendRankClientProtobufMsg(uint32(pb.C2TSLogicMsgId_ts2c_clear_player_ack_id), &ack, rankClientSessId, tid, cbId)
	}

	rank, err := GRankMgr.FindGlobalRank(tid)
	if err != nil {
		elog.ErrorAf("[TsGateServer] C2TsClearPlayerReq Tid=%v PlayerId=%v,Error=%v", tid, req.PlayerId, err)
		AckFunc(frame.MSG_FAIL)
		return false
	}

	rank.Remove(req.PlayerId)
	AckFunc(frame.MSG_SUCCESS)

	return true
}
