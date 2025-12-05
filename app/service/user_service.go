package service

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"uas-backend/app/model"
	repository "uas-backend/app/repository"
	"uas-backend/helper"
	"uas-backend/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UserService interface {
	Login(c *fiber.Ctx) error
	GetAll(c *fiber.Ctx) error
	GetByID(c *fiber.Ctx) error
	Create(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
}

type userservice struct {
	userRepository repository.UserRepository
}

func NewUserService(userRepository repository.UserRepository) UserService {
	return &userservice{
		userRepository: userRepository,
	}
}

func (s *userservice) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Request body tidak valid")
	}

	user, passwordHash, err := s.userRepository.Login(req.Username)

	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Username salah")
	}

	fmt.Println("user:", user)

	if !utils.CheckPassword(req.Password, passwordHash) {
		return helper.ErrorResponse(c, fiber.StatusUnauthorized, "password salah")
	}

	token, err := utils.GenerateToken(*user)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal membuat token")
	}

	return c.JSON(fiber.Map{"token": token})
}

func (s *userservice) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	sortBy := c.Query("sortBy", "created_at")
	order := c.Query("order", "desc")
	search := c.Query("search", "")

	// Validasi page dan limit
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
