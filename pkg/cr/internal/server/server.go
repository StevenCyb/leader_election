package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	pb "github.com/StevenCyb/leader_election/pkg/cr/internal/rpc"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var ErrCallbackNotSet = errors.New("callback is not set")

// ElectedFunc is a callback function for OnElected.
type ElectedFunc func(context.Context, *pb.Message) (*emptypb.Empty, error)

// ElectionFunc is a callback function for OnElection.
type ElectionFunc func(context.Context, *pb.Message) (*emptypb.Empty, error)

// Server is a gRPC server.
type Server struct {
	pb.UnsafeCRServiceServer

	grpcServer *grpc.Server
	listener   net.Listener

	onElected  ElectedFunc
	onElection ElectionFunc
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

	pb.RegisterCRServiceServer(grpcServer, server)

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

// OnElected sets the callback function for.
func (s *Server) OnElected(callback ElectedFunc) {
	s.onElected = callback
}

// OnElection sets the callback function for Message.
func (s *Server) OnElection(callback ElectionFunc) {
	s.onElection = callback
}

// Elected implements rpc.CRServiceServer.
func (s *Server) Elected(ctx context.Context, req *pb.Message) (*emptypb.Empty, error) {
	if s.onElected == nil {
		return nil, ErrCallbackNotSet
	}

	return s.onElected(ctx, req)
}

// Election implements rpc.CRServiceServer.
func (s *Server) Election(ctx context.Context, req *pb.Message) (*emptypb.Empty, error) {
	if s.onElection == nil {
		return nil, ErrCallbackNotSet
	}

	return s.onElection(ctx, req)
}

// Ping implements rpc.LCRServiceServer.
func (s *Server) Ping(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
