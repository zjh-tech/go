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
	msgBuffChan chan []byte
	exitChan    chan struct{}
	session     ISession
	state       uint32
	writeState  uint32
}

func NewConnection(connId uint64, net INet, conn *net.TCPConn, sess ISession) *Connection {
	ELog.InfoAf("[Net][Connection] ConnID=%v Bind SessID=%v", connId, sess.GetSessID())
	return &Connection{
		connId:      connId,
		net:         net,
		conn:        conn,
		session:     sess,
		msgBuffChan: make(chan []byte, ConnectChannelMaxSize),
		exitChan:    make(chan struct{}),
		state:       ConnEstablishState,
		writeState:  IsFreeWriteState,
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
		case datas, ok := <-c.msgBuffChan:
			if !ok {
				ELog.ErrorAf("[Net][Connection] Write ConnID=%v MsgBuffChan Ok Error", c.connId)
				return
			}

			ELog.DebugAf("[Net][Connection] Write ConnID=%v,Len=%v", c.connId, len(datas))

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

func (c *Connection) DoWriter() {
	ELog.DebugAf("[Net][Connection] ConnID=%v DoWriter Start", c.connId)
	defer func() {
		ELog.DebugAf("[Net][Connection] ConnID=%v DoWriter End", c.connId)
		atomic.CompareAndSwapUint32(&c.writeState, IsWritingState, IsFreeWriteState)
	}()

	coder := c.session.GetCoder()
	packMaxLen := int(coder.GetPackageMaxLen())
	mixBuff := make([]byte, packMaxLen*2)
	mixIndex := 0
	i := 0
	totalCount := 0
	for {
		mixIndex = 0
		i = 0
		loopCount := len(c.msgBuffChan)
		if loopCount == 0 {
			//server 2 unity return
			return
		}

		if totalCount == ConnWriterSleepLoopCount {
			//server 2 server sleep
			totalCount = 0
			time.Sleep(1 * time.Microsecond)
		}
	loop:
		for ; i < loopCount; i++ {
			select {
			case datas, ok := <-c.msgBuffChan:
				if !ok {
					ELog.ErrorAf("[Net][Connection] Write Goroutine ConnID=%v MsgBuffChan Ok Error", c.connId)
					return
				}
				totalCount++
				dataLen := len(datas)
				//session send : ensure len datas < packMaxLen
				mixIndex += copy(mixBuff[mixIndex:], datas)
				ELog.DebugAf("[Net][Connection] Write Goroutine ConnID=%v,CopyLen=%v To MixBuff", c.connId, dataLen)

				if mixIndex >= packMaxLen {
					ELog.DebugAf("[Net][Connection] Write Goroutine ConnID=%v,Out Range MixBuff Len MixIndex=%v", c.connId, mixIndex)
					break loop
				}
			default:
				break loop
			}
		}

		if mixIndex != 0 {
			if _, err := c.conn.Write(mixBuff[0:mixIndex]); err != nil {
				ELog.ErrorAf("[Net][Connection] ConnID=%v Write Goroutine Exit Send Error=%v", c.connId, err)
				c.close(false)
				return
			}
			ELog.DebugAf("[Net][Connection] Write Goroutine ConnID=%v,Conn.Write MixIndex=%v", c.connId, mixIndex)
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

		headerLen := coder.GetHeaderLen()
		headBytes := make([]byte, headerLen)
		ELog.DebugAf("StartReader ConnID=%v HeaderLen=%v", c.connId, headerLen)
		if _, head_err := io.ReadFull(c.conn, headBytes); head_err != nil {
			ELog.ErrorAf("[Net][Connection] ConnID=%v Read Goroutine Exit ReadFullError=%v", c.connId, head_err)
			return
		}

		bodyLen, bodyLenErr := coder.GetBodyLen(headBytes)
		if bodyLenErr != nil {
			ELog.ErrorAf("[Net][Connection] ConnID=%v Read Goroutine Exit GetUnpackBodyLenError=%V", c.connId, bodyLenErr)
			return
		}

		ELog.DebugAf("StartReader ConnID=%v BodyLen=%v", c.connId, bodyLen)
		bodyBytes := make([]byte, bodyLen)
		if _, bodyErr := io.ReadFull(c.conn, bodyBytes); bodyErr != nil {
			ELog.ErrorAf("[Net][Connection] ConnID=%v Read Goroutine Exit ReadBodyError=%v", c.connId, bodyErr)
			return
		}

		realBodyBytes, realBodyBytesErr := coder.UnpackMsg(bodyBytes)
		if realBodyBytesErr != nil {
			ELog.ErrorAf("[Net][Connection] ConnID=%v Read Goroutine Exit DecodeBodyError=%v", c.connId, realBodyBytesErr)
			return
		}

		msgEvent := NewTcpEvent(ConnRecvMsgType, c, realBodyBytes)

		if c.session.GetSessionConcurrentFlag() {
			c.session.PushEvent(msgEvent)
		} else {
			c.net.PushEvent(msgEvent)
		}

		//atomic.AddInt64(&GRecvQps, 1)
	}
}

func (c *Connection) Start() {
	establishEvent := NewTcpEvent(ConnEstablishType, c, nil)
	if c.session.GetSessionConcurrentFlag() {
		c.session.SetConnection(c)
		c.session.StartSessionConcurrentGoroutine()
		c.session.PushEvent(establishEvent)
	} else {
		c.net.PushEvent(establishEvent)
	}

	go c.StartReader()
	//go c.StartWriter()
}

func (c *Connection) Terminate() {
	c.close(true)
}

func (c *Connection) close(terminate bool) {
	if !atomic.CompareAndSwapUint32(&c.state, ConnEstablishState, ConnClosedState) {
		return
	}

	closeEvent := NewTcpEvent(ConnCloseType, c, terminate)
	if c.session.GetSessionConcurrentFlag() {
		c.session.PushEvent(closeEvent)
	} else {
		c.net.PushEvent(closeEvent)
	}

	if terminate {
		//主动断开
		ELog.InfoAf("[Net][Connection] ConnID=%v Active Closed", c.connId)
		go func() {
			//等待发完所有消息或者超时后,关闭底层read,write
			closeTimer := time.NewTicker(100 * time.Millisecond)
			defer closeTimer.Stop()

			closeTimeoutTimer := time.NewTimer(60 * time.Second)
			defer closeTimeoutTimer.Stop()
			for {
				select {
				case <-closeTimer.C:
					{
						if len(c.msgBuffChan) <= 0 {
							c.onClose()
							return
						}
					}
				case <-closeTimeoutTimer.C:
					{
						c.onClose()
						return
					}
				}
			}
		}()
	} else {
		//被动断开
		ELog.InfoAf("[Net][Connection] ConnID=%v Passive Closed", c.connId)
		c.onClose()
	}
}

func (c *Connection) onClose() {
	if c.conn != nil {
		c.exitChan <- struct{}{} //close writer Goroutine
		c.conn.Close()           //close reader Goroutine
	}
}

func (c *Connection) AsyncSend(datas []byte) {
	if atomic.LoadUint32(&c.state) != ConnEstablishState {
		ELog.WarnAf("[Net][Connection] ConnID=%v Send Error", c.connId)
		return
	}

	c.msgBuffChan <- datas

	if atomic.CompareAndSwapUint32(&c.writeState, IsFreeWriteState, IsWritingState) {
		go c.DoWriter()
	}

	//atomic.AddInt64(&GSendQps, 1)
}

func (c *Connection) GetSession() ISession {
	return c.session
}
