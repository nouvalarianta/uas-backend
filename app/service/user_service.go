package service

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	model "uas-backend/app/model"
	repository "uas-backend/app/repository"
	helper "uas-backend/helper"
	utils "uas-backend/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UserService interface {
	Login(c *fiber.Ctx) error
	Refresh(c *fiber.Ctx) error
	Logout(c *fiber.Ctx) error
	Profile(c *fiber.Ctx) error
	GetAll(c *fiber.Ctx) error
	GetByID(c *fiber.Ctx) error
	Create(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
	ReplaceRole(c *fiber.Ctx) error
}

type userservice struct {
	userRepository repository.UserRepository
}

func NewUserService(userRepository repository.UserRepository) UserService {
	return &userservice{
		userRepository: userRepository,
	}
}

// Login godoc
// @Summary User Login
// @Description Authenticate user and get JWT tokens (access token and refresh token)
// @Tags Authentication
// @Accept json
// @Produce json
// @Param body body model.LoginRequest true "Login credentials"
// @Success 200 {object} map[string]interface{} "Login successful with token and user data"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 401 {object} map[string]interface{} "Invalid credentials"
// @Router /login [post]
func (s *userservice) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Request body tidak valid")
	}

	user, passwordHash, err := s.userRepository.Login(req.Username)

	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Username salah")
	}

	if !utils.CheckPassword(req.Password, passwordHash) {
		return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Password salah")
	}

	permissions, err := s.userRepository.GetPermissionsByRoleID(user.RoleID)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil permissions: "+err.Error())
	}

	fmt.Print(permissions)

	token, err := utils.GenerateToken(*user, permissions)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal membuat token")
	}

	refreshToken, err := utils.GenerateRefreshToken(*user, permissions)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal membuat refresh token")
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Login berhasil", fiber.Map{
		"token":        token,
		"refreshToken": refreshToken,
		"user": fiber.Map{
			"id":          user.ID,
			"username":    user.Username,
			"fullName":    user.FullName,
			"role":        user.Role.Name,
			"permissions": permissions,
		},
	})
}

// GetByID godoc
// @Summary Get User by ID
// @Description Get user details by ID (Admin only)
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "User retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid user ID"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Router /users/{id} [get]

// GetAll godoc
// @Summary List All Users
// @Description Get paginated list of all users (Admin only)
// @Tags Users
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Users retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Requires user:manage permission"
// @Router /users [get]
func (s *userservice) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	sortBy := c.Query("sortBy", "created_at")
	order := c.Query("order", "desc")
	search := c.Query("search", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	users, err := s.userRepository.GetAll(search, sortBy, order, limit, offset)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data users: "+err.Error())
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Berhasil mengambil data users", fiber.Map{
		"users": users,
		"page":  page,
		"limit": limit,
		"total": len(users),
	})
}

// GetByID godoc
// @Summary Get User by ID
// @Description Get user details by ID (Admin only)
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "User retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid user ID"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Router /users/{id} [get]
func (s *userservice) GetByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "id tidak valid")
	}

	user, err := s.userRepository.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return helper.ErrorResponse(c, fiber.StatusBadRequest, "User tidak ditemukan")
		}
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data user: "+err.Error())
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Berhasil mengambil data user", user)
}

