package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type AchievementReference struct {
	ID                 uuid.UUID    `db:"id" json:"id"`
	StudentID          uuid.UUID    `db:"student_id" json:"student_id"`
	MongoAchievementID string       `db:"mongo_achievement_id" json:"mongo_achievement_id"`
	Status             string       `db:"status" json:"status"`
	SubmittedAt        sql.NullTime `db:"submitted_at" json:"submitted_at,omitempty"`
	VerifiedAt         sql.NullTime `db:"verified_at" json:"verified_at,omitempty"`
	VerifiedBy         *uuid.UUID   `db:"verified_by" json:"verified_by,omitempty"`
	RejectionNote      string       `db:"rejection_note" json:"rejection_note,omitempty"`
	CreatedAt          time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time    `db:"updated_at" json:"updated_at"`

	Student  *Student `db:"-" json:"student,omitempty"`
	Verifier *User    `db:"-" json:"verifier,omitempty"`
}
