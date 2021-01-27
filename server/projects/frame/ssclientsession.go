package frame

import (
	"math"
	"projects/go-engine/elog"
	"projects/go-engine/enet"
	"projects/go-engine/etimer"
	"projects/go-engine/inet"
	"projects/pb"
	"projects/util"

	"github.com/golang/protobuf/proto"
)

type ISSClientMsgHandler interface {
	Init() bool
	OnHandler(msgID uint32, datas []byte, sess *SSClientSession)
	OnConnect(sess *SSClientSession)
	OnDisconnect(sess *SSClientSession)
	OnBeatHeartError(sess *SSClientSession)
}

type SSClientSession struct {
	Session
	handler              ISSClientMsgHandler
	timerRegister        etimer.ITimerRegister
	last_beat_heart_time int64
	sessState            uint32
	remoteOuter          string
}

const (
	SS_CLIENT_CHECK_BEAT_HEART_TIME_ID uint32 = 1
	SS_CLIENT_SEND_BEAT_HEART_TIME_ID  uint32 = 2
)

const (
	SS_CLIENT_HEART_TIME_DELAY      uint64 = 1000 * 1
	SS_CLIENT_SEND_BEAT_HEART_DELAY uint64 = 1000 * 20
)

const (
	SS_CLIENT_BEAT_HEART_MAX_TIME int64 = 1000 * 60 * 5
)

func NewSSClientSession(handler ISSClientMsgHandler) *SSClientSession {
	sess := &SSClientSession{
		handler:              handler,
		last_beat_heart_time: util.GetMillsecond(),
		timerRegister:        etimer.NewTimerRegister(),
	}
	sess.SetListenType()
	sess.Session.SessionOnHandler = sess
	return sess
}

func (s *SSClientSession) SetRemoteOuter(outer string) {
	s.remoteOuter = outer
}

func (s *SSClientSession) OnEstablish() {
	elog.InfoAf("SSClientSession %v Establish", s.GetSessID())
	s.factory.AddSession(s)
	s.handler.OnConnect(s)

	s.timerRegister.AddRepeatTimer(SS_CLIENT_CHECK_BEAT_HEART_TIME_ID, SS_CLIENT_HEART_TIME_DELAY, "SSClientSession-BeatHeartCheck", func(v ...interface{}) {
		now := util.GetMillsecond()
		if (s.last_beat_heart_time + SS_CLIENT_BEAT_HEART_MAX_TIME) < now {
			elog.ErrorAf("[SSClientSession] SessID=%v BeatHeart Check Exception", s.GetSessID())
			s.handler.OnBeatHeartError(s)
			s.Terminate()
		} else {
			elog.DebugAf("[SSClientSession] SessID=%v BeatHeart Success", s.GetSessID())
		}
	}, []interface{}{}, true)

	if s.IsConnectType() {
		s.timerRegister.AddRepeatTimer(SS_CLIENT_SEND_BEAT_HEART_TIME_ID, SS_CLIENT_SEND_BEAT_HEART_DELAY, "SSClientSession-SendBeatHeart", func(v ...interface{}) {
			elog.DebugAf("[SSClientSession] SessID=%v Send Beat Heart", s.GetSessID())
			s.AsyncSendMsg(uint32(pb.S2SBaseMsgId_s2s_client_session_ping_id), nil, nil)
		}, []interface{}{}, true)
	}
}

func (s *SSClientSession) OnTerminate() {
	elog.InfoAf("SSClientSession %v Terminate", s.GetSessID())
	s.timerRegister.KillAllTimer()
	factory := s.GetSessionFactory()
	ssclientfactory := factory.(*SSClientSessionMgr)
	ssclientfactory.RemoveSession(s.GetSessID())
	s.handler.OnDisconnect(s)
}

func (s *SSClientSession) OnHandler(msgID uint32, attach_datas []byte, datas []byte) {
	if msgID == uint32(pb.S2SBaseMsgId_s2s_client_session_ping_id) {
		elog.DebugAf("[SSClientSession] SessionID=%v RECV SS_CLIENT_PING_MSG_ID", s.GetSessID())
		s.last_beat_heart_time = util.GetMillsecond()
		s.AsyncSendMsg(uint32(pb.S2SBaseMsgId_s2s_client_session_pong_id), nil, nil)
		return
	} else if msgID == uint32(pb.S2SBaseMsgId_s2s_client_session_pong_id) {
		elog.DebugAf("[SSClientSession] SessionID=%v RECV SS_CLIENT_PONG_MSG_ID", s.GetSessID())
		s.last_beat_heart_time = util.GetMillsecond()
		return
	}

	s.handler.OnHandler(msgID, datas, s)
	s.last_beat_heart_time = util.GetMillsecond()
}

func (s *SSClientSession) SendMsg(msgID uint32, datas []byte, attach inet.IAttachParas) bool {
	return s.AsyncSendMsg(msgID, datas, attach)
}

