package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	pb "leadelection/pkg/hs/internal/rpc"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var ErrCallbackNotSet = errors.New("callback is not set")

// ProbeFunc is a callback function for Probe.
type ProbeFunc func(context.Context, *pb.ProbeMessage) (*emptypb.Empty, error)

// ReplyFunc is a callback function for Reply.
type ReplyFunc func(context.Context, *pb.ReplyMessage) (*emptypb.Empty, error)

// TerminateFunc is a callback function for Terminate.
type TerminateFunc func(context.Context, *pb.TerminateMessage) (*emptypb.Empty, error)

// Server is a gRPC server.
type Server struct {
	pb.UnsafeHSServiceServer

	grpcServer *grpc.Server
	listener   net.Listener

	onProbe     ProbeFunc
	onReply     ReplyFunc
	onTerminate TerminateFunc
}

// New creates a new gRPC server.
func New(listen string, opt ...grpc.ServerOption) (*Server, error) {
	listener, err := net.Listen("tcp", listen)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %v", err)
	}

	time.Sleep(time.Millisecond * 100) // give the listener time to start

	grpcServer := grpc.NewServer(opt...)
	server := &Server{
		grpcServer: grpcServer,
		listener:   listener,
	}

	pb.RegisterHSServiceServer(grpcServer, server)

	return server, nil
}

// Start starts the gRPC server.
func (s *Server) Start() error {
	if err := s.grpcServer.Serve(s.listener); err != nil {
		return fmt.Errorf("failed to serve gRPC: %v", err)
	}

	return nil
}

// Close closes the gRPC server.
func (s *Server) Close() {
	if s.listener != nil {
		s.listener.Close()
	}
	if s.grpcServer != nil {
		s.grpcServer.Stop()
	}
}

// OnProbe sets the callback function for probe.
func (s *Server) OnProbe(callback ProbeFunc) {
	s.onProbe = callback
}

// OnReply sets the callback function for reply.
func (s *Server) OnReply(callback ReplyFunc) {
	s.onReply = callback
}

// OnTerminate sets the callback function for terminate.
func (s *Server) OnTerminate(callback TerminateFunc) {
	s.onTerminate = callback
}

// Probe implements rpc.HSServiceServer.
func (s *Server) Probe(ctx context.Context, req *pb.ProbeMessage) (*emptypb.Empty, error) {
	if s.onProbe == nil {
		return nil, ErrCallbackNotSet
	}

	return s.onProbe(ctx, req)
}

// Reply implements rpc.HSServiceServer.
func (s *Server) Reply(ctx context.Context, req *pb.ReplyMessage) (*emptypb.Empty, error) {
	if s.onReply == nil {
		return nil, ErrCallbackNotSet
	}

	return s.onReply(ctx, req)
}

// Terminate implements rpc.HSServiceServer.
func (s *Server) Terminate(ctx context.Context, req *pb.TerminateMessage) (*emptypb.Empty, error) {
	if s.onTerminate == nil {
		return nil, ErrCallbackNotSet
	}

	return s.onTerminate(ctx, req)
}

// Ping implements rpc.LCRServiceServer.
func (s *Server) Ping(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
