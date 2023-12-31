// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        (unknown)
// source: zanzigo/v1/zanzigo.proto

package zanzigov1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Tuple struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ObjectType      string `protobuf:"bytes,1,opt,name=object_type,json=objectType,proto3" json:"object_type,omitempty"`
	ObjectId        string `protobuf:"bytes,2,opt,name=object_id,json=objectId,proto3" json:"object_id,omitempty"`
	ObjectRelation  string `protobuf:"bytes,3,opt,name=object_relation,json=objectRelation,proto3" json:"object_relation,omitempty"`
	SubjectType     string `protobuf:"bytes,4,opt,name=subject_type,json=subjectType,proto3" json:"subject_type,omitempty"`
	SubjectId       string `protobuf:"bytes,5,opt,name=subject_id,json=subjectId,proto3" json:"subject_id,omitempty"`
	SubjectRelation string `protobuf:"bytes,6,opt,name=subject_relation,json=subjectRelation,proto3" json:"subject_relation,omitempty"`
}

func (x *Tuple) Reset() {
	*x = Tuple{}
	if protoimpl.UnsafeEnabled {
		mi := &file_zanzigo_v1_zanzigo_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Tuple) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Tuple) ProtoMessage() {}

func (x *Tuple) ProtoReflect() protoreflect.Message {
	mi := &file_zanzigo_v1_zanzigo_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Tuple.ProtoReflect.Descriptor instead.
func (*Tuple) Descriptor() ([]byte, []int) {
	return file_zanzigo_v1_zanzigo_proto_rawDescGZIP(), []int{0}
}

func (x *Tuple) GetObjectType() string {
	if x != nil {
		return x.ObjectType
	}
	return ""
}

func (x *Tuple) GetObjectId() string {
	if x != nil {
		return x.ObjectId
	}
	return ""
}

func (x *Tuple) GetObjectRelation() string {
	if x != nil {
		return x.ObjectRelation
	}
	return ""
}

func (x *Tuple) GetSubjectType() string {
	if x != nil {
		return x.SubjectType
	}
	return ""
}

func (x *Tuple) GetSubjectId() string {
	if x != nil {
		return x.SubjectId
	}
	return ""
}

func (x *Tuple) GetSubjectRelation() string {
	if x != nil {
		return x.SubjectRelation
	}
	return ""
}

type WriteRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Tuple *Tuple `protobuf:"bytes,1,opt,name=tuple,proto3" json:"tuple,omitempty"`
}

func (x *WriteRequest) Reset() {
	*x = WriteRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_zanzigo_v1_zanzigo_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WriteRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WriteRequest) ProtoMessage() {}

func (x *WriteRequest) ProtoReflect() protoreflect.Message {
	mi := &file_zanzigo_v1_zanzigo_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WriteRequest.ProtoReflect.Descriptor instead.
func (*WriteRequest) Descriptor() ([]byte, []int) {
	return file_zanzigo_v1_zanzigo_proto_rawDescGZIP(), []int{1}
}

func (x *WriteRequest) GetTuple() *Tuple {
	if x != nil {
		return x.Tuple
	}
	return nil
}

type WriteResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *WriteResponse) Reset() {
	*x = WriteResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_zanzigo_v1_zanzigo_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WriteResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WriteResponse) ProtoMessage() {}

func (x *WriteResponse) ProtoReflect() protoreflect.Message {
	mi := &file_zanzigo_v1_zanzigo_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WriteResponse.ProtoReflect.Descriptor instead.
func (*WriteResponse) Descriptor() ([]byte, []int) {
	return file_zanzigo_v1_zanzigo_proto_rawDescGZIP(), []int{2}
}

type ReadRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Tuple *Tuple `protobuf:"bytes,1,opt,name=tuple,proto3" json:"tuple,omitempty"`
}

func (x *ReadRequest) Reset() {
	*x = ReadRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_zanzigo_v1_zanzigo_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReadRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReadRequest) ProtoMessage() {}

