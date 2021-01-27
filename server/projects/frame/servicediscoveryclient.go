package frame

import (
	"math"
	"projects/go-engine/elog"
	"projects/go-engine/enet"
	"projects/go-engine/etimer"
	"projects/pb"

	"github.com/golang/protobuf/proto"
)

const (
	SD_CLIENT_SEND_REQ_TIMER_ID  uint32 = 1
	SD_CLIENT_RECONNECT_TIMER_ID uint32 = 2
)

const (
	SD_CLIENT_SEND_REQ_TIMER_DELAY  uint64 = 1000 * 3
	SD_CLIENT_RECONNECT_TIMER_DELAY uint64 = 1000 * 10
)

type SDCbFunc func(...interface{})

type ServiceDiscoveryClient struct {
	addr                 string
	server_id            uint64
	token                string
	cb_func              SDCbFunc
	init_flag            bool
	time_register        etimer.ITimerRegister
	ssclient_session_mgr *SSClientSessionMgr
	ssclient_session     *SSClientSession
	session_id           uint64
}

func (s *ServiceDiscoveryClient) Init(_addr string, _server_id uint64, _token string, _cb SDCbFunc) bool {
	s.addr = _addr
	s.server_id = _server_id
	s.token = _token
	s.init_flag = false
	s.cb_func = _cb
	s.ssclient_session_mgr = GSSClientSessionMgr
	s.session_id = s.ssclient_session_mgr.SSClientConnect(_addr, GSDServerSession, nil)
	elog.InfoAf("[ServiceDiscoveryClient] Connect SessionID=%v Addr=%v", s.session_id, _addr)

	s.time_register.AddRepeatTimer(SD_CLIENT_SEND_REQ_TIMER_ID, SD_CLIENT_SEND_REQ_TIMER_DELAY, "SDClient-SendSDReq", func(v ...interface{}) {
		if s.ssclient_session == nil {
			return
		}

		serverID := v[0].(uint64)
		token := v[1].(string)
		req := &pb.ServiceDiscoveryReq{
			ServerId: serverID,
			Token:    token,
		}
		s.ssclient_session.SendProtoMsg(uint32(pb.S2SBaseMsgId_service_discovery_req_id), req, nil)
		elog.DebugAf("[ServiceDiscoveryClient] Send ServiceDiscoveryReq ServerID=%v", serverID)
	}, []interface{}{_server_id, _token}, true)

	s.time_register.AddRepeatTimer(SD_CLIENT_RECONNECT_TIMER_ID, SD_CLIENT_RECONNECT_TIMER_DELAY, "SDClient-CheckReconnect", func(v ...interface{}) {
		if s.ssclient_session != nil {
			return
		}

		if s.ssclient_session_mgr.IsInConnectCache(s.session_id) == false && s.ssclient_session_mgr.IsExistSessionOfSessID(s.session_id) == false {
			s.session_id = s.ssclient_session_mgr.SSClientConnect(s.addr, GSDServerSession, nil)
			elog.InfoAf("[ServiceDiscoveryClient] ReConnect SessionID=%v Addr=%v", s.session_id, s.addr)
		}
	}, []interface{}{}, true)

	return true
}

func (s *ServiceDiscoveryClient) SetSSClientSession(sess *SSClientSession) {
	s.ssclient_session = sess
}

func (s *ServiceDiscoveryClient) GetInitFlag() bool {
	return s.init_flag
}

func (s *ServiceDiscoveryClient) SetInitFlag() {
	s.init_flag = true
}

func (s *ServiceDiscoveryClient) GetCbFunc() SDCbFunc {
	return s.cb_func
}

//---------------------------------------------------------------------------------

type ServiceDiscoveryFunc func(datas []byte, sess *SSClientSession)

type SDServerSession struct {
	dealer *IDDealer
}

func (c *SDServerSession) Init() bool {
	c.dealer.RegisterHandler(uint32(pb.S2SBaseMsgId_service_discovery_ack_id), ServiceDiscoveryFunc(OnHandlerServiceDiscoveryAck))
	return true
}

