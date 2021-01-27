package main

import (
	"fmt"
	"projects/frame"
	"projects/go-engine/edb"
	"projects/go-engine/elog"
	"projects/go-engine/etimer"
	"projects/pb"
	"projects/rank/tscommon"
	skip_list "projects/thirds/skiplist"
)

func RankGreaterCompare(l, r interface{}) int {
	lRankItem := l.(*pb.RankItem)
	rRankItem := r.(*pb.RankItem)

	if lRankItem.SortField1 < rRankItem.SortField1 {
		return -1
	} else if lRankItem.SortField1 > rRankItem.SortField1 {
		return 1
	}

	if lRankItem.SortField2 < rRankItem.SortField2 {
		return -1
	} else if lRankItem.SortField2 > rRankItem.SortField2 {
		return 1
	}

	if lRankItem.SortField3 < rRankItem.SortField3 {
		return -1
	} else if lRankItem.SortField3 > rRankItem.SortField3 {
		return 1
	}

	if lRankItem.SortField4 < rRankItem.SortField4 {
		return -1
	} else if lRankItem.SortField4 > rRankItem.SortField4 {
		return 1
	}

	if lRankItem.SortField5 < rRankItem.SortField5 {
		return -1
	} else if lRankItem.SortField5 > rRankItem.SortField5 {
		return 1
	}

	if lRankItem.PlayerId < rRankItem.PlayerId {
		return -1
	} else if lRankItem.PlayerId > rRankItem.PlayerId {
		return 1
	}

	return 0
}

const (
	DB_ADD_OR_UDP_RECORD uint32 = 1
	DB_DEL_RECORD        uint32 = 2
)

type DBRankItem struct {
	PlayerID uint64
	RankItem *pb.RankItem
	Flag     uint32
}

const (
	RANK_SAVE_TIMER_ID uint32 = 1
)

const (
	RANK_SAVE_TIMER_DELAY uint64 = 1000 * 10
)

type Rank struct {
	tid           uint32
	rankList      skip_list.SkipList      //RankSortKey - Playerid
	rankMap       map[uint64]*pb.RankItem //Playerid - pb.RankItem
	dbRankItems   []*DBRankItem
	allDelFlag    bool
	timerRegister etimer.ITimerRegister
}

func NewRank(tid uint32) *Rank {
	return &Rank{
		tid:           tid,
		rankList:      skip_list.New(RankGreaterCompare, 32),
		rankMap:       make(map[uint64]*pb.RankItem),
		dbRankItems:   make([]*DBRankItem, 0),
		timerRegister: etimer.NewTimerRegister(),
	}
}

func (r *Rank) Init() {
	r.timerRegister.AddRepeatTimer(RANK_SAVE_TIMER_ID, RANK_SAVE_TIMER_DELAY, "Rank-Save", func(v ...interface{}) {
		r.SaveDB()
	}, []interface{}{}, true)
}

func (r *Rank) GetAllDelFlag() bool {
	return r.allDelFlag
}

func (r *Rank) SetAllDelFlag(flag bool) {
	r.allDelFlag = flag
}

func (r *Rank) EnsureRankListSize() {
	cfg, ok := tscommon.GRankCfg.RankAtrrMap[r.tid]
	if ok {
		if r.rankList.Size() > int(cfg.RankSize) {
			delSize := r.rankList.Size() - int(cfg.RankSize)
			for i := 0; i < delSize; i++ {
				delKey, delValue := r.rankList.Min()
				if delKey != nil {
					delPlayerId := delValue.(uint64)
					delete(r.rankMap, delPlayerId)
					elog.InfoAf("[Rank] Tid=%v Up RankSize Remove PlayerId=%v", r.tid, delPlayerId)
					r.rankList.Erase(delKey)
					r.AddDbItem(DB_DEL_RECORD, delPlayerId, nil)
				}
			}
		}
	}
}

func (r *Rank) Update(item *pb.RankItem, loadflag bool) {
	if item == nil {
		return
	}

	if item.PlayerId == 0 {
		return
	}

	oldRankSortKey, keyOk := r.rankMap[item.PlayerId]
	if keyOk {
		r.rankList.Erase(oldRankSortKey)
	}

	cfg, cfgOk := tscommon.GRankCfg.RankAtrrMap[r.tid]
	if !cfgOk {
		return
	}

	if r.rankList.Size() >= int(cfg.RankSize) {
		minKey, _ := r.rankList.Min()
		if RankGreaterCompare(minKey, item) > 0 {
			elog.InfoAf("Min=%+v,Update=%+v", minKey, item)
			return
		}
	}

	r.rankMap[item.PlayerId] = item
	if err := r.rankList.Insert(item, item.PlayerId); err != nil {
		elog.WarnAf("[Rank] Tid=%v RankList Exist", r.tid, item.PlayerId)
		return
	}

	if !loadflag {
		elog.InfoAf("[Rank] Tid=%v  AddOrUpdate RankItem=%+v", r.tid, item)
		r.AddDbItem(DB_ADD_OR_UDP_RECORD, item.PlayerId, item)
		r.EnsureRankListSize()
	}

	for iter := r.rankList.Begin(); iter != r.rankList.End(); iter = iter.Next() {
		rankItem := iter.Key().(*pb.RankItem)
		elog.InfoAf("[Rank] Tid=%v RankItem=%+v", r.tid, rankItem)
	}
}

