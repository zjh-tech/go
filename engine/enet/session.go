package enet

import (
	"encoding/json"

	"github.com/golang/protobuf/proto"
)

type Session struct {
	ISessionOnHandler
	conn                  IConnection
	sessId                uint64
	attach                interface{}
	coder                 ICoder
	sessType              SessionType
	factory               ISessionFactory
	evtQueue              IEventQueue
	sessionConcurrentFlag bool
	exitChan              chan struct{}
}

func (s *Session) SetSessionConcurrentFlag(flag bool) {
	s.sessionConcurrentFlag = flag
	if s.sessionConcurrentFlag {
		s.evtQueue = newEventQueue(NetChannelMaxSize)
	}
}

func (s *Session) GetSessionConcurrentFlag() bool {
	return s.sessionConcurrentFlag
}

func (s *Session) PushEvent(evt IEvent) {
	s.evtQueue.PushEvent(evt)
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

func (s *Session) StartSessionConcurrentGoroutine() {
	connID := s.conn.GetConnID()
	s.exitChan = make(chan struct{})
	ELog.InfoAf("[Net][Session] SessID=%v ConnID=%v ProcessMsg Goroutine Start", s.sessId, connID)

	go func() {
		defer func() {
			ELog.InfoAf("[Net][Session] SessID=%v ConnID=%v ProcessMsg Goroutine Exit", s.sessId, connID)
			closeEvent := NewTcpEvent(ConnCloseType, s.conn, nil)
			closeEvent.ProcessMsg()
		}()

		for {
			select {
			case evt, ok := <-s.evtQueue.GetEventQueue():
				if !ok {
					return
				}
				tcpEvt := evt.(*TcpEvent)
				tcpEvt.ProcessMsg()
			case <-s.exitChan:
				return
			}
		}
	}()
}

func (s *Session) StopSessionConcurrentGoroutine() {
	s.exitChan <- struct{}{}
}

func (s *Session) Terminate() {
	if s.conn != nil {
		s.conn.Terminate()
		ELog.InfoAf("[Session] Terminate SesssionID=%v", s.GetSessID())
	}
}

func (s *Session) SendMsg(msgId uint32, datas []byte) bool {
	if s.conn == nil {
		return false
	}

	allDatas, err := s.coder.PackMsg(msgId, datas)
	if err != nil {
		ELog.ErrorAf("[Session] SesssionID=%v  SendMsg PackMsg Error=%v", s.GetSessID(), err)
		return false
	}

	if len(allDatas) >= int(s.coder.GetPackageMaxLen()) {
		ELog.ErrorAf("[Session] SesssionID=%v SendMsg MsgId=%v Out Range PackMsg Max Len", s.GetSessID(), msgId)
		return false
	}

	ELog.DebugAf("[Net][Session] SendMsg MsgId=%v,Datas=%v", msgId, datas)
	s.conn.AsyncSend(allDatas)
	return true
}

func (s *Session) SendProtoMsg(msgId uint32, msg proto.Message) bool {
	if s.conn == nil {
		return false
	}

	datas, err := proto.Marshal(msg)
	if err != nil {
		ELog.ErrorAf("[Net] Msg=%v Proto.Marshal Err %v ", msgId, err)
		return false
	}

	ELog.DebugAf("[Net][Session] SendProtoMsg MsgId=%v,Protobuf=%v", msgId, msg)
	return s.SendMsg(msgId, datas)
}

func (s *Session) SendJsonMsg(msgId uint32, js interface{}) bool {
	if s.conn == nil {
		return false
	}

	datas, err := json.Marshal(js)
	if err != nil {
		ELog.ErrorAf("[Net] Msg=%v Json.Marshal Err %v ", msgId, err)
		return false
	}

	ELog.DebugAf("[Net][Session] SendJsonMsg MsgId=%v,Json=%v", msgId, js)
	return s.SendMsg(msgId, datas)
}
