package model

import (
	"time"

	"github.com/google/uuid"
)

type Student struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	UserID       uuid.UUID  `db:"user_id" json:"user_id"`
	StudentID    string     `db:"student_id" json:"student_id"`
	ProgramStudy string     `db:"program_study" json:"program_study,omitempty"`
	AcademicYear string     `db:"academic_year" json:"academic_year,omitempty"`
	AdvisorID    *uuid.UUID `db:"advisor_id" json:"advisor_id,omitempty"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`

	User    *User     `db:"-" json:"user,omitempty"`
	Advisor *Lecturer `db:"-" json:"advisor,omitempty"`
}

type UpdateAdvisorRequest struct {
	AdvisorID string `json:"advisor_id" validate:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
}
