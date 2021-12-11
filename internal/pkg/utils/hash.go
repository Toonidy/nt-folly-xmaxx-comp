package utils

import (
	"crypto/sha256"
	"fmt"
)

// HashData generates a sha256 checksum of given data.
func HashData(data []byte) ([]byte, error) {
	h := sha256.New()
	_, err := h.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate md5 hash: %w", err)
	}
	return h.Sum(nil), nil
}
