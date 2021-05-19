// Code generated by protoc-gen-go.
// source: slb.proto
// DO NOT EDIT!

/*
Package slbpb is a generated protocol buffer package.

It is generated from these files:
	slb.proto

It has these top-level messages:
	SdServerSpec
	S2LbServiceDiscoveryReq
	Lb2SServiceDiscoveryAck
	S2LbSelectMinServerReq
	Lb2SSelectMinServerAck
*/
package slbpb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
const _ = proto.ProtoPackageIsVersion1

type S2LBMsgID int32

const (
	S2LBMsgID_s2lb_invalid_msg_id           S2LBMsgID = 0
	S2LBMsgID_s2lb_service_discovery_req_id S2LBMsgID = 1
	S2LBMsgID_lb2s_service_discovery_ack_id S2LBMsgID = 2
	S2LBMsgID_s2lb_select_min_server_req_id S2LBMsgID = 3
	S2LBMsgID_lb2s_select_min_server_ack_id S2LBMsgID = 4
)

var S2LBMsgID_name = map[int32]string{
	0: "s2lb_invalid_msg_id",
	1: "s2lb_service_discovery_req_id",
	2: "lb2s_service_discovery_ack_id",
	3: "s2lb_select_min_server_req_id",
	4: "lb2s_select_min_server_ack_id",
}
var S2LBMsgID_value = map[string]int32{
	"s2lb_invalid_msg_id":           0,
	"s2lb_service_discovery_req_id": 1,
	"lb2s_service_discovery_ack_id": 2,
	"s2lb_select_min_server_req_id": 3,
	"lb2s_select_min_server_ack_id": 4,
}

func (x S2LBMsgID) String() string {
	return proto.EnumName(S2LBMsgID_name, int32(x))
}
func (S2LBMsgID) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type SdServerSpec struct {
	ServerId       uint64 `protobuf:"varint,1,opt,name=server_id,json=serverId" json:"server_id,omitempty"`
	ServerType     uint32 `protobuf:"varint,2,opt,name=server_type,json=serverType" json:"server_type,omitempty"`
	Token          string `protobuf:"bytes,3,opt,name=token" json:"token,omitempty"`
	S2SInterListen string `protobuf:"bytes,4,opt,name=s2s_inter_listen,json=s2sInterListen" json:"s2s_inter_listen,omitempty"`
	S2SOuterListen string `protobuf:"bytes,5,opt,name=s2s_outer_listen,json=s2sOuterListen" json:"s2s_outer_listen,omitempty"`
	State          uint32 `protobuf:"varint,6,opt,name=state" json:"state,omitempty"`
}

func (m *SdServerSpec) Reset()                    { *m = SdServerSpec{} }
func (m *SdServerSpec) String() string            { return proto.CompactTextString(m) }
func (*SdServerSpec) ProtoMessage()               {}
func (*SdServerSpec) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type S2LbServiceDiscoveryReq struct {
	SdServerInfo *SdServerSpec `protobuf:"bytes,1,opt,name=sd_server_info,json=sdServerInfo" json:"sd_server_info,omitempty"`
}

func (m *S2LbServiceDiscoveryReq) Reset()                    { *m = S2LbServiceDiscoveryReq{} }
func (m *S2LbServiceDiscoveryReq) String() string            { return proto.CompactTextString(m) }
func (*S2LbServiceDiscoveryReq) ProtoMessage()               {}
func (*S2LbServiceDiscoveryReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *S2LbServiceDiscoveryReq) GetSdServerInfo() *SdServerSpec {
	if m != nil {
		return m.SdServerInfo
	}
	return nil
}

type Lb2SServiceDiscoveryAck struct {
	SdServerList []*SdServerSpec `protobuf:"bytes,1,rep,name=sd_server_list,json=sdServerList" json:"sd_server_list,omitempty"`
}

