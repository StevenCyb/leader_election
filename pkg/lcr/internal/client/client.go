package client

import (
	"context"
	"fmt"

	pb "github.com/StevenCyb/leader_election/pkg/lcr/internal/rpc"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Client is a gRPC client wrapper.
type Client struct {
	conn       *grpc.ClientConn
	grpcClient pb.LCRServiceClient
}

// New creates a new gRPC client.
func New(address string, opts ...grpc.DialOption) (*Client, error) {
	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %v", err)
	}

	grpcClient := pb.NewLCRServiceClient(conn)

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

// RPC for sending a message to the next process (to the left).
func (c *Client) Message(ctx context.Context, in *pb.LCRMessage) (*pb.LCRResponse, error) {
	if c.grpcClient == nil {
		return nil, nil
	}

	return c.grpcClient.Message(ctx, in)
}

// RPC to notify the current process that the leader has been elected and termination is starting.
func (c *Client) NotifyTermination(ctx context.Context, in *pb.LCRMessage) (*pb.LCRResponse, error) {
	if c.grpcClient == nil {
		return nil, nil
	}

	return c.grpcClient.NotifyTermination(ctx, in)
}

// RPC to ping the current node to check if it is still alive.
func (c *Client) Ping(ctx context.Context) error {
	if c.grpcClient == nil {
		return nil
	}
	_, err := c.grpcClient.Ping(ctx, &emptypb.Empty{})

	return err
}
