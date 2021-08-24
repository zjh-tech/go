package main

import (
	"time"

	"github.com/zjh-tech/go-frame/engine/enet"
)

type ClientFunc func(datas []byte, sess *enet.CSSession) bool

type ClientMsgHandler struct {
	dealer *enet.IDDealer
}

func (c *ClientMsgHandler) Init() bool {
	c.dealer.RegisterHandler(C2STestResMsgId, ClientFunc(OnHandlerC2STestPressureRes))
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
	time.Sleep(10 * time.Millisecond)

	for i := 0; i < GCfg.LoopCount; i++ {
		SendC2STestPressureReq(sess)
	}

	// go func() {
	// 	sendTimer := time.NewTicker(1 * time.Second)
	// 	defer sendTimer.Stop()

	// 	for {
	// 		select {
	// 		case <-sendTimer.C:
	// 			{
	// 				for i := 0; i < GCfg.LoopCount; i++ {
	// 					SendC2STestPressureReq(sess)
	// 				}
	// 			}
	// 		}
	// 	}
	// }()
}

func (c *ClientMsgHandler) OnDisconnect(sess *enet.CSSession) {
}

func (c *ClientMsgHandler) OnBeatHeartError(sess *enet.CSSession) {

}
func OnHandlerC2STestPressureRes(datas []byte, sess *enet.CSSession) bool {
	// res := &C2STestPressureRes{}
	// unmarshalErr := json.Unmarshal(datas, res)
	// if unmarshalErr != nil {
	// 	return false
	// }

	// ELog.DebugAf("C2STestPressureRes=%+v", res)
	//注释掉，增加客户端的消费能力
	SendC2STestPressureReq(sess)
	return true
}

func SendC2STestPressureReq(sess *enet.CSSession) {
	req := &C2STestPressureReq{}
	req.StrValue = "test string value"
	req.Uint64Value = 100
	sess.SendJsonMsg(C2STestReqMsgId, req)
	ELog.DebugAf("C2STestPressureReq=%+v", req)
}

var GClientMsgHandler *ClientMsgHandler

func init() {
	GClientMsgHandler = &ClientMsgHandler{
		dealer: enet.NewIDDealer(),
	}
	GClientMsgHandler.Init()
}
