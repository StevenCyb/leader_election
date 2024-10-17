package client

import (
	"context"
	"fmt"

	pb "github.com/StevenCyb/leader_election/pkg/cr/internal/rpc"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Client is a gRPC client wrapper.
type Client struct {
	conn       *grpc.ClientConn
	grpcClient pb.CRServiceClient
}

// New creates a new gRPC client.
func New(address string, opts ...grpc.DialOption) (*Client, error) {
	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %v", err)
	}

	grpcClient := pb.NewCRServiceClient(conn)

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

// RPC to request the current leader.
func (c *Client) Election(ctx context.Context, in *pb.Message) (*emptypb.Empty, error) {
	if c.grpcClient == nil {
		return nil, nil
	}

	return c.grpcClient.Election(ctx, in)
}

// RPC to notify the current node that it is the leader.
func (c *Client) Elected(ctx context.Context, in *pb.Message) (*emptypb.Empty, error) {
	if c.grpcClient == nil {
		return nil, nil
	}

	return c.grpcClient.Elected(ctx, in)
}

// RPC to ping the current node to check if it is still alive.
func (c *Client) Ping(ctx context.Context) error {
	if c.grpcClient == nil {
		return nil
	}
	_, err := c.grpcClient.Ping(ctx, &emptypb.Empty{})

	return err
}
