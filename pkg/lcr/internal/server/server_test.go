package server

import (
	"context"
	"fmt"
	"leadelection/pkg/internal"
	"leadelection/pkg/lcr/internal/client"
	"leadelection/pkg/lcr/internal/rpc"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestServer(t *testing.T) {
	t.Parallel()

	port := internal.GetFreeLocalPort(t)
	listen := fmt.Sprintf("localhost:%d", port)
	s, err := New(listen)
	assert.NoError(t, err)

	c, err := client.New(listen, grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)

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
			assert.NoError(t, err)
		}
	}()

	time.Sleep(1 * time.Second)

	for i := 0; i < 10; i++ {
		assert.NoError(t, c.Ping(context.Background()))
		resp, err := c.Message(context.Background(), &rpc.LCRMessage{})
		assert.NoError(t, err)
		assert.Equal(t, rpc.Status_RECEIVED, resp.Status)
		resp, err = c.NotifyTermination(context.Background(), &rpc.LCRMessage{})
		assert.NoError(t, err)
		assert.Equal(t, rpc.Status_RECEIVED, resp.Status)
	}
}
