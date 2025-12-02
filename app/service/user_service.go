package service

import (
	"fmt"
	"uas-backend/app/model"
	repository "uas-backend/app/repository"
	"uas-backend/helper"
	"uas-backend/utils"

	// "github.com/google/uuid"
	"github.com/gofiber/fiber/v2"
)

type UserService interface {
	Login(c *fiber.Ctx) error
	// GetAll(c *fiber.Ctx) error
	// GetByID(c *fiber.Ctx) error
	// GetByUsername(c *fiber.Ctx) error
	// Create(c *fiber.Ctx) error
	// Update(c *fiber.Ctx) error
	// Delete(c *fiber.Ctx) error
}

type userservice struct {
	userRepository repository.UserRepository
	// validate *validator.Validate
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

	fmt.Println("Retrieved user:", user)
	fmt.Println("Stored password hash:", passwordHash)
	fmt.Println("Provided password:", req.Password)

	if !utils.CheckPassword(req.Password, passwordHash) {
		return helper.ErrorResponse(c, fiber.StatusUnauthorized, "password salah")
	}

	token, err := utils.GenerateToken(*user)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal membuat token")
	}

	return c.JSON(fiber.Map{"token": token})
}
