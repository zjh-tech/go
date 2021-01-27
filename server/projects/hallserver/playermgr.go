package main

import "projects/go-engine/elog"

type PlayerMgr struct {
	playersByAccountID map[uint64]*Player
	playersByPlayerID  map[uint64]*Player
}

func (p *PlayerMgr) AddPlayerByAccountID(player *Player) {
	p.playersByAccountID[player.GetAccountId()] = player
	elog.InfoAf("[PlayerMgr] AddPlayer AccountID=%v", player.GetAccountId())
}

func (p *PlayerMgr) SelectPlayerByPlayerID(player *Player) {
	for _, baseInfo := range player.GetPlayerBaseInfos() {
		delete(p.playersByPlayerID, baseInfo.Playerid)
	}
	p.playersByAccountID[player.curPlayerId] = player
	elog.InfoAf("[PlayerMgr] SelectPlayer AccountID=%v PlayerID=%v", player.GetAccountId(), player.curPlayerId)
}

func (p *PlayerMgr) FindPlayerByAccountID(accountID uint64) *Player {
	if player, ok := p.playersByAccountID[accountID]; ok {
		return player
	}

	return nil
}

func (p *PlayerMgr) FindPlayerByPlayerID(playerID uint64) *Player {
	if player, ok := p.playersByPlayerID[playerID]; ok {
		return player
	}

	return nil
}

func (p *PlayerMgr) RemovePlayer(player *Player) {
	elog.InfoAf("[PlayerMgr] RemovePlayer AccountID=%v", player.accountId)
	delete(p.playersByAccountID, player.accountId)
	for _, baseInfo := range player.GetPlayerBaseInfos() {
		delete(p.playersByPlayerID, baseInfo.Playerid)
		elog.InfoAf("[PlayerMgr] RemovePlayer AccountID=%v PlayerID=%v", player.accountId, player.curPlayerId)
	}
}

var GPlayerMgr *PlayerMgr

func init() {
	GPlayerMgr = &PlayerMgr{
		playersByAccountID: make(map[uint64]*Player),
		playersByPlayerID:  make(map[uint64]*Player),
	}
}
