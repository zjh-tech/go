package main

import (
	"projects/engine/etimer"
	"projects/frame"
)

type GameClientFunc func(datas []byte, sess *frame.CSSession) bool

type ClientMsgHandler struct {
	dealer        *frame.IDDealer
	timerRegister etimer.ITimerRegister
	Qps           uint64
}

func (c *ClientMsgHandler) Init() bool {
	c.timerRegister = etimer.NewTimerRegister()
	c.dealer.RegisterHandler(uint32(10), GameClientFunc(OnHandlerCsTestReq))

	c.timerRegister.AddRepeatTimer(uint32(1), 1000, "", func(v ...interface{}) {
		ELog.InfoAf("Qps=%v", c.Qps)
		c.Qps = 0
	}, []interface{}{}, true)

	return true
}

func (c *ClientMsgHandler) OnHandler(msgID uint32, datas []byte, sess *frame.CSSession) {
	defer func() {
		if err := recover(); err != nil {
			ELog.ErrorAf("ClientMsgHandler onHandler MsgID = %v Error=%v", msgID, err)
		}
	}()

	dealer := c.dealer.FindHandler(msgID)
	if dealer != nil {
		dealer.(GameClientFunc)(datas, sess)
		return
	}
}

func (c *ClientMsgHandler) OnConnect(sess *frame.CSSession) {

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

	// GClientMsgHandler.Qps++

	// res := pb.CsGameLoginReq{}
	// res.Accountid = 1
	// res.Token = []byte("2")
	// res.Loginserverid = 3
	// sess.SendProtoMsg(uint32(10), &res)
	return true
}

var GClientMsgHandler *ClientMsgHandler

func init() {
	GClientMsgHandler = &ClientMsgHandler{
		dealer: frame.NewIDDealer(),
	}
}
