package enet

import (
	"fmt"
	"net"
)

type ConnEvent struct {
	addr string
	sess ISession
}

type Net struct {
	evtQueue     IEventQueue
	httpEvtQueue IEventQueue
	connQueue    chan ConnEvent
}

func newNet(maxEvtCount uint32, maxConnCount uint32) *Net {
	return &Net{
		evtQueue:     newEventQueue(maxEvtCount),
		httpEvtQueue: newEventQueue(maxConnCount),
		connQueue:    make(chan ConnEvent, maxConnCount),
	}
}

func (n *Net) Init() bool {
	go func() {
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
			}
		}
	}()

	return true
}

func (n *Net) UnInit() {
	ELog.Info("[Net] Stop")
}

func (n *Net) PushEvent(evt IEvent) {
	n.evtQueue.PushEvent(evt)
}

func (n *Net) PushSingleHttpEvent(http_evt IHttpEvent) {
	n.httpEvtQueue.PushEvent(http_evt)
}

func (n *Net) PushMultiHttpEvent(http_evt IHttpEvent) {
	http_evt.ProcessMsg()
}

func (n *Net) Listen(addr string, factory ISessionFactory, listenMaxCount int) bool {
	tcp_addr, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		message := fmt.Sprintf("[Net] Addr=%v ResolveTCPAddr Error=%v", addr, err)
		ELog.Errorf(message)
		panic(message)
	}

	listen, listen_err := net.ListenTCP("tcp4", tcp_addr)
	if listen_err != nil {
		message := fmt.Sprintf("[Net] Addr=%v ListenTCP Error=%v", tcp_addr, listen_err)
		ELog.Errorf(message)
		panic(message)
	}

	ELog.Infof("[Net] Addr=%v ListenTCP Success", tcp_addr)

	go func(sessfactory ISessionFactory, listen *net.TCPListener, listenMaxCount int) {
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

			conn := GConnectionMgr.Create(n, netConn, session)
			session.SetConnection(conn)
			go conn.Start()
		}
	}(factory, listen, listenMaxCount)

	return true
}

func (n *Net) Connect(addr string, sess ISession) {
	connEvt := ConnEvent{
		addr: addr,
		sess: sess,
	}
	n.connQueue <- connEvt
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

	if GSDKSessionMgr != nil {
		GSDKSessionMgr.Update()
	}
}

var GNet *Net

func init() {
	GNet = newNet(NetChannelMaxSize, NetMaxConnectSize)
}
