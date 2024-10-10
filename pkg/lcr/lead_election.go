package lcr

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"leadelection/pkg/lcr/internal/client"
	pb "leadelection/pkg/lcr/internal/rpc"
	"leadelection/pkg/lcr/internal/server"
	"sync"
	"time"
)

var ErrNodeAlreadyExists = errors.New("node already exists")
var ErrNodeNotExists = errors.New("node not exists")
var ErrTimeout = errors.New("timeout")
var ErrTLSNode = errors.New("node is TLS, use AddTLSNode instead")
var ErrNonTLSNode = errors.New("node is not TLS, use AddNode instead")

// Define SortedList as a slice of uint64.
type orderedUIDList []uint64

// AddOrdered adds a new value to the SortedList, keeping it sorted.
func (o *orderedUIDList) AddOrdered(uid uint64) {
	i := 0

	for i < len(*o) && (*o)[i] < uid {
		i++
	}

	*o = append((*o)[:i], append([]uint64{uid}, (*o)[i:]...)...)
}

// GetNext returns the next value in the SortedList, wrapping around.
func (o orderedUIDList) GetNext(index int) uint64 {
	return o[index%len(o)]
}

// FindNeighbor finds the index of the next value to the given number.
func (o orderedUIDList) FindNeighbor(number uint64) *int {
	for i, v := range o {
		if v == number {
			nextIndex := (i + 1) % len(o)

			return &nextIndex
		}
	}

	return nil
}

// LeadElection represents the LeLann-Chang-Roberts (LCR) algorithm for leader election.
type LeadElection struct {
	uid             uint64
	isTLS           bool
	performElection bool
	orderedUIDs     orderedUIDList
	nodes           map[uint64]*client.Client
	neighborIndex   int
	server          *server.Server
	leader          *uint64
	mutex           sync.Mutex // BUG this will cause race condition on election etc.
}

// New creates a new LeadElection instance.
func New(uid uint64, listen string) (*LeadElection, error) {
	s, err := server.New(listen)
	if err != nil {
		return nil, err
	}

	le := &LeadElection{
		uid:         uid,
		server:      s,
		isTLS:       false,
		orderedUIDs: orderedUIDList{uid},
		nodes:       map[uint64]*client.Client{uid: nil},
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
		uid:         uid,
		isTLS:       true,
		server:      s,
		orderedUIDs: orderedUIDList{uid},
		nodes:       map[uint64]*client.Client{uid: nil},
	}

	s.OnNotifyTermination(le.onTermination)
	s.OnMessage(le.onMessage)

	return le, nil
}

// Start starts the cluster.
func (le *LeadElection) Start(delay time.Duration) error {
	le.mutex.Lock()
	defer le.mutex.Unlock()

	err := le.server.Start()
	if err == nil {
		time.Sleep(delay)
		le.startElectionUnsecured()
	}

	return err
}

// Stop stops the cluster.
func (le *LeadElection) Stop() {
	le.mutex.Lock()
	defer le.mutex.Unlock()

	le.server.Close()

	for _, cl := range le.nodes {
		cl.Close()
	}

	le.orderedUIDs = orderedUIDList{le.uid}
	le.nodes = map[uint64]*client.Client{le.uid: nil}
}

// GetLeaderSync returns the current leader's UID.
// If the current node is not the leader, it pings the leader to verify its availability.
// If the leader is unreachable, it starts a new election.
// Returns error if the leader is not found within 5 seconds.
func (le *LeadElection) GetLeaderSync() (uint64, error) {
	le.mutex.Lock()
	defer le.mutex.Unlock()

	return le.getLeaderSyncUnsecured()
}

// IsLeader returns synchronously if the current node is the leader.
// If the leader is someone else, to verify, this method pings the leader and starts a new election if it is unreachable.
// Returns error if the leader is not found within 5 seconds.
func (le *LeadElection) IsLeaderSync() (bool, error) {
	le.mutex.Lock()
	defer le.mutex.Unlock()

	if leader, err := le.getLeaderSyncUnsecured(); err != nil {
		return false, err
	} else {
		return leader == le.uid, nil
	}
}

