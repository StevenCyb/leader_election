package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	pb "leadelection/pkg/lcr/internal/rpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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

// NewTLS creates a new gRPC client with TLS.
func NewTLS(address string, certPool *x509.CertPool) (*Client, error) {
	cred := credentials.NewTLS(&tls.Config{ServerName: "", RootCAs: certPool})

	return New(address, grpc.WithTransportCredentials(cred))
}

// Close closes the gRPC client.
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

// RPC for sending a message to the next process (to the left).
func (c *Client) Message(ctx context.Context, in *pb.LCRMessage) (*pb.LCRResponse, error) {
	return c.grpcClient.Message(ctx, in)
}

// RPC to notify the current process that the leader has been elected and termination is starting.
func (c *Client) NotifyTermination(ctx context.Context, in *pb.LCRMessage) (*pb.LCRResponse, error) {
	return c.grpcClient.NotifyTermination(ctx, in)
}

// RPC to ping the current node to check if it is still alive.
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.grpcClient.Ping(ctx, &emptypb.Empty{})

	return err
}
