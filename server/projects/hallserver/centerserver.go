package main

import (
	"projects/frame"
	"projects/go-engine/elog"
	"projects/pb"

	"github.com/golang/protobuf/proto"
)

type CenterServerFunc func(datas []byte, c *CenterServer) bool

type CenterServer struct {
	frame.LogicServer
	dealer *frame.IDDealer
}

func NewCenterServer() *CenterServer {
	center := &CenterServer{
		dealer: frame.NewIDDealer(),
	}
	center.Init()
	return center
}
func (c *CenterServer) Init() bool {
	c.dealer.RegisterHandler(uint32(pb.S2SLogicMsgId_ch_select_player_list_ack_id), CenterServerFunc(OnHandlerChSelectPlayerListAck))
	c.dealer.RegisterHandler(uint32(pb.EClient2GameMsgId_sc_create_player_ack_id), CenterServerFunc(OnHandlerCgCgCreatePlayerAck))

	return true
}

func (c *CenterServer) OnHandler(msgID uint32, attach_datas []byte, datas []byte, sess *frame.SSServerSession) {
	defer func() {
		if err := recover(); err != nil {
			elog.ErrorAf("CenterServer onHandler MsgID = %v Error=%v", msgID, err)
		}
	}()

	dealer := c.dealer.FindHandler(msgID)
	if dealer == nil {
		elog.ErrorAf("CenterServer MsgHandler Can Not Find MsgID = %v", msgID)
		return
	}
}

func (c *CenterServer) OnEstablish(serversess *frame.SSServerSession) {
	elog.InfoAf("CenterServer OnEstablish Remote [ID=%v,Type=%v,Ip=%v] ", serversess.GetRemoteServerID(), serversess.GetRemoteServerType(), serversess.GetRemoteOuter())
}

func (c *CenterServer) OnTerminate(serversess *frame.SSServerSession) {
	elog.InfoAf("CenterServer OnTerminate Remote [ID=%v,Type=%v,Ip=%v] ", serversess.GetRemoteServerID(), serversess.GetRemoteServerType(), serversess.GetRemoteOuter())
}

func OnHandlerChSelectPlayerListAck(datas []byte, c *CenterServer) bool {
	centerAck := pb.ChSelectPlayerListAck{}
	err := proto.Unmarshal(datas, &centerAck)
	if err != nil {
		return false
	}

	player := GPlayerMgr.FindPlayerByAccountID(centerAck.Accountid)
	if player == nil {
		elog.ErrorAf("[Login] ChSelectPlayerListAck Not Find Accountid=%v Player", centerAck.Accountid)
		return false
	}

	gateAck := pb.HgSelectHallAck{}
	gateAck.Accountid = player.accountId
	gateAck.Playeruid = player.gatePlayerUId
	player.Send2Gateway(uint32(pb.S2SLogicMsgId_hg_select_hall_ack_id), &gateAck)

	clientAck := pb.ScGameLoginAck{}
	clientAck.Accountid = centerAck.Accountid
	clientAck.Errorcode = centerAck.Errorcode
	clientAck.Baseinfolist = make([]*pb.PlayerBaseInfo, 0)
	for _, Playerbase := range centerAck.Baseinfolist {
		clientPlayerBase := &pb.PlayerBaseInfo{}
		clientPlayerBase.Playerid = Playerbase.Playerid
		clientPlayerBase.Playername = Playerbase.Playername
		clientAck.Baseinfolist = append(clientAck.Baseinfolist, clientPlayerBase)
	}
	player.SetPlayerBaseInfos(clientAck.Baseinfolist)
	player.Send2Client(uint32(pb.EClient2GameMsgId_sc_game_login_ack_id), &clientAck)
	elog.InfoAf("[Login] Send PlayerList AccountId=%v", player.accountId)
	return true
}

func OnHandlerCgCgCreatePlayerAck(datas []byte, c *CenterServer) bool {
	centerAck := pb.ChCreatePlayerAck{}
	err := proto.Unmarshal(datas, &centerAck)
	if err != nil {
		return false
	}

	player := GPlayerMgr.FindPlayerByAccountID(centerAck.Accountid)
	if player == nil {
		elog.ErrorAf("[Login] ChCreatePlayerAck Not Find Accountid=%v Player", centerAck.Accountid)
		return false
	}

	var ClientAckFunc = func(errorCode uint32) {
		clientAck := pb.ScCreatePlayerAck{}
		clientAck.Errorcode = errorCode
		if centerAck.Errorcode == frame.MSG_SUCCESS {
			clientAck.Baseinfolist = make([]*pb.PlayerBaseInfo, 0)
			player.FillPlayerBaseInfos(clientAck.Baseinfolist)
		}
		player.Send2Client(uint32(pb.EClient2GameMsgId_sc_create_player_ack_id), &clientAck)
	}

	if centerAck.Errorcode == frame.MSG_FAIL {
		ClientAckFunc(frame.MSG_FAIL)
		return false
	}

	if player.GetPlayerBaseLen() == 0 {
		player.FirstCreate()
	}

	dbReq := pb.HdCreatePlayerReq{}
	dbReq.Accountid = player.accountId
	dbReq.Playerid = centerAck.Playerid
	dbReq.Playername = centerAck.Playername
	if frame.GSSServerSessionMgr.SendProtoMsgBySessionID(player.GetDBSessionID(), uint32(pb.S2SLogicMsgId_hd_create_player_req_id), &dbReq) == false {
		elog.ErrorAf("CreatePlayer AccountID=%v PlayerId=%v PlayerName=%v Error", player.accountId, centerAck.Playerid, centerAck.Playername)
	}
	return true
}
