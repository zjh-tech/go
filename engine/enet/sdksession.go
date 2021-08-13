package enet

import (
	"math"

	"github.com/golang/protobuf/proto"
)

type ISdkMsgHandler interface {
	Init() bool
	OnHandler(msgId uint32, datas []byte, sess *SDKSession)
	OnConnect(sess *SDKSession)
	OnDisconnect(sess *SDKSession)
	OnBeatHeartError(sess *SDKSession)
}

type SDKSession struct {
	Session
	handler                ISdkMsgHandler
	lastSendBeatHeartTime  int64
	lastCheckBeatHeartTime int64
	remote_outer           string
}

const (
	SdkBeatHeartMaxTime  int64 = 1000 * 60 * 2
	SdkSendBeatHeartTime int64 = 1000 * 20
	SdkMgrUpdateTime     int64 = 1000 * 1
)

func NewSDKSession(handler ISdkMsgHandler, isListenFlag bool) *SDKSession {
	sess := &SDKSession{
		handler:                handler,
		lastCheckBeatHeartTime: getMillsecond(),
	}
	sess.Session.ISessionOnHandler = sess
	if isListenFlag {
		sess.SetListenType()
	} else {
		sess.SetConnectType()
	}
	return sess
}

func (s *SDKSession) SetRemoteOuter(remote_outer string) {
	s.remote_outer = remote_outer
}

func (s *SDKSession) OnEstablish() {
	ELog.InfoAf("SDKSession %v Establish", s.GetSessID())
	s.factory.AddSession(s)
	s.handler.OnConnect(s)
}

func (s *SDKSession) Update() {
	now := getMillsecond()
	if (s.lastCheckBeatHeartTime + SdkBeatHeartMaxTime) < now {
		ELog.ErrorAf("[SDKSession] SessID=%v BeatHeart Check Exception", s.GetSessID())
		s.handler.OnBeatHeartError(s)
		s.Terminate()
		return
	}

	if s.IsConnectType() {
		if (s.lastSendBeatHeartTime + SdkSendBeatHeartTime) >= now {
			s.lastSendBeatHeartTime = now
			ELog.DebugAf("[SDKSession] SessID=%v Send Beat Heart", s.GetSessID())
			s.AsyncSendMsg(SDKSessionPingId, nil)
		}
	}
}

func (s *SDKSession) OnTerminate() {
	ELog.InfoAf("SDKSession %v Terminate", s.GetSessID())
	factory := s.GetSessionFactory()
	ssclientfactory := factory.(*SDKSessionMgr)
	ssclientfactory.RemoveSession(s.GetSessID())
	s.handler.OnDisconnect(s)
}

func (s *SDKSession) OnHandler(msgId uint32, datas []byte) {
	if msgId == SDKSessionPingId {
		ELog.DebugAf("[SDKSession] SessionID=%v RECV Ping ", s.GetSessID())
		s.lastCheckBeatHeartTime = getMillsecond()
		s.AsyncSendMsg(SDKSessionPongId, nil)
		return
	} else if msgId == SDKSessionPongId {
		ELog.DebugAf("[SDKSession] SessionID=%v RECV Pong", s.GetSessID())
		s.lastCheckBeatHeartTime = getMillsecond()
		return
	}

	s.handler.OnHandler(msgId, datas, s)
	s.lastCheckBeatHeartTime = getMillsecond()
}

type SCClientSessionCache struct {
	sessionId   uint64
	addr        string
	connectTick int64
}

type SDKSessionMgr struct {
	nextId         uint64
	sessMap        map[uint64]ISession
	handler        ISdkMsgHandler
	coder          ICoder
	cacheMap       map[uint64]*SCClientSessionCache
	lastUpdateTime int64
}

func NewSDKSessionMgr() *SDKSessionMgr {
	return &SDKSessionMgr{
		nextId:         1,
		sessMap:        make(map[uint64]ISession),
		cacheMap:       make(map[uint64]*SCClientSessionCache),
		lastUpdateTime: getMillsecond(),
	}
}

