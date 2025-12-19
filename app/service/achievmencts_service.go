package service

import (
	"context"
	"strconv"
	model "uas-backend/app/model"
	repository "uas-backend/app/repository"
	helper "uas-backend/helper"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AchievementService interface {
	GetAll(c *fiber.Ctx) error
	GetByID(c *fiber.Ctx) error
	Create(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
	Submit(c *fiber.Ctx) error
	Verify(c *fiber.Ctx) error
	Reject(c *fiber.Ctx) error
	UploadAttachment(c *fiber.Ctx) error
}

type achievementService struct {
	achievementRepo          repository.AchievementRepository
	achievementReferenceRepo repository.AchievementReferenceRepository
}

func NewAchievementService(
	achievementRepo repository.AchievementRepository,
	achievementReferenceRepo repository.AchievementReferenceRepository,
) AchievementService {
	return &achievementService{
		achievementRepo:          achievementRepo,
		achievementReferenceRepo: achievementReferenceRepo,
	}
}

// GetAll godoc
// @Summary List All Achievements
// @Description Get paginated list of all achievements with optional filters
// @Tags Achievements
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param status query string false "Filter by status (draft, submitted, verified, rejected)"
// @Param studentId query string false "Filter by student ID (UUID)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Achievements retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /achievements [get]
func (s *achievementService) GetAll(c *fiber.Ctx) error {
	ctx := context.Background()

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	skip := int64((page - 1) * limit)

	status := c.Query("status")
	studentID := c.Query("studentId")

	var refs []*model.AchievementReference
	var err error

	if studentID != "" {
		studentUUID, parseErr := uuid.Parse(studentID)
		if parseErr != nil {
			return helper.ErrorResponse(c, fiber.StatusBadRequest, "Student ID tidak valid")
		}
		refs, _, err = s.achievementReferenceRepo.GetByStudentID(studentUUID, status, 1000, 0)
	} else {
		refs, _, err = s.achievementReferenceRepo.GetAll(status, 1000, 0)
	}

	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data references: "+err.Error())
	}

	validMongoIDs := make(map[string]bool)
	for _, ref := range refs {
		validMongoIDs[ref.MongoAchievementID] = true
	}

	filter := make(map[string]interface{})

	if studentID != "" {
		filter["studentId"] = studentID
	}

	if achievementType := c.Query("type"); achievementType != "" {
		filter["achievementType"] = achievementType
	}

	if search := c.Query("search"); search != "" {
		filter["$or"] = []map[string]interface{}{
			{"title": map[string]interface{}{"$regex": search, "$options": "i"}},
			{"description": map[string]interface{}{"$regex": search, "$options": "i"}},
		}
	}

	achievements, _, err := s.achievementRepo.GetAll(ctx, filter, skip, int64(limit))
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data: "+err.Error())
	}

	var filteredAchievements []*model.Achievement
	for _, achievement := range achievements {
		if validMongoIDs[achievement.ID.Hex()] {
			filteredAchievements = append(filteredAchievements, achievement)
		}
	}

	total := int64(len(filteredAchievements))

	return helper.SuccessResponse(c, fiber.StatusOK, "Berhasil mengambil data achievements", fiber.Map{
		"achievements": filteredAchievements,
		"meta": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		}})

}

// GetByID godoc
// @Summary Get Achievement by ID
// @Description Get achievement details by MongoDB ObjectID
// @Tags Achievements
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID (MongoDB ObjectID)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Achievement retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid achievement ID"
// @Failure 404 {object} map[string]interface{} "Achievement not found"
// @Router /achievements/{id} [get]
func (s *achievementService) GetByID(c *fiber.Ctx) error {
	ctx := context.Background()
	idParam := c.Params("id")

	objectID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "ID tidak valid")
	}

	ref, err := s.achievementReferenceRepo.GetByMongoID(objectID.Hex())
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil reference: "+err.Error())
	}
	if ref == nil {
		return helper.ErrorResponse(c, fiber.StatusNotFound, "Achievement tidak ditemukan atau sudah dihapus")
	}

	achievement, err := s.achievementRepo.GetByID(ctx, objectID)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data: "+err.Error())
	}

	if achievement == nil {
		return helper.ErrorResponse(c, fiber.StatusNotFound, "Achievement tidak ditemukan")
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Berhasil mengambil data achievement", fiber.Map{
		"achievement": achievement,
		"reference":   ref,
	})
}

