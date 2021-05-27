package frame

import (
	"math"

	"github.com/golang/protobuf/proto"
	"github.com/zjh-tech/go-frame/base/util"
	"github.com/zjh-tech/go-frame/engine/enet"
	"github.com/zjh-tech/go-frame/engine/etimer"
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
	handler           ISdkMsgHandler
	timerRegister     etimer.ITimerRegister
	lastBeatHeartTime int64
	remote_outer      string
}

const (
	SdkCheckBeatHeartTimerId uint32 = 1
	SdkSendBeatHeartTimeId   uint32 = 2
)

const (
	SdkHeartTimeDelay     uint64 = 1000 * 1
	SdkSendBeatHeartDelay uint64 = 1000 * 20
)

const (
	SdkBeatHeartMaxTime int64 = 1000 * 60 * 5
)

func NewSDKSession(handler ISdkMsgHandler) *SDKSession {
	sess := &SDKSession{
		handler:           handler,
		lastBeatHeartTime: util.GetMillsecond(),
		timerRegister:     etimer.NewTimerRegister(),
	}
	sess.SetListenType()
	sess.Session.ISessionOnHandler = sess
	return sess
}

func (s *SDKSession) SetRemoteOuter(remote_outer string) {
	s.remote_outer = remote_outer
}

func (s *SDKSession) OnEstablish() {
	ELog.InfoAf("SDKSession %v Establish", s.GetSessID())
	s.factory.AddSession(s)
	s.handler.OnConnect(s)

	s.timerRegister.AddRepeatTimer(SdkCheckBeatHeartTimerId, SdkHeartTimeDelay, "SDKSession-BeatHeartCheck", func(v ...interface{}) {
		now := util.GetMillsecond()
		if (s.lastBeatHeartTime + SdkBeatHeartMaxTime) < now {
			ELog.ErrorAf("[SDKSession] SessID=%v BeatHeart Check Exception", s.GetSessID())
			s.handler.OnBeatHeartError(s)
			s.Terminate()
		} else {
			ELog.DebugAf("[SDKSession] SessID=%v BeatHeart Success", s.GetSessID())
		}
	}, []interface{}{}, true)

	if s.IsConnectType() {
		s.timerRegister.AddRepeatTimer(SdkSendBeatHeartTimeId, SdkSendBeatHeartDelay, "SDKSession-SendBeatHeart", func(v ...interface{}) {
			ELog.DebugAf("[SDKSession] SessID=%v Send Beat Heart", s.GetSessID())
			s.AsyncSendMsg(SDKSessionPingId, nil)
		}, []interface{}{}, true)
	}
}

func (s *SDKSession) OnTerminate() {
	ELog.InfoAf("SDKSession %v Terminate", s.GetSessID())
	s.timerRegister.KillAllTimer()
	factory := s.GetSessionFactory()
	ssclientfactory := factory.(*SDKSessionMgr)
	ssclientfactory.RemoveSession(s.GetSessID())
	s.handler.OnDisconnect(s)
}

func (s *SDKSession) OnHandler(msgId uint32, datas []byte) {
	if msgId == SDKSessionPingId {
		ELog.DebugAf("[SDKSession] SessionID=%v RECV Ping ", s.GetSessID())
		s.lastBeatHeartTime = util.GetMillsecond()
		s.AsyncSendMsg(SDKSessionPongId, nil)
		return
	} else if msgId == SDKSessionPongId {
		ELog.DebugAf("[SDKSession] SessionID=%v RECV Pong", s.GetSessID())
		s.lastBeatHeartTime = util.GetMillsecond()
		return
	}

	s.handler.OnHandler(msgId, datas, s)
	s.lastBeatHeartTime = util.GetMillsecond()
}

type SCClientSessionCache struct {
	session_id   uint64
	addr         string
	connect_tick int64
}

const (
	SDK_MGR_CACHE_TIMER_ID uint32 = 1
)

