package utils

import (
	"testing"
	"time"
)

func TestAddToBlacklist(t *testing.T) {
	token := "test-token-123"
	expiresAt := time.Now().Add(time.Hour)

	AddToBlacklist(token, expiresAt)

	if !IsBlacklisted(token) {
		t.Error("Token should be blacklisted after AddToBlacklist")
	}
}

func TestIsBlacklisted(t *testing.T) {
	tests := []struct {
		name  string
		token string
		setup func(string)
		want  bool
	}{
		{
			name:  "Non-blacklisted token",
			token: "non-blacklisted-token",
			setup: func(token string) {},
			want:  false,
		},
		{
			name:  "Blacklisted token",
			token: "blacklisted-token",
			setup: func(token string) {
				AddToBlacklist(token, time.Now().Add(time.Hour))
			},
			want: true,
		},
		{
			name:  "Expired blacklisted token",
			token: "expired-token",
			setup: func(token string) {
				AddToBlacklist(token, time.Now().Add(-time.Hour)) // Expired 1 hour ago
			},
			want: false, // Should be removed from blacklist
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(tt.token)
			if got := IsBlacklisted(tt.token); got != tt.want {
				t.Errorf("IsBlacklisted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlacklistConcurrency(t *testing.T) {
	// Test concurrent access to blacklist
	done := make(chan bool)

	// Multiple goroutines adding tokens
	for i := 0; i < 10; i++ {
		go func(id int) {
			token := time.Now().String() + string(rune(id))
			AddToBlacklist(token, time.Now().Add(time.Hour))
			IsBlacklisted(token)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestCleanupExpiredTokens(t *testing.T) {
	// Add some expired and non-expired tokens
	expiredToken1 := "expired-1"
	expiredToken2 := "expired-2"
	validToken := "valid-token"

	AddToBlacklist(expiredToken1, time.Now().Add(-time.Hour))
	AddToBlacklist(expiredToken2, time.Now().Add(-time.Minute))
	AddToBlacklist(validToken, time.Now().Add(time.Hour))

	// Run cleanup
	CleanupExpiredTokens()

	// Valid token should still be blacklisted
	if !IsBlacklisted(validToken) {
		t.Error("Valid token should still be blacklisted after cleanup")
	}

	// Expired tokens should be removed (IsBlacklisted checks expiry)
	if IsBlacklisted(expiredToken1) {
		t.Error("Expired token 1 should not be blacklisted after cleanup")
	}
	if IsBlacklisted(expiredToken2) {
		t.Error("Expired token 2 should not be blacklisted after cleanup")
	}
}

func TestBlacklistMultipleTokens(t *testing.T) {
	tokens := []string{
		"token-1",
		"token-2",
		"token-3",
	}
	expiresAt := time.Now().Add(time.Hour)

	// Add multiple tokens
	for _, token := range tokens {
		AddToBlacklist(token, expiresAt)
	}

	// Verify all are blacklisted
	for _, token := range tokens {
		if !IsBlacklisted(token) {
			t.Errorf("Token %s should be blacklisted", token)
		}
	}
}

func TestBlacklistSameTokenTwice(t *testing.T) {
	token := "duplicate-token"
	firstExpiry := time.Now().Add(time.Hour)
	secondExpiry := time.Now().Add(2 * time.Hour)

	// Add same token twice with different expiry times
	AddToBlacklist(token, firstExpiry)
	AddToBlacklist(token, secondExpiry)

	if !IsBlacklisted(token) {
		t.Error("Token should be blacklisted")
	}
}

func TestBlacklistEmptyToken(t *testing.T) {
	token := ""
	expiresAt := time.Now().Add(time.Hour)

	AddToBlacklist(token, expiresAt)

	if !IsBlacklisted(token) {
		t.Error("Empty token should be blacklisted if added")
	}
}

func TestBlacklistExpiryBoundary(t *testing.T) {
	token := "boundary-token"
	// Set expiry to very close to now
	expiresAt := time.Now().Add(10 * time.Millisecond)

	AddToBlacklist(token, expiresAt)

	// Should be blacklisted immediately
	if !IsBlacklisted(token) {
		t.Error("Token should be blacklisted immediately after adding")
	}

	// Wait for expiry
	time.Sleep(20 * time.Millisecond)

	// Should no longer be blacklisted
	if IsBlacklisted(token) {
		t.Error("Token should not be blacklisted after expiry")
	}
}
