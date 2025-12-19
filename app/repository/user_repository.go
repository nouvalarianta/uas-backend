package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
	model "uas-backend/app/model"

	"github.com/google/uuid"
)

type UserRepository interface {
	Login(username string) (*model.User, string, error)
	GetAll(search, sortBy, order string, limit, offset int) ([]*model.User, error)
	GetByID(id uuid.UUID) (*model.User, error)
	CheckUsernameExists(username string) (bool, error)
	CheckEmailExists(email string) (bool, error)
	CheckRoleExists(roleID uuid.UUID) (bool, error)
	Create(username, email, passwordHash, fullName string, roleID uuid.UUID) (*model.User, error)
	Update(id uuid.UUID, req *model.UpdateUserRequest) (*model.User, error)
	Delete(id uuid.UUID) error
	GetPermissionsByRoleID(roleID uuid.UUID) ([]string, error)
	ReplaceRole(id uuid.UUID, roleID uuid.UUID) (*model.User, error)
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

func (r *userrepository) GetAll(search, sortBy, order string, limit, offset int) ([]*model.User, error) {
	allowedSortBy := map[string]bool{
		"id":         true,
		"username":   true,
		"email":      true,
		"fullname":   true,
		"role_id":    true,
		"is_active":  true,
		"created_at": true,
		"updated_at": true,
	}
	if !allowedSortBy[sortBy] {
		sortBy = "created_at"
	}

	order = strings.ToUpper(order)
	if order != "ASC" && order != "DESC" {
		order = "DESC"
	}

	baseQuery := `
		SELECT u.id, u.username, u.email, u.full_name, u.role_id, u.is_active, u.created_at, u.updated_at
		FROM users u
		WHERE u.is_active IS NULL
	`
	whereClause := ""
	args := []interface{}{}
	argCounter := 1

	if search != "" {
		whereClause = fmt.Sprintf(" AND (u.username ILIKE $%d OR u.email ILIKE $%d OR u.full_name ILIKE $%d)", argCounter, argCounter, argCounter)
		args = append(args, "%"+search+"%")
		argCounter++
	}

	query := fmt.Sprintf("%s %s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		baseQuery, whereClause, sortBy, order, argCounter, argCounter+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userlist []*model.User
	for rows.Next() {
		user := &model.User{}
		var createdAt, updatedAt sql.NullTime

		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.FullName,
			&user.RoleID, &user.IsActive, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, err
		}
		if createdAt.Valid {
			user.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			user.UpdatedAt = updatedAt.Time
		}
		userlist = append(userlist, user)
	}

	return userlist, nil
}

func (r *userrepository) GetByID(id uuid.UUID) (*model.User, error) {
	user := &model.User{
		Role: &model.Role{},
	}
	var createdAt, updatedAt sql.NullTime

	query := `
		SELECT u.id, u.username, u.email, u.full_name, u.role_id, u.is_active, u.created_at, u.updated_at,
		       r.id, r.name
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1 AND u.is_active IS NULL
	`

	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.FullName,
		&user.RoleID, &user.IsActive, &createdAt, &updatedAt,
		&user.Role.ID, &user.Role.Name,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if createdAt.Valid {
		user.CreatedAt = createdAt.Time
	}

	if updatedAt.Valid {
		user.UpdatedAt = updatedAt.Time
	}

	return user, nil
}

// CheckUsernameExists cek apakah username sudah ada
func (r *userrepository) CheckUsernameExists(username string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`
	err := r.db.QueryRow(query, username).Scan(&exists)
	return exists, err
}

// cek apakah email sudah ada atau belum
func (r *userrepository) CheckEmailExists(email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	err := r.db.QueryRow(query, email).Scan(&exists)
	return exists, err
}

// cek apakah role ada atau tidak
func (r *userrepository) CheckRoleExists(roleID uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM roles WHERE id = $1)`
	err := r.db.QueryRow(query, roleID).Scan(&exists)
	return exists, err
}

func (r *userrepository) Create(username, email, passwordHash, fullName string, roleID uuid.UUID) (*model.User, error) {
	user := &model.User{
		ID:        uuid.New(),
		Username:  username,
		Email:     email,
		FullName:  fullName,
		RoleID:    roleID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	query := `
		INSERT INTO users (id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Exec(query,
		user.ID, user.Username, user.Email, passwordHash, user.FullName,
		user.RoleID, nil, user.CreatedAt, user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userrepository) Update(id uuid.UUID, req *model.UpdateUserRequest) (*model.User, error) {
	now := time.Now()
	user := &model.User{}

	query := `
		UPDATE users
		SET username = $1, email = $2, full_name = $3, role_id = $4, updated_at = $5
		WHERE id = $6 AND is_active IS NULL
		RETURNING id, username, email, full_name, role_id, is_active, created_at, updated_at
	`

	err := r.db.QueryRow(
		query, req.Username, req.Email, req.FullName, req.RoleID, now, id,
	).Scan(
		&user.ID, &user.Username, &user.Email, &user.FullName, &user.RoleID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)

	return user, err
}

func (r *userrepository) Delete(id uuid.UUID) error {
	query := `UPDATE users SET is_active = $1, updated_at = $2 WHERE id = $3 AND is_active IS NULL`
	now := time.Now()
	result, err := r.db.Exec(query, now, now, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *userrepository) GetPermissionsByRoleID(roleID uuid.UUID) ([]string, error) {
	query := `
		SELECT p.name
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.name
	`

	rows, err := r.db.Query(query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var permissionName string
		if err := rows.Scan(&permissionName); err != nil {
			return nil, err
		}
		permissions = append(permissions, permissionName)
	}

	return permissions, nil
}

func (r *userrepository) ReplaceRole(id uuid.UUID, roleID uuid.UUID) (*model.User, error) {
	now := time.Now()
	user := &model.User{}

	query := `
		UPDATE users
		SET role_id = $1, updated_at = $2
		WHERE id = $3 AND is_active IS NULL
		RETURNING id, role_id, updated_at
	`

	err := r.db.QueryRow(
		query, roleID, now, id,
	).Scan(
		&user.ID, &user.RoleID, &user.UpdatedAt,
	)

	return user, err

}