func (x *ReadRequest) ProtoReflect() protoreflect.Message {
	mi := &file_zanzigo_v1_zanzigo_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReadRequest.ProtoReflect.Descriptor instead.
func (*ReadRequest) Descriptor() ([]byte, []int) {
	return file_zanzigo_v1_zanzigo_proto_rawDescGZIP(), []int{3}
}

func (x *ReadRequest) GetTuple() *Tuple {
	if x != nil {
		return x.Tuple
	}
	return nil
}

type ReadResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Uuid string `protobuf:"bytes,1,opt,name=uuid,proto3" json:"uuid,omitempty"`
}

func (x *ReadResponse) Reset() {
	*x = ReadResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_zanzigo_v1_zanzigo_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReadResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReadResponse) ProtoMessage() {}

func (x *ReadResponse) ProtoReflect() protoreflect.Message {
	mi := &file_zanzigo_v1_zanzigo_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReadResponse.ProtoReflect.Descriptor instead.
func (*ReadResponse) Descriptor() ([]byte, []int) {
	return file_zanzigo_v1_zanzigo_proto_rawDescGZIP(), []int{4}
}

func (x *ReadResponse) GetUuid() string {
	if x != nil {
		return x.Uuid
	}
	return ""
}

type CheckRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Tuple *Tuple `protobuf:"bytes,1,opt,name=tuple,proto3" json:"tuple,omitempty"`
}

func (x *CheckRequest) Reset() {
	*x = CheckRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_zanzigo_v1_zanzigo_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CheckRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CheckRequest) ProtoMessage() {}

func (x *CheckRequest) ProtoReflect() protoreflect.Message {
	mi := &file_zanzigo_v1_zanzigo_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CheckRequest.ProtoReflect.Descriptor instead.
func (*CheckRequest) Descriptor() ([]byte, []int) {
	return file_zanzigo_v1_zanzigo_proto_rawDescGZIP(), []int{5}
}

func (x *CheckRequest) GetTuple() *Tuple {
	if x != nil {
		return x.Tuple
	}
	return nil
}

type CheckResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Result bool `protobuf:"varint,1,opt,name=result,proto3" json:"result,omitempty"`
}

func (x *CheckResponse) Reset() {
	*x = CheckResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_zanzigo_v1_zanzigo_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CheckResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CheckResponse) ProtoMessage() {}

