package main

import (
	"projects/frame"
	"projects/go-engine/elog"
	"projects/pb"
	"strings"

	"github.com/golang/protobuf/proto"
)

type GatewayFunc func(datas []byte, g *GatewayServer) bool
type ClientFunc func(datas []byte, p *Player, g *GatewayServer) bool

type GatewayServer struct {
	frame.LogicServer
	gatewayDealer *frame.IDDealer
	clientDealer  *frame.IDDealer
}

func NewGatewayServer() *GatewayServer {
	gateway := &GatewayServer{
		gatewayDealer: frame.NewIDDealer(),
		clientDealer:  frame.NewIDDealer(),
	}
	gateway.Init()
	return gateway
}

func (g *GatewayServer) Init() bool {
	g.clientDealer.RegisterHandler(uint32(pb.EClient2GameMsgId_cs_create_player_req_id), ClientFunc(OnHandlerCsCreatePlayerReq))
	g.clientDealer.RegisterHandler(uint32(pb.EClient2GameMsgId_cs_select_player_req_id), ClientFunc(OnHandlerCsSelectPlayerReq))
	g.clientDealer.RegisterHandler(uint32(pb.EClient2GameMsgId_cs_gm_req_id), ClientFunc(OnHandlerCsGmReq))

	g.gatewayDealer.RegisterHandler(uint32(pb.S2SLogicMsgId_gh_select_hall_req_id), GatewayFunc(OnHandlerGhSelectHallReq))
	g.gatewayDealer.RegisterHandler(uint32(pb.S2SLogicMsgId_s2s_gh_tranfer_msg_id), GatewayFunc(OnHandlerS2SGhTranferMsg))
	g.gatewayDealer.RegisterHandler(uint32(pb.S2SLogicMsgId_s2s_gh_exit_game_req_id), ClientFunc(OnHandlerS2SGhExitGameReq))
	g.gatewayDealer.RegisterHandler(uint32(pb.S2SLogicMsgId_s2s_gh_reconnect_game_req_id), ClientFunc(OnHandlerS2SGhReconnectGameReq))
	g.gatewayDealer.RegisterHandler(uint32(pb.S2SLogicMsgId_s2s_gh_kick_player_req_id), ClientFunc(OnHandlerS2SGhKickPlayerReq))
	g.gatewayDealer.RegisterHandler(uint32(pb.S2SLogicMsgId_s2s_lgh_kick_player_req_id), ClientFunc(OnHandlerS2SLghKickPlayerReq))

	return true
}

