package sdk

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

type RankClient struct {
	gatewayAddr  string
	gatewayToken string

	gatewayVerifyOk bool
	gatewaySessID   uint64

	cbNextId      uint64
	cbMap         map[uint64]*RankCbItem
	timerRegister etimer.ITimerRegister

	handler *RankClientMsgHandler
	sess    *frame.SSClientSession
}

func NewRankClient() *RankClient {
	rank_client := &RankClient{
		cbNextId:      1,
		cbMap:         make(map[uint64]*RankCbItem),
		timerRegister: etimer.NewTimerRegister(),
		handler:       NewRankClientMsgHandler(),
		sess:          nil,
	}

	return rank_client
}

func (r *RankClient) Reset() {
	r.gatewayVerifyOk = false
	r.gatewaySessID = 0
}

func (r *RankClient) SetSession(sess *frame.SSClientSession) {
	r.sess = sess
}

//-----------------------------------------------------------------------------------------------
func (r *RankClient) Init(remoteAddr string, token string) bool {
	r.gatewayAddr = remoteAddr
	r.gatewayToken = token

	r.timerRegister.AddRepeatTimer(RANK_CLIENT_CALLBACK_TIMEOUT_TIMER_ID, RANK_CLIENT_CALLBACK_TIMEOUT_TIMER_DELAY, "RankClient-CallBackTimeOut", func(v ...interface{}) {
		now := util.GetMillsecond()
		for cbId, cbItem := range r.cbMap {
			if cbItem.Tick+int64(RANK_CLIENT_CALLBACK_TIMEOUT_TIMER_DELAY) < now {
				elog.WarnAf("[RankClient] TimeOut Tid=%v, Msg=%+v", cbItem.Tid, cbItem.Msg)
				delete(r.cbMap, cbId)
			}
		}
	}, []interface{}{}, true)

	r.timerRegister.AddRepeatTimer(RANK_CLIENT_RECONNECT_TIME_ID, RANK_CLIENT_RECONNECT_TIME_DELAY, "RankClient-ReConnect", func(v ...interface{}) {
		if (frame.GSSClientSessionMgr.IsInConnectCache(r.gatewaySessID) == false) && (frame.GSSClientSessionMgr.IsExistSessionOfSessID(r.gatewaySessID) == false) {
			//重连
			r.Reset()
			elog.InfoAf("[RankClient] Reconnect Remote Addr=%v", r.gatewayAddr)
			r.gatewaySessID = frame.GSSClientSessionMgr.SSClientConnect(r.gatewayAddr, r.handler, nil)
		}
	}, []interface{}{}, true)

	return true
}

//pb.Ts2CUpdateRankAck
func (r *RankClient) SendUpdateRankReq(tid uint32, rankItem *pb.RankItem, cb RankCbFunc, args RankCbArgType) {
	req := pb.C2RUpdateRankReq{
		RankInfo: rankItem,
	}
	r.sendRankMsg(tid, uint32(pb.CRLogicMsgId_c2r_update_rank_req_id), &req, cb, args)
}

//pb.Ts2CQueryRankAck
func (r *RankClient) SendQueryRankReq(tid uint32, playerid uint64, topn uint32, cb RankCbFunc, args RankCbArgType) {
	req := pb.C2RQueryRankReq{
		PlayerId: playerid,
		Topn:     topn,
	}
	r.sendRankMsg(tid, uint32(pb.CRLogicMsgId_c2r_query_rank_req_id), &req, cb, args)
}

//pb.Ts2CClearAllAck
func (r *RankClient) SendClearAllReq(tid uint32, cb RankCbFunc, args RankCbArgType) {
	req := pb.C2RClearAllReq{}
	r.sendRankMsg(tid, uint32(pb.CRLogicMsgId_c2r_clear_all_req_id), &req, cb, args)
}

//pb.Ts2CClearPlayerAck
func (r *RankClient) SendClearPlayerReq(tid uint32, PlayerId uint64, cb RankCbFunc, args RankCbArgType) {
	req := pb.C2RClearPlayerReq{
		PlayerId: PlayerId,
	}
	r.sendRankMsg(tid, uint32(pb.CRLogicMsgId_c2r_clear_player_req_id), &req, cb, args)
}

//-----------------------------------------------------------------------------------------------

func (r *RankClient) GetCbMap() map[uint64]*RankCbItem {
	return r.cbMap
}

