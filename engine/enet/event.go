package enet

//Tcp
type TcpEvent struct {
	event_type uint32
	conn       IConnection
	datas      interface{}
}

func NewTcpEvent(t uint32, c IConnection, datas interface{}) *TcpEvent {
	return &TcpEvent{
		event_type: t,
		conn:       c,
		datas:      datas,
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

	if t.event_type == ConnEstablishType {
		session.SetConnection(t.conn)
		session.OnEstablish()
	} else if t.event_type == ConnRecvMsgType {
		datas := t.datas.([]byte)
		session.GetCoder().ProcessMsg(datas, session)
	} else if t.event_type == ConnCloseType {
		GConnectionMgr.Remove(t.conn.GetConnID())
		session.SetConnection(nil)
		session.OnTerminate()
	}
	return true
}

//Http
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

func (h *HttpEvent) ProcessMsg() bool {
	if h.http_conn == nil {
		ELog.ErrorA("[Net] ProcessMsg Run HttpConnection Is Nil")
		return false
	}

	h.http_conn.OnHandler(h.msg_id, h.datas)
	return true
}
