// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.27.3
// source: lcr.proto

package rpc

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	LCRService_Message_FullMethodName           = "/lcr.LCRService/Message"
	LCRService_NotifyTermination_FullMethodName = "/lcr.LCRService/NotifyTermination"
	LCRService_Ping_FullMethodName              = "/lcr.LCRService/Ping"
)

// LCRServiceClient is the client API for LCRService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// The LCR service definition.
type LCRServiceClient interface {
	// RPC for sending a message to the next process (to the left).
	Message(ctx context.Context, in *LCRMessage, opts ...grpc.CallOption) (*LCRResponse, error)
	// RPC to notify the current process that the leader has been elected and termination is starting.
	NotifyTermination(ctx context.Context, in *LCRMessage, opts ...grpc.CallOption) (*LCRResponse, error)
	// Ping RPC to check if the node is alive.
	Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type lCRServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewLCRServiceClient(cc grpc.ClientConnInterface) LCRServiceClient {
	return &lCRServiceClient{cc}
}

func (c *lCRServiceClient) Message(ctx context.Context, in *LCRMessage, opts ...grpc.CallOption) (*LCRResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(LCRResponse)
	err := c.cc.Invoke(ctx, LCRService_Message_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *lCRServiceClient) NotifyTermination(ctx context.Context, in *LCRMessage, opts ...grpc.CallOption) (*LCRResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(LCRResponse)
	err := c.cc.Invoke(ctx, LCRService_NotifyTermination_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *lCRServiceClient) Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, LCRService_Ping_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LCRServiceServer is the server API for LCRService service.
// All implementations must embed UnimplementedLCRServiceServer
// for forward compatibility.
//
// The LCR service definition.
type LCRServiceServer interface {
	// RPC for sending a message to the next process (to the left).
	Message(context.Context, *LCRMessage) (*LCRResponse, error)
	// RPC to notify the current process that the leader has been elected and termination is starting.
	NotifyTermination(context.Context, *LCRMessage) (*LCRResponse, error)
	// Ping RPC to check if the node is alive.
	Ping(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
	mustEmbedUnimplementedLCRServiceServer()
}

// UnimplementedLCRServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedLCRServiceServer struct{}

func (UnimplementedLCRServiceServer) Message(context.Context, *LCRMessage) (*LCRResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Message not implemented")
}
func (UnimplementedLCRServiceServer) NotifyTermination(context.Context, *LCRMessage) (*LCRResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NotifyTermination not implemented")
}
func (UnimplementedLCRServiceServer) Ping(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedLCRServiceServer) mustEmbedUnimplementedLCRServiceServer() {}
func (UnimplementedLCRServiceServer) testEmbeddedByValue()                    {}

// UnsafeLCRServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to LCRServiceServer will
// result in compilation errors.
type UnsafeLCRServiceServer interface {
	mustEmbedUnimplementedLCRServiceServer()
}

func RegisterLCRServiceServer(s grpc.ServiceRegistrar, srv LCRServiceServer) {
	// If the following call pancis, it indicates UnimplementedLCRServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&LCRService_ServiceDesc, srv)
}

func _LCRService_Message_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LCRMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LCRServiceServer).Message(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LCRService_Message_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LCRServiceServer).Message(ctx, req.(*LCRMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _LCRService_NotifyTermination_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LCRMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LCRServiceServer).NotifyTermination(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LCRService_NotifyTermination_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LCRServiceServer).NotifyTermination(ctx, req.(*LCRMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _LCRService_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LCRServiceServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LCRService_Ping_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LCRServiceServer).Ping(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// LCRService_ServiceDesc is the grpc.ServiceDesc for LCRService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var LCRService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "lcr.LCRService",
	HandlerType: (*LCRServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Message",
			Handler:    _LCRService_Message_Handler,
		},
		{
			MethodName: "NotifyTermination",
			Handler:    _LCRService_NotifyTermination_Handler,
		},
		{
			MethodName: "Ping",
			Handler:    _LCRService_Ping_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "lcr.proto",
}
