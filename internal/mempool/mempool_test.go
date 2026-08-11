package mempool

import (
	"testing"

	"github.com/blockchain2/internal/blockchain"
	"github.com/blockchain2/internal/crypto"
)

func createTestTx(t *testing.T) *blockchain.Transaction {
	t.Helper()
	pub, priv, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}
	return blockchain.NewTransaction("recipient_address_1234567890abcdef1234", 10, 1, pub, priv)
}

func TestPoolAddAndSize(t *testing.T) {
	pool := NewPool()

	if pool.Size() != 0 {
		t.Errorf("New pool should have size 0, got %d", pool.Size())
	}

	tx := createTestTx(t)
	added := pool.Add(tx)
	if !added {
		t.Error("First add should return true")
	}

	if pool.Size() != 1 {
		t.Errorf("Pool should have size 1, got %d", pool.Size())
	}
}

func TestPoolDeDuplication(t *testing.T) {
	pool := NewPool()
	tx := createTestTx(t)

	pool.Add(tx)
	added := pool.Add(tx)
	if added {
		t.Error("Duplicate add should return false")
	}

	if pool.Size() != 1 {
		t.Errorf("Pool should still have size 1, got %d", pool.Size())
	}
}

func TestPoolContains(t *testing.T) {
	pool := NewPool()
	tx := createTestTx(t)

	if pool.Contains(tx.ID) {
		t.Error("Pool should not contain transaction before add")
	}

	pool.Add(tx)

	if !pool.Contains(tx.ID) {
		t.Error("Pool should contain transaction after add")
	}
}

func TestPoolRemove(t *testing.T) {
	pool := NewPool()
	tx := createTestTx(t)
	pool.Add(tx)

	pool.Remove([]string{tx.ID})

	if pool.Size() != 0 {
		t.Errorf("Pool should be empty after remove, got %d", pool.Size())
	}
}

func TestPoolRemoveByBlock(t *testing.T) {
	pool := NewPool()
	tx := createTestTx(t)
	pool.Add(tx)

	block := &blockchain.Block{
		Transactions: []*blockchain.Transaction{
			blockchain.NewCoinbaseTransaction("miner", 50),
			tx,
		},
	}

	pool.RemoveByBlock(block)

	if pool.Size() != 0 {
		t.Errorf("Pool should be empty after block removal, got %d", pool.Size())
	}
}

func TestPoolGetPending(t *testing.T) {
	pool := NewPool()

	tx1 := createTestTx(t)
	tx2 := createTestTx(t)
	pool.Add(tx1)
	pool.Add(tx2)

	pending := pool.GetPending()
	if len(pending) != 2 {
		t.Errorf("Expected 2 pending, got %d", len(pending))
	}
}

func TestPoolReturnOrphaned(t *testing.T) {
	pool := NewPool()
	tx := createTestTx(t)

	pool.ReturnOrphaned([]*blockchain.Transaction{tx})

	if pool.Size() != 1 {
		t.Errorf("Pool should have 1 orphaned tx, got %d", pool.Size())
	}
}

func TestPoolClear(t *testing.T) {
	pool := NewPool()
	pool.Add(createTestTx(t))
	pool.Add(createTestTx(t))

	pool.Clear()

	if pool.Size() != 0 {
		t.Errorf("Pool should be empty after clear, got %d", pool.Size())
	}
}

// TestPoolConcurrentAccess tests thread safety of the pool.
func TestPoolConcurrentAccess(t *testing.T) {
	pool := NewPool()
	done := make(chan bool)

	// Concurrent adds
	for i := 0; i < 100; i++ {
		go func() {
			tx := createTestTx(t)
			pool.Add(tx)
			pool.Size()
			pool.GetPending()
			pool.Contains(tx.ID)
			done <- true
		}()
	}

	for i := 0; i < 100; i++ {
		<-done
	}

	if pool.Size() != 100 {
		t.Errorf("Expected 100 transactions, got %d", pool.Size())
	}
}
