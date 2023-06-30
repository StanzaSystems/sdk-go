// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        (unknown)
// source: stanza/hub/v1/common.proto

package hubv1

import (
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options"
	_ "google.golang.org/genproto/googleapis/api/annotations"
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

type Health int32

const (
	Health_HEALTH_UNSPECIFIED Health = 0
	Health_HEALTH_OK          Health = 1
	Health_HEALTH_OVERLOAD    Health = 2
	Health_HEALTH_DOWN        Health = 3
)

// Enum value maps for Health.
var (
	Health_name = map[int32]string{
		0: "HEALTH_UNSPECIFIED",
		1: "HEALTH_OK",
		2: "HEALTH_OVERLOAD",
		3: "HEALTH_DOWN",
	}
	Health_value = map[string]int32{
		"HEALTH_UNSPECIFIED": 0,
		"HEALTH_OK":          1,
		"HEALTH_OVERLOAD":    2,
		"HEALTH_DOWN":        3,
	}
)

func (x Health) Enum() *Health {
	p := new(Health)
	*p = x
	return p
}

func (x Health) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Health) Descriptor() protoreflect.EnumDescriptor {
	return file_stanza_hub_v1_common_proto_enumTypes[0].Descriptor()
}

func (Health) Type() protoreflect.EnumType {
	return &file_stanza_hub_v1_common_proto_enumTypes[0]
}

func (x Health) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Health.Descriptor instead.
func (Health) EnumDescriptor() ([]byte, []int) {
	return file_stanza_hub_v1_common_proto_rawDescGZIP(), []int{0}
}

type State int32

const (
	State_STATE_UNSPECIFIED State = 0
	State_STATE_ENABLED     State = 1
	State_STATE_DISABLED    State = 2
)

// Enum value maps for State.
var (
	State_name = map[int32]string{
		0: "STATE_UNSPECIFIED",
		1: "STATE_ENABLED",
		2: "STATE_DISABLED",
	}
	State_value = map[string]int32{
		"STATE_UNSPECIFIED": 0,
		"STATE_ENABLED":     1,
		"STATE_DISABLED":    2,
	}
)

func (x State) Enum() *State {
	p := new(State)
	*p = x
	return p
}

func (x State) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (State) Descriptor() protoreflect.EnumDescriptor {
	return file_stanza_hub_v1_common_proto_enumTypes[1].Descriptor()
}

func (State) Type() protoreflect.EnumType {
	return &file_stanza_hub_v1_common_proto_enumTypes[1]
}

func (x State) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use State.Descriptor instead.
func (State) EnumDescriptor() ([]byte, []int) {
	return file_stanza_hub_v1_common_proto_rawDescGZIP(), []int{1}
}

type DecoratorSelector struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Environment string `protobuf:"bytes,1,opt,name=environment,proto3" json:"environment,omitempty"`
	Name        string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Tags        []*Tag `protobuf:"bytes,6,rep,name=tags,proto3" json:"tags,omitempty"`
}

func (x *DecoratorSelector) Reset() {
	*x = DecoratorSelector{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stanza_hub_v1_common_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DecoratorSelector) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DecoratorSelector) ProtoMessage() {}

func (x *DecoratorSelector) ProtoReflect() protoreflect.Message {
	mi := &file_stanza_hub_v1_common_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DecoratorSelector.ProtoReflect.Descriptor instead.
func (*DecoratorSelector) Descriptor() ([]byte, []int) {
	return file_stanza_hub_v1_common_proto_rawDescGZIP(), []int{0}
}

func (x *DecoratorSelector) GetEnvironment() string {
	if x != nil {
		return x.Environment
	}
	return ""
}

func (x *DecoratorSelector) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *DecoratorSelector) GetTags() []*Tag {
	if x != nil {
		return x.Tags
	}
	return nil
}

type FeatureSelector struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Environment string   `protobuf:"bytes,1,opt,name=environment,proto3" json:"environment,omitempty"`
	Names       []string `protobuf:"bytes,2,rep,name=names,proto3" json:"names,omitempty"`
	Tags        []*Tag   `protobuf:"bytes,6,rep,name=tags,proto3" json:"tags,omitempty"`
}

