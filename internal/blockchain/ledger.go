package blockchain

import (
	"fmt"
	"sync"
)

// Ledger tracks account balances derived from the blockchain.
type Ledger struct {
	balances     map[string]uint64
	processedTxs map[string]bool
	mu           sync.RWMutex
}

// NewLedger creates a new empty ledger.
func NewLedger() *Ledger {
	return &Ledger{
		balances:     make(map[string]uint64),
		processedTxs: make(map[string]bool),
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

	// Reset balances and processed txs
	l.balances = make(map[string]uint64)
	l.processedTxs = make(map[string]bool)

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
// It applies transactions to a temporary clone first to ensure atomicity.
func (l *Ledger) applyBlock(block *Block) error {
	// 1. Create a clone of the current state
	tempBalances := make(map[string]uint64, len(l.balances))
	for k, v := range l.balances {
		tempBalances[k] = v
	}

	tempProcessed := make(map[string]bool, len(l.processedTxs))
	for k, v := range l.processedTxs {
		tempProcessed[k] = v
	}

	var totalFees uint64
	var coinbaseTx *Transaction

	// 2. Apply transactions to the clone
	for i, tx := range block.Transactions {
		if tx.IsCoinbase() {
			if i != 0 {
				return fmt.Errorf("coinbase transaction must be first")
			}
			coinbaseTx = tx
			continue
		}

		if tempProcessed[tx.ID] {
			return fmt.Errorf("transaction replay detected: %s", tx.ID)
		}

		totalDebit := tx.Amount + tx.Fee
		if tempBalances[tx.From] < totalDebit {
			return fmt.Errorf("insufficient balance for tx %s: have %d, need %d",
				tx.ID, tempBalances[tx.From], totalDebit)
		}

		tempBalances[tx.From] -= totalDebit
		tempBalances[tx.To] += tx.Amount
		tempProcessed[tx.ID] = true
		totalFees += tx.Fee
	}

	// 3. Validate and apply coinbase
	if coinbaseTx != nil {
		maxAllowedReward := MiningReward + totalFees
		if coinbaseTx.Amount > maxAllowedReward {
			return fmt.Errorf("coinbase reward exceeds allowed maximum: got %d, max %d", coinbaseTx.Amount, maxAllowedReward)
		}
		tempBalances[coinbaseTx.To] += coinbaseTx.Amount
		tempProcessed[coinbaseTx.ID] = true
	}

	// 4. Commit changes (Atomic Update)
	l.balances = tempBalances
	l.processedTxs = tempProcessed

	return nil
}

// CanApplyTransaction checks if a transaction can be applied without actually applying it.
// It also checks for replay protection.
func (l *Ledger) CanApplyTransaction(tx *Transaction) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if tx.IsCoinbase() {
		return true
	}

	if l.processedTxs[tx.ID] {
		return false // Replay protection
	}

	totalDebit := tx.Amount + tx.Fee
	return l.balances[tx.From] >= totalDebit
}

// FilterValidTransactions takes a list of pending transactions and returns only
// those that are valid against the current ledger state without double-spending.
func (l *Ledger) FilterValidTransactions(txs []*Transaction) []*Transaction {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Clone the state to simulate applying transactions
	tempBalances := make(map[string]uint64, len(l.balances))
	for k, v := range l.balances {
		tempBalances[k] = v
	}
	tempProcessed := make(map[string]bool, len(l.processedTxs))
	for k, v := range l.processedTxs {
		tempProcessed[k] = v
	}

	validTxs := make([]*Transaction, 0, len(txs))

	for _, tx := range txs {
		if tx.IsCoinbase() {
			continue // Handled separately
		}

		if tempProcessed[tx.ID] {
			continue // Replay protection
		}

		totalDebit := tx.Amount + tx.Fee
		if tempBalances[tx.From] < totalDebit {
			continue // Insufficient funds (might be caused by a previous tx in this list)
		}

		// Apply to temp state
		tempBalances[tx.From] -= totalDebit
		tempBalances[tx.To] += tx.Amount
		tempProcessed[tx.ID] = true

		validTxs = append(validTxs, tx)
	}

	return validTxs
}
