package common

import (
	"fmt"
	"projects/config"
	"projects/go-engine/elog"
	"projects/pb"
)

func SendTipMsg(obj GameObject, tipId uint32, v ...interface{}) {
	tipInfo := config.GConfigMgr.GetTipByID(tipId)
	if tipInfo == nil {
		elog.WarnAf("[Tip] TipId=%v Not Find", tipId)
		return
	}

	ntf := pb.ScTipNtf{}
	ntf.Content = fmt.Sprintf(tipInfo.Str, v...)
	obj.Send2Client(uint32(pb.EClient2GameMsgId_sc_tip_ntf_id), &ntf)
}
