package node

import (
	"context"
	"log"
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
}

// NewNode creates a new blockchain node with all components initialised.
func NewNode(config Config) *Node {
	// Set defaults
	if config.Difficulty == 0 {
		config.Difficulty = blockchain.DefaultDifficulty
	}

	// Create core components
	chain := blockchain.NewChain()
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
	}

	// Create HTTP server
	server := network.NewServer(config.ListenAddr, chain, pool, peerManager, gossiper, chainSync)

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
		}
		if len(orphanedTxs) > 0 {
			n.mempool.ReturnOrphaned(orphanedTxs)
		}
	}

	// Start periodic sync
	go n.periodicSync(ctx)

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

// periodicSync runs chain synchronisation with peers on a timer.
func (n *Node) periodicSync(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-n.done:
			return
		case <-ticker.C:
			added, orphanedTxs := n.chainSync.SyncWithAllPeers()
			if added > 0 {
				log.Printf("[NODE] Periodic sync: added %d blocks", added)
			}
			if len(orphanedTxs) > 0 {
				n.mempool.ReturnOrphaned(orphanedTxs)
			}
		}
	}
}
