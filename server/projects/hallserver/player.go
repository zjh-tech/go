package main

import (
	"projects/frame"
	"projects/go-engine/elog"
	"projects/go-engine/etimer"
	"projects/pb"
	"projects/util"

	"github.com/golang/protobuf/proto"
)

const (
	MaxPlayerCount int = 1
)

type Player struct {
	accountId     uint64
	clientSessId  uint64
	gatePlayerUId uint64
	gatewaySessId uint64
	timeRegister  etimer.ITimerRegister
	dbSessionId   uint64

	playerbaseInfos []*pb.PlayerBaseInfo //玩家列表
	curPlayerId     uint64
	curPlayerInfo   *PlayerAllInfo
}

func NewPlayer(accountid uint64, clientSessId uint64, gateplayeruid uint64, gatewaySessId uint64) *Player {
	p := &Player{
		accountId:       accountid,
		clientSessId:    clientSessId,
		gatePlayerUId:   gateplayeruid,
		gatewaySessId:   gatewaySessId,
		playerbaseInfos: make([]*pb.PlayerBaseInfo, 0),
	}
	return p
}

func (p *Player) GetDBSessionID() uint64 {
	if p.dbSessionId == 0 {
		p.dbSessionId = frame.GSSServerSessionMgr.GetSessionIdByHashIdAndSrvType(p.accountId, frame.DB_SERVER_TYPE)
	}

	return p.dbSessionId
}

func (p *Player) GetAccountId() uint64 {
	return p.accountId
}

func (p *Player) GetGatewaySessID() uint64 {
	return p.clientSessId
}

func (p *Player) SetReconnParas(clientSessId uint64) {
	p.clientSessId = clientSessId
}

func (p *Player) ResetPlayerbaseInfos() {
	p.playerbaseInfos = make([]*pb.PlayerBaseInfo, 0)
}

func (p *Player) GetPlayerBaseLen() int {
	return len(p.playerbaseInfos)
}

func (p *Player) GetPlayerBaseInfos() []*pb.PlayerBaseInfo {
	return p.playerbaseInfos
}

func (p *Player) SetPlayerBaseInfos(infos []*pb.PlayerBaseInfo) {
	p.playerbaseInfos = infos
}

func (p *Player) AddPlayerBaseInfo(info *pb.PlayerBaseInfo) {
	p.playerbaseInfos = append(p.playerbaseInfos, info)
	elog.InfoA("[Player] AccountID=%v Create Player Ok PlayerBaseInfo=%v", p.accountId, info)
}

func (p *Player) FillPlayerBaseInfos(infos []*pb.PlayerBaseInfo) {
	for _, info := range p.playerbaseInfos {
		infos = append(infos, info)
	}
}

func (p *Player) FirstCreate() {

	//Save
}

func (p *Player) OnLine() {
	p.SendScServerTimeNtf()
}

func (p *Player) OffLine() {

}

func (p *Player) SelectPlayer(playerId uint64) {

}

func (p *Player) ProcessGM(contents []string) {

}

func (p *Player) SendScServerTimeNtf() {
	ntf := pb.ScServerTimeNtf{}
	ntf.Millsecond = util.GetMillsecond()
	p.Send2Client(uint32(pb.EClient2GameMsgId_sc_server_time_ntf_id), &ntf)
}

func (p *Player) Send2Client(msgID uint32, msg proto.Message) bool {
	datas, err := proto.Marshal(msg)
	if err != nil {
		elog.ErrorAf("[Player] Send2Client Msg=%v Marshal Err %v ", msgID, err)
		return false
	}

	tranfer := pb.S2S2CHgcTranferMsg{}
	tranfer.Msgid = msgID
	tranfer.Datas = datas
	return frame.GSSServerSessionMgr.SendProtoMsg(p.clientSessId, uint32(pb.S2SLogicMsgId_s2s_hgc_tranfer_msg_id), &tranfer)
}

func (p *Player) Send2Gateway(msgID uint32, msg proto.Message) bool {
	return frame.GSSServerSessionMgr.SendProtoMsg(p.gatewaySessId, msgID, msg)
}

func (p *Player) Send2CenterHashAccountId(msgID uint32, msg proto.Message) bool {
	datas, err := proto.Marshal(msg)
	if err != nil {
		elog.ErrorAf("[Player] Send2Center Msg=%v Marshal Err %v ", msgID, err)
		return false
	}

	tranfer := pb.S2SHcTranferMsg{}
	tranfer.Msgid = msgID
	tranfer.Datas = datas
	return frame.GSSServerSessionMgr.SendProtoMsgByHashIdAndSrvType(p.accountId, frame.CENTER_SERVER_TYPE,
		uint32(pb.S2SLogicMsgId_s2s_hc_tranfer_msg_id), &tranfer)
}
