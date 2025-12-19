package middleware

import (
	"net/http/httptest"
	"os"
	"testing"
	"uas-backend/app/model"
	"uas-backend/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func setupTestApp() *fiber.App {
	os.Setenv("JWT_SECRET_KEY", "test-secret-key")
	os.Setenv("JWT_EXPIRATION_MINUTE", "60")
	return fiber.New()
}

func teardownTest() {
	os.Unsetenv("JWT_SECRET_KEY")
	os.Unsetenv("JWT_EXPIRATION_MINUTE")
}

func generateTestToken(permissions []string) string {
	user := model.User{
		ID:       uuid.New(),
		Username: "testuser",
		RoleID:   uuid.New(),
		Role: &model.Role{
			ID:   uuid.New(),
			Name: "TestRole",
		},
	}
	token, _ := utils.GenerateToken(user, permissions)
	return token
}

func TestRequirePermission_NoAuthHeader(t *testing.T) {
	app := setupTestApp()
	defer teardownTest()

	app.Get("/test", RequirePermission("user:read"), func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", fiber.StatusUnauthorized, resp.StatusCode)
	}
}

func TestRequirePermission_InvalidHeaderFormat(t *testing.T) {
	app := setupTestApp()
	defer teardownTest()

	app.Get("/test", RequirePermission("user:read"), func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	tests := []struct {
		name   string
		header string
	}{
		{
			name:   "Missing Bearer prefix",
			header: "token-without-bearer",
		},
		{
			name:   "Wrong prefix",
			header: "Basic token123",
		},
		{
			name:   "Only Bearer",
			header: "Bearer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", tt.header)

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}

			if resp.StatusCode != fiber.StatusUnauthorized {
				t.Errorf("Expected status %d for %s, got %d", fiber.StatusUnauthorized, tt.name, resp.StatusCode)
			}
		})
	}
}

func TestRequirePermission_ValidToken(t *testing.T) {
	app := setupTestApp()
	defer teardownTest()

	token := generateTestToken([]string{"user:read", "user:write"})

	app.Get("/test", RequirePermission("user:read"), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}
}

func TestRequirePermission_MissingPermission(t *testing.T) {
	app := setupTestApp()
	defer teardownTest()

	token := generateTestToken([]string{"user:read"})

	app.Get("/test", RequirePermission("user:write"), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != fiber.StatusForbidden {
		t.Errorf("Expected status %d, got %d", fiber.StatusForbidden, resp.StatusCode)
	}
}

func TestRequirePermission_NoPermissionsRequired(t *testing.T) {
	app := setupTestApp()
	defer teardownTest()

	token := generateTestToken([]string{"user:read"})

	app.Get("/test", RequirePermission(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status %d for authenticated request, got %d", fiber.StatusOK, resp.StatusCode)
	}
}

func TestRequirePermission_BlacklistedToken(t *testing.T) {
	app := setupTestApp()
	defer teardownTest()

	token := generateTestToken([]string{"user:read"})

	claims, _ := utils.ParseToken(token)
	utils.AddToBlacklist(token, claims.ExpiresAt.Time)

	app.Get("/test", RequirePermission("user:read"), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status %d for blacklisted token, got %d", fiber.StatusUnauthorized, resp.StatusCode)
	}
}

func TestRequirePermission_MultiplePermissions(t *testing.T) {
	app := setupTestApp()
	defer teardownTest()

	token := generateTestToken([]string{"user:read", "user:write", "user:delete"})

	app.Get("/test", RequirePermission("user:read", "user:write"), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}
}
