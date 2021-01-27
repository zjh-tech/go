package inet

import "net"

type IConnection interface {
	GetConnID() uint64
	GetSession() ISession
	Start()
	Send(datas []byte)
	AsyncSend(datas []byte)
	Terminate()
	OnClose()
}

type IConnectionMgr interface {
	Create(net INet, conn *net.TCPConn, sess ISession) IConnection
	Remove(id uint64)
	GetConnCount() int
}
