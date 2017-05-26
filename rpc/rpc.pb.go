// Code generated by protoc-gen-go.
// source: rpc.proto
// DO NOT EDIT!

/*
Package rpc is a generated protocol buffer package.

It is generated from these files:
	rpc.proto

It has these top-level messages:
	BlockRPCData
	LocateRequest
	LocateResponse
	NegotiateRequest
	NegotiateResponse
*/
package rpc

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import structs "github.com/hexablock/blockring/structs"
import chord "github.com/ipkg/go-chord"

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

type BlockRPCData struct {
	ID       []byte                  `protobuf:"bytes,1,opt,name=ID,json=iD,proto3" json:"ID,omitempty"`
	Block    *structs.Block          `protobuf:"bytes,2,opt,name=Block,json=block" json:"Block,omitempty"`
	Location *structs.Location       `protobuf:"bytes,3,opt,name=Location,json=location" json:"Location,omitempty"`
	Options  *structs.RequestOptions `protobuf:"bytes,4,opt,name=Options,json=options" json:"Options,omitempty"`
}

func (m *BlockRPCData) Reset()                    { *m = BlockRPCData{} }
func (m *BlockRPCData) String() string            { return proto.CompactTextString(m) }
func (*BlockRPCData) ProtoMessage()               {}
func (*BlockRPCData) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *BlockRPCData) GetID() []byte {
	if m != nil {
		return m.ID
	}
	return nil
}

func (m *BlockRPCData) GetBlock() *structs.Block {
	if m != nil {
		return m.Block
	}
	return nil
}

func (m *BlockRPCData) GetLocation() *structs.Location {
	if m != nil {
		return m.Location
	}
	return nil
}

func (m *BlockRPCData) GetOptions() *structs.RequestOptions {
	if m != nil {
		return m.Options
	}
	return nil
}

type LocateRequest struct {
	Key []byte `protobuf:"bytes,1,opt,name=Key,json=key,proto3" json:"Key,omitempty"`
	N   int32  `protobuf:"varint,2,opt,name=N,json=n" json:"N,omitempty"`
}

func (m *LocateRequest) Reset()                    { *m = LocateRequest{} }
func (m *LocateRequest) String() string            { return proto.CompactTextString(m) }
func (*LocateRequest) ProtoMessage()               {}
func (*LocateRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *LocateRequest) GetKey() []byte {
	if m != nil {
		return m.Key
	}
	return nil
}

func (m *LocateRequest) GetN() int32 {
	if m != nil {
		return m.N
	}
	return 0
}

type LocateResponse struct {
	KeyHash     []byte              `protobuf:"bytes,1,opt,name=KeyHash,json=keyHash,proto3" json:"KeyHash,omitempty"`
	Predecessor *chord.Vnode        `protobuf:"bytes,2,opt,name=Predecessor,json=predecessor" json:"Predecessor,omitempty"`
	Successors  []*chord.Vnode      `protobuf:"bytes,3,rep,name=Successors,json=successors" json:"Successors,omitempty"`
	Locations   []*structs.Location `protobuf:"bytes,4,rep,name=Locations,json=locations" json:"Locations,omitempty"`
}

func (m *LocateResponse) Reset()                    { *m = LocateResponse{} }
func (m *LocateResponse) String() string            { return proto.CompactTextString(m) }
func (*LocateResponse) ProtoMessage()               {}
func (*LocateResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *LocateResponse) GetKeyHash() []byte {
	if m != nil {
		return m.KeyHash
	}
	return nil
}

func (m *LocateResponse) GetPredecessor() *chord.Vnode {
	if m != nil {
		return m.Predecessor
	}
	return nil
}

func (m *LocateResponse) GetSuccessors() []*chord.Vnode {
	if m != nil {
		return m.Successors
	}
	return nil
}

func (m *LocateResponse) GetLocations() []*structs.Location {
	if m != nil {
		return m.Locations
	}
	return nil
}

type NegotiateRequest struct {
	Key []byte `protobuf:"bytes,1,opt,name=Key,json=key,proto3" json:"Key,omitempty"`
}

func (m *NegotiateRequest) Reset()                    { *m = NegotiateRequest{} }
func (m *NegotiateRequest) String() string            { return proto.CompactTextString(m) }
func (*NegotiateRequest) ProtoMessage()               {}
func (*NegotiateRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *NegotiateRequest) GetKey() []byte {
	if m != nil {
		return m.Key
	}
	return nil
}

