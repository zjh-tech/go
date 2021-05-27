package frame

type IServerMsgHandler interface {
	OnHandler(msgId uint32, datas []byte, sess *SSSession)
}

type ILogicServer interface {
	IServerMsgHandler
	OnEstablish(serversess *SSSession)
	OnTerminate(serversess *SSSession)
	SetServerSession(serversess *SSSession)
	GetServerSession() *SSSession
}

type LogicServer struct {
	serverSess *SSSession
}

func (l *LogicServer) SetServerSession(serversess *SSSession) {
	l.serverSess = serversess
}

func (l *LogicServer) GetServerSession() *SSSession {
	return l.serverSess
}

type ILogicServerFactory interface {
	SetLogicServer(serversess *SSSession)
}
