package enet

import (
	"fmt"
	"time"
)

const (
	ConnEstablishType uint32 = iota
	ConnRecvMsgType
	ConnCloseType
)

type SessionType uint32

const (
	SessConnectType SessionType = iota
	SessListenType
)

//var GSendQps int64 = 0
//var GRecvQps int64 = 0

const (
	ConnEstablishState uint32 = iota
	ConnClosedState
)

const (
	MsgSuccess uint32 = 0
	MsgFail    uint32 = 1
)

func getMillsecond() int64 {
	return time.Now().UnixNano() / 1e6
}

const NetMajorVersion = 1
const NetMinorVersion = 1

type NetVersion struct {
}

func (n *NetVersion) GetVersion() string {
	return fmt.Sprintf("Net Version: %v.%v", NetMajorVersion, NetMinorVersion)
}

var GNetVersion *NetVersion

func init() {
	GNetVersion = &NetVersion{}
}
