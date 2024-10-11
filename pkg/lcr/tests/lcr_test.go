package tests

import (
	"fmt"
	"leadelection/pkg/internal"
	"leadelection/pkg/lcr"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestLeadElection_LCR(t *testing.T) {
	t.Parallel()

	portA := internal.GetFreeLocalPort(t)
	portB := internal.GetFreeLocalPort(t)
	portC := internal.GetFreeLocalPort(t)

	listenA := fmt.Sprintf("localhost:%d", portA)
	listenB := fmt.Sprintf("localhost:%d", portB)
	listenC := fmt.Sprintf("localhost:%d", portC)

	leA, err := lcr.New(1, listenA)
	assert.NoError(t, err)
	leB, err := lcr.New(2, listenB)
	assert.NoError(t, err)
	leC, err := lcr.New(3, listenC)
	assert.NoError(t, err)

	assert.NoError(t, leA.AddNode(2, listenB, grpc.WithTransportCredentials(insecure.NewCredentials())))
	assert.NoError(t, leA.AddNode(3, listenC, grpc.WithTransportCredentials(insecure.NewCredentials())))

	assert.NoError(t, leB.AddNode(1, listenA, grpc.WithTransportCredentials(insecure.NewCredentials())))
	assert.NoError(t, leB.AddNode(3, listenC, grpc.WithTransportCredentials(insecure.NewCredentials())))

	assert.NoError(t, leC.AddNode(1, listenA, grpc.WithTransportCredentials(insecure.NewCredentials())))
	assert.NoError(t, leC.AddNode(2, listenB, grpc.WithTransportCredentials(insecure.NewCredentials())))

	t.Cleanup(func() {
		leA.Stop()
		leB.Stop()
		leC.Stop()
	})

	go func() {
		assert.NoError(t, leA.Start(time.Microsecond*500, time.Microsecond*100))
	}()
	go func() {
		assert.NoError(t, leB.Start(time.Microsecond*500, time.Microsecond*100))
	}()
	go func() {
		assert.NoError(t, leC.Start(time.Microsecond*500, time.Microsecond*100))
	}()

	time.Sleep(time.Second)

	assert.Equal(t, 3, leA.GetLeader())
	assert.Equal(t, 3, leB.GetLeader())
	assert.Equal(t, 3, leC.GetLeader())
	// assert.False(t, leA.IsLeader())
	// assert.False(t, leB.IsLeader())
	// assert.True(t, leC.IsLeader())
}
