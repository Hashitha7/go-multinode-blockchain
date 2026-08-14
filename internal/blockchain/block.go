package blockchain

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	bcrypto "github.com/blockchain2/internal/crypto"
)

// DefaultDifficulty is the number of leading zeros required in the block hash.
const DefaultDifficulty = 4

// Block represents a block in the blockchain.
type Block struct {
	Index        uint64         `json:"index"`
	Timestamp    int64          `json:"timestamp"`
	Transactions []*Transaction `json:"transactions"`
	PrevHash     string         `json:"prev_hash"`
	Hash         string         `json:"hash"`
	Nonce        uint64         `json:"nonce"`
	MinerAddress string         `json:"miner_address"`
	Difficulty   int            `json:"difficulty"`
}

// blockData is used for hash calculation — excludes the Hash field itself.
type blockData struct {
	Index        uint64         `json:"index"`
	Timestamp    int64          `json:"timestamp"`
	Transactions []*Transaction `json:"transactions"`
	PrevHash     string         `json:"prev_hash"`
	Nonce        uint64         `json:"nonce"`
	MinerAddress string         `json:"miner_address"`
	Difficulty   int            `json:"difficulty"`
}

// NewBlock creates a new unmined block.
func NewBlock(index uint64, transactions []*Transaction, prevHash, minerAddress string, difficulty int) *Block {
	return &Block{
		Index:        index,
		Timestamp:    time.Now().UnixNano(),
		Transactions: transactions,
		PrevHash:     prevHash,
		Nonce:        0,
		MinerAddress: minerAddress,
		Difficulty:   difficulty,
	}
}

// CalculateHash computes the SHA-256 hash of the block (excluding the Hash field).
func (b *Block) CalculateHash() string {
	data := blockData{
		Index:        b.Index,
		Timestamp:    b.Timestamp,
		Transactions: b.Transactions,
		PrevHash:     b.PrevHash,
		Nonce:        b.Nonce,
		MinerAddress: b.MinerAddress,
		Difficulty:   b.Difficulty,
	}
	bytes, _ := json.Marshal(data)
	return bcrypto.HashData(bytes)
}

// Mine performs proof-of-work mining on the block.
// It increments the nonce until the hash has the required number of leading zeros.
// The done channel can be used to stop mining externally.
func (b *Block) Mine(done <-chan struct{}) bool {
	target := strings.Repeat("0", b.Difficulty)
	for {
		select {
		case <-done:
			return false // Mining was cancelled
		default:
			b.Hash = b.CalculateHash()
			if strings.HasPrefix(b.Hash, target) {
				return true // Found a valid hash
			}
			b.Nonce++
		}
	}
}

// HasValidPoW checks if the block hash satisfies the proof-of-work requirement.
// It verifies the hash has at least b.Difficulty leading zeros.
// Note: Validate() enforces that b.Difficulty equals the network's DefaultDifficulty.
func (b *Block) HasValidPoW() bool {
	target := strings.Repeat("0", b.Difficulty)
	return strings.HasPrefix(b.Hash, target)
}

// Validate checks the block against the previous block.
func (b *Block) Validate(prevBlock *Block) error {
	// Check index
	if b.Index != prevBlock.Index+1 {
		return fmt.Errorf("invalid block index: got %d, want %d", b.Index, prevBlock.Index+1)
	}

	// Check previous hash link
	if b.PrevHash != prevBlock.Hash {
		return fmt.Errorf("invalid previous hash: got %s, want %s", b.PrevHash, prevBlock.Hash)
	}

	// Check hash is correctly computed
	computed := b.CalculateHash()
	if b.Hash != computed {
		return fmt.Errorf("invalid block hash: got %s, computed %s", b.Hash, computed)
	}

	// Enforce network difficulty
	if b.Difficulty != DefaultDifficulty {
		return fmt.Errorf("invalid block difficulty: got %d, want %d", b.Difficulty, DefaultDifficulty)
	}

	// Check proof of work
	if !b.HasValidPoW() {
		return fmt.Errorf("block does not satisfy proof of work (difficulty %d)", b.Difficulty)
	}

	// Validate all transactions
	for i, tx := range b.Transactions {
		if tx.IsCoinbase() {
			if i != 0 {
				return fmt.Errorf("coinbase transaction must be first in block")
			}
			continue
		}
		if err := tx.Verify(); err != nil {
			return fmt.Errorf("invalid transaction %s at index %d: %w", tx.ID, i, err)
		}
	}

	return nil
}

// ValidateGenesis checks if a block is a valid genesis block.
func (b *Block) ValidateGenesis() error {
	if b.Index != 0 {
		return fmt.Errorf("genesis block must have index 0")
	}
	if b.PrevHash != "" {
		return fmt.Errorf("genesis block must have empty previous hash")
	}
	computed := b.CalculateHash()
	if b.Hash != computed {
		return fmt.Errorf("invalid genesis block hash: got %s, computed %s", b.Hash, computed)
	}
	if b.Difficulty != 1 {
		return fmt.Errorf("invalid genesis difficulty: got %d, want 1", b.Difficulty)
	}
	return nil
}
