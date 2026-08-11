package blockchain

import (
	"testing"

	"github.com/blockchain2/internal/crypto"
)

func TestNewTransaction(t *testing.T) {
	pub, priv, _ := crypto.GenerateKeyPair()
	to := "recipient_address_1234567890abcdef1234"

	tx := NewTransaction(to, 100, 1, pub, priv)

	if tx.ID == "" {
		t.Error("Transaction ID should not be empty")
	}
	if tx.From == "" {
		t.Error("From address should not be empty")
	}
	if tx.To != to {
		t.Errorf("To address mismatch: got %s, want %s", tx.To, to)
	}
	if tx.Amount != 100 {
		t.Errorf("Amount mismatch: got %d, want 100", tx.Amount)
	}
	if tx.Fee != 1 {
		t.Errorf("Fee mismatch: got %d, want 1", tx.Fee)
	}
	if tx.PubKey == "" {
		t.Error("PubKey should not be empty")
	}
	if tx.Signature == "" {
		t.Error("Signature should not be empty")
	}
}

func TestTransactionVerify(t *testing.T) {
	pub, priv, _ := crypto.GenerateKeyPair()
	tx := NewTransaction("recipient", 50, 1, pub, priv)

	// Should verify successfully
	if err := tx.Verify(); err != nil {
		t.Errorf("Valid transaction failed verification: %v", err)
	}
}

func TestTransactionVerifyInvalidSignature(t *testing.T) {
	pub, priv, _ := crypto.GenerateKeyPair()
	tx := NewTransaction("recipient", 50, 1, pub, priv)

	// Tamper with amount
	tx.Amount = 999

	// Should fail verification
	if err := tx.Verify(); err == nil {
		t.Error("Tampered transaction should fail verification")
	}
}

func TestTransactionVerifyWrongKey(t *testing.T) {
	pub, priv, _ := crypto.GenerateKeyPair()
	tx := NewTransaction("recipient", 50, 1, pub, priv)

	// Replace pubkey with a different key
	pub2, _, _ := crypto.GenerateKeyPair()
	tx.PubKey = crypto.PublicKeyToHex(pub2)

	// Should fail (address won't match)
	if err := tx.Verify(); err == nil {
		t.Error("Transaction with wrong key should fail verification")
	}
}

func TestCoinbaseTransaction(t *testing.T) {
	cb := NewCoinbaseTransaction("miner_addr", 50)

	if !cb.IsCoinbase() {
		t.Error("Coinbase transaction should be identified as coinbase")
	}
	if cb.To != "miner_addr" {
		t.Errorf("Coinbase To mismatch: got %s, want miner_addr", cb.To)
	}
	if cb.Amount != 50 {
		t.Errorf("Coinbase Amount mismatch: got %d, want 50", cb.Amount)
	}
}

func TestGenesisBlock(t *testing.T) {
	genesis := GenesisBlock()

	if genesis.Index != 0 {
		t.Errorf("Genesis index should be 0, got %d", genesis.Index)
	}
	if genesis.PrevHash != "" {
		t.Errorf("Genesis prev_hash should be empty, got %s", genesis.PrevHash)
	}
	if genesis.Hash == "" {
		t.Error("Genesis hash should not be empty")
	}

	// Should be deterministic
	genesis2 := GenesisBlock()
	if genesis.Hash != genesis2.Hash {
		t.Error("Genesis block should be deterministic")
	}

	// Should pass validation
	if err := genesis.ValidateGenesis(); err != nil {
		t.Errorf("Genesis block validation failed: %v", err)
	}
}

func TestBlockMining(t *testing.T) {
	genesis := GenesisBlock()

	block := NewBlock(1, []*Transaction{}, genesis.Hash, "miner", 1) // Low difficulty for test
	done := make(chan struct{})
	defer close(done)

	success := block.Mine(done)
	if !success {
		t.Error("Mining should succeed")
	}
	if block.Hash == "" {
		t.Error("Block hash should not be empty after mining")
	}
	if !block.HasValidPoW() {
		t.Error("Mined block should have valid proof of work")
	}
}

func TestBlockValidation(t *testing.T) {
	genesis := GenesisBlock()

	block := NewBlock(1, []*Transaction{}, genesis.Hash, "miner", 1)
	done := make(chan struct{})
	defer close(done)
	block.Mine(done)

	// Should validate against genesis
	if err := block.Validate(genesis); err != nil {
		t.Errorf("Valid block failed validation: %v", err)
	}
}

func TestBlockValidationInvalidIndex(t *testing.T) {
	genesis := GenesisBlock()

	block := NewBlock(5, []*Transaction{}, genesis.Hash, "miner", 1) // Wrong index
	done := make(chan struct{})
	defer close(done)
	block.Mine(done)

	if err := block.Validate(genesis); err == nil {
		t.Error("Block with wrong index should fail validation")
	}
}

