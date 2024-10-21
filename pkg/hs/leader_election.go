package hs

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/StevenCyb/leader_election/pkg/hs/internal/client"
	"github.com/StevenCyb/leader_election/pkg/hs/internal/server"
	"github.com/StevenCyb/leader_election/pkg/internal"
	"github.com/StevenCyb/leader_election/pkg/log"

	pb "github.com/StevenCyb/leader_election/pkg/hs/internal/rpc"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var ErrNodeAlreadyExists = errors.New("node already exists")
var ErrNodeNotExists = errors.New("node not exists")
var ErrNotEnoughNeighbors = errors.New("not enough neighbors")

// OnLeaderChangeFunc is a callback function that is called when the leader changes.
type OnLeaderChangeFunc func(leader *uint64)

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
	leftNeighbors        []internal.Tuple[uint64, *client.Client]
	rightNeighbors       []internal.Tuple[uint64, *client.Client]
	leaderUID            *uint64
	leaderClient         *client.Client
	gotReplyFromLeft     bool
	gotReplyFromRight    bool
}

// New creates a new LeadElection instance.
func New(uid uint64, listen string, logger log.ILogger, opt ...grpc.ServerOption) (*LeadElection, error) {
	s, err := server.New(listen, opt...)
	if err != nil {
		return nil, err
	}

	le := &LeadElection{
		uid:            uid,
		listen:         listen,
		logger:         logger,
		server:         s,
		stop:           make(chan struct{}),
		leftNeighbors:  []internal.Tuple[uint64, *client.Client]{},
		rightNeighbors: []internal.Tuple[uint64, *client.Client]{},
	}

	s.OnProbe(le.onProbe)
	s.OnReply(le.onReply)
	s.OnTerminate(le.onTerminate)

	return le, nil
}

// MustStart starts the leader election service or panics if server can not start.
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
				le.neighborsMutex.RLock()
				if le.leaderUID == nil {
					le.neighborsMutex.RUnlock()
					le.startElection()
					time.Sleep(checkInterval)
					continue
				}
				if *le.leaderUID == le.uid {
					le.neighborsMutex.RUnlock()
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
		cl.Second.Close()
	}
	for _, cl := range le.rightNeighbors {
		cl.Second.Close()
	}

	le.leaderUID = nil
	le.leaderClient = nil
	le.stop = make(chan struct{})
	le.leftNeighbors = []internal.Tuple[uint64, *client.Client]{}
	le.rightNeighbors = []internal.Tuple[uint64, *client.Client]{}
	le.performElectionMutex = sync.Mutex{}
}

// AddLeftNeighborNode adds a left neighbor node to the cluster.
func (le *LeadElection) AddLeftNeighborNode(uid uint64, addr string, opts ...grpc.DialOption) error {
	le.neighborsMutex.Lock()
	defer le.neighborsMutex.Unlock()

	for _, neighbor := range le.leftNeighbors {
		if neighbor.First == uid {
			return ErrNodeAlreadyExists
		}
	}

	cl, err := client.New(addr, opts...)
	if err != nil {
		return err
	}

	le.leftNeighbors = append(le.leftNeighbors, internal.Tuple[uint64, *client.Client]{First: uid, Second: cl})
	le.logger.Tracef("Added left neighbor, now have %d neighbors", len(le.leftNeighbors))

	return nil
}

// AddRightNeighborNode adds a right neighbor node to the cluster.
func (le *LeadElection) AddRightNeighborNode(uid uint64, addr string, opts ...grpc.DialOption) error {
	le.neighborsMutex.Lock()
	defer le.neighborsMutex.Unlock()

	for _, neighbor := range le.rightNeighbors {
		if neighbor.First == uid {
			return ErrNodeAlreadyExists
		}
	}

	cl, err := client.New(addr, opts...)
	if err != nil {
		return err
	}

	le.rightNeighbors = append(le.rightNeighbors, internal.Tuple[uint64, *client.Client]{First: uid, Second: cl})
	le.logger.Tracef("Added right neighbor, now have %d neighbors", len(le.rightNeighbors))

	return nil
}

// OnLeaderChange sets the callback function that is called when the leader changes.
// Leader is nil if no leader known yet or old leader is not reachable and new election is ongoing.
func (le *LeadElection) OnLeaderChange(f OnLeaderChangeFunc) {
	le.onLeaderChangeFunc = f
}

// IsLeader returns if the current node is the leader.
func (le *LeadElection) IsLeader() bool {
	le.neighborsMutex.RLock()
	defer le.neighborsMutex.Unlock()
	return le.leaderUID != nil && *le.leaderUID == le.uid
}

// GetLeaderSync returns the current leader's UID.
// Might be nil if no leader known yet.
func (le *LeadElection) GetLeader() *uint64 {
	return le.leaderUID
}

