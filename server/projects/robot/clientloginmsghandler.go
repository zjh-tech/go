package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net/http"
	"projects/frame"
	"projects/go-engine/elog"
	"projects/pb"
	"projects/util"
	"time"

	"github.com/golang/protobuf/proto"
)

func SendLoginHttpsReq(url string, msgID uint32, datas []byte) {
	buff := bytes.NewBuffer([]byte{})
	binary.Write(buff, binary.BigEndian, msgID)
	binary.Write(buff, binary.BigEndian, datas)

	client := &http.Client{}
	//if frame.LOGIN_HTTPS_FLAG == true {
	//	client.Transport = GLoginSys.transport
	//}
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
		elog.InfoAf("[Login] AckMsgId=%v", ackMsgID)
		GLoginSysHandler.OnHandler(ackMsgID, body[ackMsgIDLen:])
	} else {
		elog.InfoAf("[ServiceDiscovery] Res Error=%v", len(body))
	}
}

type LoginSys struct {
	//httpclient证书只能初始化一次,httpsserver会出现too many open files
	// ==>lsof -p 进程id | wc -l 统计open files
	transport *http.Transport
	url       string
}

func (l *LoginSys) Init(caCert string, url string) bool {
	login_https_flag := false
	if login_https_flag == true {
		l.url += "https://" + url
		certPool := x509.NewCertPool()
		cert, ioErr := ioutil.ReadFile(caCert)
		if ioErr != nil {
			elog.ErrorAf("Https CaCert %v ioErr=%v", caCert, ioErr)
			return false
		}
		certPool.AppendCertsFromPEM(cert)

		l.transport = &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: certPool},
		}
		elog.InfoA("[LoginSys] Https Client Init Ok")
	} else {
		l.url += "http://" + url
		elog.InfoA("[LoginSys] Http Client Init Ok")
	}

	return true
}

func (l *LoginSys) SendCsAccountRegisterReq(accountName string, password string) {
	req := &pb.CsAccountRegisterReq{}
	req.AccountName = accountName
	req.Password = password
	datas, _ := proto.Marshal(req)
	elog.InfoAf("[LoginSys] SendCsAccountRegisterReq AccountName=%v", accountName)
	SendLoginHttpsReq(l.url, uint32(pb.EClient2GameMsgId_cs_account_register_req_id), datas)
}

func (l *LoginSys) SendCsAccountLoginReq(accountName string, password string) {
	elog.InfoAf("[LoginSys] SendCsAccountLoginReq AccountName=%v", accountName)
	req := &pb.CsAccountLoginReq{}
	req.AccountName = accountName
	req.Password = password
	datas, _ := proto.Marshal(req)
	SendLoginHttpsReq(l.url, uint32(pb.EClient2GameMsgId_cs_account_login_req_id), datas)
}

var GLoginSys *LoginSys

//---------------------------------------------------
type LoginSysFunc func(datas []byte)

type LoginSysHandler struct {
	dealer *frame.IDDealer
}

func (l *LoginSysHandler) Init() bool {
	l.dealer.RegisterHandler(uint32(pb.EClient2GameMsgId_sc_account_register_ack_id), LoginSysFunc(OnHandlerScAccountRegisterAck))
	l.dealer.RegisterHandler(uint32(pb.EClient2GameMsgId_sc_account_login_ack_id), LoginSysFunc(OnHandlerScAccountLoginAck))
	return true
}

func (l *LoginSysHandler) OnHandler(msgID uint32, datas []byte) {
	dealer := l.dealer.FindHandler(msgID)
	if dealer == nil {
		elog.ErrorAf("LoginSysHandler Can Not Find MsgID = %v", msgID)
		return
	}

	dealer.(LoginSysFunc)(datas)
}

var GLoginSysHandler *LoginSysHandler
var GScAccountLoginAck *pb.ScAccountLoginAck

func init() {
	GLoginSys = &LoginSys{}

	GLoginSysHandler = &LoginSysHandler{
		dealer: frame.NewIDDealer(),
	}
	GLoginSysHandler.Init()
}

func OnHandlerScAccountRegisterAck(datas []byte) {
	ack := &pb.ScAccountRegisterAck{}
	err := proto.Unmarshal(datas, ack)
	if err != nil {
		elog.ErrorAf("[LoginSysHandler] ScAccountRegisterAck Error=%v", err)
		return
	}

	if ack.ErrorCode == uint32(pb.EScErrorCode_com_success) {
		go GLoginSys.SendCsAccountLoginReq(ack.AccountName, "123456")
	}
}

func OnHandlerScAccountLoginAck(datas []byte) {
	ack := &pb.ScAccountLoginAck{}
	err := proto.Unmarshal(datas, ack)
	if err != nil {
		elog.ErrorAf("[LoginSysHandler] ScAccountLoginAck Error=%v", err)
		return
	}

	if ack.ErrorCode == uint32(pb.EScErrorCode_com_success) {
		GScAccountLoginAck = ack
		elog.InfoAf("GatewayIp=%v,GatewayPort=%v", ack.GatewayIp, ack.GatewayPort)
		gatewayAddr := fmt.Sprintf("%s:%d", ack.GatewayIp, ack.GatewayPort)
		frame.GCSClientSessionMgr.Connect(gatewayAddr, GClientGateMsgHandler, nil)
	}
}
