package enet

import "runtime"

//C++ 并发数量静态初始化
//Go  每个Session有个Channel,每个Entity有个Channel,并发的数量可以动态(建议使用这种模式)

type MsgHandlerPool struct {
	pool_size       int
	event_chan_sets []chan IEvent
}

func NewMsgHandlerPool(pool_size, chan_size int) *MsgHandlerPool {
	cpu_num := runtime.NumCPU()
	if pool_size <= cpu_num {
		pool_size = cpu_num
	}

	chan_min_size := 100000
	if chan_size <= chan_min_size {
		chan_size = chan_min_size
	}

	obj := &MsgHandlerPool{
		pool_size:       pool_size,
		event_chan_sets: make([]chan IEvent, 0),
	}

	for i := 0; i < pool_size; i++ {
		event_chan := make(chan IEvent, chan_size)
		obj.event_chan_sets = append(obj.event_chan_sets, event_chan)
	}

	return obj
}

func (a *MsgHandlerPool) Init() bool {
	for i := 0; i < a.pool_size; i++ {
		go a.run(i)
	}

	return true
}

func (a *MsgHandlerPool) run(index int) {
	evt_queue := a.event_chan_sets[index]
	for {
		select {
		case evt := <-evt_queue:
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

	worker_id := int(sess.GetSessID()) % a.pool_size
	a.event_chan_sets[worker_id] <- evt
	return true
}

func DoMsgHandler(evt IEvent) {
	if evt == nil {
		return
	}

	evt.ProcessMsg()
}

var GMsgHandlerPool *MsgHandlerPool
