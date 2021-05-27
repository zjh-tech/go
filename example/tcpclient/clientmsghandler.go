package main

import (
	"github.com/zjh-tech/go-frame/engine/etimer"
	"github.com/zjh-tech/go-frame/frame"
)

type ClientFunc func(datas []byte, sess *frame.CSSession) bool

type ClientMsgHandler struct {
	dealer        *frame.IDDealer
	timerRegister etimer.ITimerRegister
}

func (c *ClientMsgHandler) Init() bool {
	c.timerRegister = etimer.NewTimerRegister()
	c.dealer.RegisterHandler(uint32(10), ClientFunc(OnHandlerCsTestReq))
	return true
}

func (c *ClientMsgHandler) OnHandler(msgId uint32, datas []byte, sess *frame.CSSession) {
	defer func() {
		if err := recover(); err != nil {
			ELog.ErrorAf("ClientGateMsgHandler onHandler MsgID = %v Error", msgId)
		}
	}()

	dealer := c.dealer.FindHandler(msgId)
	if dealer == nil {
		ELog.ErrorAf("ClientGateMsgHandler Can Not Find MsgID = %v", msgId)
		return
	}

	dealer.(ClientFunc)(datas, sess)
}

func (c *ClientMsgHandler) OnConnect(sess *frame.CSSession) {
	ELog.InfoA("Connect  Success")
	for i := 0; i < 100000; i++ {
		SendCsTestReq(sess)
	}
}

func (c *ClientMsgHandler) OnDisconnect(sess *frame.CSSession) {
}

func (c *ClientMsgHandler) OnBeatHeartError(sess *frame.CSSession) {

}
func OnHandlerCsTestReq(datas []byte, sess *frame.CSSession) bool {
	// req := pb.CsGameLoginReq{}
	// unmarshalErr := proto.Unmarshal(datas, &req)
	// if unmarshalErr != nil {
	// 	return false
	// }
	// //elog.InfoAf("TestReq AccountID=%v Token=%v LoginServerID=%v", req.Accountid, req.Token, req.Loginserverid)

	// res := pb.CsGameLoginReq{}
	// res.Accountid = 1
	// res.Token = []byte("2")
	// res.Loginserverid = 3
	// sess.SendProtoMsg(uint32(10), &res)
	return true
}

func SendCsTestReq(sess *frame.CSSession) {
	// req := pb.CsGameLoginReq{}
	// req.Accountid = 1
	// req.Token = []byte("2")
	// req.Loginserverid = 3
	// sess.SendProtoMsg(uint32(10), &req)
}

var GClientMsgHandler *ClientMsgHandler

func init() {
	GClientMsgHandler = &ClientMsgHandler{
		dealer: frame.NewIDDealer(),
	}
	GClientMsgHandler.Init()
}
