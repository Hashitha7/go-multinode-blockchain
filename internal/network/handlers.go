package network

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/blockchain2/internal/blockchain"
	"github.com/blockchain2/internal/mempool"
)

// Handlers holds all HTTP API handler dependencies.
type Handlers struct {
	chain           *blockchain.Chain
	mempool         *mempool.Pool
	peerManager     *PeerManager
	gossiper        *Gossiper
	chainSync       *ChainSync
	selfAddr        string
	onBlockAccepted func() // Called when a new block is added to the chain
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(
	chain *blockchain.Chain,
	pool *mempool.Pool,
	peerManager *PeerManager,
	gossiper *Gossiper,
	chainSync *ChainSync,
	selfAddr string,
	onBlockAccepted func(),
) *Handlers {
	return &Handlers{
		chain:           chain,
		mempool:         pool,
		peerManager:     peerManager,
		gossiper:        gossiper,
		chainSync:       chainSync,
		selfAddr:        selfAddr,
		onBlockAccepted: onBlockAccepted,
	}
}

// HandleSubmitTransaction handles POST /tx — submit or receive a gossiped transaction.
func (h *Handlers) HandleSubmitTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var tx blockchain.Transaction
	if err := json.Unmarshal(body, &tx); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Check de-duplication via gossiper
	if h.gossiper.IsSeenTx(tx.ID) {
		// Already seen — don't forward again (prevents loops)
		writeJSON(w, http.StatusOK, map[string]string{
			"status": "already_seen",
			"tx_id":  tx.ID,
		})
		return
	}

	// Verify the transaction signature (FR-2)
	if err := tx.Verify(); err != nil {
		http.Error(w, fmt.Sprintf("Invalid transaction: %v", err), http.StatusBadRequest)
		return
	}

	// Verify the transaction can be applied against the current ledger (Point 1)
	if !h.chain.GetLedger().CanApplyTransaction(&tx) {
		http.Error(w, "Transaction invalid against current ledger (insufficient funds or replay)", http.StatusBadRequest)
		return
	}

	// Add to mempool (FR-3)
	isNew := h.mempool.Add(&tx)
	if !isNew {
		writeJSON(w, http.StatusOK, map[string]string{
			"status": "duplicate",
			"tx_id":  tx.ID,
		})
		return
	}

	// Gossip to peers, excluding the source (FR-3)
	sourcePeer := r.Header.Get("X-Source-Peer")
	h.gossiper.GossipTransaction(&tx, sourcePeer)

	log.Printf("[API] Accepted transaction %s... (from: %s)", tx.ID[:16], tx.From[:16])
	writeJSON(w, http.StatusCreated, map[string]string{
		"status": "accepted",
		"tx_id":  tx.ID,
	})
}

// HandleReceiveBlock handles POST /block — receive a gossiped block.
func (h *Handlers) HandleReceiveBlock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var block blockchain.Block
	if err := json.Unmarshal(body, &block); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Check de-duplication
	if h.gossiper.IsSeenBlock(block.Hash) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "already_seen"})
		return
	}

	// Try to add the block or determine if we need sync
	added, needsSync, err := h.chain.TryAddOrReorg(&block)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to process block: %v", err), http.StatusBadRequest)
		return
	}

	sourcePeer := r.Header.Get("X-Source-Peer")

	if added {
		// Remove mined transactions from mempool (FR-9)
		h.mempool.RemoveByBlock(&block)

		// Forward to other peers (FR-4)
		h.gossiper.GossipBlock(&block, sourcePeer)

		// Notify node to restart miner
		if h.onBlockAccepted != nil {
			h.onBlockAccepted()
		}

		log.Printf("[API] Accepted block #%d from peer %s", block.Index, sourcePeer)
		writeJSON(w, http.StatusCreated, map[string]string{
			"status": "accepted",
			"height": fmt.Sprintf("%d", block.Index),
		})
		return
	}

	if needsSync {
		// Block doesn't extend our chain — sync with the source peer
		log.Printf("[API] Block #%d requires sync with peer %s", block.Index, sourcePeer)
		go func() {
			if sourcePeer != "" {
				_, orphanedTxs, err := h.chainSync.SyncWithPeer(sourcePeer)
				if err != nil {
					log.Printf("[API] Background sync with %s failed: %v", sourcePeer, err)
					return
				}
				// Return orphaned transactions to mempool
				if len(orphanedTxs) > 0 {
					h.mempool.ReturnOrphaned(orphanedTxs)
				}
			}
		}()

		writeJSON(w, http.StatusOK, map[string]string{"status": "syncing"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ignored"})
}

// HandleGetChain handles GET /chain — dump the full chain.
func (h *Handlers) HandleGetChain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	blocks := h.chain.GetAllBlocks()
	writeJSON(w, http.StatusOK, blocks)
}

// HandleGetChainHeight handles GET /chain/height — current height and head hash.
func (h *Handlers) HandleGetChainHeight(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, http.StatusOK, PeerChainInfo{
		Height:   h.chain.GetHeight(),
		HeadHash: h.chain.GetHeadHash(),
	})
}

// HandleGetBlocks handles GET /chain/blocks?from=N — get blocks after height N.
func (h *Handlers) HandleGetBlocks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fromStr := r.URL.Query().Get("from")
	from, err := strconv.ParseUint(fromStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid 'from' parameter", http.StatusBadRequest)
		return
	}

	blocks := h.chain.GetBlocksAfter(from)
	if blocks == nil {
		blocks = []*blockchain.Block{}
	}
	writeJSON(w, http.StatusOK, blocks)
}

// HandleGetPeers handles GET /peers — list known peers.
func (h *Handlers) HandleGetPeers(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		peers := h.peerManager.GetPeers()
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"peers": peers,
			"count": len(peers),
		})
		return
	}

	// POST /peers — exchange peer lists
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var req struct {
			Peers []string `json:"peers"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		newPeers := h.peerManager.ExchangePeers(req.Peers, h.selfAddr)
		ourPeers := h.peerManager.GetPeers()

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"peers":     ourPeers,
			"new_peers": len(newPeers),
		})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// HandleGetMempool handles GET /mempool — pending pool info.
func (h *Handlers) HandleGetMempool(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pending := h.mempool.GetPending()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"size":         h.mempool.Size(),
		"transactions": pending,
	})
}

// HandleGetBalances handles GET /balances — all account balances.
func (h *Handlers) HandleGetBalances(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	balances := h.chain.GetLedger().GetAllBalances()
	writeJSON(w, http.StatusOK, balances)
}

// HandleGetStatus handles GET /status — node status overview.
func (h *Handlers) HandleGetStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"address":      h.selfAddr,
		"height":       h.chain.GetHeight(),
		"head_hash":    h.chain.GetHeadHash(),
		"peer_count":   h.peerManager.PeerCount(),
		"mempool_size": h.mempool.Size(),
		"chain_length": h.chain.Length(),
	})
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
