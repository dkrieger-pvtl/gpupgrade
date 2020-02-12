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

type UpgradePrimariesRequest struct {
	SourceBinDir         string         `protobuf:"bytes,1,opt,name=SourceBinDir" json:"SourceBinDir,omitempty"`
	TargetBinDir         string         `protobuf:"bytes,2,opt,name=TargetBinDir" json:"TargetBinDir,omitempty"`
	TargetVersion        string         `protobuf:"bytes,3,opt,name=TargetVersion" json:"TargetVersion,omitempty"`
	DataDirPairs         []*DataDirPair `protobuf:"bytes,4,rep,name=DataDirPairs" json:"DataDirPairs,omitempty"`
	CheckOnly            bool           `protobuf:"varint,5,opt,name=CheckOnly" json:"CheckOnly,omitempty"`
	UseLinkMode          bool           `protobuf:"varint,6,opt,name=UseLinkMode" json:"UseLinkMode,omitempty"`
	MasterBackupDir      string         `protobuf:"bytes,7,opt,name=MasterBackupDir" json:"MasterBackupDir,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *UpgradePrimariesRequest) Reset()         { *m = UpgradePrimariesRequest{} }
func (m *UpgradePrimariesRequest) String() string { return proto.CompactTextString(m) }
func (*UpgradePrimariesRequest) ProtoMessage()    {}
func (*UpgradePrimariesRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_hub_to_agent_ce7efdff2736ce20, []int{0}
}
func (m *UpgradePrimariesRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UpgradePrimariesRequest.Unmarshal(m, b)
}
func (m *UpgradePrimariesRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UpgradePrimariesRequest.Marshal(b, m, deterministic)
}
func (dst *UpgradePrimariesRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UpgradePrimariesRequest.Merge(dst, src)
}
func (m *UpgradePrimariesRequest) XXX_Size() int {
	return xxx_messageInfo_UpgradePrimariesRequest.Size(m)
}
func (m *UpgradePrimariesRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_UpgradePrimariesRequest.DiscardUnknown(m)
}

var xxx_messageInfo_UpgradePrimariesRequest proto.InternalMessageInfo

func (m *UpgradePrimariesRequest) GetSourceBinDir() string {
	if m != nil {
		return m.SourceBinDir
	}
	return ""
}

func (m *UpgradePrimariesRequest) GetTargetBinDir() string {
	if m != nil {
		return m.TargetBinDir
	}
	return ""
}

func (m *UpgradePrimariesRequest) GetTargetVersion() string {
	if m != nil {
		return m.TargetVersion
	}
	return ""
}

func (m *UpgradePrimariesRequest) GetDataDirPairs() []*DataDirPair {
	if m != nil {
		return m.DataDirPairs
	}
	return nil
}

func (m *UpgradePrimariesRequest) GetCheckOnly() bool {
	if m != nil {
		return m.CheckOnly
	}
	return false
}

func (m *UpgradePrimariesRequest) GetUseLinkMode() bool {
	if m != nil {
		return m.UseLinkMode
	}
	return false
}

func (m *UpgradePrimariesRequest) GetMasterBackupDir() string {
	if m != nil {
		return m.MasterBackupDir
	}
	return ""
}

type DataDirPair struct {
	SourceDataDir        string   `protobuf:"bytes,1,opt,name=SourceDataDir" json:"SourceDataDir,omitempty"`
	TargetDataDir        string   `protobuf:"bytes,2,opt,name=TargetDataDir" json:"TargetDataDir,omitempty"`
	SourcePort           int32    `protobuf:"varint,3,opt,name=SourcePort" json:"SourcePort,omitempty"`
	TargetPort           int32    `protobuf:"varint,4,opt,name=TargetPort" json:"TargetPort,omitempty"`
	Content              int32    `protobuf:"varint,5,opt,name=Content" json:"Content,omitempty"`
	DBID                 int32    `protobuf:"varint,6,opt,name=DBID" json:"DBID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DataDirPair) Reset()         { *m = DataDirPair{} }
func (m *DataDirPair) String() string { return proto.CompactTextString(m) }
func (*DataDirPair) ProtoMessage()    {}
func (*DataDirPair) Descriptor() ([]byte, []int) {
	return fileDescriptor_hub_to_agent_ce7efdff2736ce20, []int{1}
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

func (m *DataDirPair) GetSourceDataDir() string {
	if m != nil {
		return m.SourceDataDir
	}
	return ""
}

func (m *DataDirPair) GetTargetDataDir() string {
	if m != nil {
		return m.TargetDataDir
	}
	return ""
}

func (m *DataDirPair) GetSourcePort() int32 {
	if m != nil {
		return m.SourcePort
	}
	return 0
}

func (m *DataDirPair) GetTargetPort() int32 {
	if m != nil {
		return m.TargetPort
	}
	return 0
}

func (m *DataDirPair) GetContent() int32 {
	if m != nil {
		return m.Content
	}
	return 0
}

func (m *DataDirPair) GetDBID() int32 {
	if m != nil {
		return m.DBID
	}
	return 0
}

type UpgradePrimariesReply struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *UpgradePrimariesReply) Reset()         { *m = UpgradePrimariesReply{} }
func (m *UpgradePrimariesReply) String() string { return proto.CompactTextString(m) }
func (*UpgradePrimariesReply) ProtoMessage()    {}
func (*UpgradePrimariesReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_hub_to_agent_ce7efdff2736ce20, []int{2}
}
func (m *UpgradePrimariesReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UpgradePrimariesReply.Unmarshal(m, b)
}
func (m *UpgradePrimariesReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UpgradePrimariesReply.Marshal(b, m, deterministic)
}
func (dst *UpgradePrimariesReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UpgradePrimariesReply.Merge(dst, src)
}
func (m *UpgradePrimariesReply) XXX_Size() int {
	return xxx_messageInfo_UpgradePrimariesReply.Size(m)
}
func (m *UpgradePrimariesReply) XXX_DiscardUnknown() {
	xxx_messageInfo_UpgradePrimariesReply.DiscardUnknown(m)
}

var xxx_messageInfo_UpgradePrimariesReply proto.InternalMessageInfo

type CreateSegmentDataDirRequest struct {
	Datadirs             []string `protobuf:"bytes,1,rep,name=datadirs" json:"datadirs,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CreateSegmentDataDirRequest) Reset()         { *m = CreateSegmentDataDirRequest{} }
func (m *CreateSegmentDataDirRequest) String() string { return proto.CompactTextString(m) }
func (*CreateSegmentDataDirRequest) ProtoMessage()    {}
func (*CreateSegmentDataDirRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_hub_to_agent_ce7efdff2736ce20, []int{3}
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
	return fileDescriptor_hub_to_agent_ce7efdff2736ce20, []int{4}
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

type StopAgentRequest struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *StopAgentRequest) Reset()         { *m = StopAgentRequest{} }
func (m *StopAgentRequest) String() string { return proto.CompactTextString(m) }
func (*StopAgentRequest) ProtoMessage()    {}
func (*StopAgentRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_hub_to_agent_ce7efdff2736ce20, []int{5}
}
func (m *StopAgentRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StopAgentRequest.Unmarshal(m, b)
}
func (m *StopAgentRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StopAgentRequest.Marshal(b, m, deterministic)
}
func (dst *StopAgentRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StopAgentRequest.Merge(dst, src)
}
func (m *StopAgentRequest) XXX_Size() int {
	return xxx_messageInfo_StopAgentRequest.Size(m)
}
func (m *StopAgentRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_StopAgentRequest.DiscardUnknown(m)
}

var xxx_messageInfo_StopAgentRequest proto.InternalMessageInfo

type StopAgentReply struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *StopAgentReply) Reset()         { *m = StopAgentReply{} }
func (m *StopAgentReply) String() string { return proto.CompactTextString(m) }
func (*StopAgentReply) ProtoMessage()    {}
func (*StopAgentReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_hub_to_agent_ce7efdff2736ce20, []int{6}
}
func (m *StopAgentReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StopAgentReply.Unmarshal(m, b)
}
func (m *StopAgentReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StopAgentReply.Marshal(b, m, deterministic)
}
func (dst *StopAgentReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StopAgentReply.Merge(dst, src)
}
func (m *StopAgentReply) XXX_Size() int {
	return xxx_messageInfo_StopAgentReply.Size(m)
}
func (m *StopAgentReply) XXX_DiscardUnknown() {
	xxx_messageInfo_StopAgentReply.DiscardUnknown(m)
}

var xxx_messageInfo_StopAgentReply proto.InternalMessageInfo

type CheckSegmentDiskSpaceRequest struct {
	Request              *CheckDiskSpaceRequest `protobuf:"bytes,1,opt,name=request" json:"request,omitempty"`
	Datadirs             []string               `protobuf:"bytes,2,rep,name=datadirs" json:"datadirs,omitempty"`
	XXX_NoUnkeyedLiteral struct{}               `json:"-"`
	XXX_unrecognized     []byte                 `json:"-"`
	XXX_sizecache        int32                  `json:"-"`
}

func (m *CheckSegmentDiskSpaceRequest) Reset()         { *m = CheckSegmentDiskSpaceRequest{} }
func (m *CheckSegmentDiskSpaceRequest) String() string { return proto.CompactTextString(m) }
func (*CheckSegmentDiskSpaceRequest) ProtoMessage()    {}
func (*CheckSegmentDiskSpaceRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_hub_to_agent_ce7efdff2736ce20, []int{7}
}
func (m *CheckSegmentDiskSpaceRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CheckSegmentDiskSpaceRequest.Unmarshal(m, b)
}
func (m *CheckSegmentDiskSpaceRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CheckSegmentDiskSpaceRequest.Marshal(b, m, deterministic)
}
func (dst *CheckSegmentDiskSpaceRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CheckSegmentDiskSpaceRequest.Merge(dst, src)
}
func (m *CheckSegmentDiskSpaceRequest) XXX_Size() int {
	return xxx_messageInfo_CheckSegmentDiskSpaceRequest.Size(m)
}
func (m *CheckSegmentDiskSpaceRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_CheckSegmentDiskSpaceRequest.DiscardUnknown(m)
}

var xxx_messageInfo_CheckSegmentDiskSpaceRequest proto.InternalMessageInfo

func (m *CheckSegmentDiskSpaceRequest) GetRequest() *CheckDiskSpaceRequest {
	if m != nil {
		return m.Request
	}
	return nil
}

func (m *CheckSegmentDiskSpaceRequest) GetDatadirs() []string {
	if m != nil {
		return m.Datadirs
	}
	return nil
}

func init() {
	proto.RegisterType((*UpgradePrimariesRequest)(nil), "idl.UpgradePrimariesRequest")
	proto.RegisterType((*DataDirPair)(nil), "idl.DataDirPair")
	proto.RegisterType((*UpgradePrimariesReply)(nil), "idl.UpgradePrimariesReply")
	proto.RegisterType((*CreateSegmentDataDirRequest)(nil), "idl.CreateSegmentDataDirRequest")
	proto.RegisterType((*CreateSegmentDataDirReply)(nil), "idl.CreateSegmentDataDirReply")
	proto.RegisterType((*StopAgentRequest)(nil), "idl.StopAgentRequest")
	proto.RegisterType((*StopAgentReply)(nil), "idl.StopAgentReply")
	proto.RegisterType((*CheckSegmentDiskSpaceRequest)(nil), "idl.CheckSegmentDiskSpaceRequest")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Agent service

type AgentClient interface {
	CheckDiskSpace(ctx context.Context, in *CheckSegmentDiskSpaceRequest, opts ...grpc.CallOption) (*CheckDiskSpaceReply, error)
	UpgradePrimaries(ctx context.Context, in *UpgradePrimariesRequest, opts ...grpc.CallOption) (*UpgradePrimariesReply, error)
	CreateSegmentDataDirectories(ctx context.Context, in *CreateSegmentDataDirRequest, opts ...grpc.CallOption) (*CreateSegmentDataDirReply, error)
	StopAgent(ctx context.Context, in *StopAgentRequest, opts ...grpc.CallOption) (*StopAgentReply, error)
}

type agentClient struct {
	cc *grpc.ClientConn
}

func NewAgentClient(cc *grpc.ClientConn) AgentClient {
	return &agentClient{cc}
}

func (c *agentClient) CheckDiskSpace(ctx context.Context, in *CheckSegmentDiskSpaceRequest, opts ...grpc.CallOption) (*CheckDiskSpaceReply, error) {
	out := new(CheckDiskSpaceReply)
	err := grpc.Invoke(ctx, "/idl.Agent/CheckDiskSpace", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentClient) UpgradePrimaries(ctx context.Context, in *UpgradePrimariesRequest, opts ...grpc.CallOption) (*UpgradePrimariesReply, error) {
	out := new(UpgradePrimariesReply)
	err := grpc.Invoke(ctx, "/idl.Agent/UpgradePrimaries", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentClient) CreateSegmentDataDirectories(ctx context.Context, in *CreateSegmentDataDirRequest, opts ...grpc.CallOption) (*CreateSegmentDataDirReply, error) {
	out := new(CreateSegmentDataDirReply)
	err := grpc.Invoke(ctx, "/idl.Agent/CreateSegmentDataDirectories", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentClient) StopAgent(ctx context.Context, in *StopAgentRequest, opts ...grpc.CallOption) (*StopAgentReply, error) {
	out := new(StopAgentReply)
	err := grpc.Invoke(ctx, "/idl.Agent/StopAgent", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Agent service

type AgentServer interface {
	CheckDiskSpace(context.Context, *CheckSegmentDiskSpaceRequest) (*CheckDiskSpaceReply, error)
	UpgradePrimaries(context.Context, *UpgradePrimariesRequest) (*UpgradePrimariesReply, error)
	CreateSegmentDataDirectories(context.Context, *CreateSegmentDataDirRequest) (*CreateSegmentDataDirReply, error)
	StopAgent(context.Context, *StopAgentRequest) (*StopAgentReply, error)
}

func RegisterAgentServer(s *grpc.Server, srv AgentServer) {
	s.RegisterService(&_Agent_serviceDesc, srv)
}

func _Agent_CheckDiskSpace_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CheckSegmentDiskSpaceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServer).CheckDiskSpace(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/idl.Agent/CheckDiskSpace",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServer).CheckDiskSpace(ctx, req.(*CheckSegmentDiskSpaceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Agent_UpgradePrimaries_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpgradePrimariesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServer).UpgradePrimaries(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/idl.Agent/UpgradePrimaries",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServer).UpgradePrimaries(ctx, req.(*UpgradePrimariesRequest))
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

func _Agent_StopAgent_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StopAgentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServer).StopAgent(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/idl.Agent/StopAgent",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServer).StopAgent(ctx, req.(*StopAgentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Agent_serviceDesc = grpc.ServiceDesc{
	ServiceName: "idl.Agent",
	HandlerType: (*AgentServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CheckDiskSpace",
			Handler:    _Agent_CheckDiskSpace_Handler,
		},
		{
			MethodName: "UpgradePrimaries",
			Handler:    _Agent_UpgradePrimaries_Handler,
		},
		{
			MethodName: "CreateSegmentDataDirectories",
			Handler:    _Agent_CreateSegmentDataDirectories_Handler,
		},
		{
			MethodName: "StopAgent",
			Handler:    _Agent_StopAgent_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "hub_to_agent.proto",
}

func init() { proto.RegisterFile("hub_to_agent.proto", fileDescriptor_hub_to_agent_ce7efdff2736ce20) }

var fileDescriptor_hub_to_agent_ce7efdff2736ce20 = []byte{
	// 503 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x54, 0x4d, 0x6f, 0x9b, 0x40,
	0x10, 0xad, 0x3f, 0x88, 0xe3, 0x71, 0x9a, 0xa2, 0xad, 0xa2, 0x50, 0x62, 0x45, 0x14, 0xf5, 0xc0,
	0xc9, 0x07, 0x37, 0x97, 0x1c, 0x6b, 0x73, 0xa9, 0xd4, 0x34, 0x16, 0x6e, 0x7a, 0x8d, 0xd6, 0x30,
	0xb2, 0x57, 0x26, 0x2c, 0x5d, 0x96, 0x43, 0x7e, 0x51, 0x7f, 0x4a, 0xfe, 0x56, 0xb5, 0xbb, 0x26,
	0x06, 0x6a, 0xe7, 0xb6, 0xfb, 0xe6, 0xcd, 0xcc, 0x9b, 0x37, 0x0b, 0x40, 0x36, 0xe5, 0xea, 0x51,
	0xf2, 0x47, 0xba, 0xc6, 0x4c, 0x4e, 0x72, 0xc1, 0x25, 0x27, 0x3d, 0x96, 0xa4, 0xae, 0x1d, 0xa7,
	0x4c, 0x05, 0x36, 0xe5, 0xca, 0xc0, 0xfe, 0xdf, 0x2e, 0x5c, 0x3e, 0xe4, 0x6b, 0x41, 0x13, 0x5c,
	0x08, 0xf6, 0x44, 0x05, 0xc3, 0x22, 0xc2, 0x3f, 0x25, 0x16, 0x92, 0xf8, 0x70, 0xb6, 0xe4, 0xa5,
	0x88, 0x71, 0xc6, 0xb2, 0x90, 0x09, 0xa7, 0xe3, 0x75, 0x82, 0x61, 0xd4, 0xc0, 0x14, 0xe7, 0x17,
	0x15, 0x6b, 0x94, 0x3b, 0x4e, 0xd7, 0x70, 0xea, 0x18, 0xf9, 0x02, 0xef, 0xcd, 0xfd, 0x37, 0x8a,
	0x82, 0xf1, 0xcc, 0xe9, 0x69, 0x52, 0x13, 0x24, 0x37, 0x70, 0x16, 0x52, 0x49, 0x43, 0x26, 0x16,
	0x94, 0x89, 0xc2, 0xe9, 0x7b, 0xbd, 0x60, 0x34, 0xb5, 0x27, 0x2c, 0x49, 0x27, 0xb5, 0x40, 0xd4,
	0x60, 0x91, 0x31, 0x0c, 0xe7, 0x1b, 0x8c, 0xb7, 0xf7, 0x59, 0xfa, 0xec, 0x58, 0x5e, 0x27, 0x38,
	0x8d, 0xf6, 0x00, 0xf1, 0x60, 0xf4, 0x50, 0xe0, 0x0f, 0x96, 0x6d, 0xef, 0x78, 0x82, 0xce, 0x89,
	0x8e, 0xd7, 0x21, 0x12, 0xc0, 0x87, 0x3b, 0x5a, 0x48, 0x14, 0x33, 0x1a, 0x6f, 0xcb, 0x5c, 0x8d,
	0x30, 0xd0, 0xea, 0xda, 0xb0, 0xff, 0xd2, 0x81, 0x51, 0xad, 0xb5, 0x9a, 0xca, 0x38, 0xb1, 0x03,
	0x77, 0xf6, 0x34, 0xc1, 0xfd, 0xec, 0x15, 0xab, 0x5b, 0x9f, 0xbd, 0x62, 0x5d, 0x03, 0x98, 0xb4,
	0x05, 0x17, 0x52, 0xdb, 0x63, 0x45, 0x35, 0x44, 0xc5, 0x4d, 0x82, 0x8e, 0xf7, 0x4d, 0x7c, 0x8f,
	0x10, 0x07, 0x06, 0x73, 0x9e, 0x49, 0xcc, 0xa4, 0xf6, 0xc0, 0x8a, 0xaa, 0x2b, 0x21, 0xd0, 0x0f,
	0x67, 0xdf, 0x43, 0x3d, 0xba, 0x15, 0xe9, 0xb3, 0x7f, 0x09, 0x17, 0xff, 0xaf, 0x3c, 0x4f, 0x9f,
	0xfd, 0x5b, 0xb8, 0x9a, 0x0b, 0xa4, 0x12, 0x97, 0xb8, 0x7e, 0xc2, 0xac, 0x92, 0x57, 0xbd, 0x07,
	0x17, 0x4e, 0x13, 0x2a, 0x69, 0xa2, 0xb6, 0xd3, 0xf1, 0x7a, 0xc1, 0x30, 0x7a, 0xbd, 0xfb, 0x57,
	0xf0, 0xe9, 0x70, 0xaa, 0xaa, 0x4b, 0xc0, 0x5e, 0x4a, 0x9e, 0x7f, 0x53, 0xcf, 0x71, 0x57, 0xcc,
	0xb7, 0xe1, 0xbc, 0x86, 0x29, 0x56, 0x0e, 0x63, 0xbd, 0xb9, 0xaa, 0x02, 0x2b, 0xb6, 0xcb, 0x9c,
	0xc6, 0x58, 0xb5, 0xbf, 0x81, 0x81, 0x30, 0x47, 0x6d, 0xf5, 0x68, 0xea, 0xea, 0xb7, 0xa1, 0x73,
	0xda, 0xe4, 0xa8, 0xa2, 0x36, 0x44, 0x77, 0x9b, 0xa2, 0xa7, 0x2f, 0x5d, 0xb0, 0xb4, 0x00, 0x72,
	0x0f, 0xe7, 0xcd, 0x3a, 0xe4, 0xf3, 0xbe, 0xf8, 0x11, 0x41, 0xae, 0x73, 0xb0, 0xbf, 0x1a, 0xe5,
	0x1d, 0xf9, 0x09, 0x76, 0xdb, 0x63, 0x32, 0xd6, 0xfc, 0x23, 0x5f, 0x9b, 0xeb, 0x1e, 0x89, 0x9a,
	0x7a, 0x2b, 0x18, 0x1f, 0xf2, 0x17, 0x63, 0xc9, 0x75, 0x6d, 0xcf, 0x68, 0x39, 0xbe, 0x3d, 0xf7,
	0xfa, 0x0d, 0x86, 0xe9, 0x71, 0x0b, 0xc3, 0xd7, 0x95, 0x90, 0x0b, 0x4d, 0x6f, 0xaf, 0xcd, 0xfd,
	0xd8, 0x86, 0x75, 0xea, 0xea, 0x44, 0xff, 0x4d, 0xbe, 0xfe, 0x0b, 0x00, 0x00, 0xff, 0xff, 0xcb,
	0xe9, 0x3a, 0xd6, 0x7a, 0x04, 0x00, 0x00,
}
