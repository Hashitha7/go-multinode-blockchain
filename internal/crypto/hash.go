package crypto

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashBytes returns the SHA-256 hash of the given data.
func HashBytes(data []byte) [32]byte {
	return sha256.Sum256(data)
}

// HashToHex returns the hexadecimal string representation of a SHA-256 hash.
func HashToHex(hash [32]byte) string {
	return hex.EncodeToString(hash[:])
}

// HashData hashes arbitrary data and returns the hex string.
func HashData(data []byte) string {
	h := HashBytes(data)
	return HashToHex(h)
}
