package ehttp

type HttpEvent struct {
	http_conn IHttpConnection
	msg_id    uint32
	datas     []byte
}

func NewHttpEvent(http_conn IHttpConnection, msg_id uint32, datas []byte) *HttpEvent {
	return &HttpEvent{
		http_conn: http_conn,
		msg_id:    msg_id,
		datas:     datas,
	}
}

func (h *HttpEvent) GetHttpConnection() IHttpConnection {
	return h.http_conn
}
func (h *HttpEvent) GetMsgID() uint32 {
	return h.msg_id
}

func (h *HttpEvent) GetDatas() []byte {
	return h.datas
}

type HttpEventQueue struct {
	http_evt_queue chan IHttpEvent
}

func new_http_event_queue(max_count uint32) *HttpEventQueue {
	return &HttpEventQueue{
		http_evt_queue: make(chan IHttpEvent, max_count),
	}
}

func (e *HttpEventQueue) PushHttpEvent(req IHttpEvent) {
	e.http_evt_queue <- req
}

func (e *HttpEventQueue) GetHttpEventQueue() chan IHttpEvent {
	return e.http_evt_queue
}