func (s *SDKSessionMgr) Init() {

}

func (s *SDKSessionMgr) Update() {
	now := getMillsecond()
	if (s.lastUpdateTime + SdkMgrUpdateTime) >= now {
		s.lastUpdateTime = now

		for sessionID, cache := range s.cacheMap {
			if cache.connectTick < now {
				ELog.InfoAf("[SDKSessionMgr] Timeout Triggle  ConnectCache Del SesssionID=%v,Addr=%v", cache.sessionId, cache.addr)
				delete(s.cacheMap, sessionID)
			}
		}
	}

	for _, session := range s.sessMap {
		sess := session.(*SDKSession)
		if sess != nil {
			sess.Update()
		}
	}
}

func (s *SDKSessionMgr) IsInConnectCache(sessionId uint64) bool {
	_, ok := s.cacheMap[sessionId]
	return ok
}

func (s *SDKSessionMgr) IsExistSessionOfSessID(sessionId uint64) bool {
	_, ok := s.sessMap[sessionId]
	return ok
}

func (s *SDKSessionMgr) CreateSession(isListenFlag bool) ISession {
	sess := NewSDKSession(s.handler, isListenFlag)
	sess.SetSessID(s.nextId)
	sess.SetCoder(s.coder)
	sess.SetSessionFactory(s)
	ELog.InfoAf("[SDKSessionMgr] CreateSession SessID=%v", sess.GetSessID())
	s.nextId++
	return sess
}

func (s *SDKSessionMgr) FindSession(id uint64) ISession {
	if id == 0 {
		return nil
	}

	if sess, ok := s.sessMap[id]; ok {
		return sess
	}

	return nil
}

func (s *SDKSessionMgr) GetSessionCount() int {
	return len(s.sessMap)
}

func (s *SDKSessionMgr) AddSession(session ISession) {
	s.sessMap[session.GetSessID()] = session
	if info, ok := s.cacheMap[session.GetSessID()]; ok {
		ELog.InfoAf("[SDKSessionMgr] AddSession Triggle ConnectCache Del SessionID=%v,ServerType=%v", session.GetSessID(), info.addr)
		delete(s.cacheMap, session.GetSessID())
	}
}

func (s *SDKSessionMgr) RemoveSession(id uint64) {
	delete(s.sessMap, id)
}

func (s *SDKSessionMgr) Count() int {
	return len(s.sessMap)
}

func (s *SDKSessionMgr) SendProtoMsgBySessionID(sessionID uint64, msgId uint32, msg proto.Message) {
	serversess, ok := s.sessMap[sessionID]
	if ok {
		serversess.AsyncSendProtoMsg(msgId, msg)
	}
}

func (s *SDKSessionMgr) SdkConnect(addr string, handler ISdkMsgHandler, coder ICoder) uint64 {
	if coder == nil {
		coder = NewCoder()
	}

	s.handler = handler
	s.coder = coder
	sess := s.CreateSession(false)
	sdkSess := sess.(*SDKSession)
	sdkSess.SetRemoteOuter(addr)
	sdkSess.SetConnectType()

	cache := &SCClientSessionCache{
		sessionId:   sess.GetSessID(),
		addr:        addr,
		connectTick: getMillsecond() + SSOnceConnectMaxTime,
	}
	s.cacheMap[sess.GetSessID()] = cache
	ELog.InfoAf("[SDKSessionMgr]ConnectCache Add SessionID=%v,Addr=%v", sess.GetSessID(), addr)
	GNet.Connect(addr, sess)

	return sess.GetSessID()
}

func (s *SDKSessionMgr) SdkListen(addr string, handler ISdkMsgHandler, coder ICoder) bool {
	if coder == nil {
		coder = NewCoder()
	}
	handler.Init()
	s.handler = handler
	s.coder = coder
	return GNet.Listen(addr, s, math.MaxInt32)
}

var GSDKSessionMgr *SDKSessionMgr
