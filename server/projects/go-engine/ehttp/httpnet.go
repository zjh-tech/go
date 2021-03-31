package ehttp

import (
	"projects/go-engine/elog"
)

type HttpNet struct {
	http_evt_queue IHttpEventQueue
}

func new_httpnet(max_count uint32) *HttpNet {
	return &HttpNet{
		http_evt_queue: new_http_event_queue(max_count),
	}
}

func (h *HttpNet) PushHttpEvent(http_evt IHttpEvent) {
	h.http_evt_queue.PushHttpEvent(http_evt)
}

func (h *HttpNet) Run(loop_count int) bool {
	for i := 0; i < loop_count; i++ {
		select {
		case httpEvt, ok := <-h.http_evt_queue.GetHttpEventQueue():
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
	GHttpNet = new_httpnet(1024 * 4)
}
