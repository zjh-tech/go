package inet

const (
	ConnEstablishType uint32 = iota
	ConnRecvMsgType
	ConnCloseType
)

type IEvent interface {
	GetType() uint32
	GetConn() IConnection
	GetDatas() interface{}
}

type IEventQueue interface {
	PushEvent(req IEvent)
	GetEventQueue() chan IEvent
}
