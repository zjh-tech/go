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

const (
	SESS_VERIFY_STATE uint32 = iota
	SESS_ESTABLISH_STATE
	SESS_CLOSE_STATE
)

const (
	S2S_CHECK_BEAT_HEART_TIMER_ID uint32 = 1
	S2S_SEND_BEAT_HEART_TIMER_ID  uint32 = 2
)

const (
	S2S_BEAT_HEART_TIMER_DELAY uint64 = 1000 * 10
	S2S_BEAT_SEND_HEART_DELAY  uint64 = 1000 * 30
)

const (
	S2SMGR_OUTPUT_TIMER_ID uint32 = 1
)

const (
	S2SMGR_OUTPUT_TIMER_DELAY uint64 = 1000 * 60
)

const (
	S2S_BEAT_HEART_MAX_TIME   uint64 = 1000 * 60 * 5
	S2S_ONCE_CONNECT_MAX_TIME        = 1000 * 10
)

type SSServerSession struct {
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

func NewSSServerSession() *SSServerSession {
	session := &SSServerSession{
		timerRegister:        etimer.NewTimerRegister(),
		sessState:            SESS_CLOSE_STATE,
		last_beat_heart_time: util.GetMillsecond(),
	}
	session.SetListenType()
	session.Session.SessionOnHandler = session
	return session
}

//---------------------------------------------------------------------
func (s *SSServerSession) SetRemoteServerId(serverID uint64) {
	s.remoteServerID = serverID
}

func (s *SSServerSession) GetRemoteServerID() uint64 {
	return s.remoteServerID
}

func (s *SSServerSession) SetRemoteServerType(serverType uint32) {
	s.remoteServerType = serverType
}

func (s *SSServerSession) SetRemoteServerTypeStr(serverTypeStr string) {
	s.remoteServerTypeStr = serverTypeStr
}

func (s *SSServerSession) GetRemoteServerType() uint32 {
	return s.remoteServerType
}

func (s *SSServerSession) SetRemoteOuter(outer string) {
	s.remoteOuter = outer
}

func (s *SSServerSession) GetRemoteOuter() string {
	return s.remoteOuter
}

func (s *SSServerSession) SetRemoteToken(token string) {
	s.remoteToken = token
}

func (s *SSServerSession) GetRemoteToken() string {
	return s.remoteToken
}

func (s *SSServerSession) SetLogicServer(logicServer ILogicServer) {
	s.logicServer = logicServer
}

func (s *SSServerSession) SendMsg(msgID uint32, datas []byte, attach inet.IAttachParas) bool {
	return s.AsyncSendMsg(msgID, datas, attach)
}

func (s *SSServerSession) SendProtoMsg(msgID uint32, msg proto.Message, attach inet.IAttachParas) bool {
	return s.AsyncSendProtoMsg(msgID, msg, attach)
}

//---------------------------------------------------------------------
func (s *SSServerSession) OnEstablish() {
	s.factory.AddSession(s)
	s.sessState = SESS_VERIFY_STATE
	if s.IsConnectType() {
		elog.InfoAf("[SSServerSession] Remote [ID=%v,Type=%v,Ip=%v] Establish Send Verify Req", s.remoteServerID, s.remoteServerType, s.remoteOuter)
		req := &pb.S2SServerSessionVeriryReq{
			ServerId:      GServer.GetLocalServerID(),
			ServerType:    GServer.GetLocalServerType(),
			ServerTypeStr: GServerCfg.ServiceName,
			Ip:            GServer.GetLocalIp(),
			//Token:         eutil.Md5([]byte(s.remoteToken)),
			Token: s.remoteToken,
		}
		s.AsyncSendProtoMsg(uint32(pb.S2SBaseMsgId_s2s_server_session_veriry_req_id), req, nil)
		return
	}
}

func (s *SSServerSession) OnVerify() {
	s.sessState = SESS_ESTABLISH_STATE
	elog.InfoAf("[SSServerSession] Remote [ID=%v,Type=%v,Ip=%v] Verify Ok", s.remoteServerID, s.remoteServerType, s.remoteOuter)

	s.timerRegister.AddRepeatTimer(S2S_CHECK_BEAT_HEART_TIMER_ID, S2S_BEAT_HEART_TIMER_DELAY, "SSServerSession-CheckBeatHeart", func(v ...interface{}) {
		now := util.GetMillsecond()
		if (s.last_beat_heart_time + int64(S2S_BEAT_HEART_MAX_TIME)) < now {
			elog.ErrorAf("[SSServerSession] Remote [ID=%v,Type=%v,Ip=%v] BeatHeart Exception", s.remoteServerID, s.remoteServerType, s.remoteOuter)
			s.Terminate()
		}
	}, []interface{}{}, true)

	if s.IsConnectType() {
		s.timerRegister.AddRepeatTimer(S2S_SEND_BEAT_HEART_TIMER_ID, S2S_BEAT_SEND_HEART_DELAY, "SSServerSession-SendBeatHeart", func(v ...interface{}) {
			s.AsyncSendMsg(uint32(pb.S2SBaseMsgId_s2s_server_session_ping_id), nil, nil)
			elog.DebugAf("[SSServerSession] Remote [ID=%v,Type=%v,Ip=%v] Send Ping", s.remoteServerID, s.remoteServerType, s.remoteOuter)
		}, []interface{}{}, true)
	}

	factory := s.GetSessionFactory()
	ssserverfactory := factory.(*SSServerSessionMgr)
	ssserverfactory.GetLogicServerFactory().SetLogicServer(s)
	s.logicServer.SetServerSession(s)
	s.logicServer.OnEstablish(s)
}

func (s *SSServerSession) OnTerminate() {
	if s.remoteServerID == 0 {
		elog.InfoAf("[SSServerSession] SessID=%v  Terminate", s.sess_id)
	} else {
		elog.InfoAf("[SSServerSession] SessID=%v [ID=%v,Type=%v,Ip=%v] Terminate", s.sess_id, s.remoteServerID, s.remoteServerType, s.remoteOuter)
	}

	s.timerRegister.KillAllTimer()
	factory := s.GetSessionFactory()
	ssserverfactory := factory.(*SSServerSessionMgr)
	ssserverfactory.RemoveSession(s.sess_id)
	if s.sessState == SESS_ESTABLISH_STATE {
		s.logicServer.SetServerSession(nil)
		s.logicServer.OnTerminate(s)
	}
	s.sessState = SESS_CLOSE_STATE
}

func (s *SSServerSession) OnHandler(msgID uint32, attach_datas []byte, datas []byte) {
	if msgID == uint32(pb.S2SBaseMsgId_s2s_server_session_veriry_req_id) && s.IsListenType() {
		verifyReq := &pb.S2SServerSessionVeriryReq{}
		err := proto.Unmarshal(datas, verifyReq)
		if err != nil {
			elog.InfoAf("[SSServerSession] Verify Req Proto Unmarshal Error")
			return
		}

		var VerifyResFunc = func(errorCode uint32) {
			if errorCode == MSG_FAIL {
				s.Terminate()
				return
			}
			s.AsyncSendMsg(uint32(pb.S2SBaseMsgId_s2s_server_session_veriry_ack_id), nil, nil)
		}

		factory := s.GetSessionFactory()
		ssserverfactory := factory.(*SSServerSessionMgr)
		if ssserverfactory.FindSessionByServerId(verifyReq.ServerId) != nil {
			//相同的配置的ServerID服务器接入:保留旧的连接,断开新的连接
			elog.InfoAf("SSServerSession VerifyReq ServerId=%v Already Exist", verifyReq.ServerId)
			VerifyResFunc(MSG_FAIL)
			return
		}

		s.SetRemoteServerId(verifyReq.ServerId)
		s.SetRemoteServerType(verifyReq.ServerType)
		s.SetRemoteServerTypeStr(verifyReq.ServerTypeStr)
		s.SetRemoteOuter(verifyReq.Ip)

		//if verifyReq.Token != eutil.Md5([]byte(GServer.GetLocalToken())) {
		if verifyReq.Token != GServer.GetLocalToken() {
			elog.ErrorAf("[SSServerSession] Remote [ID=%v,Type=%v,Ip=%v] Recv Verify Error", s.remoteServerID, s.remoteServerType, s.remoteOuter)
			VerifyResFunc(MSG_FAIL)
			return
		}

		elog.InfoAf("[SSServerSession] Remote [ID=%v,Type=%v,Ip=%v] Recv Verify Ok", s.remoteServerID, s.remoteServerType, s.remoteOuter)
		s.OnVerify()
		VerifyResFunc(MSG_SUCCESS)
		return
	}

	if msgID == uint32(pb.S2SBaseMsgId_s2s_server_session_veriry_ack_id) && s.IsConnectType() {
		elog.InfoAf("[SSServerSession] Remote [ID=%v,Type=%v,Ip=%v] Recv Verify Ack Ok", s.remoteServerID, s.remoteServerType, s.remoteOuter)
		s.OnVerify()
		return
	}

	if msgID == uint32(pb.S2SBaseMsgId_s2s_server_session_ping_id) && s.IsListenType() {
		elog.DebugAf("[SSServerSession] Remote [ID=%v,Type=%v,Ip=%v] Recv Ping Send Pong", s.remoteServerID, s.remoteServerType, s.remoteOuter)
		s.last_beat_heart_time = util.GetMillsecond()
		s.AsyncSendMsg(uint32(pb.S2SBaseMsgId_s2s_server_session_pong_id), nil, nil)
		return
	}

	if msgID == uint32(pb.S2SBaseMsgId_s2s_server_session_pong_id) && s.IsConnectType() {
		elog.DebugAf("[SSServerSession] Remote [ID=%v,Type=%v,Ip=%v] Recv Pong", s.remoteServerID, s.remoteServerType, s.remoteOuter)
		s.last_beat_heart_time = util.GetMillsecond()
		return
	}

	s.last_beat_heart_time = util.GetMillsecond()
	s.logicServer.OnHandler(msgID, attach_datas, datas, s)
}

//----------------------------------------------------------------------
type SSServerSessionCache struct {
	ServerID      uint64
	ServerType    uint32
	ServerTypeStr string
	ConnectTick   int64
}

type SSServerSessionMgr struct {
	nextId             uint64
	sessMap            map[uint64]inet.ISession
	logicServerFactory ILogicServerFactory
	timerRegister      etimer.ITimerRegister
	connectingCache    map[uint64]*SSServerSessionCache
}

func (s *SSServerSessionMgr) Init() {
	s.timerRegister.AddRepeatTimer(S2SMGR_OUTPUT_TIMER_ID, S2SMGR_OUTPUT_TIMER_DELAY, "SSServerSessionMgr-OutPut", func(v ...interface{}) {
		now := util.GetMillsecond()
		for _, session := range s.sessMap {
			serversess := session.(*SSServerSession)
			elog.InfoAf("[SSServerSessionMgr] OutPut ServerId=%v,ServerType=%v", serversess.remoteServerID, serversess.remoteServerTypeStr)
		}

		for serverID, cache := range s.connectingCache {
			if cache.ConnectTick < now {
				elog.InfoAf("[SSServerSessionMgr] Timeout Triggle  ConnectCache Del ServerId=%v,ServerType=%v", cache.ServerID, cache.ServerTypeStr)
				delete(s.connectingCache, serverID)
			}
		}
	}, []interface{}{}, true)
}

func (s *SSServerSessionMgr) CreateSession() inet.ISession {
	sess := NewSSServerSession()
	sess.SetSessID(s.nextId)
	sess.SetCoder(NewCoder())
	sess.SetSessionFactory(s)
	s.nextId++
	elog.InfoAf("[SSServerSessionMgr] CreateSession SessID=%v", sess.GetSessID())
	return sess
}

func (s *SSServerSessionMgr) FindLogicServerByServerType(serverType uint32) []ILogicServer {
	sessArray := make([]ILogicServer, 0)
	for _, session := range s.sessMap {
		serversess := session.(*SSServerSession)
		if serversess.remoteServerType == serverType {
			sessArray = append(sessArray, serversess.logicServer)
		}
	}

	return sessArray
}

func (s *SSServerSessionMgr) FindSessionByServerId(serverId uint64) inet.ISession {
	for _, session := range s.sessMap {
		serversess := session.(*SSServerSession)
		if serversess.remoteServerID == serverId {
			return serversess
		}
	}

	return nil
}

func (s *SSServerSessionMgr) FindSession(id uint64) inet.ISession {
	if id == 0 {
		return nil
	}

	if sess, ok := s.sessMap[id]; ok {
		return sess
	}
	return nil
}

func (s *SSServerSessionMgr) IsInConnectCache(serverID uint64) bool {
	_, ok := s.connectingCache[serverID]
	return ok
}

func (s *SSServerSessionMgr) AddSession(session inet.ISession) {
	s.sessMap[session.GetSessID()] = session
	serversess := session.(*SSServerSession)
	if _, ok := s.connectingCache[serversess.GetRemoteServerID()]; ok {
		elog.InfoAf("[SSServerSessionMgr] AddSession Triggle ConnectCache Del ServerId=%v,ServerType=%v", serversess.GetRemoteServerID(), serversess.GetRemoteServerType())

		delete(s.connectingCache, serversess.GetRemoteServerID())
	}
}

func (s *SSServerSessionMgr) RemoveSession(id uint64) {
	if session, ok := s.sessMap[id]; ok {
		sess := session.(*SSServerSession)
		if sess.remoteServerID == 0 {
			elog.InfoAf("[SSServerSessionMgr] Remove SessID=%v UnInit SSServerSession", sess.GetSessID())
		} else {
			elog.InfoAf("[SSServerSessionMgr] Remove SessID=%v [ID=%v,Type=%v,Ip=%v] SSServerSession", sess.GetSessID(), sess.remoteServerID, sess.remoteServerType, sess.remoteOuter)
		}
		delete(s.sessMap, id)
	}
}

func (s *SSServerSessionMgr) SetLogicServerFactory(factory ILogicServerFactory) {
	s.logicServerFactory = factory
}

func (s *SSServerSessionMgr) GetLogicServerFactory() ILogicServerFactory {
	return s.logicServerFactory
}

func (s *SSServerSessionMgr) SendMsg(serverId uint64, msgID uint32, datas []byte) {
	for _, session := range s.sessMap {
		serversess := session.(*SSServerSession)
		if serversess.remoteServerID == serverId {
			serversess.AsyncSendMsg(msgID, datas, nil)
			return
		}
	}
}

func (s *SSServerSessionMgr) SendProtoMsg(serverId uint64, msgID uint32, msg proto.Message) bool {
	if serverId == 0 {
		return false
	}

	for _, session := range s.sessMap {
		serversess := session.(*SSServerSession)
		if serversess.remoteServerID == serverId {
			serversess.AsyncSendProtoMsg(msgID, msg, nil)
			return false
		}
	}

	return true
}

func (s *SSServerSessionMgr) SendProtoMsgBySessionID(sessionID uint64, msgID uint32, msg proto.Message) bool {
	serversess, ok := s.sessMap[sessionID]
	if ok {
		return serversess.AsyncSendProtoMsg(msgID, msg, nil)
	}

	return false
}

func (s *SSServerSessionMgr) BroadMsg(serverType uint32, msgID uint32, datas []byte) {
	for _, session := range s.sessMap {
		serversess := session.(*SSServerSession)
		if serversess.remoteServerType == serverType {
			serversess.AsyncSendMsg(msgID, datas, nil)
		}
	}
}

func (s *SSServerSessionMgr) BroadProtoMsg(serverType uint32, msgID uint32, msg proto.Message) {
	for _, session := range s.sessMap {
		serversess := session.(*SSServerSession)
		if serversess.remoteServerType == serverType {
			serversess.AsyncSendProtoMsg(msgID, msg, nil)
		}
	}
}

func (s *SSServerSessionMgr) GetSessionIdByHashIdAndSrvType(hashId uint64, serverType uint32) uint64 {
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
		elog.ErrorAf("[SSServerSessionMgr] GetSessionIdByHashId ServerType=%v,HashId=%v Error", serverType, hashId)
	}

	return sessionId
}

