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

// Server is a gRPC server.
type Server struct {
	pb.UnsafeRaftServiceServer

	grpcServer *grpc.Server
	listener   net.Listener

	onLeaderAnnouncement RequestVoteCallback
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

// OnProbe sets the callback function for probe.
func (s *Server) OnLeaderAnnouncement(callback RequestVoteCallback) {
	s.onLeaderAnnouncement = callback
}

// OnRequestVote sets the callback function for request vote.
func (s *Server) OnRequestVote(callback RequestVoteCallback) {
	s.onLeaderAnnouncement = callback
}

// RequestVote implements rpc.RaftServiceServer.
func (s *Server) RequestVote(ctx context.Context, req *pb.VoteMessage) (*pb.VoteResponse, error) {
	if s.onLeaderAnnouncement == nil {
		return nil, ErrCallbackNotSet
	}

	return s.onLeaderAnnouncement(ctx, req)
}

// Heartbeat implements rpc.RaftServiceServer.
func (s *Server) Heartbeat(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
