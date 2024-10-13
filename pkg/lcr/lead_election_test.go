package lcr

import (
	"fmt"
	"leadelection/pkg/internal"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ClusterTester struct {
	t                 *testing.T
	instances         map[uint64]*LeadElection
	inactiveInstances map[uint64]*LeadElection
	currentLeader     uint64
}

func (c *ClusterTester) Cleanup() {
	c.t.Helper()

	for _, instance := range c.instances {
		instance.Stop()
	}
}

func (c *ClusterTester) AddInstance(id uint64) {
	c.t.Helper()

	port := internal.GetFreeLocalPort(c.t)
	listen := fmt.Sprintf("localhost:%d", port)
	newInstance, err := New(id, listen)
	require.NoError(c.t, err)

	for _, instance := range c.instances {
		require.NoError(c.t, instance.AddNode(id, listen, grpc.WithTransportCredentials(insecure.NewCredentials())))
		require.NoError(c.t, newInstance.AddNode(instance.uid, instance.listen, grpc.WithTransportCredentials(insecure.NewCredentials())))
	}
	for _, instance := range c.inactiveInstances {
		require.NoError(c.t, instance.AddNode(id, listen, grpc.WithTransportCredentials(insecure.NewCredentials())))
	}

	newInstance.MustStart(time.Millisecond*500, time.Millisecond*100)

	c.instances[id] = newInstance
}

func (c *ClusterTester) Kill(id uint64) {
	c.t.Helper()

	instance, ok := c.instances[id]
	require.True(c.t, ok)

	instance.Stop()
	delete(c.instances, id)
	c.inactiveInstances[id] = instance
}

func (c *ClusterTester) Revive(id uint64) {
	c.t.Helper()

	inactiveInstance, ok := c.inactiveInstances[id]
	require.True(c.t, ok)

	delete(c.inactiveInstances, id)

	newInstance, err := New(id, inactiveInstance.listen)
	require.NoError(c.t, err)

	newInstance.OnLeaderChange(func(leader *uint64) {
		if leader != nil {
			require.Equal(c.t, c.currentLeader, *leader)
		}
	})

	for _, instance := range c.instances {
		require.NoError(c.t, instance.RemoveNode(id), fmt.Sprintf("failed to remove node %d from instance %d", id, instance.uid))
		require.NoError(c.t, instance.AddNode(inactiveInstance.uid, inactiveInstance.listen, grpc.WithTransportCredentials(insecure.NewCredentials())))
		require.NoError(c.t, newInstance.AddNode(instance.uid, instance.listen, grpc.WithTransportCredentials(insecure.NewCredentials())))
	}

	newInstance.MustStart(time.Millisecond*500, time.Millisecond*100)

	c.instances[id] = newInstance
}

func (c *ClusterTester) ExpectLeader(delay time.Duration, expect uint64) {
	c.t.Helper()

	c.currentLeader = expect
	time.Sleep(delay)

	for _, instance := range c.instances {
		actual := instance.GetLeader()

		require.NotNil(c.t, actual, fmt.Sprintf("leader in nil in instance %d", instance.uid))
		require.Equal(c.t, expect, *actual, fmt.Sprintf("leader mismatch in instance %d", instance.uid))
	}
}

func TestLeadElection_LCR_Simple(t *testing.T) {
	t.Parallel()

	ct := ClusterTester{
		t:                 t,
		instances:         make(map[uint64]*LeadElection),
		inactiveInstances: make(map[uint64]*LeadElection),
	}
	t.Cleanup(ct.Cleanup)

	ct.AddInstance(10)
	ct.ExpectLeader(time.Second, 10)
	ct.AddInstance(20)
	ct.ExpectLeader(time.Second*2, 20)
	ct.AddInstance(30)
	ct.ExpectLeader(time.Second*2, 30)
}

func TestLeadElection_LCR_DeadLeader(t *testing.T) {
	t.Parallel()

	ct := ClusterTester{
		t:                 t,
		instances:         make(map[uint64]*LeadElection),
		inactiveInstances: make(map[uint64]*LeadElection),
	}
	t.Cleanup(ct.Cleanup)

	ct.AddInstance(10)
	ct.AddInstance(20)
	ct.AddInstance(30)
	ct.ExpectLeader(time.Second*2, 30)
	ct.Kill(30)
	ct.ExpectLeader(time.Second*2, 20)
}

func TestLeadElection_LCR_DeadLeader_Revived(t *testing.T) {
	t.Parallel()

	ct := ClusterTester{
		t:                 t,
		instances:         make(map[uint64]*LeadElection),
		inactiveInstances: make(map[uint64]*LeadElection),
	}
	t.Cleanup(ct.Cleanup)

	ct.AddInstance(10)
	ct.AddInstance(20)
	ct.ExpectLeader(time.Second*2, 20)
	ct.Kill(20)
	ct.ExpectLeader(time.Second*2, 10)
	ct.Revive(20)
	ct.ExpectLeader(time.Second*2, 20)
}
