// Code generated by protoc-gen-go. DO NOT EDIT.
// source: proto/feature/segment.proto

package feature

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Segment_Status int32

const (
	Segment_INITIAL   Segment_Status = 0
	Segment_UPLOADING Segment_Status = 1
	Segment_SUCEEDED  Segment_Status = 2
	Segment_FAILED    Segment_Status = 3
)

var Segment_Status_name = map[int32]string{
	0: "INITIAL",
	1: "UPLOADING",
	2: "SUCEEDED",
	3: "FAILED",
}

var Segment_Status_value = map[string]int32{
	"INITIAL":   0,
	"UPLOADING": 1,
	"SUCEEDED":  2,
	"FAILED":    3,
}

func (x Segment_Status) String() string {
	return proto.EnumName(Segment_Status_name, int32(x))
}

func (Segment_Status) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_8ad8286456030bff, []int{0, 0}
}

type SegmentUser_State int32

const (
	SegmentUser_INCLUDED SegmentUser_State = 0
	SegmentUser_EXCLUDED SegmentUser_State = 1 // Deprecated: Do not use.
)

var SegmentUser_State_name = map[int32]string{
	0: "INCLUDED",
	1: "EXCLUDED",
}

var SegmentUser_State_value = map[string]int32{
	"INCLUDED": 0,
	"EXCLUDED": 1,
}

func (x SegmentUser_State) String() string {
	return proto.EnumName(SegmentUser_State_name, int32(x))
}

func (SegmentUser_State) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_8ad8286456030bff, []int{1, 0}
}

type Segment struct {
	Id                   string         `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name                 string         `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Description          string         `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"`
	Rules                []*Rule        `protobuf:"bytes,4,rep,name=rules,proto3" json:"rules,omitempty"`
	CreatedAt            int64          `protobuf:"varint,5,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	UpdatedAt            int64          `protobuf:"varint,6,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
	Version              int64          `protobuf:"varint,7,opt,name=version,proto3" json:"version,omitempty"` // Deprecated: Do not use.
	Deleted              bool           `protobuf:"varint,8,opt,name=deleted,proto3" json:"deleted,omitempty"`
	IncludedUserCount    int64          `protobuf:"varint,9,opt,name=included_user_count,json=includedUserCount,proto3" json:"included_user_count,omitempty"`
	ExcludedUserCount    int64          `protobuf:"varint,10,opt,name=excluded_user_count,json=excludedUserCount,proto3" json:"excluded_user_count,omitempty"` // Deprecated: Do not use.
	Status               Segment_Status `protobuf:"varint,11,opt,name=status,proto3,enum=bucketeer.feature.Segment_Status" json:"status,omitempty"`
	IsInUseStatus        bool           `protobuf:"varint,12,opt,name=is_in_use_status,json=isInUseStatus,proto3" json:"is_in_use_status,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *Segment) Reset()         { *m = Segment{} }
func (m *Segment) String() string { return proto.CompactTextString(m) }
func (*Segment) ProtoMessage()    {}
func (*Segment) Descriptor() ([]byte, []int) {
	return fileDescriptor_8ad8286456030bff, []int{0}
}

func (m *Segment) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Segment.Unmarshal(m, b)
}
func (m *Segment) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Segment.Marshal(b, m, deterministic)
}
func (m *Segment) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Segment.Merge(m, src)
}
func (m *Segment) XXX_Size() int {
	return xxx_messageInfo_Segment.Size(m)
}
func (m *Segment) XXX_DiscardUnknown() {
	xxx_messageInfo_Segment.DiscardUnknown(m)
}

var xxx_messageInfo_Segment proto.InternalMessageInfo

func (m *Segment) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Segment) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Segment) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *Segment) GetRules() []*Rule {
	if m != nil {
		return m.Rules
	}
	return nil
}

func (m *Segment) GetCreatedAt() int64 {
	if m != nil {
		return m.CreatedAt
	}
	return 0
}

func (m *Segment) GetUpdatedAt() int64 {
	if m != nil {
		return m.UpdatedAt
	}
	return 0
}

// Deprecated: Do not use.
func (m *Segment) GetVersion() int64 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *Segment) GetDeleted() bool {
	if m != nil {
		return m.Deleted
	}
	return false
}

func (m *Segment) GetIncludedUserCount() int64 {
	if m != nil {
		return m.IncludedUserCount
	}
	return 0
}

// Deprecated: Do not use.
func (m *Segment) GetExcludedUserCount() int64 {
	if m != nil {
		return m.ExcludedUserCount
	}
	return 0
}

func (m *Segment) GetStatus() Segment_Status {
	if m != nil {
		return m.Status
	}
	return Segment_INITIAL
}

func (m *Segment) GetIsInUseStatus() bool {
	if m != nil {
		return m.IsInUseStatus
	}
	return false
}

type SegmentUser struct {
	Id                   string            `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	SegmentId            string            `protobuf:"bytes,2,opt,name=segment_id,json=segmentId,proto3" json:"segment_id,omitempty"`
	UserId               string            `protobuf:"bytes,3,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	State                SegmentUser_State `protobuf:"varint,4,opt,name=state,proto3,enum=bucketeer.feature.SegmentUser_State" json:"state,omitempty"`
	Deleted              bool              `protobuf:"varint,5,opt,name=deleted,proto3" json:"deleted,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *SegmentUser) Reset()         { *m = SegmentUser{} }
func (m *SegmentUser) String() string { return proto.CompactTextString(m) }
func (*SegmentUser) ProtoMessage()    {}
func (*SegmentUser) Descriptor() ([]byte, []int) {
	return fileDescriptor_8ad8286456030bff, []int{1}
}

func (m *SegmentUser) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SegmentUser.Unmarshal(m, b)
}
func (m *SegmentUser) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SegmentUser.Marshal(b, m, deterministic)
}
func (m *SegmentUser) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SegmentUser.Merge(m, src)
}
func (m *SegmentUser) XXX_Size() int {
	return xxx_messageInfo_SegmentUser.Size(m)
}
func (m *SegmentUser) XXX_DiscardUnknown() {
	xxx_messageInfo_SegmentUser.DiscardUnknown(m)
}

