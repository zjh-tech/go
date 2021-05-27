package enet

import (
	"io"
	"net"
	"sync/atomic"
	"time"
)

type Connection struct {
	connId      uint64
	net         INet
	conn        *net.TCPConn
	msgChan     chan []byte
	msgBuffChan chan []byte
	exitChan    chan struct{}
	session     ISession
	state       uint32
}

func NewConnection(connId uint64, net INet, conn *net.TCPConn, sess ISession) *Connection {
	maxMsgChansize := 500000
	ELog.InfoAf("[Net][Connection] ConnID=%v Attach SessID=%v", connId, sess.GetSessID())
	return &Connection{
		connId:      connId,
		net:         net,
		conn:        conn,
		session:     sess,
		msgChan:     make(chan []byte),
		msgBuffChan: make(chan []byte, maxMsgChansize),
		exitChan:    make(chan struct{}),
		state:       ConnEstablishState,
	}
}

func (c *Connection) GetConnID() uint64 {
	return c.connId
}

func (c *Connection) StartWriter() {
	ELog.InfoAf("[Net][Connection] ConnID=%v Write Goroutine Start", c.connId)

	defer c.close(false)
	for {
		select {
		case datas := <-c.msgChan:
			if _, err := c.conn.Write(datas); err != nil {
				ELog.ErrorAf("[Net][Connection] ConnID=%v Write Goroutine Exit SendError=%v", c.connId, err)
				return
			}
		case datas, _ := <-c.msgBuffChan:
			ELog.DebugAf("StartWriter ConnID=%v,Len=%v", c.connId, len(datas))
			if _, err := c.conn.Write(datas); err != nil {
				ELog.ErrorAf("[Net][Connection] ConnID=%v Write Goroutine Exit SendBuffError=%v", c.connId, err)
				return
			}
		case <-c.exitChan:
			{
				ELog.ErrorAf("[Net][Connection] ConnID=%v Write Goroutine  Exit", c.connId)
				return
			}
		}
	}
}

func (c *Connection) StartReader() {
	ELog.InfoAf("[Net][Connection] ConnID=%v Read Goroutine Start", c.connId)
	defer c.close(false)

	for {
		if atomic.LoadUint32(&c.state) == ConnClosedState {
			ELog.InfoAf("[Net][Connection] ConnID=%v Read Goroutine Exit", c.connId)
			return
		}

		coder := c.session.GetCoder()

		header_len := coder.GetHeaderLen()
		head_bytes := make([]byte, header_len)
		ELog.DebugAf("StartReader ConnID=%v HeaderLen=%v", c.connId, header_len)
		if _, head_err := io.ReadFull(c.conn, head_bytes); head_err != nil {
			ELog.ErrorAf("[Net][Connection] ConnID=%v Read Goroutine Exit ReadFullError=%v", c.connId, head_err)
			return
		}

		body_len, bodylen_err := coder.GetBodyLen(head_bytes)
		if bodylen_err != nil {
			ELog.ErrorAf("[Net][Connection] ConnID=%v Read Goroutine Exit GetUnpackBodyLenError=%V", c.connId, bodylen_err)
			return
		}

		ELog.DebugAf("StartReader ConnID=%v BodyLen=%v", c.connId, body_len)
		body_bytes := make([]byte, body_len)
		if _, body_err := io.ReadFull(c.conn, body_bytes); body_err != nil {
			ELog.ErrorAf("[Net][Connection] ConnID=%v Read Goroutine Exit ReadBodyError=%v", c.connId, body_err)
			return
		}

		decode_datas, decode_err := coder.DecodeBody(body_bytes)
		if decode_err != nil {
			ELog.ErrorAf("[Net][Connection] ConnID=%v Read Goroutine Exit DecodeBodyError=%v", c.connId, decode_err)
			return
		}

		unzip_datas, unzip_err := coder.UnzipBody(decode_datas)
		if unzip_err != nil {
			ELog.ErrorAf("[Net][Connection] ConnID=%v Read Goroutine Exit UnzipBodyError=%v", c.connId, unzip_err)
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
		ELog.InfoAf("[Net][Connection] ConnID=%v Active Closed", c.connId)
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
						if len(c.msgChan) <= 0 && len(c.msgBuffChan) <= 0 {
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
		ELog.InfoAf("[Net][Connection] ConnID=%v Passive Closed", c.connId)
		c.on_close()
	}
}

func (c *Connection) on_close() {
	if c.conn != nil {
		c.exitChan <- struct{}{} //close writer Goroutine
		c.conn.Close()           //close reader Goroutine
	}
}

func (c *Connection) Send(datas []byte) {
	if atomic.LoadUint32(&c.state) != ConnEstablishState {
		ELog.WarnAf("[Net][Connection] ConnID=%v Send Error", c.connId)
		return
	}

	c.msgChan <- datas
	//atomic.AddInt64(&GSendQps, 1)
}

func (c *Connection) AsyncSend(datas []byte) {
	c.msgBuffChan <- datas
	//atomic.AddInt64(&GSendQps, 1)
}

func (c *Connection) GetSession() ISession {
	return c.session
}
