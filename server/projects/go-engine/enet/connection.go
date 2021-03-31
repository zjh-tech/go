package enet

import (
	"io"
	"net"
	"projects/go-engine/elog"
	"projects/go-engine/inet"
	"sync/atomic"
)

//var GSendQps int64 = 0
//var GRecvQps int64 = 0

const (
	ConnEstablishState uint32 = iota
	ConnClosedState
)

type Connection struct {
	conn_id          uint64
	net              inet.INet
	conn             *net.TCPConn
	msg_chan         chan []byte
	msg_buff_chan    chan []byte
	active_exit_chan chan bool
	session          inet.ISession
	state            uint32
}

func NewConnection(conn_id uint64, net inet.INet, conn *net.TCPConn, sess inet.ISession) *Connection {
	maxMsgChanSize := 500000
	elog.InfoAf("[Net][Connection] ConnID=%v Attach SessID=%v", conn_id, sess.GetSessID())
	return &Connection{
		conn_id:          conn_id,
		net:              net,
		conn:             conn,
		session:          sess,
		msg_chan:         make(chan []byte),
		msg_buff_chan:    make(chan []byte, maxMsgChanSize),
		active_exit_chan: make(chan bool),
		state:            ConnEstablishState,
	}
}

func (c *Connection) GetConnID() uint64 {
	return c.conn_id
}

func (c *Connection) StartWriter() {
	elog.InfoAf("[Net][Connection] ConnID=%v Write Goroutine Start", c.conn_id)

	defer c.close(false)
	active_exit_flag := false
	for {
		if active_exit_flag && len(c.msg_chan) == 0 && len(c.msg_buff_chan) == 0 {
			elog.InfoAf("[Net][Connection] ConnID=%v Write Goroutine Active Exit", c.conn_id)
			return
		}

		select {
		case datas := <-c.msg_chan:
			if _, err := c.conn.Write(datas); err != nil {
				elog.ErrorAf("[Net][Connection] ConnID=%v Write Goroutine Exit SendError=%v", c.conn_id, err)
				return
			}
		case datas, _ := <-c.msg_buff_chan:
			if _, err := c.conn.Write(datas); err != nil {
				elog.ErrorAf("[Net][Connection] ConnID=%v Write Goroutine Exit SendBuffError=%v", c.conn_id, err)
				return
			}
		case flag, _ := <-c.active_exit_chan:
			if !flag {
				elog.ErrorAf("[Net][Connection] ConnID=%v Write Goroutine Passive Exit", c.conn_id)
				return
			}

			active_exit_flag = true
		}
	}
}

func (c *Connection) StartReader() {
	elog.InfoAf("[Net][Connection] ConnID=%v Read Goroutine Start", c.conn_id)
	defer c.close(false)

	for {
		if atomic.LoadUint32(&c.state) == ConnClosedState {
			elog.InfoAf("[Net][Connection] ConnID=%v Read Goroutine Exit", c.conn_id)
			return
		}

		coder := c.session.GetCoder()
		head_bytes := make([]byte, coder.GetHeaderLen())
		if _, head_err := io.ReadFull(c.conn, head_bytes); head_err != nil {
			elog.ErrorAf("[Net][Connection] ConnID=%v Read Goroutine Exit ReadFullError=%v", c.conn_id, head_err)
			return
		}

		bodylen, bodylen_err := coder.GetBodyLen(head_bytes)
		if bodylen_err != nil {
			elog.ErrorAf("[Net][Connection] ConnID=%v Read Goroutine Exit GetUnpackBodyLenError=%V", c.conn_id, bodylen_err)
			return
		}

		body_bytes := make([]byte, bodylen)
		if _, body_err := io.ReadFull(c.conn, body_bytes); body_err != nil {
			elog.ErrorAf("[Net][Connection] ConnID=%v Read Goroutine Exit ReadBodyError=%v", c.conn_id, body_err)
			return
		}

		decode_datas, decode_err := coder.DecodeBody(body_bytes)
		if decode_err != nil {
			elog.ErrorAf("[Net][Connection] ConnID=%v Read Goroutine Exit DecodeBodyError=%v", c.conn_id, decode_err)
			return
		}

		unzip_datas, unzip_err := coder.UnzipBody(decode_datas)
		if unzip_err != nil {
			elog.ErrorAf("[Net][Connection] ConnID=%v Read Goroutine Exit UnzipBodyError=%v", c.conn_id, unzip_err)
			return
		}

		msg_event := NewEvent(inet.ConnRecvMsgType, c, unzip_datas)
		c.net.PushEvent(msg_event)

		//atomic.AddInt64(&GRecvQps, 1)
	}
}

func (c *Connection) Start() {
	establishEvent := NewEvent(inet.ConnEstablishType, c, nil)
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

	closeEvent := NewEvent(inet.ConnCloseType, c, nil)
	c.net.PushEvent(closeEvent)

	if terminate {
		elog.InfoAf("[Net][Connection] ConnID=%v Active Closed", c.conn_id)
		c.active_exit_chan <- true
	} else {
		elog.InfoAf("[Net][Connection] ConnID=%v Passive Closed", c.conn_id)
		c.active_exit_chan <- false
	}
}

func (c *Connection) OnClose() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *Connection) Send(datas []byte) {
	if atomic.LoadUint32(&c.state) != ConnEstablishState {
		elog.WarnAf("[Net][Connection] ConnID=%v Send Error", c.conn_id)
		return
	}

	c.msg_chan <- datas
	//atomic.AddInt64(&GSendQps, 1)
}

func (c *Connection) AsyncSend(datas []byte) {
	c.msg_buff_chan <- datas
	//atomic.AddInt64(&GSendQps, 1)
}

func (c *Connection) GetSession() inet.ISession {
	return c.session
}
