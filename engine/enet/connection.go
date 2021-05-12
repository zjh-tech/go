package enet

import (
	"io"
	"net"
	"sync/atomic"
	"time"
)

type Connection struct {
	conn_id       uint64
	net           INet
	conn          *net.TCPConn
	msg_chan      chan []byte
	msg_buff_chan chan []byte
	exit_chan     chan struct{}
	session       ISession
	state         uint32
}

func NewConnection(conn_id uint64, net INet, conn *net.TCPConn, sess ISession) *Connection {
	max_msg_chan_size := 500000
	ELog.InfoAf("[Net][Connection] ConnID=%v Attach SessID=%v", conn_id, sess.GetSessID())
	return &Connection{
		conn_id:       conn_id,
		net:           net,
		conn:          conn,
		session:       sess,
		msg_chan:      make(chan []byte),
		msg_buff_chan: make(chan []byte, max_msg_chan_size),
		exit_chan:     make(chan struct{}),
		state:         ConnEstablishState,
	}
}

func (c *Connection) GetConnID() uint64 {
	return c.conn_id
}

func (c *Connection) StartWriter() {
	ELog.InfoAf("[Net][Connection] ConnID=%v Write Goroutine Start", c.conn_id)

	defer c.close(false)
	for {
		select {
		case datas := <-c.msg_chan:
			if _, err := c.conn.Write(datas); err != nil {
				ELog.ErrorAf("[Net][Connection] ConnID=%v Write Goroutine Exit SendError=%v", c.conn_id, err)
				return
			}
		case datas, _ := <-c.msg_buff_chan:
			ELog.DebugAf("StartWriter ConnID=%v,Len=%v", c.conn_id, len(datas))
			if _, err := c.conn.Write(datas); err != nil {
				ELog.ErrorAf("[Net][Connection] ConnID=%v Write Goroutine Exit SendBuffError=%v", c.conn_id, err)
				return
			}
		case <-c.exit_chan:
			{
				ELog.ErrorAf("[Net][Connection] ConnID=%v Write Goroutine  Exit", c.conn_id)
				return
			}
		}
	}
}

func (c *Connection) StartReader() {
	ELog.InfoAf("[Net][Connection] ConnID=%v Read Goroutine Start", c.conn_id)
	defer c.close(false)

	for {
		if atomic.LoadUint32(&c.state) == ConnClosedState {
			ELog.InfoAf("[Net][Connection] ConnID=%v Read Goroutine Exit", c.conn_id)
			return
		}

		coder := c.session.GetCoder()

		header_len := coder.GetHeaderLen()
		head_bytes := make([]byte, header_len)
		ELog.DebugAf("StartReader ConnID=%v HeaderLen=%v", c.conn_id, header_len)
		if _, head_err := io.ReadFull(c.conn, head_bytes); head_err != nil {
			ELog.ErrorAf("[Net][Connection] ConnID=%v Read Goroutine Exit ReadFullError=%v", c.conn_id, head_err)
			return
		}

		body_len, bodylen_err := coder.GetBodyLen(head_bytes)
		if bodylen_err != nil {
			ELog.ErrorAf("[Net][Connection] ConnID=%v Read Goroutine Exit GetUnpackBodyLenError=%V", c.conn_id, bodylen_err)
			return
		}

		ELog.DebugAf("StartReader ConnID=%v BodyLen=%v", c.conn_id, body_len)
		body_bytes := make([]byte, body_len)
		if _, body_err := io.ReadFull(c.conn, body_bytes); body_err != nil {
			ELog.ErrorAf("[Net][Connection] ConnID=%v Read Goroutine Exit ReadBodyError=%v", c.conn_id, body_err)
			return
		}

		decode_datas, decode_err := coder.DecodeBody(body_bytes)
		if decode_err != nil {
			ELog.ErrorAf("[Net][Connection] ConnID=%v Read Goroutine Exit DecodeBodyError=%v", c.conn_id, decode_err)
			return
		}

		unzip_datas, unzip_err := coder.UnzipBody(decode_datas)
		if unzip_err != nil {
			ELog.ErrorAf("[Net][Connection] ConnID=%v Read Goroutine Exit UnzipBodyError=%v", c.conn_id, unzip_err)
			return
		}

		msg_event := NewTcpEvent(ConnRecvMsgType, c, unzip_datas)
		c.net.PushEvent(msg_event)

		//atomic.AddInt64(&GRecvQps, 1)
	}
}

func (c *Connection) Start() {
	establishEvent := NewTcpEvent(ConnEstablishType, c, nil)
	c.net.PushEvent(establishEvent)

	go c.StartReader()
	go c.StartWriter()
}

func (c *Connection) Terminate() {
	c.close(true)
}

func (c *Connection) close(terminate bool) {
	if !atomic.CompareAndSwapUint32(&c.state, ConnEstablishState, ConnClosedState) {
		return
	}

	closeEvent := NewTcpEvent(ConnCloseType, c, terminate)
	c.net.PushEvent(closeEvent) //业务logci处理CloseEvent

	if terminate {
		//主动断开
		ELog.InfoAf("[Net][Connection] ConnID=%v Active Closed", c.conn_id)
		go func() {
			//等待发完所有消息或者超时后,关闭底层read,write
			close_timer := time.NewTicker(100 * time.Millisecond)
			defer close_timer.Stop()

			close_timeout_timer := time.NewTimer(60 * time.Second)
			defer close_timeout_timer.Stop()
			for {
				select {
				case <-close_timer.C:
					{
						if len(c.msg_chan) <= 0 && len(c.msg_buff_chan) <= 0 {
							c.on_close()
						}
					}
				case <-close_timeout_timer.C:
					{
						c.on_close()
						return
					}
				}
			}
		}()
	} else {
		//被动断开
		ELog.InfoAf("[Net][Connection] ConnID=%v Passive Closed", c.conn_id)
		c.on_close()
	}
}

func (c *Connection) on_close() {
	if c.conn != nil {
		c.exit_chan <- struct{}{} //close writer Goroutine
		c.conn.Close()            //close reader Goroutine
	}
}

func (c *Connection) Send(datas []byte) {
	if atomic.LoadUint32(&c.state) != ConnEstablishState {
		ELog.WarnAf("[Net][Connection] ConnID=%v Send Error", c.conn_id)
		return
	}

	c.msg_chan <- datas
	//atomic.AddInt64(&GSendQps, 1)
}

func (c *Connection) AsyncSend(datas []byte) {
	c.msg_buff_chan <- datas
	//atomic.AddInt64(&GSendQps, 1)
}

func (c *Connection) GetSession() ISession {
	return c.session
}
