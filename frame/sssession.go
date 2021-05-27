package frame

import (
	"encoding/json"

	"github.com/golang/protobuf/proto"
	"github.com/zjh-tech/go-frame/base/util"
	"github.com/zjh-tech/go-frame/engine/enet"
	"github.com/zjh-tech/go-frame/engine/etimer"
)

const (
	SessVerifyState uint32 = iota
	SessEstablishState
	SessCloseState
)

const (
	SSCheckBeatHeartTimerId uint32 = 1
	SSSendBeatHeartTimerId  uint32 = 2
)

const (
	SSBeatHeartTimerDelay uint64 = 1000 * 10
	SSBeatSendHeartDelay  uint64 = 1000 * 30
)

const (
	SSMgrOutputTimerId uint32 = 1
)

const (
	SSMgrOutputTimerDelay uint64 = 1000 * 60
)

const (
	SSBeatHeartMaxTime   uint64 = 1000 * 60 * 5
	SSOnceConnectMaxTime        = 1000 * 10
)

type LocalSessionSpec struct {
	ServerID      uint64
	ServerType    uint32
	ServerTypeStr string
	Ip            string
	Token         string
}

type RemoteSessionSpec struct {
	ServerID      uint64
	ServerType    uint32
	ServerTypeStr string
	Ip            string
}

type SSSession struct {
	Session
	timerRegister     etimer.ITimerRegister
	sessState         uint32
	lastBeatHeartTime int64
	localSpec         LocalSessionSpec
	remoteSpec        RemoteSessionSpec
	logicServer       ILogicServer
}

func NewSSSession() *SSSession {
	session := &SSSession{
		timerRegister:     etimer.NewTimerRegister(),
		sessState:         SessCloseState,
		lastBeatHeartTime: util.GetMillsecond(),
	}
	session.SetListenType()
	session.Session.ISessionOnHandler = session
	return session
}

//------------------------------------------------------------
func (s *SSSession) SetLocalSpec(localSpec LocalSessionSpec) {
	s.localSpec = localSpec
}

func (s *SSSession) SetRemoteSpec(remoteSpec RemoteSessionSpec) {
	s.remoteSpec = remoteSpec
}

func (s *SSSession) GetRemoteServerID() uint64 {
	return s.remoteSpec.ServerID
}

func (s *SSSession) GetRemoteServerType() uint32 {
	return s.remoteSpec.ServerType
}

func (s *SSSession) SetLogicServer(logicServer ILogicServer) {
	s.logicServer = logicServer
}

//---------------------------------------------------------------------
func (s *SSSession) OnEstablish() {
	s.factory.AddSession(s)
	s.sessState = SessVerifyState
	if s.IsConnectType() {
		ELog.InfoAf("[SSSession] Remote [ID=%v,Type=%v,Ip=%v] Establish Send Verify Req", s.remoteSpec.ServerID, s.remoteSpec.ServerType, s.remoteSpec.Ip)
		req := &S2SSessionVerifyReq{
			ServerId:      s.localSpec.ServerID,
			ServerType:    s.localSpec.ServerType,
			ServerTypeStr: s.localSpec.ServerTypeStr,
			Ip:            s.localSpec.Ip,
			Token:         s.localSpec.Token,
		}

		datas, marshalErr := json.Marshal(req)
		if marshalErr == nil {
			s.AsyncSendMsg(S2SSessionVerifyReqId, datas)
		}
		return
	}
}

