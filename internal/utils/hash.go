package utils

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

// HashRequest creates a hash of the request data for idempotency
func HashRequest(data interface{}) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request data: %w", err)
	}

	hash := sha256.Sum256(jsonData)
	return fmt.Sprintf("%x", hash), nil
}
