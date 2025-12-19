package service

import (
	"strconv"
	repository "uas-backend/app/repository"
	helper "uas-backend/helper"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type LecturerService interface {
	GetAll(c *fiber.Ctx) error
	GetByID(c *fiber.Ctx) error
	GetAdvisees(c *fiber.Ctx) error
}

type lecturerService struct {
	lecturerRepo repository.LecturerRepository
}

func NewLecturerService(lecturerRepo repository.LecturerRepository) LecturerService {
	return &lecturerService{
		lecturerRepo: lecturerRepo,
	}
}

// GetAll godoc
// @Summary List All Lecturers
// @Description Get paginated list of all lecturers
// @Tags Lecturers
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Lecturers retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /lecturers [get]
func (s *lecturerService) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	skip := (page - 1) * limit

	lecturers, total, err := s.lecturerRepo.GetAll(limit, skip)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data: "+err.Error())
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Berhasil mengambil data lecturers", fiber.Map{
		"data": lecturers,
		"meta": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetByID godoc
// @Summary Get Lecturer by ID
// @Description Get lecturer details by ID
// @Tags Lecturers
// @Accept json
// @Produce json
// @Param id path string true "Lecturer ID (UUID)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Lecturer retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid lecturer ID"
// @Failure 404 {object} map[string]interface{} "Lecturer not found"
// @Router /lecturers/{id} [get]
func (s *lecturerService) GetByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	lecturerID, err := uuid.Parse(idParam)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "ID tidak valid")
	}

	lecturer, err := s.lecturerRepo.GetByID(lecturerID)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data: "+err.Error())
	}

	if lecturer == nil {
		return helper.ErrorResponse(c, fiber.StatusNotFound, "Lecturer tidak ditemukan")
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Berhasil mengambil data lecturer", lecturer)
}

// GetAdvisees godoc
// @Summary Get Lecturer's Advisees
// @Description Get all students advised by a specific lecturer
// @Tags Lecturers
// @Accept json
// @Produce json
// @Param id path string true "Lecturer ID (UUID)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Advisees retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid lecturer ID"
// @Failure 404 {object} map[string]interface{} "Lecturer not found"
// @Router /lecturers/{id}/advisees [get]
func (s *lecturerService) GetAdvisees(c *fiber.Ctx) error {
	idParam := c.Params("id")
	lecturerID, err := uuid.Parse(idParam)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "ID tidak valid")
	}

	// Check if lecturer exists
	lecturer, err := s.lecturerRepo.GetByID(lecturerID)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data lecturer: "+err.Error())
	}
	if lecturer == nil {
		return helper.ErrorResponse(c, fiber.StatusNotFound, "Lecturer tidak ditemukan")
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	skip := (page - 1) * limit

	// Get advisees
	students, total, err := s.lecturerRepo.GetAdvisees(lecturerID, limit, skip)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data advisees: "+err.Error())
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Berhasil mengambil data advisees", fiber.Map{
		"data": students,
		"meta": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}
