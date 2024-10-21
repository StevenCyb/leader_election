package raft

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/StevenCyb/leader_election/pkg/log"
	"github.com/StevenCyb/leader_election/pkg/raft/internal"
	"github.com/StevenCyb/leader_election/pkg/raft/internal/client"
	pb "github.com/StevenCyb/leader_election/pkg/raft/internal/rpc"
	"github.com/StevenCyb/leader_election/pkg/raft/internal/server"
	"golang.org/x/exp/rand"
	"google.golang.org/grpc"
)

var ErrNodeAlreadyExists = errors.New("node already exists")
var ErrNodeNotExists = errors.New("node not exists")

// LeaderElectionConflictResolveFunc is a callback function that is called when a leader conflict is detected.
// The return value is used to determine if the leader should be changed.
type LeaderElectionConflictResolveFunc func(potentialLeader uint64) bool

// OnLeaderChangeFunc is a callback function that is called when the leader changes.
type OnLeaderChangeFunc func(leader *uint64)

type LeadElection struct {
	// Unchanged fields
	uid                       uint64
	listen                    string
	server                    *server.Server
	onLeaderChangeFunc        OnLeaderChangeFunc
	logger                    log.ILogger
	leaderConflictResolveFunc LeaderElectionConflictResolveFunc
	// Dynamic fields
	stop               chan struct{}
	nodesMutex         sync.RWMutex
	nodes              map[uint64]*client.Client
	electionPhaseMutex sync.Mutex
	term               uint64
	leaderUID          *uint64
	votedFor           *uint64
	timer              internal.Timer
}

// New creates a new LeadElection instance.
func New(uid uint64, listen string, logger log.ILogger, leaderConflictResolveFunc LeaderElectionConflictResolveFunc, opt ...grpc.ServerOption) (*LeadElection, error) {
	s, err := server.New(listen, opt...)
	if err != nil {
		return nil, err
	}

	le := &LeadElection{
		uid:                       uid,
		listen:                    listen,
		logger:                    logger,
		server:                    s,
		leaderConflictResolveFunc: leaderConflictResolveFunc,
		stop:                      make(chan struct{}),
		nodes:                     make(map[uint64]*client.Client),
		timer:                     internal.Timer{},
	}

	s.OnRequestVote(le.onRequestVote)
	s.OnHeartbeat(le.onHeartbeat)

	return le, nil
}

// MustStart starts the leader election or panics if server can not start.
// `heartbeatTimeout` is the interval between heartbeats and
// the `electionTimeout` is random between 50ms-100ms + `heartbeatTimeout`.
func (le *LeadElection) MustStart(heartbeatTimeout time.Duration) {
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

	go le.watchLeader(heartbeatTimeout)
}

