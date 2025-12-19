package service

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"os"
	"testing"
	"uas-backend/app/mocks"
	"uas-backend/app/model"
	"uas-backend/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupTestEnv() {
	os.Setenv("JWT_SECRET_KEY", "test-secret-key")
	os.Setenv("JWT_EXPIRATION_MINUTE", "60")
}

func teardownTestEnv() {
	os.Unsetenv("JWT_SECRET_KEY")
	os.Unsetenv("JWT_EXPIRATION_MINUTE")
}

func TestLogin_Success(t *testing.T) {
	setupTestEnv()
	defer teardownTestEnv()

	mockRepo := new(mocks.MockUserRepository)
	service := NewUserService(mockRepo)

	app := fiber.New()
	app.Post("/login", service.Login)

	userID := uuid.New()
	roleID := uuid.New()
	user := &model.User{
		ID:       userID,
		Username: "testuser",
		Email:    "test@example.com",
		FullName: "Test User",
		RoleID:   roleID,
		IsActive: nil,
		Role: &model.Role{
			ID:          roleID,
			Name:        "Admin",
			Description: "Administrator",
		},
	}
	passwordHash := "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
	permissions := []string{"user:read", "user:write"}

	mockRepo.On("Login", "testuser").Return(user, passwordHash, nil)
	mockRepo.On("GetPermissionsByRoleID", roleID).Return(permissions, nil)

	reqBody := model.LoginRequest{
		Username: "testuser",
		Password: "password",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &response)

	assert.Equal(t, "success", response["status"])
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	assert.NotEmpty(t, data["token"])
	assert.NotEmpty(t, data["refreshToken"])

	mockRepo.AssertExpectations(t)
}

