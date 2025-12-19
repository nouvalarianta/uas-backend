package service

import (
	"context"
	"strconv"
	"uas-backend/app/repository"
	helper "uas-backend/helper"

	"github.com/gofiber/fiber/v2"
)

type ReportService interface {
	GetStatistics(c *fiber.Ctx) error
	GetStudentReport(c *fiber.Ctx) error
}

type reportService struct {
	reportRepo repository.ReportRepository
}

func NewReportService(reportRepo repository.ReportRepository) ReportService {
	return &reportService{
		reportRepo: reportRepo,
	}
}

// GetStatistics godoc
// @Summary Get System Statistics
// @Description Get overall achievement statistics with optional year filter
// @Tags Reports
// @Accept json
// @Produce json
// @Param year query int false "Filter by year (e.g., 2024)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Statistics retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /reports/statistics [get]
func (s *reportService) GetStatistics(c *fiber.Ctx) error {
	ctx := context.Background()

	// Optional year filter
	var year *int
	if yearStr := c.Query("year"); yearStr != "" {
		if yearInt, err := strconv.Atoi(yearStr); err == nil {
			year = &yearInt
		}
	}

	stats, err := s.reportRepo.GetStatistics(ctx, year)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil statistik: "+err.Error())
	}

	responseData := fiber.Map{
		"data": stats,
	}

	if year != nil {
		responseData["year"] = *year
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Statistik berhasil diambil", responseData)
}

// GetStudentReport godoc
// @Summary Get Student Achievement Report
// @Description Get achievement statistics for a specific student
// @Tags Reports
// @Accept json
// @Produce json
// @Param id path string true "Student ID (UUID)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Student report retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid student ID"
// @Failure 404 {object} map[string]interface{} "Student not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /reports/students/{id} [get]
func (s *reportService) GetStudentReport(c *fiber.Ctx) error {
	studentID := c.Params("id")
	if studentID == "" {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Student ID diperlukan")
	}

	stats, err := s.reportRepo.GetStudentAchievementStats(studentID)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil report student: "+err.Error())
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Report student berhasil diambil", fiber.Map{
		"student_id": studentID,
		"data":       stats,
	})
}