func (x *FeatureSelector) Reset() {
	*x = FeatureSelector{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stanza_hub_v1_common_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FeatureSelector) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FeatureSelector) ProtoMessage() {}

func (x *FeatureSelector) ProtoReflect() protoreflect.Message {
	mi := &file_stanza_hub_v1_common_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FeatureSelector.ProtoReflect.Descriptor instead.
func (*FeatureSelector) Descriptor() ([]byte, []int) {
	return file_stanza_hub_v1_common_proto_rawDescGZIP(), []int{1}
}

func (x *FeatureSelector) GetEnvironment() string {
	if x != nil {
		return x.Environment
	}
	return ""
}

func (x *FeatureSelector) GetNames() []string {
	if x != nil {
		return x.Names
	}
	return nil
}

func (x *FeatureSelector) GetTags() []*Tag {
	if x != nil {
		return x.Tags
	}
	return nil
}

type DecoratorFeatureSelector struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Environment   string  `protobuf:"bytes,1,opt,name=environment,proto3" json:"environment,omitempty"`
	DecoratorName string  `protobuf:"bytes,2,opt,name=decorator_name,json=decoratorName,proto3" json:"decorator_name,omitempty"`
	FeatureName   *string `protobuf:"bytes,3,opt,name=feature_name,json=featureName,proto3,oneof" json:"feature_name,omitempty"`
	Tags          []*Tag  `protobuf:"bytes,6,rep,name=tags,proto3" json:"tags,omitempty"`
}

func (x *DecoratorFeatureSelector) Reset() {
	*x = DecoratorFeatureSelector{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stanza_hub_v1_common_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DecoratorFeatureSelector) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DecoratorFeatureSelector) ProtoMessage() {}

func (x *DecoratorFeatureSelector) ProtoReflect() protoreflect.Message {
	mi := &file_stanza_hub_v1_common_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DecoratorFeatureSelector.ProtoReflect.Descriptor instead.
func (*DecoratorFeatureSelector) Descriptor() ([]byte, []int) {
	return file_stanza_hub_v1_common_proto_rawDescGZIP(), []int{2}
}

func (x *DecoratorFeatureSelector) GetEnvironment() string {
	if x != nil {
		return x.Environment
	}
	return ""
}

func (x *DecoratorFeatureSelector) GetDecoratorName() string {
	if x != nil {
		return x.DecoratorName
	}
	return ""
}

func (x *DecoratorFeatureSelector) GetFeatureName() string {
	if x != nil && x.FeatureName != nil {
		return *x.FeatureName
	}
	return ""
}

func (x *DecoratorFeatureSelector) GetTags() []*Tag {
	if x != nil {
		return x.Tags
	}
	return nil
}

type DecoratorServiceSelector struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Environment    string `protobuf:"bytes,1,opt,name=environment,proto3" json:"environment,omitempty"`
	DecoratorName  string `protobuf:"bytes,2,opt,name=decorator_name,json=decoratorName,proto3" json:"decorator_name,omitempty"`
	ServiceName    string `protobuf:"bytes,3,opt,name=service_name,json=serviceName,proto3" json:"service_name,omitempty"`
	ServiceRelease string `protobuf:"bytes,4,opt,name=service_release,json=serviceRelease,proto3" json:"service_release,omitempty"`
	Tags           []*Tag `protobuf:"bytes,6,rep,name=tags,proto3" json:"tags,omitempty"`
}

func (x *DecoratorServiceSelector) Reset() {
	*x = DecoratorServiceSelector{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stanza_hub_v1_common_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DecoratorServiceSelector) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DecoratorServiceSelector) ProtoMessage() {}

