package enet

//Tcp
type TcpEvent struct {
	eventType uint32
	conn      IConnection
	datas     interface{}
}

func NewTcpEvent(t uint32, c IConnection, datas interface{}) *TcpEvent {
	return &TcpEvent{
		eventType: t,
		conn:      c,
		datas:     datas,
	}
}

func (t *TcpEvent) GetConn() IConnection {
	return t.conn
}

func (t *TcpEvent) ProcessMsg() bool {
	if t.conn == nil {
		ELog.ErrorA("[Net] Run Conn Is Nil")
		return false
	}

	session := t.conn.GetSession()
	if session == nil {
		ELog.ErrorA("[Net] Run Session Is Nil")
		return false
	}

	if t.eventType == ConnEstablishType {
		session.SetConnection(t.conn)
		session.OnEstablish()
	} else if t.eventType == ConnRecvMsgType {
		datas := t.datas.([]byte)
		session.GetCoder().ProcessMsg(datas, session)
	} else if t.eventType == ConnCloseType {
		GConnectionMgr.Remove(t.conn.GetConnID())
		session.SetConnection(nil)
		session.OnTerminate()
	}
	return true
}

//Http
type HttpEvent struct {
	httpConn IHttpConnection
	msgId    uint32
	datas    []byte
}

func NewHttpEvent(httpConn IHttpConnection, msgId uint32, datas []byte) *HttpEvent {
	return &HttpEvent{
		httpConn: httpConn,
		msgId:    msgId,
		datas:    datas,
	}
}

func (h *HttpEvent) ProcessMsg() bool {
	if h.httpConn == nil {
		ELog.ErrorA("[Net] ProcessMsg Run HttpConnection Is Nil")
		return false
	}

	h.httpConn.OnHandler(h.msgId, h.datas)
	return true
}
