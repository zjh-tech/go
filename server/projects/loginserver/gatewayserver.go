package main

import (
	"projects/frame"
	"projects/go-engine/elog"
	"projects/go-engine/eredis"
	"projects/pb"
	"projects/util"

	"github.com/golang/protobuf/proto"
)

type GatewayFunc func(datas []byte, g *GatewayServer) bool

type GatewayServer struct {
	frame.LogicServer
	dealer           *frame.IDDealer
	Ip               string
	Port             uint32
	PlayerTotalCount uint64
}

func NewGatewayServer() *GatewayServer {
	gateway := &GatewayServer{
		dealer: frame.NewIDDealer(),
	}
	gateway.Init()
	return gateway
}

func (g *GatewayServer) Init() bool {
	g.dealer.RegisterHandler(uint32(pb.S2SLogicMsgId_gl_gateway_info_ntf_id), GatewayFunc(OnHandlerGlGatewayInfoNtf))
	g.dealer.RegisterHandler(uint32(pb.S2SLogicMsgId_s2s_hgl_kick_player_ack_id), GatewayFunc(OnHandlerS2SHglKickPlayerAck))
	g.dealer.RegisterHandler(uint32(pb.S2SLogicMsgId_gl_verify_token_req_id), GatewayFunc(OnHandlerGlVerifyTokenReq))

	return true
}

func (g *GatewayServer) OnHandler(msgID uint32, attach_datas []byte, datas []byte, sess *frame.SSServerSession) {
	defer func() {
		if err := recover(); err != nil {
			elog.ErrorAf("GatewayServer onHandler MsgID = %v Error=%v", msgID, err)
		}
	}()

	dealer := g.dealer.FindHandler(msgID)
	if dealer == nil {
		elog.ErrorAf("GatewayServer MsgHandler Can Not Find MsgID = %v", msgID)
		return
	}

	dealer.(GatewayFunc)(datas, g)
}

func (g *GatewayServer) OnEstablish(serversess *frame.SSServerSession) {
	elog.InfoAf("GatewayServer OnEstablish Remote [ID=%v,Type=%v,Ip=%v] ", serversess.GetRemoteServerID(), serversess.GetRemoteServerType(), serversess.GetRemoteOuter())
}

func (g *GatewayServer) OnTerminate(serversess *frame.SSServerSession) {
	elog.InfoAf("GatewayServer OnTerminate Remote [ID=%v,Type=%v,Ip=%v] ", serversess.GetRemoteServerID(), serversess.GetRemoteServerType(), serversess.GetRemoteOuter())
}

func OnHandlerGlGatewayInfoNtf(datas []byte, g *GatewayServer) bool {
	ntf := pb.GlGatewayInfoNtf{}
	err := proto.Unmarshal(datas, &ntf)
	if err != nil {
		return false
	}

	g.Ip = ntf.Ip
	g.Port = ntf.Port
	g.PlayerTotalCount = ntf.PlayerTotalCount
	elog.InfoAf("[GatewayServer] Ntf Ip=%v,Port=%v PlayerTotalCount=%v", ntf.Ip, ntf.Port, ntf.PlayerTotalCount)
	return true
}

func OnHandlerGlVerifyTokenReq(datas []byte, g *GatewayServer) bool {
	glLoginReq := pb.GlVerifyTokenReq{}
	err := proto.Unmarshal(datas, &glLoginReq)
	if err != nil {
		return false
	}

	verifyFlag := GTokenMgr.IsValidToken(glLoginReq.Accountid, glLoginReq.Token)
	if verifyFlag == false {
		SendLgVerifyTokenAck(glLoginReq.Accountid, glLoginReq.Token, glLoginReq.Playeruid, frame.MSG_FAIL, g.GetServerSession())
		return true
	}

	redisClient := eredis.GRedisModule.GetRedisClient(glLoginReq.Accountid)
	if redisClient == nil {
		elog.ErrorAf("[Login] AccountId=%v RedisClient Is Nil", glLoginReq.Accountid)
		return false
	}

	gatewaySrvIdKey := frame.GetRedisKey("Login", glLoginReq.Accountid, "GatewaySrvId")
	gatewaySrvIdValue, redisErr := redisClient.Get(gatewaySrvIdKey)
	if redisClient != nil {
		elog.InfoAf("[Login] AccountId=%v RedisError=%v", redisErr)
		return false
	}

	gatewaySrvID, _ := util.Str2Uint64(string(gatewaySrvIdValue))
	if gatewaySrvID != 0 {
		//踢掉旧的Gateway上的Player
		kickReq := pb.S2SLghKickPlayerReq{
			Accountid:       glLoginReq.Accountid,
			Token:           glLoginReq.Token,
			Playeruid:       glLoginReq.Playeruid,
			Newgatewaysrvid: g.GetServerSession().GetRemoteServerID(),
		}
		if frame.GSSServerSessionMgr.SendProtoMsg(gatewaySrvID, uint32(pb.S2SLogicMsgId_s2s_lgh_kick_player_req_id), &kickReq) == false {
			elog.ErrorAf("[Login] AcccountId=%v Kick Old Player OldGateway", glLoginReq.Accountid)
			return false
		}
	} else {
		//没登录过
		SendLgVerifyTokenAck(glLoginReq.Accountid, glLoginReq.Token, glLoginReq.Playeruid, frame.MSG_SUCCESS, g.GetServerSession())
	}

	return true
}

func OnHandlerS2SHglKickPlayerAck(datas []byte, g *GatewayServer) bool {
	loginAck := pb.S2SHglKickPlayerAck{}
	err := proto.Unmarshal(datas, &loginAck)
	if err != nil {
		return false
	}

	sess := frame.GSSServerSessionMgr.FindSessionByServerId(loginAck.Newgatewaysrvid)
	if sess == nil {
		return false
	}
	serverSess := sess.(*frame.SSServerSession)
	SendLgVerifyTokenAck(loginAck.Accountid, loginAck.Token, loginAck.Playeruid, frame.MSG_SUCCESS, serverSess)
	return true
}

func SendLgVerifyTokenAck(accountid uint64, token []byte, playeruid uint64, errorCode uint32, gateway *frame.SSServerSession) {
	lgLoginAck := pb.LgVerifyTokenAck{}
	lgLoginAck.Accountid = accountid
	lgLoginAck.Token = token
	lgLoginAck.Playeruid = playeruid
	lgLoginAck.Errorcode = errorCode
	gateway.SendProtoMsg(uint32(pb.S2SLogicMsgId_lg_verify_token_ack_id), &lgLoginAck, nil)
}
