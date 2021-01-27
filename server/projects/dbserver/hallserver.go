package main

import (
	"projects/frame"
	"projects/go-engine/edb"
	"projects/go-engine/elog"
	"projects/pb"
	"projects/util"

	"github.com/golang/protobuf/proto"
)

type HallFunc func(datas []byte, h *HallServer) bool

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
	h.dealer.RegisterHandler(uint32(pb.S2SLogicMsgId_hd_create_player_req_id), HallFunc(OnHandlerHdCreatePlayerReq))
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

func OnHandlerHdCreatePlayerReq(datas []byte, h *HallServer) bool {
	hallReq := pb.HdCreatePlayerReq{}
	err := proto.Unmarshal(datas, &hallReq)
	if err != nil {
		return false
	}

	type CmdParas struct {
		req pb.HdCreatePlayerReq
	}

	cmdParas := &CmdParas{
		req: hallReq,
	}

	frame.AsyncDoSqlOpt(func(conn edb.IMysqlConn, attach []interface{}) (edb.IMysqlRecordSet, int32, error) {
		paras := attach[0].(*CmdParas)
		tableName := edb.GDBModule.GetTableNameByUID("playerinfo", paras.req.Accountid)
		insert_sql := frame.BuildInsertSQL(tableName, map[string]interface{}{
			"accountid":  paras.req.Accountid,
			"playerid":   paras.req.Playerid,
			"playername": paras.req.Playername,
			"createtime": util.GetSecond(),
		})

		_, insertErr := conn.QueryWithoutResult(insert_sql)
		if insertErr != nil {
			return nil, edb.DB_EXEC_FAIL, insertErr
		}
		return nil, edb.DB_EXEC_SUCCESS, nil

	}, func(recordSet edb.IMysqlRecordSet, attach []interface{}, errorCode int32, err error) {
		paras := attach[0].(*CmdParas)
		var AckFunc = func() {
			ack := &pb.DhCreatePlayerAck{}
			ack.Playerid = paras.req.Playerid
			h.GetServerSession().SendProtoMsg(uint32(pb.S2SLogicMsgId_dh_create_player_ack_id), ack, nil)
		}

		if errorCode == edb.DB_EXEC_SUCCESS {
			AckFunc()
			return
		}
	}, []interface{}{cmdParas}, hallReq.Accountid)
	return true
}
