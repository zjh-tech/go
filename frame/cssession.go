package frame

import (
	"github.com/golang/protobuf/proto"
	"github.com/zjh-tech/go-frame/base/util"
	"github.com/zjh-tech/go-frame/engine/enet"
	"github.com/zjh-tech/go-frame/engine/etimer"
)

type ICSMsgHandler interface {
	OnHandler(msgId uint32, datas []byte, sess *CSSession)
	OnConnect(sess *CSSession)
	OnDisconnect(sess *CSSession)
	OnBeatHeartError(sess *CSSession)
}

type CSSession struct {
	Session
	handler           ICSMsgHandler
	timerRegister     etimer.ITimerRegister
	lastBeatHeartTime int64
}

const (
	CsBeatHeartTimeId     uint32 = 1
	CsSendBeatHeartTimeId uint32 = 2
)

const (
	CsBeatHeartTimeDelay uint64 = 1000 * 1
	CsSendBeatHeartDelay uint64 = 1000 * 20
)

const C2sBeatHeartMaxTime int64 = 1000 * 60

func NewCSSession(handler ICSMsgHandler) *CSSession {
	sess := &CSSession{
		handler:           handler,
		lastBeatHeartTime: util.GetMillsecond(),
		timerRegister:     etimer.NewTimerRegister(),
	}

	sess.SetListenType()
	sess.Session.ISessionOnHandler = sess
	return sess
}

func (c *CSSession) OnEstablish() {
	c.factory.AddSession(c)
	ELog.InfoAf("CSSession %v Establish", c.GetSessID())
	c.handler.OnConnect(c)

	c.timerRegister.AddRepeatTimer(CsBeatHeartTimeId, CsBeatHeartTimeDelay, "CSSession-BeatHeartCheck", func(v ...interface{}) {
		now := util.GetMillsecond()
		if (c.lastBeatHeartTime + C2sBeatHeartMaxTime) < now {
			ELog.ErrorAf("CSSession %v  BeatHeart Exception", c.GetSessID())
			c.handler.OnBeatHeartError(c)
			c.Terminate()
		}
	}, []interface{}{}, true)

	if c.IsConnectType() {
		c.timerRegister.AddRepeatTimer(CsSendBeatHeartTimeId, CsSendBeatHeartDelay, "CSSession-SendBeatHeart", func(v ...interface{}) {
			ELog.DebugAf("[CSSession] SessID=%v Send Beat Heart", c.GetSessID())
			c.AsyncSendMsg(C2SSessionPingId, nil)
		}, []interface{}{}, true)
	}
}

func (c *CSSession) OnTerminate() {
	ELog.InfoAf("CSSession %v Terminate", c.GetSessID())
	c.factory.RemoveSession(c.GetSessID())
	c.timerRegister.KillAllTimer()
	c.handler.OnDisconnect(c)
}

func (c *CSSession) OnHandler(msgId uint32, datas []byte) {
	if msgId == C2SSessionPingId {
		ELog.DebugAf("[CSSession] SessionID=%v RECV PING SEND PONG", c.GetSessID())
		c.lastBeatHeartTime = util.GetMillsecond()
		c.AsyncSendMsg(C2SSessionPongId, nil)
		return
	} else if msgId == C2SSessionPongId {
		ELog.DebugAf("[CSSession] SessionID=%v RECV  PONG", c.GetSessID())
		c.lastBeatHeartTime = util.GetMillsecond()
		return
	}

	c.handler.OnHandler(msgId, datas, c)
	c.lastBeatHeartTime = util.GetMillsecond()
}

type CSSessionMgr struct {
	nextId  uint64
	sessMap map[uint64]enet.ISession
	handler ICSMsgHandler
	coder   enet.ICoder
}

func NewCSSessionMgr() *CSSessionMgr {
	return &CSSessionMgr{
		nextId:  1,
		sessMap: make(map[uint64]enet.ISession),
	}
}

func (c *CSSessionMgr) CreateSession() enet.ISession {
	sess := NewCSSession(c.handler)
	sess.SetSessID(c.nextId)
	sess.SetCoder(c.coder)
	sess.SetSessionFactory(c)
	c.nextId++
	return sess
}

func (c *CSSessionMgr) AddSession(session enet.ISession) {
	c.sessMap[session.GetSessID()] = session
}

func (c *CSSessionMgr) FindSession(id uint64) enet.ISession {
	if id == 0 {
		return nil
	}

	if sess, ok := c.sessMap[id]; ok {
		return sess
	}

	return nil
}

func (c *CSSessionMgr) GetSessionCount() int {
	return len(c.sessMap)
}

func (c *CSSessionMgr) RemoveSession(id uint64) {
	if _, ok := c.sessMap[id]; ok {
		delete(c.sessMap, id)
	}
}

func (c *CSSessionMgr) Count() int {
	return len(c.sessMap)
}

func (c *CSSessionMgr) SendProtoMsgBySessionID(sessionID uint64, msgId uint32, msg proto.Message) {
	serversess, ok := c.sessMap[sessionID]
	if ok {
		serversess.AsyncSendProtoMsg(msgId, msg)
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
