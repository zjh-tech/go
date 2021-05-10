package frame

import (
	"bytes"
	"encoding/binary"

	"github.com/zjh-tech/go-frame/engine/enet"

	"github.com/golang/protobuf/proto"
)

const (
	PackageMsgIDLen int = 4
)

type SessionOnHandler interface {
	OnHandler(msgID uint32, datas []byte)
}

type Session struct {
	SessionOnHandler
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

func (s *Session) IsListenType() bool {
	if s.sess_type == enet.SESS_LISTEN_TYPE {
		return true
	}

	return false
}

func (s *Session) IsConnectType() bool {
	if s.sess_type == enet.SESS_CONNECT_TYPE {
		return true
	}

	return false
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

func (s *Session) ProcessMsg(datas []byte) {
	if len(datas) < PackageMsgIDLen {
		ELog.ErrorAf("[Session] SesssionID=%v ProcessMsg Len Error", s.GetSessID())
		return
	}

	buff := bytes.NewBuffer(datas)
	msg_id := uint32(0)
	if err := binary.Read(buff, binary.BigEndian, &msg_id); err != nil {
		ELog.ErrorAf("[Session] SesssionID=%v ProcessMsg MsgID Error=%v", s.GetSessID(), err)
		return
	}

	ELog.DebugAf("ConnID=%v,SessionID=%v,MsgID=%v", msg_id)
	msg_start_index := PackageMsgIDLen + 2
	s.OnHandler(msg_id, datas[msg_start_index:])
}

func (s *Session) AsyncSendMsg(msgID uint32, datas []byte) bool {
	if s.conn == nil {
		return false
	}

	buff := bytes.NewBuffer([]byte{})
	if err := binary.Write(buff, binary.BigEndian, msgID); err != nil {
		ELog.ErrorAf("[Session] SesssionID=%v  AsyncSendMsg MsgID Error=%v", s.GetSessID(), err)
		return false
	}

	if err := binary.Write(buff, binary.BigEndian, uint16(0)); err != nil {
		ELog.ErrorAf("[Session] SesssionID=%v  AsyncSendMsg Attach Len Error=%v", s.GetSessID(), err)
		return false
	}

	if datas != nil {
		if err := binary.Write(buff, binary.BigEndian, datas); err != nil {
			ELog.ErrorAf("[Session] SesssionID=%v  AsyncSendMsg  Datas Error=%v", s.GetSessID(), err)
			return false
		}
	}

	all_datas, err := s.coder.FillNetStream(buff.Bytes())
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
