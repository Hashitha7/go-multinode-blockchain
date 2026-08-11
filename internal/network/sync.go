package network

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/blockchain2/internal/blockchain"
)

// ChainSync handles chain synchronisation with peers.
type ChainSync struct {
	chain       *blockchain.Chain
	peerManager *PeerManager
	httpClient  *http.Client
}

// NewChainSync creates a new chain synchroniser.
func NewChainSync(chain *blockchain.Chain, peerManager *PeerManager) *ChainSync {
	return &ChainSync{
		chain:       chain,
		peerManager: peerManager,
		httpClient: &http.Client{
			Timeout: 30000000000, // 30 seconds in nanoseconds
		},
	}
}

// PeerChainInfo represents the chain status of a peer.
type PeerChainInfo struct {
	Height   uint64 `json:"height"`
	HeadHash string `json:"head_hash"`
}

// GetPeerChainInfo queries a peer's chain height and head hash.
func (cs *ChainSync) GetPeerChainInfo(peerAddr string) (*PeerChainInfo, error) {
	url := fmt.Sprintf("http://%s/chain/height", peerAddr)
	resp, err := cs.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to peer %s: %w", peerAddr, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("peer %s returned status %d", peerAddr, resp.StatusCode)
	}

	var info PeerChainInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode peer chain info: %w", err)
	}

	return &info, nil
}

// GetPeerBlocks downloads blocks from a peer starting from a given height.
func (cs *ChainSync) GetPeerBlocks(peerAddr string, fromHeight uint64) ([]*blockchain.Block, error) {
	url := fmt.Sprintf("http://%s/chain/blocks?from=%d", peerAddr, fromHeight)
	resp, err := cs.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get blocks from peer %s: %w", peerAddr, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("peer %s returned status %d: %s", peerAddr, resp.StatusCode, string(body))
	}

	var blocks []*blockchain.Block
	if err := json.NewDecoder(resp.Body).Decode(&blocks); err != nil {
		return nil, fmt.Errorf("failed to decode blocks: %w", err)
	}

	return blocks, nil
}

// GetPeerFullChain downloads the full chain from a peer.
func (cs *ChainSync) GetPeerFullChain(peerAddr string) ([]*blockchain.Block, error) {
	url := fmt.Sprintf("http://%s/chain", peerAddr)
	resp, err := cs.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain from peer %s: %w", peerAddr, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("peer %s returned status %d", peerAddr, resp.StatusCode)
	}

	var blocks []*blockchain.Block
	if err := json.NewDecoder(resp.Body).Decode(&blocks); err != nil {
		return nil, fmt.Errorf("failed to decode chain: %w", err)
	}

	return blocks, nil
}

// SyncWithPeer synchronises our chain with a specific peer.
// Returns the number of new blocks added and any orphaned transactions from reorg.
func (cs *ChainSync) SyncWithPeer(peerAddr string) (int, []*blockchain.Transaction, error) {
	// Get peer's chain info
	peerInfo, err := cs.GetPeerChainInfo(peerAddr)
	if err != nil {
		return 0, nil, err
	}

	ourHeight := cs.chain.GetHeight()

	// If peer is not ahead, nothing to do
	if peerInfo.Height <= ourHeight {
		return 0, nil, nil
	}

	log.Printf("[SYNC] Peer %s is ahead (peer: %d, us: %d), syncing...", peerAddr, peerInfo.Height, ourHeight)

	// Try to get blocks from our current height
	newBlocks, err := cs.GetPeerBlocks(peerAddr, ourHeight)
	if err != nil {
		return 0, nil, err
	}

	if len(newBlocks) == 0 {
		return 0, nil, nil
	}

	// Check if the first new block extends our chain directly
	ourHead := cs.chain.GetLatestBlock()
	firstNew := newBlocks[0]

	if firstNew.PrevHash == ourHead.Hash && firstNew.Index == ourHead.Index+1 {
		// Simple append — blocks extend our chain
		addedCount := 0
		for _, block := range newBlocks {
			if err := cs.chain.AddBlock(block); err != nil {
				log.Printf("[SYNC] Failed to add block #%d: %v", block.Index, err)
				break
			}
			addedCount++
		}
		log.Printf("[SYNC] Added %d blocks from peer %s (new height: %d)", addedCount, peerAddr, cs.chain.GetHeight())
		return addedCount, nil, nil
	}

	// Need a full reorg — get the entire chain from peer
	log.Printf("[SYNC] Fork detected with peer %s, fetching full chain for reorg...", peerAddr)
	fullChain, err := cs.GetPeerFullChain(peerAddr)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to get full chain for reorg: %w", err)
	}

	// Attempt reorganisation
	orphanedTxs, err := cs.chain.ReorganiseTo(fullChain)
	if err != nil {
		return 0, nil, fmt.Errorf("reorg failed: %w", err)
	}

	newHeight := cs.chain.GetHeight()
	blocksAdded := int(newHeight - ourHeight)
	log.Printf("[SYNC] Reorg complete. New height: %d, orphaned txs: %d", newHeight, len(orphanedTxs))

	return blocksAdded, orphanedTxs, nil
}

// SyncWithAllPeers attempts to sync with all known peers.
func (cs *ChainSync) SyncWithAllPeers() (int, []*blockchain.Transaction) {
	peers := cs.peerManager.GetPeers()
	totalAdded := 0
	var allOrphanedTxs []*blockchain.Transaction

	for _, peer := range peers {
		added, orphaned, err := cs.SyncWithPeer(peer)
		if err != nil {
			log.Printf("[SYNC] Failed to sync with peer %s: %v", peer, err)
			continue
		}
		totalAdded += added
		allOrphanedTxs = append(allOrphanedTxs, orphaned...)
	}

	return totalAdded, allOrphanedTxs
}
