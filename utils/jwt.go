package utils

import (
	"os"
	"strconv"
	"time"
	"uas-backend/app/model"

	"github.com/golang-jwt/jwt/v5"
)

type JwtCustomClaims struct {
	UserID      string   `json:"user_id"`
	RoleID      string   `json:"role_id"`
	RoleName    string   `json:"name"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

func GenerateToken(user model.User, permissions []string) (string, error) {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	expirationMinutesStr := os.Getenv("JWT_EXPIRATION_MINUTE")

	expirationMinutes, err := strconv.Atoi(expirationMinutesStr)
	if err != nil {
		expirationMinutes = 60
	}

	roleName := ""
	if user.Role != nil {
		roleName = user.Role.Name
	}

	claims := &JwtCustomClaims{
		UserID:      user.ID.String(),
		RoleID:      user.RoleID.String(),
		RoleName:    roleName,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * time.Duration(expirationMinutes))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secretKey))
}

func GenerateRefreshToken(user model.User, permissions []string) (string, error) {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	expirationDays := 7

	roleName := ""
	if user.Role != nil {
		roleName = user.Role.Name
	}

	claims := &JwtCustomClaims{
		UserID:      user.ID.String(),
		RoleID:      user.RoleID.String(),
		RoleName:    roleName,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * time.Duration(expirationDays))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secretKey))
}

func ParseToken(tokenString string) (*JwtCustomClaims, error) {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	claims := &JwtCustomClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	return claims, nil
}
