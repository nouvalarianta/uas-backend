package repository

import (
	"database/sql"
	model "uas-backend/app/model"

	"github.com/google/uuid"
)

type StudentRepository interface {
	GetAll(limit, offset int) ([]*model.Student, int, error)
	GetByID(id uuid.UUID) (*model.Student, error)
	GetByUserID(userID uuid.UUID) (*model.Student, error)
	UpdateAdvisor(studentID, advisorID uuid.UUID) error
}

type studentRepository struct {
	db *sql.DB
}

func NewStudentRepository(db *sql.DB) StudentRepository {
	return &studentRepository{db: db}
}

func (r *studentRepository) GetAll(limit, offset int) ([]*model.Student, int, error) {
	var students []*model.Student
	var total int

	countQuery := `SELECT COUNT(*) FROM students`
	err := r.db.QueryRow(countQuery).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT 
			s.id, s.user_id, s.student_id, s.program_study, s.academic_year, s.advisor_id, s.created_at,
			u.id, u.username, u.full_name, u.email, u.created_at
		FROM students s
		LEFT JOIN users u ON s.user_id = u.id
		ORDER BY s.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, limit, offset)
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

func (r *studentRepository) GetByID(id uuid.UUID) (*model.Student, error) {
	student := &model.Student{User: &model.User{}}

	query := `
		SELECT 
			s.id, s.user_id, s.student_id, s.program_study, s.academic_year, s.advisor_id, s.created_at,
			u.id, u.username, u.full_name, u.email, u.created_at
		FROM students s
		LEFT JOIN users u ON s.user_id = u.id
		WHERE s.id = $1
	`

	err := r.db.QueryRow(query, id).Scan(
		&student.ID, &student.UserID, &student.StudentID, &student.ProgramStudy,
		&student.AcademicYear, &student.AdvisorID, &student.CreatedAt,
		&student.User.ID, &student.User.Username, &student.User.FullName,
		&student.User.Email, &student.User.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if student.AdvisorID != nil {
		advisor := &model.Lecturer{User: &model.User{}}
		advisorQuery := `
			SELECT 
				l.id, l.user_id, l.lecturer_id, l.department, l.created_at,
				u.id, u.username, u.full_name, u.email, u.created_at
			FROM lecturers l
			LEFT JOIN users u ON l.user_id = u.id
			WHERE l.id = $1
		`
		err := r.db.QueryRow(advisorQuery, student.AdvisorID).Scan(
			&advisor.ID, &advisor.UserID, &advisor.LecturerID, &advisor.Department, &advisor.CreatedAt,
			&advisor.User.ID, &advisor.User.Username, &advisor.User.FullName,
			&advisor.User.Email, &advisor.User.CreatedAt,
		)
		if err == nil {
			student.Advisor = advisor
		}
	}

	return student, nil
}

func (r *studentRepository) GetByUserID(userID uuid.UUID) (*model.Student, error) {
	student := &model.Student{}

	query := `SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at FROM students WHERE user_id = $1`

	err := r.db.QueryRow(query, userID).Scan(
		&student.ID, &student.UserID, &student.StudentID, &student.ProgramStudy,
		&student.AcademicYear, &student.AdvisorID, &student.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return student, nil
}

func (r *studentRepository) UpdateAdvisor(studentID, advisorID uuid.UUID) error {
	query := `UPDATE students SET advisor_id = $1 WHERE id = $2`
	_, err := r.db.Exec(query, advisorID, studentID)
	return err
}
