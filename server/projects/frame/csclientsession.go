package frame

import (
	"projects/go-engine/elog"
	"projects/go-engine/enet"
	"projects/go-engine/etimer"
	"projects/go-engine/inet"
	"projects/pb"
	"projects/util"

	"github.com/golang/protobuf/proto"
)

type ICSClientMsgHandler interface {
	OnHandler(msgID uint32, datas []byte, sess *CSClientSession)
	OnConnect(sess *CSClientSession)
	OnDisconnect(sess *CSClientSession)
	OnBeatHeartError(sess *CSClientSession)
}

type CSClientSession struct {
	Session
	handler              ICSClientMsgHandler
	timerRegister        etimer.ITimerRegister
	last_beat_heart_time int64
}

const (
	CS_CLIENT_BEAT_HEART_TIME_ID      uint32 = 1
	CS_CLIENT_SEND_BEAT_HEART_TIME_ID uint32 = 2
)

const (
	CS_CLIENT_BEAT_HEART_TIME_DELAY uint64 = 1000 * 1
	CS_CLIENT_SEND_BEAT_HEART_DELAY uint64 = 1000 * 20
)

const C2S_BEAT_HEART_MAX_TIME int64 = 1000 * 60

func NewCSClientSession(handler ICSClientMsgHandler) *CSClientSession {
	sess := &CSClientSession{
		handler:              handler,
		last_beat_heart_time: util.GetMillsecond(),
		timerRegister:        etimer.NewTimerRegister(),
	}

	sess.SetListenType()
	return sess
}

func (c *CSClientSession) OnEstablish() {
	c.factory.AddSession(c)
	elog.InfoAf("CSClientSession %v Establish", c.GetSessID())
	c.handler.OnConnect(c)

	c.timerRegister.AddRepeatTimer(CS_CLIENT_BEAT_HEART_TIME_ID, CS_CLIENT_BEAT_HEART_TIME_DELAY, "CSClientSession-BeatHeartCheck", func(v ...interface{}) {
		now := util.GetMillsecond()
		if (c.last_beat_heart_time + C2S_BEAT_HEART_MAX_TIME) < now {
			elog.ErrorAf("CSClientSession %v  BeatHeart Exception", c.GetSessID())
			c.handler.OnBeatHeartError(c)
			c.Terminate()
		}
	}, []interface{}{}, true)

	if c.IsConnectType() {
		c.timerRegister.AddRepeatTimer(CS_CLIENT_SEND_BEAT_HEART_TIME_ID, CS_CLIENT_SEND_BEAT_HEART_DELAY, "CSClientSession-SendBeatHeart", func(v ...interface{}) {
			elog.DebugAf("[CSClientSession] SessID=%v Send Beat Heart", c.GetSessID())
			c.AsyncSendMsg(uint32(pb.EClient2GameMsgId_c2s_client_session_ping_id), nil, nil)
		}, []interface{}{}, true)
	}
}

func (c *CSClientSession) OnTerminate() {
	elog.InfoAf("CSClientSession %v Terminate", c.GetSessID())
	c.factory.RemoveSession(c.GetSessID())
	c.timerRegister.KillAllTimer()
	c.handler.OnDisconnect(c)
}

func (c *CSClientSession) OnHandler(msgID uint32, attach_datas []byte, datas []byte) {
	if msgID == uint32(pb.EClient2GameMsgId_c2s_client_session_ping_id) {
		elog.DebugAf("[CSClientSession] SessionID=%v RECV PING SEND PONG", c.GetSessID())
		c.last_beat_heart_time = util.GetMillsecond()
		c.AsyncSendMsg(uint32(pb.EClient2GameMsgId_s2c_client_session_pong_id), nil, nil)
		return
	} else if msgID == uint32(pb.EClient2GameMsgId_s2c_client_session_pong_id) {
		elog.DebugAf("[CSClientSession] SessionID=%v RECV  PONG", c.GetSessID())
		c.last_beat_heart_time = util.GetMillsecond()
		return
	}

	c.handler.OnHandler(msgID, datas, c)
	c.last_beat_heart_time = util.GetMillsecond()
}

func (c *CSClientSession) SendMsg(msgID uint32, datas []byte) bool {
	return c.AsyncSendMsg(msgID, datas, nil)
}

func (c *CSClientSession) SendProtoMsg(msgID uint32, msg proto.Message) bool {
	return c.AsyncSendProtoMsg(msgID, msg, nil)
}

type CSClientSessionMgr struct {
	nextId  uint64
	sessMap map[uint64]inet.ISession
	handler ICSClientMsgHandler
	coder   inet.ICoder
}

func (c *CSClientSessionMgr) CreateSession() inet.ISession {
	sess := NewCSClientSession(c.handler)
	sess.SetSessID(c.nextId)
	sess.SetCoder(c.coder)
	sess.SetSessionFactory(c)
	c.nextId++
	return sess
}

func (c *CSClientSessionMgr) AddSession(session inet.ISession) {
	c.sessMap[session.GetSessID()] = session
}

func (c *CSClientSessionMgr) FindSession(id uint64) inet.ISession {
	if id == 0 {
		return nil
	}

	if sess, ok := c.sessMap[id]; ok {
		return sess
	}

	return nil
}

func (c *CSClientSessionMgr) RemoveSession(id uint64) {
	if _, ok := c.sessMap[id]; ok {
		delete(c.sessMap, id)
	}
}

func (c *CSClientSessionMgr) Count() int {
	return len(c.sessMap)
}

func (c *CSClientSessionMgr) SendProtoMsgBySessionID(sessionID uint64, msgID uint32, msg proto.Message) {
	serversess, ok := c.sessMap[sessionID]
	if ok {
		serversess.AsyncSendProtoMsg(msgID, msg, nil)
	}
}

func (c *CSClientSessionMgr) Connect(addr string, handler ICSClientMsgHandler, coder inet.ICoder) {
	if coder == nil {
		coder = NewCoder()
	}

	c.coder = coder
	c.handler = handler
	sess := c.CreateSession()
	sess.SetConnectType()
	enet.GNet.Connect(addr, sess)
}

func (c *CSClientSessionMgr) Listen(addr string, handler ICSClientMsgHandler, coder inet.ICoder) bool {
	if coder == nil {
		coder = NewCoder()
	}

	c.coder = coder
	c.handler = handler
	return enet.GNet.Listen(addr, c, GServerCfg.C2SListenMaxCount)
}

var GCSClientSessionMgr *CSClientSessionMgr

func init() {
	GCSClientSessionMgr = &CSClientSessionMgr{
		nextId:  1,
		sessMap: make(map[uint64]inet.ISession),
	}
}
