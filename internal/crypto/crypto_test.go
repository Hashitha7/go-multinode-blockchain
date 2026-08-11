package crypto

import (
	"crypto/ed25519"
	"testing"
)

func TestHashBytes(t *testing.T) {
	data := []byte("hello blockchain")
	hash := HashBytes(data)
	hexStr := HashToHex(hash)

	if len(hexStr) != 64 {
		t.Errorf("Expected 64 char hex string, got %d", len(hexStr))
	}

	// Same input should produce same hash
	hash2 := HashBytes(data)
	hexStr2 := HashToHex(hash2)
	if hexStr != hexStr2 {
		t.Error("Same input produced different hashes")
	}

	// Different input should produce different hash
	hash3 := HashBytes([]byte("different data"))
	hexStr3 := HashToHex(hash3)
	if hexStr == hexStr3 {
		t.Error("Different inputs produced same hash")
	}
}

func TestHashData(t *testing.T) {
	result := HashData([]byte("test"))
	if len(result) != 64 {
		t.Errorf("Expected 64 char hex string, got %d", len(result))
	}
}

func TestGenerateKeyPair(t *testing.T) {
	pub, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	if len(pub) != ed25519.PublicKeySize {
		t.Errorf("Expected public key size %d, got %d", ed25519.PublicKeySize, len(pub))
	}

	if len(priv) != ed25519.PrivateKeySize {
		t.Errorf("Expected private key size %d, got %d", ed25519.PrivateKeySize, len(priv))
	}
}

func TestSignAndVerify(t *testing.T) {
	pub, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	message := []byte("test message")
	sig := Sign(priv, message)

	// Should verify with correct key and message
	if !Verify(pub, message, sig) {
		t.Error("Valid signature failed verification")
	}

	// Should fail with wrong message
	if Verify(pub, []byte("wrong message"), sig) {
		t.Error("Signature verified with wrong message")
	}

	// Should fail with wrong key
	pub2, _, _ := GenerateKeyPair()
	if Verify(pub2, message, sig) {
		t.Error("Signature verified with wrong public key")
	}
}

func TestAddressFromPublicKey(t *testing.T) {
	pub, _, _ := GenerateKeyPair()
	addr := AddressFromPublicKey(pub)

	if len(addr) != 40 {
		t.Errorf("Expected 40 char address, got %d", len(addr))
	}

	// Same key should produce same address
	addr2 := AddressFromPublicKey(pub)
	if addr != addr2 {
		t.Error("Same key produced different addresses")
	}
}

func TestPublicKeyHexRoundTrip(t *testing.T) {
	pub, _, _ := GenerateKeyPair()
	hexStr := PublicKeyToHex(pub)
	recovered, err := HexToPublicKey(hexStr)
	if err != nil {
		t.Fatalf("Failed to recover public key: %v", err)
	}

	if !pub.Equal(recovered) {
		t.Error("Recovered public key doesn't match original")
	}
}

func TestPrivateKeyHexRoundTrip(t *testing.T) {
	_, priv, _ := GenerateKeyPair()
	hexStr := PrivateKeyToHex(priv)
	recovered, err := HexToPrivateKey(hexStr)
	if err != nil {
		t.Fatalf("Failed to recover private key: %v", err)
	}

	if !priv.Equal(recovered) {
		t.Error("Recovered private key doesn't match original")
	}
}

func TestHexToPublicKeyInvalid(t *testing.T) {
	// Invalid hex
	_, err := HexToPublicKey("not_hex")
	if err == nil {
		t.Error("Expected error for invalid hex")
	}

	// Wrong length
	_, err = HexToPublicKey("abcd")
	if err == nil {
		t.Error("Expected error for wrong length")
	}
}