//func (s *SSServerSessionMgr) SendProtoMsgByHashIdAndSrvType(hashId uint64, serverType uint32, msgID uint32, msg proto.Message) bool {
//	logicServerArray := s.FindLogicServerByServerType(serverType)
//	logicServerLen := uint64(len(logicServerArray))
//	if logicServerLen == 0 {
//		return false
//	}
//
//	var serverSess *SSServerSession
//	logicServerIndex := hashId % logicServerLen
//	for index, logicServer := range logicServerArray {
//		if uint64(index) == logicServerIndex {
//			if logicServer.GetServerSession() != nil {
//				serverSess = logicServer.GetServerSession()
//			}
//			break
//		}
//	}
//
//	if serverSess == nil {
//		elog.ErrorAf("[SSServerSessionMgr] GetServerIdByHashId ServerType=%v,HashId=%v Error", serverType, hashId)
//		return false
//	}
//
//	serverSess.AsyncSendProtoMsg(msgID, msg, nil)
//	return true
//}

func (s *SSServerSessionMgr) SSServerConnect(ServerID uint64, ServerType uint32, ServerTypeStr string, Outer string, Token string) {
	session := s.CreateSession()
	if session != nil {
		cache := &SSServerSessionCache{
			ServerID:      ServerID,
			ServerType:    ServerType,
			ServerTypeStr: ServerTypeStr,
			ConnectTick:   util.GetMillsecond() + S2S_ONCE_CONNECT_MAX_TIME,
		}
		s.connectingCache[ServerID] = cache
		elog.InfoAf("[SSServerSessionMgr]ConnectCache Add ServerId=%v,ServerType=%v", ServerID, ServerTypeStr)

		serverSession := session.(*SSServerSession)
		serverSession.SetRemoteServerId(ServerID)
		serverSession.SetRemoteServerType(ServerType)
		serverSession.SetRemoteServerTypeStr(ServerTypeStr)
		serverSession.SetRemoteOuter(Outer)
		serverSession.SetRemoteToken(Token)
		serverSession.SetConnectType()
		enet.GNet.Connect(Outer, serverSession)
	}
}

var GSSServerSessionMgr *SSServerSessionMgr

func init() {
	GSSServerSessionMgr = &SSServerSessionMgr{
		nextId:          1,
		sessMap:         make(map[uint64]inet.ISession),
		timerRegister:   etimer.NewTimerRegister(),
		connectingCache: make(map[uint64]*SSServerSessionCache),
	}
}
