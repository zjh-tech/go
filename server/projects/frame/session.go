package frame

import (
	"bytes"
	"encoding/binary"
	"math"
	"projects/go-engine/elog"
	"projects/go-engine/inet"

	"github.com/golang/protobuf/proto"
)

const (
	PackageMsgIDLen int = 4
)

type SessionOnHandler interface {
	OnHandler(msgID uint32, attach_datas []byte, datas []byte)
}

type Session struct {
	SessionOnHandler
	conn      inet.IConnection
	sess_id   uint64
	attach    interface{}
	coder     inet.ICoder
	sess_type inet.SessionType
	factory   inet.ISessionFactory
}

func (s *Session) SetConnection(conn inet.IConnection) {
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

func (s *Session) GetCoder() inet.ICoder {
	return s.coder
}

func (s *Session) SetCoder(coder inet.ICoder) {
	s.coder = coder
}

func (s *Session) IsListenType() bool {
	if s.sess_type == inet.SESS_LISTEN_TYPE {
		return true
	}

	return false
}

func (s *Session) IsConnectType() bool {
	if s.sess_type == inet.SESS_CONNECT_TYPE {
		return true
	}

	return false
}

func (s *Session) SetConnectType() {
	s.sess_type = inet.SESS_CONNECT_TYPE
}

func (s *Session) SetListenType() {
	s.sess_type = inet.SESS_LISTEN_TYPE
}

func (s *Session) SetSessionFactory(factory inet.ISessionFactory) {
	s.factory = factory
}

func (s *Session) GetSessionFactory() inet.ISessionFactory {
	return s.factory
}

func (s *Session) Terminate() {
	if s.conn != nil {
		s.conn.Terminate()
		elog.InfoAf("[Session] Terminate SesssionID=%v", s.GetSessID())
	}
}

func (s *Session) ProcessMsg(datas []byte) {
	if len(datas) < PackageMsgIDLen {
		elog.ErrorAf("[Session] SesssionID=%v ProcessMsg Len Error", s.GetSessID())
		return
	}

	buff := bytes.NewBuffer(datas)
	msg_id := uint32(0)
	if err := binary.Read(buff, binary.BigEndian, &msg_id); err != nil {
		elog.ErrorAf("[Session] SesssionID=%v ProcessMsg MsgID Error=%v", s.GetSessID(), err)
		return
	}

	attach_len := uint16(0)
	if err := binary.Read(buff, binary.BigEndian, &attach_len); err != nil {
		elog.ErrorAf("[Session] SesssionID=%v ProcessMsg AttachLen Error=%v", s.GetSessID(), err)
		return
	}

	if attach_len != 0 {
		attachDatas := make([]byte, attach_len)
		if err := binary.Read(buff, binary.BigEndian, &attachDatas); err != nil {
			elog.ErrorAf("[Session] SesssionID=%v ProcessMsg AttachData Error=%v", s.GetSessID(), err)
			return
		}
		msg_start_index := PackageMsgIDLen + 2 + int(attach_len)
		s.OnHandler(msg_id, attachDatas, datas[msg_start_index:])
	} else {
		msg_start_index := PackageMsgIDLen + 2
		s.OnHandler(msg_id, nil, datas[msg_start_index:])
	}

}

func (s *Session) AsyncSendMsg(msgID uint32, datas []byte, attach inet.IAttachParas) bool {
	if s.conn == nil {
		return false
	}

	buff := bytes.NewBuffer([]byte{})
	if err := binary.Write(buff, binary.BigEndian, msgID); err != nil {
		elog.ErrorAf("[Session] SesssionID=%v  AsyncSendMsg MsgID Error=%v", s.GetSessID(), err)
		return false
	}

	if attach != nil {
		attach_datas := attach.FillNetStream()
		if len(attach_datas) > math.MaxUint16 {
			elog.ErrorAf("[Session] SesssionID=%v  AsyncSendMsg Attach Len Cross Range", s.GetSessID())
			return false
		}

		if err := binary.Write(buff, binary.BigEndian, uint16(len(attach_datas))); err != nil {
			elog.ErrorAf("[Session] SesssionID=%v  AsyncSendMsg Attach Len Error=%v", s.GetSessID(), err)
			return false
		}

		if err := binary.Write(buff, binary.BigEndian, attach_datas); err != nil {
			elog.ErrorAf("[Session] SesssionID=%v  AsyncSendMsg Attach Datas Error=%v", s.GetSessID(), err)
			return false
		}
	} else {
		if err := binary.Write(buff, binary.BigEndian, uint16(0)); err != nil {
			elog.ErrorAf("[Session] SesssionID=%v  AsyncSendMsg Attach Len Error=%v", s.GetSessID(), err)
			return false
		}
	}

	if datas != nil {
		if err := binary.Write(buff, binary.BigEndian, datas); err != nil {
			elog.ErrorAf("[Session] SesssionID=%v  AsyncSendMsg  Datas Error=%v", s.GetSessID(), err)
			return false
		}
	}

	all_datas, err := s.coder.FillNetStream(buff.Bytes())
	if err != nil {
		elog.ErrorAf("[Session] SesssionID=%v  AsyncSendMsg FillNetStream Error=%v", s.GetSessID(), err)
	}

	s.conn.AsyncSend(all_datas)
	return true
}

func (s *Session) AsyncSendProtoMsg(msgID uint32, msg proto.Message, attach inet.IAttachParas) bool {
	if s.conn == nil {
		return false
	}

	datas, err := proto.Marshal(msg)
	if err != nil {
		elog.ErrorAf("[Net] Msg=%v Marshal Err %v ", msgID, err)
		return false
	}

	return s.AsyncSendMsg(msgID, datas, attach)
}
