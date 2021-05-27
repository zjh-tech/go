package enet

import "runtime"

//C++ 并发数量静态初始化
//Go  每个Session有个Channel,每个Entity有个Channel,并发的数量可以动态(建议使用这种模式)

type MsgHandlerPool struct {
	poolSize      int
	eventChanSets []chan IEvent
}

func NewMsgHandlerPool(poolSize, chanSize int) *MsgHandlerPool {
	cpu_num := runtime.NumCPU()
	if poolSize <= cpu_num {
		poolSize = cpu_num
	}

	chan_min_size := 100000
	if chanSize <= chan_min_size {
		chanSize = chan_min_size
	}

	obj := &MsgHandlerPool{
		poolSize:      poolSize,
		eventChanSets: make([]chan IEvent, 0),
	}

	for i := 0; i < poolSize; i++ {
		event_chan := make(chan IEvent, chanSize)
		obj.eventChanSets = append(obj.eventChanSets, event_chan)
	}

	return obj
}

func (a *MsgHandlerPool) Init() bool {
	for i := 0; i < a.poolSize; i++ {
		go a.run(i)
	}

	return true
}

func (a *MsgHandlerPool) run(index int) {
	evtQueue := a.eventChanSets[index]
	for {
		select {
		case evt := <-evtQueue:
			{
				DoMsgHandler(evt)
			}
		}
	}
}

func (a *MsgHandlerPool) PushEvent(evt IEvent) bool {
	if evt == nil {
		return false
	}

	conn := evt.GetConn()
	if conn == nil {
		return false
	}

	sess := conn.GetSession()
	if sess == nil {
		return false
	}

	workerId := int(sess.GetSessID()) % a.poolSize
	a.eventChanSets[workerId] <- evt
	return true
}

func DoMsgHandler(evt IEvent) {
	if evt == nil {
		return
	}

	evt.ProcessMsg()
}

var GMsgHandlerPool *MsgHandlerPool
