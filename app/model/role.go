package model

import (
	"time"

	"github.com/google/uuid"
)

type Role struct {
	ID          uuid.UUID `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description,omitempty"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type Permission struct {
	ID          uuid.UUID `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Resource    string    `db:"resource" json:"resource"`
	Action      string    `db:"action" json:"action"`
	Description string    `db:"description" json:"description,omitempty"`
}

type RolePermission struct {
	RoleID       uuid.UUID `db:"role_id" json:"role_id"`
	PermissionID uuid.UUID `db:"permission_id" json:"permission_id"`
}
