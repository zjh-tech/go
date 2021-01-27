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
	evtQueue  inet.IEventQueue
	connMgr   inet.IConnectionMgr
	connQueue chan ConnEvent
}

func newNet() *Net {
	return &Net{
		evtQueue:  NewEventQueue(4096),
		connMgr:   NewConnectionMgr(),
		connQueue: make(chan ConnEvent, 4096),
	}
}

func (n *Net) Init() bool {
	go func() {
		for {
			select {
			case evt := <-n.connQueue:
				addr, err := net.ResolveTCPAddr("tcp4", evt.addr)
				if err != nil {
					elog.Errorf("[Net] Connect Addr=%v ResolveTCPAddr Error=%v", addr, err)
					continue
				}

				netConn, err := net.DialTCP("tcp4", nil, addr)
				if err != nil {
					elog.Errorf("[Net] Connect Addr=%v DialTCP Error=%v", addr, err)
					continue
				}

				conn := n.connMgr.Create(n, netConn, evt.sess)
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
	n.evtQueue.PushEvent(evt)
}

func (n *Net) Listen(addr string, factory inet.ISessionFactory, listenMaxCount int) bool {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		elog.Errorf("[Net] Addr=%v ResolveTCPAddr Error=%v", addr, err)
		return false
	}

	listen, err := net.ListenTCP("tcp4", tcpAddr)
	if err != nil {
		elog.Errorf("[Net] Addr=%v ListenTCP Error=%v", tcpAddr, err)
		return false
	}

	elog.Infof("[Net] Addr=%v ListenTCP Success", tcpAddr)

	go func(sessfactory inet.ISessionFactory, listen *net.TCPListener, listenMaxCount int) {
		for {
			netConn, err := listen.AcceptTCP()
			if err != nil {
				elog.ErrorAf("[Net] Accept Error=%v", err)
				continue
			}
			elog.InfoAf("[Net] Accept Remote Addr %v", netConn.RemoteAddr().String())

			if n.connMgr.GetConnCount() >= listenMaxCount {
				elog.ErrorA("[Net] Conn is Full")
				netConn.Close()
				continue
			}

			session := sessfactory.CreateSession()
			if session == nil {
				elog.ErrorA("[Net] CreateSession Error")
				netConn.Close()
				continue
			}

			conn := n.connMgr.Create(n, netConn, session)
			session.SetConnection(conn)
			go conn.Start()
		}
	}(factory, listen, listenMaxCount)

	return true
}

func (n *Net) Connect(addr string, sess inet.ISession) {
	connEvt := ConnEvent{
		addr: addr,
		sess: sess,
	}
	n.connQueue <- connEvt
}

func (n *Net) Run(loop_count int) bool {
	for i := 0; i < loop_count; i++ {
		select {
		case evt, ok := <-n.evtQueue.GetEventQueue():
			if !ok {
				return false
			}

			evtType := evt.GetType()
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

			if evtType == inet.ConnEstablishType {
				session.SetConnection(conn)
				session.OnEstablish()
			} else if evtType == inet.ConnRecvMsgType {
				datas := evt.GetDatas().([]byte)
				session.ProcessMsg(datas)
			} else if evtType == inet.ConnCloseType {
				conn.OnClose()
				n.connMgr.Remove(conn.GetConnID())
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
	GNet = newNet()
}
