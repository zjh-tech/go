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
	evt_queue      IEventQueue
	http_evt_queue IEventQueue
	conn_queue     chan ConnEvent
	multi_flag     bool
}

func new_net(max_evt_count uint32, max_conn_count uint32) *Net {
	return &Net{
		evt_queue:      new_event_queue(max_evt_count),
		http_evt_queue: new_event_queue(max_conn_count),
		conn_queue:     make(chan ConnEvent, max_conn_count),
		multi_flag:     false,
	}
}

func (n *Net) Init() bool {
	go func() {
		for {
			select {
			case evt := <-n.conn_queue:
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
	n.multi_flag = true
	if GMsgHandlerPool == nil {
		chan_size := 100000
		GMsgHandlerPool = NewMsgHandlerPool(runtime.NumCPU(), chan_size)
		GMsgHandlerPool.Init()
	}
}

func (n *Net) PushEvent(evt IEvent) {
	if n.multi_flag {
		GMsgHandlerPool.PushEvent(evt)
	} else {
		n.evt_queue.PushEvent(evt)
	}
}

func (n *Net) PushSingleHttpEvent(http_evt IHttpEvent) {
	n.http_evt_queue.PushEvent(http_evt)
}

func (n *Net) PushMultiHttpEvent(http_evt IHttpEvent) {
	http_evt.ProcessMsg()
}

func (n *Net) Listen(addr string, factory ISessionFactory, listen_max_count int) bool {
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

	go func(sessfactory ISessionFactory, listen *net.TCPListener, listen_max_count int) {
		for {
			net_conn, accept_err := listen.AcceptTCP()
			if accept_err != nil {
				ELog.ErrorAf("[Net] Accept Error=%v", accept_err)
				continue
			}
			ELog.InfoAf("[Net] Accept Remote Addr %v", net_conn.RemoteAddr().String())

			if sessfactory.GetSessionCount() >= listen_max_count {
				ELog.ErrorA("[Net] Conn is Full")
				net_conn.Close()
				continue
			}

			session := sessfactory.CreateSession()
			if session == nil {
				ELog.ErrorA("[Net] CreateSession Error")
				net_conn.Close()
				continue
			}

			conn := GConnectionMgr.Create(n, net_conn, session)
			session.SetConnection(conn)
			go conn.Start()
		}
	}(factory, listen, listen_max_count)

	return true
}

func (n *Net) Connect(addr string, sess ISession) {
	connEvt := ConnEvent{
		addr: addr,
		sess: sess,
	}
	n.conn_queue <- connEvt
}

func (n *Net) Run(loopCount int) bool {
	for i := 0; i < loopCount; i++ {
		select {
		case evt, ok := <-n.evt_queue.GetEventQueue():
			tcp_evt := evt.(*TcpEvent)
			if !ok {
				return false
			}

			return tcp_evt.ProcessMsg()
		case evt, ok := <-n.http_evt_queue.GetEventQueue():
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
	GNet = new_net(1024*10, 60000)
}
