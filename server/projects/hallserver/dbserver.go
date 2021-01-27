package main

import (
	"projects/frame"
	"projects/go-engine/elog"
	"projects/pb"

	"github.com/golang/protobuf/proto"
)

type DBFunc func(datas []byte, h *DBServer) bool

type DBServer struct {
	frame.LogicServer
	dealer *frame.IDDealer
}

func NewDBServer() *DBServer {
	db := &DBServer{
		dealer: frame.NewIDDealer(),
	}
	db.Init()
	return db
}

func (d *DBServer) Init() bool {
	d.dealer.RegisterHandler(uint32(pb.S2SLogicMsgId_dh_create_player_ack_id), DBFunc(OnHandlerDhCreatePlayerAck))
	return true
}

func (d *DBServer) OnHandler(msgID uint32, attach_datas []byte, datas []byte, sess *frame.SSServerSession) {
	defer func() {
		if err := recover(); err != nil {
			elog.ErrorAf("DBServer onHandler MsgID = %v Error=%v", msgID, err)
		}
	}()

	dealer := d.dealer.FindHandler(msgID)
	if dealer == nil {
		elog.ErrorAf("DBServer MsgHandler Can Not Find MsgID = %v", msgID)
		return
	}

	dealer.(DBFunc)(datas, d)
}

func (d *DBServer) OnEstablish(serversess *frame.SSServerSession) {
	elog.InfoAf("DBServer OnEstablish Remote [ID=%v,Type=%v,Ip=%v]", serversess.GetRemoteServerID(), serversess.GetRemoteServerType(), serversess.GetRemoteOuter())
}

func (d *DBServer) OnTerminate(serversess *frame.SSServerSession) {
	elog.InfoAf("DBServer OnTerminate Remote [ID=%v,Type=%v,Ip=%v]", serversess.GetRemoteServerID(), serversess.GetRemoteServerType(), serversess.GetRemoteOuter())
}

func OnHandlerDhCreatePlayerAck(datas []byte, d *DBServer) bool {
	dbAck := pb.DhCreatePlayerAck{}
	err := proto.Unmarshal(datas, &dbAck)
	if err != nil {
		return false
	}

	player := GPlayerMgr.FindPlayerByAccountID(dbAck.Accountid)
	if player == nil {
		elog.ErrorAf("[Login] DhCreatePlayerAck Not Find Accountid=%v Player", dbAck.Accountid)
		return false
	}

	info := &pb.PlayerBaseInfo{}
	info.Playername = dbAck.Playername
	info.Playerid = dbAck.Playerid
	player.AddPlayerBaseInfo(info)

	clientAck := pb.ScCreatePlayerAck{}
	clientAck.Playername = dbAck.Playername
	clientAck.Errorcode = frame.MSG_SUCCESS
	clientAck.Baseinfolist = make([]*pb.PlayerBaseInfo, 0)
	player.FillPlayerBaseInfos(clientAck.Baseinfolist)
	player.Send2Client(uint32(pb.EClient2GameMsgId_sc_create_player_ack_id), &clientAck)
	return true
}
