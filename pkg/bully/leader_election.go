package bully

import (
	"context"
	"errors"
	"fmt"
	"leadelection/pkg/bully/internal/client"
	pb "leadelection/pkg/bully/internal/rpc"
	"leadelection/pkg/bully/internal/server"
	"leadelection/pkg/log"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var ErrNodeAlreadyExists = errors.New("node already exists")
var ErrNodeNotExists = errors.New("node not exists")

// OnLeaderChangeFunc is a callback function that is called when the leader changes.
type OnLeaderChangeFunc func(leader *uint64)

// LeadElection represents the bully leader election algorithm.
type LeadElection struct {
	// Unchanged fields
	uid                uint64
	listen             string
	server             *server.Server
	onLeaderChangeFunc OnLeaderChangeFunc
	logger             log.ILogger
	// Dynamic fields
	stop         chan struct{}
	leaderUID    *uint64
	ontoElection sync.Mutex
	nodesMutex   sync.RWMutex
	lowerNodes   map[uint64]*client.Client
	higherNode   map[uint64]*client.Client
}

// New creates a new LeadElection instance.
func New(uid uint64, listen string, logger log.ILogger, opt ...grpc.ServerOption) (*LeadElection, error) {
	s, err := server.New(listen, opt...)
	if err != nil {
		return nil, err
	}

	le := &LeadElection{
		uid:        uid,
		listen:     listen,
		logger:     logger,
		server:     s,
		stop:       make(chan struct{}),
		lowerNodes: make(map[uint64]*client.Client),
		higherNode: make(map[uint64]*client.Client),
	}

	s.OnLeaderAnnouncement(le.onLeaderAnnouncement)
	s.OnElection(le.onElection)

	return le, nil
}

// MustStart starts the the leader elecvice or panics if server can not start.
// New leader will be elected if current leader is unknown or unreachable within 5 seconds.
func (le *LeadElection) MustStart(delay time.Duration, checkInterval time.Duration) {
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

	go le.watchLeader(delay, checkInterval)
}

// Stop stops the service.
func (le *LeadElection) Stop() {
	le.logger.Info("Stop leader election service")

	close(le.stop)
	le.server.Close()

	le.ontoElection.Lock()
	defer le.ontoElection.Unlock()
	le.nodesMutex.Lock()
	defer le.nodesMutex.Unlock()

	for _, cl := range le.lowerNodes {
		cl.Close()
	}
	for _, cl := range le.higherNode {
		cl.Close()
	}

	le.leaderUID = nil
	le.stop = make(chan struct{})
	le.lowerNodes = make(map[uint64]*client.Client)
	le.higherNode = make(map[uint64]*client.Client)
}

// OnLeaderChange sets the callback function that is called when the leader changes.
// Leader is nil if no leader known yet or old leader is not reachable and new election is ongoing.
func (le *LeadElection) OnLeaderChange(f OnLeaderChangeFunc) {
	le.onLeaderChangeFunc = f
}

// AddNode adds a new node to the cluster.
func (le *LeadElection) AddNode(uid uint64, addr string, opts ...grpc.DialOption) error {
	le.nodesMutex.Lock()
	defer le.nodesMutex.Unlock()

	if _, ok := le.lowerNodes[uid]; ok {
		return ErrNodeAlreadyExists
	}
	if _, ok := le.higherNode[uid]; ok {
		return ErrNodeAlreadyExists
	}

	cl, err := client.New(addr, opts...)
	if err != nil {
		return err
	}

	if uid < le.uid {
		le.logger.Infof("Added node %d, to lower group", uid)
		le.lowerNodes[uid] = cl
	} else if uid > le.uid {
		le.logger.Infof("Added node %d, to higher group", uid)
		le.higherNode[uid] = cl
	} else {
		return ErrNodeAlreadyExists
	}

	return nil
}

// RemoveNode removes a node from the cluster.
func (le *LeadElection) RemoveNode(uid uint64) error {
	le.nodesMutex.Lock()
	defer le.nodesMutex.Unlock()

	if _, ok := le.lowerNodes[uid]; ok {
		le.logger.Infof("Removed node %d, from lower group", uid)
		delete(le.lowerNodes, uid)
	} else if _, ok := le.higherNode[uid]; ok {
		le.logger.Infof("Removed node %d, from higher group", uid)
		delete(le.higherNode, uid)
	} else {
		return ErrNodeNotExists
	}

	return nil
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

func (le *LeadElection) onLeaderAnnouncement(_ context.Context, req *pb.LeaderAnnouncementMessage) (*emptypb.Empty, error) {
	le.logger.Debugf("Leader announcement received: %d", req.Uid)
	le.leaderUID = &req.Uid
	if le.onLeaderChangeFunc != nil {
		le.onLeaderChangeFunc(le.leaderUID)
	}

	return &emptypb.Empty{}, nil
}

func (le *LeadElection) onElection(_ context.Context) (*emptypb.Empty, error) {
	le.logger.Debug("Got election request, start election")

	le.ontoElection.TryLock()
	defer le.ontoElection.Unlock()

	wg := sync.WaitGroup{}
	var mu sync.Mutex
	var highestUID *uint64

	for uid, nodeClient := range le.higherNode {
		wg.Add(1)
		go func() {
			defer wg.Done()
			le.logger.Tracef("Send election request to %d", uid)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			if err := nodeClient.Elect(ctx); err != nil {
				le.logger.Tracef("Node %d is not reachable", uid)
			} else {
				le.logger.Tracef("Node %d is reachable and might be the leader", uid)
				mu.Lock()
				if highestUID == nil || *highestUID < uid {
					le.logger.Tracef("Node %d is the highest UID that responded till now", uid)
					tmp := uid
					highestUID = &tmp
				}
				mu.Unlock()
			}
			cancel()
		}()
	}
	wg.Wait()

	if highestUID == nil {
		le.logger.Debug("No Leader found, declare self as Leader")
		le.leaderUID = &le.uid
		for _, nodeClient := range le.lowerNodes {
			go func(nodeClient *client.Client) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				//nolint:errcheck
				nodeClient.LeaderAnnouncement(ctx, &pb.LeaderAnnouncementMessage{Uid: le.uid})
			}(nodeClient)
		}
	} else {
		le.logger.Debugf("Found potential Leader with UID %d", *highestUID)
	}

	return &emptypb.Empty{}, nil
}

func (le *LeadElection) startElection() {
	if !le.ontoElection.TryLock() {
		return
	}

	le.logger.Infof("Start election")
	le.leaderUID = nil

	//nolint:errcheck
	le.onElection(context.Background())
}

func (le *LeadElection) watchLeader(delay time.Duration, checkInterval time.Duration) {
	time.Sleep(delay)
	le.logger.Debug("Start watching leader")
	for {
		select {
		case <-le.stop:
			return
		default:
			le.nodesMutex.RLock()
			if le.leaderUID == nil {
				le.nodesMutex.RUnlock()
				le.startElection()
				time.Sleep(checkInterval)
				continue
			}
			if *le.leaderUID == le.uid {
				le.nodesMutex.RUnlock()
				time.Sleep(checkInterval)
				continue
			}

			leaderClient := le.higherNode[*le.leaderUID]
			if leaderClient == nil {
				panic(fmt.Errorf("leader %d not found in neighbors", *le.leaderUID))
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			if err := leaderClient.Ping(ctx); err != nil {
				le.logger.Infof("Leader %d is not reachable", *le.leaderUID)
				le.nodesMutex.RUnlock()
				le.startElection()
			} else {
				le.nodesMutex.RUnlock()
			}
			cancel()
			time.Sleep(checkInterval)
		}
	}
}
