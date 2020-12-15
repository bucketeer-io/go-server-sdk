// Code generated by protoc-gen-go. DO NOT EDIT.
// source: proto/feature/feature.proto

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

type Feature struct {
	Id                    string               `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name                  string               `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Description           string               `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"`
	Enabled               bool                 `protobuf:"varint,4,opt,name=enabled,proto3" json:"enabled,omitempty"`
	Deleted               bool                 `protobuf:"varint,5,opt,name=deleted,proto3" json:"deleted,omitempty"`
	EvaluationUndelayable bool                 `protobuf:"varint,6,opt,name=evaluation_undelayable,json=evaluationUndelayable,proto3" json:"evaluation_undelayable,omitempty"` // Deprecated: Do not use.
	Ttl                   int32                `protobuf:"varint,7,opt,name=ttl,proto3" json:"ttl,omitempty"`
	Version               int32                `protobuf:"varint,8,opt,name=version,proto3" json:"version,omitempty"`
	CreatedAt             int64                `protobuf:"varint,9,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	UpdatedAt             int64                `protobuf:"varint,10,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
	Variations            []*Variation         `protobuf:"bytes,11,rep,name=variations,proto3" json:"variations,omitempty"`
	Targets               []*Target            `protobuf:"bytes,12,rep,name=targets,proto3" json:"targets,omitempty"`
	Rules                 []*Rule              `protobuf:"bytes,13,rep,name=rules,proto3" json:"rules,omitempty"`
	DefaultStrategy       *Strategy            `protobuf:"bytes,14,opt,name=default_strategy,json=defaultStrategy,proto3" json:"default_strategy,omitempty"`
	OffVariation          string               `protobuf:"bytes,15,opt,name=off_variation,json=offVariation,proto3" json:"off_variation,omitempty"`
	Tags                  []string             `protobuf:"bytes,16,rep,name=tags,proto3" json:"tags,omitempty"`
	LastUsedInfo          *FeatureLastUsedInfo `protobuf:"bytes,17,opt,name=last_used_info,json=lastUsedInfo,proto3" json:"last_used_info,omitempty"`
	XXX_NoUnkeyedLiteral  struct{}             `json:"-"`
	XXX_unrecognized      []byte               `json:"-"`
	XXX_sizecache         int32                `json:"-"`
}

func (m *Feature) Reset()         { *m = Feature{} }
func (m *Feature) String() string { return proto.CompactTextString(m) }
func (*Feature) ProtoMessage()    {}
func (*Feature) Descriptor() ([]byte, []int) {
	return fileDescriptor_01855a1c92e13124, []int{0}
}

func (m *Feature) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Feature.Unmarshal(m, b)
}
func (m *Feature) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Feature.Marshal(b, m, deterministic)
}
func (m *Feature) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Feature.Merge(m, src)
}
func (m *Feature) XXX_Size() int {
	return xxx_messageInfo_Feature.Size(m)
}
func (m *Feature) XXX_DiscardUnknown() {
	xxx_messageInfo_Feature.DiscardUnknown(m)
}

var xxx_messageInfo_Feature proto.InternalMessageInfo

func (m *Feature) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Feature) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Feature) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *Feature) GetEnabled() bool {
	if m != nil {
		return m.Enabled
	}
	return false
}

func (m *Feature) GetDeleted() bool {
	if m != nil {
		return m.Deleted
	}
	return false
}

// Deprecated: Do not use.
func (m *Feature) GetEvaluationUndelayable() bool {
	if m != nil {
		return m.EvaluationUndelayable
	}
	return false
}

func (m *Feature) GetTtl() int32 {
	if m != nil {
		return m.Ttl
	}
	return 0
}

func (m *Feature) GetVersion() int32 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *Feature) GetCreatedAt() int64 {
	if m != nil {
		return m.CreatedAt
	}
	return 0
}

func (m *Feature) GetUpdatedAt() int64 {
	if m != nil {
		return m.UpdatedAt
	}
	return 0
}

func (m *Feature) GetVariations() []*Variation {
	if m != nil {
		return m.Variations
	}
	return nil
}

func (m *Feature) GetTargets() []*Target {
	if m != nil {
		return m.Targets
	}
	return nil
}

func (m *Feature) GetRules() []*Rule {
	if m != nil {
		return m.Rules
	}
	return nil
}

func (m *Feature) GetDefaultStrategy() *Strategy {
	if m != nil {
		return m.DefaultStrategy
	}
	return nil
}

func (m *Feature) GetOffVariation() string {
	if m != nil {
		return m.OffVariation
	}
	return ""
}

