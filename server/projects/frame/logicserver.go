package frame

type IServerMsgHandler interface {
	OnHandler(msgID uint32, attach_datas []byte, datas []byte, sess *SSServerSession)
}

type ILogicServer interface {
	IServerMsgHandler
	OnEstablish(serversess *SSServerSession)
	OnTerminate(serversess *SSServerSession)
	SetServerSession(serversess *SSServerSession)
	GetServerSession() *SSServerSession
}

type LogicServer struct {
	server_sess *SSServerSession
}

func (l *LogicServer) SetServerSession(serversess *SSServerSession) {
	l.server_sess = serversess
}

func (l *LogicServer) GetServerSession() *SSServerSession {
	return l.server_sess
}

type ILogicServerFactory interface {
	SetLogicServer(serversess *SSServerSession)
}
