package lcr

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"leadelection/pkg/lcr/internal"
	"leadelection/pkg/lcr/internal/client"
	pb "leadelection/pkg/lcr/internal/rpc"
	"leadelection/pkg/lcr/internal/server"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var ErrNodeAlreadyExists = errors.New("node already exists")
var ErrNodeNotExists = errors.New("node not exists")
var ErrTimeout = errors.New("timeout")
var ErrTLSNode = errors.New("node is TLS, use AddTLSNode instead")
var ErrNonTLSNode = errors.New("node is not TLS, use AddNode instead")

// LeadElection represents the LeLann-Chang-Roberts (LCR) algorithm for leader election.
type LeadElection struct {
	// Unchanged fields
	uid    uint64
	listen string
	isTLS  bool
	server *server.Server
	// Dynamic fields
	performElectionMutex sync.Mutex
	stop                 chan struct{}
	nodesMutex           sync.RWMutex
	nodes                map[uint64]*client.Client
	orderedNodeUIDs      internal.OrderedUIDList
	neighborNodeIndex    int
	leader               *uint64
}

// New creates a new LeadElection instance.
func New(uid uint64, listen string) (*LeadElection, error) {
	s, err := server.New(listen)
	if err != nil {
		return nil, err
	}

	le := &LeadElection{
		uid:             uid,
		listen:          listen,
		server:          s,
		isTLS:           false,
		stop:            make(chan struct{}),
		orderedNodeUIDs: internal.OrderedUIDList{uid},
		nodes:           map[uint64]*client.Client{uid: nil},
	}

	s.OnNotifyTermination(le.onTermination)
	s.OnMessage(le.onMessage)

	return le, nil
}

// NewTLS creates a new LeadElection instance with TLS.
func NewTLS(uid uint64, listen string, cert tls.Certificate) (*LeadElection, error) {
	s, err := server.NewTLS(listen, cert)
	if err != nil {
		return nil, err
	}

	le := &LeadElection{
		uid:             uid,
		listen:          listen,
		server:          s,
		isTLS:           true,
		stop:            make(chan struct{}),
		orderedNodeUIDs: internal.OrderedUIDList{uid},
		nodes:           map[uint64]*client.Client{uid: nil},
	}

	s.OnNotifyTermination(le.onTermination)
	s.OnMessage(le.onMessage)

	return le, nil
}

// MustStart starts the the leader election service or panics if server can not start.
// New leader will be elected if current leader is unknown or unreachable within 5 seconds.
func (le *LeadElection) MustStart(delay time.Duration, checkEvery time.Duration) {
	le.log("Start leader election service")
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
		le.log("Start watching leader")
		for {
			select {
			case <-le.stop:
				return
			default:
				if le.leader == nil {
					le.startElection()
					time.Sleep(checkEvery)
					continue
				}
				if *le.leader == le.uid {
					time.Sleep(checkEvery)
					continue
				}
				le.nodesMutex.RLock()
				if le.nodes[*le.leader] == nil {
					continue
				}
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
				if err := le.nodes[*le.leader].Ping(ctx); err != nil {
					le.log("Leader %d is not reachable, start election", *le.leader)
					le.nodesMutex.RUnlock()
					le.startElection()
				} else {
					le.nodesMutex.RUnlock()
				}
				cancel()
				time.Sleep(checkEvery)
			}
		}
	}()
}

// Stop stops the cluster.
func (le *LeadElection) Stop() {
	le.log("Stop leader election service")

	close(le.stop)
	le.server.Close()

	le.nodesMutex.Lock()
	defer func() {
		le.nodesMutex.Unlock()
	}()

	for _, cl := range le.nodes {
		if cl != nil {
			cl.Close()
		}
	}

	le.stop = make(chan struct{})
	le.orderedNodeUIDs = internal.OrderedUIDList{le.uid}
	le.nodes = map[uint64]*client.Client{le.uid: nil}
	le.performElectionMutex = sync.Mutex{}
}

