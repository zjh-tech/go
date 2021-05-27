package frame

// import (
// 	"bytes"
// 	"encoding/binary"
// 	"io/ioutil"
// 	"net/http"
// 	"time"

// 	"github.com/zjh-tech/go-frame/base/util"
// 	"github.com/zjh-tech/go-frame/engine/enet"

// 	"github.com/golang/protobuf/proto"
// )

// func SendSingleHttpReq(url string, msgId uint32, msg proto.Message) {
// 	datas, err := proto.Marshal(msg)
// 	if err != nil {
// 		ELog.ErrorAf("[Net] SendSingleHttpReq Msg=%v Marshal Err %v ", msgId, err)
// 		return
// 	}
// 	go send_http_req(url, msgId, datas, false, false)
// }

// func SendMultiHttpReq(url string, msgId uint32, msg proto.Message) {
// 	datas, err := proto.Marshal(msg)
// 	if err != nil {
// 		ELog.ErrorAf("[Net] SendMultiHttpReq Msg=%v Marshal Err %v ", msgId, err)
// 		return
// 	}
// 	go send_http_req(url, msgId, datas, true, false)
// }

// func SendSDHttpReq(url string, msgId uint32, msg proto.Message) {
// 	datas, err := proto.Marshal(msg)
// 	if err != nil {
// 		ELog.ErrorAf("[Net] SendSDHttpReq Msg=%v Marshal Err %v ", msgId, err)
// 		return
// 	}
// 	go send_http_req(url, msgId, datas, false, true)
// }

// func send_http_req(url string, msgId uint32, datas []byte, multiFlag bool, sd_flag bool) {
// 	buff := bytes.NewBuffer([]byte{})
// 	binary.Write(buff, binary.BigEndian, msgId)
// 	binary.Write(buff, binary.BigEndian, datas)

// 	client := &http.Client{}
// 	client.Timeout = time.Second

// 	resp, res_err := client.Post(url, "application/octet-stream", buff)
// 	if res_err != nil {
// 		ELog.ErrorAf("Http Post Url=%v ResErr=%v", url, res_err)
// 		return
// 	}

// 	defer resp.Body.Close()
// 	body, body_err := ioutil.ReadAll(resp.Body)
// 	if body_err != nil {
// 		ELog.ErrorAf("Http Url=%v BodyErr=%v", url, body_err)
// 		return
// 	}

// 	ack_msgId_len := 4
// 	if len(body) > ack_msgId_len {
// 		ack_msgId := util.NetBytesToUint32(body)
// 		if sd_flag {
// 			http_event := enet.NewHttpEvent(GSDHttpServerSession, ack_msgId, body[ack_msgId_len:])
// 			enet.GNet.PushSingleHttpEvent(http_event)
// 		} else {
// 			if multiFlag {
// 				http_event := enet.NewHttpEvent(GMultiHttpConnection, ack_msgId, body[ack_msgId_len:])
// 				enet.GNet.PushMultiHttpEvent(http_event)
// 			} else {
// 				http_event := enet.NewHttpEvent(GSingleHttpConnection, ack_msgId, body[ack_msgId_len:])
// 				enet.GNet.PushSingleHttpEvent(http_event)
// 			}
// 		}
// 	}
// }

// var GSDHttpServerSession *SDHttpServerSession  //SD
// var GSingleHttpConnection enet.IHttpConnection //Logic Single
// var GMultiHttpConnection enet.IHttpConnection  //Logic Multi

// func init() {
// 	GSDHttpServerSession = &SDHttpServerSession{
// 		dealer: NewIDDealer(),
// 	}

// 	GSDHttpServerSession.Init()
// }
