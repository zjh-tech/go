package main

import (
	"projects/frame"
	"projects/go-engine/elog"

	"github.com/golang/protobuf/proto"
)

type PlayerState uint32

const (
	InitialState PlayerState = iota
)

type Player struct {
	uid          uint64
	sess         *frame.CSClientSession
	state        PlayerState
	accountid    uint64
	tranferFlag  bool
	hallServerId uint64
	isReconnFlag bool
}

func NewPlayer(uid uint64) *Player {
	p := &Player{
		uid:          uid,
		sess:         nil,
		state:        InitialState,
		tranferFlag:  false,
		isReconnFlag: false,
	}
	return p
}

func (p *Player) GetUID() uint64 {
	return p.uid
}

func (p *Player) GetSess() *frame.CSClientSession {
	return p.sess
}

func (p *Player) SetSess(sess *frame.CSClientSession) {
	p.sess = sess
}

func (p *Player) GetSessId() uint64 {
	return p.sess.GetSessID()
}

func (p *Player) SetPlayerState(state PlayerState) {
	p.state = state
}

func (p *Player) GetPlayerState() PlayerState {
	return p.state
}

func (p *Player) GetAccountId() uint64 {
	return p.accountid
}

func (p *Player) SetAccountId(accountid uint64) {
	p.accountid = accountid
}

func (p *Player) GetHallServerId() uint64 {
	return p.hallServerId
}

func (p *Player) SetHallServerId(serverid uint64) {
	p.hallServerId = serverid
	elog.InfoAf("[Player] AccountID=%v Select HallServerId=%v", p.accountid, serverid)
}

func (p *Player) SetReconnFlag(flag bool) {
	p.isReconnFlag = flag
}

func (p *Player) Send2ClientProtoMsg(msgID uint32, msg proto.Message) bool {
	if p.sess == nil {
		elog.ErrorAf("Player Id=%v AsyncSendProtoMsg Error", p.uid)
		return false
	}

	if p.isReconnFlag == true {
		elog.WarnAf("PlayerId=%v ReconnectState AsyncSendProtoMsg", p.uid)
		return false
	}

	return p.sess.SendProtoMsg(msgID, msg)
}

func (p *Player) Send2HallProtoMsg(msgID uint32, msg proto.Message) bool {
	return frame.GSSServerSessionMgr.SendProtoMsg(p.hallServerId, msgID, msg)
}

func (p *Player) IsTranferFlag() bool {
	return p.tranferFlag
}

func (p *Player) SetTranferFlag() {
	p.tranferFlag = true
}