// AddNode adds a new node to the cluster.
func (le *LeadElection) AddNode(uid uint64, addr string, opts ...grpc.DialOption) error {
	if le.isTLS {
		return ErrNonTLSNode
	}

	cl, err := client.New(addr, opts...)
	if err != nil {
		return err
	}

	return le.addAnyNode(uid, cl)
}

// AddTLSNode adds a new node to the cluster with TLS.
func (le *LeadElection) AddTLSNode(uid uint64, addr string, certPool *x509.CertPool, opts ...grpc.DialOption) error {
	if !le.isTLS {
		return ErrTLSNode
	}

	cl, err := client.NewTLS(addr, certPool, opts...)
	if err != nil {
		return err
	}

	return le.addAnyNode(uid, cl)
}

// RemoveNode removes a node from the cluster.
func (le *LeadElection) RemoveNode(uid uint64) error {
	le.nodesMutex.Lock()
	defer func() {
		le.nodesMutex.Unlock()
	}()

	if _, ok := le.nodes[uid]; !ok {
		return ErrNodeNotExists
	}

	delete(le.nodes, uid)
	le.orderedNodeUIDs.RemoveOrdered(uid)
	if neighborIndex := le.orderedNodeUIDs.FindNeighbor(le.uid); neighborIndex != nil {
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

func (le *LeadElection) onMessage(_ context.Context, req *pb.LCRMessage) (*pb.LCRResponse, error) {
	le.log("Got message[%s] for leader %d", req.MessageId, req.Uid)
	if req.Uid == le.uid {
		le.log("Received message[%s] from self within %s, terminate with self as leader", req.MessageId, time.Since(req.StartTime.AsTime()).String())
		if resp := le.sendTermination(req); resp != nil && resp.Status == pb.Status_DISCARDED {
			return &pb.LCRResponse{Status: pb.Status_DISCARDED, MessageId: req.MessageId}, nil
		}

		return &pb.LCRResponse{Status: pb.Status_RECEIVED, MessageId: req.MessageId}, nil
	}

	if req.Uid > le.uid {
		le.log("Propagate message[%s] for leader candidate %d", req.MessageId, req.Uid)
		if resp := le.sendMessage(req); resp != nil && resp.Status == pb.Status_DISCARDED {
			le.log("Discard message[%s] for leader due neighbor discarded candidate %d", req.MessageId, req.Uid)

			return &pb.LCRResponse{Status: pb.Status_DISCARDED, MessageId: req.MessageId}, nil
		}

		return &pb.LCRResponse{Status: pb.Status_RECEIVED, MessageId: req.MessageId}, nil
	}

	// req.Uid < le.uid
	le.log("Discard message[%s] for leader %d", req.MessageId, req.Uid)

	return &pb.LCRResponse{Status: pb.Status_DISCARDED, MessageId: req.MessageId}, nil
}

func (le *LeadElection) onTermination(_ context.Context, req *pb.LCRMessage) (*pb.LCRResponse, error) {
	le.log("Got termination message[%s] for leader %d", req.MessageId, req.Uid)
	if req.Uid == le.uid {
		le.log("Terminate by message[%s] with self (%d) as leader after %s", req.MessageId, le.uid, time.Since(req.StartTime.AsTime()).String())
		le.leader = &req.Uid

		return &pb.LCRResponse{Status: pb.Status_RECEIVED, MessageId: req.MessageId}, nil
	}

	le.log("Propagate termination message[%s] for leader %d", req.MessageId, req.Uid)

	le.leader = &req.Uid
	le.log("New leader is %d", req.Uid)

	if resp := le.sendTermination(req); resp != nil && resp.Status == pb.Status_DISCARDED {
		return &pb.LCRResponse{Status: pb.Status_DISCARDED, MessageId: req.MessageId}, nil
	}

	return &pb.LCRResponse{Status: pb.Status_RECEIVED, MessageId: req.MessageId}, nil
}

func (le *LeadElection) startElection() {
	if ok := le.performElectionMutex.TryLock(); !ok {
		return
	}

	req := &pb.LCRMessage{Uid: le.uid, StartTime: timestamppb.New(time.Now()), MessageId: uuid.New().String()}
	le.log("Starting election with leader %v on message[%s]  current leader = nil", le.uid, req.MessageId)
	if resp := le.sendMessage(req); resp != nil && resp.Status == pb.Status_DISCARDED {
		le.log("Election discarded for leader %v on message[%s] after %s", le.uid, req.MessageId, time.Since(req.StartTime.AsTime()).String())
		le.performElectionMutex.Unlock()

		return
	}

	le.leader = &le.uid
	le.performElectionMutex.Unlock()
	le.log("Election accepted for self leader %v on message[%s] after %s", le.uid, req.MessageId, time.Since(req.StartTime.AsTime()).String())
}

func (le *LeadElection) sendTermination(req *pb.LCRMessage) *pb.LCRResponse {
	le.nodesMutex.RLock()
	defer func() {
		le.nodesMutex.RUnlock()
	}()

	index := le.neighborNodeIndex
	sendTo := le.orderedNodeUIDs.GetNext(index)

	for {
		node := le.nodes[sendTo]
		if node == nil {
			le.log("Reached himself on message[%s], no one else left", req.MessageId)
			return nil
		}

		le.log("Send termination message[%s] with leader %d to %d", req.MessageId, req.Uid, sendTo)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
		if resp, err := node.NotifyTermination(ctx, req); err != nil {
			le.log("Error termination sending message[%s] to %d: %v", req.MessageId, sendTo, err)
			cancel()
			index++
			sendTo = le.orderedNodeUIDs.GetNext(index)
		} else {
			cancel()
			return resp
		}
	}
}

func (le *LeadElection) sendMessage(req *pb.LCRMessage) *pb.LCRResponse {
	le.nodesMutex.RLock()
	defer func() {
		le.nodesMutex.RUnlock()
	}()

	le.leader = nil
	index := le.neighborNodeIndex
	sendTo := le.orderedNodeUIDs.GetNext(index)

	for {
		node := le.nodes[sendTo]
		if node == nil {
			le.log("Reached himself on termination message[%s], no one else left", req.MessageId)
			return nil
		}

		le.log("Send message[%s] with leader %d to %d", req.MessageId, req.Uid, sendTo)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
		if resp, err := node.Message(ctx, req); err != nil {
			le.log("Error sending message[%s] to %d: %v", req.MessageId, sendTo, err)
			cancel()
			index++
			sendTo = le.orderedNodeUIDs.GetNext(index)
		} else {
			cancel()
			return resp
		}
	}
}

func (le *LeadElection) addAnyNode(uid uint64, cl *client.Client) error {
	le.nodesMutex.Lock()
	defer func() {
		le.nodesMutex.Unlock()
	}()

	if _, ok := le.nodes[uid]; ok {
		return ErrNodeAlreadyExists
	}

	le.nodes[uid] = cl
	le.orderedNodeUIDs.AddOrdered(uid)
	if neighborIndex := le.orderedNodeUIDs.FindNeighbor(le.uid); neighborIndex != nil {
		le.neighborNodeIndex = *neighborIndex
	} else {
		panic("could not find self in ordered list, this is an internal bug, please open an issue at https://github.com/StevenCyb/go_lead_election/issues.")
	}
	le.log("Added node %d, own idex %d, neighbor index %d", uid, le.orderedNodeUIDs.GetIndexFor(le.uid), le.neighborNodeIndex)

	return nil
}

func (le *LeadElection) log(format string, a ...any) {
	fmt.Printf("[%s.%03d] [%d] %s\n", time.Now().Format("2006-01-02T15:04:05"), time.Now().Nanosecond()/1e6, le.uid, fmt.Sprintf(format, a...))
}