func (le *LeadElection) startElection() {
	if !le.performElectionMutex.TryLock() {
		return
	}

	le.logger.Infof("Start election")

	le.neighborsMutex.Lock()
	le.leaderUID = nil
	le.leaderClient = nil
	le.neighborsMutex.Unlock()
	le.gotReplyFromLeft = false
	le.gotReplyFromRight = false

	go func() {
		le.logger.Trace("Send probe to left neighbors")
		rightProbeMsg := &pb.ProbeMessage{
			Uid:       le.uid,
			Hops:      0,
			Phase:     0,
			Direction: pb.Direction_RIGHT,
		}
		le.sendProbe(rightProbeMsg, le.rightNeighbors, false)
	}()

	go func() {
		le.logger.Trace("Send probe to right neighbors")
		leftProbeMsg := &pb.ProbeMessage{
			Uid:       le.uid,
			Hops:      0,
			Phase:     0,
			Direction: pb.Direction_RIGHT,
		}
		le.sendProbe(leftProbeMsg, le.leftNeighbors, false)
	}()
}

func (le *LeadElection) onProbe(_ context.Context, req *pb.ProbeMessage) (*emptypb.Empty, error) {
	le.logger.Debugf("Received probe from %d", req.Uid)

	if req.Uid == le.uid {
		le.neighborsMutex.Lock()
		le.leaderUID = &le.uid
		le.neighborsMutex.Unlock()
		le.logger.Debug("Identify self as leader, send termination to neighbors")

		go func() {
			le.logger.Trace("Send termination to left neighbors")
			msg := &pb.TerminateMessage{Uid: *le.leaderUID, Direction: pb.Direction_LEFT}
			le.sendTermination(msg, le.leftNeighbors, false)
		}()
		go func() {
			le.logger.Trace("Send termination to right neighbors")
			msg := &pb.TerminateMessage{Uid: *le.leaderUID, Direction: pb.Direction_RIGHT}
			le.sendTermination(msg, le.rightNeighbors, false)
		}()

		return nil, nil
	}

	if req.Uid > le.uid && req.Hops < internal.Pow(2, req.Phase) {
		le.logger.Debugf("Forward probe from %d", req.Uid)
		req.Hops++
		if req.Direction == pb.Direction_LEFT {
			le.sendProbe(req, le.leftNeighbors, false)
		} else {
			le.sendProbe(req, le.rightNeighbors, false)
		}

		return nil, nil
	}

	if req.Uid > le.uid && req.Hops >= internal.Pow(2, req.Phase) {
		le.logger.Debugf("Send reply to %d", req.Uid)
		reply := &pb.ReplyMessage{
			Uid:   req.Uid,
			Phase: req.Phase,
		}

		if req.Direction == pb.Direction_LEFT {
			le.sendReply(reply, le.leftNeighbors, false)
		} else {
			le.sendReply(reply, le.rightNeighbors, false)
		}

		return nil, nil
	}

	le.logger.Debugf("Got lower leader election from %d, start election", req.Uid)
	le.startElection()

	return nil, nil
}

func (le *LeadElection) onReply(_ context.Context, req *pb.ReplyMessage) (*emptypb.Empty, error) {
	le.logger.Debugf("Received reply from %d", req.Uid)

	if req.Uid != le.uid {
		if req.Direction == pb.Direction_LEFT {
			le.logger.Debugf("Forward reply from %d to left", req.Uid)
			le.sendReply(req, le.leftNeighbors, false)
		} else {
			le.logger.Debugf("Forward reply from %d to right", req.Uid)
			le.sendReply(req, le.rightNeighbors, false)
		}

		return nil, nil
	}

	if req.Direction == pb.Direction_LEFT {
		le.logger.Trace("Got reply from left")
		le.gotReplyFromLeft = true
	} else {
		le.logger.Trace("Got reply from right")
		le.gotReplyFromRight = true
	}

	if le.gotReplyFromLeft && le.gotReplyFromRight {
		le.logger.Debugf("Got reply from both sides, go to next phase %d", req.Phase+1)
		go func() {
			le.logger.Trace("Send probe to right neighbors")
			rightProbeMsg := &pb.ProbeMessage{
				Uid:       le.uid,
				Hops:      0,
				Phase:     req.Phase + 1,
				Direction: pb.Direction_RIGHT,
			}
			le.sendProbe(rightProbeMsg, le.rightNeighbors, false)
		}()

		go func() {
			le.logger.Trace("Send probe to left neighbors")
			leftProbeMsg := &pb.ProbeMessage{
				Uid:       le.uid,
				Hops:      req.Phase + 1,
				Phase:     0,
				Direction: pb.Direction_RIGHT,
			}
			le.sendProbe(leftProbeMsg, le.leftNeighbors, false)
		}()
	}

	return nil, nil
}

func (le *LeadElection) onTerminate(_ context.Context, req *pb.TerminateMessage) (*emptypb.Empty, error) {
	le.performElectionMutex.TryLock()
	le.performElectionMutex.Unlock()

	le.logger.Debugf("Received termination from %d, set as leader", req.Uid)
	if le.leaderUID != nil && *le.leaderUID > req.Uid {
		le.logger.Debugf("Leader %d is higher than %d, ignore termination", *le.leaderUID, req.Uid)
		return nil, nil
	}

	le.neighborsMutex.Lock()
	le.leaderUID = &req.Uid
	le.neighborsMutex.Unlock()

	for _, cl := range append(le.leftNeighbors, le.rightNeighbors...) {
		if cl.First == req.Uid {
			le.leaderClient = cl.Second
			break
		}
	}

	if *le.leaderUID != le.uid {
		if req.Direction == pb.Direction_LEFT {
			le.logger.Trace("Forward termination to right neighbors")
			go func() {
				le.sendTermination(req, le.leftNeighbors, false)
			}()
		} else {
			le.logger.Trace("Forward termination to left neighbors")
			go func() {
				le.sendTermination(req, le.rightNeighbors, false)
			}()
		}

		return nil, nil
	}

	return nil, nil
}

