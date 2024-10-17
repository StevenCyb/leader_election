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
	stop       chan struct{}
	leaderUID  *uint64
	nodeMutex  sync.RWMutex
	lowerNodes map[uint64]*client.Client
	higherNode map[uint64]*client.Client
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

	return le, nil
}

// MustStart starts the the leader election service or panics if server can not start.
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

	go func() {
		time.Sleep(delay)
		le.logger.Debug("Start watching leader")
		for {
			select {
			case <-le.stop:
				return
			default:
				le.nodeMutex.RLock()
				if le.leaderUID == nil {
					le.nodeMutex.RUnlock()
					le.startElection()
					time.Sleep(checkInterval)
					continue
				}
				if *le.leaderUID == le.uid {
					le.nodeMutex.RUnlock()
					time.Sleep(checkInterval)
					continue
				}

				var leaderClient *client.Client
				for _, cl := range append(le.leftNeighbors, le.rightNeighbors...) {
					if cl.First == *le.leaderUID {
						leaderClient = cl.Second
						break
					}
				}

				if leaderClient == nil {
					panic(fmt.Errorf("leader %d not found in neighbors", *le.leaderUID))
				}

				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				if err := leaderClient.Ping(ctx); err != nil {
					le.logger.Infof("Leader %d is not reachable", *le.leaderUID)
					le.nodeMutex.RUnlock()
					le.startElection()
				} else {
					le.nodeMutex.RUnlock()
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

	le.nodeMutex.Lock()
	defer le.nodeMutex.Unlock()

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

func (le *LeadElection) onLeaderAnnouncement(_ context.Context, req *pb.LeaderAnnouncementMessage) (*emptypb.Empty, error) {
	le.logger.Infof("Leader announcement received: %d", req.Uid)
	le.leaderUID = &req.Uid
	if le.onLeaderChangeFunc != nil {
		le.onLeaderChangeFunc(le.leaderUID)
	}

	return &emptypb.Empty{}, nil
}

func (le *LeadElection) startElection() {

}
