package repository

import (
	"database/sql"
	"strconv"
	"time"
	model "uas-backend/app/model"

	"github.com/google/uuid"
)

type AchievementReferenceRepository interface {
	Create(studentID uuid.UUID, mongoAchievementID string) (*model.AchievementReference, error)
	GetByMongoID(mongoAchievementID string) (*model.AchievementReference, error)
	GetByStudentID(studentID uuid.UUID, status string, limit, offset int) ([]*model.AchievementReference, int, error)
	GetAll(status string, limit, offset int) ([]*model.AchievementReference, int, error)
	UpdateStatus(mongoAchievementID, status string) error
	Submit(mongoAchievementID string) error
	Verify(mongoAchievementID string, verifiedBy uuid.UUID, note string) error
	Reject(mongoAchievementID string, rejectionNote string) error
	Delete(mongoAchievementID string) error
}

type achievementReferenceRepository struct {
	db *sql.DB
}

func NewAchievementReferenceRepository(db *sql.DB) AchievementReferenceRepository {
	return &achievementReferenceRepository{db: db}
}

func (r *achievementReferenceRepository) Create(studentID uuid.UUID, mongoAchievementID string) (*model.AchievementReference, error) {
	ref := &model.AchievementReference{
		ID:                 uuid.New(),
		StudentID:          studentID,
		MongoAchievementID: mongoAchievementID,
		Status:             "draft",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	query := `
		INSERT INTO achievement_references (id, student_id, mongo_achievement_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, student_id, mongo_achievement_id, status, submitted_at, verified_at, verified_by, rejection_note, created_at, updated_at
	`

	err := r.db.QueryRow(query, ref.ID, ref.StudentID, ref.MongoAchievementID, ref.Status, ref.CreatedAt, ref.UpdatedAt).Scan(
		&ref.ID, &ref.StudentID, &ref.MongoAchievementID, &ref.Status,
		&ref.SubmittedAt, &ref.VerifiedAt, &ref.VerifiedBy, &ref.RejectionNote,
		&ref.CreatedAt, &ref.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return ref, nil
}

func (r *achievementReferenceRepository) GetByMongoID(mongoAchievementID string) (*model.AchievementReference, error) {
	ref := &model.AchievementReference{}
	query := `
		SELECT id, student_id, mongo_achievement_id, status, submitted_at, verified_at, verified_by, rejection_note, created_at, updated_at
		FROM achievement_references
		WHERE mongo_achievement_id = $1 AND status != 'deleted'
	`

	err := r.db.QueryRow(query, mongoAchievementID).Scan(
		&ref.ID, &ref.StudentID, &ref.MongoAchievementID, &ref.Status,
		&ref.SubmittedAt, &ref.VerifiedAt, &ref.VerifiedBy, &ref.RejectionNote,
		&ref.CreatedAt, &ref.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return ref, nil
}

func (r *achievementReferenceRepository) GetByStudentID(studentID uuid.UUID, status string, limit, offset int) ([]*model.AchievementReference, int, error) {
	var refs []*model.AchievementReference
	var total int

	countQuery := `SELECT COUNT(*) FROM achievement_references WHERE student_id = $1 AND status != 'deleted'`
	args := []interface{}{studentID}

	if status != "" {
		countQuery += " AND status = $2"
		args = append(args, status)
	}

	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, student_id, mongo_achievement_id, status, submitted_at, verified_at, verified_by, rejection_note, created_at, updated_at
		FROM achievement_references
		WHERE student_id = $1 AND status != 'deleted'
	`

	queryArgs := []interface{}{studentID}
	argPos := 2

	if status != "" {
		query += " AND status = $" + strconv.Itoa(argPos)
		queryArgs = append(queryArgs, status)
		argPos++
	}

	query += " ORDER BY created_at DESC LIMIT $" + strconv.Itoa(argPos) + " OFFSET $" + strconv.Itoa(argPos+1)
	queryArgs = append(queryArgs, limit, offset)

	rows, err := r.db.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		ref := &model.AchievementReference{}
		err := rows.Scan(
			&ref.ID, &ref.StudentID, &ref.MongoAchievementID, &ref.Status,
			&ref.SubmittedAt, &ref.VerifiedAt, &ref.VerifiedBy, &ref.RejectionNote,
			&ref.CreatedAt, &ref.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		refs = append(refs, ref)
	}

	return refs, total, nil
}

func (r *achievementReferenceRepository) GetAll(status string, limit, offset int) ([]*model.AchievementReference, int, error) {
	var refs []*model.AchievementReference
	var total int

	// Count total
	countQuery := `SELECT COUNT(*) FROM achievement_references WHERE status != 'deleted'`
	args := []interface{}{}

	if status != "" {
		countQuery += " AND status = $1"
		args = append(args, status)
	}

	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get data
	query := `
		SELECT ar.id, ar.student_id, ar.mongo_achievement_id, ar.status, ar.submitted_at, ar.verified_at, ar.verified_by, ar.rejection_note, ar.created_at, ar.updated_at
		FROM achievement_references ar
		WHERE status != 'deleted'
	`

	queryArgs := []interface{}{}
	argPos := 1

	if status != "" {
		query += " AND ar.status = $" + strconv.Itoa(argPos)
		queryArgs = append(queryArgs, status)
		argPos++
	}

	query += " ORDER BY ar.created_at DESC LIMIT $" + strconv.Itoa(argPos) + " OFFSET $" + strconv.Itoa(argPos+1)
	queryArgs = append(queryArgs, limit, offset)

	rows, err := r.db.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		ref := &model.AchievementReference{}
		err := rows.Scan(
			&ref.ID, &ref.StudentID, &ref.MongoAchievementID, &ref.Status,
			&ref.SubmittedAt, &ref.VerifiedAt, &ref.VerifiedBy, &ref.RejectionNote,
			&ref.CreatedAt, &ref.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		refs = append(refs, ref)
	}

	return refs, total, nil
}

func (r *achievementReferenceRepository) UpdateStatus(mongoAchievementID, status string) error {
	query := `UPDATE achievement_references SET status = $1, updated_at = $2 WHERE mongo_achievement_id = $3`
	_, err := r.db.Exec(query, status, time.Now(), mongoAchievementID)
	return err
}

func (r *achievementReferenceRepository) Submit(mongoAchievementID string) error {
	query := `UPDATE achievement_references SET status = 'submitted', submitted_at = $1, updated_at = $2 WHERE mongo_achievement_id = $3`
	_, err := r.db.Exec(query, time.Now(), time.Now(), mongoAchievementID)
	return err
}

func (r *achievementReferenceRepository) Verify(mongoAchievementID string, verifiedBy uuid.UUID, note string) error {
	var query string
	var args []interface{}

	if note != "" {
		query = `UPDATE achievement_references SET status = 'verified', verified_by = $1, verified_at = $2, rejection_note = $3, updated_at = $4 WHERE mongo_achievement_id = $5`
		args = []interface{}{verifiedBy, time.Now(), note, time.Now(), mongoAchievementID}
	} else {
		query = `UPDATE achievement_references SET status = 'verified', verified_by = $1, verified_at = $2, updated_at = $3 WHERE mongo_achievement_id = $4`
		args = []interface{}{verifiedBy, time.Now(), time.Now(), mongoAchievementID}
	}

	_, err := r.db.Exec(query, args...)
	return err
}

func (r *achievementReferenceRepository) Reject(mongoAchievementID string, rejectionNote string) error {
	query := `UPDATE achievement_references SET status = 'rejected', rejection_note = $1, updated_at = $2 WHERE mongo_achievement_id = $3`
	_, err := r.db.Exec(query, rejectionNote, time.Now(), mongoAchievementID)
	return err
}

func (r *achievementReferenceRepository) Delete(mongoAchievementID string) error {
	query := `UPDATE achievement_references SET status = 'deleted', updated_at = $1 WHERE mongo_achievement_id = $2`
	_, err := r.db.Exec(query, time.Now(), mongoAchievementID)
	return err
}
