package bully

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Node struct representing a node in the system.
type Node struct {
	ID         int     // Unique ID of the node
	IsAlive    bool    // Whether the node is up or down
	Nodes      []*Node // List of other nodes in the system
	LeaderID   int     // Current leader ID
	mutex      sync.Mutex
	inElection bool // Is the node currently in an election
}

// NewNode creates a new node with a given ID.
func NewNode(id int, nodes []*Node) *Node {
	return &Node{
		ID:         id,
		IsAlive:    true, // All nodes start alive
		Nodes:      nodes,
		LeaderID:   -1,    // No leader initially
		inElection: false, // Not in an election initially
	}
}

// StartElection initiates an election by the current node.
func (n *Node) StartElection() {
	n.mutex.Lock()
	if n.inElection {
		n.mutex.Unlock()
		return // Prevent re-entering the election if already in one
	}
	n.inElection = true
	n.mutex.Unlock()

	fmt.Printf("Node %d starts an election.\n", n.ID)

	higherNodes := false
	for _, node := range n.Nodes {
		if node.ID > n.ID && node.IsAlive {
			fmt.Printf("Node %d sends election message to Node %d.\n", n.ID, node.ID)
			response := node.ElectionMessage(n.ID)
			if response {
				higherNodes = true
			}
		}
	}

	if !higherNodes {
		n.BecomeLeader()
	}

	n.mutex.Lock()
	n.inElection = false
	n.mutex.Unlock()
}

// ElectionMessage simulates receiving an election message from another node.
func (n *Node) ElectionMessage(callerID int) bool {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	if !n.IsAlive || n.inElection {
		return false // Do not respond if node is down or already in election
	}

	fmt.Printf("Node %d received election message from Node %d.\n", n.ID, callerID)
	go n.StartElection() // Start an election if I received a message from a lower ID node
	return true
}

// BecomeLeader marks this node as the leader and announces it to all other nodes.
func (n *Node) BecomeLeader() {
	n.mutex.Lock()
	n.LeaderID = n.ID
	n.mutex.Unlock()
	fmt.Printf("Node %d is now the leader.\n", n.ID)

	for _, node := range n.Nodes {
		if node.IsAlive {
			node.AnnounceLeader(n.ID)
		}
	}
}

// AnnounceLeader receives the leader announcement.
func (n *Node) AnnounceLeader(leaderID int) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	if !n.IsAlive {
		return
	}

	n.LeaderID = leaderID
	fmt.Printf("Node %d recognizes Node %d as the leader.\n", n.ID, leaderID)
}

// SimulateFailure turns off a node to simulate failure.
func (n *Node) SimulateFailure() {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.IsAlive = false
	fmt.Printf("Node %d has failed.\n", n.ID)
}

// SimulateRecovery brings a node back to life.
func (n *Node) SimulateRecovery() {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.IsAlive = true
	fmt.Printf("Node %d has recovered.\n", n.ID)
	go n.StartElection() // Start election upon recovery
}

// CheckLeader checks if the current leader is alive.
func (n *Node) CheckLeader() {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	if !n.IsAlive {
		return
	}

	for _, node := range n.Nodes {
		if node.ID == n.LeaderID && node.IsAlive {
			fmt.Printf("Node %d confirms that Node %d is still the leader.\n", n.ID, n.LeaderID)
			return
		}
	}

	fmt.Printf("Node %d detects that the leader (Node %d) has failed. Starting election...\n", n.ID, n.LeaderID)
	go n.StartElection()
}

func TestBully(t *testing.T) {
	// Create a list of nodes in the system
	var nodes []*Node
	for i := 1; i <= 5; i++ {
		node := NewNode(i, nodes)
		nodes = append(nodes, node)
	}

	// Update the nodes to know about each other
	for _, node := range nodes {
		node.Nodes = nodes
	}

	// Test Initial Election (Node 1 initiates an election)
	nodes[0].StartElection()

	// Wait for the election to complete
	time.Sleep(2 * time.Second)

	// Verify that Node 5 (highest-ID) is the leader
	require.Equal(t, 5, nodes[4].LeaderID, "Node 5 should be the leader")
	for _, node := range nodes {
		require.Equal(t, 5, node.LeaderID, "All nodes should recognize Node 5 as the leader")
	}

	// Simulate the failure of the current leader (Node 5)
	nodes[4].SimulateFailure()

	// Node 1 starts another election after leader failure
	nodes[0].StartElection()

	// Wait for the election to complete
	time.Sleep(2 * time.Second)

	// Verify that Node 4 (next highest-ID) becomes the leader
	require.Equal(t, 4, nodes[3].LeaderID, "Node 4 should be the new leader after Node 5 fails")
	for _, node := range nodes {
		if node.IsAlive {
			require.Equal(t, 4, node.LeaderID, "All alive nodes should recognize Node 4 as the leader")
		}
	}

	// Simulate the recovery of Node 5
	nodes[4].SimulateRecovery()

	// Wait for the election initiated by Node 5 to complete
	time.Sleep(2 * time.Second)

	// Verify that Node 5 is re-elected as the leader after recovery
	require.Equal(t, 5, nodes[4].LeaderID, "Node 5 should reclaim leadership after recovery")
	for _, node := range nodes {
		require.Equal(t, 5, node.LeaderID, "All nodes should recognize Node 5 as the leader after recovery")
	}
}
