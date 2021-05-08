package frame

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"net/http"
	"projects/base/util"
	"projects/engine/enet"
	"time"

	"github.com/golang/protobuf/proto"
)

func SendSingleHttpReq(url string, msg_id uint32, msg proto.Message) {
	datas, err := proto.Marshal(msg)
	if err != nil {
		ELog.ErrorAf("[Net] SendSingleHttpReq Msg=%v Marshal Err %v ", msg_id, err)
		return
	}
	go send_http_req(url, msg_id, datas, false, false)
}

func SendMultiHttpReq(url string, msg_id uint32, msg proto.Message) {
	datas, err := proto.Marshal(msg)
	if err != nil {
		ELog.ErrorAf("[Net] SendMultiHttpReq Msg=%v Marshal Err %v ", msg_id, err)
		return
	}
	go send_http_req(url, msg_id, datas, true, false)
}

func SendSDHttpReq(url string, msg_id uint32, msg proto.Message) {
	datas, err := proto.Marshal(msg)
	if err != nil {
		ELog.ErrorAf("[Net] SendSDHttpReq Msg=%v Marshal Err %v ", msg_id, err)
		return
	}
	go send_http_req(url, msg_id, datas, false, true)
}

func send_http_req(url string, msg_id uint32, datas []byte, multi_flag bool, sd_flag bool) {
	buff := bytes.NewBuffer([]byte{})
	binary.Write(buff, binary.BigEndian, msg_id)
	binary.Write(buff, binary.BigEndian, datas)

	client := &http.Client{}
	client.Timeout = time.Second

	resp, res_err := client.Post(url, "application/octet-stream", buff)
	if res_err != nil {
		ELog.ErrorAf("Http Post Url=%v ResErr=%v", url, res_err)
		return
	}

	defer resp.Body.Close()
	body, body_err := ioutil.ReadAll(resp.Body)
	if body_err != nil {
		ELog.ErrorAf("Http Url=%v BodyErr=%v", url, body_err)
		return
	}

	ack_msg_id_len := 4
	if len(body) > ack_msg_id_len {
		ack_msg_id := util.NetBytesToUint32(body)
		if sd_flag {
			http_event := enet.NewHttpEvent(GSDHttpServerSession, ack_msg_id, body[ack_msg_id_len:])
			enet.GNet.PushSingleHttpEvent(http_event)
		} else {
			if multi_flag {
				http_event := enet.NewHttpEvent(GMultiHttpConnection, ack_msg_id, body[ack_msg_id_len:])
				enet.GNet.PushMultiHttpEvent(http_event)
			} else {
				http_event := enet.NewHttpEvent(GSingleHttpConnection, ack_msg_id, body[ack_msg_id_len:])
				enet.GNet.PushSingleHttpEvent(http_event)
			}
		}
	}
}

var GSDHttpServerSession *SDHttpServerSession  //SD
var GSingleHttpConnection enet.IHttpConnection //Logic Single
var GMultiHttpConnection enet.IHttpConnection  //Logic Multi

func init() {
	GSDHttpServerSession = &SDHttpServerSession{
		dealer: NewIDDealer(),
	}

	GSDHttpServerSession.Init()
}
