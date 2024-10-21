// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.28.2
// source: zab.proto

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
	ZabService_Ping_FullMethodName = "/zab.ZabService/Ping"
)

// ZabServiceClient is the client API for ZabService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// The zab service definition.
type ZabServiceClient interface {
	// Ping RPC to check if the node is alive.
	Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type zabServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewZabServiceClient(cc grpc.ClientConnInterface) ZabServiceClient {
	return &zabServiceClient{cc}
}

func (c *zabServiceClient) Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, ZabService_Ping_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ZabServiceServer is the server API for ZabService service.
// All implementations must embed UnimplementedZabServiceServer
// for forward compatibility.
//
// The zab service definition.
type ZabServiceServer interface {
	// Ping RPC to check if the node is alive.
	Ping(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
	mustEmbedUnimplementedZabServiceServer()
}

// UnimplementedZabServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedZabServiceServer struct{}

func (UnimplementedZabServiceServer) Ping(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedZabServiceServer) mustEmbedUnimplementedZabServiceServer() {}
func (UnimplementedZabServiceServer) testEmbeddedByValue()                    {}

// UnsafeZabServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ZabServiceServer will
// result in compilation errors.
type UnsafeZabServiceServer interface {
	mustEmbedUnimplementedZabServiceServer()
}

func RegisterZabServiceServer(s grpc.ServiceRegistrar, srv ZabServiceServer) {
	// If the following call pancis, it indicates UnimplementedZabServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&ZabService_ServiceDesc, srv)
}

func _ZabService_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ZabServiceServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ZabService_Ping_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ZabServiceServer).Ping(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// ZabService_ServiceDesc is the grpc.ServiceDesc for ZabService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ZabService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "zab.ZabService",
	HandlerType: (*ZabServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _ZabService_Ping_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "zab.proto",
}
