// Code generated by protoc-gen-go. DO NOT EDIT.
// source: proto/api/api.proto

package api // import "github.com/duolacloud/microbase/proto/api"

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import any "github.com/golang/protobuf/ptypes/any"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Status struct {
	Code                 int32      `protobuf:"varint,1,opt,name=code,proto3" json:"code,omitempty"`
	Message              string     `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	Details              []*any.Any `protobuf:"bytes,3,rep,name=details,proto3" json:"details,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *Status) Reset()         { *m = Status{} }
func (m *Status) String() string { return proto.CompactTextString(m) }
func (*Status) ProtoMessage()    {}
func (*Status) Descriptor() ([]byte, []int) {
	return fileDescriptor_api_958048e13e84e500, []int{0}
}
func (m *Status) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Status.Unmarshal(m, b)
}
func (m *Status) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Status.Marshal(b, m, deterministic)
}
func (dst *Status) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Status.Merge(dst, src)
}
func (m *Status) XXX_Size() int {
	return xxx_messageInfo_Status.Size(m)
}
func (m *Status) XXX_DiscardUnknown() {
	xxx_messageInfo_Status.DiscardUnknown(m)
}

var xxx_messageInfo_Status proto.InternalMessageInfo

func (m *Status) GetCode() int32 {
	if m != nil {
		return m.Code
	}
	return 0
}

func (m *Status) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func (m *Status) GetDetails() []*any.Any {
	if m != nil {
		return m.Details
	}
	return nil
}

func init() {
	proto.RegisterType((*Status)(nil), "api.Status")
}

func init() { proto.RegisterFile("proto/api/api.proto", fileDescriptor_api_958048e13e84e500) }

var fileDescriptor_api_958048e13e84e500 = []byte{
	// 187 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x34, 0x8e, 0x3d, 0x6f, 0x83, 0x30,
	0x10, 0x86, 0x45, 0x69, 0x41, 0x75, 0x37, 0xb7, 0x83, 0xdb, 0x09, 0x75, 0x62, 0xa9, 0x2d, 0xb5,
	0x63, 0xa7, 0xe4, 0x27, 0x90, 0x2d, 0xdb, 0xf9, 0x03, 0xc7, 0x92, 0xe1, 0x2c, 0x6c, 0x0f, 0xfc,
	0xfb, 0x08, 0x13, 0x86, 0x93, 0xee, 0xb9, 0x7b, 0xa5, 0xf7, 0x21, 0xef, 0x61, 0xc1, 0x84, 0x02,
	0x82, 0xdb, 0x86, 0x17, 0xa2, 0x35, 0x04, 0xf7, 0xf5, 0x69, 0x11, 0xad, 0x37, 0xa2, 0x9c, 0x64,
	0x1e, 0x05, 0xcc, 0xeb, 0xfe, 0xff, 0x1e, 0x49, 0x73, 0x49, 0x90, 0x72, 0xa4, 0x94, 0x3c, 0x2b,
	0xd4, 0x86, 0x55, 0x5d, 0xd5, 0xbf, 0x0c, 0x65, 0xa7, 0x8c, 0xb4, 0x93, 0x89, 0x11, 0xac, 0x61,
	0x4f, 0x5d, 0xd5, 0xbf, 0x0e, 0x07, 0x52, 0x4e, 0x5a, 0x6d, 0x12, 0x38, 0x1f, 0x59, 0xdd, 0xd5,
	0xfd, 0xdb, 0xef, 0x07, 0xdf, 0x4b, 0xf8, 0x51, 0xc2, 0x4f, 0xf3, 0x3a, 0x1c, 0xa1, 0xb3, 0xb8,
	0xfe, 0x58, 0x97, 0x6e, 0x59, 0x72, 0x85, 0x93, 0xd0, 0x19, 0x3d, 0x28, 0x8f, 0x59, 0x8b, 0xc9,
	0xa9, 0x05, 0x25, 0xc4, 0x87, 0xdd, 0xa6, 0xfe, 0x0f, 0xc1, 0xc9, 0xa6, 0xe0, 0xdf, 0x3d, 0x00,
	0x00, 0xff, 0xff, 0x59, 0x98, 0x8b, 0x7a, 0xd6, 0x00, 0x00, 0x00,
}
