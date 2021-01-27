package main

import (
	"math"
	"projects/frame"
	"projects/go-engine/elog"
	"projects/pb"

	"github.com/golang/protobuf/proto"
)

type LoginServerFunc func(datas []byte, g *LoginServer) bool

type LoginServer struct {
	frame.LogicServer
	dealer *frame.IDDealer
}

func NewLoginServer() *LoginServer {
	login := &LoginServer{
		dealer: frame.NewIDDealer(),
	}
	login.Init()
	return login
}
func (g *LoginServer) Init() bool {
	g.dealer.RegisterHandler(uint32(pb.S2SLogicMsgId_s2s_lgh_kick_player_req_id), LoginServerFunc(OnHandlerS2SLghKickPlayerReq))
	g.dealer.RegisterHandler(uint32(pb.S2SLogicMsgId_lg_verify_token_ack_id), LoginServerFunc(OnHandlerLgVerifyTokenAck))

	return true
}

func (c *LoginServer) OnHandler(msgID uint32, attach_datas []byte, datas []byte, sess *frame.SSServerSession) {
	defer func() {
		if err := recover(); err != nil {
			elog.ErrorAf("LoginServer onHandler MsgID = %v Error=%v", msgID, err)
		}
	}()

	dealer := c.dealer.FindHandler(msgID)
	if dealer == nil {
		elog.ErrorAf("LoginServer MsgHandler Can Not Find MsgID = %v", msgID)
		return
	}

	dealer.(LoginServerFunc)(datas, c)
}

func (g *LoginServer) OnEstablish(serversess *frame.SSServerSession) {
	elog.InfoAf("LoginServer OnEstablish Remote [ID=%v,Type=%v,Ip=%v] ", serversess.GetRemoteServerID(), serversess.GetRemoteServerType(), serversess.GetRemoteOuter())
}

func (g *LoginServer) OnTerminate(serversess *frame.SSServerSession) {
	elog.InfoAf("LoginServer OnTerminate Remote [ID=%v,Type=%v,Ip=%v] ", serversess.GetRemoteServerID(), serversess.GetRemoteServerType(), serversess.GetRemoteOuter())
}

func OnHandlerS2SLghKickPlayerReq(datas []byte, l *LoginServer) bool {
	loginReq := pb.S2SLghKickPlayerReq{}
	err := proto.Unmarshal(datas, &loginReq)
	if err != nil {
		return false
	}
	loginReq.Loginsrvid = l.GetServerSession().GetRemoteServerID()

	var LoginAckFunc = func() {
		loginAck := pb.S2SHglKickPlayerAck{
			Accountid:       loginReq.Accountid,
			Playeruid:       loginReq.Playeruid,
			Token:           loginReq.Token,
			Newgatewaysrvid: loginReq.Newgatewaysrvid,
		}
		l.GetServerSession().SendProtoMsg(uint32(pb.S2SLogicMsgId_s2s_hgl_kick_player_ack_id), &loginAck, nil)
	}

	player := GPlayerMgr.FindPlayerByAccountID(loginReq.Accountid)
	if player == nil {
		elog.InfoAf("[Login] Not Find AccountId=%v Player", loginReq.Accountid)
		LoginAckFunc()
		return false
	}

	clientSess := player.GetSess()
	if clientSess != nil {
		elog.InfoAf("[OffLine] Player AccountId=%v, SessId=%v Exit Game", player.accountid, clientSess.GetSessID())
		clientSess.Terminate()
	}

	if player.Send2HallProtoMsg(uint32(pb.S2SLogicMsgId_s2s_lgh_kick_player_req_id), &loginReq) == false {
		elog.InfoAf("[Login]  AccountId=%v Send2Hall S2SLghKickPlayerReq Error", loginReq.Accountid)
		LoginAckFunc()
		return false
	}
	elog.InfoAf("[Login] AccountId=%v Send2Hall S2SLghKickPlayerReq Success", loginReq.Accountid)
	return true
}

func OnHandlerLgVerifyTokenAck(datas []byte, l *LoginServer) bool {
	loginAck := pb.LgVerifyTokenAck{}
	err := proto.Unmarshal(datas, &loginAck)
	if err != nil {
		return false
	}

	player := GPlayerMgr.FindPlayer(loginAck.Playeruid)
	if player == nil {
		elog.WarnAf("[Login] VerifyToken Not Find PlayerId=%v Player", loginAck.Playeruid)
		return false
	}

	var FailAck = func() {
		ack := pb.ScGameLoginAck{
			Accountid: loginAck.Accountid,
			Errorcode: frame.MSG_FAIL,
		}
		player.Send2ClientProtoMsg(uint32(pb.EClient2GameMsgId_sc_game_login_ack_id), &ack)
	}
	if loginAck.Errorcode != frame.MSG_SUCCESS {
		FailAck()
		return false
	}

	player.SetAccountId(loginAck.Accountid)
	elog.InfoAf("[Login] ClientSessId=%v AccountId=%v", player.GetSessId(), player.accountid)

	hallServerId := SelectMinHall()
	if hallServerId == 0 {
		return false
	}

	player.SetHallServerId(hallServerId)

	hallReq := pb.GhSelectHallReq{}
	hallReq.Accountid = loginAck.Accountid
	hallReq.ClientsessId = player.GetSessId()
	hallReq.Playeruid = loginAck.Playeruid
	frame.GSSServerSessionMgr.SendProtoMsg(hallServerId, uint32(pb.S2SLogicMsgId_gh_select_hall_req_id), &hallReq)

	return true
}

func SelectMinHall() uint64 {
	hallMap := frame.GSSServerSessionMgr.FindLogicServerByServerType(frame.HALL_SERVER_TYPE)
	minCount := uint64(math.MaxUint64)
	minHallServerId := uint64(0)
	for _, logicserver := range hallMap {
		hall := logicserver.(*HallServer)
		if hall.GetPlayerTotalCount() < minCount {
			minCount = hall.GetPlayerTotalCount()
			minHallServerId = hall.GetServerSession().GetRemoteServerID()
		}
	}

	return minHallServerId
}
