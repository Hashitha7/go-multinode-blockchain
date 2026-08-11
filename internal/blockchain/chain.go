package blockchain

import (
	"fmt"
	"log"
	"sync"
)

// Chain manages the blockchain state including blocks, ledger, and fork resolution.
type Chain struct {
	blocks     []*Block
	sideChains map[string]*Block // hash -> block (orphans/competing blocks)
	ledger     *Ledger
	mu         sync.RWMutex
}

// NewChain creates a new chain starting with the genesis block.
func NewChain() *Chain {
	genesis := GenesisBlock()
	c := &Chain{
		blocks:     []*Block{genesis},
		sideChains: make(map[string]*Block),
		ledger:     NewLedger(),
	}
	return c
}

// GetHeight returns the current chain height (index of the latest block).
func (c *Chain) GetHeight() uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.blocks[len(c.blocks)-1].Index
}

// GetLatestBlock returns the most recent block in the chain.
func (c *Chain) GetLatestBlock() *Block {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.blocks[len(c.blocks)-1]
}

// GetBlockByIndex returns the block at the given index, or nil if out of range.
func (c *Chain) GetBlockByIndex(index uint64) *Block {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if index >= uint64(len(c.blocks)) {
		return nil
	}
	return c.blocks[index]
}

// GetBlocksAfter returns all blocks with index > afterIndex.
// Used for chain synchronisation.
func (c *Chain) GetBlocksAfter(afterIndex uint64) []*Block {
	c.mu.RLock()
	defer c.mu.RUnlock()

	startIdx := int(afterIndex + 1)
	if startIdx >= len(c.blocks) {
		return nil
	}

	result := make([]*Block, len(c.blocks)-startIdx)
	copy(result, c.blocks[startIdx:])
	return result
}

// GetAllBlocks returns a copy of all blocks in the chain.
func (c *Chain) GetAllBlocks() []*Block {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]*Block, len(c.blocks))
	copy(result, c.blocks)
	return result
}

// GetHeadHash returns the hash of the latest block.
func (c *Chain) GetHeadHash() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.blocks[len(c.blocks)-1].Hash
}

// GetLedger returns the ledger associated with this chain.
func (c *Chain) GetLedger() *Ledger {
	return c.ledger
}

// calculateWork computes the total proof of work for a slice of blocks.
func calculateWork(blocks []*Block) uint64 {
	var work uint64
	for _, b := range blocks {
		// A simple approximation: work = 2^difficulty
		work += 1 << b.Difficulty
	}
	return work
}

// AddBlock validates and appends a block to the chain.
// Returns an error if the block is invalid or doesn't extend the current chain.
func (c *Chain) AddBlock(block *Block) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	latest := c.blocks[len(c.blocks)-1]

	// Validate the block against the current tip
	if err := block.Validate(latest); err != nil {
		return fmt.Errorf("block validation failed: %w", err)
	}

	// Apply block transactions to ledger
	if err := c.ledger.ApplyBlock(block); err != nil {
		return fmt.Errorf("failed to apply block to ledger: %w", err)
	}

	c.blocks = append(c.blocks, block)
	log.Printf("[CHAIN] Added block #%d (hash: %s...)", block.Index, block.Hash[:16])
	return nil
}

