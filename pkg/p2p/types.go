package p2p

import (
	"context"
	"time"
)

// NetworkManager interface for P2P networking
type NetworkManager interface {
	Start(ctx context.Context) error
	Stop() error
	GetPeers() []string
	SendMessage(peerID string, data []byte) error
	BroadcastMessage(data []byte) error
}

// BasicPeerInfo represents basic information about a peer
type BasicPeerInfo struct {
	ID       string    `json:"id"`
	Address  string    `json:"address"`
	LastSeen time.Time `json:"last_seen"`
	Status   string    `json:"status"`
}

// NetworkStats represents P2P network statistics
type NetworkStats struct {
	PeerCount      int       `json:"peer_count"`
	MessagesSent   int64     `json:"messages_sent"`
	MessagesRecv   int64     `json:"messages_recv"`
	BytesSent      int64     `json:"bytes_sent"`
	BytesReceived  int64     `json:"bytes_received"`
	LastUpdate     time.Time `json:"last_update"`
}

// MockNetworkManager is a simple mock implementation
type MockNetworkManager struct {
	peers []string
}

func NewMockNetworkManager() *MockNetworkManager {
	return &MockNetworkManager{
		peers: []string{},
	}
}

func (m *MockNetworkManager) Start(ctx context.Context) error {
	return nil
}

func (m *MockNetworkManager) Stop() error {
	return nil
}

func (m *MockNetworkManager) GetPeers() []string {
	return m.peers
}

func (m *MockNetworkManager) SendMessage(peerID string, data []byte) error {
	return nil
}

func (m *MockNetworkManager) BroadcastMessage(data []byte) error {
	return nil
}