package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	Username  string     `db:"username" json:"username"`
	Email     string     `db:"email" json:"email"`
	FullName  string     `db:"full_name" json:"full_name"`
	RoleID    uuid.UUID  `db:"role_id" json:"role_id"`
	IsActive  *time.Time `db:"is_active" json:"is_active,omitempty"` // NULL = active, NOT NULL = deleted at timestamp
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`

	Role *Role `db:"-" json:"role,omitempty"`
}

// CreateUserRequest untuk request body saat create user
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	FullName string `json:"full_name" validate:"required,min=3,max=100"`
	RoleID   string `json:"role_id" validate:"required,uuid"` // String UUID dari frontend
}

type UpdateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	FullName string `json:"full_name" validate:"required,min=3,max=100"`
	RoleID   string `json:"role_id" validate:"required,uuid"` // String UUID dari frontend
}

type LoginRequest struct {
	Username string `json:"username" example:"admin"`
	Password string `json:"password" example:"password123"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type ReplaceRoleRequest struct {
	RoleID string `json:"role_id" validate:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
}
