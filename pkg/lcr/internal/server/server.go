package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"time"

	pb "leadelection/pkg/lcr/internal/rpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/emptypb"
)

var ErrCallbackNotSet = errors.New("callback is not set")

// NotifyTerminationFunc is a callback function for NotifyTermination.
type NotifyTerminationFunc func(context.Context, *pb.LCRMessage) (*pb.LCRResponse, error)

// MessageFunc is a callback function for Message.
type MessageFunc func(context.Context, *pb.LCRMessage) (*pb.LCRResponse, error)

// Server is a gRPC server.
type Server struct {
	pb.UnsafeLCRServiceServer

	grpcServer *grpc.Server
	listener   net.Listener

	onNotifyTermination NotifyTerminationFunc
	onMessage           MessageFunc
}

// New creates a new gRPC server.
func New(listen string, opt ...grpc.ServerOption) (*Server, error) {
	listener, err := net.Listen("tcp", listen)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %v", err)
	}

	time.Sleep(time.Millisecond * 100) // give the listener time to start

	grpcServer := grpc.NewServer(opt...)
	pb.RegisterLCRServiceServer(grpcServer, &Server{})

	return &Server{
		grpcServer: grpcServer,
		listener:   listener,
	}, nil
}

// NewTLS creates a new gRPC server with TLS.
func NewTLS(listen string, cert tls.Certificate) (*Server, error) {
	cred := credentials.NewTLS(&tls.Config{Certificates: []tls.Certificate{cert}})

	return New(listen, grpc.Creds(cred))
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

// OnNotifyTermination sets the callback function for NotifyTermination.
func (s *Server) OnNotifyTermination(callback NotifyTerminationFunc) {
	s.onNotifyTermination = callback
}

// OnMessage sets the callback function for Message.
func (s *Server) OnMessage(callback MessageFunc) {
	s.onMessage = callback
}

// NotifyTermination implements rpc.LCRServiceServer.
func (s *Server) NotifyTermination(ctx context.Context, req *pb.LCRMessage) (*pb.LCRResponse, error) {
	if s.onNotifyTermination == nil {
		return nil, ErrCallbackNotSet
	}

	return s.onNotifyTermination(ctx, req)
}

// Message implements rpc.LCRServiceServer.
func (s *Server) Message(ctx context.Context, req *pb.LCRMessage) (*pb.LCRResponse, error) {
	if s.onMessage == nil {
		return nil, ErrCallbackNotSet
	}

	return s.onMessage(ctx, req)
}

// Ping implements rpc.LCRServiceServer.
func (s *Server) Ping(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
