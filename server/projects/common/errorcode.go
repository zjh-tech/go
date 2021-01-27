package common

import (
	"fmt"
	"projects/pb"
)

func SendErrorCodeMsg(obj GameObject, errorCode pb.EScErrorCode, v ...interface{}) {
	ntf := pb.ScErrorcodeNtf{}
	ntf.Errorcode = errorCode
	parasLen := len(v)
	for i := 0; i < parasLen; i++ {
		para := fmt.Sprintf("%v", v[i])
		ntf.Paras = append(ntf.Paras, para)
	}
	obj.Send2Client(uint32(pb.EClient2GameMsgId_sc_tip_ntf_id), &ntf)
}
