package repository

import (
	"database/sql"
	model "uas-backend/app/model"

	"github.com/google/uuid"
)

type LecturerRepository interface {
	GetAll(limit, offset int) ([]*model.Lecturer, int, error)
	GetByID(id uuid.UUID) (*model.Lecturer, error)
	GetByUserID(userID uuid.UUID) (*model.Lecturer, error)
	GetAdvisees(lecturerID uuid.UUID, limit, offset int) ([]*model.Student, int, error)
}

type lecturerRepository struct {
	db *sql.DB
}

func NewLecturerRepository(db *sql.DB) LecturerRepository {
	return &lecturerRepository{db: db}
}

func (r *lecturerRepository) GetAll(limit, offset int) ([]*model.Lecturer, int, error) {
	var lecturers []*model.Lecturer
	var total int

	countQuery := `SELECT COUNT(*) FROM lecturers`
	err := r.db.QueryRow(countQuery).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT 
			l.id, l.user_id, l.lecturer_id, l.department, l.created_at,
			u.id, u.username, u.full_name, u.email, u.created_at
		FROM lecturers l
		LEFT JOIN users u ON l.user_id = u.id
		ORDER BY l.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		lecturer := &model.Lecturer{User: &model.User{}}
		err := rows.Scan(
			&lecturer.ID, &lecturer.UserID, &lecturer.LecturerID, &lecturer.Department, &lecturer.CreatedAt,
			&lecturer.User.ID, &lecturer.User.Username, &lecturer.User.FullName,
			&lecturer.User.Email, &lecturer.User.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		lecturers = append(lecturers, lecturer)
	}

	return lecturers, total, nil
}

func (r *lecturerRepository) GetByID(id uuid.UUID) (*model.Lecturer, error) {
	lecturer := &model.Lecturer{User: &model.User{}}

	query := `
		SELECT 
			l.id, l.user_id, l.lecturer_id, l.department, l.created_at,
			u.id, u.username, u.full_name, u.email, u.created_at
		FROM lecturers l
		LEFT JOIN users u ON l.user_id = u.id
		WHERE l.id = $1
	`

	err := r.db.QueryRow(query, id).Scan(
		&lecturer.ID, &lecturer.UserID, &lecturer.LecturerID, &lecturer.Department, &lecturer.CreatedAt,
		&lecturer.User.ID, &lecturer.User.Username, &lecturer.User.FullName,
		&lecturer.User.Email, &lecturer.User.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return lecturer, nil
}

func (r *lecturerRepository) GetByUserID(userID uuid.UUID) (*model.Lecturer, error) {
	lecturer := &model.Lecturer{}

	query := `SELECT id, user_id, lecturer_id, department, created_at FROM lecturers WHERE user_id = $1`

	err := r.db.QueryRow(query, userID).Scan(
		&lecturer.ID, &lecturer.UserID, &lecturer.LecturerID, &lecturer.Department, &lecturer.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return lecturer, nil
}

func (r *lecturerRepository) GetAdvisees(lecturerID uuid.UUID, limit, offset int) ([]*model.Student, int, error) {
	var students []*model.Student
	var total int
	countQuery := `SELECT COUNT(*) FROM students WHERE advisor_id = $1`
	err := r.db.QueryRow(countQuery, lecturerID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT 
			s.id, s.user_id, s.student_id, s.program_study, s.academic_year, s.advisor_id, s.created_at,
			u.id, u.username, u.full_name, u.email, u.created_at
		FROM students s
		LEFT JOIN users u ON s.user_id = u.id
		WHERE s.advisor_id = $1
		ORDER BY s.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, lecturerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		student := &model.Student{User: &model.User{}}
		err := rows.Scan(
			&student.ID, &student.UserID, &student.StudentID, &student.ProgramStudy,
			&student.AcademicYear, &student.AdvisorID, &student.CreatedAt,
			&student.User.ID, &student.User.Username, &student.User.FullName,
			&student.User.Email, &student.User.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		students = append(students, student)
	}

	return students, total, nil
}
