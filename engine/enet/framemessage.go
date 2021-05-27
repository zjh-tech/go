package enet

const (
	S2SEngineMinMsgId = 1

	S2SSessionVerifyReqId = 2
	S2SSessionVerifyResId = 3
	S2SSessionPingId      = 4
	S2SSessionPongId      = 5

	S2SEngineMaxMsgId = 100
)

type S2SSessionVerifyReq struct {
	ServerId      uint64
	ServerType    uint32
	ServerTypeStr string
	Ip            string
	Token         string
}

type S2SSessionVerifyRes struct {
}

type S2SSessionPing struct {
}

type S2SSessionPong struct {
}

//--------c2s--------------
const (
	C2SEngineMinMsgId = 1

	C2SSessionPingId = 2
	C2SSessionPongId = 3

	C2SEngineMaxMsgId = 100
)

type C2SSessionPing struct {
}

type C2SSessionPong struct {
}

//--------sdk--------------
const (
	SDkEngineMinMsgId = 1

	SDKSessionPingId = 2
	SDKSessionPongId = 3

	SDkEngineMaxMsgId = 100
)

type SDKSessionPing struct {
}

type SDKSessionPong struct {
}
