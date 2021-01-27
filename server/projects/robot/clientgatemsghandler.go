package main

import (
	"projects/frame"
	"projects/go-engine/elog"
	"projects/go-engine/etimer"
	"projects/pb"
)

type ClientGateFunc func(datas []byte, sess *frame.CSClientSession) bool

type ClientGateMsgHandler struct {
	dealer        *frame.IDDealer
	timerRegister etimer.ITimerRegister
}

func (c *ClientGateMsgHandler) Init() bool {
	c.timerRegister = etimer.NewTimerRegister()

	return true
}

func (c *ClientGateMsgHandler) OnHandler(msgID uint32, datas []byte, sess *frame.CSClientSession) {
	defer func() {
		if err := recover(); err != nil {
			elog.ErrorAf("Robot ClientGateMsgHandler onHandler MsgID = %v Error", msgID)
		}
	}()

	dealer := c.dealer.FindHandler(msgID)
	if dealer == nil {
		elog.ErrorAf("Robot ClientGateMsgHandler Can Not Find MsgID = %v", msgID)
		return
	}

	dealer.(ClientGateFunc)(datas, sess)
}

func (c *ClientGateMsgHandler) OnConnect(sess *frame.CSClientSession) {
	GGatewayClient = sess
	elog.InfoA("Connect GatewayServer Success")

	SendCsGameLoginReq(sess)
}

func (c *ClientGateMsgHandler) OnDisconnect(sess *frame.CSClientSession) {
	GGatewayClient = nil
	elog.InfoA("GatewayServer  DisConnect")
	c.timerRegister.KillAllTimer()
}

func (c *ClientGateMsgHandler) OnBeatHeartError(sess *frame.CSClientSession) {

}

func SendCsGameLoginReq(sess *frame.CSClientSession) {
	req := pb.CsGameLoginReq{}
	req.Accountid = GScAccountLoginAck.Accountid
	req.Token = GScAccountLoginAck.Token
	req.Loginserverid = GScAccountLoginAck.LoginServerId
	sess.SendProtoMsg(uint32(pb.EClient2GameMsgId_cs_game_login_req_id), &req)
}

var GClientGateMsgHandler *ClientGateMsgHandler
var GGatewayClient *frame.CSClientSession

func init() {
	GClientGateMsgHandler = &ClientGateMsgHandler{
		dealer: frame.NewIDDealer(),
	}
	GClientGateMsgHandler.Init()
}