func (x *CheckResponse) ProtoReflect() protoreflect.Message {
	mi := &file_zanzigo_v1_zanzigo_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CheckResponse.ProtoReflect.Descriptor instead.
func (*CheckResponse) Descriptor() ([]byte, []int) {
	return file_zanzigo_v1_zanzigo_proto_rawDescGZIP(), []int{6}
}

func (x *CheckResponse) GetResult() bool {
	if x != nil {
		return x.Result
	}
	return false
}

var File_zanzigo_v1_zanzigo_proto protoreflect.FileDescriptor

var file_zanzigo_v1_zanzigo_proto_rawDesc = []byte{
	0x0a, 0x18, 0x7a, 0x61, 0x6e, 0x7a, 0x69, 0x67, 0x6f, 0x2f, 0x76, 0x31, 0x2f, 0x7a, 0x61, 0x6e,
	0x7a, 0x69, 0x67, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x7a, 0x61, 0x6e, 0x7a,
	0x69, 0x67, 0x6f, 0x2e, 0x76, 0x31, 0x22, 0xdb, 0x01, 0x0a, 0x05, 0x54, 0x75, 0x70, 0x6c, 0x65,
	0x12, 0x1f, 0x0a, 0x0b, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x54, 0x79, 0x70,
	0x65, 0x12, 0x1b, 0x0a, 0x09, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x64, 0x12, 0x27,
	0x0a, 0x0f, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x72, 0x65, 0x6c, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x52,
	0x65, 0x6c, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x21, 0x0a, 0x0c, 0x73, 0x75, 0x62, 0x6a, 0x65,
	0x63, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x73,
	0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x75,
	0x62, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09,
	0x73, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x64, 0x12, 0x29, 0x0a, 0x10, 0x73, 0x75, 0x62,
	0x6a, 0x65, 0x63, 0x74, 0x5f, 0x72, 0x65, 0x6c, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x06, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0f, 0x73, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x52, 0x65, 0x6c, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x22, 0x37, 0x0a, 0x0c, 0x57, 0x72, 0x69, 0x74, 0x65, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x27, 0x0a, 0x05, 0x74, 0x75, 0x70, 0x6c, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x7a, 0x61, 0x6e, 0x7a, 0x69, 0x67, 0x6f, 0x2e, 0x76, 0x31,
	0x2e, 0x54, 0x75, 0x70, 0x6c, 0x65, 0x52, 0x05, 0x74, 0x75, 0x70, 0x6c, 0x65, 0x22, 0x0f, 0x0a,
	0x0d, 0x57, 0x72, 0x69, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x36,
	0x0a, 0x0b, 0x52, 0x65, 0x61, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x27, 0x0a,
	0x05, 0x74, 0x75, 0x70, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x7a,
	0x61, 0x6e, 0x7a, 0x69, 0x67, 0x6f, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x75, 0x70, 0x6c, 0x65, 0x52,
	0x05, 0x74, 0x75, 0x70, 0x6c, 0x65, 0x22, 0x22, 0x0a, 0x0c, 0x52, 0x65, 0x61, 0x64, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x75, 0x75, 0x69, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x75, 0x75, 0x69, 0x64, 0x22, 0x37, 0x0a, 0x0c, 0x43, 0x68,
	0x65, 0x63, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x27, 0x0a, 0x05, 0x74, 0x75,
	0x70, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x7a, 0x61, 0x6e, 0x7a,
	0x69, 0x67, 0x6f, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x75, 0x70, 0x6c, 0x65, 0x52, 0x05, 0x74, 0x75,
	0x70, 0x6c, 0x65, 0x22, 0x27, 0x0a, 0x0d, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x32, 0xcd, 0x01, 0x0a,
	0x0e, 0x5a, 0x61, 0x6e, 0x7a, 0x69, 0x67, 0x6f, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12,
	0x3e, 0x0a, 0x05, 0x57, 0x72, 0x69, 0x74, 0x65, 0x12, 0x18, 0x2e, 0x7a, 0x61, 0x6e, 0x7a, 0x69,
	0x67, 0x6f, 0x2e, 0x76, 0x31, 0x2e, 0x57, 0x72, 0x69, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x19, 0x2e, 0x7a, 0x61, 0x6e, 0x7a, 0x69, 0x67, 0x6f, 0x2e, 0x76, 0x31, 0x2e,
	0x57, 0x72, 0x69, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12,
	0x3b, 0x0a, 0x04, 0x52, 0x65, 0x61, 0x64, 0x12, 0x17, 0x2e, 0x7a, 0x61, 0x6e, 0x7a, 0x69, 0x67,
	0x6f, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x61, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x18, 0x2e, 0x7a, 0x61, 0x6e, 0x7a, 0x69, 0x67, 0x6f, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65,
	0x61, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x3e, 0x0a, 0x05,
	0x43, 0x68, 0x65, 0x63, 0x6b, 0x12, 0x18, 0x2e, 0x7a, 0x61, 0x6e, 0x7a, 0x69, 0x67, 0x6f, 0x2e,
	0x76, 0x31, 0x2e, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x19, 0x2e, 0x7a, 0x61, 0x6e, 0x7a, 0x69, 0x67, 0x6f, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x68, 0x65,
	0x63, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x34, 0x5a, 0x32,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x74, 0x72, 0x65, 0x76, 0x65,
	0x78, 0x2f, 0x7a, 0x61, 0x6e, 0x7a, 0x69, 0x67, 0x6f, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x7a, 0x61,
	0x6e, 0x7a, 0x69, 0x67, 0x6f, 0x2f, 0x76, 0x31, 0x3b, 0x7a, 0x61, 0x6e, 0x7a, 0x69, 0x67, 0x6f,
	0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_zanzigo_v1_zanzigo_proto_rawDescOnce sync.Once
	file_zanzigo_v1_zanzigo_proto_rawDescData = file_zanzigo_v1_zanzigo_proto_rawDesc
)

func file_zanzigo_v1_zanzigo_proto_rawDescGZIP() []byte {
	file_zanzigo_v1_zanzigo_proto_rawDescOnce.Do(func() {
		file_zanzigo_v1_zanzigo_proto_rawDescData = protoimpl.X.CompressGZIP(file_zanzigo_v1_zanzigo_proto_rawDescData)
	})
	return file_zanzigo_v1_zanzigo_proto_rawDescData
}

var file_zanzigo_v1_zanzigo_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_zanzigo_v1_zanzigo_proto_goTypes = []interface{}{
	(*Tuple)(nil),         // 0: zanzigo.v1.Tuple
	(*WriteRequest)(nil),  // 1: zanzigo.v1.WriteRequest
	(*WriteResponse)(nil), // 2: zanzigo.v1.WriteResponse
	(*ReadRequest)(nil),   // 3: zanzigo.v1.ReadRequest
	(*ReadResponse)(nil),  // 4: zanzigo.v1.ReadResponse
	(*CheckRequest)(nil),  // 5: zanzigo.v1.CheckRequest
	(*CheckResponse)(nil), // 6: zanzigo.v1.CheckResponse
}
var file_zanzigo_v1_zanzigo_proto_depIdxs = []int32{
	0, // 0: zanzigo.v1.WriteRequest.tuple:type_name -> zanzigo.v1.Tuple
	0, // 1: zanzigo.v1.ReadRequest.tuple:type_name -> zanzigo.v1.Tuple
	0, // 2: zanzigo.v1.CheckRequest.tuple:type_name -> zanzigo.v1.Tuple
	1, // 3: zanzigo.v1.ZanzigoService.Write:input_type -> zanzigo.v1.WriteRequest
	3, // 4: zanzigo.v1.ZanzigoService.Read:input_type -> zanzigo.v1.ReadRequest
	5, // 5: zanzigo.v1.ZanzigoService.Check:input_type -> zanzigo.v1.CheckRequest
	2, // 6: zanzigo.v1.ZanzigoService.Write:output_type -> zanzigo.v1.WriteResponse
	4, // 7: zanzigo.v1.ZanzigoService.Read:output_type -> zanzigo.v1.ReadResponse
	6, // 8: zanzigo.v1.ZanzigoService.Check:output_type -> zanzigo.v1.CheckResponse
	6, // [6:9] is the sub-list for method output_type
	3, // [3:6] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_zanzigo_v1_zanzigo_proto_init() }
func file_zanzigo_v1_zanzigo_proto_init() {
	if File_zanzigo_v1_zanzigo_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_zanzigo_v1_zanzigo_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Tuple); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_zanzigo_v1_zanzigo_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WriteRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_zanzigo_v1_zanzigo_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WriteResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_zanzigo_v1_zanzigo_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReadRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_zanzigo_v1_zanzigo_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReadResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_zanzigo_v1_zanzigo_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CheckRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_zanzigo_v1_zanzigo_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CheckResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_zanzigo_v1_zanzigo_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_zanzigo_v1_zanzigo_proto_goTypes,
		DependencyIndexes: file_zanzigo_v1_zanzigo_proto_depIdxs,
		MessageInfos:      file_zanzigo_v1_zanzigo_proto_msgTypes,
	}.Build()
	File_zanzigo_v1_zanzigo_proto = out.File
	file_zanzigo_v1_zanzigo_proto_rawDesc = nil
	file_zanzigo_v1_zanzigo_proto_goTypes = nil
	file_zanzigo_v1_zanzigo_proto_depIdxs = nil
}
