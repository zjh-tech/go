package frame

import (
	"projects/base/util"
	"projects/engine/enet"
	"projects/engine/etimer"
	"projects/pb"

	"github.com/golang/protobuf/proto"
)

type ICSMsgHandler interface {
	OnHandler(msgID uint32, datas []byte, sess *CSSession)
	OnConnect(sess *CSSession)
	OnDisconnect(sess *CSSession)
	OnBeatHeartError(sess *CSSession)
}

type CSSession struct {
	Session
	handler              ICSMsgHandler
	timer_register       etimer.ITimerRegister
	last_beat_heart_time int64
}

const (
	CS_BEAT_HEART_TIME_ID      uint32 = 1
	CS_SEND_BEAT_HEART_TIME_ID uint32 = 2
)

const (
	CS_BEAT_HEART_TIME_DELAY uint64 = 1000 * 1
	CS_SEND_BEAT_HEART_DELAY uint64 = 1000 * 20
)

const C2S_BEAT_HEART_MAX_TIME int64 = 1000 * 60

func NewCSSession(handler ICSMsgHandler) *CSSession {
	sess := &CSSession{
		handler:              handler,
		last_beat_heart_time: util.GetMillsecond(),
		timer_register:       etimer.NewTimerRegister(),
	}

	sess.SetListenType()
	sess.Session.SessionOnHandler = sess
	return sess
}

func (c *CSSession) OnEstablish() {
	c.factory.AddSession(c)
	ELog.InfoAf("CSSession %v Establish", c.GetSessID())
	c.handler.OnConnect(c)

	c.timer_register.AddRepeatTimer(CS_BEAT_HEART_TIME_ID, CS_BEAT_HEART_TIME_DELAY, "CSSession-BeatHeartCheck", func(v ...interface{}) {
		now := util.GetMillsecond()
		if (c.last_beat_heart_time + C2S_BEAT_HEART_MAX_TIME) < now {
			ELog.ErrorAf("CSSession %v  BeatHeart Exception", c.GetSessID())
			c.handler.OnBeatHeartError(c)
			c.Terminate()
		}
	}, []interface{}{}, true)

	if c.IsConnectType() {
		c.timer_register.AddRepeatTimer(CS_SEND_BEAT_HEART_TIME_ID, CS_SEND_BEAT_HEART_DELAY, "CSSession-SendBeatHeart", func(v ...interface{}) {
			ELog.DebugAf("[CSSession] SessID=%v Send Beat Heart", c.GetSessID())
			c.AsyncSendMsg(uint32(pb.EClient2GameMsgId_c2s_client_session_ping_id), nil)
		}, []interface{}{}, true)
	}
}

func (c *CSSession) OnTerminate() {
	ELog.InfoAf("CSSession %v Terminate", c.GetSessID())
	c.factory.RemoveSession(c.GetSessID())
	c.timer_register.KillAllTimer()
	c.handler.OnDisconnect(c)
}

func (c *CSSession) OnHandler(msgID uint32, datas []byte) {
	if msgID == uint32(pb.EClient2GameMsgId_c2s_client_session_ping_id) {
		ELog.DebugAf("[CSSession] SessionID=%v RECV PING SEND PONG", c.GetSessID())
		c.last_beat_heart_time = util.GetMillsecond()
		c.AsyncSendMsg(uint32(pb.EClient2GameMsgId_s2c_client_session_pong_id), nil)
		return
	} else if msgID == uint32(pb.EClient2GameMsgId_s2c_client_session_pong_id) {
		ELog.DebugAf("[CSSession] SessionID=%v RECV  PONG", c.GetSessID())
		c.last_beat_heart_time = util.GetMillsecond()
		return
	}

	c.handler.OnHandler(msgID, datas, c)
	c.last_beat_heart_time = util.GetMillsecond()
}

func (c *CSSession) SendMsg(msgID uint32, datas []byte) bool {
	return c.AsyncSendMsg(msgID, datas)
}

func (c *CSSession) SendProtoMsg(msgID uint32, msg proto.Message) bool {
	return c.AsyncSendProtoMsg(msgID, msg)
}

type CSSessionMgr struct {
	next_id  uint64
	sess_map map[uint64]enet.ISession
	handler  ICSMsgHandler
	coder    enet.ICoder
}

func NewCSSessionMgr() *CSSessionMgr {
	return &CSSessionMgr{
		next_id:  1,
		sess_map: make(map[uint64]enet.ISession),
	}
}

func (c *CSSessionMgr) CreateSession() enet.ISession {
	sess := NewCSSession(c.handler)
	sess.SetSessID(c.next_id)
	sess.SetCoder(c.coder)
	sess.SetSessionFactory(c)
	c.next_id++
	return sess
}

func (c *CSSessionMgr) AddSession(session enet.ISession) {
	c.sess_map[session.GetSessID()] = session
}

func (c *CSSessionMgr) FindSession(id uint64) enet.ISession {
	if id == 0 {
		return nil
	}

	if sess, ok := c.sess_map[id]; ok {
		return sess
	}

	return nil
}

func (c *CSSessionMgr) RemoveSession(id uint64) {
	if _, ok := c.sess_map[id]; ok {
		delete(c.sess_map, id)
	}
}

func (c *CSSessionMgr) Count() int {
	return len(c.sess_map)
}

func (c *CSSessionMgr) SendProtoMsgBySessionID(sessionID uint64, msgID uint32, msg proto.Message) {
	serversess, ok := c.sess_map[sessionID]
	if ok {
		serversess.AsyncSendProtoMsg(msgID, msg)
	}
}

func (c *CSSessionMgr) Connect(addr string, handler ICSMsgHandler, coder enet.ICoder) {
	if coder == nil {
		coder = NewCoder()
	}

	c.coder = coder
	c.handler = handler
	sess := c.CreateSession()
	sess.SetConnectType()
	enet.GNet.Connect(addr, sess)
}

func (c *CSSessionMgr) Listen(addr string, handler ICSMsgHandler, coder enet.ICoder, listenMaxCount int) bool {
	if coder == nil {
		coder = NewCoder()
	}

	c.coder = coder
	c.handler = handler
	return enet.GNet.Listen(addr, c, listenMaxCount)
}

var GCSSessionMgr *CSSessionMgr

func init() {
	GCSSessionMgr = NewCSSessionMgr()
}
