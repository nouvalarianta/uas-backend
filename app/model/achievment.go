package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AchievementReference struct {
	ID                 uuid.UUID    `db:"id" json:"id"`
	StudentID          uuid.UUID    `db:"student_id" json:"student_id"`
	MongoAchievementID string       `db:"mongo_achievement_id" json:"mongo_achievement_id"`
	Status             string       `db:"status" json:"status"` // draft, submitted, verified, rejected, deleted
	SubmittedAt        sql.NullTime `db:"submitted_at" json:"submitted_at,omitempty"`
	VerifiedAt         sql.NullTime `db:"verified_at" json:"verified_at,omitempty"`
	VerifiedBy         *uuid.UUID   `db:"verified_by" json:"verified_by,omitempty"`
	RejectionNote      *string      `db:"rejection_note" json:"rejection_note,omitempty"`
	CreatedAt          time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time    `db:"updated_at" json:"updated_at"`

	Student  *Student `db:"-" json:"student,omitempty"`
	Verifier *User    `db:"-" json:"verifier,omitempty"`
}

// Achievement - MongoDB Document
type Achievement struct {
	ID              primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	StudentID       string                 `bson:"studentId" json:"studentId"`
	AchievementType string                 `bson:"achievementType" json:"achievementType"`
	Title           string                 `bson:"title" json:"title"`
	Description     string                 `bson:"description,omitempty" json:"description,omitempty"`
	Details         map[string]interface{} `bson:"details,omitempty" json:"details,omitempty"`
	Attachments     []Attachment           `bson:"attachments,omitempty" json:"attachments,omitempty"`
	Points          float64                `bson:"points,omitempty" json:"points,omitempty"`
	Tags            []string               `bson:"tags,omitempty" json:"tags,omitempty"`
	CreatedAt       time.Time              `bson:"createdAt" json:"createdAt"`
	UpdatedAt       time.Time              `bson:"updatedAt" json:"updatedAt"`
}

type Attachment struct {
	FileName   string    `bson:"fileName" json:"fileName"`
	FileType   string    `bson:"fileType" json:"fileType"`
	FileSize   int64     `bson:"fileSize" json:"fileSize"`
	FileURL    string    `bson:"fileUrl,omitempty" json:"fileUrl,omitempty"`
	UploadedAt time.Time `bson:"uploadedAt" json:"uploadedAt"`
}

// Request DTOs
type CreateAchievementRequest struct {
	StudentID       string                 `json:"studentId" validate:"required,uuid"`
	AchievementType string                 `json:"achievementType" validate:"required,oneof=academic competition organization publication certification other"`
	Title           string                 `json:"title" validate:"required,min=3,max=500"`
	Description     string                 `json:"description"`
	Details         map[string]interface{} `json:"details"`
	Points          float64                `json:"points" validate:"min=0"`
	Tags            []string               `json:"tags"`
}

type UpdateAchievementRequest struct {
	Title       string                 `json:"title" validate:"omitempty,min=3,max=500"`
	Description string                 `json:"description"`
	Details     map[string]interface{} `json:"details"`
	Points      float64                `json:"points" validate:"min=0"`
	Tags        []string               `json:"tags"`
}

type SubmitAchievementRequest struct {
	// Empty for now, just trigger status change
}

type VerifyAchievementRequest struct {
	VerifiedBy       string `json:"verifiedBy" validate:"required,uuid"`
	VerificationNote string `json:"verificationNote"`
}

type RejectAchievementRequest struct {
	RejectionNote string `json:"rejectionNote" validate:"required"`
}

type UploadAttachmentRequest struct {
	FileName string `json:"fileName" validate:"required"`
	FileType string `json:"fileType" validate:"required"`
	FileSize int64  `json:"fileSize" validate:"required"`
	FileURL  string `json:"fileUrl"`
}
