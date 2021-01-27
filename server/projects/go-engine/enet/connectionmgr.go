package enet

import (
	"net"
	"projects/go-engine/elog"
	"projects/go-engine/inet"
	"sync"
)

type ConnectionMgr struct {
	conns      map[uint64]inet.IConnection
	connLocker sync.RWMutex
	nextId     uint64
}

func NewConnectionMgr() *ConnectionMgr {
	return &ConnectionMgr{
		conns:  make(map[uint64]inet.IConnection),
		nextId: 0,
	}
}

func (c *ConnectionMgr) Create(net inet.INet, netConn *net.TCPConn, sess inet.ISession) inet.IConnection {
	c.connLocker.Lock()
	defer c.connLocker.Unlock()
	c.nextId++
	conn := NewConnection(c.nextId, net, netConn, sess)
	c.conns[conn.GetConnID()] = conn
	elog.InfoAf("[Net][ConnectionMgr] Add ConnID=%v Connection", conn.connID)
	return conn
}

func (c *ConnectionMgr) Remove(id uint64) {
	c.connLocker.Lock()
	defer c.connLocker.Unlock()

	delete(c.conns, id)
	elog.InfoAf("[Net][ConnectionMgr] Remove ConnID=%v Connection", id)
}

func (c *ConnectionMgr) GetConnCount() int {
	c.connLocker.Lock()
	defer c.connLocker.Unlock()

	return len(c.conns)
}
