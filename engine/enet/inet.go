package enet

import (
	"net"

	"github.com/golang/protobuf/proto"
)

type INet interface {
	PushEvent(IEvent)
	PushSingleHttpEvent(IHttpEvent)
	PushMultiHttpEvent(IHttpEvent)
	Connect(addr string, sess ISession)
	Listen(addr string, factory ISessionFactory, listen_max_count int) bool
	Run(loop_count int) bool
}

type ICoder interface {
	GetHeaderLen() uint32
	GetBodyLen(datas []byte) (uint32, error)

	EnCodeBody(datas []byte) ([]byte, bool)
	DecodeBody(datas []byte) ([]byte, error)

	ZipBody(datas []byte) ([]byte, bool)
	UnzipBody(datas []byte) ([]byte, error)

	FillNetStream(datas []byte) ([]byte, error)
}

type IEventQueue interface {
	PushEvent(req interface{})
	GetEventQueue() chan interface{}
}

//Tcp
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

type IEvent interface {
	GetType() uint32
	GetConn() IConnection
	GetDatas() interface{}
}

//ISession
type ISession interface {
	SetConnection(conn IConnection)

	GetSessID() uint64

	OnEstablish()

	OnTerminate()

	ProcessMsg(datas []byte)

	GetCoder() ICoder

	SetCoder(coder ICoder)

	IsListenType() bool

	IsConnectType() bool

	SetConnectType()

	SetListenType()

	SetSessionFactory(factory ISessionFactory)

	GetSessionFactory() ISessionFactory

	AsyncSendMsg(msgID uint32, datas []byte) bool

	AsyncSendProtoMsg(msgID uint32, msg proto.Message) bool

	//主动断开
	Terminate()
}

//ISessionFactory
type ISessionFactory interface {
	CreateSession() ISession
	AddSession(session ISession)
	RemoveSession(id uint64)
}

//Http
type IHttpConnection interface {
	OnHandler(msg_id uint32, datas []byte)
}

type IHttpEvent interface {
	GetHttpConnection() IHttpConnection
	GetMsgID() uint32
	GetDatas() []byte
}