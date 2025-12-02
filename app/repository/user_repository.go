package repository

import (
	"database/sql"
	"uas-backend/app/model"
	// "strings"
	// "time"
	// "github.com/google/uuid"
)

type UserRepository interface {
	Login(username string) (*model.User, string, error)
	// GetAll() ([]*model.User, error)
	// GetByID(id uuid.UUID) (*model.User, error)
	// GetByUsername(username string) (*model.User, error)
	// Create(user *model.User) error
	// Update(user *model.User) error
	// Delete(id uuid.UUID) error
}

type userrepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userrepository{db: db}
}

func (r *userrepository) Login(username string) (*model.User, string, error) {
	var user model.User
	var passwordHash string
	var role model.Role

	// Query by username and JOIN with roles table to get role information
	query := `
		SELECT 
			u.id, u.username, u.email, u.full_name, u.role_id, u.is_active, u.password_hash,
			r.id, r.name, r.description
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.username = $1
	`

	err := r.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.FullName, &user.RoleID, &user.IsActive, &passwordHash,
		&role.ID, &role.Name, &role.Description,
	)
	if err != nil {
		return nil, "", err
	}

	user.Role = &role
	return &user, passwordHash, nil
}
