package network

import (
	"log"
	"sync"
)

// PeerManager manages the list of known peers.
type PeerManager struct {
	peers map[string]bool // peer address -> active
	mu    sync.RWMutex
}

// NewPeerManager creates a new peer manager.
func NewPeerManager() *PeerManager {
	return &PeerManager{
		peers: make(map[string]bool),
	}
}

// AddPeer adds a peer to the known list.
func (pm *PeerManager) AddPeer(address string) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.peers[address]; exists {
		return false
	}

	pm.peers[address] = true
	log.Printf("[PEERS] Added peer: %s", address)
	return true
}

// RemovePeer removes a peer from the known list.
func (pm *PeerManager) RemovePeer(address string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	delete(pm.peers, address)
	log.Printf("[PEERS] Removed peer: %s", address)
}

// GetPeers returns a list of all known peer addresses.
func (pm *PeerManager) GetPeers() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make([]string, 0, len(pm.peers))
	for addr := range pm.peers {
		result = append(result, addr)
	}
	return result
}

// PeerCount returns the number of known peers.
func (pm *PeerManager) PeerCount() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.peers)
}

// ExchangePeers merges a list of remote peers into our peer list.
// Returns the list of newly added peers.
func (pm *PeerManager) ExchangePeers(remotePeers []string, selfAddr string) []string {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	newPeers := make([]string, 0)
	for _, addr := range remotePeers {
		// Don't add ourselves
		if addr == selfAddr {
			continue
		}
		if _, exists := pm.peers[addr]; !exists {
			pm.peers[addr] = true
			newPeers = append(newPeers, addr)
			log.Printf("[PEERS] Discovered peer via exchange: %s", addr)
		}
	}
	return newPeers
}

// HasPeer checks if a peer is known.
func (pm *PeerManager) HasPeer(address string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	_, exists := pm.peers[address]
	return exists
}
