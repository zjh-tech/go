package main

import (
	"math/rand"
	"projects/frame"
	"projects/go-engine/ehttp"
	"projects/go-engine/elog"
	"projects/go-engine/enet"
	"projects/go-engine/etimer"
	"projects/pb"
	"projects/sdk"
	"projects/util"
	"time"

	"github.com/golang/protobuf/proto"
)

type RankRobot struct {
	timerRegister etimer.ITimerRegister
}

func (r *RankRobot) Init() bool {
	r.timerRegister = etimer.NewTimerRegister()

	elog.Init("./log", 1, nil)
	elog.Info("Log System Init Success")

	rand.Seed(time.Now().UnixNano())

	if !enet.GNet.Init() {
		elog.Error("Net Init Error")
		return false
	}
	elog.Info("Net Init Success")

	sdk.GRankClient.Init("127.0.0.1:3001", "123456")

	elog.Info("Init Success")

	r.AddSendRankReqTimer()
	return true
}

func (r *RankRobot) Run() {
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

func (t *RankRobot) UnInit() {
	elog.UnInit(nil)
}

const (
	SEND_RANK_REQ_TIME_ID uint32 = 1
)

const (
	SS_CLIENT_HEART_TIME_DELAY uint64 = 1000 * 10
)

var GPlayerID uint64
var GSortFiled2 int64

func (t *RankRobot) AddSendRankReqTimer() {
	t.timerRegister.AddRepeatTimer(SEND_RANK_REQ_TIME_ID, SS_CLIENT_HEART_TIME_DELAY, "TsRankClient-SendRankReq", func(v ...interface{}) {
		tid := uint32(0)
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
		sdk.GRankClient.SendUpdateRankReq(tid, rankItem, func(attach ...interface{}) {
			tempTid := attach[0].(uint32)
			ackDatas := attach[1].([]byte)
			ack := pb.R2CUpdateRankAck{}
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

func main() {
	var rank RankRobot
	if rank.Init() {
		rank.Run()
	}

	rank.UnInit()
}
