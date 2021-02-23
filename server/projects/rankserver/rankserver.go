package main

import (
	"fmt"
	"projects/frame"
	"projects/go-engine/edb"
	"projects/go-engine/ehttp"
	"projects/go-engine/elog"
	"projects/go-engine/enet"
	"projects/go-engine/etimer"
	"projects/pb"
	"projects/util"
	"time"
)

const (
	DB_DEFAULT_UID uint64 = 0
)

type RankServer struct {
	frame.Server
}

func (r *RankServer) Init() bool {
	if !r.Server.Init() {
		return false
	}

	if !enet.GNet.Init() {
		elog.Error("RankServer Net Init Error")
		return false
	}
	elog.Info("RankServer Net Init Success")

	if rankCfg, err := ReadRankCfg("./config/rank_config.xml"); err != nil {
		elog.Errorf("RankServer Load RankConfig xml Error=%v", err)
		return false
	} else {
		GRankCfg = rankCfg
	}

	if err := frame.GDatabaseCfgMgr.Load("./db_cfg.xml"); err != nil {
		elog.Error(err)
		return false
	}

	if err := edb.GDBModule.Init(frame.GDatabaseCfgMgr.DBConnMaxCount, frame.GDatabaseCfgMgr.DBTableMaxCount, frame.GDatabaseCfgMgr.DBConnSpecs); err != nil {
		elog.Error(err)
		return false
	}

	tids := GRankCfg.GetTIds()
	GRankMgr.Init(tids)
	r.LoadRankByIndex(0, tids)

	frame.GSSClientSessionMgr.SSClientListen(frame.GServerCfg.RankServerAddr, NewRankClient(), nil)

	elog.Info("RankServer Init Success")
	return true
}

func (r *RankServer) LoadRankByIndex(index int, tids []uint32) {
	if len(tids) == index {
		elog.Error("RankServer LoadRankByIndex Ok")
		return
	}

	type CmdParas struct {
		index int
		tids  []uint32
	}

	cmdParas := &CmdParas{
		index: index,
		tids:  tids,
	}

	frame.AsyncDoSqlOpt(func(conn edb.IMysqlConn, attach []interface{}) (edb.IMysqlRecordSet, int32, error) {
		paras := attach[0].(*CmdParas)
		tableName := fmt.Sprintf("rank_%v", paras.tids[paras.index])
		select_sql := frame.BuildSelectSQL(tableName, []string{
			"playerid",
			"sortfield1",
			"sortfield2",
			"sortfield3",
			"sortfield4",
			"sortfield5",
			"attachdatas",
		}, nil)

		result, selectErr := conn.QueryWithResult(select_sql)
		if selectErr != nil {
			return nil, edb.DB_EXEC_FAIL, selectErr
		}

		return result, edb.DB_EXEC_SUCCESS, nil
	}, func(recordSet edb.IMysqlRecordSet, attach []interface{}, errorCode int32, err error) {
		paras := attach[0].(*CmdParas)
		tid := paras.tids[paras.index]
		if errorCode == edb.DB_EXEC_FAIL {
			elog.ErrorAf("[LoadRankByIndex] Tid=%v Error", tid)
			frame.GServer.Quit()
			return
		}

		if errorCode == edb.DB_EXEC_SUCCESS {
			rank, _ := GRankMgr.FindGlobalRank(tid)
			if rank == nil {
				elog.WarnAf("[LoadRankByIndex] Tid=%v Rank Error", tid)
				frame.GServer.Quit()
				return
			}

			rc := recordSet.GetRecordSet()
			rcLen := len(rc)
			for j := 0; j < rcLen; j++ {
				rankItem := &pb.RankItem{}
				rankItem.PlayerId, _ = util.Str2Uint64(rc[j]["Playerid"])
				rankItem.SortField1, _ = util.Str2Int64(rc[j]["sortfield1"])
				rankItem.SortField2, _ = util.Str2Int64(rc[j]["sortfield2"])
				rankItem.SortField3, _ = util.Str2Int64(rc[j]["sortfield3"])
				rankItem.SortField4, _ = util.Str2Int64(rc[j]["sortfield4"])
				rankItem.SortField5, _ = util.Str2Int64(rc[j]["sortfield5"])
				rankItem.AttachDatas = []byte(rc[j]["attachdatas"])
				rank.Update(rankItem, true)
			}
			paras.index++
			r.LoadRankByIndex(paras.index, paras.tids)
			return
		}
	}, []interface{}{cmdParas}, DB_DEFAULT_UID)
}

func (t *RankServer) Run() {
	busy := false
	net_module := enet.GNet
	http_net_module := ehttp.GHttpNet
	db_module := edb.GDBModule
	timer_module := etimer.GTimerMgr
	meter := frame.NewTimeMeter(frame.METER_LOOP_COUNT)

	for !t.Server.IsQuit() {
		busy = false
		meter.Clear()

		if net_module.Run(frame.NET_LOOP_COUNT) {
			busy = true
		}
		meter.Stamp()

		if http_net_module.Run(frame.HTTP_LOOP_COUNT) {
			busy = true
		}
		meter.Stamp()

		if db_module.Run(frame.DB_LOOP_COUNT) {
			busy = true
		}
		meter.Stamp()

		if timer_module.Update(frame.TIMER_LOOP_COUNT) {
			busy = true
		}
		meter.CheckOut()

		if !busy {
			time.Sleep(1 * time.Millisecond)
		}
	}
}
