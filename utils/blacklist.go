package utils

import (
	"sync"
	"time"
)

// TokenBlacklist stores blacklisted tokens in memory
type TokenBlacklist struct {
	tokens map[string]time.Time
	mu     sync.RWMutex
}

var blacklist = &TokenBlacklist{
	tokens: make(map[string]time.Time),
}

// AddToBlacklist adds a token to the blacklist with its expiration time
func AddToBlacklist(token string, expiresAt time.Time) {
	blacklist.mu.Lock()
	defer blacklist.mu.Unlock()
	blacklist.tokens[token] = expiresAt
}

// IsBlacklisted checks if a token is in the blacklist
func IsBlacklisted(token string) bool {
	blacklist.mu.RLock()
	defer blacklist.mu.RUnlock()

	expiresAt, exists := blacklist.tokens[token]
	if !exists {
		return false
	}

	// Check if the token has expired, if so remove it from blacklist
	if time.Now().After(expiresAt) {
		blacklist.mu.RUnlock()
		blacklist.mu.Lock()
		delete(blacklist.tokens, token)
		blacklist.mu.Unlock()
		blacklist.mu.RLock()
		return false
	}

	return true
}

// CleanupExpiredTokens removes expired tokens from the blacklist
// Should be called periodically
func CleanupExpiredTokens() {
	blacklist.mu.Lock()
	defer blacklist.mu.Unlock()

	now := time.Now()
	for token, expiresAt := range blacklist.tokens {
		if now.After(expiresAt) {
			delete(blacklist.tokens, token)
		}
	}
}

// StartCleanupRoutine starts a background goroutine to cleanup expired tokens
func StartCleanupRoutine() {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			CleanupExpiredTokens()
		}
	}()
}
