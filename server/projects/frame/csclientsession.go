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
	timer_register       etimer.ITimerRegister
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
		timer_register:       etimer.NewTimerRegister(),
	}

	sess.SetListenType()
	sess.Session.SessionOnHandler = sess
	return sess
}

func (c *CSClientSession) OnEstablish() {
	c.factory.AddSession(c)
	elog.InfoAf("CSClientSession %v Establish", c.GetSessID())
	c.handler.OnConnect(c)

	c.timer_register.AddRepeatTimer(CS_CLIENT_BEAT_HEART_TIME_ID, CS_CLIENT_BEAT_HEART_TIME_DELAY, "CSClientSession-BeatHeartCheck", func(v ...interface{}) {
		now := util.GetMillsecond()
		if (c.last_beat_heart_time + C2S_BEAT_HEART_MAX_TIME) < now {
			elog.ErrorAf("CSClientSession %v  BeatHeart Exception", c.GetSessID())
			c.handler.OnBeatHeartError(c)
			c.Terminate()
		}
	}, []interface{}{}, true)

	if c.IsConnectType() {
		c.timer_register.AddRepeatTimer(CS_CLIENT_SEND_BEAT_HEART_TIME_ID, CS_CLIENT_SEND_BEAT_HEART_DELAY, "CSClientSession-SendBeatHeart", func(v ...interface{}) {
			elog.DebugAf("[CSClientSession] SessID=%v Send Beat Heart", c.GetSessID())
			c.AsyncSendMsg(uint32(pb.EClient2GameMsgId_c2s_client_session_ping_id), nil, nil)
		}, []interface{}{}, true)
	}
}

func (c *CSClientSession) OnTerminate() {
	elog.InfoAf("CSClientSession %v Terminate", c.GetSessID())
	c.factory.RemoveSession(c.GetSessID())
	c.timer_register.KillAllTimer()
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
	next_id  uint64
	sess_map map[uint64]inet.ISession
	handler  ICSClientMsgHandler
	coder    inet.ICoder
}

func NewCSClientSessionMgr() *CSClientSessionMgr {
	return &CSClientSessionMgr{
		next_id:  1,
		sess_map: make(map[uint64]inet.ISession),
	}
}

func (c *CSClientSessionMgr) CreateSession() inet.ISession {
	sess := NewCSClientSession(c.handler)
	sess.SetSessID(c.next_id)
	sess.SetCoder(c.coder)
	sess.SetSessionFactory(c)
	c.next_id++
	return sess
}

func (c *CSClientSessionMgr) AddSession(session inet.ISession) {
	c.sess_map[session.GetSessID()] = session
}

func (c *CSClientSessionMgr) FindSession(id uint64) inet.ISession {
	if id == 0 {
		return nil
	}

	if sess, ok := c.sess_map[id]; ok {
		return sess
	}

	return nil
}

func (c *CSClientSessionMgr) RemoveSession(id uint64) {
	if _, ok := c.sess_map[id]; ok {
		delete(c.sess_map, id)
	}
}

func (c *CSClientSessionMgr) Count() int {
	return len(c.sess_map)
}

func (c *CSClientSessionMgr) SendProtoMsgBySessionID(sessionID uint64, msgID uint32, msg proto.Message) {
	serversess, ok := c.sess_map[sessionID]
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

func (c *CSClientSessionMgr) Listen(addr string, handler ICSClientMsgHandler, coder inet.ICoder, listenMaxCount int) bool {
	if coder == nil {
		coder = NewCoder()
	}

	c.coder = coder
	c.handler = handler
	return enet.GNet.Listen(addr, c, listenMaxCount)
}

var GCSClientSessionMgr *CSClientSessionMgr

func init() {
	GCSClientSessionMgr = NewCSClientSessionMgr()
}
