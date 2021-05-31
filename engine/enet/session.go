package enet

import (
	"github.com/golang/protobuf/proto"
)

type Session struct {
	ISessionOnHandler
	conn     IConnection
	sessId   uint64
	attach   interface{}
	coder    ICoder
	sessType SessionType
	factory  ISessionFactory
}

func (s *Session) SetConnection(conn IConnection) {
	s.conn = conn
}

func (s *Session) SetSessID(sessId uint64) {
	s.sessId = sessId
}

func (s *Session) GetSessID() uint64 {
	return s.sessId
}

func (s *Session) GetAttach() interface{} {
	return s.attach
}

func (s *Session) SetAttach(attach interface{}) {
	s.attach = attach
}

func (s *Session) GetCoder() ICoder {
	return s.coder
}

func (s *Session) SetCoder(coder ICoder) {
	s.coder = coder
}

func (s *Session) GetSessionOnHandler() ISessionOnHandler {
	return s.ISessionOnHandler
}

func (s *Session) IsListenType() bool {
	if s.sessType == SessListenType {
		return true
	} else {
		return false
	}
}

func (s *Session) IsConnectType() bool {
	if s.sessType == SessConnectType {
		return true
	} else {
		return false
	}
}

func (s *Session) SetConnectType() {
	s.sessType = SessConnectType
}

func (s *Session) SetListenType() {
	s.sessType = SessListenType
}

func (s *Session) SetSessionFactory(factory ISessionFactory) {
	s.factory = factory
}

func (s *Session) GetSessionFactory() ISessionFactory {
	return s.factory
}

func (s *Session) Terminate() {
	if s.conn != nil {
		s.conn.Terminate()
		ELog.InfoAf("[Session] Terminate SesssionID=%v", s.GetSessID())
	}
}

func (s *Session) AsyncSendMsg(msgId uint32, datas []byte) bool {
	if s.conn == nil {
		return false
	}

	allDatas, err := s.coder.PackMsg(msgId, datas)
	if err != nil {
		ELog.ErrorAf("[Session] SesssionID=%v  SendMsg PackMsg Error=%v", s.GetSessID(), err)
	}

	s.conn.AsyncSend(allDatas)
	return true
}

func (s *Session) AsyncSendProtoMsg(msgId uint32, msg proto.Message) bool {
	if s.conn == nil {
		return false
	}

	datas, err := proto.Marshal(msg)
	if err != nil {
		ELog.ErrorAf("[Net] Msg=%v Marshal Err %v ", msgId, err)
		return false
	}

	return s.AsyncSendMsg(msgId, datas)
}