func (le *LeadElection) sendProbe(probeMsg *pb.ProbeMessage, neighbors []internal.Tuple[uint64, *client.Client], reverted bool) {
	sendFkt := func(uid uint64, probeMsg *pb.ProbeMessage, cl *client.Client) bool {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		if err := cl.Probe(ctx, probeMsg); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				le.logger.Warningf("Timeout sending probe to node %d, continuing with next", uid)
				cancel()
				return false
			}
			le.logger.Errorf("Failed to send probe to node %d: %v", uid, err)
		}
		cancel()
		le.logger.Tracef("Sent probe from %d to %d", le.uid, uid)

		return true
	}

	if !reverted {
		for _, cl := range neighbors {
			if sendFkt(cl.First, probeMsg, cl.Second) {
				return
			}
		}
	} else {
		for i := len(neighbors) - 1; i >= 0; i-- {
			cl := neighbors[i]
			if sendFkt(cl.First, probeMsg, cl.Second) {
				return
			}
		}
	}

	if !reverted && probeMsg.Direction == pb.Direction_LEFT && len(le.rightNeighbors) > 0 {
		le.sendProbe(probeMsg, le.rightNeighbors, true)
		return
	} else if !reverted && probeMsg.Direction == pb.Direction_RIGHT && len(le.leftNeighbors) > 0 {
		le.sendProbe(probeMsg, le.leftNeighbors, true)
		return
	}

	le.logger.Trace("Send probe to self, no one left to send")
	if _, err := le.onProbe(context.Background(), probeMsg); err != nil {
		le.logger.Errorf("Failed to send probe to self: %v", err)
	}
}

func (le *LeadElection) sendReply(replyMsg *pb.ReplyMessage, neighbors []internal.Tuple[uint64, *client.Client], reverted bool) {
	sendFkt := func(uid uint64, replyMsg *pb.ReplyMessage, cl *client.Client) bool {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		if err := cl.Reply(ctx, replyMsg); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				le.logger.Warningf("Timeout sending reply to node %d, continuing with next", uid)
				cancel()
				return false
			}
			le.logger.Errorf("Failed to send reply to node %d: %v", uid, err)
		}
		cancel()
		le.logger.Tracef("Sent reply from %d to %d", replyMsg.Uid, uid)

		return true
	}

	if !reverted {
		for _, cl := range neighbors {
			if sendFkt(cl.First, replyMsg, cl.Second) {
				return
			}
		}
	} else {
		for i := len(neighbors) - 1; i >= 0; i-- {
			cl := neighbors[i]
			if sendFkt(cl.First, replyMsg, cl.Second) {
				return
			}
		}
	}

	if !reverted && replyMsg.Direction == pb.Direction_LEFT && len(le.rightNeighbors) > 0 {
		le.sendReply(replyMsg, le.rightNeighbors, true)
		return
	} else if !reverted && replyMsg.Direction == pb.Direction_RIGHT && len(le.leftNeighbors) > 0 {
		le.sendReply(replyMsg, le.leftNeighbors, true)
		return
	}
}

func (le *LeadElection) sendTermination(terminateMsg *pb.TerminateMessage, neighbors []internal.Tuple[uint64, *client.Client], reverted bool) {
	sendFkt := func(uid uint64, terminateMsg *pb.TerminateMessage, cl *client.Client) bool {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		if err := cl.Terminate(ctx, terminateMsg); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				le.logger.Warningf("Timeout sending termination to node %d, continuing with next", uid)
				cancel()
				return false
			}
			le.logger.Errorf("Failed to send termination to node %d: %v", uid, err)
		}
		cancel()
		le.logger.Tracef("Sent termination from %d to %d", terminateMsg.Uid, uid)

		return true
	}

	if !reverted {
		for _, cl := range neighbors {
			if sendFkt(cl.First, terminateMsg, cl.Second) {
				return
			}
		}
	} else {
		for i := len(neighbors) - 1; i >= 0; i-- {
			cl := neighbors[i]
			if sendFkt(cl.First, terminateMsg, cl.Second) {
				return
			}
		}
	}

	if !reverted && terminateMsg.Direction == pb.Direction_LEFT && len(le.rightNeighbors) > 0 {
		le.sendTermination(terminateMsg, le.rightNeighbors, true)
		return
	} else if !reverted && terminateMsg.Direction == pb.Direction_RIGHT && len(le.leftNeighbors) > 0 {
		le.sendTermination(terminateMsg, le.leftNeighbors, true)
		return
	}
}
