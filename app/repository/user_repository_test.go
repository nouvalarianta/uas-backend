package repository

import (
	"database/sql"
	"testing"
	"uas-backend/app/model"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestLogin_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	userID := uuid.New()
	roleID := uuid.New()
	username := "testuser"
	email := "test@example.com"
	fullName := "Test User"
	passwordHash := "$2a$10$abcdefghijklmnopqrstuvwxyz"

	rows := sqlmock.NewRows([]string{
		"id", "username", "email", "full_name", "role_id", "is_active", "password_hash",
		"role_id", "role_name", "role_description",
	}).AddRow(
		userID, username, email, fullName, roleID, nil, passwordHash,
		roleID, "Admin", "Administrator Role",
	)

	mock.ExpectQuery("SELECT (.+) FROM users u LEFT JOIN roles r").
		WithArgs(username).
		WillReturnRows(rows)

	user, hash, err := repo.Login(username)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, username, user.Username)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, passwordHash, hash)
	assert.NotNil(t, user.Role)
	assert.Equal(t, "Admin", user.Role.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLogin_UserNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	username := "nonexistent"

	mock.ExpectQuery("SELECT (.+) FROM users u LEFT JOIN roles r").
		WithArgs(username).
		WillReturnError(sql.ErrNoRows)

	user, hash, err := repo.Login(username)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Empty(t, hash)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	userID := uuid.New()
	roleID := uuid.New()

	rows := sqlmock.NewRows([]string{
		"id", "username", "email", "full_name", "role_id", "is_active", "created_at", "updated_at",
		"role_id", "role_name",
	}).AddRow(
		userID, "testuser", "test@example.com", "Test User", roleID, nil, sql.NullTime{Valid: true}, sql.NullTime{Valid: true},
		roleID, "Admin",
	)

	mock.ExpectQuery("SELECT (.+) FROM users u LEFT JOIN roles r").
		WithArgs(userID).
		WillReturnRows(rows)

	user, err := repo.GetByID(userID)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, "testuser", user.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	userID := uuid.New()

	mock.ExpectQuery("SELECT (.+) FROM users u LEFT JOIN roles r").
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetByID(userID)

	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckUsernameExists(t *testing.T) {
	tests := []struct {
		name     string
		username string
		exists   bool
		wantErr  bool
	}{
		{
			name:     "Username exists",
			username: "existinguser",
			exists:   true,
			wantErr:  false,
		},
		{
			name:     "Username does not exist",
			username: "newuser",
			exists:   false,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			repo := NewUserRepository(db)

			rows := sqlmock.NewRows([]string{"exists"}).AddRow(tt.exists)
			mock.ExpectQuery("SELECT EXISTS").
				WithArgs(tt.username).
				WillReturnRows(rows)

			exists, err := repo.CheckUsernameExists(tt.username)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.exists, exists)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCheckEmailExists(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		exists  bool
		wantErr bool
	}{
		{
			name:    "Email exists",
			email:   "existing@example.com",
			exists:  true,
			wantErr: false,
		},
		{
			name:    "Email does not exist",
			email:   "new@example.com",
			exists:  false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			repo := NewUserRepository(db)

			rows := sqlmock.NewRows([]string{"exists"}).AddRow(tt.exists)
			mock.ExpectQuery("SELECT EXISTS").
				WithArgs(tt.email).
				WillReturnRows(rows)

			exists, err := repo.CheckEmailExists(tt.email)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.exists, exists)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCreate_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	roleID := uuid.New()
	username := "newuser"
	email := "new@example.com"
	passwordHash := "$2a$10$hash"
	fullName := "New User"

	mock.ExpectExec("INSERT INTO users").
		WithArgs(sqlmock.AnyArg(), username, email, passwordHash, fullName, roleID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	user, err := repo.Create(username, email, passwordHash, fullName, roleID)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, username, user.Username)
	assert.Equal(t, email, user.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDelete_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	userID := uuid.New()

	mock.ExpectExec("UPDATE users SET deleted_at").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Delete(userID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPermissionsByRoleID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	roleID := uuid.New()

	rows := sqlmock.NewRows([]string{"permission_name"}).
		AddRow("user:read").
		AddRow("user:write").
		AddRow("user:delete")

	mock.ExpectQuery("SELECT p.name as permission_name").
		WithArgs(roleID).
		WillReturnRows(rows)

	permissions, err := repo.GetPermissionsByRoleID(roleID)

	assert.NoError(t, err)
	assert.Len(t, permissions, 3)
	assert.Contains(t, permissions, "user:read")
	assert.Contains(t, permissions, "user:write")
	assert.Contains(t, permissions, "user:delete")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPermissionsByRoleID_NoPermissions(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	roleID := uuid.New()

	rows := sqlmock.NewRows([]string{"permission_name"})

	mock.ExpectQuery("SELECT p.name as permission_name").
		WithArgs(roleID).
		WillReturnRows(rows)

	permissions, err := repo.GetPermissionsByRoleID(roleID)

	assert.NoError(t, err)
	assert.Len(t, permissions, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdate_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	userID := uuid.New()
	roleID := uuid.New()

	updateReq := &model.UpdateUserRequest{
		Username: "updateduser",
		Email:    "updated@example.com",
		FullName: "Updated User",
		RoleID:   roleID.String(),
	}

	rows := sqlmock.NewRows([]string{
		"id", "username", "email", "full_name", "role_id", "is_active", "created_at", "updated_at",
	}).AddRow(
		userID, "updateduser", "updated@example.com", "Updated User", roleID, nil, sql.NullTime{Valid: true}, sql.NullTime{Valid: true},
	)

	mock.ExpectQuery("UPDATE users SET").
		WithArgs("updateduser", "updated@example.com", "Updated User", roleID.String(), sqlmock.AnyArg(), userID).
		WillReturnRows(rows)

	user, err := repo.Update(userID, updateReq)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "updateduser", user.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}
