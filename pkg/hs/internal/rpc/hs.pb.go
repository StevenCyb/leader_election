// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        v5.27.3
// source: hs.proto

package rpc

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Enum for message status.
type Status int32

const (
	Status_UNKNOWN        Status = 0
	Status_RECEIVED       Status = 1
	Status_DISCARDED      Status = 2
	Status_PROCESSED      Status = 3
	Status_REPLY_SENT     Status = 4
	Status_LEADER_ELECTED Status = 5
)

// Enum value maps for Status.
var (
	Status_name = map[int32]string{
		0: "UNKNOWN",
		1: "RECEIVED",
		2: "DISCARDED",
		3: "PROCESSED",
		4: "REPLY_SENT",
		5: "LEADER_ELECTED",
	}
	Status_value = map[string]int32{
		"UNKNOWN":        0,
		"RECEIVED":       1,
		"DISCARDED":      2,
		"PROCESSED":      3,
		"REPLY_SENT":     4,
		"LEADER_ELECTED": 5,
	}
)

func (x Status) Enum() *Status {
	p := new(Status)
	*p = x
	return p
}

func (x Status) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Status) Descriptor() protoreflect.EnumDescriptor {
	return file_hs_proto_enumTypes[0].Descriptor()
}

func (Status) Type() protoreflect.EnumType {
	return &file_hs_proto_enumTypes[0]
}

func (x Status) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Status.Descriptor instead.
func (Status) EnumDescriptor() ([]byte, []int) {
	return file_hs_proto_rawDescGZIP(), []int{0}
}

// Enum for message direction.
type Direction int32

const (
	Direction_LEFT  Direction = 0
	Direction_RIGHT Direction = 1
)

// Enum value maps for Direction.
var (
	Direction_name = map[int32]string{
		0: "LEFT",
		1: "RIGHT",
	}
	Direction_value = map[string]int32{
		"LEFT":  0,
		"RIGHT": 1,
	}
)

func (x Direction) Enum() *Direction {
	p := new(Direction)
	*p = x
	return p
}

func (x Direction) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Direction) Descriptor() protoreflect.EnumDescriptor {
	return file_hs_proto_enumTypes[1].Descriptor()
}

func (Direction) Type() protoreflect.EnumType {
	return &file_hs_proto_enumTypes[1]
}

func (x Direction) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Direction.Descriptor instead.
func (Direction) EnumDescriptor() ([]byte, []int) {
	return file_hs_proto_rawDescGZIP(), []int{1}
}

// Message to be sent to the HS service as a probe.
type ProbeMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Uid       uint64    `protobuf:"varint,1,opt,name=uid,proto3" json:"uid,omitempty"`
	Hops      uint32    `protobuf:"varint,2,opt,name=hops,proto3" json:"hops,omitempty"`
	Direction Direction `protobuf:"varint,3,opt,name=direction,proto3,enum=hs.Direction" json:"direction,omitempty"`
	Phase     uint32    `protobuf:"varint,4,opt,name=phase,proto3" json:"phase,omitempty"`
}

func (x *ProbeMessage) Reset() {
	*x = ProbeMessage{}
	mi := &file_hs_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProbeMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProbeMessage) ProtoMessage() {}

func (x *ProbeMessage) ProtoReflect() protoreflect.Message {
	mi := &file_hs_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProbeMessage.ProtoReflect.Descriptor instead.
func (*ProbeMessage) Descriptor() ([]byte, []int) {
	return file_hs_proto_rawDescGZIP(), []int{0}
}

func (x *ProbeMessage) GetUid() uint64 {
	if x != nil {
		return x.Uid
	}
	return 0
}

func (x *ProbeMessage) GetHops() uint32 {
	if x != nil {
		return x.Hops
	}
	return 0
}

func (x *ProbeMessage) GetDirection() Direction {
	if x != nil {
		return x.Direction
	}
	return Direction_LEFT
}

func (x *ProbeMessage) GetPhase() uint32 {
	if x != nil {
		return x.Phase
	}
	return 0
}

// Message to be sent to the HS service as a reply.
type ReplyMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Uid   uint64 `protobuf:"varint,1,opt,name=uid,proto3" json:"uid,omitempty"`
	Phase uint32 `protobuf:"varint,2,opt,name=phase,proto3" json:"phase,omitempty"`
}

func (x *ReplyMessage) Reset() {
	*x = ReplyMessage{}
	mi := &file_hs_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ReplyMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReplyMessage) ProtoMessage() {}