func (x *DecoratorServiceSelector) ProtoReflect() protoreflect.Message {
	mi := &file_stanza_hub_v1_common_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DecoratorServiceSelector.ProtoReflect.Descriptor instead.
func (*DecoratorServiceSelector) Descriptor() ([]byte, []int) {
	return file_stanza_hub_v1_common_proto_rawDescGZIP(), []int{3}
}

func (x *DecoratorServiceSelector) GetEnvironment() string {
	if x != nil {
		return x.Environment
	}
	return ""
}

func (x *DecoratorServiceSelector) GetDecoratorName() string {
	if x != nil {
		return x.DecoratorName
	}
	return ""
}

func (x *DecoratorServiceSelector) GetServiceName() string {
	if x != nil {
		return x.ServiceName
	}
	return ""
}

func (x *DecoratorServiceSelector) GetServiceRelease() string {
	if x != nil {
		return x.ServiceRelease
	}
	return ""
}

func (x *DecoratorServiceSelector) GetTags() []*Tag {
	if x != nil {
		return x.Tags
	}
	return nil
}

type ServiceSelector struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Environment string  `protobuf:"bytes,1,opt,name=environment,proto3" json:"environment,omitempty"`
	Name        string  `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Release     *string `protobuf:"bytes,3,opt,name=release,proto3,oneof" json:"release,omitempty"`
	Tags        []*Tag  `protobuf:"bytes,6,rep,name=tags,proto3" json:"tags,omitempty"`
}

func (x *ServiceSelector) Reset() {
	*x = ServiceSelector{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stanza_hub_v1_common_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServiceSelector) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServiceSelector) ProtoMessage() {}

func (x *ServiceSelector) ProtoReflect() protoreflect.Message {
	mi := &file_stanza_hub_v1_common_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServiceSelector.ProtoReflect.Descriptor instead.
func (*ServiceSelector) Descriptor() ([]byte, []int) {
	return file_stanza_hub_v1_common_proto_rawDescGZIP(), []int{4}
}

func (x *ServiceSelector) GetEnvironment() string {
	if x != nil {
		return x.Environment
	}
	return ""
}

func (x *ServiceSelector) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *ServiceSelector) GetRelease() string {
	if x != nil && x.Release != nil {
		return *x.Release
	}
	return ""
}

func (x *ServiceSelector) GetTags() []*Tag {
	if x != nil {
		return x.Tags
	}
	return nil
}

type HealthByPriority struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Priority uint32 `protobuf:"varint,1,opt,name=priority,proto3" json:"priority,omitempty"`
	Health   Health `protobuf:"varint,2,opt,name=health,proto3,enum=stanza.hub.v1.Health" json:"health,omitempty"`
}

func (x *HealthByPriority) Reset() {
	*x = HealthByPriority{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stanza_hub_v1_common_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HealthByPriority) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HealthByPriority) ProtoMessage() {}

func (x *HealthByPriority) ProtoReflect() protoreflect.Message {
	mi := &file_stanza_hub_v1_common_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HealthByPriority.ProtoReflect.Descriptor instead.
func (*HealthByPriority) Descriptor() ([]byte, []int) {
	return file_stanza_hub_v1_common_proto_rawDescGZIP(), []int{5}
}

func (x *HealthByPriority) GetPriority() uint32 {
	if x != nil {
		return x.Priority
	}
	return 0
}

func (x *HealthByPriority) GetHealth() Health {
	if x != nil {
		return x.Health
	}
	return Health_HEALTH_UNSPECIFIED
}

type Tag struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key   string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value string `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *Tag) Reset() {
	*x = Tag{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stanza_hub_v1_common_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Tag) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Tag) ProtoMessage() {}

func (x *Tag) ProtoReflect() protoreflect.Message {
	mi := &file_stanza_hub_v1_common_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Tag.ProtoReflect.Descriptor instead.
func (*Tag) Descriptor() ([]byte, []int) {
	return file_stanza_hub_v1_common_proto_rawDescGZIP(), []int{6}
}

func (x *Tag) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *Tag) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

var File_stanza_hub_v1_common_proto protoreflect.FileDescriptor

