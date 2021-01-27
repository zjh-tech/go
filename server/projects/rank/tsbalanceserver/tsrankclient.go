package main

import (
	"projects/frame"
	"projects/go-engine/elog"
	"projects/pb"

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
	t.dealer.RegisterHandler(uint32(pb.C2TSLogicMsgId_c2ts_select_tsgate_req_id), TsRankClientFunc(OnHandlerC2TsSelectTsgateReq))
	return true
}

func (t *TsRankClient) OnHandler(msgID uint32, datas []byte, sess *frame.SSClientSession) {
	//defer func() {
	//	if err := recover(); err != nil {
	//		elog.ErrorAf("TsRankClient onHandler MsgID = %v Error=%v", msgID, err)
	//	}
	//}()

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

func GetMinLoadTsGateway() *TsGatewayServer {
	var minLoadGateway *TsGatewayServer = nil
	minClientCount := uint32(0)
	tsgateways := frame.GSSServerSessionMgr.FindLogicServerByServerType(uint32(frame.TS_RANK_GATEWAY_SERVER_TYPE))
	for _, logicServer := range tsgateways {
		gateway := logicServer.(*TsGatewayServer)
		if minClientCount == 0 {
			minClientCount = gateway.GetClientConnCount()
			minLoadGateway = gateway
		} else if minClientCount < gateway.GetClientConnCount() {
			minClientCount = gateway.GetClientConnCount()
			minLoadGateway = gateway
		}
	}

	return minLoadGateway
}

func OnHandlerC2TsSelectTsgateReq(datas []byte, sess *frame.SSClientSession) bool {
	req := pb.C2TsSelectTsgateReq{}
	unmarshalErr := proto.Unmarshal(datas, &req)
	if unmarshalErr != nil {
		return false
	}

	tsgateway := GetMinLoadTsGateway()

	ack := pb.Ts2CSelectTsgateAck{}
	if tsgateway == nil {
		elog.Errorf("MinLoadTsGateway Error")
	} else {
		ack.RemoteAddr = tsgateway.RemoteAddr
		ack.GatewayToken = tsgateway.Token
		elog.InfoAf("Select MinLoadTsGateway RemoteServerID=%v RemoteServerType=%v RemoteAddr=%v ",
			tsgateway.GetServerSession().GetRemoteServerID(), tsgateway.GetServerSession().GetRemoteServerType(),
			tsgateway.RemoteAddr)
	}
	sess.AsyncSendProtoMsg(uint32(pb.C2TSLogicMsgId_ts2c_select_tsgate_ack_id), &ack, nil)

	return true
}

var GTsRankClient *TsRankClient

func init() {
	GTsRankClient = NewTsRankClient()
	GTsRankClient.Init()
}
