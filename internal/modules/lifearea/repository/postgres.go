package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/M1ralai/notly-api/internal/modules/lifearea/domain"
	"github.com/jmoiron/sqlx"
)

type postgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) LifeAreaRepository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(ctx context.Context, lifeArea *domain.LifeArea) (*domain.LifeArea, error) {
	query := `
		INSERT INTO life_areas (user_id, name, icon, color, display_order, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	now := time.Now()
	model := FromDomain(lifeArea)

	err := r.db.QueryRowxContext(
		ctx, query,
		model.UserID,
		model.Name,
		model.Icon,
		model.Color,
		model.DisplayOrder,
		now,
	).Scan(&model.ID, &model.CreatedAt)

	if err != nil {
		return nil, err
	}

	return model.ToDomain(), nil
}

func (r *postgresRepository) GetByID(ctx context.Context, id int) (*domain.LifeArea, error) {
	query := `
		SELECT id, user_id, name, icon, color, display_order, created_at
		FROM life_areas
		WHERE id = $1 AND deleted_at IS NULL
	`

	var model LifeAreaModel
	err := r.db.GetContext(ctx, &model, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

func (r *postgresRepository) GetByUserID(ctx context.Context, userID int) ([]*domain.LifeArea, error) {
	query := `
		SELECT id, user_id, name, icon, color, display_order, created_at
		FROM life_areas
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY display_order ASC, created_at ASC
	`

	var models []LifeAreaModel
	err := r.db.SelectContext(ctx, &models, query, userID)
	if err != nil {
		return nil, err
	}

	areas := make([]*domain.LifeArea, len(models))
	for i, m := range models {
		areas[i] = m.ToDomain()
	}

	return areas, nil
}

func (r *postgresRepository) Update(ctx context.Context, lifeArea *domain.LifeArea) error {
	query := `
		UPDATE life_areas
		SET name = $1, icon = $2, color = $3, display_order = $4, updated_at = $5
		WHERE id = $6
	`

	model := FromDomain(lifeArea)
	_, err := r.db.ExecContext(
		ctx, query,
		model.Name,
		model.Icon,
		model.Color,
		model.DisplayOrder,
		time.Now(),
		model.ID,
	)

	return err
}

func (r *postgresRepository) Delete(ctx context.Context, id int) error {
	query := `UPDATE life_areas SET deleted_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

func (r *postgresRepository) GetUpdatedSince(ctx context.Context, userID int, since time.Time) ([]*domain.LifeArea, error) {
	query := `
		SELECT id, user_id, name, icon, color, display_order, created_at
		FROM life_areas
		WHERE user_id = $1 AND (COALESCE(updated_at, created_at) > $2) AND deleted_at IS NULL
		ORDER BY display_order ASC
	`
	var models []LifeAreaModel
	err := r.db.SelectContext(ctx, &models, query, userID, since)
	if err != nil {
		return nil, err
	}
	areas := make([]*domain.LifeArea, len(models))
	for i, m := range models {
		areas[i] = m.ToDomain()
	}
	return areas, nil
}

func (r *postgresRepository) GetDeletedSince(ctx context.Context, userID int, since time.Time) ([]int, error) {
	query := `
		SELECT id FROM life_areas
		WHERE user_id = $1 AND deleted_at > $2
	`
	var ids []int
	err := r.db.SelectContext(ctx, &ids, query, userID, since)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if ids == nil {
		ids = make([]int, 0)
	}
	return ids, nil
}
