// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        (unknown)
// source: stanza/hub/v1/usage.proto

package hubv1

import (
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// MODE_SUM: query results across various axes (features, tags, apikeys) are added up and one timeseries is returned.
// MODE_REPORT: individual timeseries are returned for specified query axes.
// If not specified then queries will default to MODE_SUM for all axes.
type QueryMode int32

const (
	QueryMode_QUERY_MODE_UNSPECIFIED QueryMode = 0
	QueryMode_QUERY_MODE_SUM         QueryMode = 1
	QueryMode_QUERY_MODE_REPORT      QueryMode = 2
)

// Enum value maps for QueryMode.
var (
	QueryMode_name = map[int32]string{
		0: "QUERY_MODE_UNSPECIFIED",
		1: "QUERY_MODE_SUM",
		2: "QUERY_MODE_REPORT",
	}
	QueryMode_value = map[string]int32{
		"QUERY_MODE_UNSPECIFIED": 0,
		"QUERY_MODE_SUM":         1,
		"QUERY_MODE_REPORT":      2,
	}
)

func (x QueryMode) Enum() *QueryMode {
	p := new(QueryMode)
	*p = x
	return p
}

func (x QueryMode) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (QueryMode) Descriptor() protoreflect.EnumDescriptor {
	return file_stanza_hub_v1_usage_proto_enumTypes[0].Descriptor()
}

func (QueryMode) Type() protoreflect.EnumType {
	return &file_stanza_hub_v1_usage_proto_enumTypes[0]
}

func (x QueryMode) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use QueryMode.Descriptor instead.
func (QueryMode) EnumDescriptor() ([]byte, []int) {
	return file_stanza_hub_v1_usage_proto_rawDescGZIP(), []int{0}
}

// Usage query.
type GetUsageRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// If specified, only stats relating to the tags and features in selector will be returned.
	//
	//	If only guard and environment are specified, then stats relating to all tags and features will be returned.
	Environment       string                 `protobuf:"bytes,1,opt,name=environment,proto3" json:"environment,omitempty"`
	Guard             *string                `protobuf:"bytes,2,opt,name=guard,proto3,oneof" json:"guard,omitempty"` // Query for stats for this specific guard. If not specified then stats for all guards are returned.
	GuardQueryMode    *QueryMode             `protobuf:"varint,3,opt,name=guard_query_mode,json=guardQueryMode,proto3,enum=stanza.hub.v1.QueryMode,oneof" json:"guard_query_mode,omitempty"`
	StartTs           *timestamppb.Timestamp `protobuf:"bytes,4,opt,name=start_ts,json=startTs,proto3" json:"start_ts,omitempty"`
	EndTs             *timestamppb.Timestamp `protobuf:"bytes,5,opt,name=end_ts,json=endTs,proto3" json:"end_ts,omitempty"`
	Apikey            *string                `protobuf:"bytes,6,opt,name=apikey,proto3,oneof" json:"apikey,omitempty"`   // Query for stats where this specific APIKey was used. If not specified then stats for all APIKeys are returned.
	Feature           *string                `protobuf:"bytes,7,opt,name=feature,proto3,oneof" json:"feature,omitempty"` // Query for stats about a specific feature. If not specified then stats for all features are returned.
	FeatureQueryMode  *QueryMode             `protobuf:"varint,8,opt,name=feature_query_mode,json=featureQueryMode,proto3,enum=stanza.hub.v1.QueryMode,oneof" json:"feature_query_mode,omitempty"`
	Priority          *int32                 `protobuf:"varint,9,opt,name=priority,proto3,oneof" json:"priority,omitempty"` // Query for stats about a specific priority level. If not specified then stats for all priorities are returned.
	PriorityQueryMode *QueryMode             `protobuf:"varint,10,opt,name=priority_query_mode,json=priorityQueryMode,proto3,enum=stanza.hub.v1.QueryMode,oneof" json:"priority_query_mode,omitempty"`
	ReportTags        []string               `protobuf:"bytes,11,rep,name=report_tags,json=reportTags,proto3" json:"report_tags,omitempty"`                   // Tags matching listed tag keys will be reported (individual timeseries returned for each value).
	Tags              []*Tag                 `protobuf:"bytes,12,rep,name=tags,proto3" json:"tags,omitempty"`                                                 // Only stats relating to the specified tags will be returned.
	ReportAllTags     *bool                  `protobuf:"varint,13,opt,name=report_all_tags,json=reportAllTags,proto3,oneof" json:"report_all_tags,omitempty"` // Report all tag values for all tags as separate timeseries. Overrides tags and report_tags params.
	Step              *string                `protobuf:"bytes,14,opt,name=step,proto3,oneof" json:"step,omitempty"`                                           // 1m to 1w - m is minutes; h hours; d days; w weeks (7d). Defaults to a step that results in <100 results. Minimum step 1m.
}

func (x *GetUsageRequest) Reset() {
	*x = GetUsageRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stanza_hub_v1_usage_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetUsageRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetUsageRequest) ProtoMessage() {}

