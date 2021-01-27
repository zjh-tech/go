package main

import (
	"math/rand"
	"projects/frame"
	"projects/go-engine/ehttp"
	"projects/go-engine/elog"
	"projects/go-engine/enet"
	"projects/go-engine/etimer"
	"projects/pb"
	"projects/rank/tsrankclient"
	"projects/util"
	"time"

	"github.com/golang/protobuf/proto"
)

type TsRobot struct {
	timerRegister etimer.ITimerRegister
}

func (t *TsRobot) Init() bool {
	t.timerRegister = etimer.NewTimerRegister()

	//Log
	elog.InitLog("./log", 0, nil, nil)
	elog.Info("Server Log System Init Success")

	rand.Seed(time.Now().UnixNano())

	if !enet.GNet.Init() {
		elog.Error("TsRobot Net Init Error")
		return false
	}
	elog.Info("TsRobot Net Init Success")

	tsrankclient.GTsRankClient.Init("192.168.92.143:20000", "123456")

	elog.Info("TsRobot Init Success")
	t.AddSendRankReqTimer()

	return true
}

func (r *TsRobot) Run() {
	busy := false
	net_module := enet.GNet
	http_net_module := ehttp.GHttpNet
	timer_module := etimer.GTimerMgr
	for {
		busy = false

		if net_module.Run(frame.NET_LOOP_COUNT) {
			busy = true
		}

		if http_net_module.Run(frame.HTTP_LOOP_COUNT) {
			busy = true
		}

		if timer_module.Update(frame.TIMER_LOOP_COUNT) {
			busy = true
		}

		if !busy {
			time.Sleep(1 * time.Millisecond)
		}
	}
}

func (t *TsRobot) UnInit() {

}

const (
	SEND_RANK_REQ_TIME_ID uint32 = 1
)

const (
	SS_CLIENT_HEART_TIME_DELAY uint64 = 1000 * 10
)

var GPlayerID uint64
var GSortFiled2 int64

func (t *TsRobot) AddSendRankReqTimer() {
	t.timerRegister.AddRepeatTimer(SEND_RANK_REQ_TIME_ID, SS_CLIENT_HEART_TIME_DELAY, "TsRankClient-SendRankReq", func(v ...interface{}) {
		tid := uint32(1)
		rankItem := &pb.RankItem{}
		GPlayerID++
		rankItem.PlayerId = GPlayerID
		sortFiled1 := int64(util.GetRandom(1, 100000))
		if GPlayerID <= 2 {
			rankItem.SortField1 = 100000
			rankItem.SortField2 = 100000
		} else if GPlayerID == 3 {
			rankItem.SortField1 = 100001
			rankItem.SortField2 = sortFiled1
		} else {
			rankItem.SortField1 = sortFiled1
			GSortFiled2++
			rankItem.SortField2 = GSortFiled2
		}

		elog.InfoAf("RankReq Send PlayerId=%v SortFiled1=%v", GPlayerID, sortFiled1)
		tsrankclient.GTsRankClient.SendUpdateRankReq(tid, rankItem, func(attach ...interface{}) {
			tempTid := attach[0].(uint32)
			ackDatas := attach[1].([]byte)
			ack := pb.Ts2CUpdateRankAck{}
			unmarshalErr := proto.Unmarshal(ackDatas, &ack)
			if unmarshalErr != nil {
				return
			}

			cbArgs := attach[2].([]interface{})
			tempPlayerID := cbArgs[0].(uint64)
			tempSortFiled1 := cbArgs[1].(int64)
			elog.InfoAf("RankReq CallBack Tid=%v PlayerId=%v SortFiled1=%v ack=%+v", tempTid, tempPlayerID, tempSortFiled1, ack)
		}, []interface{}{GPlayerID, sortFiled1})
	}, []interface{}{}, true)
}
