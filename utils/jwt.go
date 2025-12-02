package utils

import (
	"os"
	"strconv"
	"time"
	"uas-backend/app/model"

	"github.com/golang-jwt/jwt/v5"
)

type JwtCustomClaims struct {
	UserID   string `json:"user_id"`
	RoleID   string `json:"role_id"`
	RoleName string `json:"name"`
	jwt.RegisteredClaims
}

func GenerateToken(user model.User) (string, error) {
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
		UserID:   user.ID.String(),
		RoleID:   user.RoleID.String(),
		RoleName: roleName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * time.Duration(expirationMinutes))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secretKey))
}
