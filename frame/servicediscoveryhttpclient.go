package frame

import (
	"math"

	"github.com/zjh-tech/go-frame/engine/enet"
	"github.com/zjh-tech/go-frame/engine/etimer"
	"github.com/zjh-tech/go-frame/frame/framepb"

	"github.com/golang/protobuf/proto"
)

const (
	SERVICEDISCOVERY_HTTP_TIMER_ID uint32 = 1
)

const (
	SERVICEDISCOVERY_HTTP_TIMER_DELAY uint64 = 1000 * 3
)

type SDHttpCbFunc func(...interface{})

type ServiceDiscoveryHttpClient struct {
	ServiceInitFlag bool
	time_register   etimer.ITimerRegister
	cb              SDHttpCbFunc
}

func (s *ServiceDiscoveryHttpClient) Init(url string, server_id uint64, token string, cb SDHttpCbFunc) bool {
	s.cb = cb
	var full_url string
	full_url += "http://" + url
	ELog.InfoA("[SDClient] Http Client Init Ok")

	s.time_register.AddRepeatTimer(SERVICEDISCOVERY_HTTP_TIMER_ID, SERVICEDISCOVERY_HTTP_TIMER_DELAY, "SDClient-SendSDHttpReq", func(v ...interface{}) {
		temp_url := v[0].(string)
		temp_server_id := v[1].(uint64)
		temp_token := v[2].(string)
		req := &framepb.ServiceDiscoveryReq{
			ServerId: temp_server_id,
			Token:    temp_token,
		}
		SendSDHttpReq(temp_url, uint32(framepb.S2SBaseMsgId_service_discovery_req_id), req)
	}, []interface{}{full_url, server_id, token}, true)

	return true
}

//-----------------------------------------------------------
type ServiceDiscoveryHttpFunc func(datas []byte)

type SDHttpServerSession struct {
	dealer *IDDealer
}

func (c *SDHttpServerSession) Init() bool {
	c.dealer.RegisterHandler(uint32(framepb.S2SBaseMsgId_service_discovery_ack_id), ServiceDiscoveryHttpFunc(OnHandlerHttpServiceDiscoveryAck))
	return true
}

func (c *SDHttpServerSession) OnHandler(msgId uint32, datas []byte) {
	dealer := c.dealer.FindHandler(msgId)
	if dealer == nil {
		ELog.ErrorAf("SDHttpServerSession Can Not Find MsgID = %v", msgId)
		return
	}

	dealer.(ServiceDiscoveryHttpFunc)(datas)
}

func OnHandlerHttpServiceDiscoveryAck(datas []byte) {
	ack := &framepb.ServiceDiscoveryAck{}
	err := proto.Unmarshal(datas, ack)
	if err != nil {
		ELog.ErrorAf("[ServiceDiscovery] Http UpdAck Error=%v", err)
		return
	}

	ELog.DebugAf("[ServiceDiscovery] Http UpdAck %+v", ack.SdInfo)

	if ack.RebuildFlag == true {
		ELog.InfoA("[ServiceDiscovery] Http Service List Rebuilding")
		return
	}

	if ack.VerifyFlag == false {
		ELog.InfoA("[ServiceDiscovery] Http Token Verify Error")
		return
	}

	//Listen
	if GServiceDiscoveryHttpClient.ServiceInitFlag == false {
		GServiceDiscoveryHttpClient.ServiceInitFlag = true
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

		GServerCfg.C2SInterListen = ack.SdInfo.C2SInterListen
		GServerCfg.C2SOuterListen = ack.SdInfo.C2SOuterListen
		GServerCfg.S2SHttpServerUrl = ack.SdInfo.S2SHttpSurl
		GServerCfg.S2SHttpClientUrl1 = ack.SdInfo.S2SHttpCurl1
		GServerCfg.S2SHttpClientUrl2 = ack.SdInfo.S2SHttpCurl2

		GServerCfg.C2SHttpsUrl = ack.SdInfo.C2SHttpsUrl
		GServerCfg.C2SHttpsCert = ack.SdInfo.C2SHttpsCert
		GServerCfg.C2SHttpsKey = ack.SdInfo.C2SHttpsKey

		GServerCfg.SDK_TCP_INTER = ack.SdInfo.SdkTcpInter
		GServerCfg.SDK_TCP_OUTER = ack.SdInfo.SdkTcpOut
		GServerCfg.SDKHttpsUrl = ack.SdInfo.SdkHttpsUrtl
		GServerCfg.SDKHttpsCert = ack.SdInfo.SdkHttpsCert
		GServerCfg.SDKHttpsKey = ack.SdInfo.SdkHttpsKey
		ELog.InfoAf("[ServiceDiscovery] ServerCfg=%+v", GServerCfg)

		if GServiceDiscoveryHttpClient.cb != nil {
			GServiceDiscoveryHttpClient.cb()
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
			ELog.InfoAf("[ServiceDiscovery] Http Add Conn=%+v", newAttr)
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

var GServiceDiscoveryHttpClient *ServiceDiscoveryHttpClient

func init() {
	GServiceDiscoveryHttpClient = &ServiceDiscoveryHttpClient{
		ServiceInitFlag: false,
		time_register:   etimer.NewTimerRegister(),
	}

}
