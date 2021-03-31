package enet

import (
	"net"
	"projects/go-engine/elog"
	"projects/go-engine/inet"
)

type ConnEvent struct {
	addr string
	sess inet.ISession
}

type Net struct {
	evt_queue  inet.IEventQueue
	conn_mgr   inet.IConnectionMgr
	conn_queue chan ConnEvent
}

func new_net(max_evt_count uint32, max_conn_count uint32) *Net {
	return &Net{
		evt_queue:  new_event_queue(max_evt_count),
		conn_mgr:   NewConnectionMgr(),
		conn_queue: make(chan ConnEvent, max_conn_count),
	}
}

func (n *Net) Init() bool {
	go func() {
		for {
			select {
			case evt := <-n.conn_queue:
				addr, err := net.ResolveTCPAddr("tcp4", evt.addr)
				if err != nil {
					elog.Errorf("[Net] Connect Addr=%v ResolveTCPAddr Error=%v", addr, err)
					continue
				}

				netConn, dial_err := net.DialTCP("tcp4", nil, addr)
				if dial_err != nil {
					elog.Errorf("[Net] Connect Addr=%v DialTCP Error=%v", addr, dial_err)
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
	elog.Info("[Net] Stop")
}

func (n *Net) PushEvent(evt inet.IEvent) {
	n.evt_queue.PushEvent(evt)
}

func (n *Net) Listen(addr string, factory inet.ISessionFactory, listen_max_count int) bool {
	tcp_addr, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		elog.Errorf("[Net] Addr=%v ResolveTCPAddr Error=%v", addr, err)
		return false
	}

	listen, listen_err := net.ListenTCP("tcp4", tcp_addr)
	if listen_err != nil {
		elog.Errorf("[Net] Addr=%v ListenTCP Error=%v", tcp_addr, listen_err)
		return false
	}

	elog.Infof("[Net] Addr=%v ListenTCP Success", tcp_addr)

	go func(sessfactory inet.ISessionFactory, listen *net.TCPListener, listen_max_count int) {
		for {
			net_conn, accept_err := listen.AcceptTCP()
			if accept_err != nil {
				elog.ErrorAf("[Net] Accept Error=%v", accept_err)
				continue
			}
			elog.InfoAf("[Net] Accept Remote Addr %v", net_conn.RemoteAddr().String())

			if n.conn_mgr.GetConnCount() >= listen_max_count {
				elog.ErrorA("[Net] Conn is Full")
				net_conn.Close()
				continue
			}

			session := sessfactory.CreateSession()
			if session == nil {
				elog.ErrorA("[Net] CreateSession Error")
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

func (n *Net) Connect(addr string, sess inet.ISession) {
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
			if !ok {
				return false
			}

			evt_type := evt.GetType()
			conn := evt.GetConn()
			if conn == nil {
				elog.ErrorA("[Net] Run Conn Is Nil")
				return false
			}

			session := conn.GetSession()
			if session == nil {
				elog.ErrorA("[Net] Run Session Is Nil")
				return false
			}

			if evt_type == inet.ConnEstablishType {
				session.SetConnection(conn)
				session.OnEstablish()
			} else if evt_type == inet.ConnRecvMsgType {
				datas := evt.GetDatas().([]byte)
				session.ProcessMsg(datas)
			} else if evt_type == inet.ConnCloseType {
				conn.OnClose()
				n.conn_mgr.Remove(conn.GetConnID())
				session.SetConnection(nil)
				session.OnTerminate()
			}

			return true
		default:
			return false
		}
	}
	elog.ErrorA("[Net] Run Error")
	return false
}

var GNet *Net

func init() {
	GNet = new_net(1024*10, 60000)
}
