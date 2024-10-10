package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	pb "leadelection/pkg/peer/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Client is a gRPC client wrapper.
type Client struct {
	address    string
	conn       *grpc.ClientConn
	grpcClient pb.PeerServiceClient
}

// New creates a new gRPC client.
func New(address string, opts ...grpc.DialOption) (*Client, error) {
	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %v", err)
	}

	grpcClient := pb.NewPeerServiceClient(conn)

	return &Client{
		address:    address,
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

func (c Client) GetAddress() string {
	return c.address
}

func (c *Client) Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return c.grpcClient.Ping(ctx, in, opts...)
}
