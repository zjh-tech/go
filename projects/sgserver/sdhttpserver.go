package main

import (
	"io/ioutil"
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/zjh-tech/go-frame/base/util"
	"github.com/zjh-tech/go-frame/frame"
	"github.com/zjh-tech/go-frame/frame/framepb"
)

type SDHttpServerFunc func(datas []byte, w http.ResponseWriter)

type SDHttpServer struct {
	dealer *frame.IDDealer
}

func NewSDHttpServer() *SDHttpServer {
	sdhttpserver := &SDHttpServer{
		dealer: frame.NewIDDealer(),
	}
	sdhttpserver.Init()
	return sdhttpserver
}

func (s *SDHttpServer) Init() bool {
	s.dealer.RegisterHandler(uint32(framepb.S2SBaseMsgId_service_discovery_req_id), SDHttpServerFunc(OnHandlerServiceDiscoveryHttpReq))
	return true
}

func (s *SDHttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ELog.ErrorAf("SDHttpHandler Error=%v", err)
		return
	}

	if len(body) < 4 {
		ELog.ErrorAf("SDHttpHandler Body MsgID Len Error")
		return
	}

	msgID := util.NetBytesToUint32(body)
	msgIDLen := 4
	dealer := s.dealer.FindHandler(msgID)
	if dealer == nil {
		ELog.ErrorAf("RegistryServer HttpsMsgHandler Can Not Find MsgID = %v", msgID)
		return
	}

	dealer.(SDHttpServerFunc)(body[msgIDLen:], w)
}

func FillServiceDiscoveryConn(serviceDiscovery *framepb.ServiceDiscovery) {
	Attr, ok := GRegistryCfg.AttrMap[serviceDiscovery.ServerId]
	if !ok {
		return
	}

	endDatas, connOk := GRegistryCfg.ConntionMap[Attr.ServerType]
	if !connOk {
		return
	}

	for _, serviceInfo := range GServiceDiscoveryServer.UseServices {
		//Exclude Warn Service
		if serviceInfo.WarnFlag == true {
			continue
		}

		usedServerAttr, usedOk := GRegistryCfg.AttrMap[serviceInfo.ServerId]
		if !usedOk {
			continue
		}

		//Is Connect Relation
		isConnFlag := false
		for _, endServerType := range endDatas {
			if endServerType == usedServerAttr.ServerType {
				isConnFlag = true
				break
			}
		}

		if isConnFlag {
			connServerAttr := framepb.SdConnAttr{
				ServerId:      usedServerAttr.ServerID,
				ServerType:    usedServerAttr.ServerType,
				ServerTypeStr: usedServerAttr.ServerTypeStr,
				Inter:         usedServerAttr.S2S_TCP_Inter,
				Outer:         usedServerAttr.S2S_TCP_Outer,
				Token:         usedServerAttr.Token,
			}
			serviceDiscovery.ConnList = append(serviceDiscovery.ConnList, &connServerAttr)
		}
	}
}

func OnHandlerServiceDiscoveryHttpReq(datas []byte, w http.ResponseWriter) {
	var AckFunc = func(ack *framepb.ServiceDiscoveryAck, w http.ResponseWriter) {
		w.Write(util.Uint32ToNetBytes(uint32(framepb.S2SBaseMsgId_service_discovery_ack_id)))
		res, _ := proto.Marshal(ack)
		w.Write(res)
	}

	ack := &framepb.ServiceDiscoveryAck{}
	ack.RebuildFlag = GServiceDiscoveryServer.RebuildFlag
	ack.VerifyFlag = false

	req := &framepb.ServiceDiscoveryReq{}
	err := proto.Unmarshal(datas, req)
	if err != nil {
		AckFunc(ack, w)
		ELog.ErrorAf("[ServiceDiscovery] Http UdpReq Protobuf Unmarshal=%v", req.ServerId)
		return
	}

	Attr, attOk := GRegistryCfg.AttrMap[req.ServerId]
	if !attOk {
		AckFunc(ack, w)
		ELog.ErrorAf("[ServiceDiscovery] Http RegistryCfg Not Find ServerId=%v", req.ServerId)
		return
	}

	if req.Token != Attr.Token {
		AckFunc(ack, w)
		ELog.ErrorAf("[ServiceDiscovery] Http  ServerId=%v Token Error", req.ServerId)
		return
	}

	ack.VerifyFlag = true

	GServiceDiscoveryServer.Mutex.Lock()
	defer GServiceDiscoveryServer.Mutex.Unlock()

	_, usedOk := GServiceDiscoveryServer.UseServices[req.ServerId]
	if !usedOk {
		GServiceDiscoveryServer.AddUsedService(req.ServerId)
	}

	GServiceDiscoveryServer.RemoveWarnService(req.ServerId)
	if GServiceDiscoveryServer.RebuildFlag == false {
		ack.SdInfo = &framepb.ServiceDiscovery{}
		ack.SdInfo.ServerId = req.ServerId
		ack.SdInfo.S2SInterListen = Attr.S2S_TCP_Inter
		ack.SdInfo.S2SOuterListen = Attr.S2S_TCP_Outer
		ack.SdInfo.S2SHttpSurl = Attr.S2S_Http_SUrl
		ack.SdInfo.S2SHttpCurl1 = Attr.S2S_Http_CUrl1
		ack.SdInfo.S2SHttpCurl2 = Attr.S2S_Http_CUrl2

		FillServiceDiscoveryConn(ack.SdInfo)

		ack.SdInfo.C2SInterListen = Attr.C2S_TCP_Inter
		ack.SdInfo.C2SOuterListen = Attr.C2S_TCP_Outer
		ack.SdInfo.C2SHttpsUrl = Attr.C2SHttpsUrl
		ack.SdInfo.C2SHttpsCert = Attr.C2SHttpsCert
		ack.SdInfo.C2SHttpsKey = Attr.C2SHttpsKey

		ack.SdInfo.SdkTcpInter = Attr.SDK_TCP_Inter
		ack.SdInfo.SdkTcpOut = Attr.SDK_TCP_Outer
		ack.SdInfo.SdkHttpsUrtl = Attr.SDKHttpsUrl
		ack.SdInfo.SdkHttpsCert = Attr.SDKHttpsCert
		ack.SdInfo.SdkHttpsKey = Attr.SDKHttpsKey

	}

	AckFunc(ack, w)
	ELog.InfoAf("[ServiceDiscovery] Http UdpServiceDiscoveryAck=%+v", ack)
}