func TestLogin_InvalidRequestBody(t *testing.T) {
	setupTestEnv()
	defer teardownTestEnv()

	mockRepo := new(mocks.MockUserRepository)
	service := NewUserService(mockRepo)

	app := fiber.New()
	app.Post("/login", service.Login)

	req := httptest.NewRequest("POST", "/login", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestLogin_WrongUsername(t *testing.T) {
	setupTestEnv()
	defer teardownTestEnv()

	mockRepo := new(mocks.MockUserRepository)
	service := NewUserService(mockRepo)

	app := fiber.New()
	app.Post("/login", service.Login)

	mockRepo.On("Login", "wronguser").Return(nil, "", sql.ErrNoRows)

	reqBody := model.LoginRequest{
		Username: "wronguser",
		Password: "password",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	mockRepo.AssertExpectations(t)
}

func TestLogin_WrongPassword(t *testing.T) {
	setupTestEnv()
	defer teardownTestEnv()

	mockRepo := new(mocks.MockUserRepository)
	service := NewUserService(mockRepo)

	app := fiber.New()
	app.Post("/login", service.Login)

	userID := uuid.New()
	roleID := uuid.New()
	user := &model.User{
		ID:       userID,
		Username: "testuser",
		RoleID:   roleID,
		Role: &model.Role{
			ID:   roleID,
			Name: "Admin",
		},
	}
	passwordHash := "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

	mockRepo.On("Login", "testuser").Return(user, passwordHash, nil)

	reqBody := model.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	mockRepo.AssertExpectations(t)
}

func TestGetByID_Success(t *testing.T) {
	setupTestEnv()
	defer teardownTestEnv()

	mockRepo := new(mocks.MockUserRepository)
	service := NewUserService(mockRepo)

	app := fiber.New()
	app.Get("/users/:id", service.GetByID)

	userID := uuid.New()
	roleID := uuid.New()
	user := &model.User{
		ID:       userID,
		Username: "testuser",
		Email:    "test@example.com",
		FullName: "Test User",
		RoleID:   roleID,
		IsActive: nil,
		Role: &model.Role{
			ID:   roleID,
			Name: "Admin",
		},
	}

	mockRepo.On("GetByID", userID).Return(user, nil)

	req := httptest.NewRequest("GET", "/users/"+userID.String(), nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &response)

	assert.Equal(t, "success", response["status"])
	mockRepo.AssertExpectations(t)
}

func TestGetByID_InvalidUUID(t *testing.T) {
	setupTestEnv()
	defer teardownTestEnv()

	mockRepo := new(mocks.MockUserRepository)
	service := NewUserService(mockRepo)

	app := fiber.New()
	app.Get("/users/:id", service.GetByID)

	req := httptest.NewRequest("GET", "/users/invalid-uuid", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestGetByID_NotFound(t *testing.T) {
	setupTestEnv()
	defer teardownTestEnv()

	mockRepo := new(mocks.MockUserRepository)
	service := NewUserService(mockRepo)

	app := fiber.New()
	app.Get("/users/:id", service.GetByID)

	userID := uuid.New()
	mockRepo.On("GetByID", userID).Return(nil, sql.ErrNoRows)

	req := httptest.NewRequest("GET", "/users/"+userID.String(), nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	mockRepo.AssertExpectations(t)
}

func TestLogout_Success(t *testing.T) {
	setupTestEnv()
	defer teardownTestEnv()

	mockRepo := new(mocks.MockUserRepository)
	service := NewUserService(mockRepo)

	app := fiber.New()
	app.Post("/logout", func(c *fiber.Ctx) error {
		userID := uuid.New()
		roleID := uuid.New()
		permissions := []string{"user:read"}

		user := model.User{
			ID:     userID,
			RoleID: roleID,
			Role: &model.Role{
				ID:   roleID,
				Name: "Admin",
			},
		}

		token, _ := utils.GenerateToken(user, permissions)
		claims, _ := utils.ParseToken(token)

		c.Locals("claims", claims)
		c.Locals("token", token)
		return service.Logout(c)
	})

	req := httptest.NewRequest("POST", "/logout", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &response)

	assert.Equal(t, "success", response["status"])
}

func TestRefresh_Success(t *testing.T) {
	setupTestEnv()
	defer teardownTestEnv()

	mockRepo := new(mocks.MockUserRepository)
	service := NewUserService(mockRepo)

	app := fiber.New()
	app.Post("/refresh", func(c *fiber.Ctx) error {
		userID := uuid.New()
		roleID := uuid.New()
		permissions := []string{"user:read", "user:write"}

		user := model.User{
			ID:       userID,
			Username: "testuser",
			RoleID:   roleID,
			Role: &model.Role{
				ID:   roleID,
				Name: "Admin",
			},
		}

		refreshToken, _ := utils.GenerateRefreshToken(user, permissions)
		claims, _ := utils.ParseToken(refreshToken)

		c.Locals("claims", claims)
		return service.Refresh(c)
	})

	mockRepo.On("GetPermissionsByRoleID", mock.AnythingOfType("uuid.UUID")).
		Return([]string{"user:read", "user:write"}, nil)

	req := httptest.NewRequest("POST", "/refresh", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &response)

	assert.Equal(t, "success", response["status"])
	assert.NotNil(t, response["data"])

	mockRepo.AssertExpectations(t)
}

func TestDelete_Success(t *testing.T) {
	setupTestEnv()
	defer teardownTestEnv()

	mockRepo := new(mocks.MockUserRepository)
	service := NewUserService(mockRepo)

	app := fiber.New()
	app.Delete("/users/:id", service.Delete)

	userID := uuid.New()
	mockRepo.On("Delete", userID).Return(nil)

	req := httptest.NewRequest("DELETE", "/users/"+userID.String(), nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	mockRepo.AssertExpectations(t)
}

func TestDelete_Error(t *testing.T) {
	setupTestEnv()
	defer teardownTestEnv()

	mockRepo := new(mocks.MockUserRepository)
	service := NewUserService(mockRepo)

	app := fiber.New()
	app.Delete("/users/:id", service.Delete)

	userID := uuid.New()
	mockRepo.On("Delete", userID).Return(errors.New("database error"))

	req := httptest.NewRequest("DELETE", "/users/"+userID.String(), nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	mockRepo.AssertExpectations(t)
}
