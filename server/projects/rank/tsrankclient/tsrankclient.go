package tsrankclient

import (
	"projects/frame"
	"projects/go-engine/elog"
	"projects/go-engine/etimer"
	"projects/pb"
	"projects/util"

	"github.com/golang/protobuf/proto"
)

//RankCbFunc args[0]=Tid args[1]=datas args[2]=cbArgs
type RankCbFunc func(...interface{})

type RankCbArgType = []interface{}

type RankCbItem struct {
	Tid   uint32
	MsgID uint32
	Msg   proto.Message
	Func  RankCbFunc
	Agrs  RankCbArgType
	Tick  int64
}

func NewRankCbItem(cb RankCbFunc, args RankCbArgType, tid uint32, msgID uint32, msg proto.Message) *RankCbItem {
	return &RankCbItem{
		Func:  cb,
		Agrs:  args,
		Tick:  util.GetMillsecond(),
		Tid:   tid,
		MsgID: msgID,
		Msg:   msg,
	}
}

const (
	RANK_CLIENT_CALLBACK_TIMEOUT_TIMER_ID uint32 = 1
	RANK_CLIENT_RECONNECT_TIME_ID         uint32 = 2
)

const (
	RANK_CLIENT_CALLBACK_TIMEOUT_TIMER_DELAY uint64 = 1000 * 60 * 10
	RANK_CLIENT_RECONNECT_TIME_DELAY         uint64 = 1000 * 10
)

type TsRankClient struct {
	cbNextId          uint64
	cbMap             map[uint64]*RankCbItem
	timerRegister     etimer.ITimerRegister
	tsBalanceAddr     string
	tsBalanceToken    string
	tsGatewayToken    string
	tsGatewayVerifyOk bool
	tsBalanceSessID   uint64
	tsGatewaySessID   uint64
}

func NewTsRankClient() *TsRankClient {
	return &TsRankClient{
		cbNextId:      1,
		cbMap:         make(map[uint64]*RankCbItem),
		timerRegister: etimer.NewTimerRegister(),
	}
}

func (t *TsRankClient) GettsBalanceAddr() string {
	return t.tsBalanceAddr
}

func (t *TsRankClient) GettsBalanceToken() string {
	return t.tsBalanceToken
}

func (t *TsRankClient) Reset() {
	t.tsGatewayToken = ""
	t.tsGatewayVerifyOk = false
	t.tsBalanceSessID = 0
	t.tsGatewaySessID = 0
}

//-----------------------------------------------------------------------------------------------
func (t *TsRankClient) Init(remoteAddr string, tsBalanceToken string) bool {
	t.tsBalanceAddr = remoteAddr
	t.tsBalanceToken = tsBalanceToken
	t.tsBalanceSessID = frame.GSSClientSessionMgr.SSClientConnect(remoteAddr, GRankBalanceMsgHandler, nil)

	t.timerRegister.AddRepeatTimer(RANK_CLIENT_CALLBACK_TIMEOUT_TIMER_ID, RANK_CLIENT_CALLBACK_TIMEOUT_TIMER_DELAY, "TsRankClient-CallBackTimeOut", func(v ...interface{}) {
		now := util.GetMillsecond()
		for cbId, cbItem := range t.cbMap {
			if cbItem.Tick+int64(RANK_CLIENT_CALLBACK_TIMEOUT_TIMER_DELAY) < now {
				elog.WarnAf("[TsRankClient] TimeOut Tid=%v, Msg=%+v", cbItem.Tid, cbItem.Msg)
				delete(t.cbMap, cbId)
			}
		}
	}, []interface{}{}, true)

	t.timerRegister.AddRepeatTimer(RANK_CLIENT_RECONNECT_TIME_ID, RANK_CLIENT_RECONNECT_TIME_DELAY, "TsRankClient-ReConnect", func(v ...interface{}) {
		if (frame.GSSClientSessionMgr.IsInConnectCache(t.tsBalanceSessID) == false) && (frame.GSSClientSessionMgr.IsExistSessionOfSessID(t.tsBalanceSessID) == false) && (frame.GSSClientSessionMgr.IsInConnectCache(t.tsGatewaySessID) == false) && (frame.GSSClientSessionMgr.IsExistSessionOfSessID(t.tsGatewaySessID) == false) {
			//重连
			t.Reset()
			elog.InfoAf("[TsRankClient] Reconnect Remote Addr=%v", t.tsBalanceAddr)
			t.tsBalanceSessID = frame.GSSClientSessionMgr.SSClientConnect(remoteAddr, GRankBalanceMsgHandler, nil)
		}
	}, []interface{}{}, true)

	return true
}

//pb.Ts2CUpdateRankAck
func (t *TsRankClient) SendUpdateRankReq(tid uint32, rankItem *pb.RankItem, cb RankCbFunc, args RankCbArgType) {
	req := pb.C2TsUpdateRankReq{
		RankInfo: rankItem,
	}
	t.sendRankMsg(tid, uint32(pb.C2TSLogicMsgId_c2ts_update_rank_req_id), &req, cb, args)
}

//pb.Ts2CQueryRankAck
func (t *TsRankClient) SendQueryRankReq(tid uint32, playerid uint64, topn uint32, cb RankCbFunc, args RankCbArgType) {
	req := pb.C2TsQueryRankReq{
		PlayerId: playerid,
		Topn:     topn,
	}
	t.sendRankMsg(tid, uint32(pb.C2TSLogicMsgId_c2ts_query_rank_req_id), &req, cb, args)
}

//pb.Ts2CClearAllAck
func (t *TsRankClient) SendClearAllReq(tid uint32, cb RankCbFunc, args RankCbArgType) {
	req := pb.C2TsClearAllReq{}
	t.sendRankMsg(tid, uint32(pb.C2TSLogicMsgId_c2ts_clear_all_req_id), &req, cb, args)
}

//pb.Ts2CClearPlayerAck
func (t *TsRankClient) SendClearPlayerReq(tid uint32, PlayerId uint64, cb RankCbFunc, args RankCbArgType) {
	req := pb.C2TsClearPlayerReq{
		PlayerId: PlayerId,
	}
	t.sendRankMsg(tid, uint32(pb.C2TSLogicMsgId_c2ts_clear_player_req_id), &req, cb, args)
}

//-----------------------------------------------------------------------------------------------

func (t *TsRankClient) GetCbMap() map[uint64]*RankCbItem {
	return t.cbMap
}

func (t *TsRankClient) sendRankMsg(tid uint32, msgID uint32, msg proto.Message, cb RankCbFunc, args RankCbArgType) bool {
	datas, err := proto.Marshal(msg)
	if err != nil {
		elog.ErrorAf("[Net] Msg=%v Marshal Err %v ", msgID, err)
		return false
	}

	cg_tranfer := pb.C2TsCgTranferMsg{}
	cg_tranfer.Tid = tid
	cg_tranfer.Msgid = msgID
	cg_tranfer.Datas = datas

	if GRankGatewaySession == nil {
		elog.ErrorA("[TsRankClient] Rank Gateway Not Find ClientSession")
		return false
	}

	if t.tsGatewayVerifyOk == false {
		return false
	}

	cbItem := NewRankCbItem(cb, args, tid, msgID, msg)
	t.cbNextId++
	t.cbMap[t.cbNextId] = cbItem
	cg_tranfer.Cbid = t.cbNextId
	GRankGatewaySession.SendProtoMsg(uint32(pb.C2TSLogicMsgId_c2ts_cg_tranfer_msg_id), &cg_tranfer, nil)

	return true
}

var GTsRankClient *TsRankClient

func init() {
	GTsRankClient = NewTsRankClient()
}