var file_stanza_hub_v1_common_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x73, 0x74, 0x61, 0x6e, 0x7a, 0x61, 0x2f, 0x68, 0x75, 0x62, 0x2f, 0x76, 0x31, 0x2f,
	0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0d, 0x73, 0x74,
	0x61, 0x6e, 0x7a, 0x61, 0x2e, 0x68, 0x75, 0x62, 0x2e, 0x76, 0x31, 0x1a, 0x1f, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x62, 0x65,
	0x68, 0x61, 0x76, 0x69, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69,
	0x76, 0x32, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x7b, 0x0a, 0x11,
	0x44, 0x65, 0x63, 0x6f, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f,
	0x72, 0x12, 0x25, 0x0a, 0x0b, 0x65, 0x6e, 0x76, 0x69, 0x72, 0x6f, 0x6e, 0x6d, 0x65, 0x6e, 0x74,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x0b, 0x65, 0x6e, 0x76,
	0x69, 0x72, 0x6f, 0x6e, 0x6d, 0x65, 0x6e, 0x74, 0x12, 0x17, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x12, 0x26, 0x0a, 0x04, 0x74, 0x61, 0x67, 0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x12, 0x2e, 0x73, 0x74, 0x61, 0x6e, 0x7a, 0x61, 0x2e, 0x68, 0x75, 0x62, 0x2e, 0x76, 0x31, 0x2e,
	0x54, 0x61, 0x67, 0x52, 0x04, 0x74, 0x61, 0x67, 0x73, 0x22, 0x76, 0x0a, 0x0f, 0x46, 0x65, 0x61,
	0x74, 0x75, 0x72, 0x65, 0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x12, 0x25, 0x0a, 0x0b,
	0x65, 0x6e, 0x76, 0x69, 0x72, 0x6f, 0x6e, 0x6d, 0x65, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x0b, 0x65, 0x6e, 0x76, 0x69, 0x72, 0x6f, 0x6e, 0x6d,
	0x65, 0x6e, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03,
	0x28, 0x09, 0x52, 0x05, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x12, 0x26, 0x0a, 0x04, 0x74, 0x61, 0x67,
	0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x73, 0x74, 0x61, 0x6e, 0x7a, 0x61,
	0x2e, 0x68, 0x75, 0x62, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x61, 0x67, 0x52, 0x04, 0x74, 0x61, 0x67,
	0x73, 0x22, 0xce, 0x01, 0x0a, 0x18, 0x44, 0x65, 0x63, 0x6f, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x46,
	0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x12, 0x25,
	0x0a, 0x0b, 0x65, 0x6e, 0x76, 0x69, 0x72, 0x6f, 0x6e, 0x6d, 0x65, 0x6e, 0x74, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x0b, 0x65, 0x6e, 0x76, 0x69, 0x72, 0x6f,
	0x6e, 0x6d, 0x65, 0x6e, 0x74, 0x12, 0x2a, 0x0a, 0x0e, 0x64, 0x65, 0x63, 0x6f, 0x72, 0x61, 0x74,
	0x6f, 0x72, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xe0,
	0x41, 0x02, 0x52, 0x0d, 0x64, 0x65, 0x63, 0x6f, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x4e, 0x61, 0x6d,
	0x65, 0x12, 0x26, 0x0a, 0x0c, 0x66, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x5f, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x0b, 0x66, 0x65, 0x61, 0x74, 0x75,
	0x72, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x88, 0x01, 0x01, 0x12, 0x26, 0x0a, 0x04, 0x74, 0x61, 0x67,
	0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x73, 0x74, 0x61, 0x6e, 0x7a, 0x61,
	0x2e, 0x68, 0x75, 0x62, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x61, 0x67, 0x52, 0x04, 0x74, 0x61, 0x67,
	0x73, 0x42, 0x0f, 0x0a, 0x0d, 0x5f, 0x66, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x5f, 0x6e, 0x61,
	0x6d, 0x65, 0x22, 0xeb, 0x01, 0x0a, 0x18, 0x44, 0x65, 0x63, 0x6f, 0x72, 0x61, 0x74, 0x6f, 0x72,
	0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x12,
	0x25, 0x0a, 0x0b, 0x65, 0x6e, 0x76, 0x69, 0x72, 0x6f, 0x6e, 0x6d, 0x65, 0x6e, 0x74, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x0b, 0x65, 0x6e, 0x76, 0x69, 0x72,
	0x6f, 0x6e, 0x6d, 0x65, 0x6e, 0x74, 0x12, 0x2a, 0x0a, 0x0e, 0x64, 0x65, 0x63, 0x6f, 0x72, 0x61,
	0x74, 0x6f, 0x72, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03,
	0xe0, 0x41, 0x02, 0x52, 0x0d, 0x64, 0x65, 0x63, 0x6f, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x4e, 0x61,
	0x6d, 0x65, 0x12, 0x26, 0x0a, 0x0c, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x0b, 0x73,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x2c, 0x0a, 0x0f, 0x73, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x72, 0x65, 0x6c, 0x65, 0x61, 0x73, 0x65, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x09, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x0e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x52, 0x65, 0x6c, 0x65, 0x61, 0x73, 0x65, 0x12, 0x26, 0x0a, 0x04, 0x74, 0x61, 0x67, 0x73,
	0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x73, 0x74, 0x61, 0x6e, 0x7a, 0x61, 0x2e,
	0x68, 0x75, 0x62, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x61, 0x67, 0x52, 0x04, 0x74, 0x61, 0x67, 0x73,
	0x22, 0xa4, 0x01, 0x0a, 0x0f, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x53, 0x65, 0x6c, 0x65,
	0x63, 0x74, 0x6f, 0x72, 0x12, 0x25, 0x0a, 0x0b, 0x65, 0x6e, 0x76, 0x69, 0x72, 0x6f, 0x6e, 0x6d,
	0x65, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x0b,
	0x65, 0x6e, 0x76, 0x69, 0x72, 0x6f, 0x6e, 0x6d, 0x65, 0x6e, 0x74, 0x12, 0x17, 0x0a, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1d, 0x0a, 0x07, 0x72, 0x65, 0x6c, 0x65, 0x61, 0x73, 0x65, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x07, 0x72, 0x65, 0x6c, 0x65, 0x61, 0x73, 0x65,
	0x88, 0x01, 0x01, 0x12, 0x26, 0x0a, 0x04, 0x74, 0x61, 0x67, 0x73, 0x18, 0x06, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x12, 0x2e, 0x73, 0x74, 0x61, 0x6e, 0x7a, 0x61, 0x2e, 0x68, 0x75, 0x62, 0x2e, 0x76,
	0x31, 0x2e, 0x54, 0x61, 0x67, 0x52, 0x04, 0x74, 0x61, 0x67, 0x73, 0x42, 0x0a, 0x0a, 0x08, 0x5f,
	0x72, 0x65, 0x6c, 0x65, 0x61, 0x73, 0x65, 0x22, 0x5d, 0x0a, 0x10, 0x48, 0x65, 0x61, 0x6c, 0x74,
	0x68, 0x42, 0x79, 0x50, 0x72, 0x69, 0x6f, 0x72, 0x69, 0x74, 0x79, 0x12, 0x1a, 0x0a, 0x08, 0x70,
	0x72, 0x69, 0x6f, 0x72, 0x69, 0x74, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x08, 0x70,
	0x72, 0x69, 0x6f, 0x72, 0x69, 0x74, 0x79, 0x12, 0x2d, 0x0a, 0x06, 0x68, 0x65, 0x61, 0x6c, 0x74,
	0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x15, 0x2e, 0x73, 0x74, 0x61, 0x6e, 0x7a, 0x61,
	0x2e, 0x68, 0x75, 0x62, 0x2e, 0x76, 0x31, 0x2e, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x52, 0x06,
	0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x22, 0x2d, 0x0a, 0x03, 0x54, 0x61, 0x67, 0x12, 0x10, 0x0a,
	0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12,
	0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x2a, 0x55, 0x0a, 0x06, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x12,
	0x16, 0x0a, 0x12, 0x48, 0x45, 0x41, 0x4c, 0x54, 0x48, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43,
	0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x0d, 0x0a, 0x09, 0x48, 0x45, 0x41, 0x4c, 0x54,
	0x48, 0x5f, 0x4f, 0x4b, 0x10, 0x01, 0x12, 0x13, 0x0a, 0x0f, 0x48, 0x45, 0x41, 0x4c, 0x54, 0x48,
	0x5f, 0x4f, 0x56, 0x45, 0x52, 0x4c, 0x4f, 0x41, 0x44, 0x10, 0x02, 0x12, 0x0f, 0x0a, 0x0b, 0x48,
	0x45, 0x41, 0x4c, 0x54, 0x48, 0x5f, 0x44, 0x4f, 0x57, 0x4e, 0x10, 0x03, 0x2a, 0x45, 0x0a, 0x05,
	0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x15, 0x0a, 0x11, 0x53, 0x54, 0x41, 0x54, 0x45, 0x5f, 0x55,
	0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x11, 0x0a, 0x0d,
	0x53, 0x54, 0x41, 0x54, 0x45, 0x5f, 0x45, 0x4e, 0x41, 0x42, 0x4c, 0x45, 0x44, 0x10, 0x01, 0x12,
	0x12, 0x0a, 0x0e, 0x53, 0x54, 0x41, 0x54, 0x45, 0x5f, 0x44, 0x49, 0x53, 0x41, 0x42, 0x4c, 0x45,
	0x44, 0x10, 0x02, 0x42, 0xa6, 0x05, 0x92, 0x41, 0xf1, 0x03, 0x12, 0x4f, 0x0a, 0x0e, 0x53, 0x74,
	0x61, 0x6e, 0x7a, 0x61, 0x20, 0x48, 0x75, 0x62, 0x20, 0x41, 0x50, 0x49, 0x22, 0x38, 0x0a, 0x06,
	0x53, 0x74, 0x61, 0x6e, 0x7a, 0x61, 0x12, 0x16, 0x68, 0x74, 0x74, 0x70, 0x73, 0x3a, 0x2f, 0x2f,
	0x73, 0x74, 0x61, 0x6e, 0x7a, 0x61, 0x2e, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x73, 0x1a, 0x16,
	0x73, 0x75, 0x70, 0x70, 0x6f, 0x72, 0x74, 0x40, 0x73, 0x74, 0x61, 0x6e, 0x7a, 0x61, 0x2e, 0x73,
	0x79, 0x73, 0x74, 0x65, 0x6d, 0x73, 0x32, 0x03, 0x31, 0x2e, 0x30, 0x2a, 0x01, 0x02, 0x32, 0x10,
	0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x6a, 0x73, 0x6f, 0x6e,
	0x3a, 0x10, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x6a, 0x73,
	0x6f, 0x6e, 0x52, 0x55, 0x0a, 0x03, 0x34, 0x32, 0x39, 0x12, 0x4e, 0x0a, 0x2f, 0x54, 0x6f, 0x6f,
	0x20, 0x4d, 0x61, 0x6e, 0x79, 0x20, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x73, 0x2c, 0x20,
	0x74, 0x72, 0x79, 0x20, 0x61, 0x67, 0x61, 0x69, 0x6e, 0x20, 0x61, 0x66, 0x74, 0x65, 0x72, 0x20,
	0x52, 0x65, 0x74, 0x72, 0x79, 0x2d, 0x41, 0x66, 0x74, 0x65, 0x72, 0x2e, 0x12, 0x1b, 0x0a, 0x19,
	0x1a, 0x17, 0x23, 0x2f, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f,
	0x72, 0x70, 0x63, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x3b, 0x0a, 0x03, 0x35, 0x30, 0x30,
	0x12, 0x34, 0x0a, 0x15, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x20, 0x53, 0x65, 0x72,
	0x76, 0x65, 0x72, 0x20, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x1b, 0x0a, 0x19, 0x1a, 0x17, 0x23,
	0x2f, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x72, 0x70, 0x63,
	0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x57, 0x0a, 0x03, 0x35, 0x30, 0x33, 0x12, 0x50, 0x0a,
	0x31, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x20, 0x55, 0x6e, 0x61, 0x76, 0x61, 0x69, 0x6c,
	0x61, 0x62, 0x6c, 0x65, 0x2c, 0x20, 0x74, 0x72, 0x79, 0x20, 0x61, 0x67, 0x61, 0x69, 0x6e, 0x20,
	0x61, 0x66, 0x74, 0x65, 0x72, 0x20, 0x52, 0x65, 0x74, 0x72, 0x79, 0x2d, 0x41, 0x66, 0x74, 0x65,
	0x72, 0x2e, 0x12, 0x1b, 0x0a, 0x19, 0x1a, 0x17, 0x23, 0x2f, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x69,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x72, 0x70, 0x63, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x5a,
	0x22, 0x0a, 0x20, 0x0a, 0x0a, 0x41, 0x70, 0x69, 0x4b, 0x65, 0x79, 0x41, 0x75, 0x74, 0x68, 0x12,
	0x12, 0x08, 0x02, 0x1a, 0x0c, 0x58, 0x2d, 0x53, 0x74, 0x61, 0x6e, 0x7a, 0x61, 0x2d, 0x4b, 0x65,
	0x79, 0x20, 0x02, 0x62, 0x10, 0x0a, 0x0e, 0x0a, 0x0a, 0x41, 0x70, 0x69, 0x4b, 0x65, 0x79, 0x41,
	0x75, 0x74, 0x68, 0x12, 0x00, 0x72, 0x54, 0x0a, 0x0e, 0x53, 0x74, 0x61, 0x6e, 0x7a, 0x61, 0x20,
	0x48, 0x75, 0x62, 0x20, 0x41, 0x50, 0x49, 0x12, 0x42, 0x68, 0x74, 0x74, 0x70, 0x73, 0x3a, 0x2f,
	0x2f, 0x73, 0x74, 0x61, 0x6e, 0x7a, 0x61, 0x2e, 0x73, 0x74, 0x6f, 0x70, 0x6c, 0x69, 0x67, 0x68,
	0x74, 0x2e, 0x69, 0x6f, 0x2f, 0x64, 0x6f, 0x63, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x73, 0x2f, 0x32,
	0x39, 0x31, 0x61, 0x32, 0x63, 0x66, 0x66, 0x39, 0x64, 0x31, 0x35, 0x36, 0x2d, 0x73, 0x74, 0x61,
	0x6e, 0x7a, 0x61, 0x2d, 0x68, 0x75, 0x62, 0x2d, 0x61, 0x70, 0x69, 0x0a, 0x11, 0x63, 0x6f, 0x6d,
	0x2e, 0x73, 0x74, 0x61, 0x6e, 0x7a, 0x61, 0x2e, 0x68, 0x75, 0x62, 0x2e, 0x76, 0x31, 0x42, 0x0b,
	0x43, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x39, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x53, 0x74, 0x61, 0x6e, 0x7a, 0x61,
	0x53, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x73, 0x2f, 0x73, 0x64, 0x6b, 0x2d, 0x67, 0x6f, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x73, 0x74, 0x61, 0x6e, 0x7a, 0x61, 0x2f, 0x68, 0x75, 0x62, 0x2f,
	0x76, 0x31, 0x3b, 0x68, 0x75, 0x62, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x53, 0x48, 0x58, 0xaa, 0x02,
	0x0d, 0x53, 0x74, 0x61, 0x6e, 0x7a, 0x61, 0x2e, 0x48, 0x75, 0x62, 0x2e, 0x56, 0x31, 0xca, 0x02,
	0x0d, 0x53, 0x74, 0x61, 0x6e, 0x7a, 0x61, 0x5c, 0x48, 0x75, 0x62, 0x5c, 0x56, 0x31, 0xe2, 0x02,
	0x19, 0x53, 0x74, 0x61, 0x6e, 0x7a, 0x61, 0x5c, 0x48, 0x75, 0x62, 0x5c, 0x56, 0x31, 0x5c, 0x47,
	0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x0f, 0x53, 0x74, 0x61,
	0x6e, 0x7a, 0x61, 0x3a, 0x3a, 0x48, 0x75, 0x62, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_stanza_hub_v1_common_proto_rawDescOnce sync.Once
	file_stanza_hub_v1_common_proto_rawDescData = file_stanza_hub_v1_common_proto_rawDesc
)

