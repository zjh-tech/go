package main

import (
	"io/ioutil"
	"net/http"
	"projects/frame"
	"projects/go-engine/elog"
	"projects/pb"
	"projects/util"

	"github.com/golang/protobuf/proto"
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
	s.dealer.RegisterHandler(uint32(pb.S2SBaseMsgId_service_discovery_req_id), SDHttpServerFunc(OnHandlerServiceDiscoveryHttpReq))
	return true
}

func (s *SDHttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		elog.ErrorAf("SDHttpHandler Error=%v", err)
		return
	}

	if len(body) < 4 {
		elog.ErrorAf("SDHttpHandler Body MsgID Len Error")
		return
	}

	msgID := util.NetBytesToUint32(body)
	msgIDLen := 4
	dealer := s.dealer.FindHandler(msgID)
	if dealer == nil {
		elog.ErrorAf("RegistryServer HttpsMsgHandler Can Not Find MsgID = %v", msgID)
		return
	}

	dealer.(SDHttpServerFunc)(body[msgIDLen:], w)
}

func FillServiceDiscoveryConn(serviceDiscovery *pb.ServiceDiscovery) {
	s2sAttr, ok := GRegistryCfg.S2SAttrMap[serviceDiscovery.ServerId]
	if !ok {
		return
	}

	endDatas, connOk := GRegistryCfg.ConntionMap[s2sAttr.ServerType]
	if !connOk {
		return
	}

	for _, serviceInfo := range GServiceDiscoveryServer.UseServices {
		//Exclude Warn Service
		if serviceInfo.WarnFlag == true {
			continue
		}

		usedServerAttr, usedOk := GRegistryCfg.S2SAttrMap[serviceInfo.ServerId]
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
			connServerAttr := pb.SdConnAttr{
				ServerId:      usedServerAttr.ServerID,
				ServerType:    usedServerAttr.ServerType,
				ServerTypeStr: usedServerAttr.ServerTypeStr,
				Inter:         usedServerAttr.Inter,
				Outer:         usedServerAttr.Outer,
				Token:         usedServerAttr.Token,
			}
			serviceDiscovery.ConnList = append(serviceDiscovery.ConnList, &connServerAttr)
		}
	}
}

func OnHandlerServiceDiscoveryHttpReq(datas []byte, w http.ResponseWriter) {
	var AckFunc = func(ack *pb.ServiceDiscoveryAck, w http.ResponseWriter) {
		w.Write(util.Uint32ToNetBytes(uint32(pb.S2SBaseMsgId_service_discovery_ack_id)))
		res, _ := proto.Marshal(ack)
		w.Write(res)
	}

	ack := &pb.ServiceDiscoveryAck{}
	ack.RebuildFlag = GServiceDiscoveryServer.RebuildFlag
	ack.VerifyFlag = false

	req := &pb.ServiceDiscoveryReq{}
	err := proto.Unmarshal(datas, req)
	if err != nil {
		AckFunc(ack, w)
		elog.ErrorAf("[ServiceDiscovery] Http UdpReq Protobuf Unmarshal=%v", req.ServerId)
		return
	}

	s2sAttr, attOk := GRegistryCfg.S2SAttrMap[req.ServerId]
	if !attOk {
		AckFunc(ack, w)
		elog.ErrorAf("[ServiceDiscovery] Http RegistryCfg Not Find ServerId=%v", req.ServerId)
		return
	}

	if req.Token != s2sAttr.Token {
		AckFunc(ack, w)
		elog.ErrorAf("[ServiceDiscovery] Http  ServerId=%v Token Error", req.ServerId)
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
		ack.SdInfo = &pb.ServiceDiscovery{}
		ack.SdInfo.ServerId = req.ServerId
		ack.SdInfo.S2SInterListen = s2sAttr.Inter
		ack.SdInfo.S2SOuterListen = s2sAttr.Outer
		FillServiceDiscoveryConn(ack.SdInfo)

		c2sAttr, c2sAttOk := GRegistryCfg.C2SAttrMap[req.ServerId]
		if c2sAttOk {
			ack.SdInfo.C2SInterListen = c2sAttr.Inter
			ack.SdInfo.C2SOuterListen = c2sAttr.Outer
			ack.SdInfo.C2SMaxCount = uint32(c2sAttr.MaxCount)
			ack.SdInfo.C2SHttpsUrl = c2sAttr.C2SHttpsUrl
			ack.SdInfo.C2SHttpsCert = c2sAttr.C2SHttpsCert
			ack.SdInfo.C2SHttpsKey = c2sAttr.C2SHttpsKey
		}
	}

	AckFunc(ack, w)
	elog.InfoAf("[ServiceDiscovery] Http UdpServiceDiscoveryAck=%+v", ack)
}
