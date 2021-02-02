package frame

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"math"
	"net/http"
	"projects/go-engine/ehttp"
	"projects/go-engine/elog"
	"projects/go-engine/enet"
	"projects/go-engine/etimer"
	"projects/pb"
	"projects/util"
	"time"

	"github.com/golang/protobuf/proto"
)

const (
	SERVICEDISCOVERY_HTTP_TIMER_ID uint32 = 1
)

const (
	SERVICEDISCOVERY_HTTP_TIMER_DELAY uint64 = 1000 * 3
)

type SDHttpCbFunc func(...interface{})

func SendServiceDiscoveryHttpReq(url string, msgID uint32, datas []byte) {
	buff := bytes.NewBuffer([]byte{})
	binary.Write(buff, binary.BigEndian, msgID)
	binary.Write(buff, binary.BigEndian, datas)

	client := &http.Client{}
	client.Timeout = time.Second

	resp, resErr := client.Post(url, "application/octet-stream", buff)
	if resErr != nil {
		elog.ErrorAf("Http Post Url=%v resErr=%v", url, resErr)
		return
	}

	defer resp.Body.Close()
	body, bodyErr := ioutil.ReadAll(resp.Body)
	if bodyErr != nil {
		elog.ErrorAf("Http Url=%v bodyErr=%v", url, bodyErr)
		return
	}

	ackMsgIDLen := 4
	if len(body) > ackMsgIDLen {
		ackMsgID := util.NetBytesToUint32(body)
		httpEvent := ehttp.NewHttpEvent(GSDHttpServerSession, ackMsgID, body[ackMsgIDLen:])
		ehttp.GHttpNet.PushHttpEvent(httpEvent)
	} else {
		elog.InfoAf("[ServiceDiscovery] Res Error=%v", len(body))
	}
}

type ServiceDiscoveryHttpClient struct {
	ServiceInitFlag bool
	timeRegister    etimer.ITimerRegister
	cb              SDHttpCbFunc
}

func (s *ServiceDiscoveryHttpClient) Init(_url string, _serverID uint64, _token string, cb SDHttpCbFunc) bool {
	s.cb = cb
	var full_url string
	full_url += "http://" + _url
	elog.InfoA("[SDClient] Http Client Init Ok")

	s.timeRegister.AddRepeatTimer(SERVICEDISCOVERY_HTTP_TIMER_ID, SERVICEDISCOVERY_HTTP_TIMER_DELAY, "SDClient-SendSDHttpReq", func(v ...interface{}) {
		url := v[0].(string)
		serverID := v[1].(uint64)
		token := v[2].(string)
		req := &pb.ServiceDiscoveryReq{
			ServerId: serverID,
			Token:    token,
		}
		datas, _ := proto.Marshal(req)
		go SendServiceDiscoveryHttpReq(url, uint32(pb.S2SBaseMsgId_service_discovery_req_id), datas)
	}, []interface{}{full_url, _serverID, _token}, true)

	return true
}

//-----------------------------------------------------------
type ServiceDiscoveryHttpFunc func(datas []byte)

type SDHttpServerSession struct {
	dealer *IDDealer
}

func (c *SDHttpServerSession) Init() bool {
	c.dealer.RegisterHandler(uint32(pb.S2SBaseMsgId_service_discovery_ack_id), ServiceDiscoveryHttpFunc(OnHandlerHttpServiceDiscoveryAck))
	return true
}

func (c *SDHttpServerSession) OnHandler(msgID uint32, datas []byte) {
	dealer := c.dealer.FindHandler(msgID)
	if dealer == nil {
		elog.ErrorAf("SDHttpServerSession Can Not Find MsgID = %v", msgID)
		return
	}

	dealer.(ServiceDiscoveryHttpFunc)(datas)
}

func OnHandlerHttpServiceDiscoveryAck(datas []byte) {
	ack := &pb.ServiceDiscoveryAck{}
	err := proto.Unmarshal(datas, ack)
	if err != nil {
		elog.ErrorAf("[ServiceDiscovery] Http UpdAck Error=%v", err)
		return
	}

	elog.DebugAf("[ServiceDiscovery] Http UpdAck %+v", ack.SdInfo)

	if ack.RebuildFlag == true {
		elog.InfoA("[ServiceDiscovery] Http Service List Rebuilding")
		return
	}

	if ack.VerifyFlag == false {
		elog.InfoA("[ServiceDiscovery] Http Token Verify Error")
		return
	}

	//Listen
	if GServiceDiscoveryHttpClient.ServiceInitFlag == false {
		GServiceDiscoveryHttpClient.ServiceInitFlag = true
		if ack.SdInfo.S2SInterListen != "" && ack.SdInfo.S2SOuterListen != "" {
			if enet.GNet.Listen(ack.SdInfo.S2SInterListen, GSSServerSessionMgr, math.MaxUint16) == false {
				elog.Errorf("[ServiceDiscovery] Http Listen %v", ack.SdInfo.S2SInterListen)
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

		if GServiceDiscoveryHttpClient.cb != nil {
			GServiceDiscoveryHttpClient.cb()
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
			elog.InfoAf("[ServiceDiscovery] Http Add Conn=%+v", newAttr)
			GSSServerSessionMgr.SSServerConnect(newAttr.ServerId, newAttr.ServerType, newAttr.ServerTypeStr, newAttr.Outer, newAttr.Token)
		}
	}

	//Del Connect Logic
	//如果ServerSession之间连接断开,都会在GServerSessionMgr删除,找不到，无需服务发现删除
}

//--------------------------------------------------
var GSDHttpServerSession *SDHttpServerSession
var GServiceDiscoveryHttpClient *ServiceDiscoveryHttpClient

func init() {
	GServiceDiscoveryHttpClient = &ServiceDiscoveryHttpClient{
		ServiceInitFlag: false,
		timeRegister:    etimer.NewTimerRegister(),
	}

	GSDHttpServerSession = &SDHttpServerSession{
		dealer: NewIDDealer(),
	}

	GSDHttpServerSession.Init()
}
