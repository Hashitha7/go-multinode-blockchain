package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateKeyPair generates a new Ed25519 key pair.
func GenerateKeyPair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate key pair: %w", err)
	}
	return pub, priv, nil
}

// AddressFromPublicKey derives an address string from an Ed25519 public key.
// The address is the hex-encoded SHA-256 hash of the public key, truncated to 40 chars.
func AddressFromPublicKey(pub ed25519.PublicKey) string {
	hash := HashBytes(pub)
	return hex.EncodeToString(hash[:20]) // 40 hex chars, similar to Ethereum addresses
}

// Sign signs a message using the given Ed25519 private key.
func Sign(privateKey ed25519.PrivateKey, message []byte) []byte {
	return ed25519.Sign(privateKey, message)
}

// Verify checks an Ed25519 signature against a public key and message.
func Verify(publicKey ed25519.PublicKey, message []byte, signature []byte) bool {
	return ed25519.Verify(publicKey, message, signature)
}

// PublicKeyToHex converts an Ed25519 public key to a hex string.
func PublicKeyToHex(pub ed25519.PublicKey) string {
	return hex.EncodeToString(pub)
}

// HexToPublicKey converts a hex string back to an Ed25519 public key.
func HexToPublicKey(hexStr string) (ed25519.PublicKey, error) {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("invalid hex string: %w", err)
	}
	if len(bytes) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key length: got %d, want %d", len(bytes), ed25519.PublicKeySize)
	}
	return ed25519.PublicKey(bytes), nil
}

// PrivateKeyToHex converts an Ed25519 private key to a hex string.
func PrivateKeyToHex(priv ed25519.PrivateKey) string {
	return hex.EncodeToString(priv)
}

// HexToPrivateKey converts a hex string back to an Ed25519 private key.
func HexToPrivateKey(hexStr string) (ed25519.PrivateKey, error) {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("invalid hex string: %w", err)
	}
	if len(bytes) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key length: got %d, want %d", len(bytes), ed25519.PrivateKeySize)
	}
	return ed25519.PrivateKey(bytes), nil
}
