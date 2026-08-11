package blockchain

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	bcrypto "github.com/blockchain2/internal/crypto"
)

// Transaction represents a signed transaction in the blockchain.
type Transaction struct {
	ID        string `json:"id"`        // SHA-256 hash of transaction content
	From      string `json:"from"`      // Sender address
	To        string `json:"to"`        // Recipient address
	Amount    uint64 `json:"amount"`    // Amount to transfer
	Fee       uint64 `json:"fee"`       // Transaction fee
	Timestamp int64  `json:"timestamp"` // Unix timestamp
	PubKey    string `json:"pubkey"`    // Sender's public key (hex)
	Signature string `json:"signature"` // Ed25519 signature (hex)
}

// txPayload is the data that gets signed (everything except ID and Signature).
type txPayload struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Amount    uint64 `json:"amount"`
	Fee       uint64 `json:"fee"`
	Timestamp int64  `json:"timestamp"`
	PubKey    string `json:"pubkey"`
}

// NewTransaction creates a new signed transaction.
func NewTransaction(to string, amount, fee uint64, pubKey ed25519.PublicKey, privKey ed25519.PrivateKey) *Transaction {
	from := bcrypto.AddressFromPublicKey(pubKey)
	tx := &Transaction{
		From:      from,
		To:        to,
		Amount:    amount,
		Fee:       fee,
		Timestamp: time.Now().UnixNano(),
		PubKey:    bcrypto.PublicKeyToHex(pubKey),
	}

	// Sign the transaction
	payload := tx.payloadBytes()
	tx.Signature = hex.EncodeToString(bcrypto.Sign(privKey, payload))

	// Compute transaction ID
	tx.ID = tx.computeID()

	return tx
}

// payloadBytes returns the serialized payload for signing.
func (tx *Transaction) payloadBytes() []byte {
	p := txPayload{
		From:      tx.From,
		To:        tx.To,
		Amount:    tx.Amount,
		Fee:       tx.Fee,
		Timestamp: tx.Timestamp,
		PubKey:    tx.PubKey,
	}
	data, _ := json.Marshal(p)
	return data
}

// computeID computes the transaction ID from its full content.
func (tx *Transaction) computeID() string {
	data := struct {
		From      string `json:"from"`
		To        string `json:"to"`
		Amount    uint64 `json:"amount"`
		Fee       uint64 `json:"fee"`
		Timestamp int64  `json:"timestamp"`
		PubKey    string `json:"pubkey"`
		Signature string `json:"signature"`
	}{
		From:      tx.From,
		To:        tx.To,
		Amount:    tx.Amount,
		Fee:       tx.Fee,
		Timestamp: tx.Timestamp,
		PubKey:    tx.PubKey,
		Signature: tx.Signature,
	}
	bytes, _ := json.Marshal(data)
	return bcrypto.HashData(bytes)
}

// Verify checks that the transaction signature is valid and the address matches the public key.
func (tx *Transaction) Verify() error {
	// Decode public key
	pubKey, err := bcrypto.HexToPublicKey(tx.PubKey)
	if err != nil {
		return fmt.Errorf("invalid public key: %w", err)
	}

	// Verify the sender address matches the public key
	expectedAddr := bcrypto.AddressFromPublicKey(pubKey)
	if tx.From != expectedAddr {
		return fmt.Errorf("sender address does not match public key: got %s, want %s", tx.From, expectedAddr)
	}

	// Decode signature
	sig, err := hex.DecodeString(tx.Signature)
	if err != nil {
		return fmt.Errorf("invalid signature hex: %w", err)
	}

	// Verify signature
	payload := tx.payloadBytes()
	if !bcrypto.Verify(pubKey, payload, sig) {
		return fmt.Errorf("invalid signature")
	}

	// Basic validation
	if tx.Amount == 0 {
		return fmt.Errorf("transaction amount must be positive")
	}

	return nil
}

// IsCoinbase returns true if this is a mining reward transaction (no sender).
func (tx *Transaction) IsCoinbase() bool {
	return tx.From == "" && tx.PubKey == "" && tx.Signature == ""
}

// NewCoinbaseTransaction creates a mining reward transaction.
func NewCoinbaseTransaction(to string, reward uint64) *Transaction {
	tx := &Transaction{
		From:      "",
		To:        to,
		Amount:    reward,
		Fee:       0,
		Timestamp: time.Now().UnixNano(),
		PubKey:    "",
		Signature: "",
	}
	tx.ID = tx.computeID()
	return tx
}
