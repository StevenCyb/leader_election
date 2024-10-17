package cr

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/StevenCyb/leader_election/pkg/cr/internal/client"
	pb "github.com/StevenCyb/leader_election/pkg/cr/internal/rpc"
	"github.com/StevenCyb/leader_election/pkg/cr/internal/server"
	"github.com/StevenCyb/leader_election/pkg/internal"
	"github.com/StevenCyb/leader_election/pkg/log"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var ErrNodeAlreadyExists = errors.New("node already exists")
var ErrNodeNotExists = errors.New("node not exists")
var ErrTimeout = errors.New("timeout")
var ErrTLSNode = errors.New("node is TLS, use AddTLSNode instead")
var ErrNonTLSNode = errors.New("node is not TLS, use AddNode instead")

// OnLeaderChangeFunc is a callback function that is called when the leader changes.
type OnLeaderChangeFunc func(leader *uint64)

type Stage byte

const (
	Idle Stage = iota
	Election
	LeaderAnnouncement
)

// LeadElection represents the LeLann-Chang-Roberts (LCR) algorithm for leader election.
type LeadElection struct {
	// Unchanged fields
	uid                uint64
	listen             string
	server             *server.Server
	onLeaderChangeFunc OnLeaderChangeFunc
	logger             log.ILogger
	// Dynamic fields
	stageMutex        sync.Mutex
	stage             Stage
	participant       bool
	stop              chan struct{}
	nodesMutex        sync.RWMutex
	nodes             map[uint64]*client.Client
	orderedNodeUIDs   internal.OrderedList[uint64]
	neighborNodeIndex int
	leader            *uint64
}

// New creates a new LeadElection instance.
func New(uid uint64, listen string, logger log.ILogger, opt ...grpc.ServerOption) (*LeadElection, error) {
	s, err := server.New(listen, opt...)
	if err != nil {
		return nil, err
	}

	le := &LeadElection{
		uid:             uid,
		listen:          listen,
		logger:          logger,
		server:          s,
		stage:           Idle,
		stop:            make(chan struct{}),
		orderedNodeUIDs: internal.OrderedList[uint64]{uid},
		nodes:           map[uint64]*client.Client{uid: nil},
	}

	s.OnElection(le.onElection)
	s.OnElected(le.onElected)

	return le, nil
}

