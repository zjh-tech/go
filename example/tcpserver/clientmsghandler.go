package main

import (
	"encoding/json"
	"sync/atomic"
	"time"

	"github.com/zjh-tech/go-frame/engine/enet"
)

type GameClientFunc func(datas []byte, sess *enet.CSSession) bool

type ClientMsgHandler struct {
	dealer *enet.IDDealer
}

func (c *ClientMsgHandler) Init() bool {
	c.dealer.RegisterHandler(C2STestReqMsgId, GameClientFunc(OnHandlerC2STestPressureReq))

	return true
}

func (c *ClientMsgHandler) OnHandler(msgId uint32, datas []byte, sess *enet.CSSession) {
	defer func() {
		if err := recover(); err != nil {
			ELog.ErrorAf("ClientMsgHandler onHandler MsgID = %v Error=%v", msgId, err)
		}
	}()

	dealer := c.dealer.FindHandler(msgId)
	if dealer != nil {
		dealer.(GameClientFunc)(datas, sess)
		return
	}
}

func (c *ClientMsgHandler) OnConnect(sess *enet.CSSession) {
	ELog.InfoAf("OnConnect SessionID=%v", sess.GetSessID())
}

func (c *ClientMsgHandler) OnDisconnect(sess *enet.CSSession) {
	ELog.InfoAf("OnDisconnect SessionID=%v", sess.GetSessID())
}

func (c *ClientMsgHandler) OnBeatHeartError(sess *enet.CSSession) {

}

func OnHandlerC2STestPressureReq(datas []byte, sess *enet.CSSession) bool {
	req := &C2STestPressureReq{}
	unmarshalErr := json.Unmarshal(datas, &req)
	if unmarshalErr != nil {
		return false
	}
	//ELog.DebugAf("C2STestPressureReq=%+v", req)

	atomic.AddInt64(&GRecvLogicQps, 1)

	res := &C2STestPressureRes{
		StrValue:    req.StrValue,
		Uint64Value: req.Uint64Value,
	}
	sess.SendJsonMsg(C2STestResMsgId, res)
	atomic.AddInt64(&GSendLogicQps, 1)
	//ELog.DebugAf("C2STestPressureRes=%+v", res)
	return true
}

var GSendLogicQps int64
var GRecvLogicQps int64

func PrintQps() {
	go func() {
		qpsTimer := time.NewTicker(1 * time.Second)
		defer qpsTimer.Stop()

		for {
			select {
			case <-qpsTimer.C:
				{
					ELog.InfoAf("RecvLogicQps=%v RecvQps=%v SendLogicQps=%v SendQps=%v", GRecvLogicQps, enet.GRecvQps, GSendLogicQps, enet.GSendQps)
					atomic.StoreInt64(&GRecvLogicQps, 0)
					atomic.StoreInt64(&enet.GRecvQps, 0)
					atomic.StoreInt64(&GSendLogicQps, 0)
					atomic.StoreInt64(&enet.GSendQps, 0)
				}
			}
		}
	}()
}

var GClientMsgHandler *ClientMsgHandler

func init() {
	GClientMsgHandler = &ClientMsgHandler{
		dealer: enet.NewIDDealer(),
	}

	GClientMsgHandler.Init()
}