func (m *Lb2SServiceDiscoveryAck) Reset()                    { *m = Lb2SServiceDiscoveryAck{} }
func (m *Lb2SServiceDiscoveryAck) String() string            { return proto.CompactTextString(m) }
func (*Lb2SServiceDiscoveryAck) ProtoMessage()               {}
func (*Lb2SServiceDiscoveryAck) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *Lb2SServiceDiscoveryAck) GetSdServerList() []*SdServerSpec {
	if m != nil {
		return m.SdServerList
	}
	return nil
}

type S2LbSelectMinServerReq struct {
	ServerType uint32 `protobuf:"varint,1,opt,name=server_type,json=serverType" json:"server_type,omitempty"`
}

func (m *S2LbSelectMinServerReq) Reset()                    { *m = S2LbSelectMinServerReq{} }
func (m *S2LbSelectMinServerReq) String() string            { return proto.CompactTextString(m) }
func (*S2LbSelectMinServerReq) ProtoMessage()               {}
func (*S2LbSelectMinServerReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

type Lb2SSelectMinServerAck struct {
	ServerType uint32        `protobuf:"varint,1,opt,name=server_type,json=serverType" json:"server_type,omitempty"`
	Errorcode  uint32        `protobuf:"varint,2,opt,name=errorcode" json:"errorcode,omitempty"`
	MinServer  *SdServerSpec `protobuf:"bytes,3,opt,name=min_server,json=minServer" json:"min_server,omitempty"`
}

func (m *Lb2SSelectMinServerAck) Reset()                    { *m = Lb2SSelectMinServerAck{} }
func (m *Lb2SSelectMinServerAck) String() string            { return proto.CompactTextString(m) }
func (*Lb2SSelectMinServerAck) ProtoMessage()               {}
func (*Lb2SSelectMinServerAck) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *Lb2SSelectMinServerAck) GetMinServer() *SdServerSpec {
	if m != nil {
		return m.MinServer
	}
	return nil
}

func init() {
	proto.RegisterType((*SdServerSpec)(nil), "slbpb.sd_server_spec")
	proto.RegisterType((*S2LbServiceDiscoveryReq)(nil), "slbpb.s2lb_service_discovery_req")
	proto.RegisterType((*Lb2SServiceDiscoveryAck)(nil), "slbpb.lb2s_service_discovery_ack")
	proto.RegisterType((*S2LbSelectMinServerReq)(nil), "slbpb.s2lb_select_min_server_req")
	proto.RegisterType((*Lb2SSelectMinServerAck)(nil), "slbpb.lb2s_select_min_server_ack")
	proto.RegisterEnum("slbpb.S2LBMsgID", S2LBMsgID_name, S2LBMsgID_value)
}

