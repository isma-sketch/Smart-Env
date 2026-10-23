package generator

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

// GenerateSecret generates a secure random hex string of given byte length.
func GenerateSecret(bytes int) string {
	if bytes <= 0 {
		bytes = 32
	}
	b := make([]byte, bytes)
	if _, err := rand.Read(b); err != nil {
		// fallback
		return "change-me-" + fmt.Sprintf("%d", bytes)
	}
	return hex.EncodeToString(b)
}

// GenerateFor suggests a value for a key.
func GenerateFor(key string) string {
	u := strings.ToUpper(key)
	if strings.Contains(u, "JWT") {
		return GenerateSecret(32)
	}
	if strings.Contains(u, "PASSWORD") || strings.Contains(u, "PASS") {
		return GenerateSecret(16)
	}
	return GenerateSecret(24)
}

// IsGeneratable reports if key is a secret-like key.
func IsGeneratable(key string) bool {
	u := strings.ToUpper(key)
	for _, kw := range []string{"SECRET", "KEY", "PASSWORD", "PASS", "JWT", "TOKEN", "API_KEY", "PRIVATE"} {
		if strings.Contains(u, kw) {
			return true
		}
	}
	return false
}