func file_stanza_hub_v1_common_proto_rawDescGZIP() []byte {
	file_stanza_hub_v1_common_proto_rawDescOnce.Do(func() {
		file_stanza_hub_v1_common_proto_rawDescData = protoimpl.X.CompressGZIP(file_stanza_hub_v1_common_proto_rawDescData)
	})
	return file_stanza_hub_v1_common_proto_rawDescData
}

var file_stanza_hub_v1_common_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_stanza_hub_v1_common_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_stanza_hub_v1_common_proto_goTypes = []interface{}{
	(Health)(0),                      // 0: stanza.hub.v1.Health
	(State)(0),                       // 1: stanza.hub.v1.State
	(*DecoratorSelector)(nil),        // 2: stanza.hub.v1.DecoratorSelector
	(*FeatureSelector)(nil),          // 3: stanza.hub.v1.FeatureSelector
	(*DecoratorFeatureSelector)(nil), // 4: stanza.hub.v1.DecoratorFeatureSelector
	(*DecoratorServiceSelector)(nil), // 5: stanza.hub.v1.DecoratorServiceSelector
	(*ServiceSelector)(nil),          // 6: stanza.hub.v1.ServiceSelector
	(*HealthByPriority)(nil),         // 7: stanza.hub.v1.HealthByPriority
	(*Tag)(nil),                      // 8: stanza.hub.v1.Tag
}
var file_stanza_hub_v1_common_proto_depIdxs = []int32{
	8, // 0: stanza.hub.v1.DecoratorSelector.tags:type_name -> stanza.hub.v1.Tag
	8, // 1: stanza.hub.v1.FeatureSelector.tags:type_name -> stanza.hub.v1.Tag
	8, // 2: stanza.hub.v1.DecoratorFeatureSelector.tags:type_name -> stanza.hub.v1.Tag
	8, // 3: stanza.hub.v1.DecoratorServiceSelector.tags:type_name -> stanza.hub.v1.Tag
	8, // 4: stanza.hub.v1.ServiceSelector.tags:type_name -> stanza.hub.v1.Tag
	0, // 5: stanza.hub.v1.HealthByPriority.health:type_name -> stanza.hub.v1.Health
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_stanza_hub_v1_common_proto_init() }
func file_stanza_hub_v1_common_proto_init() {
	if File_stanza_hub_v1_common_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_stanza_hub_v1_common_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DecoratorSelector); i {
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
		file_stanza_hub_v1_common_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FeatureSelector); i {
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
		file_stanza_hub_v1_common_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DecoratorFeatureSelector); i {
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
		file_stanza_hub_v1_common_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DecoratorServiceSelector); i {
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
		file_stanza_hub_v1_common_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ServiceSelector); i {
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
		file_stanza_hub_v1_common_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HealthByPriority); i {
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
		file_stanza_hub_v1_common_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Tag); i {
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
	file_stanza_hub_v1_common_proto_msgTypes[2].OneofWrappers = []interface{}{}
	file_stanza_hub_v1_common_proto_msgTypes[4].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_stanza_hub_v1_common_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_stanza_hub_v1_common_proto_goTypes,
		DependencyIndexes: file_stanza_hub_v1_common_proto_depIdxs,
		EnumInfos:         file_stanza_hub_v1_common_proto_enumTypes,
		MessageInfos:      file_stanza_hub_v1_common_proto_msgTypes,
	}.Build()
	File_stanza_hub_v1_common_proto = out.File
	file_stanza_hub_v1_common_proto_rawDesc = nil
	file_stanza_hub_v1_common_proto_goTypes = nil
	file_stanza_hub_v1_common_proto_depIdxs = nil
}
