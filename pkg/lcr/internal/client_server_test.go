package internal

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/StevenCyb/leader_election/pkg/internal"
	"github.com/StevenCyb/leader_election/pkg/lcr/internal/client"
	"github.com/StevenCyb/leader_election/pkg/lcr/internal/rpc"
	"github.com/StevenCyb/leader_election/pkg/lcr/internal/server"

	globalInternal "github.com/StevenCyb/leader_election/pkg/internal"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/grpclog"
)

func TestClientServer(t *testing.T) {
	t.Parallel()

	port := internal.GetFreeLocalPort(t)
	listen := fmt.Sprintf("localhost:%d", port)
	s, err := server.New(listen)
	require.NoError(t, err)

	c, err := client.New(listen, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	s.OnMessage(func(_ context.Context, _ *rpc.LCRMessage) (*rpc.LCRResponse, error) {
		fmt.Println("message called")
		return &rpc.LCRResponse{
			Status: rpc.Status_RECEIVED,
		}, nil
	})
	s.OnNotifyTermination(func(ctx context.Context, l *rpc.LCRMessage) (*rpc.LCRResponse, error) {
		fmt.Println("termination called")
		return &rpc.LCRResponse{
			Status: rpc.Status_RECEIVED,
		}, nil
	})

	defer s.Close()
	go func() {
		err := s.Start()
		if err != nil && err != grpc.ErrServerStopped {
			if strings.Contains(err.Error(), "use of closed network connection") {
				// Ignore this error
				return
			}
			require.NoError(t, err)
		}
	}()

	time.Sleep(time.Second)

	for i := 0; i < 10; i++ {
		require.NoError(t, c.Ping(context.Background()))
		resp, err := c.Message(context.Background(), &rpc.LCRMessage{})
		require.NoError(t, err)
		require.Equal(t, rpc.Status_RECEIVED, resp.Status)
		resp, err = c.NotifyTermination(context.Background(), &rpc.LCRMessage{})
		require.NoError(t, err)
		require.Equal(t, rpc.Status_RECEIVED, resp.Status)
	}
}

func TestClientServer_TLS(t *testing.T) {
	t.Parallel()

	grpclog.SetLoggerV2(grpclog.NewLoggerV2WithVerbosity(os.Stderr, os.Stderr, os.Stderr, 3))

	caPrivateKeyPEM, caCertificatePEM := globalInternal.GenerateCaCert(t)
	privateKeyPEM, certificatePEM := globalInternal.GenerateCert(t, "localhost", false, caPrivateKeyPEM, caCertificatePEM)
	serverCert, err := tls.X509KeyPair([]byte(certificatePEM), []byte(privateKeyPEM))
	require.NoError(t, err)

	serverOpts := grpc.Creds(credentials.NewTLS(&tls.Config{Certificates: []tls.Certificate{serverCert}}))

	certPool := x509.NewCertPool()
	require.True(t, certPool.AppendCertsFromPEM([]byte(caCertificatePEM)))
	clientOpts := grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
		ServerName: "",
		RootCAs:    certPool,
	}))

	port := internal.GetFreeLocalPort(t)
	listen := fmt.Sprintf("localhost:%d", port)
	s, err := server.New(listen, serverOpts)
	require.NoError(t, err)

	c, err := client.New(listen, clientOpts)
	require.NoError(t, err)

	s.OnMessage(func(_ context.Context, _ *rpc.LCRMessage) (*rpc.LCRResponse, error) {
		fmt.Println("message called")
		return &rpc.LCRResponse{
			Status: rpc.Status_RECEIVED,
		}, nil
	})
	s.OnNotifyTermination(func(ctx context.Context, l *rpc.LCRMessage) (*rpc.LCRResponse, error) {
		fmt.Println("termination called")
		return &rpc.LCRResponse{
			Status: rpc.Status_RECEIVED,
		}, nil
	})

	t.Cleanup(s.Close)
	go func() {
		err := s.Start()
		if err != nil && err != grpc.ErrServerStopped {
			if strings.Contains(err.Error(), "use of closed network connection") {
				// Ignore this error
				return
			}
			require.NoError(t, err)
		}
	}()

	time.Sleep(time.Second)

	for i := 0; i < 10; i++ {
		require.NoError(t, c.Ping(context.Background()))
		resp, err := c.Message(context.Background(), &rpc.LCRMessage{})
		require.NoError(t, err)
		require.Equal(t, rpc.Status_RECEIVED, resp.Status)
		resp, err = c.NotifyTermination(context.Background(), &rpc.LCRMessage{})
		require.NoError(t, err)
		require.Equal(t, rpc.Status_RECEIVED, resp.Status)
	}
}

