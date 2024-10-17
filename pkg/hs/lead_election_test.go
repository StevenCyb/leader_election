package hs

import (
	"fmt"
	"testing"
	"time"

	"github.com/StevenCyb/leader_election/pkg/hs/internal/client"
	"github.com/StevenCyb/leader_election/pkg/internal"
	"github.com/StevenCyb/leader_election/pkg/log"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ClusterTester struct {
	t             *testing.T
	instances     map[uint64]*LeadElection
	instanceUIDs  internal.OrderedList[uint64]
	currentLeader uint64
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

	newInstance.OnLeaderChange(func(leader *uint64) {
		if leader != nil {
			require.Equal(c.t, int(c.currentLeader), int(*leader), fmt.Sprintf("leader mismatch in instance %d", id))
		}
	})

	c.instanceUIDs.AddOrdered(id)
	c.instances[id] = newInstance

	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	for id, instance := range c.instances {
		instance.leftNeighbors = []internal.Tuple[uint64, *client.Client]{}
		instance.rightNeighbors = []internal.Tuple[uint64, *client.Client]{}
		index := c.instanceUIDs.GetIndexFor(id)

		for i := index - 1; i >= 0; i-- {
			require.NoError(c.t, instance.AddLeftNeighborNode(c.instanceUIDs[i], c.instances[c.instanceUIDs[i]].listen, opt))
		}
		for i := index + 1; i < len(c.instanceUIDs); i++ {
			require.NoError(c.t, instance.AddRightNeighborNode(c.instanceUIDs[i], c.instances[c.instanceUIDs[i]].listen, opt))
		}
	}

	newInstance.MustStart(time.Millisecond*500, time.Millisecond*100)
}

func (c *ClusterTester) ExpectLeader(delay time.Duration, expect uint64) {
	c.t.Helper()

	c.currentLeader = expect
	time.Sleep(delay)

	for _, instance := range c.instances {
		actual := instance.GetLeader()

		require.NotNil(c.t, actual, fmt.Sprintf("leader in nil in instance %d", instance.uid))
		require.Equal(c.t, int(expect), int(*actual), fmt.Sprintf("leader mismatch in instance %d", instance.uid))
	}
}

func (c *ClusterTester) Kill(id uint64) {
	c.t.Helper()

	instance, ok := c.instances[id]
	require.True(c.t, ok)

	instance.Stop()
	delete(c.instances, id)
}

func TestLeadElection_HS_Single(t *testing.T) {
	t.Parallel()

	ct := ClusterTester{
		t:         t,
		instances: make(map[uint64]*LeadElection),
	}
	t.Cleanup(ct.Cleanup)

	ct.AddInstance(10)
	ct.ExpectLeader(time.Second, 10)
}

func TestLeadElection_HS_Simple(t *testing.T) {
	t.Parallel()

	ct := ClusterTester{
		t:         t,
		instances: make(map[uint64]*LeadElection),
	}
	t.Cleanup(ct.Cleanup)

	ct.AddInstance(10)
	ct.ExpectLeader(time.Second, 10)
	ct.AddInstance(20)
	ct.ExpectLeader(time.Second*2, 20)
	ct.AddInstance(15)
	ct.AddInstance(5)
	ct.ExpectLeader(time.Second*3, 20)
}

func TestLeadElection_HS_DeadLeader(t *testing.T) {
	t.Parallel()

	ct := ClusterTester{
		t:         t,
		instances: make(map[uint64]*LeadElection),
	}
	t.Cleanup(ct.Cleanup)

	ct.AddInstance(10)
	ct.AddInstance(20)
	ct.AddInstance(30)
	ct.ExpectLeader(time.Second*3, 30)
	ct.Kill(30)
	ct.ExpectLeader(time.Second*3, 20)
}
