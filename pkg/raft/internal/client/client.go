package client

import (
	"context"
	"fmt"

	pb "github.com/StevenCyb/leader_election/pkg/raft/internal/rpc"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Client is a gRPC client wrapper.
type Client struct {
	conn       *grpc.ClientConn
	grpcClient pb.RaftServiceClient
}

// New creates a new gRPC client.
func New(address string, opts ...grpc.DialOption) (*Client, error) {
	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %v", err)
	}

	grpcClient := pb.NewRaftServiceClient(conn)

	return &Client{
		grpcClient: grpcClient,
		conn:       conn,
	}, nil
}

// Close closes the gRPC client.
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

// RequestVote sends a vote request to the server.
func (c *Client) RequestVote(ctx context.Context, req *pb.VoteMessage) error {
	if c.grpcClient == nil {
		return nil
	}

	_, err := c.grpcClient.RequestVote(ctx, req)

	return err
}

// Heartbeat sends a heartbeat to the server.
func (c *Client) Heartbeat(ctx context.Context) error {
	if c.grpcClient == nil {
		return nil
	}

	_, err := c.grpcClient.Heartbeat(ctx, &emptypb.Empty{})

	return err
}
