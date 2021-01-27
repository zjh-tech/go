package inet

import "github.com/golang/protobuf/proto"

type SessionType uint32

const (
	SESS_CONNECT_TYPE SessionType = iota
	SESS_LISTEN_TYPE
)

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

	AsyncSendMsg(msgID uint32, datas []byte, attach IAttachParas) bool

	AsyncSendProtoMsg(msgID uint32, msg proto.Message, attach IAttachParas) bool

	//主动断开
	Terminate()
}

//ISessionFactory
type ISessionFactory interface {
	CreateSession() ISession
	AddSession(session ISession)
	RemoveSession(id uint64)
}
