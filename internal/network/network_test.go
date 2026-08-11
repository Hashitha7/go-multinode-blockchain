package network

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/blockchain2/internal/blockchain"
	"github.com/blockchain2/internal/crypto"
	"github.com/blockchain2/internal/mempool"
)

func setupTestHandlers() (*Handlers, *blockchain.Chain, *mempool.Pool) {
	chain := blockchain.NewChain()
	pool := mempool.NewPool()
	pm := NewPeerManager()
	gossiper := NewGossiper(pm, "localhost:9999")
	cs := NewChainSync(chain, pm)
	handlers := NewHandlers(chain, pool, pm, gossiper, cs, "localhost:9999")
	return handlers, chain, pool
}

func TestHandleGetStatus(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()

	handlers.HandleGetStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var status map[string]interface{}
	json.NewDecoder(w.Body).Decode(&status)

	if status["height"].(float64) != 0 {
		t.Errorf("Expected height 0, got %v", status["height"])
	}
}

func TestHandleGetChainHeight(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	req := httptest.NewRequest("GET", "/chain/height", nil)
	w := httptest.NewRecorder()

	handlers.HandleGetChainHeight(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var info PeerChainInfo
	json.NewDecoder(w.Body).Decode(&info)

	if info.Height != 0 {
		t.Errorf("Expected height 0, got %d", info.Height)
	}
	if info.HeadHash == "" {
		t.Error("Head hash should not be empty")
	}
}

func TestHandleSubmitTransaction(t *testing.T) {
	handlers, _, pool := setupTestHandlers()

	// Create a valid transaction
	pub, priv, _ := crypto.GenerateKeyPair()
	tx := blockchain.NewTransaction("recipient_address_1234567890abcdef1234", 10, 1, pub, priv)

	txJSON, _ := json.Marshal(tx)
	req := httptest.NewRequest("POST", "/tx", strings.NewReader(string(txJSON)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.HandleSubmitTransaction(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	if pool.Size() != 1 {
		t.Errorf("Mempool should have 1 tx, got %d", pool.Size())
	}
}

func TestHandleSubmitTransactionInvalidSignature(t *testing.T) {
	handlers, _, pool := setupTestHandlers()

	// Create a transaction and tamper with it
	pub, priv, _ := crypto.GenerateKeyPair()
	tx := blockchain.NewTransaction("recipient_address_1234567890abcdef1234", 10, 1, pub, priv)
	tx.Amount = 999 // Tamper

	txJSON, _ := json.Marshal(tx)
	req := httptest.NewRequest("POST", "/tx", strings.NewReader(string(txJSON)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.HandleSubmitTransaction(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid tx, got %d", w.Code)
	}

	if pool.Size() != 0 {
		t.Errorf("Invalid tx should not be added to mempool, size: %d", pool.Size())
	}
}

func TestHandleGetChain(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	req := httptest.NewRequest("GET", "/chain", nil)
	w := httptest.NewRecorder()

	handlers.HandleGetChain(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var blocks []*blockchain.Block
	json.NewDecoder(w.Body).Decode(&blocks)

	if len(blocks) != 1 {
		t.Errorf("Expected 1 block (genesis), got %d", len(blocks))
	}
}

func TestHandleGetMempool(t *testing.T) {
	handlers, _, pool := setupTestHandlers()

	pub, priv, _ := crypto.GenerateKeyPair()
	tx := blockchain.NewTransaction("recipient_address_1234567890abcdef1234", 10, 1, pub, priv)
	pool.Add(tx)

	req := httptest.NewRequest("GET", "/mempool", nil)
	w := httptest.NewRecorder()

	handlers.HandleGetMempool(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result map[string]interface{}
	json.NewDecoder(w.Body).Decode(&result)

	if result["size"].(float64) != 1 {
		t.Errorf("Expected mempool size 1, got %v", result["size"])
	}
}

func TestHandleGetBalances(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	req := httptest.NewRequest("GET", "/balances", nil)
	w := httptest.NewRecorder()

	handlers.HandleGetBalances(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestPeerManager(t *testing.T) {
	pm := NewPeerManager()

	added := pm.AddPeer("localhost:3001")
	if !added {
		t.Error("First add should return true")
	}

	added = pm.AddPeer("localhost:3001")
	if added {
		t.Error("Duplicate add should return false")
	}

	if pm.PeerCount() != 1 {
		t.Errorf("Expected 1 peer, got %d", pm.PeerCount())
	}

	peers := pm.GetPeers()
	if len(peers) != 1 || peers[0] != "localhost:3001" {
		t.Errorf("Expected [localhost:3001], got %v", peers)
	}

	pm.RemovePeer("localhost:3001")
	if pm.PeerCount() != 0 {
		t.Errorf("Expected 0 peers after remove, got %d", pm.PeerCount())
	}
}

func TestPeerExchange(t *testing.T) {
	pm := NewPeerManager()
	pm.AddPeer("localhost:3001")

	newPeers := pm.ExchangePeers(
		[]string{"localhost:3002", "localhost:3003", "localhost:9999"},
		"localhost:9999", // self
	)

	if len(newPeers) != 2 {
		t.Errorf("Expected 2 new peers, got %d", len(newPeers))
	}

	if pm.PeerCount() != 3 {
		t.Errorf("Expected 3 total peers, got %d", pm.PeerCount())
	}
}

func TestSeenTracker(t *testing.T) {
	st := NewSeenTracker(1 * time.Second)

	// First time should not be seen
	if st.MarkSeen("tx1") {
		t.Error("First MarkSeen should return false")
	}

	// Second time should be seen
	if !st.MarkSeen("tx1") {
		t.Error("Second MarkSeen should return true")
	}

	// Different ID should not be seen
	if st.MarkSeen("tx2") {
		t.Error("Different ID should not be seen")
	}
}

// TestMultiNodeIntegration tests two nodes communicating via HTTP.
func TestMultiNodeIntegration(t *testing.T) {
	// Create two node setups
	chain1 := blockchain.NewChain()
	pool1 := mempool.NewPool()
	pm1 := NewPeerManager()

	chain2 := blockchain.NewChain()
	pool2 := mempool.NewPool()
	pm2 := NewPeerManager()

	// Create test HTTP servers
	gossiper1 := NewGossiper(pm1, "node1")
	cs1 := NewChainSync(chain1, pm1)
	handlers1 := NewHandlers(chain1, pool1, pm1, gossiper1, cs1, "node1")

	gossiper2 := NewGossiper(pm2, "node2")
	cs2 := NewChainSync(chain2, pm2)
	handlers2 := NewHandlers(chain2, pool2, pm2, gossiper2, cs2, "node2")

	// Set up HTTP test servers
	mux1 := http.NewServeMux()
	mux1.HandleFunc("/tx", handlers1.HandleSubmitTransaction)
	mux1.HandleFunc("/block", handlers1.HandleReceiveBlock)
	mux1.HandleFunc("/chain", handlers1.HandleGetChain)
	mux1.HandleFunc("/chain/height", handlers1.HandleGetChainHeight)
	mux1.HandleFunc("/chain/blocks", handlers1.HandleGetBlocks)
	mux1.HandleFunc("/status", handlers1.HandleGetStatus)
	server1 := httptest.NewServer(mux1)
	defer server1.Close()

	mux2 := http.NewServeMux()
	mux2.HandleFunc("/tx", handlers2.HandleSubmitTransaction)
	mux2.HandleFunc("/block", handlers2.HandleReceiveBlock)
	mux2.HandleFunc("/chain", handlers2.HandleGetChain)
	mux2.HandleFunc("/chain/height", handlers2.HandleGetChainHeight)
	mux2.HandleFunc("/chain/blocks", handlers2.HandleGetBlocks)
	mux2.HandleFunc("/status", handlers2.HandleGetStatus)
	server2 := httptest.NewServer(mux2)
	defer server2.Close()

	// Extract addresses
	addr1 := server1.Listener.Addr().String()
	addr2 := server2.Listener.Addr().String()

	// Register peers
	pm1.AddPeer(addr2)
	pm2.AddPeer(addr1)

	// Mine a block on node1
	genesis := chain1.GetLatestBlock()
	block1 := blockchain.NewBlock(1, []*blockchain.Transaction{
		blockchain.NewCoinbaseTransaction("miner1", blockchain.MiningReward),
	}, genesis.Hash, "miner1", 1)
	done := make(chan struct{})
	defer close(done)
	block1.Mine(done)
	chain1.AddBlock(block1)

	// Sync node2 with node1
	syncCS2 := NewChainSync(chain2, pm2)
	added, _, err := syncCS2.SyncWithPeer(addr1)
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	if added != 1 {
		t.Errorf("Expected 1 block added during sync, got %d", added)
	}

	if chain2.GetHeight() != 1 {
		t.Errorf("Node2 height should be 1, got %d", chain2.GetHeight())
	}

	// Verify both chains have the same head
	if chain1.GetHeadHash() != chain2.GetHeadHash() {
		t.Error("Both nodes should have the same head hash after sync")
	}

	_ = fmt.Sprintf("test") // avoid unused import
}
