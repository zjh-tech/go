package frame

type IServerMsgHandler interface {
	OnHandler(msgID uint32, datas []byte, sess *SSSession)
}

type ILogicServer interface {
	IServerMsgHandler
	OnEstablish(serversess *SSSession)
	OnTerminate(serversess *SSSession)
	SetServerSession(serversess *SSSession)
	GetServerSession() *SSSession
}

type LogicServer struct {
	server_sess *SSSession
}

func (l *LogicServer) SetServerSession(serversess *SSSession) {
	l.server_sess = serversess
}

func (l *LogicServer) GetServerSession() *SSSession {
	return l.server_sess
}

type ILogicServerFactory interface {
	SetLogicServer(serversess *SSSession)
}
