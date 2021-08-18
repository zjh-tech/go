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

const ConnWriterSleepLoopCount = 10000

const (
	NetChannelMaxSize     = 10000000
	NetMaxConnectSize     = 60000
	ConnectChannelMaxSize = 1000000
	PackageDefaultMaxSize = 1024 * 64
)

const (
	IsFreeWriteState uint32 = 0
	IsWritingState   uint32 = 1
)

func getMillsecond() int64 {
	return time.Now().UnixNano() / 1e6
}

func getSecond() int64 {
	return time.Now().UnixNano() / 1e9
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