// Create godoc
// @Summary Create New User
// @Description Create a new user (Admin only)
// @Tags Users
// @Accept json
// @Produce json
// @Param body body model.CreateUserRequest true "User data"
// @Security BearerAuth
// @Success 201 {object} map[string]interface{} "User created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body or validation error"
// @Failure 409 {object} map[string]interface{} "User or email already exists"
// @Router /users [post]
func (s *userservice) Create(c *fiber.Ctx) error {
	var req model.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Request body tidak valid: "+err.Error())
	}

	// Validasi username
	if len(req.Username) < 3 || len(req.Username) > 50 {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Username harus 3-50 karakter")
	}

	if len(req.FullName) < 3 || len(req.FullName) > 100 {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Full Name harus 3-100 karakter")
	}

	// Validasi email format (simple check)
	if !strings.Contains(req.Email, "@") {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Email tidak valid")
	}

	// Validasi password
	if len(req.Password) < 6 {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Password minimal 6 karakter")
	}

	// Parse roleID
	roleID, err := uuid.Parse(req.RoleID)

	exist, _ := s.userRepository.CheckRoleExists(roleID)
	fmt.Println("role :", exist)
	if !exist {
		return helper.ErrorResponse(c, fiber.StatusConflict, "role tidak ada")
	}

	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Role ID tidak valid")
	}

	// Check username exists
	usernameExists, err := s.userRepository.CheckUsernameExists(req.Username)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal validasi username: "+err.Error())
	}
	if usernameExists {
		return helper.ErrorResponse(c, fiber.StatusConflict, "Username sudah terdaftar")
	}

	// Check email exists
	emailExists, err := s.userRepository.CheckEmailExists(req.Email)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal validasi email: "+err.Error())
	}
	if emailExists {
		return helper.ErrorResponse(c, fiber.StatusConflict, "Email sudah terdaftar")
	}

	// Hash password
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal hash password: "+err.Error())
	}

	// Create user
	user, err := s.userRepository.Create(req.Username, req.Email, passwordHash, req.FullName, roleID)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal menambah user: "+err.Error())
	}

	return helper.SuccessResponse(c, fiber.StatusCreated, "Berhasil menambah user", user)
}

// Update godoc
// @Summary Update User
// @Description Update user details (Admin only)
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Param body body model.UpdateUserRequest true "User data to update"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "User updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Router /users/{id} [put]
func (s *userservice) Update(c *fiber.Ctx) error {
	var req model.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Request body tidak valid: "+err.Error())
	}

	// cek len username
	if len(req.Username) < 3 || len(req.Username) > 50 {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Username harus 3-50 karakter")
	}

	//cek len fullname
	if len(req.FullName) < 3 || len(req.FullName) > 100 {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Fullname harus 3-100 karakter")
	}

	//cek apakah email valid
	if !strings.Contains(req.Email, "@") {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Email tidak valid")
	}

	//cek role_id ada atau ngga pada database
	RoleId, err := uuid.Parse(req.RoleID)

	exist, _ := s.userRepository.CheckRoleExists(RoleId)
	if !exist {
		return helper.ErrorResponse(c, fiber.StatusConflict, "role_id tidak ada")
	}

	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Role ID tidak valid")
	}

	usernameExist, err := s.userRepository.CheckUsernameExists(req.Username)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal validasi username: "+err.Error())
	}

	if usernameExist {
		return helper.ErrorResponse(c, fiber.StatusConflict, "Username sudah terdaftar")
	}

	emailExist, err := s.userRepository.CheckEmailExists(req.Email)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal validasi email: "+err.Error())
	}

	if emailExist {
		return helper.ErrorResponse(c, fiber.StatusConflict, "email sdah terdaftar")
	}

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "id tidak valid")
	}

	update, err := s.userRepository.Update(id, &req)

	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "gagal update users"+err.Error())
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Berhasil mengupdate Users", update)
}

// Delete godoc
// @Summary Delete User
// @Description Soft delete user by ID (Admin only)
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "User deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid user ID"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Router /users/{id} [delete]
func (s *userservice) Delete(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "ID tidak valid")
	}

	user, err := s.userRepository.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return helper.ErrorResponse(c, fiber.StatusNotFound, "User tidak ditemukan atau sudah dihapus")
		}
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data user: "+err.Error())
	}

	if user == nil {
		return helper.ErrorResponse(c, fiber.StatusNotFound, "User tidak ditemukan atau sudah dihapus")
	}

	err = s.userRepository.Delete(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return helper.ErrorResponse(c, fiber.StatusNotFound, "User tidak ditemukan atau sudah dihapus")
		}
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal menghapus user: "+err.Error())
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Berhasil menghapus user", nil)
}

