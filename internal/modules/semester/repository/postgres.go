package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/M1ralai/notly-api/internal/modules/semester/domain"
	"github.com/jmoiron/sqlx"
)

type PostgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) SemesterRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, semester *domain.Semester) (*domain.Semester, error) {
	model := FromDomain(semester)
	query := `
		INSERT INTO semesters (user_id, name, start_date, end_date, is_current, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	now := time.Now()
	err := r.db.QueryRowContext(
		ctx,
		query,
		model.UserID,
		model.Name,
		model.StartDate,
		model.EndDate,
		model.IsCurrent,
		now,
		now,
	).Scan(&model.ID, &model.CreatedAt, &model.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return model.ToDomain(), nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id int) (*domain.Semester, error) {
	var model SemesterModel
	query := `
		SELECT id, user_id, name, start_date, end_date, is_current, created_at, updated_at
		FROM semesters
		WHERE id = $1 AND deleted_at IS NULL
	`

	err := r.db.GetContext(ctx, &model, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

func (r *PostgresRepository) GetByUserID(ctx context.Context, userID int) ([]*domain.Semester, error) {
	var models []SemesterModel
	query := `
		SELECT id, user_id, name, start_date, end_date, is_current, created_at, updated_at
		FROM semesters
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY start_date DESC
	`

	err := r.db.SelectContext(ctx, &models, query, userID)
	if err != nil {
		return nil, err
	}

	semesters := make([]*domain.Semester, len(models))
	for i, model := range models {
		semesters[i] = model.ToDomain()
	}

	return semesters, nil
}

func (r *PostgresRepository) GetCurrent(ctx context.Context, userID int) (*domain.Semester, error) {
	var model SemesterModel
	query := `
		SELECT id, user_id, name, start_date, end_date, is_current, created_at, updated_at
		FROM semesters
		WHERE user_id = $1 AND is_current = TRUE AND deleted_at IS NULL
		LIMIT 1
	`

	err := r.db.GetContext(ctx, &model, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

func (r *PostgresRepository) Update(ctx context.Context, semester *domain.Semester) error {
	model := FromDomain(semester)
	query := `
		UPDATE semesters
		SET name = $1, start_date = $2, end_date = $3, is_current = $4, updated_at = $5
		WHERE id = $6
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		model.Name,
		model.StartDate,
		model.EndDate,
		model.IsCurrent,
		time.Now(),
		model.ID,
	)

	return err
}

func (r *PostgresRepository) Delete(ctx context.Context, id int) error {
	query := `UPDATE semesters SET deleted_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

func (r *PostgresRepository) GetUpdatedSince(ctx context.Context, userID int, since time.Time) ([]*domain.Semester, error) {
	var models []SemesterModel
	query := `
		SELECT id, user_id, name, start_date, end_date, is_current, created_at, updated_at
		FROM semesters
		WHERE user_id = $1 AND updated_at > $2 AND deleted_at IS NULL
		ORDER BY updated_at ASC
	`
	if err := r.db.SelectContext(ctx, &models, query, userID, since); err != nil {
		return nil, err
	}
	semesters := make([]*domain.Semester, len(models))
	for i, model := range models {
		semesters[i] = model.ToDomain()
	}
	return semesters, nil
}

func (r *PostgresRepository) GetDeletedSince(ctx context.Context, userID int, since time.Time) ([]int, error) {
	var ids []int
	query := `SELECT id FROM semesters WHERE user_id = $1 AND deleted_at > $2`
	if err := r.db.SelectContext(ctx, &ids, query, userID, since); err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if ids == nil {
		ids = make([]int, 0)
	}
	return ids, nil
}
