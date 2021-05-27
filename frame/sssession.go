package frame

import (
	"github.com/zjh-tech/go-frame/base/util"
	"github.com/zjh-tech/go-frame/engine/enet"
	"github.com/zjh-tech/go-frame/engine/etimer"
	"github.com/zjh-tech/go-frame/frame/framepb"

	"github.com/golang/protobuf/proto"
)

const (
	SESS_VERIFY_STATE uint32 = iota
	SESS_ESTABLISH_STATE
	SESS_CLOSE_STATE
)

const (
	SS_CHECK_BEAT_HEART_TIMER_ID uint32 = 1
	SS_SEND_BEAT_HEART_TIMER_ID  uint32 = 2
)

const (
	SS_BEAT_HEART_TIMER_DELAY uint64 = 1000 * 10
	SS_BEAT_SEND_HEART_DELAY  uint64 = 1000 * 30
)

const (
	SSMGR_OUTPUT_TIMER_ID uint32 = 1
)

const (
	SSMGR_OUTPUT_TIMER_DELAY uint64 = 1000 * 60
)

const (
	SS_BEAT_HEART_MAX_TIME   uint64 = 1000 * 60 * 5
	SS_ONCE_CONNECT_MAX_TIME        = 1000 * 10
)

type SSSession struct {
	Session
	timerRegister        etimer.ITimerRegister
	sessState            uint32
	last_beat_heart_time int64
	remoteServerID       uint64
	remoteServerType     uint32
	remoteServerTypeStr  string
	remoteOuter          string
	remoteToken          string
	logicServer          ILogicServer
}

func NewSSSession() *SSSession {
	session := &SSSession{
		timerRegister:        etimer.NewTimerRegister(),
		sessState:            SESS_CLOSE_STATE,
		last_beat_heart_time: util.GetMillsecond(),
	}
	session.SetListenType()
	session.Session.ISessionOnHandler = session
	return session
}

//---------------------------------------------------------------------
func (s *SSSession) SetRemoteServerId(serverID uint64) {
	s.remoteServerID = serverID
}

func (s *SSSession) GetRemoteServerID() uint64 {
	return s.remoteServerID
}

func (s *SSSession) SetRemoteServerType(serverType uint32) {
	s.remoteServerType = serverType
}

func (s *SSSession) SetRemoteServerTypeStr(serverTypeStr string) {
	s.remoteServerTypeStr = serverTypeStr
}

func (s *SSSession) GetRemoteServerType() uint32 {
	return s.remoteServerType
}

func (s *SSSession) SetRemoteOuter(outer string) {
	s.remoteOuter = outer
}

func (s *SSSession) GetRemoteOuter() string {
	return s.remoteOuter
}

func (s *SSSession) SetRemoteToken(token string) {
	s.remoteToken = token
}

func (s *SSSession) GetRemoteToken() string {
	return s.remoteToken
}

func (s *SSSession) SetLogicServer(logicServer ILogicServer) {
	s.logicServer = logicServer
}

//---------------------------------------------------------------------
func (s *SSSession) OnEstablish() {
	s.factory.AddSession(s)
	s.sessState = SESS_VERIFY_STATE
	if s.IsConnectType() {
		ELog.InfoAf("[SSSession] Remote [ID=%v,Type=%v,Ip=%v] Establish Send Verify Req", s.remoteServerID, s.remoteServerType, s.remoteOuter)
		req := &framepb.S2SServerSessionVeriryReq{
			ServerId:      GServer.GetLocalServerID(),
			ServerType:    GServer.GetLocalServerType(),
			ServerTypeStr: GServerCfg.ServiceName,
			Ip:            GServer.GetLocalIp(),
			//Token:         eutil.Md5([]byte(s.remoteToken)),
			Token: s.remoteToken,
		}
		s.AsyncSendProtoMsg(uint32(framepb.S2SBaseMsgId_s2s_server_session_veriry_req_id), req)
		return
	}
}