// Create godoc
// @Summary Create New Achievement
// @Description Create a new achievement (draft status)
// @Tags Achievements
// @Accept json
// @Produce json
// @Param body body model.CreateAchievementRequest true "Achievement data"
// @Security BearerAuth
// @Success 201 {object} map[string]interface{} "Achievement created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /achievements [post]
func (s *achievementService) Create(c *fiber.Ctx) error {
	ctx := context.Background()

	var req model.CreateAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Request body tidak valid: "+err.Error())
	}

	studentUUID, err := uuid.Parse(req.StudentID)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Student ID tidak valid")
	}

	achievement := &model.Achievement{
		StudentID:       req.StudentID,
		AchievementType: req.AchievementType,
		Title:           req.Title,
		Description:     req.Description,
		Details:         req.Details,
		Points:          req.Points,
		Tags:            req.Tags,
		Attachments:     []model.Attachment{},
	}

	createdAchievement, err := s.achievementRepo.Create(ctx, achievement)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal membuat achievement: "+err.Error())
	}

	ref, err := s.achievementReferenceRepo.Create(studentUUID, createdAchievement.ID.Hex())
	if err != nil {
		s.achievementRepo.Delete(ctx, createdAchievement.ID)
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal membuat reference: "+err.Error())
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Achievement berhasil dibuat", fiber.Map{
		"achievement": createdAchievement,
		"reference":   ref,
	})
}

// Update godoc
// @Summary Update Achievement
// @Description Update achievement details (only draft status can be updated)
// @Tags Achievements
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID (MongoDB ObjectID)"
// @Param body body model.UpdateAchievementRequest true "Achievement data to update"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Achievement updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 403 {object} map[string]interface{} "Only draft achievements can be updated"
// @Failure 404 {object} map[string]interface{} "Achievement not found"
// @Router /achievements/{id} [put]
func (s *achievementService) Update(c *fiber.Ctx) error {
	ctx := context.Background()
	idParam := c.Params("id")

	objectID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "ID tidak valid")
	}

	var req model.UpdateAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Request body tidak valid: "+err.Error())
	}

	existing, err := s.achievementRepo.GetByID(ctx, objectID)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data: "+err.Error())
	}
	if existing == nil {
		return helper.ErrorResponse(c, fiber.StatusNotFound, "Achievement tidak ditemukan")
	}

	ref, _ := s.achievementReferenceRepo.GetByMongoID(objectID.Hex())
	if ref != nil && ref.Status != "draft" {
		return helper.ErrorResponse(c, fiber.StatusForbidden, "Hanya achievement dengan status draft yang dapat diubah")
	}

	updatedAchievement, err := s.achievementRepo.Update(ctx, objectID, &req)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengupdate achievement: "+err.Error())
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Achievement berhasil diupdate", updatedAchievement)
}

// Delete godoc
// @Summary Delete Achievement
// @Description Soft delete achievement (only draft/submitted status can be deleted)
// @Tags Achievements
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID (MongoDB ObjectID)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Achievement deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid achievement ID"
// @Failure 403 {object} map[string]interface{} "Verified/rejected achievements cannot be deleted"
// @Failure 404 {object} map[string]interface{} "Achievement not found"
// @Router /achievements/{id} [delete]
func (s *achievementService) Delete(c *fiber.Ctx) error {
	ctx := context.Background()
	idParam := c.Params("id")

	objectID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "ID tidak valid")
	}

	ref, err := s.achievementReferenceRepo.GetByMongoID(objectID.Hex())
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil reference: "+err.Error())
	}
	if ref == nil {
		return helper.ErrorResponse(c, fiber.StatusNotFound, "Achievement reference tidak ditemukan")
	}

	if ref.Status == "deleted" {
		return helper.ErrorResponse(c, fiber.StatusNotFound, "Achievement sudah dihapus")
	}

	if ref.Status == "verified" {
		return helper.ErrorResponse(c, fiber.StatusForbidden, "Achievement yang sudah diverifikasi tidak dapat dihapus")
	}
	if ref.Status == "rejected" {
		return helper.ErrorResponse(c, fiber.StatusForbidden, "Achievement yang sudah direject tidak dapat dihapus")
	}

	if ref.Status != "draft" && ref.Status != "submitted" {
		return helper.ErrorResponse(c, fiber.StatusForbidden, "Hanya achievement dengan status draft atau submitted yang dapat dihapus")
	}

	existing, err := s.achievementRepo.GetByID(ctx, objectID)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data: "+err.Error())
	}
	if existing == nil {
		return helper.ErrorResponse(c, fiber.StatusNotFound, "Achievement tidak ditemukan")
	}

	if err := s.achievementReferenceRepo.Delete(objectID.Hex()); err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal menghapus achievement: "+err.Error())
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Achievement berhasil dihapus", nil)
}

// Submit godoc
// @Summary Submit Achievement
// @Description Submit achievement for verification (changes status from draft to submitted)
// @Tags Achievements
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID (MongoDB ObjectID)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Achievement submitted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid achievement ID"
// @Failure 403 {object} map[string]interface{} "Only draft achievements can be submitted"
// @Failure 404 {object} map[string]interface{} "Achievement not found"
// @Router /achievements/{id}/submit [post]
func (s *achievementService) Submit(c *fiber.Ctx) error {
	idParam := c.Params("id")

	objectID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "ID tidak valid")
	}

	ref, err := s.achievementReferenceRepo.GetByMongoID(objectID.Hex())
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data: "+err.Error())
	}
	if ref == nil {
		return helper.ErrorResponse(c, fiber.StatusNotFound, "Achievement reference tidak ditemukan")
	}

	if ref.Status != "draft" {
		return helper.ErrorResponse(c, fiber.StatusForbidden, "Hanya achievement dengan status draft yang dapat disubmit")
	}

	if err := s.achievementReferenceRepo.Submit(objectID.Hex()); err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal submit achievement: "+err.Error())
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Achievement berhasil disubmit untuk verifikasi", nil)
}

