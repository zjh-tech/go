package tscommon

import (
	"errors"
	"fmt"
	"projects/thirds/etree"
	"projects/util"
)

type RankAtrr struct {
	Id           uint32
	GameId       uint32
	FuncId       uint32
	GameDesc     string
	FuncDesc     string
	RankSize     uint32
	RankServerId uint64
}

type RankCfg struct {
	RankAtrrMap map[uint32]*RankAtrr
}

func NewRankCfg() *RankCfg {
	return &RankCfg{
		RankAtrrMap: make(map[uint32]*RankAtrr),
	}
}

func (r *RankCfg) GetTIds(serverId uint64) []uint32 {
	tids := make([]uint32, 0)
	for tid, _ := range r.RankAtrrMap {
		tids = append(tids, tid)
	}
	return tids
}

var GRankCfg *RankCfg

func ReadRankCfg(path string) (*RankCfg, error) {
	doc := etree.NewDocument()
	err := doc.ReadFromFile(path)
	if err != nil {
		return nil, err
	}

	cfgElem := doc.SelectElement("config")
	if cfgElem == nil {
		return nil, errors.New("rank_config Xml Config Error")
	}

	cfg := NewRankCfg()
	if cfgElem != nil {
		for _, rankElem := range cfgElem.FindElements("rank") {
			rankAtrr := &RankAtrr{}
			for _, attr := range rankElem.Attr {
				if attr.Key == "Id" {
					rankAtrr.Id, _ = util.Str2Uint32(attr.Value)
				} else if attr.Key == "GameId" {
					rankAtrr.GameId, _ = util.Str2Uint32(attr.Value)
				} else if attr.Key == "FuncId" {
					rankAtrr.FuncId, _ = util.Str2Uint32(attr.Value)
				} else if attr.Key == "GameDesc" {
					rankAtrr.GameDesc = attr.Value
				} else if attr.Key == "FuncDesc" {
					rankAtrr.FuncDesc = attr.Value
				} else if attr.Key == "RankSize" {
					rankAtrr.RankSize, _ = util.Str2Uint32(attr.Value)
				} else if attr.Key == "RankSrvId" {
					rankAtrr.RankServerId, _ = util.Str2Uint64(attr.Value)
				} else {
					errStr := fmt.Sprintf("rank_config Xml %v Error", attr.Value)
					return nil, errors.New(errStr)
				}
			}
			cfg.RankAtrrMap[rankAtrr.Id] = rankAtrr
		}
	}

	return cfg, nil
}

func init() {
	GRankCfg = NewRankCfg()
}
