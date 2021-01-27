package main

import (
	"projects/frame"
	"projects/go-engine/edb"
	"projects/go-engine/elog"
	"projects/pb"
	"projects/util"

	"github.com/golang/protobuf/proto"
)

type HallFunc func(datas []byte, g *HallServer) bool

type HallServer struct {
	frame.LogicServer
	dealer *frame.IDDealer
}

func NewHallServer() *HallServer {
	hall := &HallServer{
		dealer: frame.NewIDDealer(),
	}
	hall.Init()
	return hall
}

func (h *HallServer) Init() bool {
	h.dealer.RegisterHandler(uint32(pb.S2SLogicMsgId_hc_select_player_list_req_id), HallFunc(OnHandlerGcSelectPlayerListReq))
	h.dealer.RegisterHandler(uint32(pb.S2SLogicMsgId_hc_create_player_req_id), HallFunc(OnHandlerHcCreatePlayerReq))
	return true
}

func (h *HallServer) OnHandler(msgID uint32, attach_datas []byte, datas []byte, sess *frame.SSServerSession) {
	defer func() {
		if err := recover(); err != nil {
			elog.ErrorAf("HallServer onHandler MsgID = %v Error=%v", msgID, err)
		}
	}()

	dealer := h.dealer.FindHandler(msgID)
	if dealer == nil {
		elog.ErrorAf("HallServer MsgHandler Can Not Find MsgID = %v", msgID)
		return
	}

	dealer.(HallFunc)(datas, h)
}

func (h *HallServer) OnEstablish(serversess *frame.SSServerSession) {
	elog.InfoAf("HallServer OnEstablish Remote [ID=%v,Type=%v,Ip=%v] ", serversess.GetRemoteServerID(), serversess.GetRemoteServerType(), serversess.GetRemoteOuter())
}

func (h *HallServer) OnTerminate(serversess *frame.SSServerSession) {
	elog.InfoAf("HallServer OnTerminate Remote [ID=%v,Type=%v,Ip=%v] ", serversess.GetRemoteServerID(), serversess.GetRemoteServerType(), serversess.GetRemoteOuter())
}

func OnHandlerGcSelectPlayerListReq(datas []byte, h *HallServer) bool {
	req := pb.HcSelectPlayerListReq{}
	unmarshalErr := proto.Unmarshal(datas, &req)
	if unmarshalErr != nil {
		return false
	}

	type CmdParas struct {
		req      pb.HcSelectPlayerListReq
		nameId   uint64
		PlayerId uint64
	}

	cmdParas := &CmdParas{
		req: req,
	}

	frame.AsyncDoSqlOpt(func(conn edb.IMysqlConn, attach []interface{}) (edb.IMysqlRecordSet, int32, error) {
		paras := attach[0].(*CmdParas)
		TableName := edb.GDBModule.GetTableNameByUID("Playerbase", paras.req.Accountid)
		select_sql := frame.BuildSelectSQL(TableName, []string{
			"Playerid",
			"Playername",
		}, map[string]interface{}{
			"accountid": paras.req.Accountid,
		})
		result, selectErr := conn.QueryWithResult(select_sql)
		if selectErr != nil {
			return nil, edb.DB_EXEC_FAIL, selectErr
		}
		return result, edb.DB_EXEC_SUCCESS, nil

	}, func(recordSet edb.IMysqlRecordSet, attach []interface{}, errorCode int32, err error) {
		paras := attach[0].(*CmdParas)
		var AckFunc = func(req pb.HcSelectPlayerListReq, baselist []*pb.SsPlayerBaseInfo, errorCode uint32) {
			ack := pb.ChSelectPlayerListAck{}
			ack.Accountid = req.Accountid
			ack.Playeruid = req.Playeruid
			ack.Baseinfolist = baselist
			h.GetServerSession().SendProtoMsg(uint32(pb.S2SLogicMsgId_ch_select_player_list_ack_id), &ack, nil)
		}

		baselist := make([]*pb.SsPlayerBaseInfo, 0)
		if errorCode == edb.DB_EXEC_FAIL {
			elog.ErrorAf("[SelectPlayerList] Error=%v", err)
			AckFunc(paras.req, baselist, frame.MSG_FAIL)
			return
		}

		if errorCode != edb.DB_EXEC_SUCCESS {
			return
		}

		rc := recordSet.GetRecordSet()
		for _, record := range rc {
			Playerbase := &pb.SsPlayerBaseInfo{}
			Playerbase.Playerid, _ = util.Str2Uint64(record["playerid"])
			Playerbase.Playername = record["playername"]
			baselist = append(baselist, Playerbase)
		}
		AckFunc(paras.req, baselist, frame.MSG_SUCCESS)
	}, []interface{}{cmdParas}, cmdParas.req.Accountid)

	return true
}

