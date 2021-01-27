package main

import (
	"projects/frame"
	"projects/go-engine/elog"
	"projects/pb"

	"github.com/golang/protobuf/proto"
)

type SDServerFunc func(datas []byte, s *frame.SSClientSession)

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
	s.dealer.RegisterHandler(uint32(pb.S2SBaseMsgId_service_discovery_req_id), SDServerFunc(OnHandlerServiceDiscoveryReq))
	return true
}
func (s *SDServer) OnHandler(msgID uint32, datas []byte, sess *frame.SSClientSession) {
	defer func() {
		if err := recover(); err != nil {
			elog.ErrorAf("SDServer onHandler MsgID = %v Error=%v", msgID, err)
		}
	}()

	dealer := s.dealer.FindHandler(msgID)
	if dealer == nil {
		elog.ErrorAf("SDServer MsgHandler Can Not Find MsgID = %v", msgID)
		return
	}

	dealer.(SDServerFunc)(datas, sess)
}

func (s *SDServer) OnConnect(sess *frame.SSClientSession) {

}

func (s *SDServer) OnDisconnect(sess *frame.SSClientSession) {

}

func (s *SDServer) OnBeatHeartError(sess *frame.SSClientSession) {

}

func OnHandlerServiceDiscoveryReq(datas []byte, sess *frame.SSClientSession) {
	ack := &pb.ServiceDiscoveryAck{}
	ack.RebuildFlag = GServiceDiscoveryServer.RebuildFlag
	ack.VerifyFlag = false

	var AckFunc = func() {
		sess.SendProtoMsg(uint32(pb.S2SBaseMsgId_service_discovery_ack_id), ack, nil)
	}

	req := &pb.ServiceDiscoveryReq{}
	err := proto.Unmarshal(datas, req)
	if err != nil {
		AckFunc()
		elog.ErrorAf("[ServiceDiscovery] UdpReq Protobuf Unmarshal=%v", req.ServerId)
		return
	}

	s2sAttr, attOk := GRegistryCfg.S2SAttrMap[req.ServerId]
	if !attOk {
		AckFunc()
		elog.ErrorAf("[ServiceDiscovery]  RegistryCfg Not Find ServerId=%v", req.ServerId)
		return
	}

	if req.Token != s2sAttr.Token {
		AckFunc()
		elog.ErrorAf("[ServiceDiscovery]  ServerId=%v Token Error", req.ServerId)
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
		ack.SdInfo = &pb.ServiceDiscovery{}
		ack.SdInfo.ServerId = req.ServerId
		ack.SdInfo.S2SInterListen = s2sAttr.Inter
		ack.SdInfo.S2SOuterListen = s2sAttr.Outer
		FillServiceDiscoveryConn(ack.SdInfo)

		c2sAttr, c2sAttOk := GRegistryCfg.C2SAttrMap[req.ServerId]
		if c2sAttOk {
			ack.SdInfo.C2SInterListen = c2sAttr.Inter
			ack.SdInfo.C2SOuterListen = c2sAttr.Outer
			ack.SdInfo.C2SMaxCount = uint32(c2sAttr.MaxCount)
			ack.SdInfo.C2SHttpsUrl = c2sAttr.C2SHttpsUrl
			ack.SdInfo.C2SHttpsCert = c2sAttr.C2SHttpsCert
			ack.SdInfo.C2SHttpsKey = c2sAttr.C2SHttpsKey
		}
	}

	AckFunc()
	elog.InfoAf("[ServiceDiscovery] UdpServiceDiscoveryAck=%+v", ack)
}
