package main

import (
	"projects/frame"
	"projects/go-engine/elog"
	"projects/go-engine/etimer"
	"projects/pb"
	"projects/util"
	"strings"
)

const (
	PLAYERMGR_OUTPUT_TIMER_ID uint32 = 1
	PLAYERMGR_REPORT_TIMER_ID uint32 = 2
)

const (
	PLAYERMGR_OUTPUT_TIMER_DELAY uint64 = 1000 * 60
	PLAYERMGR_REPORT_TIMER_DELAY uint64 = 1000 * 60
)

type PlayerMgr struct {
	players       map[uint64]*Player
	nextId        uint64
	timerRegister etimer.ITimerRegister
}

func (p *PlayerMgr) Init() {
	p.timerRegister.AddRepeatTimer(PLAYERMGR_OUTPUT_TIMER_ID, PLAYERMGR_OUTPUT_TIMER_DELAY, "PlayerMgr-OutPut", func(v ...interface{}) {
		elog.InfoAf("[PlayerMgr] OutPut Player TotalCount=%v", len(p.players))
	}, []interface{}{}, true)

	p.timerRegister.AddRepeatTimer(PLAYERMGR_REPORT_TIMER_ID, PLAYERMGR_REPORT_TIMER_DELAY, "PlayerMgr-OutPut", func(v ...interface{}) {
		ntf := &pb.GlGatewayInfoNtf{}
		listenSlices := strings.Split(frame.GServerCfg.C2SOuterListen, ":")
		if len(listenSlices) != 2 {
			elog.ErrorAf("[PlayerMgr] Report C2SOuterListen Split Error")
			return
		}
		ntf.Ip = listenSlices[0]
		ntf.Port, _ = util.Str2Uint32(listenSlices[1])
		frame.GSSServerSessionMgr.BroadProtoMsg(frame.LOGIN_SERVER_TYPE, uint32(pb.S2SLogicMsgId_gl_gateway_info_ntf_id), ntf)
	}, []interface{}{}, true)

}

func (p *PlayerMgr) CreatePlayer() *Player {
	p.nextId++
	player := NewPlayer(p.nextId)
	p.players[player.GetUID()] = player
	return player
}

func (p *PlayerMgr) FindPlayer(uid uint64) *Player {
	player, ok := p.players[uid]
	if ok {
		return player
	}

	return nil
}

func (p *PlayerMgr) FindPlayerByAccountID(accountID uint64) *Player {
	for _, player := range p.players {
		if player.accountid == accountID {
			return player
		}
	}
	return nil
}

func (p *PlayerMgr) RemovePlayer(uid uint64) {
	if _, ok := p.players[uid]; ok {
		delete(p.players, uid)
	}
}

func (p *PlayerMgr) GetAllPlayer() map[uint64]*Player {
	return p.players
}

var GPlayerMgr *PlayerMgr

func init() {
	GPlayerMgr = &PlayerMgr{
		players:       make(map[uint64]*Player),
		nextId:        0,
		timerRegister: etimer.NewTimerRegister(),
	}
}
