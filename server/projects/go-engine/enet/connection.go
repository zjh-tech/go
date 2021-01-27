package enet

import (
	"io"
	"net"
	"projects/go-engine/elog"
	"projects/go-engine/inet"
	"sync/atomic"
)

var GSendQps int64 = 0
var GRecvQps int64 = 0

const (
	ConnEstablishState uint32 = iota
	ConnClosedState
)

type Connection struct {
	connID         uint64
	net            inet.INet
	conn           *net.TCPConn
	msgChan        chan []byte
	msgBuffChan    chan []byte
	activeExitChan chan bool
	session        inet.ISession
	state          uint32
}

func NewConnection(connID uint64, net inet.INet, conn *net.TCPConn, sess inet.ISession) *Connection {
	maxMsgChanSize := 500000
	elog.InfoAf("[Net][Connection] ConnID=%v Attach SessID=%v", connID, sess.GetSessID())
	return &Connection{
		connID:         connID,
		net:            net,
		conn:           conn,
		session:        sess,
		msgChan:        make(chan []byte),
		msgBuffChan:    make(chan []byte, maxMsgChanSize),
		activeExitChan: make(chan bool),
		state:          ConnEstablishState,
	}
}

func (c *Connection) GetConnID() uint64 {
	return c.connID
}

func (c *Connection) StartWriter() {
	elog.InfoAf("[Net][Connection] ConnID=%v Write Goroutine Start", c.connID)

	defer c.close(false)
	activeExitflag := false
	for {
		if activeExitflag && len(c.msgChan) == 0 && len(c.msgBuffChan) == 0 {
			elog.InfoAf("[Net][Connection] ConnID=%v Write Goroutine Active Exit", c.connID)
			return
		}

		select {
		case datas := <-c.msgChan:
			if _, err := c.conn.Write(datas); err != nil {
				elog.ErrorAf("[Net][Connection] ConnID=%v Write Goroutine Exit SendError=%v", c.connID, err)
				return
			}
		case datas, _ := <-c.msgBuffChan:
			if _, err := c.conn.Write(datas); err != nil {
				elog.ErrorAf("[Net][Connection] ConnID=%v Write Goroutine Exit SendBuffError=%v", c.connID, err)
				return
			}
		case flag, _ := <-c.activeExitChan:
			if !flag {
				elog.ErrorAf("[Net][Connection] ConnID=%v Write Goroutine Passive Exit", c.connID)
				return
			}

			activeExitflag = true
		}
	}
}

func (c *Connection) StartReader() {
	elog.InfoAf("[Net][Connection] ConnID=%v Read Goroutine Start", c.connID)
	defer c.close(false)

	for {
		if atomic.LoadUint32(&c.state) == ConnClosedState {
			elog.InfoAf("[Net][Connection] ConnID=%v Read Goroutine Exit", c.connID)
			return
		}

		coder := c.session.GetCoder()
		headBytes := make([]byte, coder.GetHeaderLen())
		if _, headErr := io.ReadFull(c.conn, headBytes); headErr != nil {
			elog.ErrorAf("[Net][Connection] ConnID=%v Read Goroutine Exit ReadFullError=%v", c.connID, headErr)
			return
		}

		bodylen, bodyLenErr := coder.GetBodyLen(headBytes)
		if bodyLenErr != nil {
			elog.ErrorAf("[Net][Connection] ConnID=%v Read Goroutine Exit GetUnpackBodyLenError=%V", c.connID, bodyLenErr)
			return
		}

		bodyBytes := make([]byte, bodylen)
		if _, bodyErr := io.ReadFull(c.conn, bodyBytes); bodyErr != nil {
			elog.ErrorAf("[Net][Connection] ConnID=%v Read Goroutine Exit ReadBodyError=%v", c.connID, bodyErr)
			return
		}

		decodeDatas, decodeErr := coder.DecodeBody(bodyBytes)
		if decodeErr != nil {
			elog.ErrorAf("[Net][Connection] ConnID=%v Read Goroutine Exit DecodeBodyError=%v", c.connID, decodeErr)
			return
		}

		unzipDatas, unzipErr := coder.UnzipBody(decodeDatas)
		if unzipErr != nil {
			elog.ErrorAf("[Net][Connection] ConnID=%v Read Goroutine Exit UnzipBodyError=%v", c.connID, unzipErr)
			return
		}

		msgEvent := NewEvent(inet.ConnRecvMsgType, c, unzipDatas)
		c.net.PushEvent(msgEvent)

		atomic.AddInt64(&GRecvQps, 1)
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
		elog.InfoAf("[Net][Connection] ConnID=%v Active Closed", c.connID)
		c.activeExitChan <- true
	} else {
		elog.InfoAf("[Net][Connection] ConnID=%v Passive Closed", c.connID)
		c.activeExitChan <- false
	}
}

func (c *Connection) OnClose() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *Connection) Send(datas []byte) {
	if atomic.LoadUint32(&c.state) != ConnEstablishState {
		elog.WarnAf("[Net][Connection] ConnID=%v Send Error", c.connID)
		return
	}

	c.msgChan <- datas
	atomic.AddInt64(&GSendQps, 1)
}

func (c *Connection) AsyncSend(datas []byte) {
	c.msgBuffChan <- datas
	atomic.AddInt64(&GSendQps, 1)
}

func (c *Connection) GetSession() inet.ISession {
	return c.session
}
