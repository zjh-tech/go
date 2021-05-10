package frame

import (
	"math"

	"github.com/zjh-tech/go-frame/engine/enet"
	"github.com/zjh-tech/go-frame/engine/etimer"
	"github.com/zjh-tech/go-frame/frame/framepb"

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
	ssclient_session_mgr *SDKSessionMgr
	ssclient_session     *SDKSession
	session_id           uint64
}

func (s *ServiceDiscoveryClient) Init(addr string, server_id uint64, token string, cb_func SDCbFunc) bool {
	s.addr = addr
	s.server_id = server_id
	s.token = token
	s.init_flag = false
	s.cb_func = cb_func
	s.ssclient_session_mgr = GSDKSessionMgr
	s.session_id = s.ssclient_session_mgr.SSClientConnect(addr, GSDServerSession, nil)
	ELog.InfoAf("[ServiceDiscoveryClient] Connect SessionID=%v Addr=%v", s.session_id, addr)

	s.time_register.AddRepeatTimer(SD_CLIENT_SEND_REQ_TIMER_ID, SD_CLIENT_SEND_REQ_TIMER_DELAY, "SDClient-SendSDReq", func(v ...interface{}) {
		if s.ssclient_session == nil {
			return
		}

		serverID := v[0].(uint64)
		token := v[1].(string)
		req := &framepb.ServiceDiscoveryReq{
			ServerId: serverID,
			Token:    token,
		}
		s.ssclient_session.AsyncSendProtoMsg(uint32(framepb.S2SBaseMsgId_service_discovery_req_id), req)
		ELog.DebugAf("[ServiceDiscoveryClient] Send ServiceDiscoveryReq ServerID=%v", serverID)
	}, []interface{}{server_id, token}, true)

	s.time_register.AddRepeatTimer(SD_CLIENT_RECONNECT_TIMER_ID, SD_CLIENT_RECONNECT_TIMER_DELAY, "SDClient-CheckReconnect", func(v ...interface{}) {
		if s.ssclient_session != nil {
			return
		}

		if s.ssclient_session_mgr.IsInConnectCache(s.session_id) == false && s.ssclient_session_mgr.IsExistSessionOfSessID(s.session_id) == false {
			s.session_id = s.ssclient_session_mgr.SSClientConnect(s.addr, GSDServerSession, nil)
			ELog.InfoAf("[ServiceDiscoveryClient] ReConnect SessionID=%v Addr=%v", s.session_id, s.addr)
		}
	}, []interface{}{}, true)

	return true
}

func (s *ServiceDiscoveryClient) SetSDKSession(sess *SDKSession) {
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

type ServiceDiscoveryFunc func(datas []byte, sess *SDKSession)

type SDServerSession struct {
	dealer *IDDealer
}

func (c *SDServerSession) Init() bool {
	c.dealer.RegisterHandler(uint32(framepb.S2SBaseMsgId_service_discovery_ack_id), ServiceDiscoveryFunc(OnHandlerServiceDiscoveryAck))
	return true
}

func (c *SDServerSession) OnHandler(msgID uint32, datas []byte, sess *SDKSession) {
	dealer := c.dealer.FindHandler(msgID)
	if dealer == nil {
		ELog.ErrorAf("SDServerSession Can Not Find MsgID = %v", msgID)
		return
	}

	dealer.(ServiceDiscoveryFunc)(datas, sess)
}

func (c *SDServerSession) OnConnect(sess *SDKSession) {
	GServiceDiscoveryClient.SetSDKSession(sess)
}

func (c *SDServerSession) OnDisconnect(sess *SDKSession) {
	GServiceDiscoveryClient.SetSDKSession(nil)
}

func (c *SDServerSession) OnBeatHeartError(sess *SDKSession) {

}

func OnHandlerServiceDiscoveryAck(datas []byte, sess *SDKSession) {
	ack := &framepb.ServiceDiscoveryAck{}
	err := proto.Unmarshal(datas, ack)
	if err != nil {
		ELog.ErrorAf("[ServiceDiscovery] UpdAck Error=%v", err)
		return
	}

	ELog.DebugAf("[ServiceDiscovery] UpdAck %+v", ack.SdInfo)

	if ack.RebuildFlag == true {
		ELog.InfoA("[ServiceDiscovery] Service List Rebuilding")
		return
	}

	if ack.VerifyFlag == false {
		ELog.InfoA("[ServiceDiscovery] Token Verify Error")
		return
	}

	//Listen
	if GServiceDiscoveryClient.GetInitFlag() == false {
		GServiceDiscoveryClient.SetInitFlag()

		if ack.SdInfo.S2SInterListen != "" {
			if enet.GNet.Listen(ack.SdInfo.S2SInterListen, GSSSessionMgr, math.MaxUint16) == false {
				ELog.Errorf("[ServiceDiscovery] Http Listen %v", ack.SdInfo.S2SInterListen)
				GServer.Quit()
				return
			}
		} else if ack.SdInfo.S2SOuterListen != "" {
			if enet.GNet.Listen(ack.SdInfo.S2SOuterListen, GSSSessionMgr, math.MaxUint16) == false {
				ELog.Errorf("[ServiceDiscovery] Http Listen %v", ack.SdInfo.S2SOuterListen)
				GServer.Quit()
				return
			}
		}

		GServerCfg.S2SHttpServerUrl = ack.SdInfo.S2SHttpSurl
		GServerCfg.S2SHttpClientUrl1 = ack.SdInfo.S2SHttpCurl1
		GServerCfg.S2SHttpClientUrl2 = ack.SdInfo.S2SHttpCurl2

		GServerCfg.C2SInterListen = ack.SdInfo.C2SInterListen
		GServerCfg.C2SOuterListen = ack.SdInfo.C2SOuterListen
		GServerCfg.C2SHttpsUrl = ack.SdInfo.C2SHttpsUrl
		GServerCfg.C2SHttpsCert = ack.SdInfo.C2SHttpsCert
		GServerCfg.C2SHttpsKey = ack.SdInfo.C2SHttpsKey

		GServerCfg.SDK_TCP_INTER = ack.SdInfo.SdkTcpInter
		GServerCfg.SDK_TCP_OUTER = ack.SdInfo.SdkTcpOut
		GServerCfg.SDKHttpsUrl = ack.SdInfo.SdkHttpsUrtl
		GServerCfg.SDKHttpsCert = ack.SdInfo.SdkHttpsCert
		GServerCfg.SDKHttpsKey = ack.SdInfo.SdkHttpsKey

		ELog.InfoAf("[ServiceDiscovery] ServerCfg=%+v", GServerCfg)

		if GServiceDiscoveryClient.GetCbFunc() != nil {
			GServiceDiscoveryClient.GetCbFunc()()
		}
	}

	//Add Connect
	for _, newAttr := range ack.SdInfo.ConnList {
		existFlag := false
		if GSSSessionMgr.FindSessionByServerId(newAttr.ServerId) != nil || GSSSessionMgr.IsInConnectCache(newAttr.ServerId) == true {
			//连接成功 或者 连接中
			existFlag = true
		}

		if existFlag == false {
			ELog.InfoAf("[ServiceDiscovery] Add Conn=%+v", newAttr)
			if newAttr.Inter != "" {
				GSSSessionMgr.SSServerConnect(newAttr.ServerId, newAttr.ServerType, newAttr.ServerTypeStr, newAttr.Inter, newAttr.Token)
			} else if newAttr.Outer != "" {
				GSSSessionMgr.SSServerConnect(newAttr.ServerId, newAttr.ServerType, newAttr.ServerTypeStr, newAttr.Outer, newAttr.Token)
			}
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
