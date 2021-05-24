package frame

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/zjh-tech/go-frame/engine/enet"
)

const (
	PackageHeaderLen uint32 = 6 //MsgHeader(bool + bool + uint32)
)

type MsgHeader struct {
	EncodeFlag bool
	ZipFlag    bool
	BodyLen    uint32
}

type Coder struct {
	msg_header MsgHeader
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
	if err := binary.Read(buff, binary.BigEndian, &c.msg_header.EncodeFlag); err != nil {
		return 0, err
	}
	if err := binary.Read(buff, binary.BigEndian, &c.msg_header.ZipFlag); err != nil {
		return 0, err
	}
	if err := binary.Read(buff, binary.BigEndian, &c.msg_header.BodyLen); err != nil {
		return 0, err
	}

	return c.msg_header.BodyLen, nil
}

func (c *Coder) EnCodeBody(datas []byte) ([]byte, bool) {
	return datas, false
}

func (c *Coder) DecodeBody(datas []byte) ([]byte, error) {
	if c.msg_header.EncodeFlag == false {
		return datas, nil
	}

	return nil, errors.New("DecodeBody Error")
}

func (c *Coder) ZipBody(datas []byte) ([]byte, bool) {
	return datas, false
}

func (c *Coder) UnzipBody(datas []byte) ([]byte, error) {
	if c.msg_header.EncodeFlag == false {
		return datas, nil
	}

	return nil, errors.New("UnzipBody Error")
}

func (c *Coder) ProcessMsg(datas []byte, sess enet.ISession) {
	if len(datas) < PackageMsgIDLen {
		ELog.ErrorAf("[Session] SesssionID=%v ProcessMsg Len Error", sess.GetSessID())
		return
	}

	buff := bytes.NewBuffer(datas)
	msg_id := uint32(0)
	if err := binary.Read(buff, binary.BigEndian, &msg_id); err != nil {
		ELog.ErrorAf("[Session] SesssionID=%v ProcessMsg MsgID Error=%v", sess.GetSessID(), err)
		return
	}

	ELog.DebugAf("ConnID=%v,SessionID=%v,MsgID=%v", msg_id)
	msg_start_index := PackageMsgIDLen + 2
	sess.GetSessionOnHandler().OnHandler(msg_id, datas[msg_start_index:])
}

func (c *Coder) FillNetStream(msgID uint32, datas []byte) ([]byte, error) {
	encodeDatas, encodeflag := c.EnCodeBody(datas)
	zipDatas, zipflag := c.ZipBody(encodeDatas)

	header := &MsgHeader{}
	header.EncodeFlag = encodeflag
	header.ZipFlag = zipflag
	header.BodyLen = uint32(len(zipDatas))

	buff := bytes.NewBuffer([]byte{})
	if err := binary.Write(buff, binary.BigEndian, header.EncodeFlag); err != nil {
		return nil, err
	}
	if err := binary.Write(buff, binary.BigEndian, header.ZipFlag); err != nil {
		return nil, err
	}
	if err := binary.Write(buff, binary.BigEndian, header.BodyLen); err != nil {
		return nil, err
	}
	if err := binary.Write(buff, binary.BigEndian, zipDatas); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}
