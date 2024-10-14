package hs

import (
	"context"
	"errors"
	"leadelection/pkg/hs/internal/client"
	"leadelection/pkg/hs/internal/server"
	"leadelection/pkg/log"
	"strings"
	"sync"
	"time"

	pb "leadelection/pkg/hs/internal/rpc"

	"google.golang.org/grpc"
)

var ErrNodeAlreadyExists = errors.New("node already exists")
var ErrNodeNotExists = errors.New("node not exists")
var ErrNotEnoughNeighbors = errors.New("not enough neighbors")

// OnLeaderChangeFunc is a callback function that is called when the leader changes.
type OnLeaderChangeFunc func(leader *string)

// LeadElection represents the Hirschenberg and Sinclair leader election algorithm.
type LeadElection struct {
	// Unchanged fields
	uid                uint64
	listen             string
	server             *server.Server
	onLeaderChangeFunc OnLeaderChangeFunc
	logger             log.ILogger
	// Dynamic fields
	performElectionMutex sync.Mutex
	stop                 chan struct{}
	neighborsMutex       sync.RWMutex
	leftNeighbors        map[string]*client.Client
	rightNeighbors       map[string]*client.Client
	leader               *string
	leaderClient         *client.Client
}

// MustStart starts the the leader election service or panics if server can not start.
// New leader will be elected if current leader is unknown or unreachable within 5 seconds.
// If not at least one neighbor on each side is added, the leader election will fail.
func (le *LeadElection) MustStart(delay time.Duration, checkInterval time.Duration) {
	if len(le.leftNeighbors) == 0 || len(le.rightNeighbors) == 0 {
		panic(ErrNotEnoughNeighbors)
	}

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

	go func() {
		time.Sleep(delay)
		le.logger.Debug("Start watching leader")
		for {
			select {
			case <-le.stop:
				return
			default:
				if le.leader == nil {
					le.startElection()
					time.Sleep(checkInterval)
					continue
				}
				if *le.leader == le.listen {
					time.Sleep(checkInterval)
					continue
				}
				le.neighborsMutex.RLock()
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
				if err := le.leaderClient.Ping(ctx); err != nil {
					le.logger.Infof("Leader %d is not reachable, start election", *le.leader)
					le.neighborsMutex.RUnlock()
					le.startElection()
				} else {
					le.neighborsMutex.RUnlock()
				}
				cancel()
				time.Sleep(checkInterval)
			}
		}
	}()
}

// Stop stops the service.
func (le *LeadElection) Stop() {
	le.logger.Info("Stop leader election service")

	close(le.stop)
	le.server.Close()

	le.neighborsMutex.Lock()
	defer le.neighborsMutex.Unlock()

	for _, cl := range le.leftNeighbors {
		if cl != nil {
			cl.Close()
		}
	}
	for _, cl := range le.rightNeighbors {
		if cl != nil {
			cl.Close()
		}
	}

	le.leader = nil
	le.leaderClient = nil
	le.stop = make(chan struct{})
	le.leftNeighbors = map[string]*client.Client{}
	le.rightNeighbors = map[string]*client.Client{}
	le.performElectionMutex = sync.Mutex{}
}

// AddLeftNeighborNode adds a left neighbor node to the cluster.
func (le *LeadElection) AddLeftNeighborNode(addr string, opts ...grpc.DialOption) error {
	le.neighborsMutex.Lock()
	defer le.neighborsMutex.Unlock()

	if _, ok := le.leftNeighbors[addr]; ok {
		return ErrNodeAlreadyExists
	}

	cl, err := client.New(addr, opts...)
	if err != nil {
		return err
	}

	le.leftNeighbors[addr] = cl
	le.logger.Tracef("Added left neighbor node. now have", len(le.leftNeighbors), "left neighbors")

	return nil
}

// AddRightNeighborNode adds a right neighbor node to the cluster.
func (le *LeadElection) AddRightNeighborNode(addr string, opts ...grpc.DialOption) error {
	le.neighborsMutex.Lock()
	defer le.neighborsMutex.Unlock()

	if _, ok := le.leftNeighbors[addr]; ok {
		return ErrNodeAlreadyExists
	}

	cl, err := client.New(addr, opts...)
	if err != nil {
		return err
	}

	le.rightNeighbors[addr] = cl
	le.logger.Tracef("Added right neighbor node. now have", len(le.rightNeighbors), "left neighbors")

	return nil
}

// OnLeaderChange sets the callback function that is called when the leader changes.
// Leader is nil if no leader known yet or old leader is not reachable and new election is ongoing.
func (le *LeadElection) OnLeaderChange(f OnLeaderChangeFunc) {
	le.onLeaderChangeFunc = f
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
	}

	s.OnProbe(le.onProbe)
	s.OnReply(le.onReply)
	s.OnTerminate(le.onTerminate)

	return le, nil
}

// IsLeader returns if the current node is the leader.
func (le *LeadElection) IsLeader() bool {
	return le.leader != nil && *le.leader == le.listen
}

// GetLeaderSync returns the current leader's listen address.
// Might be nil if no leader known yet.
func (le *LeadElection) GetLeader() *string {
	return le.leader
}

func (le *LeadElection) startElection() {
	if ok := le.performElectionMutex.TryLock(); !ok {
		return
	}

	// TODO

	le.performElectionMutex.Unlock()
}

func (le *LeadElection) onProbe(context.Context, *pb.ProbeMessage) (*pb.HSResponse, error) {
	// TODO
	return nil, nil
}

func (le *LeadElection) onReply(context.Context, *pb.ReplyMessage) (*pb.HSResponse, error) {
	// TODO
	return nil, nil
}

func (le *LeadElection) onTerminate(context.Context, *pb.TerminateMessage) (*pb.HSResponse, error) {
	// TODO
	return nil, nil
}

func (le *LeadElection) sendProbe(probeMsg *pb.ProbeMessage, neighbors map[string]*client.Client, direction pb.Direction) {

}
