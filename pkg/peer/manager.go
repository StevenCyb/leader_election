package peer

import (
	"errors"
	"leadelection/pkg/peer/client"
	"leadelection/pkg/peer/server"
)

var ErrPeerNotFound = errors.New("peer not found")
var ErrPeerAlreadyExists = errors.New("peer already exists")

type Manager struct {
	peers  map[string]*client.Client
	server *server.Server
}

func New(server *server.Server) *Manager {
	return &Manager{
		peers:  make(map[string]*client.Client),
		server: server,
	}
}

func (m *Manager) Close() {
	for _, peer := range m.peers {
		peer.Close()
	}

	m.peers = make(map[string]*client.Client)

	m.server.Close()
}

func (m *Manager) AddPeer(peer *client.Client) error {
	address := peer.GetAddress()
	if _, ok := m.peers[address]; ok {
		return ErrPeerAlreadyExists
	}

	m.peers[address] = peer

	return nil
}

func (m *Manager) RemovePeer(address string) error {
	peer, ok := m.peers[address]
	if !ok {
		return ErrPeerNotFound
	}

	peer.Close()
	delete(m.peers, address)

	return nil
}

func (m *Manager) GetPeer(address string) (*client.Client, error) {
	peer, ok := m.peers[address]
	if !ok {
		return nil, ErrPeerNotFound
	}

	return peer, nil
}
