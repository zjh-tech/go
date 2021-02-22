package main

import (
	"projects/frame"
	"projects/go-engine/elog"
	"projects/go-engine/etimer"
	"projects/pb"

	"github.com/golang/protobuf/proto"
)

type ClientFunc func(datas []byte, sess *frame.CSClientSession) bool

type ClientMsgHandler struct {
	dealer        *frame.IDDealer
	timerRegister etimer.ITimerRegister
}

func (c *ClientMsgHandler) Init() bool {
	c.timerRegister = etimer.NewTimerRegister()
	c.dealer.RegisterHandler(uint32(10), ClientFunc(OnHandlerCsTestReq))
	return true
}

func (c *ClientMsgHandler) OnHandler(msgID uint32, datas []byte, sess *frame.CSClientSession) {
	defer func() {
		if err := recover(); err != nil {
			elog.ErrorAf("ClientGateMsgHandler onHandler MsgID = %v Error", msgID)
		}
	}()

	dealer := c.dealer.FindHandler(msgID)
	if dealer == nil {
		elog.ErrorAf("ClientGateMsgHandler Can Not Find MsgID = %v", msgID)
		return
	}

	dealer.(ClientFunc)(datas, sess)
}

func (c *ClientMsgHandler) OnConnect(sess *frame.CSClientSession) {
	elog.InfoA("Connect  Success")
	for i := 0; i < 100000; i++ {
		SendCsTestReq(sess)
	}
}

func (c *ClientMsgHandler) OnDisconnect(sess *frame.CSClientSession) {
}

func (c *ClientMsgHandler) OnBeatHeartError(sess *frame.CSClientSession) {

}
func OnHandlerCsTestReq(datas []byte, sess *frame.CSClientSession) bool {
	req := pb.CsGameLoginReq{}
	unmarshalErr := proto.Unmarshal(datas, &req)
	if unmarshalErr != nil {
		return false
	}
	//elog.InfoAf("TestReq AccountID=%v Token=%v LoginServerID=%v", req.Accountid, req.Token, req.Loginserverid)

	res := pb.CsGameLoginReq{}
	res.Accountid = 1
	res.Token = []byte("2")
	res.Loginserverid = 3
	sess.SendProtoMsg(uint32(10), &res)
	return true
}

func SendCsTestReq(sess *frame.CSClientSession) {
	req := pb.CsGameLoginReq{}
	req.Accountid = 1
	req.Token = []byte("2")
	req.Loginserverid = 3
	sess.SendProtoMsg(uint32(10), &req)
}

var GClientMsgHandler *ClientMsgHandler

func init() {
	GClientMsgHandler = &ClientMsgHandler{
		dealer: frame.NewIDDealer(),
	}
	GClientMsgHandler.Init()
}
