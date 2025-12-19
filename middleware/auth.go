package middleware

import (
	"os"
	"strings"
	helper "uas-backend/helper"
	utils "uas-backend/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func RequirePermission(requiredPermissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, "tidak ada authorization header")
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, "invalid authorization header format")
		}

		tokenString := parts[1]

		if utils.IsBlacklisted(tokenString) {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, "token telah di-logout")
		}

		secretkey := os.Getenv("JWT_SECRET_KEY")
		claims := &utils.JwtCustomClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) { return []byte(secretkey), nil })

		if err != nil || !token.Valid {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, "token invalid atau expired")
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("role_id", claims.RoleID)
		c.Locals("role_name", claims.RoleName)
		c.Locals("permissions", claims.Permissions)

		// If no permissions required, just check if user is authenticated
		if len(requiredPermissions) == 0 {
			return c.Next()
		}

		userPermissions := make(map[string]bool)
		for _, perm := range claims.Permissions {
			userPermissions[perm] = true
		}

		for _, requiredPerm := range requiredPermissions {
			if !userPermissions[requiredPerm] {
				return helper.ErrorResponse(c, fiber.StatusForbidden, "anda tidak memiliki akses untuk resource ini")
			}
		}

		return c.Next()
	}
}
