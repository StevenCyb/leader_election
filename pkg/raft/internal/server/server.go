package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	pb "github.com/StevenCyb/leader_election/pkg/raft/internal/rpc"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var ErrCallbackNotSet = errors.New("callback is not set")

// RequestVoteCallback is a callback function for leader announcement.
type RequestVoteCallback func(context.Context, *pb.VoteMessage) (*pb.VoteResponse, error)

// HeartbeatCallback is a callback function for leader announcement.
type HeartbeatCallback func(context.Context, *pb.HeartbeatMessage) error

// Server is a gRPC server.
type Server struct {
	pb.UnsafeRaftServiceServer

	grpcServer *grpc.Server
	listener   net.Listener

	onRequestVote RequestVoteCallback
	onHeartbeat   HeartbeatCallback
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

	pb.RegisterRaftServiceServer(grpcServer, server)

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

// OnRequestVote sets the callback function for request vote.
func (s *Server) OnRequestVote(callback RequestVoteCallback) {
	s.onRequestVote = callback
}

// OnHeartbeat sets the callback function for heartbeat.
func (s *Server) OnHeartbeat(callback HeartbeatCallback) {
	s.onHeartbeat = callback
}

// RequestVote implements rpc.RaftServiceServer.
func (s *Server) RequestVote(ctx context.Context, req *pb.VoteMessage) (*pb.VoteResponse, error) {
	if s.onRequestVote == nil {
		return nil, ErrCallbackNotSet
	}

	return s.onRequestVote(ctx, req)
}

// Heartbeat implements rpc.RaftServiceServer.
func (s *Server) Heartbeat(ctx context.Context, req *pb.HeartbeatMessage) (*emptypb.Empty, error) {
	if s.onRequestVote == nil {
		return nil, ErrCallbackNotSet
	}

	return &emptypb.Empty{}, s.onHeartbeat(ctx, req)
}