func (s *SSClientSession) SendProtoMsg(msgID uint32, msg proto.Message, attach inet.IAttachParas) bool {
	return s.AsyncSendProtoMsg(msgID, msg, attach)
}

type SCClientSessionCache struct {
	SessionID   uint64
	Addr        string
	ConnectTick int64
}

const (
	S2S_CLIENTMGR_CACHE_TIMER_ID uint32 = 1
)

const (
	S2S_CLIENTMGR_CACHE_TIMER_DELAY uint64 = 1000 * 1
)

type SSClientSessionMgr struct {
	nextId        uint64
	sessMap       map[uint64]inet.ISession
	handler       ISSClientMsgHandler
	coder         inet.ICoder
	cacheMap      map[uint64]*SCClientSessionCache
	timerRegister etimer.ITimerRegister
}

func (s *SSClientSessionMgr) Init() {
	s.timerRegister.AddRepeatTimer(S2S_CLIENTMGR_CACHE_TIMER_ID, S2S_CLIENTMGR_CACHE_TIMER_DELAY, "SSClientSessionMgr-Cache", func(v ...interface{}) {
		now := util.GetMillsecond()
		for sessionID, cache := range s.cacheMap {
			if cache.ConnectTick < now {
				elog.InfoAf("[SSClientSessionMgr] Timeout Triggle  ConnectCache Del SesssionID=%v,Addr=%v", cache.SessionID, cache.Addr)
				delete(s.cacheMap, sessionID)
			}
		}
	}, []interface{}{}, true)
}

func (s *SSClientSessionMgr) IsInConnectCache(session_id uint64) bool {
	_, ok := s.cacheMap[session_id]
	return ok
}

func (s *SSClientSessionMgr) IsExistSessionOfSessID(session_id uint64) bool {
	_, ok := s.sessMap[session_id]
	return ok
}

func (s *SSClientSessionMgr) CreateSession() inet.ISession {
	sess := NewSSClientSession(s.handler)
	sess.SetSessID(s.nextId)
	sess.SetCoder(s.coder)
	sess.SetSessionFactory(s)
	elog.InfoAf("[SSClientSessionMgr] CreateSession SessID=%v", sess.GetSessID())
	s.nextId++
	return sess
}

func (s *SSClientSessionMgr) FindSession(id uint64) inet.ISession {
	if id == 0 {
		return nil
	}

	if sess, ok := s.sessMap[id]; ok {
		return sess
	}

	return nil
}

func (s *SSClientSessionMgr) AddSession(session inet.ISession) {
	s.sessMap[session.GetSessID()] = session
	if info, ok := s.cacheMap[session.GetSessID()]; ok {
		elog.InfoAf("[SSClientSessionMgr] AddSession Triggle ConnectCache Del SessionID=%v,ServerType=%v", session.GetSessID(), info.Addr)
		delete(s.cacheMap, session.GetSessID())
	}
}

func (s *SSClientSessionMgr) RemoveSession(id uint64) {
	if _, ok := s.sessMap[id]; ok {
		delete(s.sessMap, id)
	}
}

func (s *SSClientSessionMgr) Count() int {
	return len(s.sessMap)
}

func (s *SSClientSessionMgr) AsyncSendProtoMsgBySessionID(sessionID uint64, msgID uint32, msg proto.Message) {
	serversess, ok := s.sessMap[sessionID]
	if ok {
		serversess.AsyncSendProtoMsg(msgID, msg, nil)
	}
}

func (s *SSClientSessionMgr) SSClientConnect(addr string, handler ISSClientMsgHandler, coder inet.ICoder) uint64 {
	if coder == nil {
		coder = NewCoder()
	}

	//handler.Init()
	s.handler = handler
	s.coder = coder
	sess := s.CreateSession()
	ssClientSess := sess.(*SSClientSession)
	ssClientSess.SetRemoteOuter(addr)
	ssClientSess.SetConnectType()

	cache := &SCClientSessionCache{
		SessionID:   sess.GetSessID(),
		Addr:        addr,
		ConnectTick: util.GetMillsecond() + S2S_ONCE_CONNECT_MAX_TIME,
	}
	s.cacheMap[sess.GetSessID()] = cache
	elog.InfoAf("[SSClientSessionMgr]ConnectCache Add SessionID=%v,Addr=%v", sess.GetSessID(), addr)
	enet.GNet.Connect(addr, sess)

	return sess.GetSessID()
}

func (s *SSClientSessionMgr) SSClientListen(addr string, handler ISSClientMsgHandler, coder inet.ICoder) bool {
	if coder == nil {
		coder = NewCoder()
	}
	handler.Init()
	s.handler = handler
	s.coder = coder
	return enet.GNet.Listen(addr, s, math.MaxInt32)
}

var GSSClientSessionMgr *SSClientSessionMgr

func init() {
	GSSClientSessionMgr = &SSClientSessionMgr{
		nextId:        1,
		sessMap:       make(map[uint64]inet.ISession),
		timerRegister: etimer.NewTimerRegister(),
		cacheMap:      make(map[uint64]*SCClientSessionCache),
	}
}
