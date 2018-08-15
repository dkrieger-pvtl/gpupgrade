// Code generated by protoc-gen-go. DO NOT EDIT.
// source: hub_to_agent.proto

package idl

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type UpgradeConvertPrimarySegmentsRequest struct {
	OldBinDir            string         `protobuf:"bytes,1,opt,name=OldBinDir,proto3" json:"OldBinDir,omitempty"`
	NewBinDir            string         `protobuf:"bytes,2,opt,name=NewBinDir,proto3" json:"NewBinDir,omitempty"`
	NewVersion           string         `protobuf:"bytes,3,opt,name=NewVersion,proto3" json:"NewVersion,omitempty"`
	DataDirPairs         []*DataDirPair `protobuf:"bytes,4,rep,name=DataDirPairs,proto3" json:"DataDirPairs,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *UpgradeConvertPrimarySegmentsRequest) Reset()         { *m = UpgradeConvertPrimarySegmentsRequest{} }
func (m *UpgradeConvertPrimarySegmentsRequest) String() string { return proto.CompactTextString(m) }
func (*UpgradeConvertPrimarySegmentsRequest) ProtoMessage()    {}
func (*UpgradeConvertPrimarySegmentsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_hub_to_agent_a0218dcec408dedf, []int{0}
}
func (m *UpgradeConvertPrimarySegmentsRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UpgradeConvertPrimarySegmentsRequest.Unmarshal(m, b)
}
func (m *UpgradeConvertPrimarySegmentsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UpgradeConvertPrimarySegmentsRequest.Marshal(b, m, deterministic)
}
func (dst *UpgradeConvertPrimarySegmentsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UpgradeConvertPrimarySegmentsRequest.Merge(dst, src)
}
func (m *UpgradeConvertPrimarySegmentsRequest) XXX_Size() int {
	return xxx_messageInfo_UpgradeConvertPrimarySegmentsRequest.Size(m)
}
func (m *UpgradeConvertPrimarySegmentsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_UpgradeConvertPrimarySegmentsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_UpgradeConvertPrimarySegmentsRequest proto.InternalMessageInfo

func (m *UpgradeConvertPrimarySegmentsRequest) GetOldBinDir() string {
	if m != nil {
		return m.OldBinDir
	}
	return ""
}

func (m *UpgradeConvertPrimarySegmentsRequest) GetNewBinDir() string {
	if m != nil {
		return m.NewBinDir
	}
	return ""
}

func (m *UpgradeConvertPrimarySegmentsRequest) GetNewVersion() string {
	if m != nil {
		return m.NewVersion
	}
	return ""
}

func (m *UpgradeConvertPrimarySegmentsRequest) GetDataDirPairs() []*DataDirPair {
	if m != nil {
		return m.DataDirPairs
	}
	return nil
}

type DataDirPair struct {
	OldDataDir           string   `protobuf:"bytes,1,opt,name=OldDataDir,proto3" json:"OldDataDir,omitempty"`
	NewDataDir           string   `protobuf:"bytes,2,opt,name=NewDataDir,proto3" json:"NewDataDir,omitempty"`
	OldPort              int32    `protobuf:"varint,3,opt,name=OldPort,proto3" json:"OldPort,omitempty"`
	NewPort              int32    `protobuf:"varint,4,opt,name=NewPort,proto3" json:"NewPort,omitempty"`
	Content              int32    `protobuf:"varint,5,opt,name=Content,proto3" json:"Content,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DataDirPair) Reset()         { *m = DataDirPair{} }
func (m *DataDirPair) String() string { return proto.CompactTextString(m) }
func (*DataDirPair) ProtoMessage()    {}
func (*DataDirPair) Descriptor() ([]byte, []int) {
	return fileDescriptor_hub_to_agent_a0218dcec408dedf, []int{1}
}
func (m *DataDirPair) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DataDirPair.Unmarshal(m, b)
}
func (m *DataDirPair) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DataDirPair.Marshal(b, m, deterministic)
}
func (dst *DataDirPair) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DataDirPair.Merge(dst, src)
}
func (m *DataDirPair) XXX_Size() int {
	return xxx_messageInfo_DataDirPair.Size(m)
}
func (m *DataDirPair) XXX_DiscardUnknown() {
	xxx_messageInfo_DataDirPair.DiscardUnknown(m)
}

var xxx_messageInfo_DataDirPair proto.InternalMessageInfo

func (m *DataDirPair) GetOldDataDir() string {
	if m != nil {
		return m.OldDataDir
	}
	return ""
}

func (m *DataDirPair) GetNewDataDir() string {
	if m != nil {
		return m.NewDataDir
	}
	return ""
}

func (m *DataDirPair) GetOldPort() int32 {
	if m != nil {
		return m.OldPort
	}
	return 0
}

func (m *DataDirPair) GetNewPort() int32 {
	if m != nil {
		return m.NewPort
	}
	return 0
}

func (m *DataDirPair) GetContent() int32 {
	if m != nil {
		return m.Content
	}
	return 0
}

