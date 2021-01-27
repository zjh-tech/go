package main

import (
	"fmt"
	"projects/go-engine/elog"
	"projects/go-engine/etimer"
	"projects/go-engine/eutil"
	"projects/util"
)

const (
	TOKEN_TIMEOUT_TIMER_ID uint32 = 1
)

const (
	TOKEN_TIMEOUT_TIMER_DELAY uint64 = 1000 * 60
)

const (
	TOKEN_TIMEOUT_TIME int64 = 1000 * 60 * 60
)

type AccountToken struct {
	Token []byte
	Tick  int64
}

func NewAccountToken() *AccountToken {
	return &AccountToken{
		Tick: util.GetMillsecond(),
	}
}

type TokenMgr struct {
	tokenMap      map[uint64]*AccountToken //accountid - AccountToken
	timerRegister etimer.ITimerRegister
	tokenCount    uint64
	desKey        []byte
}

func RandomDesKey() []byte {
	key := make([]byte, 0)
	keyBlock := [...]byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', '1', '2', '3', '4', '5', '6', '7', '8', '9'}
	for i := 0; i < 8; i++ {
		index := util.GetRandom(0, len(keyBlock)-1)
		key = append(key, keyBlock[index])
	}
	return key
}

func NewTokenMgr() *TokenMgr {
	return &TokenMgr{
		tokenMap:      make(map[uint64]*AccountToken),
		timerRegister: etimer.NewTimerRegister(),
		tokenCount:    1,
		desKey:        RandomDesKey(),
	}
}

func (t *TokenMgr) Init() {
	t.timerRegister.AddRepeatTimer(TOKEN_TIMEOUT_TIMER_ID, TOKEN_TIMEOUT_TIMER_DELAY, "TokenMgr-TokenTimeOut", func(v ...interface{}) {
		now := util.GetMillsecond()
		for accountid, info := range t.tokenMap {
			if info.Tick+TOKEN_TIMEOUT_TIME < now {
				elog.InfoAf("[TokenMgr] AccountID=%v Token TimeOut", accountid)
				delete(t.tokenMap, accountid)
			}
		}
	}, []interface{}{}, true)

	elog.InfoAf("[TokenMgr] DesKey=%v", t.desKey)
}

func (t *TokenMgr) GenerateToken(accountid uint64) []byte {
	//algorithm genenrate token
	accountToken := NewAccountToken()
	src := fmt.Sprintf("%v%v%v", accountid, t.tokenCount, util.GetRandom(0, 100))
	t.tokenCount++
	accountToken.Token, _ = eutil.EncryptDes([]byte(src), string(t.desKey))

	t.tokenMap[accountid] = accountToken
	return accountToken.Token
}

func (t *TokenMgr) IsValidToken(accountid uint64, token []byte) bool {
	info, ok := t.tokenMap[accountid]
	if !ok {
		return false
	}

	if len(info.Token) != len(token) {
		return false
	}

	for i := 0; i < len(token); i++ {
		if info.Token[i] != token[i] {
			return false
		}
	}

	return true
}

var GTokenMgr *TokenMgr

func init() {
	GTokenMgr = NewTokenMgr()
}