// MustStart starts the the leader election service or panics if server can not start.
// New leader will be elected if current leader is unknown or unreachable within 5 seconds.
func (le *LeadElection) MustStart(delay time.Duration, checkInterval time.Duration) {
	le.logger.Info("Start leader election service")
	go func() {
		err := le.server.Start()
		le.logger.Trace("Server stopped")
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
				if *le.leader == le.uid {
					time.Sleep(checkInterval)
					continue
				}
				le.nodesMutex.RLock()
				if le.nodes[*le.leader] == nil {
					continue
				}
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				if err := le.nodes[*le.leader].Ping(ctx); err != nil {
					le.logger.Infof("Leader %d is not reachable, start election", *le.leader)
					le.nodesMutex.RUnlock()
					le.startElection()
				} else {
					le.nodesMutex.RUnlock()
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

	le.nodesMutex.Lock()
	defer le.nodesMutex.Unlock()

	for _, cl := range le.nodes {
		if cl != nil {
			cl.Close()
		}
	}

	le.stage = Idle
	le.participant = false
	le.stop = make(chan struct{})
	le.orderedNodeUIDs = internal.OrderedList[uint64]{le.uid}
	le.nodes = map[uint64]*client.Client{le.uid: nil}
	le.stageMutex = sync.Mutex{}
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

	if _, ok := le.nodes[uid]; ok {
		return ErrNodeAlreadyExists
	}

	cl, err := client.New(addr, opts...)
	if err != nil {
		return err
	}

	le.nodes[uid] = cl
	le.orderedNodeUIDs.AddOrdered(uid)
	if neighborIndex := le.orderedNodeUIDs.GetIndexRightOfValue(le.uid); neighborIndex != nil {
		le.neighborNodeIndex = *neighborIndex
	} else {
		panic("could not find self in ordered list, this is an internal bug, please open an issue at https://github.com/StevenCyb/go_lead_election/issues.")
	}
	le.logger.Tracef("Added node %d, own idex %d, neighbor index %d", uid, le.orderedNodeUIDs.GetIndexFor(le.uid), le.neighborNodeIndex)

	return nil
}

// RemoveNode removes a node from the cluster.
func (le *LeadElection) RemoveNode(uid uint64) error {
	le.nodesMutex.Lock()
	defer le.nodesMutex.Unlock()

	if _, ok := le.nodes[uid]; !ok {
		return ErrNodeNotExists
	}

	delete(le.nodes, uid)
	le.orderedNodeUIDs.RemoveOrdered(uid)
	if neighborIndex := le.orderedNodeUIDs.GetIndexLeftOfValue(le.uid); neighborIndex != nil {
		le.neighborNodeIndex = *neighborIndex
	} else {
		panic("could not find self in ordered list, this is an internal bug, please open an issue at https://github.com/StevenCyb/go_lead_election/issues.")
	}

	return nil
}

// IsLeader returns if the current node is the leader.
func (le *LeadElection) IsLeader() bool {
	return le.leader != nil && *le.leader == le.uid
}

// GetLeaderSync returns the current leader's UID.
// Might be nil if no leader known yet.
func (le *LeadElection) GetLeader() *uint64 {
	return le.leader
}

func (le *LeadElection) startElection() {
	if !le.stageMutex.TryLock() {
		return
	}
	defer le.stageMutex.Unlock()

	if le.stage != Idle || le.participant {
		return
	}

	le.logger.Info("Start election")

	le.participant = true
	le.stage = Election
	msg := &pb.Message{Uid: le.uid}

	go le.sendElection(msg)
}

func (le *LeadElection) onElection(ctx context.Context, req *pb.Message) (*emptypb.Empty, error) {
	le.logger.Debugf("Received election message with leader %d", req.Uid)
	le.stageMutex.Lock()
	defer le.stageMutex.Unlock()

	if req.Uid > le.uid {
		le.participant = true
		le.logger.Tracef("Forward election message to next node with leader %d", req.Uid)
		go le.sendElection(req)
	} else if req.Uid < le.uid && !le.participant {
		le.logger.Tracef("Forward election message to next node with replaced leader %d->%d", req.Uid, le.uid)
		le.participant = true
		go le.sendElection(&pb.Message{Uid: le.uid})
	} else if req.Uid == le.uid {
		// own election message
		le.logger.Tracef("Own election message, start leader announcement")
		le.stage = LeaderAnnouncement
		le.leader = &req.Uid
		go le.sendElected(req)
	} else {
		le.logger.Tracef("Ignore election message with leader %d", req.Uid)
	}

	return &emptypb.Empty{}, nil
}

func (le *LeadElection) onElected(ctx context.Context, req *pb.Message) (*emptypb.Empty, error) {
	le.logger.Debugf("Received elected message with leader %d", req.Uid)
	le.stageMutex.Lock()
	defer le.stageMutex.Unlock()

	le.participant = false
	le.stage = Idle

	if req.Uid != le.uid {
		le.logger.Tracef("Forward elected message to next node with leader %d", req.Uid)
		le.leader = &req.Uid
		go le.sendElected(req)
	}

	return &emptypb.Empty{}, nil
}

func (le *LeadElection) sendElection(req *pb.Message) {
	le.nodesMutex.RLock()
	defer le.nodesMutex.RUnlock()

	index := le.neighborNodeIndex
	sendTo := le.orderedNodeUIDs.GetValueForIndexLooped(index)

	for {
		node := le.nodes[sendTo]
		if node == nil {
			break
		}

		le.logger.Tracef("Send election message with leader %d to %d", req.Uid, sendTo)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		if _, err := node.Election(ctx, req); err != nil {
			le.logger.Tracef("Error sending election message to %d: %v", sendTo, err)
			cancel()
			index++
			sendTo = le.orderedNodeUIDs.GetValueForIndexLooped(index)
		} else {
			cancel()
			le.logger.Tracef("Election message sent to %d", sendTo)
			return
		}
	}

	le.logger.Trace("Reached himself on sending election")

	//nolint:errcheck
	go le.onElection(context.Background(), req)
}

func (le *LeadElection) sendElected(req *pb.Message) {
	le.nodesMutex.RLock()
	defer le.nodesMutex.RUnlock()

	index := le.neighborNodeIndex
	sendTo := le.orderedNodeUIDs.GetValueForIndexLooped(index)

	for {
		node := le.nodes[sendTo]
		if node == nil {
			break
		}

		le.logger.Tracef("Send elected announcement message with leader %d to %d", req.Uid, sendTo)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		if _, err := node.Elected(ctx, req); err != nil {
			le.logger.Tracef("Error sending elected announcement message to %d: %v", sendTo, err)
			cancel()
			index++
			sendTo = le.orderedNodeUIDs.GetValueForIndexLooped(index)
		} else {
			cancel()
			le.logger.Tracef("Elected announcement message sent to %d", sendTo)
			return
		}
	}

	le.logger.Trace("Reached himself on sending elected announcement")

	//nolint:errcheck
	go le.onElected(context.Background(), req)
}