type UpgradeConvertPrimarySegmentsReply struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *UpgradeConvertPrimarySegmentsReply) Reset()         { *m = UpgradeConvertPrimarySegmentsReply{} }
func (m *UpgradeConvertPrimarySegmentsReply) String() string { return proto.CompactTextString(m) }
func (*UpgradeConvertPrimarySegmentsReply) ProtoMessage()    {}
func (*UpgradeConvertPrimarySegmentsReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_hub_to_agent_a0218dcec408dedf, []int{2}
}
func (m *UpgradeConvertPrimarySegmentsReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UpgradeConvertPrimarySegmentsReply.Unmarshal(m, b)
}
func (m *UpgradeConvertPrimarySegmentsReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UpgradeConvertPrimarySegmentsReply.Marshal(b, m, deterministic)
}
func (dst *UpgradeConvertPrimarySegmentsReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UpgradeConvertPrimarySegmentsReply.Merge(dst, src)
}
func (m *UpgradeConvertPrimarySegmentsReply) XXX_Size() int {
	return xxx_messageInfo_UpgradeConvertPrimarySegmentsReply.Size(m)
}
func (m *UpgradeConvertPrimarySegmentsReply) XXX_DiscardUnknown() {
	xxx_messageInfo_UpgradeConvertPrimarySegmentsReply.DiscardUnknown(m)
}

var xxx_messageInfo_UpgradeConvertPrimarySegmentsReply proto.InternalMessageInfo

type PingAgentsRequest struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PingAgentsRequest) Reset()         { *m = PingAgentsRequest{} }
func (m *PingAgentsRequest) String() string { return proto.CompactTextString(m) }
func (*PingAgentsRequest) ProtoMessage()    {}
func (*PingAgentsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_hub_to_agent_a0218dcec408dedf, []int{3}
}
func (m *PingAgentsRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PingAgentsRequest.Unmarshal(m, b)
}
func (m *PingAgentsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PingAgentsRequest.Marshal(b, m, deterministic)
}
func (dst *PingAgentsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PingAgentsRequest.Merge(dst, src)
}
func (m *PingAgentsRequest) XXX_Size() int {
	return xxx_messageInfo_PingAgentsRequest.Size(m)
}
func (m *PingAgentsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_PingAgentsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_PingAgentsRequest proto.InternalMessageInfo

type PingAgentsReply struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PingAgentsReply) Reset()         { *m = PingAgentsReply{} }
func (m *PingAgentsReply) String() string { return proto.CompactTextString(m) }
func (*PingAgentsReply) ProtoMessage()    {}
func (*PingAgentsReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_hub_to_agent_a0218dcec408dedf, []int{4}
}
func (m *PingAgentsReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PingAgentsReply.Unmarshal(m, b)
}
func (m *PingAgentsReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PingAgentsReply.Marshal(b, m, deterministic)
}
func (dst *PingAgentsReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PingAgentsReply.Merge(dst, src)
}
func (m *PingAgentsReply) XXX_Size() int {
	return xxx_messageInfo_PingAgentsReply.Size(m)
}
func (m *PingAgentsReply) XXX_DiscardUnknown() {
	xxx_messageInfo_PingAgentsReply.DiscardUnknown(m)
}

var xxx_messageInfo_PingAgentsReply proto.InternalMessageInfo

type CheckUpgradeStatusRequest struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CheckUpgradeStatusRequest) Reset()         { *m = CheckUpgradeStatusRequest{} }
func (m *CheckUpgradeStatusRequest) String() string { return proto.CompactTextString(m) }
func (*CheckUpgradeStatusRequest) ProtoMessage()    {}
func (*CheckUpgradeStatusRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_hub_to_agent_a0218dcec408dedf, []int{5}
}
func (m *CheckUpgradeStatusRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CheckUpgradeStatusRequest.Unmarshal(m, b)
}
func (m *CheckUpgradeStatusRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CheckUpgradeStatusRequest.Marshal(b, m, deterministic)
}
func (dst *CheckUpgradeStatusRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CheckUpgradeStatusRequest.Merge(dst, src)
}
func (m *CheckUpgradeStatusRequest) XXX_Size() int {
	return xxx_messageInfo_CheckUpgradeStatusRequest.Size(m)
}
func (m *CheckUpgradeStatusRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_CheckUpgradeStatusRequest.DiscardUnknown(m)
}

var xxx_messageInfo_CheckUpgradeStatusRequest proto.InternalMessageInfo