func (x *ReplyMessage) ProtoReflect() protoreflect.Message {
	mi := &file_hs_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReplyMessage.ProtoReflect.Descriptor instead.
func (*ReplyMessage) Descriptor() ([]byte, []int) {
	return file_hs_proto_rawDescGZIP(), []int{1}
}

func (x *ReplyMessage) GetUid() uint64 {
	if x != nil {
		return x.Uid
	}
	return 0
}

func (x *ReplyMessage) GetPhase() uint32 {
	if x != nil {
		return x.Phase
	}
	return 0
}

// Message to terminate the election process.
type TerminateMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Uid uint64 `protobuf:"varint,1,opt,name=uid,proto3" json:"uid,omitempty"`
}

func (x *TerminateMessage) Reset() {
	*x = TerminateMessage{}
	mi := &file_hs_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TerminateMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TerminateMessage) ProtoMessage() {}

func (x *TerminateMessage) ProtoReflect() protoreflect.Message {
	mi := &file_hs_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TerminateMessage.ProtoReflect.Descriptor instead.
func (*TerminateMessage) Descriptor() ([]byte, []int) {
	return file_hs_proto_rawDescGZIP(), []int{2}
}

func (x *TerminateMessage) GetUid() uint64 {
	if x != nil {
		return x.Uid
	}
	return 0
}

// Response from the HS service.
type HSResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status Status `protobuf:"varint,1,opt,name=status,proto3,enum=hs.Status" json:"status,omitempty"`
}

func (x *HSResponse) Reset() {
	*x = HSResponse{}
	mi := &file_hs_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HSResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HSResponse) ProtoMessage() {}

