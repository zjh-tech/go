package main

import (
	"errors"
	"fmt"
)

type RankMgr struct {
	globalRanks map[uint32]*Rank //RankAtrr.Id作为实例ID
}

func NewRankMgr() *RankMgr {
	return &RankMgr{
		globalRanks: make(map[uint32]*Rank),
	}
}

func (r *RankMgr) Init(tids []uint32) {
	for _, tid := range tids {
		rank := NewRank(tid)
		rank.Init()
		r.globalRanks[tid] = rank
	}
}

func (r *RankMgr) FindGlobalRank(tid uint32) (*Rank, error) {
	if GRankCfg == nil {
		return nil, errors.New("[RankMgr] FindGlobalRank GRankCfg Error")
	}

	_, ok := GRankCfg.RankAtrrMap[tid]
	if !ok {
		errStr := fmt.Sprintf("[RankMgr] FindGlobalRank Tid=%v Error", tid)
		return nil, errors.New(errStr)
	}

	globalRank, globalOk := r.globalRanks[tid]
	if globalOk {
		return globalRank, nil
	}

	return nil, nil
}

var GRankMgr *RankMgr

func init() {
	GRankMgr = NewRankMgr()
}