const (
	SDK_MGR_CACHE_TIMER_DELAY uint64 = 1000 * 1
)

type SDKSessionMgr struct {
	nextId        uint64
	sessMap       map[uint64]enet.ISession
	handler       ISdkMsgHandler
	coder         enet.ICoder
	cache_map     map[uint64]*SCClientSessionCache
	timerRegister etimer.ITimerRegister
}

func NewSDKSessionMgr() *SDKSessionMgr {
	return &SDKSessionMgr{
		nextId:        1,
		sessMap:       make(map[uint64]enet.ISession),
		timerRegister: etimer.NewTimerRegister(),
		cache_map:     make(map[uint64]*SCClientSessionCache),
	}
}

func (s *SDKSessionMgr) Init() {
	s.timerRegister.AddRepeatTimer(SDK_MGR_CACHE_TIMER_ID, SDK_MGR_CACHE_TIMER_DELAY, "SDKSessionMgr-Cache", func(v ...interface{}) {
		now := util.GetMillsecond()
		for sessionID, cache := range s.cache_map {
			if cache.connect_tick < now {
				ELog.InfoAf("[SDKSessionMgr] Timeout Triggle  ConnectCache Del SesssionID=%v,Addr=%v", cache.session_id, cache.addr)
				delete(s.cache_map, sessionID)
			}
		}
	}, []interface{}{}, true)
}

func (s *SDKSessionMgr) IsInConnectCache(session_id uint64) bool {
	_, ok := s.cache_map[session_id]
	return ok
}

func (s *SDKSessionMgr) IsExistSessionOfSessID(session_id uint64) bool {
	_, ok := s.sessMap[session_id]
	return ok
}

func (s *SDKSessionMgr) CreateSession() enet.ISession {
	sess := NewSDKSession(s.handler)
	sess.SetSessID(s.nextId)
	sess.SetCoder(s.coder)
	sess.SetSessionFactory(s)
	ELog.InfoAf("[SDKSessionMgr] CreateSession SessID=%v", sess.GetSessID())
	s.nextId++
	return sess
}

func (s *SDKSessionMgr) FindSession(id uint64) enet.ISession {
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

func (s *SDKSessionMgr) AddSession(session enet.ISession) {
	s.sessMap[session.GetSessID()] = session
	if info, ok := s.cache_map[session.GetSessID()]; ok {
		ELog.InfoAf("[SDKSessionMgr] AddSession Triggle ConnectCache Del SessionID=%v,ServerType=%v", session.GetSessID(), info.addr)
		delete(s.cache_map, session.GetSessID())
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

func (s *SDKSessionMgr) SSClientConnect(addr string, handler ISdkMsgHandler, coder enet.ICoder) uint64 {
	if coder == nil {
		coder = NewCoder()
	}

	//handler.Init()
	s.handler = handler
	s.coder = coder
	sess := s.CreateSession()
	ssClientSess := sess.(*SDKSession)
	ssClientSess.SetRemoteOuter(addr)
	ssClientSess.SetConnectType()

	cache := &SCClientSessionCache{
		session_id:   sess.GetSessID(),
		addr:         addr,
		connect_tick: util.GetMillsecond() + SSOnceConnectMaxTime,
	}
	s.cache_map[sess.GetSessID()] = cache
	ELog.InfoAf("[SDKSessionMgr]ConnectCache Add SessionID=%v,Addr=%v", sess.GetSessID(), addr)
	enet.GNet.Connect(addr, sess)

	return sess.GetSessID()
}

func (s *SDKSessionMgr) SSClientListen(addr string, handler ISdkMsgHandler, coder enet.ICoder) bool {
	if coder == nil {
		coder = NewCoder()
	}
	handler.Init()
	s.handler = handler
	s.coder = coder
	return enet.GNet.Listen(addr, s, math.MaxInt32)
}

var GSDKSessionMgr *SDKSessionMgr

func init() {
	GSDKSessionMgr = NewSDKSessionMgr()
}