var fileDescriptor0 = []byte{
	// 382 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x8c, 0x93, 0xbd, 0x4e, 0xeb, 0x30,
	0x1c, 0xc5, 0x6f, 0xfa, 0xa5, 0x1b, 0xf7, 0xde, 0xaa, 0x32, 0x20, 0x22, 0x3e, 0x04, 0x64, 0xaa,
	0x18, 0x3a, 0x04, 0x36, 0xc4, 0x82, 0x58, 0x2a, 0x15, 0x21, 0xa5, 0x2c, 0x4c, 0x56, 0x13, 0x9b,
	0xca, 0x6a, 0x1a, 0x07, 0xdb, 0xad, 0xd4, 0xc7, 0xe0, 0x4d, 0x78, 0x12, 0x9e, 0x89, 0xbf, 0x6d,
	0xda, 0xa6, 0xa5, 0x15, 0x6c, 0xf1, 0xdf, 0x3f, 0x9f, 0x9c, 0x73, 0x9c, 0x20, 0x5f, 0x65, 0x49,
	0xb7, 0x90, 0x42, 0x0b, 0x5c, 0x87, 0xc7, 0x22, 0x09, 0x3f, 0x3c, 0xd4, 0x52, 0x94, 0x28, 0x26,
	0x67, 0x4c, 0x12, 0x55, 0xb0, 0x14, 0x1f, 0x03, 0xe6, 0x96, 0x9c, 0x06, 0xde, 0xb9, 0xd7, 0xa9,
	0xc5, 0x7f, 0xdd, 0xa0, 0x47, 0xf1, 0x19, 0x6a, 0x7e, 0x6d, 0xea, 0x79, 0xc1, 0x82, 0x0a, 0x6c,
	0xff, 0x8f, 0x91, 0x1b, 0x3d, 0xc1, 0x04, 0xef, 0xa3, 0xba, 0x16, 0x63, 0x96, 0x07, 0x55, 0xd8,
	0xf2, 0x63, 0xb7, 0xc0, 0x1d, 0xd4, 0x56, 0x91, 0x22, 0x3c, 0xd7, 0x70, 0x32, 0xe3, 0x4a, 0x03,
	0x50, 0xb3, 0x40, 0x0b, 0xe6, 0x3d, 0x33, 0xee, 0xdb, 0xe9, 0x82, 0x14, 0xd3, 0x12, 0x59, 0x5f,
	0x92, 0x8f, 0xd3, 0x15, 0x09, 0x6f, 0x52, 0x7a, 0xa8, 0x59, 0xd0, 0xb0, 0x26, 0xdc, 0x22, 0x7c,
	0x46, 0x47, 0x2a, 0xca, 0x12, 0x9b, 0x88, 0xa7, 0x8c, 0x50, 0xae, 0x52, 0x01, 0xe6, 0xe6, 0x44,
	0xb2, 0x57, 0x7c, 0x53, 0x4e, 0xcb, 0xf3, 0x17, 0x61, 0x03, 0x36, 0xa3, 0x83, 0xae, 0xad, 0xa3,
	0xbb, 0x5e, 0x45, 0xfc, 0x4f, 0xd1, 0x81, 0x4b, 0x0e, 0xa8, 0x91, 0xce, 0x12, 0xf0, 0xf6, 0x5d,
	0x7a, 0x98, 0x8e, 0xd7, 0xa5, 0x8d, 0x71, 0x90, 0xae, 0xfe, 0x42, 0xda, 0xa4, 0x09, 0x6f, 0x97,
	0xae, 0x33, 0x96, 0x6a, 0x32, 0xe1, 0xf9, 0x02, 0x36, 0xae, 0x37, 0x4a, 0xf7, 0x36, 0x4b, 0x0f,
	0xdf, 0xbc, 0xa5, 0xb5, 0xcd, 0xf3, 0xc6, 0xda, 0x4f, 0xe7, 0xf1, 0x09, 0xf2, 0x99, 0x94, 0x42,
	0xa6, 0x82, 0x2e, 0xee, 0x74, 0x35, 0xc0, 0xd7, 0x08, 0xad, 0x04, 0xed, 0xbd, 0xee, 0x4c, 0xe5,
	0x03, 0xe8, 0x62, 0x5d, 0xbe, 0x7b, 0xc8, 0x1f, 0x44, 0xfd, 0xbb, 0x07, 0x35, 0xea, 0xdd, 0xe3,
	0x43, 0xb4, 0x67, 0x03, 0xf2, 0x7c, 0x36, 0xcc, 0x38, 0x25, 0x13, 0x35, 0x82, 0xcf, 0xab, 0xfd,
	0x07, 0x5f, 0xa0, 0xd3, 0xdd, 0xf7, 0x65, 0x10, 0xcf, 0x20, 0xbb, 0x7b, 0x37, 0x48, 0xa5, 0xa4,
	0xb2, 0xa5, 0x3f, 0x83, 0x54, 0x4b, 0x2a, 0x5b, 0x2a, 0x32, 0x48, 0x2d, 0x69, 0xd8, 0x5f, 0xe3,
	0xea, 0x33, 0x00, 0x00, 0xff, 0xff, 0x19, 0x40, 0x53, 0xca, 0x27, 0x03, 0x00, 0x00,
}