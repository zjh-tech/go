package common

import "github.com/golang/protobuf/proto"

type GameObject interface {
	Send2Client(msgID uint32, msg proto.Message) bool
}
