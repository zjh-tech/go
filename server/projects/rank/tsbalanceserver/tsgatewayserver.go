package main

import (
	"projects/frame"
	"projects/go-engine/elog"
	"projects/pb"

	"github.com/golang/protobuf/proto"
)

type TsGatewayServerFunc func(datas []byte, t *TsGatewayServer) bool

type TsGatewayServer struct {
	frame.LogicServer
	dealer *frame.IDDealer

	RemoteAddr      string
	ClientConnCount uint32
	Token           string
}

func NewTsGatewayServer() *TsGatewayServer {
	gateway := &TsGatewayServer{
		dealer: frame.NewIDDealer(),
	}
	gateway.Init()
	return gateway
}
func (t *TsGatewayServer) Init() bool {
	t.dealer.RegisterHandler(uint32(pb.TS2TSLogicMsgId_ts2ts_gateway_info_ntf_id), TsGatewayServerFunc(OnHandlerTs2TsGatewayInfoNtf))
	return true
}

func (t *TsGatewayServer) OnHandler(msgID uint32, attach_datas []byte, datas []byte, sess *frame.SSServerSession) {
	defer func() {
		if err := recover(); err != nil {
			elog.ErrorAf("TsGatewayServer onHandler MsgID = %v Error=%v", msgID, err)
		}
	}()

	dealer := t.dealer.FindHandler(msgID)
	if dealer == nil {
		elog.ErrorAf("TsGatewayServer MsgHandler Can Not Find MsgID = %v", msgID)
		return
	}
	dealer.(TsGatewayServerFunc)(datas, t)
}

func (t *TsGatewayServer) OnEstablish(serversess *frame.SSServerSession) {
	elog.InfoAf("TsGatewayServer OnEstablish Remote [ID=%v,Type=%v,Ip=%v] ", serversess.GetRemoteServerID(), serversess.GetRemoteServerType(), serversess.GetRemoteOuter())
}

func (t *TsGatewayServer) OnTerminate(serversess *frame.SSServerSession) {
	elog.InfoAf("TsGatewayServer OnTerminate Remote [ID=%v,Type=%v,Ip=%v] ", serversess.GetRemoteServerID(), serversess.GetRemoteServerType(), serversess.GetRemoteOuter())
}

func (t *TsGatewayServer) GetClientConnCount() uint32 {
	return t.ClientConnCount
}

func OnHandlerTs2TsGatewayInfoNtf(datas []byte, t *TsGatewayServer) bool {
	ntf := pb.Ts2TsGatewayInfoNtf{}
	unmarshalErr := proto.Unmarshal(datas, &ntf)
	if unmarshalErr != nil {
		return false
	}

	t.ClientConnCount = ntf.ClientConnCount
	t.RemoteAddr = ntf.RemoteAddr
	t.Token = ntf.Token
	elog.InfoAf("TsGatewayServer RemoteAddr=%v,ClientConnCount=%v", t.RemoteAddr, t.ClientConnCount)

	return true
}