var xxx_messageInfo_SegmentUser proto.InternalMessageInfo

func (m *SegmentUser) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *SegmentUser) GetSegmentId() string {
	if m != nil {
		return m.SegmentId
	}
	return ""
}

func (m *SegmentUser) GetUserId() string {
	if m != nil {
		return m.UserId
	}
	return ""
}

func (m *SegmentUser) GetState() SegmentUser_State {
	if m != nil {
		return m.State
	}
	return SegmentUser_INCLUDED
}

func (m *SegmentUser) GetDeleted() bool {
	if m != nil {
		return m.Deleted
	}
	return false
}

type SegmentUsers struct {
	SegmentId            string         `protobuf:"bytes,1,opt,name=segment_id,json=segmentId,proto3" json:"segment_id,omitempty"`
	Users                []*SegmentUser `protobuf:"bytes,2,rep,name=users,proto3" json:"users,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *SegmentUsers) Reset()         { *m = SegmentUsers{} }
func (m *SegmentUsers) String() string { return proto.CompactTextString(m) }
func (*SegmentUsers) ProtoMessage()    {}
func (*SegmentUsers) Descriptor() ([]byte, []int) {
	return fileDescriptor_8ad8286456030bff, []int{2}
}

func (m *SegmentUsers) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SegmentUsers.Unmarshal(m, b)
}
func (m *SegmentUsers) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SegmentUsers.Marshal(b, m, deterministic)
}
func (m *SegmentUsers) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SegmentUsers.Merge(m, src)
}
func (m *SegmentUsers) XXX_Size() int {
	return xxx_messageInfo_SegmentUsers.Size(m)
}
func (m *SegmentUsers) XXX_DiscardUnknown() {
	xxx_messageInfo_SegmentUsers.DiscardUnknown(m)
}

var xxx_messageInfo_SegmentUsers proto.InternalMessageInfo

func (m *SegmentUsers) GetSegmentId() string {
	if m != nil {
		return m.SegmentId
	}
	return ""
}

func (m *SegmentUsers) GetUsers() []*SegmentUser {
	if m != nil {
		return m.Users
	}
	return nil
}

func init() {
	proto.RegisterEnum("bucketeer.feature.Segment_Status", Segment_Status_name, Segment_Status_value)
	proto.RegisterEnum("bucketeer.feature.SegmentUser_State", SegmentUser_State_name, SegmentUser_State_value)
	proto.RegisterType((*Segment)(nil), "bucketeer.feature.Segment")
	proto.RegisterType((*SegmentUser)(nil), "bucketeer.feature.SegmentUser")
	proto.RegisterType((*SegmentUsers)(nil), "bucketeer.feature.SegmentUsers")
}

func init() { proto.RegisterFile("proto/feature/segment.proto", fileDescriptor_8ad8286456030bff) }

var fileDescriptor_8ad8286456030bff = []byte{
	// 521 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x53, 0xd1, 0x6e, 0xd3, 0x30,
	0x14, 0x5d, 0xd2, 0x25, 0x69, 0x6f, 0xb7, 0x29, 0xf5, 0x1e, 0x66, 0x0d, 0x86, 0x42, 0x85, 0xb4,
	0x08, 0x69, 0xa9, 0x54, 0x78, 0x81, 0x07, 0xa4, 0xae, 0x2d, 0x28, 0x52, 0x55, 0x50, 0x4a, 0x25,
	0xc4, 0x4b, 0x94, 0xc6, 0x97, 0x61, 0xd1, 0xa6, 0x55, 0xec, 0x20, 0x3e, 0x94, 0xcf, 0xe1, 0x01,
	0xd9, 0x71, 0x47, 0xbb, 0x02, 0x6f, 0xb9, 0xe7, 0x9c, 0x7b, 0x7d, 0x8e, 0x7d, 0x03, 0x8f, 0x36,
	0xe5, 0x5a, 0xae, 0x7b, 0x5f, 0x30, 0x93, 0x55, 0x89, 0x3d, 0x81, 0x77, 0x2b, 0x2c, 0x64, 0xa4,
	0x51, 0xd2, 0x59, 0x54, 0xf9, 0x37, 0x94, 0x88, 0x65, 0x64, 0x04, 0x97, 0x74, 0x5f, 0x5f, 0x56,
	0x4b, 0xac, 0xc5, 0xdd, 0x5f, 0x0d, 0xf0, 0x66, 0x75, 0x3b, 0x39, 0x03, 0x9b, 0x33, 0x6a, 0x05,
	0x56, 0xd8, 0x4a, 0x6c, 0xce, 0x08, 0x81, 0xe3, 0x22, 0x5b, 0x21, 0xb5, 0x35, 0xa2, 0xbf, 0x49,
	0x00, 0x6d, 0x86, 0x22, 0x2f, 0xf9, 0x46, 0xf2, 0x75, 0x41, 0x1b, 0x9a, 0xda, 0x85, 0xc8, 0x0d,
	0x38, 0x6a, 0xbe, 0xa0, 0xc7, 0x41, 0x23, 0x6c, 0xf7, 0x2f, 0xa2, 0x03, 0x3b, 0x51, 0x52, 0x2d,
	0x31, 0xa9, 0x55, 0xe4, 0x0a, 0x20, 0x2f, 0x31, 0x93, 0xc8, 0xd2, 0x4c, 0x52, 0x27, 0xb0, 0xc2,
	0x46, 0xd2, 0x32, 0xc8, 0x40, 0x2a, 0xba, 0xda, 0xb0, 0x2d, 0xed, 0xd6, 0xb4, 0x41, 0x06, 0x92,
	0x3c, 0x06, 0xef, 0x3b, 0x96, 0x42, 0x59, 0xf1, 0x14, 0x77, 0x6b, 0x53, 0x2b, 0xd9, 0x42, 0x84,
	0x82, 0xc7, 0x70, 0x89, 0x12, 0x19, 0x6d, 0x06, 0x56, 0xd8, 0x4c, 0xb6, 0x25, 0x89, 0xe0, 0x9c,
	0x17, 0xf9, 0xb2, 0x62, 0xc8, 0xd2, 0x4a, 0x60, 0x99, 0xe6, 0xeb, 0xaa, 0x90, 0xb4, 0xa5, 0xe7,
	0x77, 0xb6, 0xd4, 0x5c, 0x60, 0x39, 0x54, 0x04, 0xe9, 0xc3, 0x39, 0xfe, 0x38, 0xd4, 0xc3, 0xfd,
	0x99, 0x9d, 0x2d, 0xfd, 0xa7, 0xe7, 0x15, 0xb8, 0x42, 0x66, 0xb2, 0x12, 0xb4, 0x1d, 0x58, 0xe1,
	0x59, 0xff, 0xe9, 0x5f, 0x6e, 0xc2, 0x5c, 0x7d, 0x34, 0xd3, 0xc2, 0xc4, 0x34, 0x90, 0x6b, 0xf0,
	0xb9, 0x48, 0x79, 0xa1, 0xce, 0x4a, 0xcd, 0x90, 0x13, 0x9d, 0xe0, 0x94, 0x8b, 0xb8, 0x98, 0x0b,
	0xac, 0x1b, 0xba, 0x6f, 0xc0, 0xad, 0xbf, 0x48, 0x1b, 0xbc, 0x78, 0x1a, 0x7f, 0x8c, 0x07, 0x13,
	0xff, 0x88, 0x9c, 0x42, 0x6b, 0xfe, 0x61, 0xf2, 0x7e, 0x30, 0x8a, 0xa7, 0xef, 0x7c, 0x8b, 0x9c,
	0x40, 0x73, 0x36, 0x1f, 0x8e, 0xc7, 0xa3, 0xf1, 0xc8, 0xb7, 0x09, 0x80, 0xfb, 0x76, 0x10, 0x4f,
	0xc6, 0x23, 0xbf, 0xd1, 0xfd, 0x69, 0x41, 0xdb, 0x78, 0x50, 0xc6, 0x0f, 0x56, 0xe0, 0x0a, 0xc0,
	0x2c, 0x57, 0xca, 0x99, 0x59, 0x84, 0x96, 0x41, 0x62, 0x46, 0x2e, 0xc0, 0xd3, 0xb7, 0xc1, 0x99,
	0xd9, 0x04, 0x57, 0x95, 0x31, 0x23, 0xaf, 0xc1, 0x51, 0xb6, 0x91, 0x1e, 0xeb, 0xe8, 0xcf, 0xfe,
	0x1d, 0x5d, 0x1d, 0xab, 0xe3, 0x63, 0x52, 0xb7, 0xec, 0xbe, 0x9a, 0xb3, 0xf7, 0x6a, 0xdd, 0x6b,
	0x70, 0xb4, 0x52, 0x05, 0x8a, 0xa7, 0xc3, 0xc9, 0x5c, 0x05, 0x3a, 0x22, 0x3e, 0x34, 0xc7, 0x9f,
	0x4c, 0x65, 0x5d, 0xda, 0x4d, 0xab, 0x9b, 0xc3, 0xc9, 0xce, 0x78, 0xf1, 0x20, 0x86, 0xf5, 0x30,
	0xc6, 0x4b, 0x70, 0x94, 0x6f, 0x41, 0x6d, 0xbd, 0xb2, 0x4f, 0xfe, 0xef, 0x36, 0xa9, 0xc5, 0xb7,
	0xcf, 0x3f, 0x87, 0x77, 0x5c, 0x7e, 0xad, 0x16, 0x51, 0xbe, 0x5e, 0xf5, 0xf2, 0xec, 0x86, 0x6d,
	0x7a, 0xf7, 0x8d, 0xbd, 0xbd, 0x3f, 0x6e, 0xe1, 0xea, 0xf2, 0xc5, 0xef, 0x00, 0x00, 0x00, 0xff,
	0xff, 0x19, 0x6a, 0xf3, 0x4e, 0xb9, 0x03, 0x00, 0x00,
}
