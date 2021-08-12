package enet

import (
	"bytes"
	"encoding/binary"
	"errors"
)

const (
	PackageHeaderLen uint32 = 4
	PackageMsgIDLen  int    = 4
)

type MsgHeader struct {
	BodyLen uint32
}

type Coder struct {
	msgHeader MsgHeader
}

func NewCoder() *Coder {
	return &Coder{}
}

func (c *Coder) GetHeaderLen() uint32 {
	return PackageHeaderLen
}

func (c *Coder) GetBodyLen(datas []byte) (uint32, error) {
	if uint32(len(datas)) < PackageHeaderLen {
		return 0, errors.New("Body Len Not Enough")
	}

	buff := bytes.NewBuffer(datas)
	if err := binary.Read(buff, binary.BigEndian, &c.msgHeader.BodyLen); err != nil {
		return 0, err
	}

	return c.msgHeader.BodyLen, nil
}

func (c *Coder) UnpackMsg(datas []byte) ([]byte, error) {
	return datas, nil
}

func (c *Coder) ProcessMsg(datas []byte, sess ISession) {
	if len(datas) < PackageMsgIDLen {
		ELog.ErrorAf("[Session] SesssionID=%v ProcessMsg Len Error", sess.GetSessID())
		return
	}

	buff := bytes.NewBuffer(datas)
	msgId := uint32(0)
	if err := binary.Read(buff, binary.BigEndian, &msgId); err != nil {
		ELog.ErrorAf("[Session] SesssionID=%v ProcessMsg MsgID Error=%v", sess.GetSessID(), err)
		return
	}

	ELog.DebugAf("ConnID=%v,SessionID=%v,MsgID=%v", msgId)
	msgStartIndex := PackageMsgIDLen + 2
	sess.GetSessionOnHandler().OnHandler(msgId, datas[msgStartIndex:])
}

func (c *Coder) PackMsg(msgId uint32, datas []byte) ([]byte, error) {
	bodyBuff := bytes.NewBuffer([]byte{})
	if err := binary.Write(bodyBuff, binary.BigEndian, msgId); err != nil {
		ELog.ErrorAf("PackMsg MsgID Error=%v", err)
	}

	if datas != nil {
		if err := binary.Write(bodyBuff, binary.BigEndian, datas); err != nil {
			ELog.ErrorAf("PackMsg Datas Error=%v", err)
		}
	}

	bodyBytes := bodyBuff.Bytes()
	header := &MsgHeader{}
	header.BodyLen = uint32(len(bodyBytes))

	buff := bytes.NewBuffer([]byte{})
	if err := binary.Write(buff, binary.BigEndian, header.BodyLen); err != nil {
		return nil, err
	}
	if err := binary.Write(buff, binary.BigEndian, bodyBytes); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}