// getLeaderSyncUnsecured is a helper method for GetLeaderSync and IsLeaderSync but without mutex.
func (le *LeadElection) getLeaderSyncUnsecured() (uint64, error) {
	if le.leader != nil && *le.leader == le.uid {
		return *le.leader, nil
	}

	if le.leader == nil {
		le.startElectionUnsecured()
	} else if err := le.nodes[*le.leader].Ping(context.Background()); err != nil {
		le.leader = nil
		le.startElectionUnsecured()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return 0, ErrTimeout
		default:
			if le.leader != nil {
				return *le.leader, nil
			}
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func (le *LeadElection) startElectionUnsecured() {
	if le.performElection {
		return
	}

	le.performElection = true
	defer func() {
		le.performElection = false
	}()

	req := &pb.LCRMessage{Uid: le.uid}
	le.sendMessageUnsecured(context.Background(), req)
}

// AddNode adds a new node to the cluster.
func (le *LeadElection) AddNode(uid uint64, addr string) error {
	if le.isTLS {
		return ErrNonTLSNode
	}

	le.mutex.Lock()
	defer le.mutex.Unlock()

	cl, err := client.New(addr)
	if err != nil {
		return err
	}

	return le.addAnyNodeUnsecured(uid, cl)
}

// AddTLSNode adds a new node to the cluster with TLS.
func (le *LeadElection) AddTLSNode(uid uint64, addr string, certPool *x509.CertPool) error {
	if !le.isTLS {
		return ErrTLSNode
	}

	le.mutex.Lock()
	defer le.mutex.Unlock()

	cl, err := client.NewTLS(addr, certPool)
	if err != nil {
		return err
	}

	return le.addAnyNodeUnsecured(uid, cl)
}

func (le *LeadElection) addAnyNodeUnsecured(uid uint64, cl *client.Client) error {
	if _, ok := le.nodes[uid]; ok {
		return ErrNodeAlreadyExists
	}

	le.nodes[uid] = cl
	le.orderedUIDs.AddOrdered(uid)
	if neighborIndex := le.orderedUIDs.FindNeighbor(le.uid); neighborIndex != nil {
		le.neighborIndex = *neighborIndex
	} else {
		delete(le.nodes, uid)
		panic("could not find self in ordered list, this is an internal bug, please open an issue at https://github.com/StevenCyb/go_lead_election/issues.")
	}

	return nil
}

// RemoveNode removes a node from the cluster.
func (le *LeadElection) RemoveNode(id uint64) error {
	cl, ok := le.nodes[id]
	if !ok {
		return ErrNodeNotExists
	}

	le.mutex.Lock()
	defer le.mutex.Unlock()

	cl.Close()
	delete(le.nodes, id)

	if neighborIndex := le.orderedUIDs.FindNeighbor(le.uid); neighborIndex != nil {
		le.neighborIndex = *neighborIndex
	} else {
		panic("could not find self in ordered list, this is an internal bug, please open an issue at https://github.com/StevenCyb/go_lead_election/issues.")
	}

	return nil
}

func (le *LeadElection) onTermination(ctx context.Context, req *pb.LCRMessage) (*pb.LCRResponse, error) {
	le.mutex.Lock()
	defer le.mutex.Unlock()

	if req.Uid == le.uid {
		le.leader = &req.Uid
		return &pb.LCRResponse{Status: pb.Status_RECEIVED}, nil
	}

	if req.Uid < le.uid {
		return &pb.LCRResponse{Status: pb.Status_DISCARDED}, nil
	}

	if len(le.orderedUIDs) > 1 {
		le.leader = &req.Uid
		le.sendTerminationUnsecured(ctx, req)
	}

	return &pb.LCRResponse{Status: pb.Status_RECEIVED}, nil
}

func (le *LeadElection) onMessage(ctx context.Context, req *pb.LCRMessage) (*pb.LCRResponse, error) {
	le.mutex.Lock()
	defer le.mutex.Unlock()

	if req.Uid == le.uid {
		le.leader = &le.uid
		le.sendTerminationUnsecured(ctx, req)

		return &pb.LCRResponse{Status: pb.Status_RECEIVED}, nil
	}

	if req.Uid > le.uid {
		le.leader = &req.Uid
		le.sendMessageUnsecured(ctx, req)
	}

	// req.Uid < le.uid
	return &pb.LCRResponse{Status: pb.Status_DISCARDED}, nil
}

func (le *LeadElection) sendTerminationUnsecured(ctx context.Context, req *pb.LCRMessage) {
	index := le.neighborIndex
	sendTo := le.orderedUIDs.GetNext(index)

	for {
		node := le.nodes[sendTo]
		if node == nil {
			// Reached himself, no one else left
			break
		}

		if resp, err := node.NotifyTermination(ctx, req); err != nil {
			index++
			sendTo = le.orderedUIDs.GetNext(index)
		} else if resp.Status == pb.Status_RECEIVED {
			break
		}
	}
}

func (le *LeadElection) sendMessageUnsecured(ctx context.Context, req *pb.LCRMessage) {
	index := le.neighborIndex
	sendTo := le.orderedUIDs.GetNext(index)

	for {
		node := le.nodes[sendTo]
		if node == nil {
			// Reached himself, no one else left
			break
		}

		if resp, err := node.Message(ctx, req); err != nil {
			index++
			sendTo = le.orderedUIDs.GetNext(index)
		} else if resp.Status == pb.Status_RECEIVED {
			break
		}
	}
}
