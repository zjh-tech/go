package main

import (
	"github.com/golang/protobuf/proto"
	"github.com/zjh-tech/go-frame/frame"
	"github.com/zjh-tech/go-frame/frame/framepb"
)

type SDServerFunc func(datas []byte, s *frame.SDKSession)

type SDServer struct {
	dealer *frame.IDDealer
}

func NewSDServer() *SDServer {
	sdserver := &SDServer{
		dealer: frame.NewIDDealer(),
	}
	sdserver.Init()
	return sdserver
}

func (s *SDServer) Init() bool {
	s.dealer.RegisterHandler(uint32(framepb.S2SBaseMsgId_service_discovery_req_id), SDServerFunc(OnHandlerServiceDiscoveryReq))
	return true
}
func (s *SDServer) OnHandler(msgID uint32, datas []byte, sess *frame.SDKSession) {
	defer func() {
		if err := recover(); err != nil {
			ELog.ErrorAf("SDServer onHandler MsgID = %v Error=%v", msgID, err)
		}
	}()

	dealer := s.dealer.FindHandler(msgID)
	if dealer == nil {
		ELog.ErrorAf("SDServer MsgHandler Can Not Find MsgID = %v", msgID)
		return
	}

	dealer.(SDServerFunc)(datas, sess)
}

func (s *SDServer) OnConnect(sess *frame.SDKSession) {

}

func (s *SDServer) OnDisconnect(sess *frame.SDKSession) {

}

func (s *SDServer) OnBeatHeartError(sess *frame.SDKSession) {

}

func OnHandlerServiceDiscoveryReq(datas []byte, sess *frame.SDKSession) {
	ack := &framepb.ServiceDiscoveryAck{}
	ack.RebuildFlag = GServiceDiscoveryServer.RebuildFlag
	ack.VerifyFlag = false

	var AckFunc = func() {
		sess.AsyncSendProtoMsg(uint32(framepb.S2SBaseMsgId_service_discovery_ack_id), ack)
	}

	req := &framepb.ServiceDiscoveryReq{}
	err := proto.Unmarshal(datas, req)
	if err != nil {
		AckFunc()
		ELog.ErrorAf("[ServiceDiscovery] UdpReq Protobuf Unmarshal=%v", req.ServerId)
		return
	}

	Attr, attOk := GRegistryCfg.AttrMap[req.ServerId]
	if !attOk {
		AckFunc()
		ELog.ErrorAf("[ServiceDiscovery]  RegistryCfg Not Find ServerId=%v", req.ServerId)
		return
	}

	if req.Token != Attr.Token {
		AckFunc()
		ELog.ErrorAf("[ServiceDiscovery]  ServerId=%v Token Error", req.ServerId)
		return
	}

	ack.VerifyFlag = true

	GServiceDiscoveryServer.Mutex.Lock()
	defer GServiceDiscoveryServer.Mutex.Unlock()

	_, usedOk := GServiceDiscoveryServer.UseServices[req.ServerId]
	if !usedOk {
		GServiceDiscoveryServer.AddUsedService(req.ServerId)
	}

	GServiceDiscoveryServer.RemoveWarnService(req.ServerId)
	if GServiceDiscoveryServer.RebuildFlag == false {
		ack.SdInfo = &framepb.ServiceDiscovery{}
		ack.SdInfo.ServerId = req.ServerId
		ack.SdInfo.S2SInterListen = Attr.S2S_TCP_Inter
		ack.SdInfo.S2SOuterListen = Attr.S2S_TCP_Outer
		ack.SdInfo.S2SHttpSurl = Attr.S2S_Http_SUrl
		ack.SdInfo.S2SHttpCurl1 = Attr.S2S_Http_CUrl1
		ack.SdInfo.S2SHttpCurl2 = Attr.S2S_Http_CUrl2
		FillServiceDiscoveryConn(ack.SdInfo)

		ack.SdInfo.C2SInterListen = Attr.C2S_TCP_Inter
		ack.SdInfo.C2SOuterListen = Attr.C2S_TCP_Outer
		ack.SdInfo.C2SHttpsUrl = Attr.C2SHttpsUrl
		ack.SdInfo.C2SHttpsCert = Attr.C2SHttpsCert
		ack.SdInfo.C2SHttpsKey = Attr.C2SHttpsKey

		ack.SdInfo.SdkTcpInter = Attr.SDK_TCP_Inter
		ack.SdInfo.SdkTcpOut = Attr.SDK_TCP_Outer
		ack.SdInfo.SdkHttpsUrtl = Attr.SDKHttpsUrl
		ack.SdInfo.SdkHttpsCert = Attr.SDKHttpsCert
		ack.SdInfo.SdkHttpsKey = Attr.SDKHttpsKey
	}

	AckFunc()
	ELog.InfoAf("[ServiceDiscovery] UdpServiceDiscoveryAck=%+v", ack)
}
