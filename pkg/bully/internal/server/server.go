package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	pb "leadelection/pkg/bully/internal/rpc"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var ErrCallbackNotSet = errors.New("callback is not set")

// LeaderAnnouncementCallback is a callback function for leader election announcement.
type LeaderAnnouncementCallback func(context.Context, *pb.LeaderAnnouncementMessage) (*emptypb.Empty, error)

// ElectionCallback is a callback function for election.
type ElectionCallback func(context.Context) (*emptypb.Empty, error)

// Server is a gRPC server.
type Server struct {
	pb.UnsafeBullyServiceServer

	grpcServer *grpc.Server
	listener   net.Listener

	onLeaderAnnouncement LeaderAnnouncementCallback
	onElection           ElectionCallback
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

	pb.RegisterBullyServiceServer(grpcServer, server)

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
func (s *Server) OnLeaderAnnouncement(callback LeaderAnnouncementCallback) {
	s.onLeaderAnnouncement = callback
}

// OnElection sets the callback function for election.
func (s *Server) OnElection(callback ElectionCallback) {
	s.onElection = callback
}

// LeaderAnnouncement implements rpc.BullyServiceServer.
func (s *Server) LeaderAnnouncement(ctx context.Context, req *pb.LeaderAnnouncementMessage) (*emptypb.Empty, error) {
	if s.onLeaderAnnouncement == nil {
		return nil, ErrCallbackNotSet
	}

	return s.onLeaderAnnouncement(ctx, req)
}

// Election implements rpc.BullyServiceServer.
func (s *Server) Elect(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	if s.onElection == nil {
		return nil, ErrCallbackNotSet
	}

	return s.onElection(ctx)
}

// Ping implements rpc.BullyServiceServer.
func (s *Server) Ping(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