func TestBlockValidationInvalidPrevHash(t *testing.T) {
	genesis := GenesisBlock()

	block := NewBlock(1, []*Transaction{}, "wrong_hash", "miner", 1)
	done := make(chan struct{})
	defer close(done)
	block.Mine(done)

	if err := block.Validate(genesis); err == nil {
		t.Error("Block with wrong prev_hash should fail validation")
	}
}

func TestChainBasicOperations(t *testing.T) {
	chain := NewChain()

	if chain.GetHeight() != 0 {
		t.Errorf("New chain should have height 0, got %d", chain.GetHeight())
	}

	if chain.Length() != 1 {
		t.Errorf("New chain should have 1 block (genesis), got %d", chain.Length())
	}

	genesis := chain.GetBlockByIndex(0)
	if genesis == nil {
		t.Fatal("Genesis block should not be nil")
	}
}

func TestChainAddBlock(t *testing.T) {
	chain := NewChain()
	genesis := chain.GetLatestBlock()

	block := NewBlock(1, []*Transaction{
		NewCoinbaseTransaction("miner", MiningReward),
	}, genesis.Hash, "miner", 1)
	done := make(chan struct{})
	defer close(done)
	block.Mine(done)

	err := chain.AddBlock(block)
	if err != nil {
		t.Errorf("Failed to add valid block: %v", err)
	}

	if chain.GetHeight() != 1 {
		t.Errorf("Chain height should be 1, got %d", chain.GetHeight())
	}
}

func TestChainReorganisation(t *testing.T) {
	// Build chain A: genesis -> block1A
	chainA := NewChain()
	genesisA := chainA.GetLatestBlock()
	block1A := NewBlock(1, []*Transaction{
		NewCoinbaseTransaction("minerA", MiningReward),
	}, genesisA.Hash, "minerA", 1)
	done := make(chan struct{})
	defer close(done)
	block1A.Mine(done)
	chainA.AddBlock(block1A)

	// Build chain B: genesis -> block1B -> block2B (longer)
	genesis := GenesisBlock()
	block1B := NewBlock(1, []*Transaction{
		NewCoinbaseTransaction("minerB", MiningReward),
	}, genesis.Hash, "minerB", 1)
	block1B.Mine(done)

	block2B := NewBlock(2, []*Transaction{
		NewCoinbaseTransaction("minerB", MiningReward),
	}, block1B.Hash, "minerB", 1)
	block2B.Mine(done)

	newBlocks := []*Block{genesis, block1B, block2B}

	// Reorg chain A to chain B
	orphanedTxs, err := chainA.ReorganiseTo(newBlocks)
	if err != nil {
		t.Fatalf("Reorg failed: %v", err)
	}

	if chainA.GetHeight() != 2 {
		t.Errorf("After reorg, height should be 2, got %d", chainA.GetHeight())
	}

	// Coinbase transactions are not returned as orphaned
	for _, tx := range orphanedTxs {
		if tx.IsCoinbase() {
			t.Error("Coinbase transactions should not be in orphaned list")
		}
	}
}

func TestLedgerBalances(t *testing.T) {
	ledger := NewLedger()

	genesis := GenesisBlock()
	block1 := &Block{
		Index:        1,
		Transactions: []*Transaction{NewCoinbaseTransaction("miner1", MiningReward)},
	}

	blocks := []*Block{genesis, block1}
	err := ledger.RebuildFromChain(blocks)
	if err != nil {
		t.Fatalf("Failed to rebuild ledger: %v", err)
	}

	balance := ledger.GetBalance("miner1")
	if balance != MiningReward {
		t.Errorf("Miner balance should be %d, got %d", MiningReward, balance)
	}
}

func TestChainGetBlocksAfter(t *testing.T) {
	chain := NewChain()
	genesis := chain.GetLatestBlock()

	// Add a couple of blocks
	block1 := NewBlock(1, []*Transaction{
		NewCoinbaseTransaction("miner", MiningReward),
	}, genesis.Hash, "miner", 1)
	done := make(chan struct{})
	defer close(done)
	block1.Mine(done)
	chain.AddBlock(block1)

	block2 := NewBlock(2, []*Transaction{
		NewCoinbaseTransaction("miner", MiningReward),
	}, block1.Hash, "miner", 1)
	block2.Mine(done)
	chain.AddBlock(block2)

	// Get blocks after genesis
	blocks := chain.GetBlocksAfter(0)
	if len(blocks) != 2 {
		t.Errorf("Expected 2 blocks after genesis, got %d", len(blocks))
	}

	// Get blocks after block 1
	blocks = chain.GetBlocksAfter(1)
	if len(blocks) != 1 {
		t.Errorf("Expected 1 block after height 1, got %d", len(blocks))
	}

	// Get blocks after latest
	blocks = chain.GetBlocksAfter(2)
	if blocks != nil {
		t.Errorf("Expected no blocks after height 2, got %d", len(blocks))
	}
}
