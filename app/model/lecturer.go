package model

import (
	"time"

	"github.com/google/uuid"
)

type Lecturer struct {
	ID         uuid.UUID `db:"id" json:"id"`
	UserID     uuid.UUID `db:"user_id" json:"user_id"`
	LecturerID string    `db:"lecturer_id" json:"lecturer_id"`
	Department string    `db:"department" json:"department,omitempty"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`

	User *User `db:"-" json:"user,omitempty"`
}
