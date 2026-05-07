package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/M1ralai/notly-api/internal/modules/goal/domain"
	"github.com/jmoiron/sqlx"
)

type postgresRepository struct{ db *sqlx.DB }

func NewPostgresRepository(db *sqlx.DB) GoalRepository { return &postgresRepository{db: db} }

func (r *postgresRepository) Create(ctx context.Context, goal *domain.Goal) (*domain.Goal, error) {
	query := `INSERT INTO goals (user_id, life_area_id, title, description, target_date, is_completed, priority, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id, created_at, updated_at`
	now := time.Now()
	model := FromDomain(goal)
	err := r.db.QueryRowxContext(ctx, query, model.UserID, model.LifeAreaID, model.Title, model.Description, model.TargetDate, false, model.Priority, now, now).Scan(&model.ID, &model.CreatedAt, &model.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *postgresRepository) GetByID(ctx context.Context, id int) (*domain.Goal, error) {
	query := `SELECT id, user_id, life_area_id, title, description, target_date, is_completed, completed_at, priority, created_at, updated_at FROM goals WHERE id = $1 AND deleted_at IS NULL`
	var model GoalModel
	err := r.db.GetContext(ctx, &model, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *postgresRepository) GetByUserID(ctx context.Context, userID int) ([]*domain.Goal, error) {
	query := `SELECT id, user_id, life_area_id, title, description, target_date, is_completed, completed_at, priority, created_at, updated_at FROM goals WHERE user_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`
	var models []GoalModel
	if err := r.db.SelectContext(ctx, &models, query, userID); err != nil {
		return nil, err
	}
	goals := make([]*domain.Goal, len(models))
	for i, m := range models {
		goals[i] = m.ToDomain()
	}
	return goals, nil
}

func (r *postgresRepository) Update(ctx context.Context, goal *domain.Goal) error {
	query := `UPDATE goals SET title = $1, description = $2, target_date = $3, is_completed = $4, completed_at = $5, priority = $6, life_area_id = $7, updated_at = $8 WHERE id = $9`
	model := FromDomain(goal)
	_, err := r.db.ExecContext(ctx, query, model.Title, model.Description, model.TargetDate, model.IsCompleted, model.CompletedAt, model.Priority, model.LifeAreaID, time.Now(), model.ID)
	return err
}

func (r *postgresRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE goals SET deleted_at = $1 WHERE id = $2`, time.Now(), id)
	return err
}

func (r *postgresRepository) GetUpdatedSince(ctx context.Context, userID int, since time.Time) ([]*domain.Goal, error) {
	query := `
		SELECT id, user_id, life_area_id, title, description, target_date, is_completed, completed_at, priority, created_at, updated_at 
		FROM goals 
		WHERE user_id = $1 AND updated_at > $2 AND deleted_at IS NULL 
		ORDER BY updated_at ASC
	`
	var models []GoalModel
	if err := r.db.SelectContext(ctx, &models, query, userID, since); err != nil {
		return nil, err
	}
	goals := make([]*domain.Goal, len(models))
	for i, m := range models {
		goals[i] = m.ToDomain()
	}
	return goals, nil
}

func (r *postgresRepository) GetDeletedSince(ctx context.Context, userID int, since time.Time) ([]int, error) {
	query := `SELECT id FROM goals WHERE user_id = $1 AND deleted_at > $2`
	var ids []int
	if err := r.db.SelectContext(ctx, &ids, query, userID, since); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if ids == nil {
		ids = make([]int, 0)
	}
	return ids, nil
}

func (r *postgresRepository) CountMilestones(ctx context.Context, goalID int) (total int, completed int, err error) {
	query := `SELECT COUNT(*) as total, COUNT(*) FILTER (WHERE is_completed = true) as completed FROM milestones WHERE goal_id = $1`
	err = r.db.QueryRowxContext(ctx, query, goalID).Scan(&total, &completed)
	return
}
