package main

import (
	"fmt"
	"time"

	"github.com/zjh-tech/go-frame/base/util"
	"github.com/zjh-tech/go-frame/engine/edb"
	"github.com/zjh-tech/go-frame/engine/elog"
	"github.com/zjh-tech/go-frame/engine/etimer"
	"github.com/zjh-tech/go-frame/frame"
)

//CREATE TABLE `account_00` (
//`accountid` bigint(20) unsigned COMMENT '账号ID',
//`username` varchar(128) NOT NULL DEFAULT '' COMMENT '账号',
//`password` varchar(128) NOT NULL DEFAULT '' COMMENT '密码',
//PRIMARY KEY (`accountid`),
//KEY (`username`)
//) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

type Server struct {
}

func (s *Server) Init() bool {
	logger := elog.NewLogger("./log", 0)
	logger.Init()
	ELog.SetLogger(logger)

	//Uid
	idMaker, idErr := frame.NewIdMaker(int64(1))
	if idErr != nil {
		ELog.Errorf("Server IdMaker Error=%v", idMaker)
		return false
	}
	frame.GIdMaker = idMaker

	ELog.Info("Server Log System Init Success")

	// if err := edb.GDatabaseCfgMgr.Load("./db_cfg.xml"); err != nil {
	// 	ELog.Error(err)
	// 	return false
	// }

	// if err := edb.GDBModule.Init(edb.GDatabaseCfgMgr.DBConnMaxCount, edb.GDatabaseCfgMgr.DBTableMaxCount, edb.GDatabaseCfgMgr.DBConnSpecs); err != nil {
	// 	ELog.Error(err)
	// 	return false
	// }

	ELog.Info("Server Init Success")

	TestSync()

	return true
}

func (s *Server) Run() {
	busy := false
	timer_module := etimer.GTimerMgr
	db_module := edb.GDBModule
	meter := frame.NewTimeMeter(frame.MeterLoopCount)

	for {
		busy = false
		meter.Clear()

		if db_module.Run(frame.DBLoopCount) {
			busy = true
		}
		meter.Stamp()

		if timer_module.Update(frame.TimerLoopCount) {
			busy = true
		}
		meter.CheckOut()

		if !busy {
			time.Sleep(1 * time.Millisecond)
		}
	}
}

func TestSync() {
	type CmdParas struct {
		UserName string
		Password string
	}

	start_index := 0
	end_index := 200000
	start_tick := util.GetMillsecond()
	qps_count := 0
	for i := start_index; i < end_index; i++ {
		user_name := fmt.Sprintf("Test%v", i)

		cmd_paras := &CmdParas{
			UserName: user_name,
			Password: "123456",
		}

		//Insert
		edb.SyncDoSqlOpt(func(conn edb.IMysqlConn, attach []interface{}) (edb.IMysqlRecordSet, int32, error) {
			paras := attach[0].(*CmdParas)
			uid := util.Hash64(paras.UserName)
			tableName := edb.GDBModule.GetTableNameByUID("account", uid)
			accountId := frame.GIdMaker.NextId()
			insert_sql := edb.BuildInsertSQL(tableName, map[string]interface{}{
				"accountid": accountId,
				"username":  paras.UserName,
				"password":  paras.Password,
			})

			_, insertErr := conn.QueryWithoutResult(insert_sql)
			if insertErr != nil {
				return nil, edb.DbExecFail, insertErr
			}
			return nil, edb.DbExecSuccess, nil
		}, func(recordSet edb.IMysqlRecordSet, attach []interface{}, errorCode int32, err error) {
			qps_count++
			end_tick := util.GetMillsecond()
			if (end_tick - start_tick) >= 1000 {
				ELog.InfoAf("Insert Qps=%v", qps_count)
				qps_count = 0
				start_tick = end_tick
			}
		}, []interface{}{cmd_paras}, util.Hash64(user_name))

	}

	qps_count = 0

	for i := start_index; i < end_index; i++ {
		user_name := fmt.Sprintf("Test%v", i)

		cmd_paras := &CmdParas{
			UserName: user_name,
			Password: "123456",
		}

		//Select
		edb.SyncDoSqlOpt(func(conn edb.IMysqlConn, attach []interface{}) (edb.IMysqlRecordSet, int32, error) {
			paras := attach[0].(*CmdParas)
			uid := util.Hash64(paras.UserName)
			tableName := edb.GDBModule.GetTableNameByUID("account", uid)
			select_sql := edb.BuildSelectSQL(tableName, []string{
				"accountid",
				"username",
				"password",
			}, map[string]interface{}{
				"username": paras.UserName,
			})

			result, selectErr := conn.QueryWithResult(select_sql)
			if selectErr != nil {
				return nil, edb.DbExecFail, selectErr
			}

			rc := result.GetRecordSet()
			if len(rc) >= 1 {
				return nil, edb.DbExecFail, nil
			}

			return nil, edb.DbExecSuccess, nil
		}, func(recordSet edb.IMysqlRecordSet, attach []interface{}, errorCode int32, err error) {
			qps_count++
			end_tick := util.GetMillsecond()
			if (end_tick - start_tick) >= 1000 {
				ELog.InfoAf("Select Qps=%v", qps_count)
				qps_count = 0
				start_tick = end_tick
			}
		}, []interface{}{cmd_paras}, util.Hash64(user_name))
	}
}

func main() {
	var server Server
	if server.Init() {
		server.Run()
	}
}
