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

func (t *TcpEvent) GetType() uint32 {
	return t.event_type
}

func (t *TcpEvent) GetConn() IConnection {
	return t.conn
}

func (t *TcpEvent) GetDatas() interface{} {
	return t.datas
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

func (h *HttpEvent) GetHttpConnection() IHttpConnection {
	return h.http_conn
}
func (h *HttpEvent) GetMsgID() uint32 {
	return h.msg_id
}

func (h *HttpEvent) GetDatas() []byte {
	return h.datas
}