func (s *SSSession) OnVerify() {
	s.sessState = SESS_ESTABLISH_STATE
	ELog.InfoAf("[SSSession] Remote [ID=%v,Type=%v,Ip=%v] Verify Ok", s.remoteServerID, s.remoteServerType, s.remoteOuter)

	s.timerRegister.AddRepeatTimer(SS_CHECK_BEAT_HEART_TIMER_ID, SS_BEAT_HEART_TIMER_DELAY, "SSSession-CheckBeatHeart", func(v ...interface{}) {
		now := util.GetMillsecond()
		if (s.last_beat_heart_time + int64(SS_BEAT_HEART_MAX_TIME)) < now {
			ELog.ErrorAf("[SSSession] Remote [ID=%v,Type=%v,Ip=%v] BeatHeart Exception", s.remoteServerID, s.remoteServerType, s.remoteOuter)
			s.Terminate()
		}
	}, []interface{}{}, true)

	if s.IsConnectType() {
		s.timerRegister.AddRepeatTimer(SS_SEND_BEAT_HEART_TIMER_ID, SS_BEAT_SEND_HEART_DELAY, "SSSession-SendBeatHeart", func(v ...interface{}) {
			s.AsyncSendMsg(uint32(framepb.S2SBaseMsgId_s2s_server_session_ping_id), nil)
			ELog.DebugAf("[SSSession] Remote [ID=%v,Type=%v,Ip=%v] Send Ping", s.remoteServerID, s.remoteServerType, s.remoteOuter)
		}, []interface{}{}, true)
	}

	factory := s.GetSessionFactory()
	ssserverfactory := factory.(*SSSessionMgr)
	ssserverfactory.GetLogicServerFactory().SetLogicServer(s)
	s.logicServer.SetServerSession(s)
	s.logicServer.OnEstablish(s)
}

func (s *SSSession) OnTerminate() {
	if s.remoteServerID == 0 {
		ELog.InfoAf("[SSSession] SessID=%v  Terminate", s.sess_id)
	} else {
		ELog.InfoAf("[SSSession] SessID=%v [ID=%v,Type=%v,Ip=%v] Terminate", s.sess_id, s.remoteServerID, s.remoteServerType, s.remoteOuter)
	}
	s.timerRegister.KillAllTimer()
	factory := s.GetSessionFactory()
	ssserverfactory := factory.(*SSSessionMgr)
	ssserverfactory.RemoveSession(s.sess_id)
	s.logicServer.SetServerSession(nil)
	s.sessState = SESS_CLOSE_STATE

	s.logicServer.OnTerminate(s)
}

func (s *SSSession) OnHandler(msgId uint32, datas []byte) {
	if msgId == uint32(framepb.S2SBaseMsgId_s2s_server_session_veriry_req_id) && s.IsListenType() {
		verifyReq := &framepb.S2SServerSessionVeriryReq{}
		err := proto.Unmarshal(datas, verifyReq)
		if err != nil {
			ELog.InfoAf("[SSSession] Verify Req Proto Unmarshal Error")
			return
		}

		var VerifyResFunc = func(errorCode uint32) {
			if errorCode == MSG_FAIL {
				s.Terminate()
				return
			}
			s.AsyncSendMsg(uint32(framepb.S2SBaseMsgId_s2s_server_session_veriry_ack_id), nil)
		}

		factory := s.GetSessionFactory()
		ssserverfactory := factory.(*SSSessionMgr)
		if ssserverfactory.FindSessionByServerId(verifyReq.ServerId) != nil {
			//相同的配置的ServerID服务器接入:保留旧的连接,断开新的连接
			ELog.InfoAf("SSSession VerifyReq ServerId=%v Already Exist", verifyReq.ServerId)
			VerifyResFunc(MSG_FAIL)
			return
		}

		s.SetRemoteServerId(verifyReq.ServerId)
		s.SetRemoteServerType(verifyReq.ServerType)
		s.SetRemoteServerTypeStr(verifyReq.ServerTypeStr)
		s.SetRemoteOuter(verifyReq.Ip)

		//if verifyReq.Token != eutil.Md5([]byte(GServer.GetLocalToken())) {
		if verifyReq.Token != GServer.GetLocalToken() {
			ELog.ErrorAf("[SSSession] Remote [ID=%v,Type=%v,Ip=%v] Recv Verify Error", s.remoteServerID, s.remoteServerType, s.remoteOuter)
			VerifyResFunc(MSG_FAIL)
			return
		}

		ELog.InfoAf("[SSSession] Remote [ID=%v,Type=%v,Ip=%v] Recv Verify Ok", s.remoteServerID, s.remoteServerType, s.remoteOuter)
		s.OnVerify()
		VerifyResFunc(MSG_SUCCESS)
		return
	}

	if msgId == uint32(framepb.S2SBaseMsgId_s2s_server_session_veriry_ack_id) && s.IsConnectType() {
		ELog.InfoAf("[SSSession] Remote [ID=%v,Type=%v,Ip=%v] Recv Verify Ack Ok", s.remoteServerID, s.remoteServerType, s.remoteOuter)
		s.OnVerify()
		return
	}

	if msgId == uint32(framepb.S2SBaseMsgId_s2s_server_session_ping_id) && s.IsListenType() {
		ELog.DebugAf("[SSSession] Remote [ID=%v,Type=%v,Ip=%v] Recv Ping Send Pong", s.remoteServerID, s.remoteServerType, s.remoteOuter)
		s.last_beat_heart_time = util.GetMillsecond()
		s.AsyncSendMsg(uint32(framepb.S2SBaseMsgId_s2s_server_session_pong_id), nil)
		return
	}

	if msgId == uint32(framepb.S2SBaseMsgId_s2s_server_session_pong_id) && s.IsConnectType() {
		ELog.DebugAf("[SSSession] Remote [ID=%v,Type=%v,Ip=%v] Recv Pong", s.remoteServerID, s.remoteServerType, s.remoteOuter)
		s.last_beat_heart_time = util.GetMillsecond()
		return
	}

	s.last_beat_heart_time = util.GetMillsecond()
	s.logicServer.OnHandler(msgId, datas, s)
}

