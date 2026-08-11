package blockchain

import (
	"fmt"
	"sync"
)

// Ledger tracks account balances derived from the blockchain.
type Ledger struct {
	balances map[string]uint64
	mu       sync.RWMutex
}

// NewLedger creates a new empty ledger.
func NewLedger() *Ledger {
	return &Ledger{
		balances: make(map[string]uint64),
	}
}

// GetBalance returns the balance for a given address.
func (l *Ledger) GetBalance(address string) uint64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.balances[address]
}

// GetAllBalances returns a copy of all account balances.
func (l *Ledger) GetAllBalances() map[string]uint64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	result := make(map[string]uint64, len(l.balances))
	for k, v := range l.balances {
		result[k] = v
	}
	return result
}

// RebuildFromChain rebuilds the entire ledger from a list of blocks.
// This is called during initialization and chain reorganisation.
func (l *Ledger) RebuildFromChain(blocks []*Block) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Reset balances
	l.balances = make(map[string]uint64)

	for _, block := range blocks {
		if err := l.applyBlock(block); err != nil {
			return fmt.Errorf("failed to apply block %d: %w", block.Index, err)
		}
	}
	return nil
}

// ApplyBlock applies a single block's transactions to the ledger.
func (l *Ledger) ApplyBlock(block *Block) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.applyBlock(block)
}

// applyBlock applies a block without locking (caller must hold the lock).
func (l *Ledger) applyBlock(block *Block) error {
	for _, tx := range block.Transactions {
		if tx.IsCoinbase() {
			// Mining reward — credit the recipient
			l.balances[tx.To] += tx.Amount
			continue
		}

		totalDebit := tx.Amount + tx.Fee
		if l.balances[tx.From] < totalDebit {
			return fmt.Errorf("insufficient balance for tx %s: have %d, need %d",
				tx.ID, l.balances[tx.From], totalDebit)
		}

		l.balances[tx.From] -= totalDebit
		l.balances[tx.To] += tx.Amount
		// Fees go to the miner (via coinbase)
	}
	return nil
}

// CanApplyTransaction checks if a transaction can be applied without actually applying it.
func (l *Ledger) CanApplyTransaction(tx *Transaction) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if tx.IsCoinbase() {
		return true
	}

	totalDebit := tx.Amount + tx.Fee
	return l.balances[tx.From] >= totalDebit
}
