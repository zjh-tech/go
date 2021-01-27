package ehttp

type IHttpConnection interface {
	OnHandler(msgID uint32, datas []byte)
}

type IHttpEvent interface {
	GetHttpConnection() IHttpConnection
	GetMsgID() uint32
	GetDatas() []byte
}

type IHttpEventQueue interface {
	PushHttpEvent(req IHttpEvent)
	GetHttpEventQueue() chan IHttpEvent
}
