package lcr

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"leadelection/pkg/lcr/internal"
	"leadelection/pkg/lcr/internal/client"
	pb "leadelection/pkg/lcr/internal/rpc"
	"leadelection/pkg/lcr/internal/server"
	"sync"
	"time"

	"google.golang.org/grpc"
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
		server:          s,
		isTLS:           false,
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
		isTLS:           true,
		server:          s,
		orderedNodeUIDs: internal.OrderedUIDList{uid},
		nodes:           map[uint64]*client.Client{uid: nil},
	}

	s.OnNotifyTermination(le.onTermination)
	s.OnMessage(le.onMessage)

	return le, nil
}

// Start starts the cluster.
func (le *LeadElection) Start(delay time.Duration, checkEvery time.Duration) error {
	err := le.server.Start()
	if err == nil {
		time.Sleep(delay)
		if le.leader == nil {
			le.startElection()
		}

		go func() {
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
					ctx, cancel := context.WithTimeout(context.Background(), checkEvery)
					le.nodesMutex.RLock()
					if err := le.nodes[*le.leader].Ping(ctx); err != nil {
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

	return err
}

// Stop stops the cluster.
func (le *LeadElection) Stop() {
	le.nodesMutex.Lock()
	defer le.nodesMutex.Unlock()

	le.server.Close()
	close(le.stop)

	for _, cl := range le.nodes {
		cl.Close()
	}

	time.Sleep(100 * time.Millisecond)
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

// IsLeader returns if the current node is the leader.
func (le *LeadElection) IsLeader() bool {
	return le.leader != nil && *le.leader == le.uid
}

// GetLeaderSync returns the current leader's UID.
// Might be nil if no leader known yet.
func (le *LeadElection) GetLeader() *uint64 {
	return le.leader
}

func (le *LeadElection) onMessage(ctx context.Context, req *pb.LCRMessage) (*pb.LCRResponse, error) {
	if req.Uid == le.uid {
		le.leader = &le.uid
		if resp := le.sendTermination(ctx, req); resp != nil && resp.Status == pb.Status_DISCARDED {
			le.leader = nil
			le.performElectionMutex.Unlock()

			return &pb.LCRResponse{Status: pb.Status_DISCARDED}, nil
		}

		return &pb.LCRResponse{Status: pb.Status_RECEIVED}, nil
	}

	if req.Uid > le.uid {
		le.leader = &req.Uid
		if resp := le.sendMessage(ctx, req); resp != nil && resp.Status == pb.Status_DISCARDED {
			le.leader = nil
			le.performElectionMutex.Unlock()

			return &pb.LCRResponse{Status: pb.Status_DISCARDED}, nil
		}
	}

	// req.Uid < le.uid
	return &pb.LCRResponse{Status: pb.Status_DISCARDED}, nil
}

func (le *LeadElection) onTermination(ctx context.Context, req *pb.LCRMessage) (*pb.LCRResponse, error) {
	if req.Uid == le.uid {
		le.leader = &req.Uid
		le.performElectionMutex.Unlock()

		return &pb.LCRResponse{Status: pb.Status_RECEIVED}, nil
	}

	go func() {
		le.leader = &req.Uid
		le.nodesMutex.RLock()
		if len(le.orderedNodeUIDs) > 1 {
			le.nodesMutex.RUnlock()
			le.sendTermination(ctx, req)
		} else {
			le.nodesMutex.RUnlock()
		}

		le.performElectionMutex.Unlock()
	}()

	return &pb.LCRResponse{Status: pb.Status_RECEIVED}, nil
}

func (le *LeadElection) startElection() {
	if ok := le.performElectionMutex.TryLock(); !ok {
		return
	}

	le.leader = nil
	if resp := le.sendMessage(context.Background(), &pb.LCRMessage{Uid: le.uid}); resp != nil && resp.Status == pb.Status_DISCARDED {
		le.performElectionMutex.Unlock()
	}
}

func (le *LeadElection) sendTermination(ctx context.Context, req *pb.LCRMessage) *pb.LCRResponse {
	le.nodesMutex.Lock()
	defer le.nodesMutex.Unlock()

	index := le.neighborNodeIndex
	sendTo := le.orderedNodeUIDs.GetNext(index)

	for {
		node := le.nodes[sendTo]
		if node == nil {
			// Reached himself, no one else left
			return nil
		}

		if resp, err := node.NotifyTermination(ctx, req); err != nil {
			index++
			sendTo = le.orderedNodeUIDs.GetNext(index)
		} else {
			return resp
		}
	}
}

func (le *LeadElection) sendMessage(ctx context.Context, req *pb.LCRMessage) *pb.LCRResponse {
	le.nodesMutex.Lock()
	defer le.nodesMutex.Unlock()

	index := le.neighborNodeIndex
	sendTo := le.orderedNodeUIDs.GetNext(index)

	for {
		node := le.nodes[sendTo]
		if node == nil {
			// Reached himself, no one else left
			return nil
		}

		if resp, err := node.Message(ctx, req); err != nil {
			index++
			sendTo = le.orderedNodeUIDs.GetNext(index)
		} else {
			return resp
		}
	}
}

func (le *LeadElection) addAnyNode(uid uint64, cl *client.Client) error {
	le.nodesMutex.Lock()
	defer le.nodesMutex.Unlock()

	if _, ok := le.nodes[uid]; ok {
		return ErrNodeAlreadyExists
	}

	le.nodes[uid] = cl
	le.orderedNodeUIDs.AddOrdered(uid)
	if neighborIndex := le.orderedNodeUIDs.FindNeighbor(le.uid); neighborIndex != nil {
		le.neighborNodeIndex = *neighborIndex
	} else {
		delete(le.nodes, uid)
		panic("could not find self in ordered list, this is an internal bug, please open an issue at https://github.com/StevenCyb/go_lead_election/issues.")
	}

	return nil
}