func TestClientServer_MutualTLS(t *testing.T) {
	t.Parallel()

	grpclog.SetLoggerV2(grpclog.NewLoggerV2WithVerbosity(os.Stderr, os.Stderr, os.Stderr, 3))

	// Generate and load certificates
	caPrivateKeyPEM, caCertificatePEM := globalInternal.GenerateCaCert(t)
	certPool := x509.NewCertPool()
	require.True(t, certPool.AppendCertsFromPEM([]byte(caCertificatePEM)))

	// Server
	serverCertKeyPEM, serverCertPEM := globalInternal.GenerateCert(t, "localhost", false, caPrivateKeyPEM, caCertificatePEM)
	serverCert, err := tls.X509KeyPair([]byte(serverCertPEM), []byte(serverCertKeyPEM))
	require.NoError(t, err)

	serverCreds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientCAs:    certPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	})
	serverOpts := grpc.Creds(serverCreds)

	port := internal.GetFreeLocalPort(t)
	listen := fmt.Sprintf("localhost:%d", port)
	s, err := server.New(listen, serverOpts)
	require.NoError(t, err)

	go func() {
		err := s.Start()
		if err != nil && err != grpc.ErrServerStopped {
			if strings.Contains(err.Error(), "use of closed network connection") {
				// Ignore this error
				return
			}
			require.NoError(t, err)
		}
	}()

	t.Cleanup(s.Close)
	time.Sleep(time.Second)

	// Client
	clientCertKeyPEM, clientCertPEM := globalInternal.GenerateCert(t, "client", true, caPrivateKeyPEM, caCertificatePEM)
	clientCert, err := tls.X509KeyPair([]byte(clientCertPEM), []byte(clientCertKeyPEM))
	require.NoError(t, err)

	clientCreds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
		ServerName:   "localhost",
	})
	clientOpts := grpc.WithTransportCredentials(clientCreds)

	c, err := client.New(listen, clientOpts)
	require.NoError(t, err)

	s.OnMessage(func(_ context.Context, _ *rpc.LCRMessage) (*rpc.LCRResponse, error) {
		fmt.Println("message called")
		return &rpc.LCRResponse{
			Status: rpc.Status_RECEIVED,
		}, nil
	})
	s.OnNotifyTermination(func(ctx context.Context, l *rpc.LCRMessage) (*rpc.LCRResponse, error) {
		fmt.Println("termination called")
		return &rpc.LCRResponse{
			Status: rpc.Status_RECEIVED,
		}, nil
	})

	t.Cleanup(s.Close)

	go func() {
		err := s.Start()
		if err != nil && err != grpc.ErrServerStopped {
			if strings.Contains(err.Error(), "use of closed network connection") {
				// Ignore this error
				return
			}
			require.NoError(t, err)
		}
	}()

	time.Sleep(time.Second * 2)

	for i := 0; i < 10; i++ {
		require.NoError(t, c.Ping(context.Background()))
		resp, err := c.Message(context.Background(), &rpc.LCRMessage{})
		require.NoError(t, err)
		require.Equal(t, rpc.Status_RECEIVED, resp.Status)
		resp, err = c.NotifyTermination(context.Background(), &rpc.LCRMessage{})
		require.NoError(t, err)
		require.Equal(t, rpc.Status_RECEIVED, resp.Status)
	}
}

func TestClientServer_MutualTLS_Unauthorized(t *testing.T) {
	t.Parallel()

	grpclog.SetLoggerV2(grpclog.NewLoggerV2WithVerbosity(os.Stderr, os.Stderr, os.Stderr, 3))

	// Generate and load certificates
	caPrivateKeyPEM, caCertificatePEM := globalInternal.GenerateCaCert(t)
	certPool := x509.NewCertPool()
	require.True(t, certPool.AppendCertsFromPEM([]byte(caCertificatePEM)))

	// Server
	serverCertKeyPEM, serverCertPEM := globalInternal.GenerateCert(t, "localhost", false, caPrivateKeyPEM, caCertificatePEM)
	serverCert, err := tls.X509KeyPair([]byte(serverCertPEM), []byte(serverCertKeyPEM))
	require.NoError(t, err)

	serverCreds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientCAs:    certPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	})
	serverOpts := grpc.Creds(serverCreds)

	port := internal.GetFreeLocalPort(t)
	listen := fmt.Sprintf("localhost:%d", port)
	s, err := server.New(listen, serverOpts)
	require.NoError(t, err)

	go func() {
		err := s.Start()
		if err != nil && err != grpc.ErrServerStopped {
			if strings.Contains(err.Error(), "use of closed network connection") {
				// Ignore this error
				return
			}
			require.NoError(t, err)
		}
	}()

	t.Cleanup(s.Close)
	time.Sleep(time.Second)

	// Client
	caPrivateKeyPEM, caCertificatePEM = globalInternal.GenerateCaCert(t)
	certPool = x509.NewCertPool()
	require.True(t, certPool.AppendCertsFromPEM([]byte(caCertificatePEM)))

	clientCertKeyPEM, clientCertPEM := globalInternal.GenerateCert(t, "client", true, caPrivateKeyPEM, caCertificatePEM)
	clientCert, err := tls.X509KeyPair([]byte(clientCertPEM), []byte(clientCertKeyPEM))
	require.NoError(t, err)

	clientCreds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
		ServerName:   "localhost",
	})
	clientOpts := grpc.WithTransportCredentials(clientCreds)

	c, err := client.New(listen, clientOpts)
	require.NoError(t, err)

	t.Cleanup(s.Close)

	go func() {
		err := s.Start()
		if err != nil && err != grpc.ErrServerStopped {
			if strings.Contains(err.Error(), "use of closed network connection") {
				// Ignore this error
				return
			}
			require.NoError(t, err)
		}
	}()

	time.Sleep(time.Second * 2)
	require.Error(t, c.Ping(context.Background()))
}