func (x *GetUsageRequest) ProtoReflect() protoreflect.Message {
	mi := &file_stanza_hub_v1_usage_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetUsageRequest.ProtoReflect.Descriptor instead.
func (*GetUsageRequest) Descriptor() ([]byte, []int) {
	return file_stanza_hub_v1_usage_proto_rawDescGZIP(), []int{0}
}

func (x *GetUsageRequest) GetEnvironment() string {
	if x != nil {
		return x.Environment
	}
	return ""
}

func (x *GetUsageRequest) GetGuard() string {
	if x != nil && x.Guard != nil {
		return *x.Guard
	}
	return ""
}

func (x *GetUsageRequest) GetGuardQueryMode() QueryMode {
	if x != nil && x.GuardQueryMode != nil {
		return *x.GuardQueryMode
	}
	return QueryMode_QUERY_MODE_UNSPECIFIED
}

func (x *GetUsageRequest) GetStartTs() *timestamppb.Timestamp {
	if x != nil {
		return x.StartTs
	}
	return nil
}

func (x *GetUsageRequest) GetEndTs() *timestamppb.Timestamp {
	if x != nil {
		return x.EndTs
	}
	return nil
}

func (x *GetUsageRequest) GetApikey() string {
	if x != nil && x.Apikey != nil {
		return *x.Apikey
	}
	return ""
}

func (x *GetUsageRequest) GetFeature() string {
	if x != nil && x.Feature != nil {
		return *x.Feature
	}
	return ""
}

func (x *GetUsageRequest) GetFeatureQueryMode() QueryMode {
	if x != nil && x.FeatureQueryMode != nil {
		return *x.FeatureQueryMode
	}
	return QueryMode_QUERY_MODE_UNSPECIFIED
}

func (x *GetUsageRequest) GetPriority() int32 {
	if x != nil && x.Priority != nil {
		return *x.Priority
	}
	return 0
}

func (x *GetUsageRequest) GetPriorityQueryMode() QueryMode {
	if x != nil && x.PriorityQueryMode != nil {
		return *x.PriorityQueryMode
	}
	return QueryMode_QUERY_MODE_UNSPECIFIED
}

func (x *GetUsageRequest) GetReportTags() []string {
	if x != nil {
		return x.ReportTags
	}
	return nil
}

func (x *GetUsageRequest) GetTags() []*Tag {
	if x != nil {
		return x.Tags
	}
	return nil
}

func (x *GetUsageRequest) GetReportAllTags() bool {
	if x != nil && x.ReportAllTags != nil {
		return *x.ReportAllTags
	}
	return false
}

func (x *GetUsageRequest) GetStep() string {
	if x != nil && x.Step != nil {
		return *x.Step
	}
	return ""
}

type GetUsageResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Result []*UsageTimeseries `protobuf:"bytes,1,rep,name=result,proto3" json:"result,omitempty"`
}

func (x *GetUsageResponse) Reset() {
	*x = GetUsageResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stanza_hub_v1_usage_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetUsageResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetUsageResponse) ProtoMessage() {}

func (x *GetUsageResponse) ProtoReflect() protoreflect.Message {
	mi := &file_stanza_hub_v1_usage_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetUsageResponse.ProtoReflect.Descriptor instead.
func (*GetUsageResponse) Descriptor() ([]byte, []int) {
	return file_stanza_hub_v1_usage_proto_rawDescGZIP(), []int{1}
}

func (x *GetUsageResponse) GetResult() []*UsageTimeseries {
	if x != nil {
		return x.Result
	}
	return nil
}

type UsageTimeseries struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data []*UsageTSDataPoint `protobuf:"bytes,1,rep,name=data,proto3" json:"data,omitempty"`
	// Axes for this timeseries - may not be applicable if there are no axes being queried in report mode
	Feature  *string `protobuf:"bytes,3,opt,name=feature,proto3,oneof" json:"feature,omitempty"`
	Priority *int32  `protobuf:"varint,4,opt,name=priority,proto3,oneof" json:"priority,omitempty"`
	Tags     []*Tag  `protobuf:"bytes,5,rep,name=tags,proto3" json:"tags,omitempty"`
	Guard    *string `protobuf:"bytes,6,opt,name=guard,proto3,oneof" json:"guard,omitempty"`
}

func (x *UsageTimeseries) Reset() {
	*x = UsageTimeseries{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stanza_hub_v1_usage_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UsageTimeseries) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UsageTimeseries) ProtoMessage() {}