type CheckUpgradeStatusReply struct {
	ProcessList          string   `protobuf:"bytes,1,opt,name=ProcessList,proto3" json:"ProcessList,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CheckUpgradeStatusReply) Reset()         { *m = CheckUpgradeStatusReply{} }
func (m *CheckUpgradeStatusReply) String() string { return proto.CompactTextString(m) }
func (*CheckUpgradeStatusReply) ProtoMessage()    {}
func (*CheckUpgradeStatusReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_hub_to_agent_a0218dcec408dedf, []int{6}
}
func (m *CheckUpgradeStatusReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CheckUpgradeStatusReply.Unmarshal(m, b)
}
func (m *CheckUpgradeStatusReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CheckUpgradeStatusReply.Marshal(b, m, deterministic)
}
func (dst *CheckUpgradeStatusReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CheckUpgradeStatusReply.Merge(dst, src)
}
func (m *CheckUpgradeStatusReply) XXX_Size() int {
	return xxx_messageInfo_CheckUpgradeStatusReply.Size(m)
}
func (m *CheckUpgradeStatusReply) XXX_DiscardUnknown() {
	xxx_messageInfo_CheckUpgradeStatusReply.DiscardUnknown(m)
}

var xxx_messageInfo_CheckUpgradeStatusReply proto.InternalMessageInfo

func (m *CheckUpgradeStatusReply) GetProcessList() string {
	if m != nil {
		return m.ProcessList
	}
	return ""
}

type CheckConversionStatusRequest struct {
	Segments             []*SegmentInfo `protobuf:"bytes,1,rep,name=Segments,proto3" json:"Segments,omitempty"`
	Hostname             string         `protobuf:"bytes,2,opt,name=Hostname,proto3" json:"Hostname,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *CheckConversionStatusRequest) Reset()         { *m = CheckConversionStatusRequest{} }
func (m *CheckConversionStatusRequest) String() string { return proto.CompactTextString(m) }
func (*CheckConversionStatusRequest) ProtoMessage()    {}
func (*CheckConversionStatusRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_hub_to_agent_a0218dcec408dedf, []int{7}
}
func (m *CheckConversionStatusRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CheckConversionStatusRequest.Unmarshal(m, b)
}
func (m *CheckConversionStatusRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CheckConversionStatusRequest.Marshal(b, m, deterministic)
}
func (dst *CheckConversionStatusRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CheckConversionStatusRequest.Merge(dst, src)
}
func (m *CheckConversionStatusRequest) XXX_Size() int {
	return xxx_messageInfo_CheckConversionStatusRequest.Size(m)
}
func (m *CheckConversionStatusRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_CheckConversionStatusRequest.DiscardUnknown(m)
}

var xxx_messageInfo_CheckConversionStatusRequest proto.InternalMessageInfo

func (m *CheckConversionStatusRequest) GetSegments() []*SegmentInfo {
	if m != nil {
		return m.Segments
	}
	return nil
}

func (m *CheckConversionStatusRequest) GetHostname() string {
	if m != nil {
		return m.Hostname
	}
	return ""
}

type SegmentInfo struct {
	Content              int32    `protobuf:"varint,1,opt,name=Content,proto3" json:"Content,omitempty"`
	Dbid                 int32    `protobuf:"varint,2,opt,name=Dbid,proto3" json:"Dbid,omitempty"`
	DataDir              string   `protobuf:"bytes,3,opt,name=DataDir,proto3" json:"DataDir,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SegmentInfo) Reset()         { *m = SegmentInfo{} }
func (m *SegmentInfo) String() string { return proto.CompactTextString(m) }
func (*SegmentInfo) ProtoMessage()    {}
func (*SegmentInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_hub_to_agent_a0218dcec408dedf, []int{8}
}
func (m *SegmentInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SegmentInfo.Unmarshal(m, b)
}
func (m *SegmentInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SegmentInfo.Marshal(b, m, deterministic)
}
func (dst *SegmentInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SegmentInfo.Merge(dst, src)
}
func (m *SegmentInfo) XXX_Size() int {
	return xxx_messageInfo_SegmentInfo.Size(m)
}
func (m *SegmentInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_SegmentInfo.DiscardUnknown(m)
}

var xxx_messageInfo_SegmentInfo proto.InternalMessageInfo

func (m *SegmentInfo) GetContent() int32 {
	if m != nil {
		return m.Content
	}
	return 0
}

func (m *SegmentInfo) GetDbid() int32 {
	if m != nil {
		return m.Dbid
	}
	return 0
}

func (m *SegmentInfo) GetDataDir() string {
	if m != nil {
		return m.DataDir
	}
	return ""
}

type CheckConversionStatusReply struct {
	Statuses             []string `protobuf:"bytes,1,rep,name=Statuses,proto3" json:"Statuses,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CheckConversionStatusReply) Reset()         { *m = CheckConversionStatusReply{} }
func (m *CheckConversionStatusReply) String() string { return proto.CompactTextString(m) }
func (*CheckConversionStatusReply) ProtoMessage()    {}
func (*CheckConversionStatusReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_hub_to_agent_a0218dcec408dedf, []int{9}
}
func (m *CheckConversionStatusReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CheckConversionStatusReply.Unmarshal(m, b)
}
func (m *CheckConversionStatusReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CheckConversionStatusReply.Marshal(b, m, deterministic)
}
func (dst *CheckConversionStatusReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CheckConversionStatusReply.Merge(dst, src)
}
func (m *CheckConversionStatusReply) XXX_Size() int {
	return xxx_messageInfo_CheckConversionStatusReply.Size(m)
}
func (m *CheckConversionStatusReply) XXX_DiscardUnknown() {
	xxx_messageInfo_CheckConversionStatusReply.DiscardUnknown(m)
}

var xxx_messageInfo_CheckConversionStatusReply proto.InternalMessageInfo

func (m *CheckConversionStatusReply) GetStatuses() []string {
	if m != nil {
		return m.Statuses
	}
	return nil
}

type FileSysUsage struct {
	Filesystem           string   `protobuf:"bytes,1,opt,name=Filesystem,proto3" json:"Filesystem,omitempty"`
	Usage                float64  `protobuf:"fixed64,2,opt,name=Usage,proto3" json:"Usage,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *FileSysUsage) Reset()         { *m = FileSysUsage{} }
func (m *FileSysUsage) String() string { return proto.CompactTextString(m) }
func (*FileSysUsage) ProtoMessage()    {}
func (*FileSysUsage) Descriptor() ([]byte, []int) {
	return fileDescriptor_hub_to_agent_a0218dcec408dedf, []int{10}
}
func (m *FileSysUsage) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FileSysUsage.Unmarshal(m, b)
}
func (m *FileSysUsage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FileSysUsage.Marshal(b, m, deterministic)
}
func (dst *FileSysUsage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FileSysUsage.Merge(dst, src)
}
func (m *FileSysUsage) XXX_Size() int {
	return xxx_messageInfo_FileSysUsage.Size(m)
}
func (m *FileSysUsage) XXX_DiscardUnknown() {
	xxx_messageInfo_FileSysUsage.DiscardUnknown(m)
}

var xxx_messageInfo_FileSysUsage proto.InternalMessageInfo

func (m *FileSysUsage) GetFilesystem() string {
	if m != nil {
		return m.Filesystem
	}
	return ""
}

func (m *FileSysUsage) GetUsage() float64 {
	if m != nil {
		return m.Usage
	}
	return 0
}

type CheckDiskSpaceRequestToAgent struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CheckDiskSpaceRequestToAgent) Reset()         { *m = CheckDiskSpaceRequestToAgent{} }
func (m *CheckDiskSpaceRequestToAgent) String() string { return proto.CompactTextString(m) }
func (*CheckDiskSpaceRequestToAgent) ProtoMessage()    {}
func (*CheckDiskSpaceRequestToAgent) Descriptor() ([]byte, []int) {
	return fileDescriptor_hub_to_agent_a0218dcec408dedf, []int{11}
}
func (m *CheckDiskSpaceRequestToAgent) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CheckDiskSpaceRequestToAgent.Unmarshal(m, b)
}
func (m *CheckDiskSpaceRequestToAgent) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CheckDiskSpaceRequestToAgent.Marshal(b, m, deterministic)
}
func (dst *CheckDiskSpaceRequestToAgent) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CheckDiskSpaceRequestToAgent.Merge(dst, src)
}
func (m *CheckDiskSpaceRequestToAgent) XXX_Size() int {
	return xxx_messageInfo_CheckDiskSpaceRequestToAgent.Size(m)
}
func (m *CheckDiskSpaceRequestToAgent) XXX_DiscardUnknown() {
	xxx_messageInfo_CheckDiskSpaceRequestToAgent.DiscardUnknown(m)
}

var xxx_messageInfo_CheckDiskSpaceRequestToAgent proto.InternalMessageInfo

type CheckDiskSpaceReplyFromAgent struct {
	ListOfFileSysUsage   []*FileSysUsage `protobuf:"bytes,1,rep,name=ListOfFileSysUsage,proto3" json:"ListOfFileSysUsage,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *CheckDiskSpaceReplyFromAgent) Reset()         { *m = CheckDiskSpaceReplyFromAgent{} }
func (m *CheckDiskSpaceReplyFromAgent) String() string { return proto.CompactTextString(m) }
func (*CheckDiskSpaceReplyFromAgent) ProtoMessage()    {}
func (*CheckDiskSpaceReplyFromAgent) Descriptor() ([]byte, []int) {
	return fileDescriptor_hub_to_agent_a0218dcec408dedf, []int{12}
}
func (m *CheckDiskSpaceReplyFromAgent) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CheckDiskSpaceReplyFromAgent.Unmarshal(m, b)
}
func (m *CheckDiskSpaceReplyFromAgent) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CheckDiskSpaceReplyFromAgent.Marshal(b, m, deterministic)
}
func (dst *CheckDiskSpaceReplyFromAgent) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CheckDiskSpaceReplyFromAgent.Merge(dst, src)
}
func (m *CheckDiskSpaceReplyFromAgent) XXX_Size() int {
	return xxx_messageInfo_CheckDiskSpaceReplyFromAgent.Size(m)
}
func (m *CheckDiskSpaceReplyFromAgent) XXX_DiscardUnknown() {
	xxx_messageInfo_CheckDiskSpaceReplyFromAgent.DiscardUnknown(m)
}

