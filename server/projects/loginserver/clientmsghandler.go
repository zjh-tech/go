package main

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"net/http"
	"projects/frame"
	"projects/go-engine/edb"
	"projects/go-engine/elog"
	"projects/pb"
	"projects/util"

	"github.com/golang/protobuf/proto"
)

type ClientMsgHandler struct{}

func (h *ClientMsgHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		elog.ErrorAf("ClientMsgHandler Error=%v", err)
		return
	}

	msgID := util.NetBytesToUint32(body)
	msgIDLen := 4
	GHttpsMsgHandler.OnHandler(msgID, body[msgIDLen:], w)
}

type HttpsFunc func(datas []byte, w http.ResponseWriter)

type HttpsMsgHandler struct {
	dealer *frame.IDDealer
}

func (c *HttpsMsgHandler) Init() bool {
	c.dealer.RegisterHandler(uint32(pb.EClient2GameMsgId_cs_account_register_req_id), HttpsFunc(OnHandlerCsAccountRegisterReq))
	c.dealer.RegisterHandler(uint32(pb.EClient2GameMsgId_cs_account_login_req_id), HttpsFunc(OnHandlerCsAccountLoginReq))
	return true
}

func (c *HttpsMsgHandler) OnHandler(msgID uint32, datas []byte, w http.ResponseWriter) {
	dealer := c.dealer.FindHandler(msgID)
	if dealer == nil {
		elog.ErrorAf("LoginServer ClientMsgHandler Can Not Find MsgID = %v", msgID)
		return
	}

	dealer.(HttpsFunc)(datas, w)
}

var GHttpsMsgHandler *HttpsMsgHandler

func init() {
	GHttpsMsgHandler = &HttpsMsgHandler{
		dealer: frame.NewIDDealer(),
	}

	GHttpsMsgHandler.Init()
}

func OnHandlerCsAccountRegisterReq(datas []byte, w http.ResponseWriter) {
	var AckFunc = func(ack *pb.ScAccountRegisterAck, w http.ResponseWriter) {
		w.Write(util.Uint32ToNetBytes(uint32(pb.EClient2GameMsgId_sc_account_register_ack_id)))
		res, _ := json.Marshal(ack)
		w.Write(res)
	}

	ack := &pb.ScAccountRegisterAck{}

	req := &pb.CsAccountRegisterReq{}
	err := proto.Unmarshal(datas, req)
	if err != nil {
		ack.ErrorCode = uint32(pb.EScErrorCode_com_fail)
		AckFunc(ack, w)
		elog.ErrorAf("[LoginSys] CsAccountRegisterReq Protobuf Unmarshal=%v", datas)
		return
	}

	ack.AccountName = req.AccountName
	if req.AccountName == "" {
		ack.ErrorCode = uint32(pb.EScErrorCode_com_fail)
		AckFunc(ack, w)
		return
	}

	const (
		AccountRegisterExist      = 0
		AccountRegisterSelectFail = 1
		AccountRegisterInsertFail = 2
	)

	type CmdParas struct {
		req *pb.CsAccountRegisterReq
	}

	cmdParas := &CmdParas{
		req: req,
	}

	frame.SyncDoSqlOpt(func(conn edb.IMysqlConn, attach []interface{}) (edb.IMysqlRecordSet, int32, error) {
		paras := attach[0].(*CmdParas)
		uid := util.Hash64(paras.req.AccountName)
		tableName := edb.GDBModule.GetTableNameByUID("accountverify", uid)
		select_sql := frame.BuildSelectSQL(tableName, []string{
			"accountid",
			"accountname",
			"password",
		}, map[string]interface{}{
			"accountname": req.AccountName,
		})

		result, selectErr := conn.QueryWithResult(select_sql)
		if selectErr != nil {
			return nil, AccountRegisterSelectFail, selectErr
		}

		rc := result.GetRecordSet()
		if len(rc) >= 1 {
			return nil, AccountRegisterExist, nil
		}

		accountId := frame.GIdMaker.NextId()
		insert_sql := frame.BuildInsertSQL(tableName, map[string]interface{}{
			"accountid":   accountId,
			"accountname": req.AccountName,
			"password":    req.Password,
			"createtime":  util.GetSecond(),
		})

		_, insertErr := conn.QueryWithoutResult(insert_sql)
		if insertErr != nil {
			return nil, AccountRegisterInsertFail, insertErr
		}
		return nil, edb.DB_EXEC_SUCCESS, nil
	}, func(recordSet edb.IMysqlRecordSet, attach []interface{}, errorCode int32, err error) {
		paras := attach[0].(*CmdParas)
		if errorCode == AccountRegisterSelectFail {
			elog.ErrorAf("[Login] AccountRegister AccountName=%v Select Error=%v", paras.req.AccountName, err)
			ack.ErrorCode = uint32(pb.EScErrorCode_com_fail)
			AckFunc(ack, w)
			return
		}

		if errorCode == AccountRegisterInsertFail {
			elog.ErrorAf("[Login] AccountRegister AccountName=%v Insert Error=%v", paras.req.AccountName, err)
			ack.ErrorCode = uint32(pb.EScErrorCode_com_fail)
			AckFunc(ack, w)
			return
		}

		if errorCode == AccountRegisterExist {
			elog.ErrorAf("[Login] AccountRegister AccountName=%v Exist", paras.req.AccountName)
			ack.ErrorCode = uint32(pb.EScErrorCode_com_accountregister_exist)
			AckFunc(ack, w)
			return
		}

		if errorCode == edb.DB_EXEC_SUCCESS {
			elog.InfoAf("[Login] AccountRegister AccountName=%v Success", paras.req.AccountName)
			ack.ErrorCode = uint32(pb.EScErrorCode_com_success)
			AckFunc(ack, w)
			return
		}
	}, []interface{}{cmdParas}, util.Hash64(req.AccountName))
}