func (x *UsageTimeseries) ProtoReflect() protoreflect.Message {
	mi := &file_stanza_hub_v1_usage_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UsageTimeseries.ProtoReflect.Descriptor instead.
func (*UsageTimeseries) Descriptor() ([]byte, []int) {
	return file_stanza_hub_v1_usage_proto_rawDescGZIP(), []int{2}
}

func (x *UsageTimeseries) GetData() []*UsageTSDataPoint {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *UsageTimeseries) GetFeature() string {
	if x != nil && x.Feature != nil {
		return *x.Feature
	}
	return ""
}

func (x *UsageTimeseries) GetPriority() int32 {
	if x != nil && x.Priority != nil {
		return *x.Priority
	}
	return 0
}

func (x *UsageTimeseries) GetTags() []*Tag {
	if x != nil {
		return x.Tags
	}
	return nil
}

func (x *UsageTimeseries) GetGuard() string {
	if x != nil && x.Guard != nil {
		return *x.Guard
	}
	return ""
}

type UsageTSDataPoint struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	StartTs            *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=start_ts,json=startTs,proto3" json:"start_ts,omitempty"`
	EndTs              *timestamppb.Timestamp `protobuf:"bytes,2,opt,name=end_ts,json=endTs,proto3" json:"end_ts,omitempty"`
	Granted            int32                  `protobuf:"varint,3,opt,name=granted,proto3" json:"granted,omitempty"`
	GrantedWeight      float32                `protobuf:"fixed32,4,opt,name=granted_weight,json=grantedWeight,proto3" json:"granted_weight,omitempty"`
	NotGranted         int32                  `protobuf:"varint,5,opt,name=not_granted,json=notGranted,proto3" json:"not_granted,omitempty"`
	NotGrantedWeight   float32                `protobuf:"fixed32,6,opt,name=not_granted_weight,json=notGrantedWeight,proto3" json:"not_granted_weight,omitempty"`
	BeBurst            *int32                 `protobuf:"varint,7,opt,name=be_burst,json=beBurst,proto3,oneof" json:"be_burst,omitempty"`
	BeBurstWeight      *float32               `protobuf:"fixed32,8,opt,name=be_burst_weight,json=beBurstWeight,proto3,oneof" json:"be_burst_weight,omitempty"`
	ParentReject       *int32                 `protobuf:"varint,9,opt,name=parent_reject,json=parentReject,proto3,oneof" json:"parent_reject,omitempty"`
	ParentRejectWeight *float32               `protobuf:"fixed32,10,opt,name=parent_reject_weight,json=parentRejectWeight,proto3,oneof" json:"parent_reject_weight,omitempty"`
}

func (x *UsageTSDataPoint) Reset() {
	*x = UsageTSDataPoint{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stanza_hub_v1_usage_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UsageTSDataPoint) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UsageTSDataPoint) ProtoMessage() {}

