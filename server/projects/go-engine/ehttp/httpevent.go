package ehttp

//HttpEvent
type HttpEvent struct {
	httpConn IHttpConnection
	msgID    uint32
	datas    []byte
}

func NewHttpEvent(httpConn IHttpConnection, msgID uint32, datas []byte) *HttpEvent {
	return &HttpEvent{
		httpConn: httpConn,
		msgID:    msgID,
		datas:    datas,
	}
}

func (h *HttpEvent) GetHttpConnection() IHttpConnection {
	return h.httpConn
}
func (h *HttpEvent) GetMsgID() uint32 {
	return h.msgID
}

func (h *HttpEvent) GetDatas() []byte {
	return h.datas
}

//HttpEventQueue
type HttpEventQueue struct {
	httpEvtQueue chan IHttpEvent
}

func NewHttpEventQueue(maxCount uint32) *HttpEventQueue {
	return &HttpEventQueue{
		httpEvtQueue: make(chan IHttpEvent, maxCount),
	}
}

func (e *HttpEventQueue) PushHttpEvent(req IHttpEvent) {
	e.httpEvtQueue <- req
}

func (e *HttpEventQueue) GetHttpEventQueue() chan IHttpEvent {
	return e.httpEvtQueue
}