func (c *SDServerSession) OnHandler(msgID uint32, datas []byte, sess *SSClientSession) {
	dealer := c.dealer.FindHandler(msgID)
	if dealer == nil {
		elog.ErrorAf("SDServerSession Can Not Find MsgID = %v", msgID)
		return
	}

	dealer.(ServiceDiscoveryFunc)(datas, sess)
}

func (c *SDServerSession) OnConnect(sess *SSClientSession) {
	GServiceDiscoveryClient.SetSSClientSession(sess)
}

func (c *SDServerSession) OnDisconnect(sess *SSClientSession) {
	GServiceDiscoveryClient.SetSSClientSession(nil)
}

func (c *SDServerSession) OnBeatHeartError(sess *SSClientSession) {

}

func OnHandlerServiceDiscoveryAck(datas []byte, sess *SSClientSession) {
	ack := &pb.ServiceDiscoveryAck{}
	err := proto.Unmarshal(datas, ack)
	if err != nil {
		elog.ErrorAf("[ServiceDiscovery] UpdAck Error=%v", err)
		return
	}

	elog.DebugAf("[ServiceDiscovery] UpdAck %+v", ack.SdInfo)

	if ack.RebuildFlag == true {
		elog.InfoA("[ServiceDiscovery] Service List Rebuilding")
		return
	}

	if ack.VerifyFlag == false {
		elog.InfoA("[ServiceDiscovery] Token Verify Error")
		return
	}

	//Listen
	if GServiceDiscoveryClient.GetInitFlag() == false {
		GServiceDiscoveryClient.SetInitFlag()
		if ack.SdInfo.S2SInterListen != "" && ack.SdInfo.S2SOuterListen != "" {
			if enet.GNet.Listen(ack.SdInfo.S2SInterListen, GSSServerSessionMgr, math.MaxUint16) == false {
				elog.Errorf("[ServiceDiscovery] Listen %v", ack.SdInfo.S2SInterListen)
				GServer.Quit()
				return
			}
		}
		GServerCfg.C2SInterListen = ack.SdInfo.C2SInterListen
		GServerCfg.C2SOuterListen = ack.SdInfo.C2SOuterListen
		GServerCfg.C2SListenMaxCount = int(ack.SdInfo.C2SMaxCount)
		GServerCfg.C2SHttpsUrl = ack.SdInfo.C2SHttpsUrl
		GServerCfg.C2SHttpsCert = ack.SdInfo.C2SHttpsCert
		GServerCfg.C2SHttpsKey = ack.SdInfo.C2SHttpsKey

		if GServiceDiscoveryClient.GetCbFunc() != nil {
			GServiceDiscoveryClient.GetCbFunc()()
		}
	}

	//Add Connect
	for _, newAttr := range ack.SdInfo.ConnList {
		existFlag := false
		if GSSServerSessionMgr.FindSessionByServerId(newAttr.ServerId) != nil || GSSServerSessionMgr.IsInConnectCache(newAttr.ServerId) == true {
			//连接成功 或者 连接中
			existFlag = true
		}

		if existFlag == false {
			elog.InfoAf("[ServiceDiscovery] Add Conn=%+v", newAttr)
			GSSServerSessionMgr.SSServerConnect(newAttr.ServerId, newAttr.ServerType, newAttr.ServerTypeStr, newAttr.Outer, newAttr.Token)
		}
	}

	//Del Connect Logic
	//如果ServerSession之间连接断开,都会在GServerSessionMgr删除,找不到，无需服务发现删除
}

//--------------------------------------------------
var GSDServerSession *SDServerSession
var GServiceDiscoveryClient *ServiceDiscoveryClient

func init() {
	GServiceDiscoveryClient = &ServiceDiscoveryClient{
		init_flag:     false,
		time_register: etimer.NewTimerRegister(),
	}

	GSDServerSession = &SDServerSession{
		dealer: NewIDDealer(),
	}

	GSDServerSession.Init()
}
