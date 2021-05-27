package frame

import (
	"github.com/golang/protobuf/proto"
	"github.com/zjh-tech/go-frame/engine/enet"
)

const (
	PackageMsgIDLen int = 4
)

type Session struct {
	enet.ISessionOnHandler
	conn      enet.IConnection
	sess_id   uint64
	attach    interface{}
	coder     enet.ICoder
	sess_type enet.SessionType
	factory   enet.ISessionFactory
}

func (s *Session) SetConnection(conn enet.IConnection) {
	s.conn = conn
}

func (s *Session) SetSessID(sess_id uint64) {
	s.sess_id = sess_id
}

func (s *Session) GetSessID() uint64 {
	return s.sess_id
}

func (s *Session) GetAttach() interface{} {
	return s.attach
}

func (s *Session) SetAttach(attach interface{}) {
	s.attach = attach
}

func (s *Session) GetCoder() enet.ICoder {
	return s.coder
}

func (s *Session) SetCoder(coder enet.ICoder) {
	s.coder = coder
}

func (s *Session) GetSessionOnHandler() enet.ISessionOnHandler {
	return s.ISessionOnHandler
}

func (s *Session) IsListenType() bool {
	if s.sess_type == enet.SESS_LISTEN_TYPE {
		return true
	} else {
		return false
	}
}

func (s *Session) IsConnectType() bool {
	if s.sess_type == enet.SESS_CONNECT_TYPE {
		return true
	} else {
		return false
	}
}

func (s *Session) SetConnectType() {
	s.sess_type = enet.SESS_CONNECT_TYPE
}

func (s *Session) SetListenType() {
	s.sess_type = enet.SESS_LISTEN_TYPE
}

func (s *Session) SetSessionFactory(factory enet.ISessionFactory) {
	s.factory = factory
}

func (s *Session) GetSessionFactory() enet.ISessionFactory {
	return s.factory
}

func (s *Session) Terminate() {
	if s.conn != nil {
		s.conn.Terminate()
		ELog.InfoAf("[Session] Terminate SesssionID=%v", s.GetSessID())
	}
}

func (s *Session) AsyncSendMsg(msgID uint32, datas []byte) bool {
	if s.conn == nil {
		return false
	}

	all_datas, err := s.coder.FillNetStream(msgID, datas)
	if err != nil {
		ELog.ErrorAf("[Session] SesssionID=%v  AsyncSendMsg FillNetStream Error=%v", s.GetSessID(), err)
	}

	s.conn.AsyncSend(all_datas)
	return true
}

func (s *Session) AsyncSendProtoMsg(msgID uint32, msg proto.Message) bool {
	if s.conn == nil {
		return false
	}

	datas, err := proto.Marshal(msg)
	if err != nil {
		ELog.ErrorAf("[Net] Msg=%v Marshal Err %v ", msgID, err)
		return false
	}

	return s.AsyncSendMsg(msgID, datas)
}