//----------------------------------------------------------------------
type SSSessionCache struct {
	ServerID      uint64
	ServerType    uint32
	ServerTypeStr string
	ConnectTick   int64
}

type SSSessionMgr struct {
	nextId             uint64
	sess_map           map[uint64]enet.ISession
	logicServerFactory ILogicServerFactory
	timer_register     etimer.ITimerRegister
	connecting_cache   map[uint64]*SSSessionCache
}

func (s *SSSessionMgr) Init() {
	s.timer_register.AddRepeatTimer(SSMGR_OUTPUT_TIMER_ID, SSMGR_OUTPUT_TIMER_DELAY, "SSSessionMgr-OutPut", func(v ...interface{}) {
		now := util.GetMillsecond()
		for _, session := range s.sess_map {
			serversess := session.(*SSSession)
			ELog.InfoAf("[SSSessionMgr] OutPut ServerId=%v,ServerType=%v", serversess.remoteServerID, serversess.remoteServerTypeStr)
		}

		for serverID, cache := range s.connecting_cache {
			if cache.ConnectTick < now {
				ELog.InfoAf("[SSSessionMgr] Timeout Triggle  ConnectCache Del ServerId=%v,ServerType=%v", cache.ServerID, cache.ServerTypeStr)
				delete(s.connecting_cache, serverID)
			}
		}
	}, []interface{}{}, true)
}

func (s *SSSessionMgr) CreateSession() enet.ISession {
	sess := NewSSSession()
	sess.SetSessID(s.nextId)
	sess.SetCoder(NewCoder())
	sess.SetSessionFactory(s)
	s.nextId++
	ELog.InfoAf("[SSSessionMgr] CreateSession SessID=%v", sess.GetSessID())
	return sess
}

func (s *SSSessionMgr) FindLogicServerByServerType(serverType uint32) []ILogicServer {
	sessArray := make([]ILogicServer, 0)
	for _, session := range s.sess_map {
		serversess := session.(*SSSession)
		if serversess.remoteServerType == serverType {
			sessArray = append(sessArray, serversess.logicServer)
		}
	}

	return sessArray
}

func (s *SSSessionMgr) FindSessionByServerId(serverId uint64) enet.ISession {
	for _, session := range s.sess_map {
		serversess := session.(*SSSession)
		if serversess.remoteServerID == serverId {
			return serversess
		}
	}

	return nil
}

func (s *SSSessionMgr) FindSession(id uint64) enet.ISession {
	if id == 0 {
		return nil
	}

	if sess, ok := s.sess_map[id]; ok {
		return sess
	}
	return nil
}

func (s *SSSessionMgr) GetSessionCount() int {
	return len(s.sess_map)
}

func (s *SSSessionMgr) IsInConnectCache(serverID uint64) bool {
	_, ok := s.connecting_cache[serverID]
	return ok
}

func (s *SSSessionMgr) AddSession(session enet.ISession) {
	s.sess_map[session.GetSessID()] = session
	serversess := session.(*SSSession)
	if _, ok := s.connecting_cache[serversess.GetRemoteServerID()]; ok {
		ELog.InfoAf("[SSSessionMgr] AddSession Triggle ConnectCache Del ServerId=%v,ServerType=%v", serversess.GetRemoteServerID(), serversess.GetRemoteServerType())

		delete(s.connecting_cache, serversess.GetRemoteServerID())
	}
}