// Stop stops the service.
func (le *LeadElection) Stop() {
	le.logger.Info("Stop leader election service")

	le.timer.Stop()
	close(le.stop)
	le.server.Close()

	le.nodesMutex.Lock()
	defer le.nodesMutex.Unlock()
	le.electionPhaseMutex.Lock()
	defer le.electionPhaseMutex.Unlock()

	for _, cl := range le.nodes {
		cl.Close()
	}

	le.stop = make(chan struct{})
	le.term = 0
	le.leaderUID = nil
	le.votedFor = nil
	le.nodes = make(map[uint64]*client.Client)
	le.timer = internal.Timer{}
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

func (le *LeadElection) onRequestVote(_ context.Context, req *pb.VoteMessage) (*pb.VoteResponse, error) {
	le.logger.Debugf("Received vote request from %d on term %d", req.Uid, req.Term)

	le.electionPhaseMutex.Lock()
	defer le.electionPhaseMutex.Unlock()

	if req.Term < le.term {
		le.logger.Debugf("Reject vote request from %d on term %d, because term is old", req.Uid, req.Term)

		resp := &pb.VoteResponse{
			Term: 0,
			Uid:  0,
		}

		return resp, nil
	} else if le.votedFor != nil {
		le.logger.Debugf("Reject vote request from %d on term %d, because already voted for %d", req.Uid, req.Term, *le.votedFor)

		resp := &pb.VoteResponse{
			Term: 0,
			Uid:  0,
		}

		return resp, nil
	}

	le.logger.Debugf("Vote for %d on term %d", req.Uid, req.Term)
	le.timer.Stop()
	le.votedFor = &req.Uid
	resp := &pb.VoteResponse{
		Term: req.Term,
		Uid:  req.Uid,
	}

	return resp, nil
}

func (le *LeadElection) onHeartbeat(_ context.Context, req *pb.HeartbeatMessage) error {
	le.logger.Debugf("Received heartbeat from %d on term %d", req.Uid, req.Term)

	le.electionPhaseMutex.Lock()
	defer le.electionPhaseMutex.Unlock()

	if req.Term > le.term || le.leaderUID == nil {
		le.logger.Debugf("Elected leader %d on term %d", req.Uid, req.Term)
		le.term = req.Term
		le.leaderUID = &req.Uid
		le.votedFor = nil
		le.timer.Reset()
		if le.onLeaderChangeFunc != nil {
			le.onLeaderChangeFunc(le.leaderUID)
		}
	} else if req.Term == le.term {
		if req.Uid != *le.leaderUID {
			if le.leaderConflictResolveFunc(req.Uid) {
				le.leaderUID = &req.Uid
			}
		}
		le.timer.Reset()
	}

	return nil
}

func (le *LeadElection) watchLeader(heartbeatInterval time.Duration) {
	le.logger.Debug("Start watching leader")
	le.timer.Set(heartbeatInterval + time.Duration(rand.Intn(150+1))*time.Millisecond)
	le.timer.OnTrigger(func() {
		le.startElection(heartbeatInterval)
	})
	le.timer.Reset()

	for {
		select {
		case <-le.stop:
			return
		default:
			le.nodesMutex.Lock()
			le.electionPhaseMutex.Lock()
			if le.leaderUID != nil && *le.leaderUID == le.uid {
				le.electionPhaseMutex.Unlock()
				for _, cl := range le.nodes {
					//nolint:errcheck
					go func(cl *client.Client) {
						ctx, cancel := context.WithTimeout(context.Background(), heartbeatInterval)
						defer cancel()

						cl.Heartbeat(ctx, &pb.HeartbeatMessage{
							Term: le.term,
							Uid:  le.uid,
						})
					}(cl)
				}
			} else {
				le.electionPhaseMutex.Unlock()
			}
			le.nodesMutex.Unlock()
			time.Sleep(heartbeatInterval)
		}
	}
}

func (le *LeadElection) startElection(heartbeatTimeout time.Duration) {
	le.electionPhaseMutex.Lock()
	if le.votedFor != nil {
		le.electionPhaseMutex.Unlock()
		return
	}

	le.term++
	le.votedFor = &le.uid
	le.logger.Debugf("Start election with term %d", le.term)
	le.electionPhaseMutex.Unlock()

	le.nodesMutex.Lock()
	majorityCount := len(le.nodes) / 2
	if len(le.nodes)%2 == 0 {
		majorityCount++
	}
	voteCount := 1
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}

	for uid, cl := range le.nodes {
		wg.Add(1)
		go func(cl *client.Client) {
			le.logger.Tracef("Request vote from %d on term %d", uid, le.term)
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), heartbeatTimeout)
			defer cancel()

			req := &pb.VoteMessage{Uid: le.uid, Term: le.term}
			if resp, err := cl.RequestVote(ctx, req); err == nil {
				if resp.Uid == le.uid && resp.Term == le.term {
					le.logger.Tracef("Received vote from %d on term %d", uid, le.term)
					mu.Lock()
					voteCount++
					mu.Unlock()
				} else {
					le.logger.Tracef("Failed to vote from %d on term %d", uid, le.term)
				}
			} else {
				le.logger.Tracef("Failed to request vote from %d on term %d: %v", uid, le.term, err)
				mu.Lock()
				voteCount++
				mu.Unlock()
			}
		}(cl)
	}

	le.nodesMutex.Unlock()
	wg.Wait()

	if voteCount >= majorityCount {
		le.logger.Debugf("Elected leader %d on term %d", le.uid, le.term)
		le.electionPhaseMutex.Lock()
		le.leaderUID = &le.uid
		le.electionPhaseMutex.Unlock()
		if le.onLeaderChangeFunc != nil {
			le.onLeaderChangeFunc(le.leaderUID)
		}
	} else {
		le.logger.Debugf("Failed to elect leader on term %d", le.term)
		le.timer.Set(heartbeatTimeout + time.Duration(rand.Intn(150+1))*time.Millisecond)
		le.timer.Reset()
	}
}
