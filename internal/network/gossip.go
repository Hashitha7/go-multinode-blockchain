package network

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/blockchain2/internal/blockchain"
)

// SeenTracker tracks recently seen transaction and block IDs for de-duplication.
type SeenTracker struct {
	seen map[string]time.Time
	mu   sync.Mutex
	ttl  time.Duration
}

// NewSeenTracker creates a new tracker with the given TTL for entries.
func NewSeenTracker(ttl time.Duration) *SeenTracker {
	st := &SeenTracker{
		seen: make(map[string]time.Time),
		ttl:  ttl,
	}
	return st
}

// HasSeen checks if an ID has been seen and is not expired, without marking it.
func (st *SeenTracker) HasSeen(id string) bool {
	st.mu.Lock()
	defer st.mu.Unlock()

	if t, exists := st.seen[id]; exists {
		if time.Since(t) < st.ttl {
			return true // Already seen and not expired
		}
	}
	return false
}

// Mark adds an ID to the seen tracker.
func (st *SeenTracker) Mark(id string) {
	st.mu.Lock()
	defer st.mu.Unlock()
	st.seen[id] = time.Now()
}

// Cleanup removes expired entries.
func (st *SeenTracker) Cleanup() {
	st.mu.Lock()
	defer st.mu.Unlock()

	now := time.Now()
	for id, t := range st.seen {
		if now.Sub(t) > st.ttl {
			delete(st.seen, id)
		}
	}
}

// StartCleanup runs periodic cleanup in the background.
func (st *SeenTracker) StartCleanup(ctx <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(st.ttl)
		defer ticker.Stop()
		for {
			select {
			case <-ctx:
				return
			case <-ticker.C:
				st.Cleanup()
			}
		}
	}()
}

// Gossiper handles broadcasting transactions and blocks to peers.
type Gossiper struct {
	peerManager *PeerManager
	seenTxs     *SeenTracker
	seenBlocks  *SeenTracker
	selfAddr    string
	httpClient  *http.Client
}

// NewGossiper creates a new gossiper.
func NewGossiper(peerManager *PeerManager, selfAddr string) *Gossiper {
	return &Gossiper{
		peerManager: peerManager,
		seenTxs:     NewSeenTracker(5 * time.Minute),
		seenBlocks:  NewSeenTracker(5 * time.Minute),
		selfAddr:    selfAddr,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GossipTransaction forwards a transaction to all peers except the source.
// Returns the number of peers it was successfully sent to.
func (g *Gossiper) GossipTransaction(tx *blockchain.Transaction, excludePeer string) int {
	// Mark as seen so we don't process it again
	g.seenTxs.Mark(tx.ID)

	peers := g.peerManager.GetPeers()
	successCount := 0

	for _, peer := range peers {
		if peer == excludePeer {
			continue
		}

		go func(peerAddr string) {
			err := g.sendTransaction(peerAddr, tx)
			if err != nil {
				log.Printf("[GOSSIP] Failed to send tx %s... to %s: %v", tx.ID[:16], peerAddr, err)
			}
		}(peer)
		successCount++
	}

	if successCount > 0 {
		log.Printf("[GOSSIP] Broadcasting tx %s... to %d peers", tx.ID[:16], successCount)
	}
	return successCount
}

// GossipBlock forwards a block to all peers except the source.
func (g *Gossiper) GossipBlock(block *blockchain.Block, excludePeer string) int {
	// Mark as seen so we don't process it again
	g.seenBlocks.Mark(block.Hash)

	peers := g.peerManager.GetPeers()
	successCount := 0

	for _, peer := range peers {
		if peer == excludePeer {
			continue
		}

		go func(peerAddr string) {
			err := g.sendBlock(peerAddr, block)
			if err != nil {
				log.Printf("[GOSSIP] Failed to send block #%d to %s: %v", block.Index, peerAddr, err)
			}
		}(peer)
		successCount++
	}

	if successCount > 0 {
		log.Printf("[GOSSIP] Broadcasting block #%d to %d peers", block.Index, successCount)
	}
	return successCount
}

// HasSeenTx checks if a transaction has been recently seen.
func (g *Gossiper) HasSeenTx(txID string) bool {
	return g.seenTxs.HasSeen(txID)
}

// MarkTx marks a transaction as seen.
func (g *Gossiper) MarkTx(txID string) {
	g.seenTxs.Mark(txID)
}

// HasSeenBlock checks if a block has been recently seen.
func (g *Gossiper) HasSeenBlock(blockHash string) bool {
	return g.seenBlocks.HasSeen(blockHash)
}

// MarkBlock marks a block as seen.
func (g *Gossiper) MarkBlock(blockHash string) {
	g.seenBlocks.Mark(blockHash)
}

// sendTransaction sends a transaction to a specific peer.
func (g *Gossiper) sendTransaction(peerAddr string, tx *blockchain.Transaction) error {
	data, err := json.Marshal(tx)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %w", err)
	}

	url := fmt.Sprintf("http://%s/tx", peerAddr)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Source-Peer", g.selfAddr)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("peer returned status %d", resp.StatusCode)
	}
	return nil
}

// sendBlock sends a block to a specific peer.
func (g *Gossiper) sendBlock(peerAddr string, block *blockchain.Block) error {
	data, err := json.Marshal(block)
	if err != nil {
		return fmt.Errorf("failed to marshal block: %w", err)
	}

	url := fmt.Sprintf("http://%s/block", peerAddr)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Source-Peer", g.selfAddr)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("peer returned status %d", resp.StatusCode)
	}
	return nil
}

// StartCleanup starts background cleanup for the seen trackers.
func (g *Gossiper) StartCleanup(done <-chan struct{}) {
	g.seenTxs.StartCleanup(done)
	g.seenBlocks.StartCleanup(done)
}
