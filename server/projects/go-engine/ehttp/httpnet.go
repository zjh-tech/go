package ehttp

import (
	"projects/go-engine/elog"
)

type HttpNet struct {
	httpEvtQueue IHttpEventQueue
}

func newHttpNet() *HttpNet {
	return &HttpNet{
		httpEvtQueue: NewHttpEventQueue(4096),
	}
}

func (h *HttpNet) PushHttpEvent(httpEvt IHttpEvent) {
	h.httpEvtQueue.PushHttpEvent(httpEvt)
}

func (h *HttpNet) Run(loop_count int) bool {
	for i := 0; i < loop_count; i++ {
		select {
		case httpEvt, ok := <-h.httpEvtQueue.GetHttpEventQueue():
			if !ok {
				return false
			}

			conn := httpEvt.GetHttpConnection()
			if conn == nil {
				elog.ErrorA("[HttpNet] Run HttpConnection Is Nil")
				return false
			}

			conn.OnHandler(httpEvt.GetMsgID(), httpEvt.GetDatas())
			return true
		default:
			return false
		}
	}
	elog.ErrorA("[HttpNet] Run Error")
	return false
}

func (h *HttpNet) UnInit() {
	elog.InfoA("[HttpNet] Stop")
}

var GHttpNet *HttpNet

func init() {
	GHttpNet = newHttpNet()
}