func (x *UsageTSDataPoint) ProtoReflect() protoreflect.Message {
	mi := &file_stanza_hub_v1_usage_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UsageTSDataPoint.ProtoReflect.Descriptor instead.
func (*UsageTSDataPoint) Descriptor() ([]byte, []int) {
	return file_stanza_hub_v1_usage_proto_rawDescGZIP(), []int{3}
}

func (x *UsageTSDataPoint) GetStartTs() *timestamppb.Timestamp {
	if x != nil {
		return x.StartTs
	}
	return nil
}

func (x *UsageTSDataPoint) GetEndTs() *timestamppb.Timestamp {
	if x != nil {
		return x.EndTs
	}
	return nil
}

func (x *UsageTSDataPoint) GetGranted() int32 {
	if x != nil {
		return x.Granted
	}
	return 0
}

func (x *UsageTSDataPoint) GetGrantedWeight() float32 {
	if x != nil {
		return x.GrantedWeight
	}
	return 0
}

func (x *UsageTSDataPoint) GetNotGranted() int32 {
	if x != nil {
		return x.NotGranted
	}
	return 0
}

func (x *UsageTSDataPoint) GetNotGrantedWeight() float32 {
	if x != nil {
		return x.NotGrantedWeight
	}
	return 0
}

func (x *UsageTSDataPoint) GetBeBurst() int32 {
	if x != nil && x.BeBurst != nil {
		return *x.BeBurst
	}
	return 0
}

func (x *UsageTSDataPoint) GetBeBurstWeight() float32 {
	if x != nil && x.BeBurstWeight != nil {
		return *x.BeBurstWeight
	}
	return 0
}

func (x *UsageTSDataPoint) GetParentReject() int32 {
	if x != nil && x.ParentReject != nil {
		return *x.ParentReject
	}
	return 0
}

func (x *UsageTSDataPoint) GetParentRejectWeight() float32 {
	if x != nil && x.ParentRejectWeight != nil {
		return *x.ParentRejectWeight
	}
	return 0
}

var File_stanza_hub_v1_usage_proto protoreflect.FileDescriptor

var file_stanza_hub_v1_usage_proto_rawDesc = []byte{
	0x0a, 0x19, 0x73, 0x74, 0x61, 0x6e, 0x7a, 0x61, 0x2f, 0x68, 0x75, 0x62, 0x2f, 0x76, 0x31, 0x2f,
	0x75, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0d, 0x73, 0x74, 0x61,
	0x6e, 0x7a, 0x61, 0x2e, 0x68, 0x75, 0x62, 0x2e, 0x76, 0x31, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x61, 0x70, 0x69, 0x2f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x62, 0x65, 0x68, 0x61, 0x76,
	0x69, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1a, 0x73, 0x74, 0x61, 0x6e,
	0x7a, 0x61, 0x2f, 0x68, 0x75, 0x62, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xa7, 0x06, 0x0a, 0x0f, 0x47, 0x65, 0x74, 0x55, 0x73,
	0x61, 0x67, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x25, 0x0a, 0x0b, 0x65, 0x6e,
	0x76, 0x69, 0x72, 0x6f, 0x6e, 0x6d, 0x65, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42,
	0x03, 0xe0, 0x41, 0x02, 0x52, 0x0b, 0x65, 0x6e, 0x76, 0x69, 0x72, 0x6f, 0x6e, 0x6d, 0x65, 0x6e,
	0x74, 0x12, 0x19, 0x0a, 0x05, 0x67, 0x75, 0x61, 0x72, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x48, 0x00, 0x52, 0x05, 0x67, 0x75, 0x61, 0x72, 0x64, 0x88, 0x01, 0x01, 0x12, 0x47, 0x0a, 0x10,
	0x67, 0x75, 0x61, 0x72, 0x64, 0x5f, 0x71, 0x75, 0x65, 0x72, 0x79, 0x5f, 0x6d, 0x6f, 0x64, 0x65,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x18, 0x2e, 0x73, 0x74, 0x61, 0x6e, 0x7a, 0x61, 0x2e,
	0x68, 0x75, 0x62, 0x2e, 0x76, 0x31, 0x2e, 0x51, 0x75, 0x65, 0x72, 0x79, 0x4d, 0x6f, 0x64, 0x65,
	0x48, 0x01, 0x52, 0x0e, 0x67, 0x75, 0x61, 0x72, 0x64, 0x51, 0x75, 0x65, 0x72, 0x79, 0x4d, 0x6f,
	0x64, 0x65, 0x88, 0x01, 0x01, 0x12, 0x3a, 0x0a, 0x08, 0x73, 0x74, 0x61, 0x72, 0x74, 0x5f, 0x74,
	0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x07, 0x73, 0x74, 0x61, 0x72, 0x74, 0x54,
	0x73, 0x12, 0x36, 0x0a, 0x06, 0x65, 0x6e, 0x64, 0x5f, 0x74, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x42, 0x03, 0xe0,
	0x41, 0x02, 0x52, 0x05, 0x65, 0x6e, 0x64, 0x54, 0x73, 0x12, 0x1b, 0x0a, 0x06, 0x61, 0x70, 0x69,
	0x6b, 0x65, 0x79, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x48, 0x02, 0x52, 0x06, 0x61, 0x70, 0x69,
	0x6b, 0x65, 0x79, 0x88, 0x01, 0x01, 0x12, 0x1d, 0x0a, 0x07, 0x66, 0x65, 0x61, 0x74, 0x75, 0x72,
	0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x48, 0x03, 0x52, 0x07, 0x66, 0x65, 0x61, 0x74, 0x75,
	0x72, 0x65, 0x88, 0x01, 0x01, 0x12, 0x4b, 0x0a, 0x12, 0x66, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65,
	0x5f, 0x71, 0x75, 0x65, 0x72, 0x79, 0x5f, 0x6d, 0x6f, 0x64, 0x65, 0x18, 0x08, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x18, 0x2e, 0x73, 0x74, 0x61, 0x6e, 0x7a, 0x61, 0x2e, 0x68, 0x75, 0x62, 0x2e, 0x76,
	0x31, 0x2e, 0x51, 0x75, 0x65, 0x72, 0x79, 0x4d, 0x6f, 0x64, 0x65, 0x48, 0x04, 0x52, 0x10, 0x66,
	0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x51, 0x75, 0x65, 0x72, 0x79, 0x4d, 0x6f, 0x64, 0x65, 0x88,
	0x01, 0x01, 0x12, 0x1f, 0x0a, 0x08, 0x70, 0x72, 0x69, 0x6f, 0x72, 0x69, 0x74, 0x79, 0x18, 0x09,
	0x20, 0x01, 0x28, 0x05, 0x48, 0x05, 0x52, 0x08, 0x70, 0x72, 0x69, 0x6f, 0x72, 0x69, 0x74, 0x79,
	0x88, 0x01, 0x01, 0x12, 0x4d, 0x0a, 0x13, 0x70, 0x72, 0x69, 0x6f, 0x72, 0x69, 0x74, 0x79, 0x5f,
	0x71, 0x75, 0x65, 0x72, 0x79, 0x5f, 0x6d, 0x6f, 0x64, 0x65, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x0e,
	0x32, 0x18, 0x2e, 0x73, 0x74, 0x61, 0x6e, 0x7a, 0x61, 0x2e, 0x68, 0x75, 0x62, 0x2e, 0x76, 0x31,
	0x2e, 0x51, 0x75, 0x65, 0x72, 0x79, 0x4d, 0x6f, 0x64, 0x65, 0x48, 0x06, 0x52, 0x11, 0x70, 0x72,
	0x69, 0x6f, 0x72, 0x69, 0x74, 0x79, 0x51, 0x75, 0x65, 0x72, 0x79, 0x4d, 0x6f, 0x64, 0x65, 0x88,
	0x01, 0x01, 0x12, 0x1f, 0x0a, 0x0b, 0x72, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x5f, 0x74, 0x61, 0x67,
	0x73, 0x18, 0x0b, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0a, 0x72, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x54,
	0x61, 0x67, 0x73, 0x12, 0x26, 0x0a, 0x04, 0x74, 0x61, 0x67, 0x73, 0x18, 0x0c, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x12, 0x2e, 0x73, 0x74, 0x61, 0x6e, 0x7a, 0x61, 0x2e, 0x68, 0x75, 0x62, 0x2e, 0x76,
	0x31, 0x2e, 0x54, 0x61, 0x67, 0x52, 0x04, 0x74, 0x61, 0x67, 0x73, 0x12, 0x2b, 0x0a, 0x0f, 0x72,
	0x65, 0x70, 0x6f, 0x72, 0x74, 0x5f, 0x61, 0x6c, 0x6c, 0x5f, 0x74, 0x61, 0x67, 0x73, 0x18, 0x0d,
	0x20, 0x01, 0x28, 0x08, 0x48, 0x07, 0x52, 0x0d, 0x72, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x41, 0x6c,
	0x6c, 0x54, 0x61, 0x67, 0x73, 0x88, 0x01, 0x01, 0x12, 0x17, 0x0a, 0x04, 0x73, 0x74, 0x65, 0x70,
	0x18, 0x0e, 0x20, 0x01, 0x28, 0x09, 0x48, 0x08, 0x52, 0x04, 0x73, 0x74, 0x65, 0x70, 0x88, 0x01,
	0x01, 0x42, 0x08, 0x0a, 0x06, 0x5f, 0x67, 0x75, 0x61, 0x72, 0x64, 0x42, 0x13, 0x0a, 0x11, 0x5f,
	0x67, 0x75, 0x61, 0x72, 0x64, 0x5f, 0x71, 0x75, 0x65, 0x72, 0x79, 0x5f, 0x6d, 0x6f, 0x64, 0x65,
	0x42, 0x09, 0x0a, 0x07, 0x5f, 0x61, 0x70, 0x69, 0x6b, 0x65, 0x79, 0x42, 0x0a, 0x0a, 0x08, 0x5f,
	0x66, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x42, 0x15, 0x0a, 0x13, 0x5f, 0x66, 0x65, 0x61, 0x74,
	0x75, 0x72, 0x65, 0x5f, 0x71, 0x75, 0x65, 0x72, 0x79, 0x5f, 0x6d, 0x6f, 0x64, 0x65, 0x42, 0x0b,
	0x0a, 0x09, 0x5f, 0x70, 0x72, 0x69, 0x6f, 0x72, 0x69, 0x74, 0x79, 0x42, 0x16, 0x0a, 0x14, 0x5f,
	0x70, 0x72, 0x69, 0x6f, 0x72, 0x69, 0x74, 0x79, 0x5f, 0x71, 0x75, 0x65, 0x72, 0x79, 0x5f, 0x6d,
	0x6f, 0x64, 0x65, 0x42, 0x12, 0x0a, 0x10, 0x5f, 0x72, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x5f, 0x61,
	0x6c, 0x6c, 0x5f, 0x74, 0x61, 0x67, 0x73, 0x42, 0x07, 0x0a, 0x05, 0x5f, 0x73, 0x74, 0x65, 0x70,
	0x22, 0x4a, 0x0a, 0x10, 0x47, 0x65, 0x74, 0x55, 0x73, 0x61, 0x67, 0x65, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x36, 0x0a, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x73, 0x74, 0x61, 0x6e, 0x7a, 0x61, 0x2e, 0x68, 0x75,
	0x62, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x73, 0x61, 0x67, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x65,
	0x72, 0x69, 0x65, 0x73, 0x52, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x22, 0xec, 0x01, 0x0a,
	0x0f, 0x55, 0x73, 0x61, 0x67, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x65, 0x72, 0x69, 0x65, 0x73,
	0x12, 0x33, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1f,
	0x2e, 0x73, 0x74, 0x61, 0x6e, 0x7a, 0x61, 0x2e, 0x68, 0x75, 0x62, 0x2e, 0x76, 0x31, 0x2e, 0x55,
	0x73, 0x61, 0x67, 0x65, 0x54, 0x53, 0x44, 0x61, 0x74, 0x61, 0x50, 0x6f, 0x69, 0x6e, 0x74, 0x52,
	0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x1d, 0x0a, 0x07, 0x66, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x07, 0x66, 0x65, 0x61, 0x74, 0x75, 0x72,
	0x65, 0x88, 0x01, 0x01, 0x12, 0x1f, 0x0a, 0x08, 0x70, 0x72, 0x69, 0x6f, 0x72, 0x69, 0x74, 0x79,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x48, 0x01, 0x52, 0x08, 0x70, 0x72, 0x69, 0x6f, 0x72, 0x69,
	0x74, 0x79, 0x88, 0x01, 0x01, 0x12, 0x26, 0x0a, 0x04, 0x74, 0x61, 0x67, 0x73, 0x18, 0x05, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x73, 0x74, 0x61, 0x6e, 0x7a, 0x61, 0x2e, 0x68, 0x75, 0x62,
	0x2e, 0x76, 0x31, 0x2e, 0x54, 0x61, 0x67, 0x52, 0x04, 0x74, 0x61, 0x67, 0x73, 0x12, 0x19, 0x0a,
	0x05, 0x67, 0x75, 0x61, 0x72, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x48, 0x02, 0x52, 0x05,
	0x67, 0x75, 0x61, 0x72, 0x64, 0x88, 0x01, 0x01, 0x42, 0x0a, 0x0a, 0x08, 0x5f, 0x66, 0x65, 0x61,
	0x74, 0x75, 0x72, 0x65, 0x42, 0x0b, 0x0a, 0x09, 0x5f, 0x70, 0x72, 0x69, 0x6f, 0x72, 0x69, 0x74,
	0x79, 0x42, 0x08, 0x0a, 0x06, 0x5f, 0x67, 0x75, 0x61, 0x72, 0x64, 0x22, 0x86, 0x04, 0x0a, 0x10,
	0x55, 0x73, 0x61, 0x67, 0x65, 0x54, 0x53, 0x44, 0x61, 0x74, 0x61, 0x50, 0x6f, 0x69, 0x6e, 0x74,
	0x12, 0x35, 0x0a, 0x08, 0x73, 0x74, 0x61, 0x72, 0x74, 0x5f, 0x74, 0x73, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x07,
	0x73, 0x74, 0x61, 0x72, 0x74, 0x54, 0x73, 0x12, 0x31, 0x0a, 0x06, 0x65, 0x6e, 0x64, 0x5f, 0x74,
	0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x52, 0x05, 0x65, 0x6e, 0x64, 0x54, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x67, 0x72,
	0x61, 0x6e, 0x74, 0x65, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x67, 0x72, 0x61,
	0x6e, 0x74, 0x65, 0x64, 0x12, 0x25, 0x0a, 0x0e, 0x67, 0x72, 0x61, 0x6e, 0x74, 0x65, 0x64, 0x5f,
	0x77, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x02, 0x52, 0x0d, 0x67, 0x72,
	0x61, 0x6e, 0x74, 0x65, 0x64, 0x57, 0x65, 0x69, 0x67, 0x68, 0x74, 0x12, 0x1f, 0x0a, 0x0b, 0x6e,
	0x6f, 0x74, 0x5f, 0x67, 0x72, 0x61, 0x6e, 0x74, 0x65, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x0a, 0x6e, 0x6f, 0x74, 0x47, 0x72, 0x61, 0x6e, 0x74, 0x65, 0x64, 0x12, 0x2c, 0x0a, 0x12,
	0x6e, 0x6f, 0x74, 0x5f, 0x67, 0x72, 0x61, 0x6e, 0x74, 0x65, 0x64, 0x5f, 0x77, 0x65, 0x69, 0x67,
	0x68, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x02, 0x52, 0x10, 0x6e, 0x6f, 0x74, 0x47, 0x72, 0x61,
	0x6e, 0x74, 0x65, 0x64, 0x57, 0x65, 0x69, 0x67, 0x68, 0x74, 0x12, 0x1e, 0x0a, 0x08, 0x62, 0x65,
	0x5f, 0x62, 0x75, 0x72, 0x73, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x05, 0x48, 0x00, 0x52, 0x07,
	0x62, 0x65, 0x42, 0x75, 0x72, 0x73, 0x74, 0x88, 0x01, 0x01, 0x12, 0x2b, 0x0a, 0x0f, 0x62, 0x65,
	0x5f, 0x62, 0x75, 0x72, 0x73, 0x74, 0x5f, 0x77, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x08, 0x20,
	0x01, 0x28, 0x02, 0x48, 0x01, 0x52, 0x0d, 0x62, 0x65, 0x42, 0x75, 0x72, 0x73, 0x74, 0x57, 0x65,
	0x69, 0x67, 0x68, 0x74, 0x88, 0x01, 0x01, 0x12, 0x28, 0x0a, 0x0d, 0x70, 0x61, 0x72, 0x65, 0x6e,
	0x74, 0x5f, 0x72, 0x65, 0x6a, 0x65, 0x63, 0x74, 0x18, 0x09, 0x20, 0x01, 0x28, 0x05, 0x48, 0x02,
	0x52, 0x0c, 0x70, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x6a, 0x65, 0x63, 0x74, 0x88, 0x01,
	0x01, 0x12, 0x35, 0x0a, 0x14, 0x70, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x5f, 0x72, 0x65, 0x6a, 0x65,
	0x63, 0x74, 0x5f, 0x77, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x02, 0x48,
	0x03, 0x52, 0x12, 0x70, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x6a, 0x65, 0x63, 0x74, 0x57,
	0x65, 0x69, 0x67, 0x68, 0x74, 0x88, 0x01, 0x01, 0x42, 0x0b, 0x0a, 0x09, 0x5f, 0x62, 0x65, 0x5f,
	0x62, 0x75, 0x72, 0x73, 0x74, 0x42, 0x12, 0x0a, 0x10, 0x5f, 0x62, 0x65, 0x5f, 0x62, 0x75, 0x72,
	0x73, 0x74, 0x5f, 0x77, 0x65, 0x69, 0x67, 0x68, 0x74, 0x42, 0x10, 0x0a, 0x0e, 0x5f, 0x70, 0x61,
	0x72, 0x65, 0x6e, 0x74, 0x5f, 0x72, 0x65, 0x6a, 0x65, 0x63, 0x74, 0x42, 0x17, 0x0a, 0x15, 0x5f,
	0x70, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x5f, 0x72, 0x65, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x77, 0x65,
	0x69, 0x67, 0x68, 0x74, 0x2a, 0x52, 0x0a, 0x09, 0x51, 0x75, 0x65, 0x72, 0x79, 0x4d, 0x6f, 0x64,
	0x65, 0x12, 0x1a, 0x0a, 0x16, 0x51, 0x55, 0x45, 0x52, 0x59, 0x5f, 0x4d, 0x4f, 0x44, 0x45, 0x5f,
	0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x12, 0x0a,
	0x0e, 0x51, 0x55, 0x45, 0x52, 0x59, 0x5f, 0x4d, 0x4f, 0x44, 0x45, 0x5f, 0x53, 0x55, 0x4d, 0x10,
	0x01, 0x12, 0x15, 0x0a, 0x11, 0x51, 0x55, 0x45, 0x52, 0x59, 0x5f, 0x4d, 0x4f, 0x44, 0x45, 0x5f,
	0x52, 0x45, 0x50, 0x4f, 0x52, 0x54, 0x10, 0x02, 0x32, 0x71, 0x0a, 0x0c, 0x55, 0x73, 0x61, 0x67,
	0x65, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x61, 0x0a, 0x08, 0x47, 0x65, 0x74, 0x55,
	0x73, 0x61, 0x67, 0x65, 0x12, 0x1e, 0x2e, 0x73, 0x74, 0x61, 0x6e, 0x7a, 0x61, 0x2e, 0x68, 0x75,
	0x62, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x55, 0x73, 0x61, 0x67, 0x65, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x1f, 0x2e, 0x73, 0x74, 0x61, 0x6e, 0x7a, 0x61, 0x2e, 0x68, 0x75,
	0x62, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x55, 0x73, 0x61, 0x67, 0x65, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x14, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x0e, 0x3a, 0x01, 0x2a,
	0x22, 0x09, 0x2f, 0x76, 0x31, 0x2f, 0x75, 0x73, 0x61, 0x67, 0x65, 0x42, 0xae, 0x01, 0x0a, 0x11,
	0x63, 0x6f, 0x6d, 0x2e, 0x73, 0x74, 0x61, 0x6e, 0x7a, 0x61, 0x2e, 0x68, 0x75, 0x62, 0x2e, 0x76,
	0x31, 0x42, 0x0a, 0x55, 0x73, 0x61, 0x67, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a,
	0x37, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x53, 0x74, 0x61, 0x6e,
	0x7a, 0x61, 0x53, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x73, 0x2f, 0x73, 0x64, 0x6b, 0x2d, 0x67, 0x6f,
	0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x73, 0x74, 0x61, 0x6e, 0x7a, 0x61, 0x2f, 0x68, 0x75, 0x62, 0x2f,
	0x76, 0x31, 0x3b, 0x68, 0x75, 0x62, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x53, 0x48, 0x58, 0xaa, 0x02,
	0x0d, 0x53, 0x74, 0x61, 0x6e, 0x7a, 0x61, 0x2e, 0x48, 0x75, 0x62, 0x2e, 0x56, 0x31, 0xca, 0x02,
	0x0d, 0x53, 0x74, 0x61, 0x6e, 0x7a, 0x61, 0x5c, 0x48, 0x75, 0x62, 0x5c, 0x56, 0x31, 0xe2, 0x02,
	0x19, 0x53, 0x74, 0x61, 0x6e, 0x7a, 0x61, 0x5c, 0x48, 0x75, 0x62, 0x5c, 0x56, 0x31, 0x5c, 0x47,
	0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x0f, 0x53, 0x74, 0x61,
	0x6e, 0x7a, 0x61, 0x3a, 0x3a, 0x48, 0x75, 0x62, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_stanza_hub_v1_usage_proto_rawDescOnce sync.Once
	file_stanza_hub_v1_usage_proto_rawDescData = file_stanza_hub_v1_usage_proto_rawDesc
)

func file_stanza_hub_v1_usage_proto_rawDescGZIP() []byte {
	file_stanza_hub_v1_usage_proto_rawDescOnce.Do(func() {
		file_stanza_hub_v1_usage_proto_rawDescData = protoimpl.X.CompressGZIP(file_stanza_hub_v1_usage_proto_rawDescData)
	})
	return file_stanza_hub_v1_usage_proto_rawDescData
}

var file_stanza_hub_v1_usage_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_stanza_hub_v1_usage_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_stanza_hub_v1_usage_proto_goTypes = []interface{}{
	(QueryMode)(0),                // 0: stanza.hub.v1.QueryMode
	(*GetUsageRequest)(nil),       // 1: stanza.hub.v1.GetUsageRequest
	(*GetUsageResponse)(nil),      // 2: stanza.hub.v1.GetUsageResponse
	(*UsageTimeseries)(nil),       // 3: stanza.hub.v1.UsageTimeseries
	(*UsageTSDataPoint)(nil),      // 4: stanza.hub.v1.UsageTSDataPoint
	(*timestamppb.Timestamp)(nil), // 5: google.protobuf.Timestamp
	(*Tag)(nil),                   // 6: stanza.hub.v1.Tag
}
var file_stanza_hub_v1_usage_proto_depIdxs = []int32{
	0,  // 0: stanza.hub.v1.GetUsageRequest.guard_query_mode:type_name -> stanza.hub.v1.QueryMode
	5,  // 1: stanza.hub.v1.GetUsageRequest.start_ts:type_name -> google.protobuf.Timestamp
	5,  // 2: stanza.hub.v1.GetUsageRequest.end_ts:type_name -> google.protobuf.Timestamp
	0,  // 3: stanza.hub.v1.GetUsageRequest.feature_query_mode:type_name -> stanza.hub.v1.QueryMode
	0,  // 4: stanza.hub.v1.GetUsageRequest.priority_query_mode:type_name -> stanza.hub.v1.QueryMode
	6,  // 5: stanza.hub.v1.GetUsageRequest.tags:type_name -> stanza.hub.v1.Tag
	3,  // 6: stanza.hub.v1.GetUsageResponse.result:type_name -> stanza.hub.v1.UsageTimeseries
	4,  // 7: stanza.hub.v1.UsageTimeseries.data:type_name -> stanza.hub.v1.UsageTSDataPoint
	6,  // 8: stanza.hub.v1.UsageTimeseries.tags:type_name -> stanza.hub.v1.Tag
	5,  // 9: stanza.hub.v1.UsageTSDataPoint.start_ts:type_name -> google.protobuf.Timestamp
	5,  // 10: stanza.hub.v1.UsageTSDataPoint.end_ts:type_name -> google.protobuf.Timestamp
	1,  // 11: stanza.hub.v1.UsageService.GetUsage:input_type -> stanza.hub.v1.GetUsageRequest
	2,  // 12: stanza.hub.v1.UsageService.GetUsage:output_type -> stanza.hub.v1.GetUsageResponse
	12, // [12:13] is the sub-list for method output_type
	11, // [11:12] is the sub-list for method input_type
	11, // [11:11] is the sub-list for extension type_name
	11, // [11:11] is the sub-list for extension extendee
	0,  // [0:11] is the sub-list for field type_name
}

func init() { file_stanza_hub_v1_usage_proto_init() }
func file_stanza_hub_v1_usage_proto_init() {
	if File_stanza_hub_v1_usage_proto != nil {
		return
	}
	file_stanza_hub_v1_common_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_stanza_hub_v1_usage_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetUsageRequest); i {
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
		file_stanza_hub_v1_usage_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetUsageResponse); i {
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
		file_stanza_hub_v1_usage_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UsageTimeseries); i {
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
		file_stanza_hub_v1_usage_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UsageTSDataPoint); i {
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
	file_stanza_hub_v1_usage_proto_msgTypes[0].OneofWrappers = []interface{}{}
	file_stanza_hub_v1_usage_proto_msgTypes[2].OneofWrappers = []interface{}{}
	file_stanza_hub_v1_usage_proto_msgTypes[3].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_stanza_hub_v1_usage_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_stanza_hub_v1_usage_proto_goTypes,
		DependencyIndexes: file_stanza_hub_v1_usage_proto_depIdxs,
		EnumInfos:         file_stanza_hub_v1_usage_proto_enumTypes,
		MessageInfos:      file_stanza_hub_v1_usage_proto_msgTypes,
	}.Build()
	File_stanza_hub_v1_usage_proto = out.File
	file_stanza_hub_v1_usage_proto_rawDesc = nil
	file_stanza_hub_v1_usage_proto_goTypes = nil
	file_stanza_hub_v1_usage_proto_depIdxs = nil
}