func (r *RankClient) sendRankMsg(tid uint32, msgID uint32, msg proto.Message, cb RankCbFunc, args RankCbArgType) bool {
	datas, err := proto.Marshal(msg)
	if err != nil {
		elog.ErrorAf("[Net] Msg=%v Marshal Err %v ", msgID, err)
		return false
	}

	cg_tranfer := pb.C2RTranferMsg{}
	cg_tranfer.Tid = tid
	cg_tranfer.Msgid = msgID
	cg_tranfer.Datas = datas

	if r.sess == nil {
		elog.ErrorA("[RankClient]  Not Find ClientSession")
		return false
	}

	if r.gatewayVerifyOk == false {
		return false
	}

	cbItem := NewRankCbItem(cb, args, tid, msgID, msg)
	r.cbNextId++
	r.cbMap[r.cbNextId] = cbItem
	cg_tranfer.Cbid = r.cbNextId
	r.sess.SendProtoMsg(uint32(pb.CRLogicMsgId_c2r_tranfer_msg_id), &cg_tranfer, nil)

	return true
}

type RankClientFunc func(datas []byte, sess *frame.SSClientSession) bool

type RankClientMsgHandler struct {
	dealer        *frame.IDDealer
	timerRegister etimer.ITimerRegister
}

func NewRankClientMsgHandler() *RankClientMsgHandler {
	handler := &RankClientMsgHandler{
		dealer:        frame.NewIDDealer(),
		timerRegister: etimer.NewTimerRegister(),
	}
	handler.Init()
	return handler
}

func (r *RankClientMsgHandler) Init() bool {
	r.dealer.RegisterHandler(uint32(pb.CRLogicMsgId_r2c_rank_verify_ack_id), RankClientFunc(OnHandlerR2CRankVerifyAck))
	r.dealer.RegisterHandler(uint32(pb.CRLogicMsgId_r2c_tranfer_msg_id), RankClientFunc(OnHandlerR2CTranferMsg))
	return true
}

func (r *RankClientMsgHandler) OnHandler(msgID uint32, datas []byte, sess *frame.SSClientSession) {
	defer func() {
		if err := recover(); err != nil {
			elog.ErrorAf("RankClientMsgHandler onHandler MsgID = %v Error=%v", msgID, err)
		}
	}()

	dealer := r.dealer.FindHandler(msgID)
	if dealer == nil {
		elog.ErrorAf("RankClientMsgHandler MsgHandler Can Not Find MsgID = %v", msgID)
		return
	}

	dealer.(RankClientFunc)(datas, sess)
}

func (r *RankClientMsgHandler) OnConnect(sess *frame.SSClientSession) {
	elog.InfoAf("[RankClientMsgHandler] SessId=%v OnConnect", sess.GetSessID())
	GRankClient.SetSession(sess)

	req := pb.C2RRankVerifyReq{}
	req.GatewayToken = GRankClient.gatewayToken
	sess.SendProtoMsg(uint32(pb.CRLogicMsgId_c2r_rank_verify_req_id), &req, nil)
}

func (r *RankClientMsgHandler) OnDisconnect(sess *frame.SSClientSession) {
	elog.InfoAf("[RankClientMsgHandler] SessId=%v OnDisconnect", sess.GetSessID())
	GRankClient.SetSession(nil)
	r.timerRegister.KillAllTimer()
}
func (r *RankClientMsgHandler) OnBeatHeartError(sess *frame.SSClientSession) {

}

func OnHandlerR2CRankVerifyAck(datas []byte, sess *frame.SSClientSession) bool {
	GRankClient.gatewayVerifyOk = true
	elog.InfoA("[RankGatewayMsgHandler] Verify Ok")
	return true
}

func OnHandlerR2CTranferMsg(datas []byte, sess *frame.SSClientSession) bool {
	tranfer := pb.C2RTranferMsg{}
	unmarshalErr := proto.Unmarshal(datas, &tranfer)
	if unmarshalErr != nil {
		return false
	}

	cbMap := GRankClient.GetCbMap()
	cbItem, ok := cbMap[tranfer.Cbid]
	if !ok {
		elog.ErrorAf("[RankClientMsgHandler] Not Find CbId=%v", tranfer.Cbid)
		return false
	}

	if cbItem.Func != nil {
		cbItem.Func(tranfer.Tid, tranfer.Datas, cbItem.Agrs)
	}

	return true
}

var GRankClient *RankClient

func init() {
	GRankClient = NewRankClient()
}
