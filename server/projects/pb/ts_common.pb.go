// Code generated by protoc-gen-go.
// source: ts_common.proto
// DO NOT EDIT!

package pb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type RankItem struct {
	PlayerId    uint64 `protobuf:"varint,1,opt,name=player_id,json=playerId" json:"player_id,omitempty"`
	AttachDatas []byte `protobuf:"bytes,2,opt,name=attach_datas,json=attachDatas,proto3" json:"attach_datas,omitempty"`
	SortField1  int64  `protobuf:"varint,3,opt,name=sort_field1,json=sortField1" json:"sort_field1,omitempty"`
	SortField2  int64  `protobuf:"varint,4,opt,name=sort_field2,json=sortField2" json:"sort_field2,omitempty"`
	SortField3  int64  `protobuf:"varint,5,opt,name=sort_field3,json=sortField3" json:"sort_field3,omitempty"`
	SortField4  int64  `protobuf:"varint,6,opt,name=sort_field4,json=sortField4" json:"sort_field4,omitempty"`
	SortField5  int64  `protobuf:"varint,7,opt,name=sort_field5,json=sortField5" json:"sort_field5,omitempty"`
}

func (m *RankItem) Reset()                    { *m = RankItem{} }
func (m *RankItem) String() string            { return proto.CompactTextString(m) }
func (*RankItem) ProtoMessage()               {}
func (*RankItem) Descriptor() ([]byte, []int) { return fileDescriptor7, []int{0} }

func init() {
	proto.RegisterType((*RankItem)(nil), "pb.rank_item")
}

var fileDescriptor7 = []byte{
	// 177 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xe2, 0xe2, 0x2f, 0x29, 0x8e, 0x4f,
	0xce, 0xcf, 0xcd, 0xcd, 0xcf, 0xd3, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x2a, 0x48, 0x52,
	0xfa, 0xc0, 0xc8, 0xc5, 0x59, 0x94, 0x98, 0x97, 0x1d, 0x9f, 0x59, 0x92, 0x9a, 0x2b, 0x24, 0xcd,
	0xc5, 0x59, 0x90, 0x93, 0x58, 0x99, 0x5a, 0x14, 0x9f, 0x99, 0x22, 0xc1, 0xa8, 0xc0, 0xa8, 0xc1,
	0x12, 0xc4, 0x01, 0x11, 0xf0, 0x4c, 0x11, 0x52, 0xe4, 0xe2, 0x49, 0x2c, 0x29, 0x49, 0x4c, 0xce,
	0x88, 0x4f, 0x49, 0x2c, 0x49, 0x2c, 0x96, 0x60, 0x02, 0xca, 0xf3, 0x04, 0x71, 0x43, 0xc4, 0x5c,
	0x40, 0x42, 0x42, 0xf2, 0x5c, 0xdc, 0xc5, 0xf9, 0x45, 0x25, 0xf1, 0x69, 0x99, 0xa9, 0x39, 0x29,
	0x86, 0x12, 0xcc, 0x40, 0x15, 0xcc, 0x41, 0x5c, 0x20, 0x21, 0x37, 0xb0, 0x08, 0xaa, 0x02, 0x23,
	0x09, 0x16, 0x34, 0x05, 0x46, 0xa8, 0x0a, 0x8c, 0x25, 0x58, 0xd1, 0x14, 0x18, 0xa3, 0x2a, 0x30,
	0x91, 0x60, 0x43, 0x53, 0x60, 0x82, 0xaa, 0xc0, 0x54, 0x82, 0x1d, 0x4d, 0x81, 0x69, 0x12, 0x1b,
	0xd8, 0xf7, 0xc6, 0x80, 0x00, 0x00, 0x00, 0xff, 0xff, 0x7f, 0xc4, 0x6d, 0x72, 0x10, 0x01, 0x00,
	0x00,
}
