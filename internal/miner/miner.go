package miner

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/blockchain2/internal/blockchain"
	"github.com/blockchain2/internal/mempool"
)

// Miner runs proof-of-work mining in a background goroutine.
type Miner struct {
	chain        *blockchain.Chain
	mempool      *mempool.Pool
	minerAddress string
	difficulty   int

	globalCtx    context.Context
	globalCancel context.CancelFunc
	mineCancel   context.CancelFunc
	done         chan struct{}
	mu           sync.Mutex
	mining       bool

	// OnBlockMined is called when a new block is successfully mined.
	// The callback should handle gossip to peers.
	OnBlockMined func(block *blockchain.Block)
}

// NewMiner creates a new miner.
func NewMiner(chain *blockchain.Chain, pool *mempool.Pool, minerAddress string, difficulty int) *Miner {
	return &Miner{
		chain:        chain,
		mempool:      pool,
		minerAddress: minerAddress,
		difficulty:   difficulty,
	}
}

// Start begins the mining loop in a background goroutine.
func (m *Miner) Start(ctx context.Context) {
	m.mu.Lock()
	if m.mining {
		m.mu.Unlock()
		return
	}
	m.mining = true
	m.done = make(chan struct{})

	globalCtx, globalCancel := context.WithCancel(ctx)
	m.globalCtx = globalCtx
	m.globalCancel = globalCancel
	m.mu.Unlock()

	go m.mineLoop()
	log.Printf("[MINER] Started mining (address: %s..., difficulty: %d)", m.minerAddress[:16], m.difficulty)
}

// Stop gracefully stops the miner.
func (m *Miner) Stop() {
	m.mu.Lock()
	if !m.mining {
		m.mu.Unlock()
		return
	}
	m.globalCancel()
	m.mining = false
	m.mu.Unlock()

	<-m.done
	log.Printf("[MINER] Stopped mining")
}

// Restart interrupts the current mining attempt (if any) and starts a fresh one.
// Called when a new block is received from the network.
func (m *Miner) Restart() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.mining {
		return
	}

	if m.mineCancel != nil {
		m.mineCancel()
	}
}

// mineLoop is the main mining loop.
func (m *Miner) mineLoop() {
	defer close(m.done)

	for {
		select {
		case <-m.globalCtx.Done():
			return
		default:
			// Create a cancellable context for the current mining attempt
			m.mu.Lock()
			mineCtx, mineCancel := context.WithCancel(m.globalCtx)
			m.mineCancel = mineCancel
			m.mu.Unlock()

			m.mineBlock(mineCtx)

			// Small delay between mining attempts to prevent CPU saturation
			select {
			case <-m.globalCtx.Done():
				return
			case <-time.After(100 * time.Millisecond):
			}
		}
	}
}

// mineBlock attempts to mine a single block.
func (m *Miner) mineBlock(ctx context.Context) {
	// Get pending transactions and filter them against current ledger state
	pendingTxs := m.mempool.GetPending()
	validTxs := m.chain.GetLedger().FilterValidTransactions(pendingTxs)

	// Create coinbase transaction (mining reward + fees)
	totalFees := uint64(0)
	for _, tx := range validTxs {
		totalFees += tx.Fee
	}
	coinbase := blockchain.NewCoinbaseTransaction(m.minerAddress, blockchain.MiningReward+totalFees)

	// Build transaction list: coinbase first, then valid pending
	txs := make([]*blockchain.Transaction, 0, len(validTxs)+1)
	txs = append(txs, coinbase)
	txs = append(txs, validTxs...)

	// Create the block
	latest := m.chain.GetLatestBlock()
	block := blockchain.NewBlock(latest.Index+1, txs, latest.Hash, m.minerAddress, m.difficulty)

	// Mine it (proof of work)
	mineDone := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			close(mineDone)
		}
	}()

	if !block.Mine(mineDone) {
		return // Mining was cancelled (e.g. new block received)
	}

	// Try to add the block to our chain
	err := m.chain.AddBlock(block)
	if err != nil {
		// Chain might have been updated while we were mining
		log.Printf("[MINER] Failed to add mined block: %v", err)
		return
	}

	// Remove mined transactions from mempool
	m.mempool.RemoveByBlock(block)

	log.Printf("[MINER] 🎉 Mined block #%d (hash: %s..., txs: %d)",
		block.Index, block.Hash[:16], len(block.Transactions))

	// Notify callback (for gossip)
	if m.OnBlockMined != nil {
		m.OnBlockMined(block)
	}
}

// IsMining returns whether the miner is currently running.
func (m *Miner) IsMining() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.mining
}