// ReplaceRole godoc
// @Summary Change User Role
// @Description Change user role (Admin only)
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Param role_id query string true "New role ID (UUID)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Role changed successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "User or role not found"
// @Router /users/{id}/role [put]
func (s *userservice) ReplaceRole(c *fiber.Ctx) error {
	idStr := c.Params("id")
	roleIDStr := c.Query("role_id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "ID user tidak valid")
	}
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Role ID tidak valid")
	}

	roleExists, err := s.userRepository.CheckRoleExists(roleID)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal cek role: "+err.Error())
	}
	if !roleExists {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Role tidak ditemukan")
	}

	user, err := s.userRepository.ReplaceRole(id, roleID)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengganti role: "+err.Error())
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Berhasil mengganti role user", user)
}

// Refresh godoc
// @Summary Refresh Access Token
// @Description Refresh access token using refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param body body model.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} map[string]interface{} "Token refreshed successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request or expired token"
// @Failure 401 {object} map[string]interface{} "Invalid refresh token"
// @Router /auth/refresh [post]
func (s *userservice) Refresh(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}

	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Request body tidak valid")
	}

	if req.RefreshToken == "" {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Refresh token diperlukan")
	}

	// Check if refresh token is blacklisted
	if utils.IsBlacklisted(req.RefreshToken) {
		return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Refresh token telah di-logout")
	}

	// Parse refresh token
	claims, err := utils.ParseToken(req.RefreshToken)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Refresh token invalid atau expired")
	}

	// Get user from database
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "User ID tidak valid")
	}

	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusNotFound, "User tidak ditemukan")
	}

	// Get permissions
	permissions, err := s.userRepository.GetPermissionsByRoleID(user.RoleID)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil permissions: "+err.Error())
	}

	// Generate new access token
	newToken, err := utils.GenerateToken(*user, permissions)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal membuat token baru")
	}

	// Generate new refresh token
	newRefreshToken, err := utils.GenerateRefreshToken(*user, permissions)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal membuat refresh token baru")
	}

	// Blacklist old refresh token
	utils.AddToBlacklist(req.RefreshToken, claims.ExpiresAt.Time)

	return helper.SuccessResponse(c, fiber.StatusOK, "Token berhasil di-refresh", fiber.Map{
		"token":        newToken,
		"refreshToken": newRefreshToken,
	})
}

// Logout godoc
// @Summary User Logout
// @Description Logout user and blacklist current token
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Logout successful"
// @Failure 401 {object} map[string]interface{} "Unauthorized or invalid token"
// @Router /auth/logout [post]
func (s *userservice) Logout(c *fiber.Ctx) error {
	// Get token from header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Token tidak ditemukan")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Format token tidak valid")
	}

	tokenString := parts[1]

	// Parse token to get expiration time
	claims, err := utils.ParseToken(tokenString)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Token invalid")
	}

	// Add token to blacklist
	utils.AddToBlacklist(tokenString, claims.ExpiresAt.Time)

	return helper.SuccessResponse(c, fiber.StatusOK, "Logout berhasil", nil)
}

func (s *userservice) Profile(c *fiber.Ctx) error {
	// Get user ID from token (set by middleware)
	userIDStr, ok := c.Locals("user_id").(string)
	if !ok {
		return helper.ErrorResponse(c, fiber.StatusUnauthorized, "User ID tidak ditemukan dalam token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "User ID tidak valid")
	}

	// Get user from database
	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusNotFound, "User tidak ditemukan")
	}

	// Get permissions
	permissions, err := s.userRepository.GetPermissionsByRoleID(user.RoleID)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil permissions: "+err.Error())
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Berhasil mengambil profile", fiber.Map{
		"id":       user.ID,
		"username": user.Username,
		"fullName": user.FullName,
		"email":    user.Email,
		"role": fiber.Map{
			"id":   user.Role.ID,
			"name": user.Role.Name,
		},
		"permissions": permissions,
		"createdAt":   user.CreatedAt,
		"updatedAt":   user.UpdatedAt,
	})
}
