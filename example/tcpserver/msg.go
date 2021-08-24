package main

const (
	C2STestReqMsgId = 1001
	C2STestResMsgId = 1002
)

type C2STestPressureReq struct {
	StrValue    string `json:"strvalue"`
	Uint64Value uint64 `json:"value"`
}

type C2STestPressureRes struct {
	StrValue    string `json:"strvalue"`
	Uint64Value uint64 `json:"value"`
}
