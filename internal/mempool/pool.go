package mempool

import (
	"log"
	"sync"

	"github.com/blockchain2/internal/blockchain"
)

// Pool is a thread-safe pending transaction pool with de-duplication.
type Pool struct {
	transactions map[string]*blockchain.Transaction // txID -> transaction
	mu           sync.RWMutex
}

// NewPool creates a new empty transaction pool.
func NewPool() *Pool {
	return &Pool{
		transactions: make(map[string]*blockchain.Transaction),
	}
}

// Add adds a transaction to the pool if it's not already present.
// Returns true if the transaction was added (new), false if it was a duplicate.
func (p *Pool) Add(tx *blockchain.Transaction) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.transactions[tx.ID]; exists {
		return false // De-duplicate
	}

	p.transactions[tx.ID] = tx
	log.Printf("[MEMPOOL] Added transaction %s... (pool size: %d)", tx.ID[:16], len(p.transactions))
	return true
}

// Remove removes transactions with the given IDs from the pool.
// Called when transactions are included in a mined block.
func (p *Pool) Remove(txIDs []string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, id := range txIDs {
		delete(p.transactions, id)
	}
}

// RemoveByBlock removes all transactions that appear in the given block.
func (p *Pool) RemoveByBlock(block *blockchain.Block) {
	ids := make([]string, 0, len(block.Transactions))
	for _, tx := range block.Transactions {
		if !tx.IsCoinbase() {
			ids = append(ids, tx.ID)
		}
	}
	p.Remove(ids)
}

// GetPending returns a copy of all pending transactions.
func (p *Pool) GetPending() []*blockchain.Transaction {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]*blockchain.Transaction, 0, len(p.transactions))
	for _, tx := range p.transactions {
		result = append(result, tx)
	}
	return result
}

// Contains checks if a transaction ID is already in the pool.
func (p *Pool) Contains(txID string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	_, exists := p.transactions[txID]
	return exists
}

// Size returns the number of pending transactions.
func (p *Pool) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.transactions)
}

// ReturnOrphaned adds back transactions that were orphaned during a chain reorganisation.
// Only adds transactions whose signatures are still valid.
func (p *Pool) ReturnOrphaned(txs []*blockchain.Transaction) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, tx := range txs {
		if tx.IsCoinbase() {
			continue
		}
		// Verify the transaction is still valid before returning it
		if err := tx.Verify(); err != nil {
			log.Printf("[MEMPOOL] Dropping orphaned tx %s...: %v", tx.ID[:16], err)
			continue
		}
		if _, exists := p.transactions[tx.ID]; !exists {
			p.transactions[tx.ID] = tx
			log.Printf("[MEMPOOL] Returned orphaned tx %s... to pool", tx.ID[:16])
		}
	}
}

// Clear removes all transactions from the pool.
func (p *Pool) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.transactions = make(map[string]*blockchain.Transaction)
}
