package blockchain

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SaveToFile serialises the entire chain to a JSON file.
func (c *Chain) SaveToFile(filename string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(c.blocks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal chain: %w", err)
	}

	// Write to temporary file first, then rename (atomic save)
	tempFile := filename + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return err
	}
	return os.Rename(tempFile, filename)
}

// LoadFromFile loads the chain from a JSON file and rebuilds the ledger.
func (c *Chain) LoadFromFile(filename string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var blocks []*Block
	if err := json.Unmarshal(data, &blocks); err != nil {
		return fmt.Errorf("failed to unmarshal chain: %w", err)
	}

	if len(blocks) == 0 {
		return fmt.Errorf("loaded chain is empty")
	}

	// Validate genesis
	if blocks[0].Hash != c.blocks[0].Hash {
		return fmt.Errorf("loaded chain has invalid genesis block")
	}

	// Validate all loaded blocks
	for i := 1; i < len(blocks); i++ {
		if err := blocks[i].Validate(blocks[i-1]); err != nil {
			return fmt.Errorf("invalid block #%d in loaded chain: %w", i, err)
		}
	}

	c.blocks = blocks

	// Rebuild ledger
	if err := c.ledger.RebuildFromChain(c.blocks); err != nil {
		return fmt.Errorf("failed to rebuild ledger from loaded chain: %w", err)
	}

	return nil
}
