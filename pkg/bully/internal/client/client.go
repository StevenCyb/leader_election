package client

import (
	"context"
	"fmt"
	pb "leadelection/pkg/bully/internal/rpc"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Client is a gRPC client wrapper.
type Client struct {
	conn       *grpc.ClientConn
	grpcClient pb.BullyServiceClient
}

// New creates a new gRPC client.
func New(address string, opts ...grpc.DialOption) (*Client, error) {
	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %v", err)
	}

	grpcClient := pb.NewBullyServiceClient(conn)

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

// LeaderAnnouncement sends a leader announcement message to the gRPC server.
func (c *Client) LeaderAnnouncement(ctx context.Context, req *pb.LeaderAnnouncementMessage) error {
	if c.grpcClient == nil {
		return nil
	}

	_, err := c.grpcClient.LeaderAnnouncement(ctx, req)

	return err
}

// Elect for sending to check if instance is alive and as for leadership.
func (c *Client) Elect(ctx context.Context) error {
	if c.grpcClient == nil {
		return nil
	}
	_, err := c.grpcClient.Elect(ctx, &emptypb.Empty{})

	return err
}

// Elect for sending to check if instance is alive and as for leadership.
func (c *Client) Ping(ctx context.Context) error {
	if c.grpcClient == nil {
		return nil
	}
	_, err := c.grpcClient.Ping(ctx, &emptypb.Empty{})

	return err
}
