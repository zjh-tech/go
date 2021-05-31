package main

import (
	"github.com/zjh-tech/go-frame/engine/enet"
	"github.com/zjh-tech/go-frame/engine/etimer"
)

type ClientFunc func(datas []byte, sess *enet.CSSession) bool

type ClientMsgHandler struct {
	dealer        *enet.IDDealer
	timerRegister etimer.ITimerRegister
}

func (c *ClientMsgHandler) Init() bool {
	c.timerRegister = etimer.NewTimerRegister(etimer.GTimerMgr)
	c.dealer.RegisterHandler(uint32(10), ClientFunc(OnHandlerCsTestReq))
	return true
}

func (c *ClientMsgHandler) OnHandler(msgId uint32, datas []byte, sess *enet.CSSession) {
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

func (c *ClientMsgHandler) OnConnect(sess *enet.CSSession) {
	ELog.InfoA("Connect  Success")
	for i := 0; i < 100000; i++ {
		SendCsTestReq(sess)
	}
}

func (c *ClientMsgHandler) OnDisconnect(sess *enet.CSSession) {
}

func (c *ClientMsgHandler) OnBeatHeartError(sess *enet.CSSession) {

}
func OnHandlerCsTestReq(datas []byte, sess *enet.CSSession) bool {
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

func SendCsTestReq(sess *enet.CSSession) {
	// req := pb.CsGameLoginReq{}
	// req.Accountid = 1
	// req.Token = []byte("2")
	// req.Loginserverid = 3
	// sess.SendProtoMsg(uint32(10), &req)
}

var GClientMsgHandler *ClientMsgHandler

func init() {
	GClientMsgHandler = &ClientMsgHandler{
		dealer: enet.NewIDDealer(),
	}
	GClientMsgHandler.Init()
}