func (g *GatewayServer) OnHandler(msgID uint32, attach_datas []byte, datas []byte, sess *frame.SSServerSession) {
	defer func() {
		if err := recover(); err != nil {
			elog.ErrorAf("GatewayServer onHandler MsgID = %v Error=%v", msgID, err)
		}
	}()

	dealer := g.gatewayDealer.FindHandler(msgID)
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

func OnHandlerS2SGhTranferMsg(datas []byte, g *GatewayServer) bool {
	tranfer := pb.S2SGhTranferMsg{}
	err := proto.Unmarshal(datas, &tranfer)
	if err != nil {
		return false
	}

	dealer := g.clientDealer.FindHandler(tranfer.Msgid)
	if dealer == nil {
		elog.ErrorAf("GatewayServer ClientMsgHandler Can Not Find MsgID = %v", tranfer.Msgid)
		return false
	}

	player := GPlayerMgr.FindPlayerByAccountID(tranfer.Accountid)
	if player == nil {
		elog.ErrorAf("GatewayServer ClientMsgHandler FindPlayerByAccountID =%v Error", tranfer.Accountid)
		return false
	}
	dealer.(ClientFunc)(datas, player, g)
	return true
}

func OnHandlerGhSelectHallReq(datas []byte, g *GatewayServer) bool {
	req := pb.GhSelectHallReq{}
	err := proto.Unmarshal(datas, &req)
	if err != nil {
		return false
	}

	player := NewPlayer(req.Accountid, req.ClientsessId, req.Playeruid, g.GetServerSession().GetSessID())
	GPlayerMgr.AddPlayerByAccountID(player)

	centerReq := pb.HcSelectPlayerListReq{
		Accountid: req.Accountid,
		Playeruid: req.ClientsessId,
	}

	if player.Send2CenterHashAccountId(uint32(pb.S2SLogicMsgId_hc_select_player_list_req_id), &centerReq) == false {
		elog.Errorf("[Login] HcSelectPlayerListReq AccountId=%v", req.Accountid)
		return false
	}
	return true
}

func OnHandlerCsCreatePlayerReq(datas []byte, p *Player, g *GatewayServer) bool {
	req := pb.CsCreatePlayerReq{}
	err := proto.Unmarshal(datas, &req)
	if err != nil {
		return false
	}

	var ClientAckFunc = func(errorcode uint32) {
		clientAck := pb.ScCreatePlayerAck{}
		clientAck.Errorcode = errorcode
		p.Send2Client(uint32(pb.EClient2GameMsgId_sc_create_player_ack_id), &clientAck)
	}

	if p.GetAccountId() == 0 {
		return false
	}

	if req.Playername != "" {
		elog.InfoAf("[Login] AccountID=%v CreatePlayerReq PlayerName Error", p.accountId)
		ClientAckFunc(frame.MSG_FAIL)
		return false
	}

	if p.GetPlayerBaseLen() >= MaxPlayerCount {
		elog.InfoAf("[Login] AccountID=%v CreatePlayerReq PlayerBaseLen Is MaxCount", p.accountId)
		ClientAckFunc(frame.MSG_FAIL)
		return false
	}

	elog.InfoAf("[Login] AccountID=%v Create Playername=%v", p.accountId, req.Playername)

	centerReq := pb.HcCreatePlayerReq{}
	centerReq.Playername = req.Playername
	centerReq.Accountid = p.GetAccountId()
	centerReq.Playerid = uint64(frame.GIdMaker.NextId())
	if p.Send2CenterHashAccountId(uint32(pb.S2SLogicMsgId_hc_create_player_req_id), &centerReq) == false {
		elog.Errorf("[Login] CreatePlayerReq AccountId=%v", p.accountId)
		return false
	}

	return true
}

func OnHandlerCsSelectPlayerReq(datas []byte, p *Player, g *GatewayServer) bool {
	req := pb.CsSelectPlayerReq{}
	err := proto.Unmarshal(datas, &req)
	if err != nil {
		return false
	}

	p.SelectPlayer(req.Playerid)
	return true
}

func OnHandlerCsGmReq(datas []byte, p *Player, g *GatewayServer) bool {
	req := pb.CsGmReq{}
	err := proto.Unmarshal(datas, &req)
	if err != nil {
		return false
	}

	if len(req.Content) == 0 {
		return false
	}

	if req.Content[0] != '@' {
		return false
	}

	elog.WarnAf("[Player] AccountId=%v PlayerId=%v GM Content=%v", p.accountId, p.curPlayerId, req.Content)
	contents := strings.Split(req.Content, "")
	if len(contents) == 0 {
		return false
	}
	p.ProcessGM(contents)
	return true
}

func PlayerExitGame(p *Player) {
	if p == nil {
		return
	}
	GPlayerMgr.RemovePlayer(p)
	p.OffLine()
}

func OnHandlerS2SGhExitGameReq(datas []byte, p *Player, g *GatewayServer) bool {
	elog.InfoAf("[OffLine] Player AccountId=%v Exit Game", p.accountId)
	PlayerExitGame(p)
	return true
}

func OnHandlerS2SGhKickPlayerReq(datas []byte, p *Player, g *GatewayServer) bool {
	elog.InfoAf("[OffLine] S2SGhKickPlayerReq Kill Player AccountId=%v", p.accountId)
	PlayerExitGame(p)
	return true
}

func OnHandlerS2SLghKickPlayerReq(datas []byte, p *Player, g *GatewayServer) bool {
	elog.InfoAf("[OffLine] S2SLghKickPlayerReq Kill Player AccountId=%v", p.accountId)
	PlayerExitGame(p)

	loginReq := pb.S2SLghKickPlayerReq{}
	err := proto.Unmarshal(datas, &loginReq)
	if err != nil {
		return false
	}

	loginAck := pb.S2SHglKickPlayerAck{
		Accountid:       loginReq.Accountid,
		Playeruid:       loginReq.Playeruid,
		Token:           loginReq.Token,
		Newgatewaysrvid: loginReq.Newgatewaysrvid,
		Loginsrvid:      loginReq.Loginsrvid,
	}
	g.GetServerSession().SendProtoMsg(uint32(pb.S2SLogicMsgId_s2s_hgl_kick_player_ack_id), &loginAck, nil)

	return true
}

func OnHandlerS2SGhReconnectGameReq(datas []byte, p *Player, g *GatewayServer) bool {
	req := pb.S2SGhReconnectGameReq{}
	err := proto.Unmarshal(datas, &req)
	if err != nil {
		return false
	}

	elog.InfoAf("[Login] AccountId=%v Reconnect Ok", p.accountId)
	p.SetReconnParas(req.ClientsessId)

	ack := pb.ScReconnectGameAck{}
	ack.Errorcode = frame.MSG_SUCCESS
	p.Send2Client(uint32(pb.EClient2GameMsgId_sc_reconnect_game_ack_id), &ack)
	return true
}