func (m *Feature) GetTags() []string {
	if m != nil {
		return m.Tags
	}
	return nil
}

func (m *Feature) GetLastUsedInfo() *FeatureLastUsedInfo {
	if m != nil {
		return m.LastUsedInfo
	}
	return nil
}

type TagFeatures struct {
	Tag                  string     `protobuf:"bytes,1,opt,name=tag,proto3" json:"tag,omitempty"`
	Features             []*Feature `protobuf:"bytes,2,rep,name=features,proto3" json:"features,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *TagFeatures) Reset()         { *m = TagFeatures{} }
func (m *TagFeatures) String() string { return proto.CompactTextString(m) }
func (*TagFeatures) ProtoMessage()    {}
func (*TagFeatures) Descriptor() ([]byte, []int) {
	return fileDescriptor_01855a1c92e13124, []int{1}
}

func (m *TagFeatures) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TagFeatures.Unmarshal(m, b)
}
func (m *TagFeatures) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TagFeatures.Marshal(b, m, deterministic)
}
func (m *TagFeatures) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TagFeatures.Merge(m, src)
}
func (m *TagFeatures) XXX_Size() int {
	return xxx_messageInfo_TagFeatures.Size(m)
}
func (m *TagFeatures) XXX_DiscardUnknown() {
	xxx_messageInfo_TagFeatures.DiscardUnknown(m)
}

var xxx_messageInfo_TagFeatures proto.InternalMessageInfo

func (m *TagFeatures) GetTag() string {
	if m != nil {
		return m.Tag
	}
	return ""
}

func (m *TagFeatures) GetFeatures() []*Feature {
	if m != nil {
		return m.Features
	}
	return nil
}

func init() {
	proto.RegisterType((*Feature)(nil), "bucketeer.feature.Feature")
	proto.RegisterType((*TagFeatures)(nil), "bucketeer.feature.TagFeatures")
}

func init() { proto.RegisterFile("proto/feature/feature.proto", fileDescriptor_01855a1c92e13124) }

var fileDescriptor_01855a1c92e13124 = []byte{
	// 503 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x93, 0x51, 0x6b, 0xdb, 0x30,
	0x10, 0xc7, 0x71, 0xd2, 0x34, 0xc9, 0x25, 0x4d, 0x53, 0xc1, 0x36, 0x2d, 0x6d, 0xc1, 0x74, 0x30,
	0x4c, 0xa1, 0x09, 0xb4, 0x30, 0x18, 0xec, 0xa5, 0x7d, 0x28, 0x0c, 0xfa, 0xa4, 0xb5, 0x1b, 0xec,
	0xc5, 0x28, 0xd1, 0xd9, 0x33, 0x73, 0xed, 0x60, 0x9d, 0x02, 0xfd, 0x28, 0xfb, 0xb6, 0x43, 0xb2,
	0x9c, 0xc6, 0x2c, 0x7b, 0xca, 0xdd, 0xfd, 0x7f, 0xff, 0xdc, 0x29, 0x77, 0x81, 0xd3, 0x75, 0x55,
	0x52, 0xb9, 0x48, 0x50, 0x92, 0xa9, 0xb0, 0xf9, 0x9c, 0xbb, 0x2a, 0x3b, 0x59, 0x9a, 0xd5, 0x6f,
	0x24, 0xc4, 0x6a, 0xee, 0x85, 0x19, 0x6f, 0xf3, 0x95, 0xc9, 0x3d, 0x3c, 0x9b, 0xb5, 0x15, 0x92,
	0x55, 0x8a, 0xe4, 0xb5, 0xf3, 0xb6, 0xb6, 0x91, 0x55, 0x26, 0x29, 0x2b, 0x0b, 0x2f, 0x9f, 0xb5,
	0x65, 0x4d, 0x95, 0x24, 0x4c, 0x5f, 0xbc, 0x7a, 0xb9, 0x77, 0xc4, 0x38, 0x97, 0x9a, 0x62, 0xa3,
	0x51, 0xc5, 0x59, 0x91, 0x94, 0x35, 0x7b, 0xf1, 0xa7, 0x07, 0xfd, 0xfb, 0x1a, 0x60, 0x13, 0xe8,
	0x64, 0x8a, 0x07, 0x61, 0x10, 0x0d, 0x45, 0x27, 0x53, 0x8c, 0xc1, 0x41, 0x21, 0x9f, 0x91, 0x77,
	0x5c, 0xc5, 0xc5, 0x2c, 0x84, 0x91, 0x42, 0xbd, 0xaa, 0xb2, 0xb5, 0x1d, 0x87, 0x77, 0x9d, 0xb4,
	0x5b, 0x62, 0x1c, 0xfa, 0x58, 0xc8, 0x65, 0x8e, 0x8a, 0x1f, 0x84, 0x41, 0x34, 0x10, 0x4d, 0x6a,
	0x15, 0x85, 0x39, 0x12, 0x2a, 0xde, 0xab, 0x15, 0x9f, 0xb2, 0xcf, 0xf0, 0x16, 0x37, 0x32, 0x37,
	0xee, 0x8d, 0xb1, 0x29, 0x14, 0xe6, 0xf2, 0xc5, 0x9a, 0xf8, 0xa1, 0x05, 0xef, 0x3a, 0x3c, 0x10,
	0x6f, 0x5e, 0x89, 0xa7, 0x57, 0x80, 0x4d, 0xa1, 0x4b, 0x94, 0xf3, 0x7e, 0x18, 0x44, 0x3d, 0x61,
	0x43, 0xdb, 0x66, 0x83, 0x95, 0xb6, 0xe3, 0x0d, 0x5c, 0xb5, 0x49, 0xd9, 0x39, 0xc0, 0xaa, 0x42,
	0x49, 0xa8, 0x62, 0x49, 0x7c, 0x18, 0x06, 0x51, 0x57, 0x0c, 0x7d, 0xe5, 0x96, 0xac, 0x6c, 0xd6,
	0xaa, 0x91, 0xa1, 0x96, 0x7d, 0xe5, 0x96, 0xd8, 0x17, 0x80, 0xed, 0x1e, 0x34, 0x1f, 0x85, 0xdd,
	0x68, 0x74, 0x7d, 0x36, 0xff, 0x67, 0xe3, 0xf3, 0xef, 0x0d, 0x24, 0x76, 0x78, 0x76, 0x03, 0xfd,
	0x7a, 0xc3, 0x9a, 0x8f, 0x9d, 0xf5, 0xfd, 0x1e, 0xeb, 0xa3, 0x23, 0x44, 0x43, 0xb2, 0x2b, 0xe8,
	0xd9, 0x83, 0xd1, 0xfc, 0xc8, 0x59, 0xde, 0xed, 0xb1, 0x08, 0x93, 0xa3, 0xa8, 0x29, 0x76, 0x0f,
	0x53, 0x85, 0x89, 0x34, 0x39, 0xc5, 0xcd, 0x49, 0xf0, 0x49, 0x18, 0x44, 0xa3, 0xeb, 0xd3, 0x3d,
	0xce, 0x6f, 0x1e, 0x11, 0xc7, 0xde, 0xd4, 0x14, 0xd8, 0x07, 0x38, 0x2a, 0x93, 0x24, 0xde, 0x4e,
	0xcf, 0x8f, 0xdd, 0x9a, 0xc7, 0x65, 0x92, 0x6c, 0x1f, 0x67, 0xaf, 0x83, 0x64, 0xaa, 0xf9, 0x34,
	0xec, 0xda, 0xeb, 0xb0, 0x31, 0x7b, 0x80, 0x49, 0xfb, 0xca, 0xf8, 0x89, 0x6b, 0xff, 0x71, 0x4f,
	0x7b, 0x7f, 0x75, 0x0f, 0x52, 0xd3, 0x93, 0x46, 0xf5, 0xb5, 0x48, 0x4a, 0x31, 0xce, 0x77, 0xb2,
	0x8b, 0x1f, 0x30, 0x7a, 0x94, 0xa9, 0xe7, 0xb4, 0xdb, 0xb4, 0x4c, 0xfd, 0x7d, 0xda, 0x90, 0x7d,
	0x82, 0x81, 0xff, 0x36, 0xcd, 0x3b, 0xee, 0x17, 0x9a, 0xfd, 0xbf, 0x91, 0xd8, 0xb2, 0x77, 0x97,
	0x3f, 0xa3, 0x34, 0xa3, 0x5f, 0x66, 0x39, 0x5f, 0x95, 0xcf, 0x8b, 0x95, 0xbc, 0x52, 0xeb, 0xc5,
	0xd6, 0xb7, 0x68, 0xfd, 0x7b, 0x96, 0x87, 0x2e, 0xbd, 0xf9, 0x1b, 0x00, 0x00, 0xff, 0xff, 0x9a,
	0x43, 0xbc, 0xf6, 0xf8, 0x03, 0x00, 0x00,
}