func OnHandlerHcCreatePlayerReq(datas []byte, h *HallServer) bool {
	// 以AccountId为Hash做基准 ==> 根据AccountId得到所有玩家列表
	req := pb.HcCreatePlayerReq{}
	unmarshalErr := proto.Unmarshal(datas, &req)
	if unmarshalErr != nil {
		return false
	}

	const (
		PlayerCreateSuccess     = 0
		PlayerCreateNameExist   = 1
		PlayerCreateUnknowError = 2
	)

	type CmdParas struct {
		req      pb.HcCreatePlayerReq
		nameId   uint64
		PlayerId uint64
	}

	cmdParas := &CmdParas{
		req: req,
	}

	var AckFunc = func(req pb.HcCreatePlayerReq, errorCode uint32) {
		ack := &pb.ChCreatePlayerAck{}
		ack.Accountid = req.Accountid
		ack.Playername = req.Playername
		ack.Playerid = req.Playerid
		ack.Errorcode = errorCode
		h.GetServerSession().SendProtoMsg(uint32(pb.EClient2GameMsgId_sc_create_player_ack_id), ack, nil)
	}

	//PlayerName Hash Conn
	frame.AsyncDoSqlOpt(func(conn edb.IMysqlConn, attach []interface{}) (edb.IMysqlRecordSet, int32, error) {
		paras := attach[0].(*CmdParas)
		//检查名字是否唯一
		verifyTableName := edb.GDBModule.GetTableNameByUID("Playernameverify", util.Hash64(paras.req.Playername))
		select_verify_sql := frame.BuildSelectSQL(verifyTableName, []string{
			"playernameid",
		}, map[string]interface{}{
			"playername": paras.req.Playername,
		})
		result, selectErr := conn.QueryWithResult(select_verify_sql)
		if selectErr != nil {
			return nil, PlayerCreateUnknowError, selectErr
		}
		rc := result.GetRecordSet()
		if len(rc) >= 1 {
			return nil, PlayerCreateNameExist, nil
		}

		//插入名字
		insert_verify_sql := frame.BuildInsertSQL(verifyTableName, map[string]interface{}{
			"Playernameid": paras.req.Playerid,
			"Playername":   paras.req.Playername,
			"accountid":    req.Accountid,
		})
		result2, insertErr := conn.QueryWithoutResult(insert_verify_sql)
		if insertErr != nil {
			return nil, PlayerCreateUnknowError, insertErr
		}
		paras.nameId = uint64(result2.GetInsertId())
		return nil, edb.DB_EXEC_SUCCESS, nil

	}, func(recordSet edb.IMysqlRecordSet, attach []interface{}, errorCode int32, err error) {
		paras := attach[0].(*CmdParas)
		if errorCode == PlayerCreateUnknowError {
			elog.ErrorAf("[CreatePlayer] AccountId=%v,PlayerName=%v PlayerCreateSelectError", paras.req.Accountid, paras.req.Playername)
			AckFunc(paras.req, PlayerCreateUnknowError)
			return
		}
		if errorCode == PlayerCreateNameExist {
			elog.ErrorAf("[CreatePlayer] AccountId=%v,PlayerName=%v PlayerCreateNameExist", paras.req.Accountid, paras.req.Playername)
			AckFunc(paras.req, PlayerCreateNameExist)
			return
		}
		if errorCode != edb.DB_EXEC_SUCCESS {
			return
		}

		//AccountId Hash Conn
		frame.AsyncDoSqlOpt(func(conn edb.IMysqlConn, attach []interface{}) (edb.IMysqlRecordSet, int32, error) {
			PlayerbaseParas := attach[0].(*CmdParas)
			//插入角色
			PlayerTableName := edb.GDBModule.GetTableNameByUID("playerbase", PlayerbaseParas.req.Accountid)
			insert_Player_sql := frame.BuildInsertSQL(PlayerTableName, map[string]interface{}{
				"playerid":   frame.GIdMaker.NextId(),
				"playername": PlayerbaseParas.req.Playername,
				"accountid":  req.Accountid,
				"createtime": util.GetSecond(),
			})
			result3, insertPlayerErr := conn.QueryWithoutResult(insert_Player_sql)
			if insertPlayerErr != nil {
				return nil, PlayerCreateUnknowError, insertPlayerErr
			}
			PlayerbaseParas.PlayerId = uint64(result3.GetInsertId())
			elog.InfoAf("[CreatePlayer] AccountId=%v,PlayerName=%v,NameId=%v,PlayerId=%v",
				PlayerbaseParas.req.Accountid, PlayerbaseParas.req.Playername, PlayerbaseParas.nameId, PlayerbaseParas.PlayerId)
			return nil, edb.DB_EXEC_SUCCESS, nil

		}, func(recordSet edb.IMysqlRecordSet, attach []interface{}, errorCode int32, err error) {
			PlayerbaseParas := attach[0].(*CmdParas)
			if errorCode == PlayerCreateUnknowError {
				elog.ErrorAf("[CreatePlayer] AccountId=%v,PlayerName=%v PlayerCreateUnknowError", paras.req.Accountid, paras.req.Playername)
				AckFunc(PlayerbaseParas.req, PlayerCreateUnknowError)
				return
			}
			if errorCode != edb.DB_EXEC_SUCCESS {
				return
			}
			AckFunc(PlayerbaseParas.req, PlayerCreateSuccess)
		}, []interface{}{paras}, paras.req.Accountid)
	}, []interface{}{cmdParas}, util.Hash64(req.Playername))

	return true
}
