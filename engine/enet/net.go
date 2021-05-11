package enet

import (
	"net"
)

type ConnEvent struct {
	addr string
	sess ISession
}

type Net struct {
	evt_queue      IEventQueue
	http_evt_queue IEventQueue
	conn_mgr       IConnectionMgr
	conn_queue     chan ConnEvent
}

func new_net(max_evt_count uint32, max_conn_count uint32) *Net {
	return &Net{
		evt_queue:      new_event_queue(max_evt_count),
		http_evt_queue: new_event_queue(max_conn_count),
		conn_mgr:       NewConnectionMgr(),
		conn_queue:     make(chan ConnEvent, max_conn_count),
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

				conn := n.conn_mgr.Create(n, netConn, evt.sess)
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
	n.evt_queue.PushEvent(evt)
}

func (n *Net) PushSingleHttpEvent(http_evt IHttpEvent) {
	n.http_evt_queue.PushEvent(http_evt)
}

func (n *Net) PushMultiHttpEvent(http_evt IHttpEvent) {
	conn := http_evt.GetHttpConnection()
	if conn == nil {
		ELog.ErrorA("[Net] PushMultiHttpEvent Run HttpConnection Is Nil")
		return
	}

	conn.OnHandler(http_evt.GetMsgID(), http_evt.GetDatas())
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

			if n.conn_mgr.GetConnCount() >= listen_max_count {
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

			conn := n.conn_mgr.Create(n, net_conn, session)
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

func (n *Net) Run(loop_count int) bool {
	for i := 0; i < loop_count; i++ {
		select {
		case evt, ok := <-n.evt_queue.GetEventQueue():
			tcp_evt := evt.(*TcpEvent)
			if !ok {
				return false
			}

			evt_type := tcp_evt.GetType()
			conn := tcp_evt.GetConn()
			if conn == nil {
				ELog.ErrorA("[Net] Run Conn Is Nil")
				return false
			}

			session := conn.GetSession()
			if session == nil {
				ELog.ErrorA("[Net] Run Session Is Nil")
				return false
			}

			if evt_type == ConnEstablishType {
				session.SetConnection(conn)
				session.OnEstablish()
			} else if evt_type == ConnRecvMsgType {
				datas := tcp_evt.GetDatas().([]byte)
				session.GetCoder().ProcessMsg(datas, session)
			} else if evt_type == ConnCloseType {
				conn.OnClose()
				n.conn_mgr.Remove(conn.GetConnID())
				session.SetConnection(nil)
				session.OnTerminate()
			}

			return true
		case evt, ok := <-n.http_evt_queue.GetEventQueue():
			if !ok {
				return false
			}
			http_evt := evt.(*HttpEvent)

			conn := http_evt.GetHttpConnection()
			if conn == nil {
				ELog.ErrorA("[Net] PushSingleHttpEvent Run HttpConnection Is Nil")
				return false
			}

			conn.OnHandler(http_evt.GetMsgID(), http_evt.GetDatas())
			return true
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