type NegotiateResponse struct {
	Successors int32    `protobuf:"varint,1,opt,name=Successors,json=successors" json:"Successors,omitempty"`
	Vnodes     int32    `protobuf:"varint,2,opt,name=Vnodes,json=vnodes" json:"Vnodes,omitempty"`
	Peers      []string `protobuf:"bytes,3,rep,name=Peers,json=peers" json:"Peers,omitempty"`
}

func (m *NegotiateResponse) Reset()                    { *m = NegotiateResponse{} }
func (m *NegotiateResponse) String() string            { return proto.CompactTextString(m) }
func (*NegotiateResponse) ProtoMessage()               {}
func (*NegotiateResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *NegotiateResponse) GetSuccessors() int32 {
	if m != nil {
		return m.Successors
	}
	return 0
}

func (m *NegotiateResponse) GetVnodes() int32 {
	if m != nil {
		return m.Vnodes
	}
	return 0
}

func (m *NegotiateResponse) GetPeers() []string {
	if m != nil {
		return m.Peers
	}
	return nil
}

func init() {
	proto.RegisterType((*BlockRPCData)(nil), "rpc.BlockRPCData")
	proto.RegisterType((*LocateRequest)(nil), "rpc.LocateRequest")
	proto.RegisterType((*LocateResponse)(nil), "rpc.LocateResponse")
	proto.RegisterType((*NegotiateRequest)(nil), "rpc.NegotiateRequest")
	proto.RegisterType((*NegotiateResponse)(nil), "rpc.NegotiateResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for BlockRPC service

type BlockRPCClient interface {
	GetBlockRPC(ctx context.Context, in *BlockRPCData, opts ...grpc.CallOption) (*BlockRPCData, error)
	SetBlockRPC(ctx context.Context, in *BlockRPCData, opts ...grpc.CallOption) (*BlockRPCData, error)
	TransferBlockRPC(ctx context.Context, in *BlockRPCData, opts ...grpc.CallOption) (*BlockRPCData, error)
	ReleaseBlockRPC(ctx context.Context, in *BlockRPCData, opts ...grpc.CallOption) (*BlockRPCData, error)
}

type blockRPCClient struct {
	cc *grpc.ClientConn
}

func NewBlockRPCClient(cc *grpc.ClientConn) BlockRPCClient {
	return &blockRPCClient{cc}
}

func (c *blockRPCClient) GetBlockRPC(ctx context.Context, in *BlockRPCData, opts ...grpc.CallOption) (*BlockRPCData, error) {
	out := new(BlockRPCData)
	err := grpc.Invoke(ctx, "/rpc.BlockRPC/GetBlockRPC", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blockRPCClient) SetBlockRPC(ctx context.Context, in *BlockRPCData, opts ...grpc.CallOption) (*BlockRPCData, error) {
	out := new(BlockRPCData)
	err := grpc.Invoke(ctx, "/rpc.BlockRPC/SetBlockRPC", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blockRPCClient) TransferBlockRPC(ctx context.Context, in *BlockRPCData, opts ...grpc.CallOption) (*BlockRPCData, error) {
	out := new(BlockRPCData)
	err := grpc.Invoke(ctx, "/rpc.BlockRPC/TransferBlockRPC", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blockRPCClient) ReleaseBlockRPC(ctx context.Context, in *BlockRPCData, opts ...grpc.CallOption) (*BlockRPCData, error) {
	out := new(BlockRPCData)
	err := grpc.Invoke(ctx, "/rpc.BlockRPC/ReleaseBlockRPC", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for BlockRPC service

type BlockRPCServer interface {
	GetBlockRPC(context.Context, *BlockRPCData) (*BlockRPCData, error)
	SetBlockRPC(context.Context, *BlockRPCData) (*BlockRPCData, error)
	TransferBlockRPC(context.Context, *BlockRPCData) (*BlockRPCData, error)
	ReleaseBlockRPC(context.Context, *BlockRPCData) (*BlockRPCData, error)
}

func RegisterBlockRPCServer(s *grpc.Server, srv BlockRPCServer) {
	s.RegisterService(&_BlockRPC_serviceDesc, srv)
}

func _BlockRPC_GetBlockRPC_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BlockRPCData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlockRPCServer).GetBlockRPC(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.BlockRPC/GetBlockRPC",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlockRPCServer).GetBlockRPC(ctx, req.(*BlockRPCData))
	}
	return interceptor(ctx, in, info, handler)
}

func _BlockRPC_SetBlockRPC_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BlockRPCData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlockRPCServer).SetBlockRPC(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.BlockRPC/SetBlockRPC",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlockRPCServer).SetBlockRPC(ctx, req.(*BlockRPCData))
	}
	return interceptor(ctx, in, info, handler)
}

func _BlockRPC_TransferBlockRPC_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BlockRPCData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlockRPCServer).TransferBlockRPC(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.BlockRPC/TransferBlockRPC",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlockRPCServer).TransferBlockRPC(ctx, req.(*BlockRPCData))
	}
	return interceptor(ctx, in, info, handler)
}

func _BlockRPC_ReleaseBlockRPC_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BlockRPCData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlockRPCServer).ReleaseBlockRPC(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.BlockRPC/ReleaseBlockRPC",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlockRPCServer).ReleaseBlockRPC(ctx, req.(*BlockRPCData))
	}
	return interceptor(ctx, in, info, handler)
}