func (r *Rank) Remove(Player_id uint64) {
	if Player_id == 0 {
		return
	}

	item, ok := r.rankMap[Player_id]
	if !ok {
		return
	}

	r.rankList.Erase(item)
	elog.InfoAf("[Rank] Tid=%v Remove PlayerId=%v", r.tid, Player_id)
	r.AddDbItem(DB_DEL_RECORD, Player_id, nil)
}

func (r *Rank) ClearAll() {
	r.rankList.Clear()
	for _, dbRankItem := range r.dbRankItems {
		elog.WarnAf("[Rank] ignore DBRankItem Flag=%v,PlayerId=%v RankItem=%+v", dbRankItem.Flag, dbRankItem.PlayerID, dbRankItem.RankItem)
	}
	r.dbRankItems = make([]*DBRankItem, 0)
	r.allDelFlag = true
	elog.InfoAf("[Rank] Tid=%v, ClearAll", r.tid)
}

func (r *Rank) AddDbItem(flag uint32, Player_id uint64, rankItem *pb.RankItem) {
	if Player_id == 0 {
		return
	}

	dbRankItem := &DBRankItem{}
	dbRankItem.Flag = flag
	dbRankItem.PlayerID = Player_id
	dbRankItem.RankItem = rankItem
	r.dbRankItems = append(r.dbRankItems, dbRankItem)
}

func (r *Rank) FillRankItems(topN uint32, ack *pb.Ts2CQueryRankAck) {
	if ack == nil {
		return
	}

	index := uint32(0)
	for iter := r.rankList.Begin(); iter != r.rankList.End(); iter = iter.Next() {
		if index < topN {
			break
		}

		Player_id := iter.Value().(uint64)
		if Player_id == 0 {
			elog.WarnA("[Rank] FillRankItems PlayerId=0 Error")
			continue
		}
		item, ok := r.rankMap[Player_id]
		if !ok {
			elog.WarnAf("[Rank] FillRankItems PlayerId=%v Error", Player_id)
			continue
		}
		index++
		ack.RankList = append(ack.RankList, item)
	}
}

func (r *Rank) SaveDB() {
	if len(r.dbRankItems) == 0 {
		return
	}
	elog.InfoAf("[Rank] Tid=%v SaveDB", r.tid)

	SaveRank(r.allDelFlag, r.tid, r.dbRankItems)
	r.dbRankItems = make([]*DBRankItem, 0)
	r.allDelFlag = false
}

func SaveRank(allDelFlag bool, tid uint32, dbItems []*DBRankItem) {
	type CmdParas struct {
		allDelFlag bool
		tid        uint32
		dbItems    []*DBRankItem
	}

	cmdParas := &CmdParas{
		allDelFlag: allDelFlag,
		tid:        tid,
		dbItems:    dbItems,
	}

	frame.AsyncDoSqlOpt(func(conn edb.IMysqlConn, attach []interface{}) (edb.IMysqlRecordSet, int32, error) {
		paras := attach[0].(*CmdParas)
		tableName := fmt.Sprintf("rank_%v", paras.tid)

		//先清空表
		if paras.allDelFlag == true {
			delete_sql := frame.BuildDeleteSQL(tableName, nil)
			_, delAllErr := conn.QueryWithoutResult(delete_sql)
			if delAllErr != nil {
				elog.ErrorAf("[Rank] Delete All Tid=%v Error=%v", paras.tid, delAllErr)
			}
		}

		//再添加或者更新
		for _, dbItem := range paras.dbItems {
			if dbItem.Flag == DB_DEL_RECORD {
				//删除一条记录
				delete_sql := frame.BuildDeleteSQL(tableName, map[string]interface{}{
					"Playerid": dbItem.PlayerID,
				})
				_, delErr := conn.QueryWithoutResult(delete_sql)
				if delErr != nil {
					elog.ErrorAf("[Rank] Delete Tid=%v PlayerId=%v RankItem=%+v Error=%v", paras.tid, dbItem.PlayerID, dbItem.RankItem, delErr)
				}
			} else if dbItem.Flag == DB_ADD_OR_UDP_RECORD {
				insertOrUpdSql := frame.BuildInsertOrUpdateSQL(tableName, map[string]interface{}{
					"Playerid":    dbItem.PlayerID,
					"sortfield1":  dbItem.RankItem.SortField1,
					"sortfield2":  dbItem.RankItem.SortField2,
					"sortfield3":  dbItem.RankItem.SortField3,
					"sortfield4":  dbItem.RankItem.SortField4,
					"sortfield5":  dbItem.RankItem.SortField5,
					"attachdatas": dbItem.RankItem.AttachDatas,
				}, []string{"Playerid"})

				_, insertOrUpdErr := conn.QueryWithoutResult(insertOrUpdSql)
				if insertOrUpdErr != nil {
					elog.ErrorAf("[Rank] Delete All Tid=%v Error=%v", paras.tid, insertOrUpdErr)
				}
			}
		}

		return nil, edb.DB_EXEC_FAIL, nil
	}, func(recordSet edb.IMysqlRecordSet, attach []interface{}, errorCode int32, err error) {

	}, []interface{}{cmdParas}, TS_DB_DEFAULT_UID)
}
