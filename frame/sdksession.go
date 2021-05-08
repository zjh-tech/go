package frame

import (
	"math"
	"projects/base/util"
	"projects/go-engine/enet"
	"projects/go-engine/etimer"
	"projects/pb"

	"github.com/golang/protobuf/proto"
)

type ISdkMsgHandler interface {
	Init() bool
	OnHandler(msg_id uint32, datas []byte, sess *SDKSession)
	OnConnect(sess *SDKSession)
	OnDisconnect(sess *SDKSession)
	OnBeatHeartError(sess *SDKSession)
}

type SDKSession struct {
	Session
	handler              ISdkMsgHandler
	timer_register       etimer.ITimerRegister
	last_beat_heart_time int64
	remote_outer         string
}

const (
	SDK_CHECK_BEAT_HEART_TIME_ID uint32 = 1
	SDK_SEND_BEAT_HEART_TIME_ID  uint32 = 2
)

const (
	SDK_HEART_TIME_DELAY      uint64 = 1000 * 1
	SDK_SEND_BEAT_HEART_DELAY uint64 = 1000 * 20
)

const (
	SDK_BEAT_HEART_MAX_TIME int64 = 1000 * 60 * 5
)

func NewSDKSession(handler ISdkMsgHandler) *SDKSession {
	sess := &SDKSession{
		handler:              handler,
		last_beat_heart_time: util.GetMillsecond(),
		timer_register:       etimer.NewTimerRegister(),
	}
	sess.SetListenType()
	sess.Session.SessionOnHandler = sess
	return sess
}

func (s *SDKSession) SetRemoteOuter(remote_outer string) {
	s.remote_outer = remote_outer
}

func (s *SDKSession) OnEstablish() {
	ELog.InfoAf("SDKSession %v Establish", s.GetSessID())
	s.factory.AddSession(s)
	s.handler.OnConnect(s)

	s.timer_register.AddRepeatTimer(SDK_CHECK_BEAT_HEART_TIME_ID, SDK_HEART_TIME_DELAY, "SDKSession-BeatHeartCheck", func(v ...interface{}) {
		now := util.GetMillsecond()
		if (s.last_beat_heart_time + SDK_BEAT_HEART_MAX_TIME) < now {
			ELog.ErrorAf("[SDKSession] SessID=%v BeatHeart Check Exception", s.GetSessID())
			s.handler.OnBeatHeartError(s)
			s.Terminate()
		} else {
			ELog.DebugAf("[SDKSession] SessID=%v BeatHeart Success", s.GetSessID())
		}
	}, []interface{}{}, true)

	if s.IsConnectType() {
		s.timer_register.AddRepeatTimer(SDK_SEND_BEAT_HEART_TIME_ID, SDK_SEND_BEAT_HEART_DELAY, "SDKSession-SendBeatHeart", func(v ...interface{}) {
			ELog.DebugAf("[SDKSession] SessID=%v Send Beat Heart", s.GetSessID())
			s.AsyncSendMsg(uint32(pb.S2SBaseMsgId_s2s_client_session_ping_id), nil)
		}, []interface{}{}, true)
	}
}

func (s *SDKSession) OnTerminate() {
	ELog.InfoAf("SDKSession %v Terminate", s.GetSessID())
	s.timer_register.KillAllTimer()
	factory := s.GetSessionFactory()
	ssclientfactory := factory.(*SDKSessionMgr)
	ssclientfactory.RemoveSession(s.GetSessID())
	s.handler.OnDisconnect(s)
}

func (s *SDKSession) OnHandler(msg_id uint32, datas []byte) {
	if msg_id == uint32(pb.S2SBaseMsgId_s2s_client_session_ping_id) {
		ELog.DebugAf("[SDKSession] SessionID=%v RECV SS_CLIENT_PING_MSG_ID", s.GetSessID())
		s.last_beat_heart_time = util.GetMillsecond()
		s.AsyncSendMsg(uint32(pb.S2SBaseMsgId_s2s_client_session_pong_id), nil)
		return
	} else if msg_id == uint32(pb.S2SBaseMsgId_s2s_client_session_pong_id) {
		ELog.DebugAf("[SDKSession] SessionID=%v RECV SS_CLIENT_PONG_MSG_ID", s.GetSessID())
		s.last_beat_heart_time = util.GetMillsecond()
		return
	}

	s.handler.OnHandler(msg_id, datas, s)
	s.last_beat_heart_time = util.GetMillsecond()
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
	next_id        uint64
	sess_map       map[uint64]enet.ISession
	handler        ISdkMsgHandler
	coder          enet.ICoder
	cache_map      map[uint64]*SCClientSessionCache
	timer_register etimer.ITimerRegister
}

func NewSDKSessionMgr() *SDKSessionMgr {
	return &SDKSessionMgr{
		next_id:        1,
		sess_map:       make(map[uint64]enet.ISession),
		timer_register: etimer.NewTimerRegister(),
		cache_map:      make(map[uint64]*SCClientSessionCache),
	}
}

func (s *SDKSessionMgr) Init() {
	s.timer_register.AddRepeatTimer(SDK_MGR_CACHE_TIMER_ID, SDK_MGR_CACHE_TIMER_DELAY, "SDKSessionMgr-Cache", func(v ...interface{}) {
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
	_, ok := s.sess_map[session_id]
	return ok
}

func (s *SDKSessionMgr) CreateSession() enet.ISession {
	sess := NewSDKSession(s.handler)
	sess.SetSessID(s.next_id)
	sess.SetCoder(s.coder)
	sess.SetSessionFactory(s)
	ELog.InfoAf("[SDKSessionMgr] CreateSession SessID=%v", sess.GetSessID())
	s.next_id++
	return sess
}

func (s *SDKSessionMgr) FindSession(id uint64) enet.ISession {
	if id == 0 {
		return nil
	}

	if sess, ok := s.sess_map[id]; ok {
		return sess
	}

	return nil
}

func (s *SDKSessionMgr) AddSession(session enet.ISession) {
	s.sess_map[session.GetSessID()] = session
	if info, ok := s.cache_map[session.GetSessID()]; ok {
		ELog.InfoAf("[SDKSessionMgr] AddSession Triggle ConnectCache Del SessionID=%v,ServerType=%v", session.GetSessID(), info.addr)
		delete(s.cache_map, session.GetSessID())
	}
}

func (s *SDKSessionMgr) RemoveSession(id uint64) {
	if _, ok := s.sess_map[id]; ok {
		delete(s.sess_map, id)
	}
}

func (s *SDKSessionMgr) Count() int {
	return len(s.sess_map)
}

func (s *SDKSessionMgr) AsyncSendProtoMsgBySessionID(sessionID uint64, msgID uint32, msg proto.Message) {
	serversess, ok := s.sess_map[sessionID]
	if ok {
		serversess.AsyncSendProtoMsg(msgID, msg)
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
		connect_tick: util.GetMillsecond() + SS_ONCE_CONNECT_MAX_TIME,
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
