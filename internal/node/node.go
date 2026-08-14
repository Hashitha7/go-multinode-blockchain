package node

import (
	"context"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/blockchain2/internal/blockchain"
	"github.com/blockchain2/internal/mempool"
	"github.com/blockchain2/internal/miner"
	"github.com/blockchain2/internal/network"
)

// Config holds the configuration for a blockchain node.
type Config struct {
	ListenAddr   string   // Address to listen on (e.g., "localhost:3001")
	Peers        []string // Initial peer addresses
	MinerAddress string   // Address to receive mining rewards
	Difficulty   int      // Mining difficulty (number of leading zeros)
	EnableMining bool     // Whether to enable mining
}

// Node is the main orchestrator that ties all components together.
type Node struct {
	config      Config
	chain       *blockchain.Chain
	mempool     *mempool.Pool
	miner       *miner.Miner
	peerManager *network.PeerManager
	gossiper    *network.Gossiper
	chainSync   *network.ChainSync
	server      *network.Server
	done        chan struct{}
	chainFile   string
}

// NewNode creates a new blockchain node with all components initialised.
func NewNode(config Config) *Node {
	// Set defaults
	if config.Difficulty == 0 {
		config.Difficulty = blockchain.DefaultDifficulty
	}

	// Extract port for data isolation
	parts := strings.Split(config.ListenAddr, ":")
	port := "default"
	if len(parts) == 2 {
		port = parts[1]
	}
	dataDir := filepath.Join(".data", "node_"+port)
	chainFile := filepath.Join(dataDir, "chain.json")

	// Create core components
	chain := blockchain.NewChain()

	// Load chain from disk (FR-11 Persistence)
	if err := chain.LoadFromFile(chainFile); err != nil {
		log.Printf("[NODE] No local chain found or failed to load: %v. Starting fresh.", err)
	} else {
		log.Printf("[NODE] Loaded chain from disk (height: %d)", chain.GetHeight())
	}

	pool := mempool.NewPool()
	peerManager := network.NewPeerManager()

	// Add initial peers
	for _, peer := range config.Peers {
		peerManager.AddPeer(peer)
	}

	// Create network components
	gossiper := network.NewGossiper(peerManager, config.ListenAddr)
	chainSync := network.NewChainSync(chain, peerManager)

	// Create miner
	m := miner.NewMiner(chain, pool, config.MinerAddress, config.Difficulty)

	// Wire up the miner's block callback to gossip
	m.OnBlockMined = func(block *blockchain.Block) {
		gossiper.GossipBlock(block, "")

		// Save chain to disk after mining
		if err := chain.SaveToFile(chainFile); err != nil {
			log.Printf("[NODE] Failed to save chain to disk: %v", err)
		}
	}

	// Create HTTP server
	server := network.NewServer(config.ListenAddr, chain, pool, peerManager, gossiper, chainSync, func() {
		m.Restart()

		// Save chain to disk after accepting a block
		if err := chain.SaveToFile(chainFile); err != nil {
			log.Printf("[NODE] Failed to save chain to disk: %v", err)
		}
	})

	return &Node{
		config:      config,
		chain:       chain,
		mempool:     pool,
		miner:       m,
		peerManager: peerManager,
		gossiper:    gossiper,
		chainSync:   chainSync,
		server:      server,
		done:        make(chan struct{}),
		chainFile:   chainFile,
	}
}

// Start starts all node components.
func (n *Node) Start(ctx context.Context) error {
	log.Printf("========================================")
	log.Printf("  Blockchain Node Starting")
	log.Printf("  Address: %s", n.config.ListenAddr)
	log.Printf("  Peers: %v", n.config.Peers)
	log.Printf("  Miner: %s", n.config.MinerAddress)
	log.Printf("  Difficulty: %d", n.config.Difficulty)
	log.Printf("  Mining: %v", n.config.EnableMining)
	log.Printf("========================================")

	// Start gossip cleanup
	n.gossiper.StartCleanup(n.done)

	// Start HTTP server in a goroutine
	go func() {
		if err := n.server.Start(); err != nil {
			log.Printf("[NODE] HTTP server error: %v", err)
		}
	}()

	// Wait a moment for the server to start
	time.Sleep(500 * time.Millisecond)

	// Initial sync with peers
	if len(n.config.Peers) > 0 {
		log.Printf("[NODE] Performing initial sync with peers...")
		added, orphanedTxs := n.chainSync.SyncWithAllPeers()
		if added > 0 {
			log.Printf("[NODE] Initial sync: added %d blocks", added)
			if err := n.chain.SaveToFile(n.chainFile); err != nil {
				log.Printf("[NODE] Failed to save chain to disk: %v", err)
			}
		}
		if len(orphanedTxs) > 0 {
			n.mempool.ReturnOrphaned(orphanedTxs)
		}
	}

	// Start periodic sync and peer health checks (FR-10)
	go n.periodicTasks(ctx)

	// Start mining if enabled
	if n.config.EnableMining {
		n.miner.Start(ctx)
	}

	// Wait for context cancellation
	<-ctx.Done()
	return n.Shutdown()
}

// Shutdown gracefully stops all node components.
func (n *Node) Shutdown() error {
	log.Printf("[NODE] Shutting down...")

	// Stop miner
	n.miner.Stop()

	// Save final chain state
	if err := n.chain.SaveToFile(n.chainFile); err != nil {
		log.Printf("[NODE] Failed to save chain on shutdown: %v", err)
	}

	// Close done channel to stop background goroutines
	close(n.done)

	// Shutdown HTTP server
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := n.server.Shutdown(shutdownCtx); err != nil {
		log.Printf("[NODE] HTTP server shutdown error: %v", err)
		return err
	}

	log.Printf("[NODE] Shutdown complete")
	return nil
}

// periodicTasks runs chain synchronisation and peer health checks (FR-10).
func (n *Node) periodicTasks(ctx context.Context) {
	syncTicker := time.NewTicker(15 * time.Second)
	healthTicker := time.NewTicker(30 * time.Second)
	defer syncTicker.Stop()
	defer healthTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-n.done:
			return
		case <-syncTicker.C:
			added, orphanedTxs := n.chainSync.SyncWithAllPeers()
			if added > 0 {
				log.Printf("[NODE] Periodic sync: added %d blocks", added)
				n.chain.SaveToFile(n.chainFile)
			}
			if len(orphanedTxs) > 0 {
				n.mempool.ReturnOrphaned(orphanedTxs)
			}
		case <-healthTicker.C:
			// FR-10: Peer health check
			peers := n.peerManager.GetPeers()
			for _, peer := range peers {
				_, err := n.chainSync.GetPeerChainInfo(peer)
				if err != nil {
					log.Printf("[PEERS] Peer %s is unreachable, removing from known peers", peer)
					n.peerManager.RemovePeer(peer)
				}
			}
		}
	}
}