func (x *HSResponse) ProtoReflect() protoreflect.Message {
	mi := &file_hs_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HSResponse.ProtoReflect.Descriptor instead.
func (*HSResponse) Descriptor() ([]byte, []int) {
	return file_hs_proto_rawDescGZIP(), []int{3}
}

func (x *HSResponse) GetStatus() Status {
	if x != nil {
		return x.Status
	}
	return Status_UNKNOWN
}

var File_hs_proto protoreflect.FileDescriptor

var file_hs_proto_rawDesc = []byte{
	0x0a, 0x08, 0x68, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x68, 0x73, 0x1a, 0x1b,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f,
	0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x77, 0x0a, 0x0c, 0x50,
	0x72, 0x6f, 0x62, 0x65, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x75,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x03, 0x75, 0x69, 0x64, 0x12, 0x12, 0x0a,
	0x04, 0x68, 0x6f, 0x70, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x68, 0x6f, 0x70,
	0x73, 0x12, 0x2b, 0x0a, 0x09, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0e, 0x32, 0x0d, 0x2e, 0x68, 0x73, 0x2e, 0x44, 0x69, 0x72, 0x65, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x52, 0x09, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x14,
	0x0a, 0x05, 0x70, 0x68, 0x61, 0x73, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x05, 0x70,
	0x68, 0x61, 0x73, 0x65, 0x22, 0x36, 0x0a, 0x0c, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x04, 0x52, 0x03, 0x75, 0x69, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x70, 0x68, 0x61, 0x73, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x05, 0x70, 0x68, 0x61, 0x73, 0x65, 0x22, 0x24, 0x0a, 0x10,
	0x54, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x61, 0x74, 0x65, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x12, 0x10, 0x0a, 0x03, 0x75, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x03, 0x75,
	0x69, 0x64, 0x22, 0x30, 0x0a, 0x0a, 0x48, 0x53, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x22, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e,
	0x32, 0x0a, 0x2e, 0x68, 0x73, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x2a, 0x65, 0x0a, 0x06, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x0b,
	0x0a, 0x07, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x00, 0x12, 0x0c, 0x0a, 0x08, 0x52,
	0x45, 0x43, 0x45, 0x49, 0x56, 0x45, 0x44, 0x10, 0x01, 0x12, 0x0d, 0x0a, 0x09, 0x44, 0x49, 0x53,
	0x43, 0x41, 0x52, 0x44, 0x45, 0x44, 0x10, 0x02, 0x12, 0x0d, 0x0a, 0x09, 0x50, 0x52, 0x4f, 0x43,
	0x45, 0x53, 0x53, 0x45, 0x44, 0x10, 0x03, 0x12, 0x0e, 0x0a, 0x0a, 0x52, 0x45, 0x50, 0x4c, 0x59,
	0x5f, 0x53, 0x45, 0x4e, 0x54, 0x10, 0x04, 0x12, 0x12, 0x0a, 0x0e, 0x4c, 0x45, 0x41, 0x44, 0x45,
	0x52, 0x5f, 0x45, 0x4c, 0x45, 0x43, 0x54, 0x45, 0x44, 0x10, 0x05, 0x2a, 0x20, 0x0a, 0x09, 0x44,
	0x69, 0x72, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x08, 0x0a, 0x04, 0x4c, 0x45, 0x46, 0x54,
	0x10, 0x00, 0x12, 0x09, 0x0a, 0x05, 0x52, 0x49, 0x47, 0x48, 0x54, 0x10, 0x01, 0x32, 0xcc, 0x01,
	0x0a, 0x09, 0x48, 0x53, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x29, 0x0a, 0x05, 0x50,
	0x72, 0x6f, 0x62, 0x65, 0x12, 0x10, 0x2e, 0x68, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x62, 0x65, 0x4d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x1a, 0x0e, 0x2e, 0x68, 0x73, 0x2e, 0x48, 0x53, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x29, 0x0a, 0x05, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12,
	0x10, 0x2e, 0x68, 0x73, 0x2e, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x1a, 0x0e, 0x2e, 0x68, 0x73, 0x2e, 0x48, 0x53, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x31, 0x0a, 0x09, 0x54, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x61, 0x74, 0x65, 0x12, 0x14,
	0x2e, 0x68, 0x73, 0x2e, 0x54, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x61, 0x74, 0x65, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x1a, 0x0e, 0x2e, 0x68, 0x73, 0x2e, 0x48, 0x53, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x36, 0x0a, 0x04, 0x50, 0x69, 0x6e, 0x67, 0x12, 0x16, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45,
	0x6d, 0x70, 0x74, 0x79, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x42, 0x15, 0x5a, 0x13,
	0x70, 0x6b, 0x67, 0x2f, 0x68, 0x73, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f,
	0x72, 0x70, 0x63, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_hs_proto_rawDescOnce sync.Once
	file_hs_proto_rawDescData = file_hs_proto_rawDesc
)

func file_hs_proto_rawDescGZIP() []byte {
	file_hs_proto_rawDescOnce.Do(func() {
		file_hs_proto_rawDescData = protoimpl.X.CompressGZIP(file_hs_proto_rawDescData)
	})
	return file_hs_proto_rawDescData
}

var file_hs_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_hs_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_hs_proto_goTypes = []any{
	(Status)(0),              // 0: hs.Status
	(Direction)(0),           // 1: hs.Direction
	(*ProbeMessage)(nil),     // 2: hs.ProbeMessage
	(*ReplyMessage)(nil),     // 3: hs.ReplyMessage
	(*TerminateMessage)(nil), // 4: hs.TerminateMessage
	(*HSResponse)(nil),       // 5: hs.HSResponse
	(*emptypb.Empty)(nil),    // 6: google.protobuf.Empty
}
var file_hs_proto_depIdxs = []int32{
	1, // 0: hs.ProbeMessage.direction:type_name -> hs.Direction
	0, // 1: hs.HSResponse.status:type_name -> hs.Status
	2, // 2: hs.HSService.Probe:input_type -> hs.ProbeMessage
	3, // 3: hs.HSService.Reply:input_type -> hs.ReplyMessage
	4, // 4: hs.HSService.Terminate:input_type -> hs.TerminateMessage
	6, // 5: hs.HSService.Ping:input_type -> google.protobuf.Empty
	5, // 6: hs.HSService.Probe:output_type -> hs.HSResponse
	5, // 7: hs.HSService.Reply:output_type -> hs.HSResponse
	5, // 8: hs.HSService.Terminate:output_type -> hs.HSResponse
	6, // 9: hs.HSService.Ping:output_type -> google.protobuf.Empty
	6, // [6:10] is the sub-list for method output_type
	2, // [2:6] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_hs_proto_init() }
func file_hs_proto_init() {
	if File_hs_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_hs_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_hs_proto_goTypes,
		DependencyIndexes: file_hs_proto_depIdxs,
		EnumInfos:         file_hs_proto_enumTypes,
		MessageInfos:      file_hs_proto_msgTypes,
	}.Build()
	File_hs_proto = out.File
	file_hs_proto_rawDesc = nil
	file_hs_proto_goTypes = nil
	file_hs_proto_depIdxs = nil
}
