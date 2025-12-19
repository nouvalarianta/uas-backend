package utils

import (
	"os"
	"testing"
	"time"
	"uas-backend/app/model"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestGenerateToken(t *testing.T) {
	// Setup
	os.Setenv("JWT_SECRET_KEY", "test-secret-key")
	os.Setenv("JWT_EXPIRATION_MINUTE", "60")
	defer os.Unsetenv("JWT_SECRET_KEY")
	defer os.Unsetenv("JWT_EXPIRATION_MINUTE")

	roleID := uuid.New()
	userID := uuid.New()
	user := model.User{
		ID:       userID,
		Username: "testuser",
		RoleID:   roleID,
		Role: &model.Role{
			ID:   roleID,
			Name: "Admin",
		},
	}
	permissions := []string{"user:read", "user:write"}

	tests := []struct {
		name        string
		user        model.User
		permissions []string
		wantErr     bool
	}{
		{
			name:        "Valid token generation",
			user:        user,
			permissions: permissions,
			wantErr:     false,
		},
		{
			name:        "Token with no permissions",
			user:        user,
			permissions: []string{},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateToken(tt.user, tt.permissions)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && token == "" {
				t.Error("GenerateToken() returned empty token")
			}
		})
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	// Setup
	os.Setenv("JWT_SECRET_KEY", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET_KEY")

	roleID := uuid.New()
	userID := uuid.New()
	user := model.User{
		ID:       userID,
		Username: "testuser",
		RoleID:   roleID,
		Role: &model.Role{
			ID:   roleID,
			Name: "Admin",
		},
	}
	permissions := []string{"user:read", "user:write"}

	token, err := GenerateRefreshToken(user, permissions)
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}
	if token == "" {
		t.Error("GenerateRefreshToken() returned empty token")
	}
}

func TestParseToken(t *testing.T) {
	// Setup
	os.Setenv("JWT_SECRET_KEY", "test-secret-key")
	os.Setenv("JWT_EXPIRATION_MINUTE", "60")
	defer os.Unsetenv("JWT_SECRET_KEY")
	defer os.Unsetenv("JWT_EXPIRATION_MINUTE")

	roleID := uuid.New()
	userID := uuid.New()
	user := model.User{
		ID:       userID,
		Username: "testuser",
		RoleID:   roleID,
		Role: &model.Role{
			ID:   roleID,
			Name: "Admin",
		},
	}
	permissions := []string{"user:read", "user:write"}

	validToken, err := GenerateToken(user, permissions)
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	tests := []struct {
		name      string
		token     string
		wantErr   bool
		checkFunc func(*testing.T, *JwtCustomClaims)
	}{
		{
			name:    "Valid token",
			token:   validToken,
			wantErr: false,
			checkFunc: func(t *testing.T, claims *JwtCustomClaims) {
				if claims.UserID != userID.String() {
					t.Errorf("Expected UserID %s, got %s", userID.String(), claims.UserID)
				}
				if claims.RoleID != roleID.String() {
					t.Errorf("Expected RoleID %s, got %s", roleID.String(), claims.RoleID)
				}
				if claims.RoleName != "Admin" {
					t.Errorf("Expected RoleName Admin, got %s", claims.RoleName)
				}
				if len(claims.Permissions) != 2 {
					t.Errorf("Expected 2 permissions, got %d", len(claims.Permissions))
				}
			},
		},
		{
			name:    "Invalid token format",
			token:   "invalid.token.format",
			wantErr: true,
		},
		{
			name:    "Empty token",
			token:   "",
			wantErr: true,
		},
		{
			name:    "Malformed token",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.malformed",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ParseToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && claims == nil {
				t.Error("ParseToken() returned nil claims for valid token")
			}
			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, claims)
			}
		})
	}
}

func TestTokenExpiration(t *testing.T) {
	// Setup
	os.Setenv("JWT_SECRET_KEY", "test-secret-key")
	os.Setenv("JWT_EXPIRATION_MINUTE", "0") // Expire immediately
	defer os.Unsetenv("JWT_SECRET_KEY")
	defer os.Unsetenv("JWT_EXPIRATION_MINUTE")

	roleID := uuid.New()
	userID := uuid.New()
	user := model.User{
		ID:       userID,
		Username: "testuser",
		RoleID:   roleID,
		Role: &model.Role{
			ID:   roleID,
			Name: "Admin",
		},
	}
	permissions := []string{"user:read"}

	token, err := GenerateToken(user, permissions)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Wait a bit to ensure expiration
	time.Sleep(100 * time.Millisecond)

	_, err = ParseToken(token)
	if err == nil {
		t.Error("ParseToken() should return error for expired token")
	}
}

func TestTokenWithDifferentSecretKey(t *testing.T) {
	// Generate token with one secret
	os.Setenv("JWT_SECRET_KEY", "secret-key-1")
	os.Setenv("JWT_EXPIRATION_MINUTE", "60")

	roleID := uuid.New()
	userID := uuid.New()
	user := model.User{
		ID:       userID,
		Username: "testuser",
		RoleID:   roleID,
		Role: &model.Role{
			ID:   roleID,
			Name: "Admin",
		},
	}
	permissions := []string{"user:read"}

	token, err := GenerateToken(user, permissions)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Try to parse with different secret
	os.Setenv("JWT_SECRET_KEY", "secret-key-2")
	defer os.Unsetenv("JWT_SECRET_KEY")
	defer os.Unsetenv("JWT_EXPIRATION_MINUTE")

	_, err = ParseToken(token)
	if err == nil {
		t.Error("ParseToken() should return error when secret key is different")
	}
}

func TestJwtCustomClaimsStructure(t *testing.T) {
	claims := &JwtCustomClaims{
		UserID:      "user-id-123",
		RoleID:      "role-id-456",
		RoleName:    "TestRole",
		Permissions: []string{"read", "write"},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	if claims.UserID != "user-id-123" {
		t.Errorf("Expected UserID user-id-123, got %s", claims.UserID)
	}
	if claims.RoleID != "role-id-456" {
		t.Errorf("Expected RoleID role-id-456, got %s", claims.RoleID)
	}
	if claims.RoleName != "TestRole" {
		t.Errorf("Expected RoleName TestRole, got %s", claims.RoleName)
	}
	if len(claims.Permissions) != 2 {
		t.Errorf("Expected 2 permissions, got %d", len(claims.Permissions))
	}
}

func TestRefreshTokenExpiration(t *testing.T) {
	os.Setenv("JWT_SECRET_KEY", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET_KEY")

	roleID := uuid.New()
	userID := uuid.New()
	user := model.User{
		ID:       userID,
		Username: "testuser",
		RoleID:   roleID,
		Role: &model.Role{
			ID:   roleID,
			Name: "Admin",
		},
	}
	permissions := []string{"user:read"}

	refreshToken, err := GenerateRefreshToken(user, permissions)
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	claims, err := ParseToken(refreshToken)
	if err != nil {
		t.Fatalf("Failed to parse refresh token: %v", err)
	}

	// Check that expiration is approximately 7 days from now
	expectedExpiry := time.Now().Add(7 * 24 * time.Hour)
	actualExpiry := claims.ExpiresAt.Time

	diff := actualExpiry.Sub(expectedExpiry).Abs()
	if diff > time.Minute {
		t.Errorf("Refresh token expiry is not around 7 days: expected %v, got %v", expectedExpiry, actualExpiry)
	}
}
