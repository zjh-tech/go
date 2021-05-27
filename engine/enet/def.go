package enet

import "fmt"

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