var _BlockRPC_serviceDesc = grpc.ServiceDesc{
	ServiceName: "rpc.BlockRPC",
	HandlerType: (*BlockRPCServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetBlockRPC",
			Handler:    _BlockRPC_GetBlockRPC_Handler,
		},
		{
			MethodName: "SetBlockRPC",
			Handler:    _BlockRPC_SetBlockRPC_Handler,
		},
		{
			MethodName: "TransferBlockRPC",
			Handler:    _BlockRPC_TransferBlockRPC_Handler,
		},
		{
			MethodName: "ReleaseBlockRPC",
			Handler:    _BlockRPC_ReleaseBlockRPC_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "rpc.proto",
}

// Client API for LocateRPC service

type LocateRPCClient interface {
	LookupKeyRPC(ctx context.Context, in *LocateRequest, opts ...grpc.CallOption) (*LocateResponse, error)
	LookupHashRPC(ctx context.Context, in *LocateRequest, opts ...grpc.CallOption) (*LocateResponse, error)
	LocateReplicatedKeyRPC(ctx context.Context, in *LocateRequest, opts ...grpc.CallOption) (*LocateResponse, error)
	LocateReplicatedHashRPC(ctx context.Context, in *LocateRequest, opts ...grpc.CallOption) (*LocateResponse, error)
	NegotiateRPC(ctx context.Context, in *NegotiateRequest, opts ...grpc.CallOption) (*NegotiateResponse, error)
}

type locateRPCClient struct {
	cc *grpc.ClientConn
}

func NewLocateRPCClient(cc *grpc.ClientConn) LocateRPCClient {
	return &locateRPCClient{cc}
}

func (c *locateRPCClient) LookupKeyRPC(ctx context.Context, in *LocateRequest, opts ...grpc.CallOption) (*LocateResponse, error) {
	out := new(LocateResponse)
	err := grpc.Invoke(ctx, "/rpc.LocateRPC/LookupKeyRPC", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *locateRPCClient) LookupHashRPC(ctx context.Context, in *LocateRequest, opts ...grpc.CallOption) (*LocateResponse, error) {
	out := new(LocateResponse)
	err := grpc.Invoke(ctx, "/rpc.LocateRPC/LookupHashRPC", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *locateRPCClient) LocateReplicatedKeyRPC(ctx context.Context, in *LocateRequest, opts ...grpc.CallOption) (*LocateResponse, error) {
	out := new(LocateResponse)
	err := grpc.Invoke(ctx, "/rpc.LocateRPC/LocateReplicatedKeyRPC", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *locateRPCClient) LocateReplicatedHashRPC(ctx context.Context, in *LocateRequest, opts ...grpc.CallOption) (*LocateResponse, error) {
	out := new(LocateResponse)
	err := grpc.Invoke(ctx, "/rpc.LocateRPC/LocateReplicatedHashRPC", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *locateRPCClient) NegotiateRPC(ctx context.Context, in *NegotiateRequest, opts ...grpc.CallOption) (*NegotiateResponse, error) {
	out := new(NegotiateResponse)
	err := grpc.Invoke(ctx, "/rpc.LocateRPC/NegotiateRPC", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for LocateRPC service

type LocateRPCServer interface {
	LookupKeyRPC(context.Context, *LocateRequest) (*LocateResponse, error)
	LookupHashRPC(context.Context, *LocateRequest) (*LocateResponse, error)
	LocateReplicatedKeyRPC(context.Context, *LocateRequest) (*LocateResponse, error)
	LocateReplicatedHashRPC(context.Context, *LocateRequest) (*LocateResponse, error)
	NegotiateRPC(context.Context, *NegotiateRequest) (*NegotiateResponse, error)
}

func RegisterLocateRPCServer(s *grpc.Server, srv LocateRPCServer) {
	s.RegisterService(&_LocateRPC_serviceDesc, srv)
}

func _LocateRPC_LookupKeyRPC_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LocateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LocateRPCServer).LookupKeyRPC(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.LocateRPC/LookupKeyRPC",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LocateRPCServer).LookupKeyRPC(ctx, req.(*LocateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LocateRPC_LookupHashRPC_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LocateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LocateRPCServer).LookupHashRPC(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.LocateRPC/LookupHashRPC",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LocateRPCServer).LookupHashRPC(ctx, req.(*LocateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LocateRPC_LocateReplicatedKeyRPC_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LocateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LocateRPCServer).LocateReplicatedKeyRPC(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.LocateRPC/LocateReplicatedKeyRPC",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LocateRPCServer).LocateReplicatedKeyRPC(ctx, req.(*LocateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LocateRPC_LocateReplicatedHashRPC_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LocateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LocateRPCServer).LocateReplicatedHashRPC(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.LocateRPC/LocateReplicatedHashRPC",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LocateRPCServer).LocateReplicatedHashRPC(ctx, req.(*LocateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LocateRPC_NegotiateRPC_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NegotiateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LocateRPCServer).NegotiateRPC(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.LocateRPC/NegotiateRPC",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LocateRPCServer).NegotiateRPC(ctx, req.(*NegotiateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _LocateRPC_serviceDesc = grpc.ServiceDesc{
	ServiceName: "rpc.LocateRPC",
	HandlerType: (*LocateRPCServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "LookupKeyRPC",
			Handler:    _LocateRPC_LookupKeyRPC_Handler,
		},
		{
			MethodName: "LookupHashRPC",
			Handler:    _LocateRPC_LookupHashRPC_Handler,
		},
		{
			MethodName: "LocateReplicatedKeyRPC",
			Handler:    _LocateRPC_LocateReplicatedKeyRPC_Handler,
		},
		{
			MethodName: "LocateReplicatedHashRPC",
			Handler:    _LocateRPC_LocateReplicatedHashRPC_Handler,
		},
		{
			MethodName: "NegotiateRPC",
			Handler:    _LocateRPC_NegotiateRPC_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "rpc.proto",
}

// Client API for LogRPC service

type LogRPCClient interface {
	NewEntryRPC(ctx context.Context, in *BlockRPCData, opts ...grpc.CallOption) (*BlockRPCData, error)
	ProposeEntryRPC(ctx context.Context, in *BlockRPCData, opts ...grpc.CallOption) (*BlockRPCData, error)
	CommitEntryRPC(ctx context.Context, in *BlockRPCData, opts ...grpc.CallOption) (*BlockRPCData, error)
	GetLogBlockRPC(ctx context.Context, in *BlockRPCData, opts ...grpc.CallOption) (*BlockRPCData, error)
	TransferLogBlockRPC(ctx context.Context, in *BlockRPCData, opts ...grpc.CallOption) (*BlockRPCData, error)
	ReleaseLogBlockRPC(ctx context.Context, in *BlockRPCData, opts ...grpc.CallOption) (*BlockRPCData, error)
}

type logRPCClient struct {
	cc *grpc.ClientConn
}

func NewLogRPCClient(cc *grpc.ClientConn) LogRPCClient {
	return &logRPCClient{cc}
}

func (c *logRPCClient) NewEntryRPC(ctx context.Context, in *BlockRPCData, opts ...grpc.CallOption) (*BlockRPCData, error) {
	out := new(BlockRPCData)
	err := grpc.Invoke(ctx, "/rpc.LogRPC/NewEntryRPC", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *logRPCClient) ProposeEntryRPC(ctx context.Context, in *BlockRPCData, opts ...grpc.CallOption) (*BlockRPCData, error) {
	out := new(BlockRPCData)
	err := grpc.Invoke(ctx, "/rpc.LogRPC/ProposeEntryRPC", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *logRPCClient) CommitEntryRPC(ctx context.Context, in *BlockRPCData, opts ...grpc.CallOption) (*BlockRPCData, error) {
	out := new(BlockRPCData)
	err := grpc.Invoke(ctx, "/rpc.LogRPC/CommitEntryRPC", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *logRPCClient) GetLogBlockRPC(ctx context.Context, in *BlockRPCData, opts ...grpc.CallOption) (*BlockRPCData, error) {
	out := new(BlockRPCData)
	err := grpc.Invoke(ctx, "/rpc.LogRPC/GetLogBlockRPC", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *logRPCClient) TransferLogBlockRPC(ctx context.Context, in *BlockRPCData, opts ...grpc.CallOption) (*BlockRPCData, error) {
	out := new(BlockRPCData)
	err := grpc.Invoke(ctx, "/rpc.LogRPC/TransferLogBlockRPC", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *logRPCClient) ReleaseLogBlockRPC(ctx context.Context, in *BlockRPCData, opts ...grpc.CallOption) (*BlockRPCData, error) {
	out := new(BlockRPCData)
	err := grpc.Invoke(ctx, "/rpc.LogRPC/ReleaseLogBlockRPC", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for LogRPC service

type LogRPCServer interface {
	NewEntryRPC(context.Context, *BlockRPCData) (*BlockRPCData, error)
	ProposeEntryRPC(context.Context, *BlockRPCData) (*BlockRPCData, error)
	CommitEntryRPC(context.Context, *BlockRPCData) (*BlockRPCData, error)
	GetLogBlockRPC(context.Context, *BlockRPCData) (*BlockRPCData, error)
	TransferLogBlockRPC(context.Context, *BlockRPCData) (*BlockRPCData, error)
	ReleaseLogBlockRPC(context.Context, *BlockRPCData) (*BlockRPCData, error)
}

func RegisterLogRPCServer(s *grpc.Server, srv LogRPCServer) {
	s.RegisterService(&_LogRPC_serviceDesc, srv)
}

func _LogRPC_NewEntryRPC_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BlockRPCData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LogRPCServer).NewEntryRPC(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.LogRPC/NewEntryRPC",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LogRPCServer).NewEntryRPC(ctx, req.(*BlockRPCData))
	}
	return interceptor(ctx, in, info, handler)
}

func _LogRPC_ProposeEntryRPC_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BlockRPCData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LogRPCServer).ProposeEntryRPC(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.LogRPC/ProposeEntryRPC",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LogRPCServer).ProposeEntryRPC(ctx, req.(*BlockRPCData))
	}
	return interceptor(ctx, in, info, handler)
}

func _LogRPC_CommitEntryRPC_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BlockRPCData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LogRPCServer).CommitEntryRPC(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.LogRPC/CommitEntryRPC",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LogRPCServer).CommitEntryRPC(ctx, req.(*BlockRPCData))
	}
	return interceptor(ctx, in, info, handler)
}

func _LogRPC_GetLogBlockRPC_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BlockRPCData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LogRPCServer).GetLogBlockRPC(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.LogRPC/GetLogBlockRPC",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LogRPCServer).GetLogBlockRPC(ctx, req.(*BlockRPCData))
	}
	return interceptor(ctx, in, info, handler)
}

func _LogRPC_TransferLogBlockRPC_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BlockRPCData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LogRPCServer).TransferLogBlockRPC(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.LogRPC/TransferLogBlockRPC",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LogRPCServer).TransferLogBlockRPC(ctx, req.(*BlockRPCData))
	}
	return interceptor(ctx, in, info, handler)
}

func _LogRPC_ReleaseLogBlockRPC_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BlockRPCData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LogRPCServer).ReleaseLogBlockRPC(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.LogRPC/ReleaseLogBlockRPC",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LogRPCServer).ReleaseLogBlockRPC(ctx, req.(*BlockRPCData))
	}
	return interceptor(ctx, in, info, handler)
}

var _LogRPC_serviceDesc = grpc.ServiceDesc{
	ServiceName: "rpc.LogRPC",
	HandlerType: (*LogRPCServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "NewEntryRPC",
			Handler:    _LogRPC_NewEntryRPC_Handler,
		},
		{
			MethodName: "ProposeEntryRPC",
			Handler:    _LogRPC_ProposeEntryRPC_Handler,
		},
		{
			MethodName: "CommitEntryRPC",
			Handler:    _LogRPC_CommitEntryRPC_Handler,
		},
		{
			MethodName: "GetLogBlockRPC",
			Handler:    _LogRPC_GetLogBlockRPC_Handler,
		},
		{
			MethodName: "TransferLogBlockRPC",
			Handler:    _LogRPC_TransferLogBlockRPC_Handler,
		},
		{
			MethodName: "ReleaseLogBlockRPC",
			Handler:    _LogRPC_ReleaseLogBlockRPC_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "rpc.proto",
}

func init() { proto.RegisterFile("rpc.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 572 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x94, 0xd1, 0x8e, 0xd2, 0x40,
	0x14, 0x86, 0x2d, 0x15, 0x58, 0x0e, 0x5d, 0xdc, 0x9d, 0x55, 0xb6, 0xe1, 0xc2, 0x90, 0x66, 0x2f,
	0xf6, 0xc2, 0x6d, 0x23, 0x6a, 0xe2, 0x1a, 0x8d, 0x89, 0x60, 0x56, 0x03, 0x41, 0xd2, 0x35, 0xde,
	0x97, 0x72, 0x2c, 0x0d, 0xd0, 0x19, 0x67, 0x06, 0x95, 0x17, 0xf2, 0x1d, 0x7c, 0x09, 0x5f, 0xc4,
	0x0b, 0x5f, 0xc1, 0x74, 0x3a, 0x65, 0x81, 0x35, 0x26, 0xe5, 0x06, 0xe6, 0xcc, 0xfc, 0xdf, 0xf0,
	0x9f, 0x33, 0x7f, 0x80, 0x1a, 0x67, 0xa1, 0xcb, 0x38, 0x95, 0x94, 0x98, 0x9c, 0x85, 0xad, 0xa7,
	0x51, 0x2c, 0xa7, 0xcb, 0xb1, 0x1b, 0xd2, 0x85, 0x37, 0xc5, 0xef, 0xc1, 0x78, 0x4e, 0xc3, 0x99,
	0xa7, 0x3e, 0x79, 0x9c, 0x44, 0x9e, 0x90, 0x7c, 0x19, 0x4a, 0x91, 0x7f, 0x67, 0x68, 0xcb, 0xd9,
	0xa0, 0x62, 0x36, 0x8b, 0xbc, 0x88, 0x5e, 0x84, 0x53, 0xca, 0x27, 0x5e, 0x82, 0x32, 0xd3, 0x38,
	0x3f, 0x0c, 0xb0, 0xde, 0xa4, 0xf7, 0xf8, 0xa3, 0x6e, 0x2f, 0x90, 0x01, 0x69, 0x40, 0xe9, 0x7d,
	0xcf, 0x36, 0xda, 0xc6, 0xb9, 0xe5, 0x97, 0xe2, 0x1e, 0x39, 0x83, 0xb2, 0x3a, 0xb7, 0x4b, 0x6d,
	0xe3, 0xbc, 0xde, 0x69, 0xb8, 0xf9, 0x6f, 0x64, 0x54, 0x59, 0x99, 0x20, 0x17, 0x70, 0x30, 0xa0,
	0x61, 0x20, 0x63, 0x9a, 0xd8, 0xa6, 0x12, 0x1e, 0xaf, 0x85, 0xf9, 0x81, 0x7f, 0x30, 0xd7, 0x2b,
	0xf2, 0x18, 0xaa, 0x1f, 0x58, 0xba, 0x12, 0xf6, 0x5d, 0xa5, 0x3e, 0x5d, 0xab, 0x7d, 0xfc, 0xb2,
	0x44, 0x21, 0xf5, 0xb1, 0x5f, 0xa5, 0xd9, 0xc2, 0xf1, 0xe0, 0x50, 0x5d, 0x84, 0x5a, 0x40, 0x8e,
	0xc0, 0xec, 0xe3, 0x4a, 0x3b, 0x35, 0x67, 0xb8, 0x22, 0x16, 0x18, 0x43, 0x65, 0xb3, 0xec, 0x1b,
	0x89, 0xf3, 0xd3, 0x80, 0x46, 0x4e, 0x08, 0x46, 0x13, 0x81, 0xc4, 0x86, 0x6a, 0x1f, 0x57, 0xef,
	0x02, 0x31, 0xd5, 0x58, 0x75, 0x96, 0x95, 0xc4, 0x85, 0xfa, 0x88, 0xe3, 0x04, 0x43, 0x14, 0x82,
	0x72, 0xdd, 0xab, 0xe5, 0xaa, 0x69, 0xb9, 0x9f, 0x12, 0x3a, 0x41, 0xbf, 0xce, 0x6e, 0x04, 0xe4,
	0x11, 0xc0, 0xf5, 0x32, 0xcc, 0x0a, 0x61, 0x9b, 0x6d, 0xf3, 0x96, 0x1c, 0xc4, 0xfa, 0x9c, 0x78,
	0x50, 0xcb, 0x87, 0x90, 0x36, 0x6c, 0xfe, 0x7b, 0x3c, 0xb5, 0x7c, 0x3c, 0xc2, 0x39, 0x83, 0xa3,
	0x21, 0x46, 0x54, 0xc6, 0xff, 0xeb, 0xd7, 0x09, 0xe0, 0x78, 0x43, 0xa5, 0x7b, 0x7c, 0xb8, 0xe5,
	0xcc, 0x50, 0xd3, 0xd8, 0xf4, 0xd2, 0x84, 0x8a, 0x32, 0x28, 0xf4, 0xa4, 0x2a, 0x5f, 0x55, 0x45,
	0xee, 0x43, 0x79, 0x84, 0xa8, 0x9b, 0xa9, 0xf9, 0x65, 0x96, 0x16, 0x9d, 0x3f, 0x06, 0x1c, 0xe4,
	0xf1, 0x20, 0xcf, 0xa0, 0x7e, 0x85, 0x72, 0x5d, 0x1e, 0xbb, 0x69, 0x4a, 0x37, 0xc3, 0xd3, 0xba,
	0xbd, 0xe5, 0xdc, 0x49, 0xb1, 0xeb, 0x3d, 0xb0, 0x17, 0x70, 0xf4, 0x91, 0x07, 0x89, 0xf8, 0x8c,
	0xbc, 0x30, 0x7b, 0x09, 0xf7, 0x7c, 0x9c, 0x63, 0x20, 0xb0, 0x28, 0xda, 0xf9, 0x55, 0xd2, 0x8f,
	0x85, 0x29, 0x75, 0x09, 0xd6, 0x80, 0xd2, 0xd9, 0x92, 0xf5, 0x71, 0x95, 0xd6, 0x44, 0x21, 0x5b,
	0x41, 0x6c, 0x9d, 0x6c, 0xed, 0x65, 0xcf, 0xa0, 0xfc, 0x1f, 0x66, 0x68, 0x1a, 0xb0, 0x82, 0x6c,
	0x17, 0x9a, 0xf9, 0x1e, 0x9b, 0xc7, 0xe9, 0x62, 0x52, 0xdc, 0x40, 0x0f, 0x4e, 0x77, 0x2f, 0xd9,
	0xc3, 0xca, 0x6b, 0xb0, 0x6e, 0x42, 0x36, 0xea, 0x92, 0x07, 0x4a, 0xb6, 0x9b, 0xce, 0x56, 0x73,
	0x77, 0x3b, 0xbf, 0xa0, 0xf3, 0xbb, 0x04, 0x95, 0x01, 0x8d, 0x74, 0x80, 0x86, 0xf8, 0xed, 0x6d,
	0x22, 0xf9, 0xaa, 0xe0, 0x6b, 0x8e, 0x38, 0x65, 0x54, 0x60, 0x61, 0xf4, 0x39, 0x34, 0xba, 0x74,
	0xb1, 0x88, 0xe5, 0x3e, 0xe4, 0x15, 0xca, 0x01, 0x8d, 0x0a, 0x87, 0xef, 0x15, 0x9c, 0xe4, 0xc1,
	0xdd, 0x07, 0x7f, 0x09, 0x44, 0x67, 0x77, 0x0f, 0x7a, 0x5c, 0x51, 0x7f, 0xeb, 0x4f, 0xfe, 0x06,
	0x00, 0x00, 0xff, 0xff, 0x6a, 0x90, 0x33, 0x80, 0x42, 0x06, 0x00, 0x00,
}