// ReorganiseTo switches to a better competing chain if it has more cumulative work.
// Returns orphaned transactions that were in the old chain but not in the new one.
// This implements FR-6 and the Cumulative Work stretch goal.
func (c *Chain) ReorganiseTo(newBlocks []*Block) ([]*Transaction, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Ensure the new chain starts from the exact same genesis block
	if newBlocks[0].Hash != c.blocks[0].Hash {
		return nil, fmt.Errorf("invalid genesis block in new chain")
	}

	// Compare cumulative work (Heaviest Chain)
	currentWork := calculateWork(c.blocks)
	newWork := calculateWork(newBlocks)

	if newWork <= currentWork {
		return nil, fmt.Errorf("new chain does not have more work: new=%d, current=%d", newWork, currentWork)
	}

	// Find the common ancestor (fork point)
	forkPoint := uint64(0)
	for i := uint64(0); i < uint64(len(c.blocks)) && i < uint64(len(newBlocks)); i++ {
		if c.blocks[i].Hash == newBlocks[i].Hash {
			forkPoint = i
		} else {
			break
		}
	}

	// Validate the new chain from the fork point
	for i := forkPoint + 1; i < uint64(len(newBlocks)); i++ {
		prevBlock := newBlocks[i-1]
		if err := newBlocks[i].Validate(prevBlock); err != nil {
			return nil, fmt.Errorf("invalid block #%d in new chain: %w", i, err)
		}
	}

	// Collect transactions from orphaned blocks (blocks in old chain after fork point)
	orphanedTxs := make([]*Transaction, 0)
	newTxIDs := make(map[string]bool)

	// Build set of transaction IDs in the new chain
	for i := forkPoint + 1; i < uint64(len(newBlocks)); i++ {
		for _, tx := range newBlocks[i].Transactions {
			newTxIDs[tx.ID] = true
		}
	}

	// Collect orphaned transactions not in the new chain
	for i := forkPoint + 1; i < uint64(len(c.blocks)); i++ {
		for _, tx := range c.blocks[i].Transactions {
			if !tx.IsCoinbase() && !newTxIDs[tx.ID] {
				orphanedTxs = append(orphanedTxs, tx)
			}
		}
	}

	// Switch to the new chain
	oldHeight := len(c.blocks) - 1
	c.blocks = make([]*Block, len(newBlocks))
	copy(c.blocks, newBlocks)

	// Rebuild the ledger from scratch
	if err := c.ledger.RebuildFromChain(c.blocks); err != nil {
		return nil, fmt.Errorf("failed to rebuild ledger after reorg: %w", err)
	}

	log.Printf("[CHAIN] Reorganised from height %d to %d (fork at #%d, %d orphaned txs)",
		oldHeight, len(c.blocks)-1, forkPoint, len(orphanedTxs))

	return orphanedTxs, nil
}

// TryAddOrReorg attempts to add a block. If it doesn't extend the current chain
// but belongs to a competing chain, it stores it in sideChains and evaluates work.
// Returns: (added bool, needsSync bool, error)
func (c *Chain) TryAddOrReorg(block *Block) (bool, bool, error) {
	c.mu.RLock()
	latest := c.blocks[len(c.blocks)-1]
	c.mu.RUnlock()

	// Case 1: Block extends our chain directly
	if block.PrevHash == latest.Hash && block.Index == latest.Index+1 {
		err := c.AddBlock(block)
		if err != nil {
			return false, false, err
		}
		return true, false, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Prevent storing invalid or deeply orphaned blocks
	if block.Index > latest.Index+50 {
		return false, true, nil // Too far ahead, trigger full sync
	}

	// Case 2 & 3: Competing block (same height or slightly ahead)
	// We store it in sideChains
	c.sideChains[block.Hash] = block

	// Try to build a chain back to our main chain to see if it's heavier
	sideChain := []*Block{block}
	curr := block
	for curr.PrevHash != "" {
		if parent, exists := c.sideChains[curr.PrevHash]; exists {
			sideChain = append([]*Block{parent}, sideChain...)
			curr = parent
		} else {
			break
		}
	}

	// If the side chain connects to our main chain, evaluate its total work
	if int(curr.Index-1) < len(c.blocks) && c.blocks[curr.Index-1].Hash == curr.PrevHash {
		// We have a full path to the main chain!
		// Build the hypothetical new main chain
		hypotheticalChain := make([]*Block, curr.Index)
		copy(hypotheticalChain, c.blocks[:curr.Index])
		hypotheticalChain = append(hypotheticalChain, sideChain...)

		currentWork := calculateWork(c.blocks)
		newWork := calculateWork(hypotheticalChain)

		if newWork > currentWork {
			// Trigger Reorg! (We need to release lock first since ReorganiseTo locks)
			c.mu.Unlock()
			_, err := c.ReorganiseTo(hypotheticalChain)
			c.mu.Lock() // Re-acquire lock for the defer
			if err != nil {
				return false, false, fmt.Errorf("failed to reorg from sidechain: %v", err)
			}
			return true, false, nil
		}
	}

	// If it doesn't connect, or isn't heavier, we just store it and maybe trigger sync
	if block.Index > latest.Index {
		return false, true, nil // Ahead, trigger sync to get missing links
	}

	return false, false, nil
}

// Length returns the number of blocks in the chain.
func (c *Chain) Length() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.blocks)
}
