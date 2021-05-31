package enet

import (
	"github.com/golang/protobuf/proto"
)

type ICSMsgHandler interface {
	OnHandler(msgId uint32, datas []byte, sess *CSSession)
	OnConnect(sess *CSSession)
	OnDisconnect(sess *CSSession)
	OnBeatHeartError(sess *CSSession)
}

type CSSession struct {
	Session
	handler                ICSMsgHandler
	lastSendBeatHeartTime  int64
	lastCheckBeatHeartTime int64
}

const (
	C2SBeatHeartMaxTime  int64 = 1000 * 60 * 2
	C2SSendBeatHeartTime int64 = 1000 * 20
)

func NewCSSession(handler ICSMsgHandler) *CSSession {
	sess := &CSSession{
		handler:                handler,
		lastCheckBeatHeartTime: getMillsecond(),
		lastSendBeatHeartTime:  getMillsecond(),
	}

	sess.SetListenType()
	sess.Session.ISessionOnHandler = sess
	return sess
}

func (c *CSSession) OnEstablish() {
	c.factory.AddSession(c)
	ELog.InfoAf("CSSession %v Establish", c.GetSessID())
	c.handler.OnConnect(c)
}

func (c *CSSession) Update() {
	now := getMillsecond()
	if (c.lastCheckBeatHeartTime + C2SBeatHeartMaxTime) < now {
		ELog.ErrorAf("CSSession %v  BeatHeart Exception", c.GetSessID())
		c.handler.OnBeatHeartError(c)
		c.Terminate()
		return
	}

	if c.IsConnectType() {
		if (c.lastSendBeatHeartTime + C2SSendBeatHeartTime) >= now {
			c.lastSendBeatHeartTime = now
			ELog.DebugAf("[CSSession] SessID=%v Send Beat Heart", c.GetSessID())
			c.AsyncSendMsg(C2SSessionPingId, nil)
		}
	}
}

func (c *CSSession) OnTerminate() {
	ELog.InfoAf("CSSession %v Terminate", c.GetSessID())
	c.factory.RemoveSession(c.GetSessID())
	c.handler.OnDisconnect(c)
}

func (c *CSSession) OnHandler(msgId uint32, datas []byte) {
	if msgId == C2SSessionPingId {
		ELog.DebugAf("[CSSession] SessionID=%v RECV PING SEND PONG", c.GetSessID())
		c.lastCheckBeatHeartTime = getMillsecond()
		c.AsyncSendMsg(C2SSessionPongId, nil)
		return
	} else if msgId == C2SSessionPongId {
		ELog.DebugAf("[CSSession] SessionID=%v RECV  PONG", c.GetSessID())
		c.lastCheckBeatHeartTime = getMillsecond()
		return
	}

	c.handler.OnHandler(msgId, datas, c)
	c.lastCheckBeatHeartTime = getMillsecond()
}

type CSSessionMgr struct {
	nextId  uint64
	sessMap map[uint64]ISession
	handler ICSMsgHandler
	coder   ICoder
}

func NewCSSessionMgr() *CSSessionMgr {
	return &CSSessionMgr{
		nextId:  1,
		sessMap: make(map[uint64]ISession),
	}
}

func (c *CSSessionMgr) Update() {
	for _, session := range c.sessMap {
		sess := session.(*CSSession)
		if sess != nil {
			sess.Update()
		}
	}
}

func (c *CSSessionMgr) CreateSession() ISession {
	sess := NewCSSession(c.handler)
	sess.SetSessID(c.nextId)
	sess.SetCoder(c.coder)
	sess.SetSessionFactory(c)
	c.nextId++
	return sess
}

func (c *CSSessionMgr) AddSession(session ISession) {
	c.sessMap[session.GetSessID()] = session
}

func (c *CSSessionMgr) FindSession(id uint64) ISession {
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

func (c *CSSessionMgr) Connect(addr string, handler ICSMsgHandler, coder ICoder) {
	if coder == nil {
		coder = NewCoder()
	}

	c.coder = coder
	c.handler = handler
	sess := c.CreateSession()
	sess.SetConnectType()
	GNet.Connect(addr, sess)
}

func (c *CSSessionMgr) Listen(addr string, handler ICSMsgHandler, coder ICoder, listenMaxCount int) bool {
	if coder == nil {
		coder = NewCoder()
	}

	c.coder = coder
	c.handler = handler
	return GNet.Listen(addr, c, listenMaxCount)
}

var GCSSessionMgr *CSSessionMgr