func (s *SSSession) OnVerify() {
	s.sessState = SessEstablishState
	ELog.InfoAf("[SSSession] Remote [ID=%v,Type=%v,Ip=%v] Verify Ok", s.remoteSpec.ServerID, s.remoteSpec.ServerType, s.remoteSpec.Ip)

	s.timerRegister.AddRepeatTimer(SSCheckBeatHeartTimerId, SSBeatHeartTimerDelay, "SSSession-CheckBeatHeart", func(v ...interface{}) {
		now := util.GetMillsecond()
		if (s.lastBeatHeartTime + int64(SSBeatHeartMaxTime)) < now {
			ELog.ErrorAf("[SSSession] Remote [ID=%v,Type=%v,Ip=%v] BeatHeart Exception", s.remoteSpec.ServerID, s.remoteSpec.ServerType, s.remoteSpec.Ip)
			s.Terminate()
		}
	}, []interface{}{}, true)

	if s.IsConnectType() {
		s.timerRegister.AddRepeatTimer(SSSendBeatHeartTimerId, SSBeatSendHeartDelay, "SSSession-SendBeatHeart", func(v ...interface{}) {
			s.AsyncSendMsg(S2SSessionPingId, nil)
			ELog.DebugAf("[SSSession] Remote [ID=%v,Type=%v,Ip=%v] Send Ping", s.remoteSpec.ServerID, s.remoteSpec.ServerType, s.remoteSpec.Ip)
		}, []interface{}{}, true)
	}

	factory := s.GetSessionFactory()
	ssserverfactory := factory.(*SSSessionMgr)
	ssserverfactory.GetLogicServerFactory().SetLogicServer(s)
	s.logicServer.SetServerSession(s)
	s.logicServer.OnEstablish(s)
}

func (s *SSSession) OnTerminate() {
	if s.remoteSpec.ServerID == 0 {
		ELog.InfoAf("[SSSession] SessID=%v  Terminate", s.sessId)
	} else {
		ELog.InfoAf("[SSSession] SessID=%v [ID=%v,Type=%v,Ip=%v] Terminate", s.sessId, s.remoteSpec.ServerID, s.remoteSpec.ServerType, s.remoteSpec.Ip)
	}
	s.timerRegister.KillAllTimer()
	factory := s.GetSessionFactory()
	ssserverfactory := factory.(*SSSessionMgr)
	ssserverfactory.RemoveSession(s.sessId)
	s.logicServer.SetServerSession(nil)
	s.sessState = SessCloseState

	s.logicServer.OnTerminate(s)
}

func (s *SSSession) OnHandler(msgId uint32, datas []byte) {
	if msgId == S2SSessionVerifyReqId && s.IsListenType() {
		verifyReq := &S2SSessionVerifyReq{}
		err := json.Unmarshal(datas, verifyReq)
		if err != nil {
			ELog.InfoAf("[SSSession] S2SSessionVerifyReq Json Unmarshal Error")
			return
		}

		var VerifyResFunc = func(errorCode uint32) {
			if errorCode == MsgFail {
				s.Terminate()
				return
			}
			s.AsyncSendMsg(S2SSessionVerifyResId, nil)
		}

		factory := s.GetSessionFactory()
		ssserverfactory := factory.(*SSSessionMgr)
		if ssserverfactory.FindSessionByServerId(verifyReq.ServerId) != nil {
			//相同的配置的ServerID服务器接入:保留旧的连接,断开新的连接
			ELog.InfoAf("SSSession VerifyReq ServerId=%v Already Exist", verifyReq.ServerId)
			VerifyResFunc(MsgFail)
			return
		}

		var remoteSpec RemoteSessionSpec
		remoteSpec.ServerID = verifyReq.ServerId
		remoteSpec.ServerType = verifyReq.ServerType
		remoteSpec.ServerTypeStr = verifyReq.ServerTypeStr
		remoteSpec.Ip = verifyReq.Ip
		s.SetRemoteSpec(remoteSpec)

		//if verifyReq.Token != eutil.Md5([]byte(GServer.GetLocalToken())) {
		if verifyReq.Token != GServer.GetLocalToken() {
			ELog.ErrorAf("[SSSession] Remote [ID=%v,Type=%v,Ip=%v] Recv Verify Error", s.remoteSpec.ServerID, s.remoteSpec.ServerType, s.remoteSpec.Ip)
			VerifyResFunc(MsgFail)
			return
		}

		ELog.InfoAf("[SSSession] Remote [ID=%v,Type=%v,Ip=%v] Recv Verify Ok", s.remoteSpec.ServerID, s.remoteSpec.ServerType, s.remoteSpec.Ip)
		s.OnVerify()
		VerifyResFunc(MsgSuccess)
		return
	}

	if msgId == S2SSessionVerifyResId && s.IsConnectType() {
		ELog.InfoAf("[SSSession] Remote [ID=%v,Type=%v,Ip=%v] Recv Verify Ack Ok", s.remoteSpec.ServerID, s.remoteSpec.ServerType, s.remoteSpec.Ip)
		s.OnVerify()
		return
	}

	if msgId == S2SSessionPingId && s.IsListenType() {
		ELog.DebugAf("[SSSession] Remote [ID=%v,Type=%v,Ip=%v] Recv Ping Send Pong", s.remoteSpec.ServerID, s.remoteSpec.ServerType, s.remoteSpec.Ip)
		s.lastBeatHeartTime = util.GetMillsecond()
		s.AsyncSendMsg(S2SSessionPongId, nil)
		return
	}

	if msgId == S2SSessionPongId && s.IsConnectType() {
		ELog.DebugAf("[SSSession] Remote [ID=%v,Type=%v,Ip=%v] Recv Pong", s.remoteSpec.ServerID, s.remoteSpec.ServerType, s.remoteSpec.Ip)
		s.lastBeatHeartTime = util.GetMillsecond()
		return
	}

	s.lastBeatHeartTime = util.GetMillsecond()
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
	sessMap            map[uint64]enet.ISession
	logicServerFactory ILogicServerFactory
	timerRegister      etimer.ITimerRegister
	connectingCache    map[uint64]*SSSessionCache
}

