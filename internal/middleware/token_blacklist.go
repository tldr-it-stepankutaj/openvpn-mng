package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// TokenBlacklist tracks invalidated JWT tokens until they expire naturally
type TokenBlacklist struct {
	mu     sync.RWMutex
	tokens map[string]time.Time // SHA-256 hash of token → expiry time
	stopCh chan struct{}
}

// NewTokenBlacklist creates a new token blacklist and starts the cleanup goroutine
func NewTokenBlacklist() *TokenBlacklist {
	b := &TokenBlacklist{
		tokens: make(map[string]time.Time),
		stopCh: make(chan struct{}),
	}
	go b.cleanupLoop()
	return b
}

// Add blacklists a token until the given expiry time
func (b *TokenBlacklist) Add(tokenString string, expiry time.Time) {
	hash := hashToken(tokenString)
	b.mu.Lock()
	b.tokens[hash] = expiry
	b.mu.Unlock()
}

// IsBlacklisted checks if a token has been blacklisted
func (b *TokenBlacklist) IsBlacklisted(tokenString string) bool {
	hash := hashToken(tokenString)
	b.mu.RLock()
	defer b.mu.RUnlock()
	_, exists := b.tokens[hash]
	return exists
}

// Stop stops the cleanup goroutine
func (b *TokenBlacklist) Stop() {
	close(b.stopCh)
}

func (b *TokenBlacklist) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			b.cleanup()
		case <-b.stopCh:
			return
		}
	}
}

func (b *TokenBlacklist) cleanup() {
	now := time.Now()
	b.mu.Lock()
	defer b.mu.Unlock()
	for hash, expiry := range b.tokens {
		if expiry.Before(now) {
			delete(b.tokens, hash)
		}
	}
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
