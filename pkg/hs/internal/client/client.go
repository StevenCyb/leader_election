package client

import (
	"context"
	"fmt"
	pb "leadelection/pkg/hs/internal/rpc"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Client is a gRPC client wrapper.
type Client struct {
	conn       *grpc.ClientConn
	grpcClient pb.HSServiceClient
}

// New creates a new gRPC client.
func New(address string, opts ...grpc.DialOption) (*Client, error) {
	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %v", err)
	}

	grpcClient := pb.NewHSServiceClient(conn)

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

// Probe for sending a message to the next process (to the left).
func (c *Client) Probe(ctx context.Context, in *pb.ProbeMessage) error {
	if c.grpcClient == nil {
		return nil
	}

	_, err := c.grpcClient.Probe(ctx, in)

	return err
}

// Reply for sending a message to the previous process (to the right).
func (c *Client) Reply(ctx context.Context, in *pb.ReplyMessage) error {
	if c.grpcClient == nil {
		return nil
	}

	_, err := c.grpcClient.Reply(ctx, in)

	return err
}

// Terminate for sending a message to terminate the process.
func (c *Client) Terminate(ctx context.Context, in *pb.TerminateMessage) error {
	if c.grpcClient == nil {
		return nil
	}

	_, err := c.grpcClient.Terminate(ctx, in)

	return err
}

// RPC to ping the current node to check if it is still alive.
func (c *Client) Ping(ctx context.Context) error {
	if c.grpcClient == nil {
		return nil
	}
	_, err := c.grpcClient.Ping(ctx, &emptypb.Empty{})

	return err
}