func (s *SSSessionMgr) RemoveSession(id uint64) {
	if session, ok := s.sess_map[id]; ok {
		sess := session.(*SSSession)
		if sess.remoteServerID == 0 {
			ELog.InfoAf("[SSSessionMgr] Remove SessID=%v UnInit SSSession", sess.GetSessID())
		} else {
			ELog.InfoAf("[SSSessionMgr] Remove SessID=%v [ID=%v,Type=%v,Ip=%v] SSSession", sess.GetSessID(), sess.remoteServerID, sess.remoteServerType, sess.remoteOuter)
		}
		delete(s.sess_map, id)
	}
}

func (s *SSSessionMgr) SetLogicServerFactory(factory ILogicServerFactory) {
	s.logicServerFactory = factory
}

func (s *SSSessionMgr) GetLogicServerFactory() ILogicServerFactory {
	return s.logicServerFactory
}

func (s *SSSessionMgr) SendMsg(serverId uint64, msgId uint32, datas []byte) {
	for _, session := range s.sess_map {
		serversess := session.(*SSSession)
		if serversess.remoteServerID == serverId {
			serversess.AsyncSendMsg(msgId, datas)
			return
		}
	}
}

func (s *SSSessionMgr) SendProtoMsg(serverId uint64, msgId uint32, msg proto.Message) bool {
	if serverId == 0 {
		return false
	}

	for _, session := range s.sess_map {
		serversess := session.(*SSSession)
		if serversess.remoteServerID == serverId {
			serversess.AsyncSendProtoMsg(msgId, msg)
			return true
		}
	}

	return false
}

func (s *SSSessionMgr) SendProtoMsgBySessionID(sessionID uint64, msgId uint32, msg proto.Message) bool {
	serversess, ok := s.sess_map[sessionID]
	if ok {
		return serversess.AsyncSendProtoMsg(msgId, msg)
	}

	return false
}

func (s *SSSessionMgr) BroadMsg(serverType uint32, msgId uint32, datas []byte) {
	for _, session := range s.sess_map {
		serversess := session.(*SSSession)
		if serversess.remoteServerType == serverType {
			serversess.AsyncSendMsg(msgId, datas)
		}
	}
}

func (s *SSSessionMgr) BroadProtoMsg(serverType uint32, msgId uint32, msg proto.Message) {
	for _, session := range s.sess_map {
		serversess := session.(*SSSession)
		if serversess.remoteServerType == serverType {
			serversess.AsyncSendProtoMsg(msgId, msg)
		}
	}
}

func (s *SSSessionMgr) GetSessionIdByHashIdAndSrvType(hashId uint64, serverType uint32) uint64 {
	sessionId := uint64(0)

	logicServerArray := s.FindLogicServerByServerType(serverType)
	logicServerLen := uint64(len(logicServerArray))
	if logicServerLen == 0 {
		return sessionId
	}

	logicServerIndex := hashId % logicServerLen
	for index, logicServer := range logicServerArray {
		if uint64(index) == logicServerIndex {
			sessionId = logicServer.GetServerSession().GetSessID()
			break
		}
	}

	if sessionId == 0 {
		ELog.ErrorAf("[SSSessionMgr] GetSessionIdByHashId ServerType=%v,HashId=%v Error", serverType, hashId)
	}

	return sessionId
}

func (s *SSSessionMgr) SSServerConnect(ServerID uint64, ServerType uint32, ServerTypeStr string, Outer string, Token string) {
	session := s.CreateSession()
	if session != nil {
		cache := &SSSessionCache{
			ServerID:      ServerID,
			ServerType:    ServerType,
			ServerTypeStr: ServerTypeStr,
			ConnectTick:   util.GetMillsecond() + SS_ONCE_CONNECT_MAX_TIME,
		}
		s.connecting_cache[ServerID] = cache
		ELog.InfoAf("[SSSessionMgr]ConnectCache Add ServerId=%v,ServerType=%v", ServerID, ServerTypeStr)

		serverSession := session.(*SSSession)
		serverSession.SetRemoteServerId(ServerID)
		serverSession.SetRemoteServerType(ServerType)
		serverSession.SetRemoteServerTypeStr(ServerTypeStr)
		serverSession.SetRemoteOuter(Outer)
		serverSession.SetRemoteToken(Token)
		serverSession.SetConnectType()
		enet.GNet.Connect(Outer, serverSession)
	}
}

var GSSSessionMgr *SSSessionMgr

func init() {
	GSSSessionMgr = &SSSessionMgr{
		nextId:           1,
		sess_map:         make(map[uint64]enet.ISession),
		timer_register:   etimer.NewTimerRegister(),
		connecting_cache: make(map[uint64]*SSSessionCache),
	}
}
