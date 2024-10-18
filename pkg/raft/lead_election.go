package raft

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/StevenCyb/leader_election/pkg/log"
	"github.com/StevenCyb/leader_election/pkg/raft/internal/client"
	pb "github.com/StevenCyb/leader_election/pkg/raft/internal/rpc"
	"github.com/StevenCyb/leader_election/pkg/raft/internal/server"
	"google.golang.org/grpc"
)

var ErrNodeAlreadyExists = errors.New("node already exists")
var ErrNodeNotExists = errors.New("node not exists")

// OnLeaderChangeFunc is a callback function that is called when the leader changes.
type OnLeaderChangeFunc func(leader *uint64)

type LeadElection struct {
	// Unchanged fields
	uid                uint64
	listen             string
	server             *server.Server
	onLeaderChangeFunc OnLeaderChangeFunc
	logger             log.ILogger
	// Dynamic fields
	stop       chan struct{}
	nodesMutex sync.RWMutex
	nodes      map[uint64]*client.Client
	// electionPhase sync.Mutex TODO
	term      uint64
	leaderUID *uint64
	votedFor  *uint64
}

// New creates a new LeadElection instance.
func New(uid uint64, listen string, logger log.ILogger, opt ...grpc.ServerOption) (*LeadElection, error) {
	s, err := server.New(listen, opt...)
	if err != nil {
		return nil, err
	}

	le := &LeadElection{
		uid:    uid,
		listen: listen,
		logger: logger,
		server: s,
		stop:   make(chan struct{}),
		nodes:  make(map[uint64]*client.Client),
	}

	s.OnRequestVote(le.onRequestVote)
	s.OnHeartbeat(le.onHeartbeat)

	return le, nil
}

// MustStart starts the leader election or panics if server can not start.
// New leader will be elected if current leader is unknown or unreachable within 5 seconds.
func (le *LeadElection) MustStart() {
	le.logger.Info("Start leader election service")
	go func() {
		err := le.server.Start()
		if err != nil && err != grpc.ErrServerStopped {
			if strings.Contains(err.Error(), "use of closed network connection") {
				// Ignore this error
				return
			}
			panic(err)
		}
	}()

	go le.watchLeader()
}

// Stop stops the service.
func (le *LeadElection) Stop() {
	le.logger.Info("Stop leader election service")

	close(le.stop)
	le.server.Close()

	le.nodesMutex.Lock()
	defer le.nodesMutex.Unlock()

	for _, cl := range le.nodes {
		cl.Close()
	}

	le.stop = make(chan struct{})
	le.term = 0
	le.leaderUID = nil
	le.votedFor = nil
	le.nodes = make(map[uint64]*client.Client)
}

// OnLeaderChange sets the callback function that is called when the leader changes.
// Leader is nil if no leader known yet or old leader is not reachable and new election is ongoing.
func (le *LeadElection) OnLeaderChange(f OnLeaderChangeFunc) {
	le.onLeaderChangeFunc = f
}

// IsLeader returns if the current node is the leader.
func (le *LeadElection) IsLeader() bool {
	return le.leaderUID != nil && *le.leaderUID == le.uid
}

// GetLeaderSync returns the current leader's UID.
// Might be nil if no leader known yet.
func (le *LeadElection) GetLeader() *uint64 {
	return le.leaderUID
}

// AddNode adds a new node to the cluster.
func (le *LeadElection) AddNode(uid uint64, addr string, opts ...grpc.DialOption) error {
	le.nodesMutex.Lock()
	defer le.nodesMutex.Unlock()

	if _, ok := le.nodes[uid]; ok {
		return ErrNodeAlreadyExists
	}

	cl, err := client.New(addr, opts...)
	if err != nil {
		return err
	}

	le.logger.Infof("Added node %d", uid)
	le.nodes[uid] = cl

	return nil
}

// RemoveNode removes a node from the cluster.
func (le *LeadElection) RemoveNode(uid uint64) error {
	le.nodesMutex.Lock()
	defer le.nodesMutex.Unlock()

	if _, ok := le.nodes[uid]; !ok {
		return ErrNodeNotExists
	}

	le.logger.Infof("Removed node %d", uid)
	delete(le.nodes, uid)

	return nil
}

func (le *LeadElection) onRequestVote(context.Context, *pb.VoteMessage) (*pb.VoteResponse, error) {
	// TODO
	return nil, nil
}

func (le *LeadElection) onHeartbeat(_ context.Context, req *pb.HeartbeatMessage) error {
	le.logger.Debugf("Received heartbeat from %d on term %d", req.Uid, req.Term)

	if req.Term > le.term {
		le.logger.Debugf("Leader is now %d", req.Uid)
		le.term = req.Term
		le.leaderUID = &req.Uid
		le.votedFor = nil
		le.onLeaderChangeFunc(le.leaderUID)
	}

	return nil
}

func (le *LeadElection) watchLeader() {
	le.logger.Debug("Start watching leader")
	for {
		select {
		case <-le.stop:
			return
		default:
			// TODO
		}
	}
}
