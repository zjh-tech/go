package enet

import (
	"net"
	"runtime"
)

type ConnEvent struct {
	addr string
	sess ISession
}

type Net struct {
	evtQueue     IEventQueue
	httpEvtQueue IEventQueue
	connQueue    chan ConnEvent
	multiFlag    bool
}

func newNet(max_evt_count uint32, max_conn_count uint32) *Net {
	return &Net{
		evtQueue:     new_event_queue(max_evt_count),
		httpEvtQueue: new_event_queue(max_conn_count),
		connQueue:    make(chan ConnEvent, max_conn_count),
		multiFlag:    false,
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

func (n *Net) SetMultiProcessMsg() {
	n.multiFlag = true
	if GMsgHandlerPool == nil {
		chanSize := 100000
		GMsgHandlerPool = NewMsgHandlerPool(runtime.NumCPU(), chanSize)
		GMsgHandlerPool.Init()
	}
}

func (n *Net) PushEvent(evt IEvent) {
	if n.multiFlag {
		GMsgHandlerPool.PushEvent(evt)
	} else {
		n.evtQueue.PushEvent(evt)
	}
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
		ELog.Errorf("[Net] Addr=%v ResolveTCPAddr Error=%v", addr, err)
		return false
	}

	listen, listen_err := net.ListenTCP("tcp4", tcp_addr)
	if listen_err != nil {
		ELog.Errorf("[Net] Addr=%v ListenTCP Error=%v", tcp_addr, listen_err)
		return false
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

			session := sessfactory.CreateSession()
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
	i := 0
	for ; i < loopCount; i++ {
		select {
		case evt, ok := <-n.evtQueue.GetEventQueue():
			tcp_evt := evt.(*TcpEvent)
			if !ok {
				return false
			}

			return tcp_evt.ProcessMsg()
		case evt, ok := <-n.httpEvtQueue.GetEventQueue():
			if !ok {
				return false
			}
			http_evt := evt.(*HttpEvent)
			return http_evt.ProcessMsg()
		default:
			return false
		}
	}
	ELog.ErrorA("[Net] Run Error")
	return false
}

var GNet *Net

func init() {
	GNet = newNet(1024*10, 60000)
}