func BuildMinGateWay(ack *pb.ScAccountLoginAck) {
	gatewayMap := frame.GSSServerSessionMgr.FindLogicServerByServerType(frame.GATEWAY_SERVER_TYPE)
	minCount := uint64(math.MaxUint64)
	for _, logicserver := range gatewayMap {
		gateway := logicserver.(*GatewayServer)
		if gateway.PlayerTotalCount < minCount {
			minCount = gateway.PlayerTotalCount
			ack.GatewayIp = gateway.Ip
			ack.GatewayPort = gateway.Port
		}
	}
}

func OnHandlerCsAccountLoginReq(datas []byte, w http.ResponseWriter) {
	var AckFunc = func(ack *pb.ScAccountLoginAck, w http.ResponseWriter) {
		w.Write(util.Uint32ToNetBytes(uint32(pb.EClient2GameMsgId_sc_account_login_ack_id)))
		res, _ := json.Marshal(ack)
		w.Write(res)
	}

	ack := &pb.ScAccountLoginAck{}

	req := &pb.CsAccountLoginReq{}
	err := proto.Unmarshal(datas, req)
	if err != nil {
		ack.ErrorCode = uint32(pb.EScErrorCode_com_fail)
		AckFunc(ack, w)
		elog.ErrorAf("[LoginSys] CsAccountLoginReq Json Unmarshal=%v", datas)
		return
	}

	ack.AccountName = req.AccountName
	if req.AccountName == "" {
		ack.ErrorCode = uint32(pb.EScErrorCode_com_fail)
		AckFunc(ack, w)
		return
	}

	type CmdParas struct {
		req *pb.CsAccountLoginReq
	}

	const (
		AccountLoginExist    = 0
		AccountLoginNonExist = 1
		AccountLoginFail     = 2
	)

	cmdParas := &CmdParas{
		req: req,
	}

	frame.SyncDoSqlOpt(func(conn edb.IMysqlConn, attach []interface{}) (edb.IMysqlRecordSet, int32, error) {
		paras := attach[0].(*CmdParas)
		uid := util.Hash64(req.AccountName)
		tableName := edb.GDBModule.GetTableNameByUID("accountverify", uid)
		select_sql := frame.BuildSelectSQL(tableName, []string{
			"accountid",
			"accountname",
			"password",
		}, map[string]interface{}{
			"accountname": paras.req.AccountName,
		})

		result, selectErr := conn.QueryWithResult(select_sql)
		if selectErr != nil {
			return nil, AccountLoginFail, selectErr
		}

		rc := result.GetRecordSet()
		if len(rc) <= 0 {
			return nil, AccountLoginNonExist, nil
		} else {
			return result, AccountLoginExist, nil
		}

	}, func(recordSet edb.IMysqlRecordSet, attach []interface{}, errorCode int32, err error) {
		paras := attach[0].(*CmdParas)

		if errorCode == AccountLoginNonExist {
			elog.ErrorAf("[Login] AccountVerify AccountName=%v Not Exist", paras.req.AccountName)
			ack.ErrorCode = uint32(pb.EScErrorCode_com_accountlogin_not_exist)
			AckFunc(ack, w)
			return
		}

		if errorCode == AccountLoginFail {
			elog.ErrorAf("[Login] AccountVerify AccountName=%v Error=%v", paras.req.AccountName, err)
			ack.ErrorCode = uint32(pb.EScErrorCode_com_fail)
			AckFunc(ack, w)
			return
		}

		if errorCode == AccountLoginExist {
			rc := recordSet.GetRecordSet()
			if len(rc) != 1 {
				elog.ErrorAf("[Login] AccountVerify AccountName=%v RecordCount=%v Error", paras.req.AccountName, len(rc))
				return
			}

			accountId, _ := util.Str2Uint64(rc[0]["accountid"])
			password := rc[0]["password"]
			password = paras.req.Password

			if password != paras.req.Password {
				elog.ErrorAf("[Login] AccountVerify AccountName=%v Password Error", paras.req.AccountName)
				ack.ErrorCode = uint32(pb.EScErrorCode_com_accountlogin_passwd_err)
				AckFunc(ack, w)
				return
			}

			elog.InfoAf("[Login] AccountVerify AccountName=%v Success", paras.req.AccountName)

			ack.ErrorCode = uint32(pb.EScErrorCode_com_success)
			ack.Accountid = accountId
			BuildMinGateWay(ack)
			ack.Token = GTokenMgr.GenerateToken(ack.Accountid)
			ack.LoginServerId = frame.GServer.GetLocalServerID()
			AckFunc(ack, w)
		}
	}, []interface{}{cmdParas}, util.Hash64(req.AccountName))
}
