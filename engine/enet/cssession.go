package enet

import (
	"github.com/golang/protobuf/proto"
)

const (
	C2SBeatHeartMaxTime  int64 = 1000 * 60 * 2
	C2SSendBeatHeartTime int64 = 1000 * 20
	CsMgrUpdateTime      int64 = 1000 * 1
	CsOnceConnectMaxTime int64 = 1000 * 10
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
	overloadModule         *OverLoadModule
}

func NewCSSession(handler ICSMsgHandler, isListenFlag bool) *CSSession {
	sess := &CSSession{
		handler:                handler,
		lastCheckBeatHeartTime: getMillsecond(),
		lastSendBeatHeartTime:  getMillsecond(),
		overloadModule:         nil,
	}
	sess.Session.ISessionOnHandler = sess
	if isListenFlag {
		sess.SetListenType()
	} else {
		sess.SetConnectType()
	}
	return sess
}

func (c *CSSession) SetOverload(intervalTime int64, limit int64) {
	c.overloadModule = NewOverLoadModule(intervalTime, limit)
	ELog.InfoAf("CSSession SessId=%v SetOverload IntervalTime=%v,Limit=%v", c.GetSessID(), intervalTime, limit)
}

func (c *CSSession) OnEstablish() {
	ELog.InfoAf("CSSession %v Establish", c.GetSessID())
	c.factory.AddSession(c)
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
	//底层提供了心跳
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

	//业务层也可以提供另外的心跳
	c.handler.OnHandler(msgId, datas, c)
	c.lastCheckBeatHeartTime = getMillsecond()

	if c.overloadModule != nil {
		c.overloadModule.AddCount()
		if c.overloadModule.IsOverLoad() {
			ELog.ErrorAf("[CSSession] SessionID=%v OverLoad", c.GetSessID())
			c.Terminate()
			return
		}
	}
}

type CSSessionCache struct {
	sessionId   uint64
	addr        string
	connectTick int64
}

type CSSessionMgr struct {
	nextId         uint64
	sessMap        map[uint64]ISession
	handler        ICSMsgHandler
	coder          ICoder
	cacheMap       map[uint64]*CSSessionCache
	lastUpdateTime int64
}

func NewCSSessionMgr() *CSSessionMgr {
	return &CSSessionMgr{
		nextId:         1,
		sessMap:        make(map[uint64]ISession),
		cacheMap:       make(map[uint64]*CSSessionCache),
		lastUpdateTime: getMillsecond(),
	}
}

func (c *CSSessionMgr) IsInConnectCache(sessionId uint64) bool {
	_, ok := c.cacheMap[sessionId]
	return ok
}

func (c *CSSessionMgr) IsExistSessionOfSessID(sessionId uint64) bool {
	_, ok := c.sessMap[sessionId]
	return ok
}

func (c *CSSessionMgr) Update() {
	now := getMillsecond()
	if (c.lastUpdateTime + CsMgrUpdateTime) >= now {
		c.lastUpdateTime = now

		for sessionID, cache := range c.cacheMap {
			if cache.connectTick < now {
				ELog.InfoAf("[CSSessionMgr] Timeout Triggle  ConnectCache Del SesssionID=%v,Addr=%v", cache.sessionId, cache.addr)
				delete(c.cacheMap, sessionID)
			}
		}
	}

	for _, session := range c.sessMap {
		sess := session.(*CSSession)
		if sess != nil {
			sess.Update()
		}
	}
}

func (c *CSSessionMgr) CreateSession(isListenFlag bool) ISession {
	sess := NewCSSession(c.handler, isListenFlag)
	sess.SetSessID(c.nextId)
	sess.SetCoder(c.coder)
	sess.SetSessionFactory(c)
	sess.SetSessionConcurrentFlag(true)
	ELog.InfoAf("[CSSessionMgr] CreateSession SessID=%v", sess.GetSessID())
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
	delete(c.sessMap, id)
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
	sess := c.CreateSession(false)
	cache := &CSSessionCache{
		sessionId:   sess.GetSessID(),
		addr:        addr,
		connectTick: getMillsecond() + CsOnceConnectMaxTime,
	}
	c.cacheMap[sess.GetSessID()] = cache
	ELog.InfoAf("[CSSessionMgr]ConnectCache Add SessionID=%v,Addr=%v", sess.GetSessID(), addr)
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
