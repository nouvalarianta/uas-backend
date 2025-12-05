package middleware

import (
	"os"
	"strings"
	helper"uas-backend/helper"
	utils"uas-backend/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func Mahasiswa() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Tidak ada authorization header")
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid authorization header format")
		}

		tokenString := parts[1]
		secretkey := os.Getenv("JWT_SECRET_KEY")

		claims := &utils.JwtCustomClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretkey), nil
		})

		if err != nil || !token.Valid {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, "token invalid")
		}

		if claims.RoleName != "Mahasiswa" {
			return helper.ErrorResponse(c, fiber.StatusForbidden, "bukan Mahasiswa")
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("role_id", claims.RoleID)
		c.Locals("name", claims.RoleName)

		return c.Next()
	}
}

func Dosen() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Tidak ada authorization header")
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid authorization header format")
		}

		tokenString := parts[1]
		secretkey := os.Getenv("JWT_SECRET_KEY")

		claims := &utils.JwtCustomClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretkey), nil
		})

		if err != nil || !token.Valid {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, "token invalid")
		}

		if claims.RoleName != "Dosen Wali" {
			return helper.ErrorResponse(c, fiber.StatusForbidden, "bukan Dosen")
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("role_id", claims.RoleID)
		c.Locals("name", claims.RoleName)

		return c.Next()
	}
}

func Admin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Tidak ada authorization header")
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid authorization header format")
		}

		tokenString := parts[1]
		secretkey := os.Getenv("JWT_SECRET_KEY")

		claims := &utils.JwtCustomClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretkey), nil
		})

		if err != nil || !token.Valid {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, "token invalid")
		}

		if claims.RoleName != "Admin" {
			return helper.ErrorResponse(c, fiber.StatusForbidden, "bukan Admin")
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("role_id", claims.RoleID)
		c.Locals("name", claims.RoleName)

		return c.Next()
	}
}

func MultiRole(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Tidak ada authorization header")
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid authorization header format")
		}

		tokenString := parts[1]
		secretkey := os.Getenv("JWT_SECRET_KEY")

		claims := &utils.JwtCustomClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretkey), nil
		})

		if err != nil || !token.Valid {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, "token invalid")
		}

		// Cek apakah role user ada di dalam allowedRoles
		userRole := claims.RoleName
		isAllowed := false
		for _, role := range allowedRoles {
			if userRole == role {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			return helper.ErrorResponse(c, fiber.StatusForbidden, "Akses ditolak: role tidak diizinkan")
		}

		// Simpan user info ke context
		c.Locals("user_id", claims.UserID)
		c.Locals("role_id", claims.RoleID)
		c.Locals("name", claims.RoleName)

		return c.Next()
	}
}
