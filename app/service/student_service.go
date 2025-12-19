package service

import (
	"context"
	"strconv"
	model "uas-backend/app/model"
	repository "uas-backend/app/repository"
	helper "uas-backend/helper"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type StudentService interface {
	GetAll(c *fiber.Ctx) error
	GetByID(c *fiber.Ctx) error
	GetAchievements(c *fiber.Ctx) error
	UpdateAdvisor(c *fiber.Ctx) error
}

type studentService struct {
	studentRepo              repository.StudentRepository
	achievementRepo          repository.AchievementRepository
	achievementReferenceRepo repository.AchievementReferenceRepository
}

func NewStudentService(
	studentRepo repository.StudentRepository,
	achievementRepo repository.AchievementRepository,
	achievementReferenceRepo repository.AchievementReferenceRepository,
) StudentService {
	return &studentService{
		studentRepo:              studentRepo,
		achievementRepo:          achievementRepo,
		achievementReferenceRepo: achievementReferenceRepo,
	}
}

// GetAll godoc
// @Summary List All Students
// @Description Get paginated list of all students
// @Tags Students
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Students retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /students [get]
func (s *studentService) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	skip := (page - 1) * limit

	students, total, err := s.studentRepo.GetAll(limit, skip)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data: "+err.Error())
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Berhasil mengambil data students", fiber.Map{
		"data": students,
		"meta": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetByID godoc
// @Summary Get Student by ID
// @Description Get student details by ID
// @Tags Students
// @Accept json
// @Produce json
// @Param id path string true "Student ID (UUID)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Student retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid student ID"
// @Failure 404 {object} map[string]interface{} "Student not found"
// @Router /students/{id} [get]
func (s *studentService) GetByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	studentID, err := uuid.Parse(idParam)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "ID tidak valid")
	}

	student, err := s.studentRepo.GetByID(studentID)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data: "+err.Error())
	}

	if student == nil {
		return helper.ErrorResponse(c, fiber.StatusNotFound, "Student tidak ditemukan")
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Berhasil mengambil data student", student)
}

// GetAchievements godoc
// @Summary Get Student Achievements
// @Description Get all achievements for a specific student with optional status filter
// @Tags Students
// @Accept json
// @Produce json
// @Param id path string true "Student ID (UUID)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param status query string false "Filter by status (draft, submitted, verified, rejected)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Achievements retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid student ID"
// @Failure 404 {object} map[string]interface{} "Student not found"
// @Router /students/{id}/achievements [get]
func (s *studentService) GetAchievements(c *fiber.Ctx) error {
	ctx := context.Background()
	idParam := c.Params("id")
	studentID, err := uuid.Parse(idParam)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "ID tidak valid")
	}

	// Check if student exists
	student, err := s.studentRepo.GetByID(studentID)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data student: "+err.Error())
	}
	if student == nil {
		return helper.ErrorResponse(c, fiber.StatusNotFound, "Student tidak ditemukan")
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Get status filter
	status := c.Query("status")

	// Get achievement references from PostgreSQL (exclude deleted)
	refs, _, err := s.achievementReferenceRepo.GetByStudentID(studentID, status, 1000, 0)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data references: "+err.Error())
	}

	// Build valid mongo IDs
	validMongoIDs := make(map[string]bool)
	for _, ref := range refs {
		validMongoIDs[ref.MongoAchievementID] = true
	}

	// Get achievements from MongoDB
	filter := map[string]interface{}{
		"studentId": studentID.String(),
	}

	if achievementType := c.Query("type"); achievementType != "" {
		filter["achievementType"] = achievementType
	}

	achievements, _, err := s.achievementRepo.GetAll(ctx, filter, 0, 1000)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data achievements: "+err.Error())
	}

	// Filter by valid IDs (exclude deleted)
	var filteredAchievements []*model.Achievement
	for _, achievement := range achievements {
		if validMongoIDs[achievement.ID.Hex()] {
			filteredAchievements = append(filteredAchievements, achievement)
		}
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Berhasil mengambil data achievements student", fiber.Map{
		"data": filteredAchievements,
		"meta": fiber.Map{
			"total": len(filteredAchievements),
		},
	})
}

// UpdateAdvisor godoc
// @Summary Update Student Advisor
// @Description Update student's advisor lecturer (Admin only)
// @Tags Students
// @Accept json
// @Produce json
// @Param id path string true "Student ID (UUID)"
// @Param body body model.UpdateAdvisorRequest true "Advisor lecturer ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Advisor updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "Student or lecturer not found"
// @Router /students/{id}/advisor [put]
func (s *studentService) UpdateAdvisor(c *fiber.Ctx) error {
	idParam := c.Params("id")
	studentID, err := uuid.Parse(idParam)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Student ID tidak valid")
	}

	var req struct {
		AdvisorID string `json:"advisor_id" validate:"required,uuid"`
	}

	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Request body tidak valid: "+err.Error())
	}

	advisorID, err := uuid.Parse(req.AdvisorID)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Advisor ID tidak valid")
	}

	// Check if student exists
	student, err := s.studentRepo.GetByID(studentID)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data student: "+err.Error())
	}
	if student == nil {
		return helper.ErrorResponse(c, fiber.StatusNotFound, "Student tidak ditemukan")
	}

	// Update advisor
	if err := s.studentRepo.UpdateAdvisor(studentID, advisorID); err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengupdate advisor: "+err.Error())
	}

	// Get updated student data
	updatedStudent, _ := s.studentRepo.GetByID(studentID)

	return helper.SuccessResponse(c, fiber.StatusOK, "Advisor berhasil diupdate", updatedStudent)
}