var xxx_messageInfo_CheckDiskSpaceReplyFromAgent proto.InternalMessageInfo

func (m *CheckDiskSpaceReplyFromAgent) GetListOfFileSysUsage() []*FileSysUsage {
	if m != nil {
		return m.ListOfFileSysUsage
	}
	return nil
}

type CreateSegmentDataDirRequest struct {
	Datadirs             []string `protobuf:"bytes,1,rep,name=datadirs,proto3" json:"datadirs,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CreateSegmentDataDirRequest) Reset()         { *m = CreateSegmentDataDirRequest{} }
func (m *CreateSegmentDataDirRequest) String() string { return proto.CompactTextString(m) }
func (*CreateSegmentDataDirRequest) ProtoMessage()    {}
func (*CreateSegmentDataDirRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_hub_to_agent_a0218dcec408dedf, []int{13}
}
func (m *CreateSegmentDataDirRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CreateSegmentDataDirRequest.Unmarshal(m, b)
}
func (m *CreateSegmentDataDirRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CreateSegmentDataDirRequest.Marshal(b, m, deterministic)
}
func (dst *CreateSegmentDataDirRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CreateSegmentDataDirRequest.Merge(dst, src)
}
func (m *CreateSegmentDataDirRequest) XXX_Size() int {
	return xxx_messageInfo_CreateSegmentDataDirRequest.Size(m)
}
func (m *CreateSegmentDataDirRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_CreateSegmentDataDirRequest.DiscardUnknown(m)
}

var xxx_messageInfo_CreateSegmentDataDirRequest proto.InternalMessageInfo

func (m *CreateSegmentDataDirRequest) GetDatadirs() []string {
	if m != nil {
		return m.Datadirs
	}
	return nil
}

type CreateSegmentDataDirReply struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CreateSegmentDataDirReply) Reset()         { *m = CreateSegmentDataDirReply{} }
func (m *CreateSegmentDataDirReply) String() string { return proto.CompactTextString(m) }
func (*CreateSegmentDataDirReply) ProtoMessage()    {}
func (*CreateSegmentDataDirReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_hub_to_agent_a0218dcec408dedf, []int{14}
}
func (m *CreateSegmentDataDirReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CreateSegmentDataDirReply.Unmarshal(m, b)
}
func (m *CreateSegmentDataDirReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CreateSegmentDataDirReply.Marshal(b, m, deterministic)
}
func (dst *CreateSegmentDataDirReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CreateSegmentDataDirReply.Merge(dst, src)
}
func (m *CreateSegmentDataDirReply) XXX_Size() int {
	return xxx_messageInfo_CreateSegmentDataDirReply.Size(m)
}
func (m *CreateSegmentDataDirReply) XXX_DiscardUnknown() {
	xxx_messageInfo_CreateSegmentDataDirReply.DiscardUnknown(m)
}

var xxx_messageInfo_CreateSegmentDataDirReply proto.InternalMessageInfo

func init() {
	proto.RegisterType((*UpgradeConvertPrimarySegmentsRequest)(nil), "idl.UpgradeConvertPrimarySegmentsRequest")
	proto.RegisterType((*DataDirPair)(nil), "idl.DataDirPair")
	proto.RegisterType((*UpgradeConvertPrimarySegmentsReply)(nil), "idl.UpgradeConvertPrimarySegmentsReply")
	proto.RegisterType((*PingAgentsRequest)(nil), "idl.PingAgentsRequest")
	proto.RegisterType((*PingAgentsReply)(nil), "idl.PingAgentsReply")
	proto.RegisterType((*CheckUpgradeStatusRequest)(nil), "idl.CheckUpgradeStatusRequest")
	proto.RegisterType((*CheckUpgradeStatusReply)(nil), "idl.CheckUpgradeStatusReply")
	proto.RegisterType((*CheckConversionStatusRequest)(nil), "idl.CheckConversionStatusRequest")
	proto.RegisterType((*SegmentInfo)(nil), "idl.SegmentInfo")
	proto.RegisterType((*CheckConversionStatusReply)(nil), "idl.CheckConversionStatusReply")
	proto.RegisterType((*FileSysUsage)(nil), "idl.FileSysUsage")
	proto.RegisterType((*CheckDiskSpaceRequestToAgent)(nil), "idl.CheckDiskSpaceRequestToAgent")
	proto.RegisterType((*CheckDiskSpaceReplyFromAgent)(nil), "idl.CheckDiskSpaceReplyFromAgent")
	proto.RegisterType((*CreateSegmentDataDirRequest)(nil), "idl.CreateSegmentDataDirRequest")
	proto.RegisterType((*CreateSegmentDataDirReply)(nil), "idl.CreateSegmentDataDirReply")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// AgentClient is the client API for Agent service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type AgentClient interface {
	CheckUpgradeStatus(ctx context.Context, in *CheckUpgradeStatusRequest, opts ...grpc.CallOption) (*CheckUpgradeStatusReply, error)
	CheckConversionStatus(ctx context.Context, in *CheckConversionStatusRequest, opts ...grpc.CallOption) (*CheckConversionStatusReply, error)
	CheckDiskSpaceOnAgents(ctx context.Context, in *CheckDiskSpaceRequestToAgent, opts ...grpc.CallOption) (*CheckDiskSpaceReplyFromAgent, error)
	PingAgents(ctx context.Context, in *PingAgentsRequest, opts ...grpc.CallOption) (*PingAgentsReply, error)
	UpgradeConvertPrimarySegments(ctx context.Context, in *UpgradeConvertPrimarySegmentsRequest, opts ...grpc.CallOption) (*UpgradeConvertPrimarySegmentsReply, error)
	CreateSegmentDataDirectories(ctx context.Context, in *CreateSegmentDataDirRequest, opts ...grpc.CallOption) (*CreateSegmentDataDirReply, error)
}

type agentClient struct {
	cc *grpc.ClientConn
}

func NewAgentClient(cc *grpc.ClientConn) AgentClient {
	return &agentClient{cc}
}

func (c *agentClient) CheckUpgradeStatus(ctx context.Context, in *CheckUpgradeStatusRequest, opts ...grpc.CallOption) (*CheckUpgradeStatusReply, error) {
	out := new(CheckUpgradeStatusReply)
	err := c.cc.Invoke(ctx, "/idl.Agent/CheckUpgradeStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentClient) CheckConversionStatus(ctx context.Context, in *CheckConversionStatusRequest, opts ...grpc.CallOption) (*CheckConversionStatusReply, error) {
	out := new(CheckConversionStatusReply)
	err := c.cc.Invoke(ctx, "/idl.Agent/CheckConversionStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentClient) CheckDiskSpaceOnAgents(ctx context.Context, in *CheckDiskSpaceRequestToAgent, opts ...grpc.CallOption) (*CheckDiskSpaceReplyFromAgent, error) {
	out := new(CheckDiskSpaceReplyFromAgent)
	err := c.cc.Invoke(ctx, "/idl.Agent/CheckDiskSpaceOnAgents", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentClient) PingAgents(ctx context.Context, in *PingAgentsRequest, opts ...grpc.CallOption) (*PingAgentsReply, error) {
	out := new(PingAgentsReply)
	err := c.cc.Invoke(ctx, "/idl.Agent/PingAgents", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentClient) UpgradeConvertPrimarySegments(ctx context.Context, in *UpgradeConvertPrimarySegmentsRequest, opts ...grpc.CallOption) (*UpgradeConvertPrimarySegmentsReply, error) {
	out := new(UpgradeConvertPrimarySegmentsReply)
	err := c.cc.Invoke(ctx, "/idl.Agent/UpgradeConvertPrimarySegments", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentClient) CreateSegmentDataDirectories(ctx context.Context, in *CreateSegmentDataDirRequest, opts ...grpc.CallOption) (*CreateSegmentDataDirReply, error) {
	out := new(CreateSegmentDataDirReply)
	err := c.cc.Invoke(ctx, "/idl.Agent/CreateSegmentDataDirectories", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AgentServer is the server API for Agent service.
type AgentServer interface {
	CheckUpgradeStatus(context.Context, *CheckUpgradeStatusRequest) (*CheckUpgradeStatusReply, error)
	CheckConversionStatus(context.Context, *CheckConversionStatusRequest) (*CheckConversionStatusReply, error)
	CheckDiskSpaceOnAgents(context.Context, *CheckDiskSpaceRequestToAgent) (*CheckDiskSpaceReplyFromAgent, error)
	PingAgents(context.Context, *PingAgentsRequest) (*PingAgentsReply, error)
	UpgradeConvertPrimarySegments(context.Context, *UpgradeConvertPrimarySegmentsRequest) (*UpgradeConvertPrimarySegmentsReply, error)
	CreateSegmentDataDirectories(context.Context, *CreateSegmentDataDirRequest) (*CreateSegmentDataDirReply, error)
}

func RegisterAgentServer(s *grpc.Server, srv AgentServer) {
	s.RegisterService(&_Agent_serviceDesc, srv)
}

func _Agent_CheckUpgradeStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CheckUpgradeStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServer).CheckUpgradeStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/idl.Agent/CheckUpgradeStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServer).CheckUpgradeStatus(ctx, req.(*CheckUpgradeStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Agent_CheckConversionStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CheckConversionStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServer).CheckConversionStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/idl.Agent/CheckConversionStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServer).CheckConversionStatus(ctx, req.(*CheckConversionStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Agent_CheckDiskSpaceOnAgents_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CheckDiskSpaceRequestToAgent)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServer).CheckDiskSpaceOnAgents(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/idl.Agent/CheckDiskSpaceOnAgents",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServer).CheckDiskSpaceOnAgents(ctx, req.(*CheckDiskSpaceRequestToAgent))
	}
	return interceptor(ctx, in, info, handler)
}

func _Agent_PingAgents_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingAgentsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServer).PingAgents(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/idl.Agent/PingAgents",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServer).PingAgents(ctx, req.(*PingAgentsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Agent_UpgradeConvertPrimarySegments_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpgradeConvertPrimarySegmentsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServer).UpgradeConvertPrimarySegments(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/idl.Agent/UpgradeConvertPrimarySegments",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServer).UpgradeConvertPrimarySegments(ctx, req.(*UpgradeConvertPrimarySegmentsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Agent_CreateSegmentDataDirectories_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateSegmentDataDirRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServer).CreateSegmentDataDirectories(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/idl.Agent/CreateSegmentDataDirectories",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServer).CreateSegmentDataDirectories(ctx, req.(*CreateSegmentDataDirRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Agent_serviceDesc = grpc.ServiceDesc{
	ServiceName: "idl.Agent",
	HandlerType: (*AgentServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CheckUpgradeStatus",
			Handler:    _Agent_CheckUpgradeStatus_Handler,
		},
		{
			MethodName: "CheckConversionStatus",
			Handler:    _Agent_CheckConversionStatus_Handler,
		},
		{
			MethodName: "CheckDiskSpaceOnAgents",
			Handler:    _Agent_CheckDiskSpaceOnAgents_Handler,
		},
		{
			MethodName: "PingAgents",
			Handler:    _Agent_PingAgents_Handler,
		},
		{
			MethodName: "UpgradeConvertPrimarySegments",
			Handler:    _Agent_UpgradeConvertPrimarySegments_Handler,
		},
		{
			MethodName: "CreateSegmentDataDirectories",
			Handler:    _Agent_CreateSegmentDataDirectories_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "hub_to_agent.proto",
}

func init() { proto.RegisterFile("hub_to_agent.proto", fileDescriptor_hub_to_agent_a0218dcec408dedf) }

var fileDescriptor_hub_to_agent_a0218dcec408dedf = []byte{
	// 638 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x55, 0x5d, 0x6f, 0xd3, 0x30,
	0x14, 0x5d, 0xb6, 0x16, 0xba, 0xdb, 0x49, 0x30, 0x33, 0x46, 0xc9, 0x4a, 0x29, 0xd6, 0x24, 0x86,
	0x84, 0xf6, 0x30, 0x78, 0x00, 0xc1, 0xcb, 0x68, 0x35, 0x81, 0x84, 0xd6, 0x2a, 0xdd, 0x78, 0x43,
	0xc3, 0x6d, 0xbc, 0xce, 0x5a, 0x1a, 0x07, 0xdb, 0x65, 0xca, 0x6f, 0xe1, 0x87, 0xc0, 0xcf, 0x43,
	0xfe, 0x48, 0xeb, 0xaa, 0xcd, 0xd8, 0x5b, 0xef, 0x39, 0xe7, 0x3a, 0x27, 0xe7, 0x5e, 0xa7, 0x80,
	0xae, 0xa6, 0xc3, 0x0b, 0xc5, 0x2f, 0xc8, 0x98, 0xa6, 0xea, 0x30, 0x13, 0x5c, 0x71, 0xb4, 0xc1,
	0xe2, 0x04, 0xff, 0x0d, 0x60, 0xff, 0x3c, 0x1b, 0x0b, 0x12, 0xd3, 0x0e, 0x4f, 0x7f, 0x51, 0xa1,
	0xfa, 0x82, 0x4d, 0x88, 0xc8, 0x07, 0x74, 0x3c, 0xa1, 0xa9, 0x92, 0x11, 0xfd, 0x39, 0xa5, 0x52,
	0xa1, 0x26, 0x6c, 0xf6, 0x92, 0xf8, 0x13, 0x4b, 0xbb, 0x4c, 0x34, 0x82, 0x76, 0x70, 0xb0, 0x19,
	0xcd, 0x01, 0xcd, 0x9e, 0xd2, 0x1b, 0xc7, 0xae, 0x5b, 0x76, 0x06, 0xa0, 0x16, 0xc0, 0x29, 0xbd,
	0xf9, 0x46, 0x85, 0x64, 0x3c, 0x6d, 0x6c, 0x18, 0xda, 0x43, 0xd0, 0x5b, 0xd8, 0xea, 0x12, 0x45,
	0xba, 0x4c, 0xf4, 0x09, 0x13, 0xb2, 0x51, 0x69, 0x6f, 0x1c, 0xd4, 0x8f, 0x1e, 0x1e, 0xb2, 0x38,
	0x39, 0xf4, 0x88, 0x68, 0x41, 0x85, 0x7f, 0x07, 0x50, 0xf7, 0x00, 0xfd, 0x94, 0x5e, 0x12, 0x3b,
	0xc4, 0x59, 0xf4, 0x10, 0xe7, 0xa2, 0xe0, 0xd7, 0x67, 0x2e, 0x0a, 0xbe, 0x01, 0xf7, 0x7b, 0x49,
	0xdc, 0xe7, 0x42, 0x19, 0x8b, 0xd5, 0xa8, 0x28, 0x35, 0x73, 0x4a, 0x6f, 0x0c, 0x53, 0xb1, 0x8c,
	0x2b, 0x35, 0xd3, 0xe1, 0xa9, 0xa2, 0xa9, 0x6a, 0x54, 0x2d, 0xe3, 0x4a, 0xbc, 0x0f, 0xf8, 0x3f,
	0xb9, 0x66, 0x49, 0x8e, 0x1f, 0xc1, 0x76, 0x9f, 0xa5, 0xe3, 0xe3, 0xb1, 0x17, 0x35, 0xde, 0x86,
	0x07, 0x3e, 0xa8, 0x75, 0x7b, 0xf0, 0xb4, 0x73, 0x45, 0x47, 0xd7, 0xee, 0xc8, 0x81, 0x22, 0x6a,
	0x3a, 0xd3, 0x7f, 0x80, 0x27, 0xab, 0xc8, 0x2c, 0xc9, 0x51, 0x1b, 0xea, 0x7d, 0xc1, 0x47, 0x54,
	0xca, 0xaf, 0x4c, 0x2a, 0x17, 0x8a, 0x0f, 0xe1, 0x2b, 0x68, 0x9a, 0x66, 0xeb, 0x52, 0x8f, 0x63,
	0xe1, 0x70, 0xf4, 0x1a, 0x6a, 0x85, 0xe5, 0x46, 0xe0, 0xcd, 0xc5, 0x81, 0x5f, 0xd2, 0x4b, 0x1e,
	0xcd, 0x14, 0x28, 0x84, 0xda, 0x67, 0x2e, 0x55, 0x4a, 0x26, 0xd4, 0x25, 0x3c, 0xab, 0xf1, 0x39,
	0xd4, 0xbd, 0x26, 0x3f, 0xba, 0x60, 0x21, 0x3a, 0x84, 0xa0, 0xd2, 0x1d, 0xb2, 0xd8, 0x1c, 0x50,
	0x8d, 0xcc, 0x6f, 0xad, 0x2e, 0x26, 0x67, 0xf7, 0xa7, 0x28, 0xf1, 0x3b, 0x08, 0x4b, 0x5e, 0x40,
	0x07, 0x10, 0x42, 0xcd, 0x96, 0xd4, 0xda, 0xdf, 0x8c, 0x66, 0x35, 0xee, 0xc2, 0xd6, 0x09, 0x4b,
	0xe8, 0x20, 0x97, 0xe7, 0x92, 0x8c, 0xa9, 0x5e, 0x10, 0x5d, 0xcb, 0x5c, 0x2a, 0x3a, 0x29, 0x16,
	0x68, 0x8e, 0xa0, 0x1d, 0xa8, 0x1a, 0xa1, 0x31, 0x16, 0x44, 0xb6, 0xc0, 0x2d, 0x17, 0x60, 0x97,
	0xc9, 0xeb, 0x41, 0x46, 0x46, 0xd4, 0x25, 0x77, 0xc6, 0xcd, 0x00, 0x31, 0x59, 0xe6, 0xb3, 0x24,
	0x3f, 0x11, 0x7c, 0x62, 0x78, 0x74, 0x0c, 0x48, 0x0f, 0xa2, 0x77, 0xe9, 0x7b, 0x71, 0x51, 0x6f,
	0x9b, 0xa8, 0x7d, 0x22, 0x5a, 0x21, 0xc6, 0xef, 0x61, 0xaf, 0x23, 0x28, 0x51, 0xd4, 0xe5, 0xeb,
	0xa2, 0x29, 0x46, 0x18, 0x42, 0x2d, 0x26, 0x8a, 0xc4, 0xfa, 0x6a, 0xb9, 0x0c, 0x8a, 0xda, 0x2c,
	0xd6, 0xca, 0xd6, 0x2c, 0xc9, 0x8f, 0xfe, 0x54, 0xa0, 0x6a, 0x4d, 0x9e, 0x01, 0x5a, 0x5e, 0x31,
	0xd4, 0x32, 0xf6, 0x4a, 0x17, 0x33, 0x6c, 0x96, 0xf2, 0x7a, 0xa7, 0xd7, 0xd0, 0x77, 0x78, 0xbc,
	0x72, 0x74, 0xe8, 0xc5, 0xbc, 0xb1, 0x64, 0x2f, 0xc3, 0xe7, 0xb7, 0x49, 0xec, 0xf1, 0x3f, 0x60,
	0x77, 0x31, 0xf9, 0x5e, 0x6a, 0xef, 0x94, 0x7f, 0x7e, 0xc9, 0xd8, 0xc2, 0xd5, 0x12, 0x7f, 0x72,
	0x78, 0x0d, 0x7d, 0x04, 0x98, 0xdf, 0x54, 0xb4, 0x6b, 0x5a, 0x96, 0xee, 0x73, 0xb8, 0xb3, 0x84,
	0x5b, 0x7f, 0x53, 0x78, 0x76, 0xeb, 0x27, 0x02, 0xbd, 0x32, 0x8d, 0x77, 0xf9, 0x3c, 0x87, 0x2f,
	0xef, 0x22, 0xb5, 0x8f, 0x1d, 0x42, 0x73, 0xd5, 0xc8, 0xe9, 0x48, 0x71, 0xc1, 0xa8, 0x44, 0x6d,
	0xfb, 0xe6, 0xe5, 0x0b, 0x15, 0xb6, 0x6e, 0x51, 0x98, 0x67, 0x0c, 0xef, 0x99, 0xbf, 0x98, 0x37,
	0xff, 0x02, 0x00, 0x00, 0xff, 0xff, 0x06, 0x30, 0xda, 0x8a, 0x78, 0x06, 0x00, 0x00,
}
