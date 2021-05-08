package enet

import (
	"net"
	"sync"
)

type ConnectionMgr struct {
	conns       map[uint64]IConnection
	conn_locker sync.RWMutex
	next_id     uint64
}

func NewConnectionMgr() *ConnectionMgr {
	return &ConnectionMgr{
		conns:   make(map[uint64]IConnection),
		next_id: 0,
	}
}

func (c *ConnectionMgr) Create(net INet, netConn *net.TCPConn, sess ISession) IConnection {
	c.conn_locker.Lock()
	defer c.conn_locker.Unlock()
	c.next_id++
	conn := NewConnection(c.next_id, net, netConn, sess)
	c.conns[conn.GetConnID()] = conn
	ELog.InfoAf("[Net][ConnectionMgr] Add ConnID=%v Connection", conn.conn_id)
	return conn
}

func (c *ConnectionMgr) Remove(id uint64) {
	c.conn_locker.Lock()
	defer c.conn_locker.Unlock()

	delete(c.conns, id)
	ELog.InfoAf("[Net][ConnectionMgr] Remove ConnID=%v Connection", id)
}

func (c *ConnectionMgr) GetConnCount() int {
	c.conn_locker.Lock()
	defer c.conn_locker.Unlock()

	return len(c.conns)
}
