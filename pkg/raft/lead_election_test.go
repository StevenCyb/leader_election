package raft

import (
	"fmt"
	"testing"
	"time"

	"github.com/StevenCyb/leader_election/pkg/internal"
	"github.com/StevenCyb/leader_election/pkg/log"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ClusterTester struct {
	t                 *testing.T
	instances         map[uint64]*LeadElection
	inactiveInstances map[uint64]*LeadElection
	leader            *uint64
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
	logger := log.New().SetName(fmt.Sprintf("I%d", id)).Build()
	newInstance, err := New(id, listen, logger)
	require.NoError(c.t, err)

	for _, instance := range c.instances {
		require.NoError(c.t, instance.AddNode(id, listen, grpc.WithTransportCredentials(insecure.NewCredentials())))
		require.NoError(c.t, newInstance.AddNode(instance.uid, instance.listen, grpc.WithTransportCredentials(insecure.NewCredentials())))
	}
	for _, instance := range c.inactiveInstances {
		require.NoError(c.t, instance.AddNode(id, listen, grpc.WithTransportCredentials(insecure.NewCredentials())))
	}

	newInstance.MustStart(time.Millisecond * 500)

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

	newInstance, err := New(id, inactiveInstance.listen, inactiveInstance.logger)
	require.NoError(c.t, err)

	newInstance.OnLeaderChange(func(leader *uint64) {
		require.NotNil(c.t, leader)
	})

	for _, instance := range c.instances {
		require.NoError(c.t, instance.RemoveNode(id), fmt.Sprintf("failed to remove node %d from instance %d", id, instance.uid))
		require.NoError(c.t, instance.AddNode(inactiveInstance.uid, inactiveInstance.listen, grpc.WithTransportCredentials(insecure.NewCredentials())))
		require.NoError(c.t, newInstance.AddNode(instance.uid, instance.listen, grpc.WithTransportCredentials(insecure.NewCredentials())))
	}

	newInstance.MustStart(time.Millisecond * 300)

	c.instances[id] = newInstance
}

func (c *ClusterTester) ExpectLeader(delay time.Duration) {
	c.t.Helper()

	time.Sleep(delay)
	c.leader = nil

	for _, instance := range c.instances {
		actual := instance.GetLeader()
		require.NotNil(c.t, actual, fmt.Sprintf("leader in nil in instance %d", instance.uid))

		if c.leader == nil {
			c.leader = actual
		}
		require.Equal(c.t, int(*c.leader), int(*actual), fmt.Sprintf("leader mismatch in instance %d", instance.uid))
	}
}

func TestLeadElection_Raft_Single(t *testing.T) {
	t.Parallel()

	ct := ClusterTester{
		t:                 t,
		instances:         make(map[uint64]*LeadElection),
		inactiveInstances: make(map[uint64]*LeadElection),
	}
	t.Cleanup(ct.Cleanup)

	ct.AddInstance(10)
	ct.ExpectLeader(time.Second)
}

func TestLeadElection_Raft_Simple(t *testing.T) {
	t.Parallel()

	ct := ClusterTester{
		t:                 t,
		instances:         make(map[uint64]*LeadElection),
		inactiveInstances: make(map[uint64]*LeadElection),
	}
	t.Cleanup(ct.Cleanup)

	ct.AddInstance(10)
	ct.ExpectLeader(time.Second)
	ct.AddInstance(20)
	ct.ExpectLeader(time.Second)
	ct.AddInstance(15)
	ct.AddInstance(5)
	ct.ExpectLeader(time.Second * 2)
}

func TestLeadElection_Raft_DeadLeader(t *testing.T) {
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
	ct.ExpectLeader(time.Second)
	ct.Kill(*ct.leader)
	ct.ExpectLeader(time.Second * 2)
}

func TestLeadElection_Raft_DeadLeader_Revived(t *testing.T) {
	t.Parallel()

	ct := ClusterTester{
		t:                 t,
		instances:         make(map[uint64]*LeadElection),
		inactiveInstances: make(map[uint64]*LeadElection),
	}
	t.Cleanup(ct.Cleanup)

	ct.AddInstance(10)
	ct.AddInstance(20)
	ct.ExpectLeader(time.Second)
	leader := *ct.leader
	ct.Kill(leader)
	ct.ExpectLeader(time.Second * 2)
	ct.Revive(leader)
	ct.ExpectLeader(time.Second * 2)
}
