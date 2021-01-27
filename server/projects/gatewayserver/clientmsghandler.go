package main

import (
	"projects/frame"
	"projects/go-engine/elog"
	"projects/pb"

	"github.com/golang/protobuf/proto"
)

type GameClientFunc func(datas []byte, player *Player, sess *frame.CSClientSession) bool

type ClientMsgHandler struct {
	dealer *frame.IDDealer
}

func (c *ClientMsgHandler) Init() bool {
	c.dealer.RegisterHandler(uint32(pb.EClient2GameMsgId_cs_game_login_req_id), GameClientFunc(OnHandlerCsGameLoginReq))
	c.dealer.RegisterHandler(uint32(pb.EClient2GameMsgId_cs_exit_game_req_id), GameClientFunc(OnHandlerCsExitGameReq))
	c.dealer.RegisterHandler(uint32(pb.EClient2GameMsgId_cs_reconnect_game_req_id), GameClientFunc(OnHandlerCsReconnectGameReq))
	return true
}

func (c *ClientMsgHandler) OnHandler(msgID uint32, datas []byte, sess *frame.CSClientSession) {
	defer func() {
		if err := recover(); err != nil {
			elog.ErrorAf("GatewayServer ClientMsgHandler onHandler MsgID = %v Error=%v", msgID, err)
		}
	}()

	playerid := sess.GetAttach().(uint64)
	player := GPlayerMgr.FindPlayer(playerid)
	if player == nil {
		elog.ErrorAf("GatewayServer ClientMsgHandler Not Find Id=%v Player", playerid)
		return
	}

	dealer := c.dealer.FindHandler(msgID)
	if dealer != nil {
		dealer.(GameClientFunc)(datas, player, sess)
		return
	} else {
		if player.IsTranferFlag() {
			tranfer := pb.S2SGhTranferMsg{}
			tranfer.Msgid = msgID
			tranfer.Datas = datas
			tranfer.Accountid = player.accountid
			player.Send2HallProtoMsg(uint32(pb.S2SLogicMsgId_s2s_gh_tranfer_msg_id), &tranfer)
		} else {
			elog.WarnAf("GatewayServer AccountId=%v TranferFlag Error", player.accountid)
		}
	}
}

func (c *ClientMsgHandler) OnConnect(sess *frame.CSClientSession) {
	player := GPlayerMgr.CreatePlayer()
	player.SetSess(sess)
	sess.SetAttach(player.GetUID())
}

func (c *ClientMsgHandler) OnDisconnect(sess *frame.CSClientSession) {
	player := c.GetPlayerByAttach(sess)
	if player == nil {
		elog.InfoAf("ClientMsgHandler Session SessID=%v Maybe Reconnect State", sess.GetSessID())
		return
	}

	player.SetSess(nil)
	GPlayerMgr.RemovePlayer(player.GetUID())
}

func (c *ClientMsgHandler) OnBeatHeartError(sess *frame.CSClientSession) {
	player := c.GetPlayerByAttach(sess)
	if player != nil {
		req := pb.S2SGhKickPlayerReq{}
		if player.Send2HallProtoMsg(uint32(pb.S2SLogicMsgId_s2s_gh_kick_player_req_id), &req) == false {
			elog.InfoAf("BeatHeartError Kick Player AccountId=%v Send Hall Error", player.accountid)
		} else {
			elog.InfoAf("BeatHeartError Kick Player AccountId=%v", player.accountid)
		}
	}
}

func (c *ClientMsgHandler) GetPlayerByAttach(sess *frame.CSClientSession) *Player {
	attach := sess.GetAttach()
	if attach == nil {
		return nil
	}
	playeruid := attach.(uint64)
	if playeruid == 0 {
		return nil
	}

	return GPlayerMgr.FindPlayer(playeruid)
}

func OnHandlerCsGameLoginReq(datas []byte, player *Player, sess *frame.CSClientSession) bool {
	req := pb.CsGameLoginReq{}
	unmarshalErr := proto.Unmarshal(datas, &req)
	if unmarshalErr != nil {
		return false
	}

	loginReq := pb.GlVerifyTokenReq{
		Accountid: req.Accountid,
		Token:     req.Token,
		Playeruid: player.uid,
	}

	if frame.GSSServerSessionMgr.SendProtoMsg(req.Loginserverid, uint32(pb.S2SLogicMsgId_gl_verify_token_req_id), &loginReq) == false {
		elog.ErrorAf("[Login] CsGameLoginReq AccountId=%v Send2Login LoginServerId=%v", req.Accountid, req.Loginserverid)
		return false
	}
	return true
}

func OnHandlerCsExitGameReq(datas []byte, player *Player, sess *frame.CSClientSession) bool {
	elog.InfoAf("[OffLine] Player AccountId=%v, SessId=%v Exit Game", player.accountid, sess.GetSessID())
	sess.Terminate()
	req := pb.S2SGhExitGameReq{}
	player.Send2HallProtoMsg(uint32(pb.S2SLogicMsgId_s2s_gh_exit_game_req_id), &req)
	return true
}

func OnHandlerCsReconnectGameReq(datas []byte, player *Player, sess *frame.CSClientSession) bool {
	req := pb.CsReconnectGameReq{}
	unmarshalErr := proto.Unmarshal(datas, &req)
	if unmarshalErr != nil {
		return false
	}

	var FailAck = func() {
		ack := pb.ScReconnectGameAck{}
		ack.Errorcode = frame.MSG_FAIL
		player.Send2ClientProtoMsg(uint32(pb.EClient2GameMsgId_sc_reconnect_game_ack_id), &ack)
	}

	lastPlayer := GPlayerMgr.FindPlayerByAccountID(req.Accountid)
	if lastPlayer == nil {
		elog.InfoAf("[Login] Reconnect AccountId=%v Not Find Player", req.Accountid)
		FailAck()
		return false
	}
	oldSess := lastPlayer.GetSess()
	if oldSess == nil {
		FailAck()
		return false
	}
	//解除
	oldSess.SetAttach(0)
	lastPlayer.SetSess(nil)
	player.SetSess(nil)
	//绑定
	lastPlayer.SetSess(sess)
	sess.SetAttach(lastPlayer.GetUID())

	GPlayerMgr.RemovePlayer(player.GetUID())
	lastPlayer.SetReconnFlag(true)
	oldSess.Terminate()
	elog.InfoAf("[Login] AccountId=%v Reconnect Ok", req.Accountid)

	gateReq := pb.S2SGhReconnectGameReq{}
	gateReq.ClientsessId = sess.GetSessID()
	if lastPlayer.Send2HallProtoMsg(uint32(pb.S2SLogicMsgId_s2s_gh_reconnect_game_req_id), &gateReq) == false {
		elog.ErrorAf("[Login] AccountId=%v Reconnect Send Hall Error", req.Accountid)
		FailAck()
		return false
	}

	return true
}

var GClientMsgHandler *ClientMsgHandler

func init() {
	GClientMsgHandler = &ClientMsgHandler{
		dealer: frame.NewIDDealer(),
	}
	GClientMsgHandler.Init()
}