// Verify godoc
// @Summary Verify Achievement
// @Description Verify achievement and award points (Lecturer only)
// @Tags Achievements
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID (MongoDB ObjectID)"
// @Param body body model.VerifyAchievementRequest true "Verification data with points"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Achievement verified successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 403 {object} map[string]interface{} "Forbidden - Requires achievement:verify permission"
// @Failure 404 {object} map[string]interface{} "Achievement not found"
// @Router /achievements/{id}/verify [post]
func (s *achievementService) Verify(c *fiber.Ctx) error {
	idParam := c.Params("id")

	objectID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "ID tidak valid")
	}

	var req model.VerifyAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Request body tidak valid: "+err.Error())
	}

	verifiedBy, err := uuid.Parse(req.VerifiedBy)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Verified by ID tidak valid")
	}

	ref, err := s.achievementReferenceRepo.GetByMongoID(objectID.Hex())
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data: "+err.Error())
	}
	if ref == nil {
		return helper.ErrorResponse(c, fiber.StatusNotFound, "Achievement reference tidak ditemukan")
	}

	if ref.Status == "draft" {
		return helper.ErrorResponse(c, fiber.StatusForbidden, "Achievement harus disubmit terlebih dahulu sebelum diverifikasi")
	}

	if ref.Status == "rejected" {
		return helper.ErrorResponse(c, fiber.StatusForbidden, "Achievement yang sudah direject tidak dapat diverifikasi")
	}

	if ref.Status == "deleted" {
		return helper.ErrorResponse(c, fiber.StatusForbidden, "Achievement yang sudah dihapus tidak dapat diverifikasi")
	}

	if ref.Status != "submitted" {
		return helper.ErrorResponse(c, fiber.StatusForbidden, "Hanya achievement dengan status submitted yang dapat diverifikasi")
	}

	if err := s.achievementReferenceRepo.Verify(objectID.Hex(), verifiedBy, req.VerificationNote); err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal verify achievement: "+err.Error())
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Achievement berhasil diverifikasi", nil)
}

// Reject godoc
// @Summary Reject Achievement
// @Description Reject achievement with note (Lecturer only)
// @Tags Achievements
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID (MongoDB ObjectID)"
// @Param body body model.RejectAchievementRequest true "Rejection note"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Achievement rejected successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 403 {object} map[string]interface{} "Forbidden - Only submitted achievements can be rejected"
// @Failure 404 {object} map[string]interface{} "Achievement not found"
// @Router /achievements/{id}/reject [post]
func (s *achievementService) Reject(c *fiber.Ctx) error {
	idParam := c.Params("id")

	objectID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "ID tidak valid")
	}

	var req model.RejectAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Request body tidak valid: "+err.Error())
	}

	if req.RejectionNote == "" {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Rejection note harus diisi")
	}

	ref, err := s.achievementReferenceRepo.GetByMongoID(objectID.Hex())
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data: "+err.Error())
	}
	if ref == nil {
		return helper.ErrorResponse(c, fiber.StatusNotFound, "Achievement reference tidak ditemukan")
	}

	if ref.Status != "submitted" {
		return helper.ErrorResponse(c, fiber.StatusForbidden, "Hanya achievement dengan status submitted yang dapat direject")
	}

	if err := s.achievementReferenceRepo.Reject(objectID.Hex(), req.RejectionNote); err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal reject achievement: "+err.Error())
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Achievement berhasil direject", nil)
}

// UploadAttachment godoc
// @Summary Upload Attachment
// @Description Upload attachment URLs for achievement
// @Tags Achievements
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID (MongoDB ObjectID)"
// @Param body body model.UploadAttachmentRequest true "Attachment URLs"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Attachment uploaded successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "Achievement not found"
// @Router /achievements/{id}/upload [post]
func (s *achievementService) UploadAttachment(c *fiber.Ctx) error {
	ctx := context.Background()
	idParam := c.Params("id")

	objectID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "ID tidak valid")
	}

	var req model.UploadAttachmentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Request body tidak valid: "+err.Error())
	}

	existing, err := s.achievementRepo.GetByID(ctx, objectID)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data: "+err.Error())
	}
	if existing == nil {
		return helper.ErrorResponse(c, fiber.StatusNotFound, "Achievement tidak ditemukan")
	}

	attachment := model.Attachment{
		FileName: req.FileName,
		FileType: req.FileType,
		FileSize: req.FileSize,
		FileURL:  req.FileURL,
	}

	if err := s.achievementRepo.AddAttachment(ctx, objectID, attachment); err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal upload attachment: "+err.Error())
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Attachment berhasil diupload", attachment)
}