func (s *SSSessionMgr) Init() {
	s.timerRegister.AddRepeatTimer(SSMgrOutputTimerId, SSMgrOutputTimerDelay, "SSSessionMgr-OutPut", func(v ...interface{}) {
		now := util.GetMillsecond()
		for _, session := range s.sessMap {
			serversess := session.(*SSSession)
			ELog.InfoAf("[SSSessionMgr] OutPut ServerId=%v,ServerType=%v", serversess.remoteSpec.ServerID, serversess.remoteSpec.ServerType)
		}

		for serverID, cache := range s.connectingCache {
			if cache.ConnectTick < now {
				ELog.InfoAf("[SSSessionMgr] Timeout Triggle  ConnectCache Del ServerId=%v,ServerType=%v", cache.ServerID, cache.ServerTypeStr)
				delete(s.connectingCache, serverID)
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
	for _, session := range s.sessMap {
		serversess := session.(*SSSession)
		if serversess.remoteSpec.ServerType == serverType {
			sessArray = append(sessArray, serversess.logicServer)
		}
	}

	return sessArray
}

func (s *SSSessionMgr) FindSessionByServerId(serverId uint64) enet.ISession {
	for _, session := range s.sessMap {
		serversess := session.(*SSSession)
		if serversess.remoteSpec.ServerID == serverId {
			return serversess
		}
	}

	return nil
}

func (s *SSSessionMgr) FindSession(id uint64) enet.ISession {
	if id == 0 {
		return nil
	}

	if sess, ok := s.sessMap[id]; ok {
		return sess
	}
	return nil
}

func (s *SSSessionMgr) GetSessionCount() int {
	return len(s.sessMap)
}

func (s *SSSessionMgr) IsInConnectCache(serverID uint64) bool {
	_, ok := s.connectingCache[serverID]
	return ok
}

func (s *SSSessionMgr) AddSession(session enet.ISession) {
	s.sessMap[session.GetSessID()] = session
	serversess := session.(*SSSession)
	if _, ok := s.connectingCache[serversess.GetRemoteServerID()]; ok {
		ELog.InfoAf("[SSSessionMgr] AddSession Triggle ConnectCache Del ServerId=%v,ServerType=%v", serversess.GetRemoteServerID(), serversess.GetRemoteServerType())

		delete(s.connectingCache, serversess.GetRemoteServerID())
	}
}

func (s *SSSessionMgr) RemoveSession(id uint64) {
	if session, ok := s.sessMap[id]; ok {
		sess := session.(*SSSession)
		if sess.remoteSpec.ServerID == 0 {
			ELog.InfoAf("[SSSessionMgr] Remove SessID=%v UnInit SSSession", sess.GetSessID())
		} else {
			ELog.InfoAf("[SSSessionMgr] Remove SessID=%v [ID=%v,Type=%v,Ip=%v] SSSession", sess.GetSessID(), sess.remoteSpec.ServerID, sess.remoteSpec.ServerType, sess.remoteSpec.Ip)
		}
		delete(s.sessMap, id)
	}
}

func (s *SSSessionMgr) SetLogicServerFactory(factory ILogicServerFactory) {
	s.logicServerFactory = factory
}

func (s *SSSessionMgr) GetLogicServerFactory() ILogicServerFactory {
	return s.logicServerFactory
}

func (s *SSSessionMgr) SendMsg(serverId uint64, msgId uint32, datas []byte) {
	for _, session := range s.sessMap {
		serversess := session.(*SSSession)
		if serversess.remoteSpec.ServerID == serverId {
			serversess.AsyncSendMsg(msgId, datas)
			return
		}
	}
}

func (s *SSSessionMgr) SendProtoMsg(serverId uint64, msgId uint32, msg proto.Message) bool {
	if serverId == 0 {
		return false
	}

	for _, session := range s.sessMap {
		serversess := session.(*SSSession)
		if serversess.remoteSpec.ServerID == serverId {
			serversess.AsyncSendProtoMsg(msgId, msg)
			return true
		}
	}

	return false
}

func (s *SSSessionMgr) SendProtoMsgBySessionID(sessionID uint64, msgId uint32, msg proto.Message) bool {
	serversess, ok := s.sessMap[sessionID]
	if ok {
		return serversess.AsyncSendProtoMsg(msgId, msg)
	}

	return false
}

func (s *SSSessionMgr) BroadMsg(serverType uint32, msgId uint32, datas []byte) {
	for _, session := range s.sessMap {
		serversess := session.(*SSSession)
		if serversess.GetRemoteServerType() == serverType {
			serversess.AsyncSendMsg(msgId, datas)
		}
	}
}

func (s *SSSessionMgr) BroadProtoMsg(serverType uint32, msgId uint32, msg proto.Message) {
	for _, session := range s.sessMap {
		serversess := session.(*SSSession)
		if serversess.GetRemoteServerType() == serverType {
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

func (s *SSSessionMgr) SSServerConnect(localSpec LocalSessionSpec, remoteSepc RemoteSessionSpec) {
	session := s.CreateSession()
	if session != nil {
		cache := &SSSessionCache{
			ServerID:      remoteSepc.ServerID,
			ServerType:    remoteSepc.ServerType,
			ServerTypeStr: remoteSepc.ServerTypeStr,
			ConnectTick:   util.GetMillsecond() + SSOnceConnectMaxTime,
		}
		s.connectingCache[remoteSepc.ServerID] = cache
		ELog.InfoAf("[SSSessionMgr]ConnectCache Add ServerId=%v,ServerType=%v", remoteSepc.ServerID, remoteSepc.ServerTypeStr)

		serverSession := session.(*SSSession)
		serverSession.SetLocalSpec(localSpec)
		serverSession.SetRemoteSpec(remoteSepc)

		serverSession.SetConnectType()
		enet.GNet.Connect(remoteSepc.Ip, serverSession)
	}
}

var GSSSessionMgr *SSSessionMgr

func init() {
	GSSSessionMgr = &SSSessionMgr{
		nextId:          1,
		sessMap:         make(map[uint64]enet.ISession),
		timerRegister:   etimer.NewTimerRegister(),
		connectingCache: make(map[uint64]*SSSessionCache),
	}
}
