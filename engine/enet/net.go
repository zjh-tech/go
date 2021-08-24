package enet

import (
	"fmt"
	"net"
	"sync/atomic"
	"time"
)

type ConnEvent struct {
	addr string
	sess ISession
}

type Net struct {
	evtQueue     IEventQueue
	httpEvtQueue IEventQueue
	connQueue    chan ConnEvent
	connState    uint32
}

func NewNet(maxEvtCount uint32, maxConnCount uint32) *Net {
	return &Net{
		evtQueue:     newEventQueue(maxEvtCount),
		httpEvtQueue: newEventQueue(maxConnCount),
		connQueue:    make(chan ConnEvent, maxConnCount),
	}
}

func (n *Net) Init() bool {
	ELog.Info("[Net] Init")
	return true
}

func (n *Net) UnInit() {
	ELog.Info("[Net] UnInit")
}

func (n *Net) PushEvent(evt IEvent) {
	n.evtQueue.PushEvent(evt)
}

func (n *Net) PushSingleHttpEvent(httpEvt IHttpEvent) {
	n.httpEvtQueue.PushEvent(httpEvt)
}

func (n *Net) PushMultiHttpEvent(httpEvt IHttpEvent) {
	httpEvt.ProcessMsg()
}

func (n *Net) Listen(addr string, factory ISessionFactory, listenMaxCount int, sessionConcurrentFlag bool) bool {
	ELog.Infof("[Net] Start ListenTCP Addr=%v", addr)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		message := fmt.Sprintf("[Net] Addr=%v ResolveTCPAddr Error=%v", addr, err)
		ELog.Errorf(message)
		time.Sleep(1 * time.Second)
		panic(message)
	}

	listen, listenErr := net.ListenTCP("tcp4", tcpAddr)
	if listenErr != nil {
		message := fmt.Sprintf("[Net] Addr=%v ListenTCP Error=%v", tcpAddr, listenErr)
		ELog.Errorf(message)
		time.Sleep(1 * time.Second)
		panic(message)
	}

	ELog.Infof("[Net] Addr=%v ListenTCP Success", tcpAddr)

	go func(n *Net, sessfactory ISessionFactory, listen *net.TCPListener, listenMaxCount int, sessionConcurrentFlag bool) {
		for {
			netConn, acceptErr := listen.AcceptTCP()
			if acceptErr != nil {
				ELog.ErrorAf("[Net] Accept Error=%v", acceptErr)
				continue
			}
			ELog.InfoAf("[Net] Accept Remote Addr %v", netConn.RemoteAddr().String())

			if sessfactory.GetSessionCount() >= listenMaxCount {
				ELog.ErrorA("[Net] Conn is Full")
				netConn.Close()
				continue
			}

			session := sessfactory.CreateSession(true)
			if session == nil {
				ELog.ErrorA("[Net] CreateSession Error")
				netConn.Close()
				continue
			}

			if sessionConcurrentFlag {
				session.SetSessionConcurrentFlag(true)
			}

			conn := GConnectionMgr.Create(n, netConn, session)
			session.SetConnection(conn)
			go conn.Start()
		}
	}(n, factory, listen, listenMaxCount, sessionConcurrentFlag)

	return true
}

func (n *Net) Connect(addr string, sess ISession) {
	connEvt := ConnEvent{
		addr: addr,
		sess: sess,
	}
	ELog.InfoAf("[Net] Connect Addr=%v In ConnQueue", addr)
	n.connQueue <- connEvt

	if atomic.CompareAndSwapUint32(&n.connState, IsFreeConnectState, IsConnectingState) {
		go n.doConnectGoroutine()
	}
}

func (n *Net) doConnectGoroutine() {
	ELog.Info("[Net] doConnect Goroutine Start")
	defer func() {
		ELog.Info("[Net] doConnect Goroutine Exit")
		atomic.CompareAndSwapUint32(&n.connState, IsConnectingState, IsFreeConnectState)
	}()

	for {
		select {
		case evt := <-n.connQueue:
			addr, err := net.ResolveTCPAddr("tcp4", evt.addr)
			if err != nil {
				ELog.Errorf("[Net] Connect Addr=%v ResolveTCPAddr Error=%v", addr, err)
				continue
			}

			netConn, dial_err := net.DialTCP("tcp4", nil, addr)
			if dial_err != nil {
				ELog.Errorf("[Net] Connect Addr=%v DialTCP Error=%v", addr, dial_err)
				continue
			}

			conn := GConnectionMgr.Create(n, netConn, evt.sess)
			go conn.Start()
		default:
			return
		}
	}
}

func (n *Net) Run(loopCount int) bool {
	n.Update()

	i := 0
	for ; i < loopCount; i++ {
		select {
		case evt, ok := <-n.evtQueue.GetEventQueue():
			if !ok {
				return false
			}
			tcpEvt := evt.(*TcpEvent)
			tcpEvt.ProcessMsg()
		case evt, ok := <-n.httpEvtQueue.GetEventQueue():
			if !ok {
				return false
			}
			httpEvt := evt.(*HttpEvent)
			httpEvt.ProcessMsg()
		default:
			return false
		}
	}
	return true
}

func (n *Net) Update() {
	if GSSSessionMgr != nil {
		GSSSessionMgr.Update()
	}

	if GCSSessionMgr != nil {
		GCSSessionMgr.Update()
	}
}

var GNet *Net

func init() {
	GNet = NewNet(NetChannelMaxSize, NetMaxConnectSize)
}
