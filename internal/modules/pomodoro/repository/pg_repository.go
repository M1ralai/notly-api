package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/M1ralai/notly-api/internal/modules/pomodoro/domain"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type pgRepository struct {
	db *sqlx.DB
}

func NewPgRepository(db *sqlx.DB) domain.Repository {
	return &pgRepository{db: db}
}

type PomodoroSessionModel struct {
	ID              int            `db:"id"`
	UserID          int            `db:"user_id"`
	CourseID        sql.NullInt64  `db:"course_id"`
	DurationMinutes int            `db:"duration_minutes"`
	Notes           sql.NullString `db:"notes"`
	CreatedAt       sql.NullTime   `db:"created_at"`
}

type PomodoroSettingsModel struct {
	UserID       int           `db:"user_id"`
	StudyPresets pq.Int64Array `db:"study_presets"`
	StudyColor   string        `db:"study_color"`
	UpdatedAt    sql.NullTime  `db:"updated_at"`
}

func (m *PomodoroSettingsModel) ToDomain() domain.PomodoroSettings {
	presets := make([]int, len(m.StudyPresets))
	for i, p := range m.StudyPresets {
		presets[i] = int(p)
	}

	return domain.PomodoroSettings{
		UserID:       m.UserID,
		StudyPresets: presets,
		StudyColor:   m.StudyColor,
		UpdatedAt:    m.UpdatedAt.Time,
	}
}

func (m *PomodoroSessionModel) ToDomain() domain.PomodoroSession {
	var courseID *int
	if m.CourseID.Valid {
		id := int(m.CourseID.Int64)
		courseID = &id
	}

	return domain.PomodoroSession{
		ID:              m.ID,
		UserID:          m.UserID,
		CourseID:        courseID,
		DurationMinutes: m.DurationMinutes,
		Notes:           m.Notes.String,
		CreatedAt:       m.CreatedAt.Time,
	}
}

func (r *pgRepository) Save(ctx context.Context, session *domain.PomodoroSession) error {
	query := `
		INSERT INTO pomodoro_sessions (user_id, course_id, duration_minutes, notes)
		VALUES (:user_id, :course_id, :duration_minutes, :notes)
		RETURNING id, created_at
	`

	model := PomodoroSessionModel{
		UserID:          session.UserID,
		DurationMinutes: session.DurationMinutes,
		Notes:           sql.NullString{String: session.Notes, Valid: session.Notes != ""},
	}

	if session.CourseID != nil {
		model.CourseID = sql.NullInt64{Int64: int64(*session.CourseID), Valid: true}
	}

	rows, err := r.db.NamedQueryContext(ctx, query, model)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&session.ID, &session.CreatedAt); err != nil {
			return err
		}
	} else {
		return errors.New("failed to retrieve inserted id")
	}

	return nil
}

func (r *pgRepository) FindAllByUserID(ctx context.Context, userID int) ([]domain.PomodoroSession, error) {
	query := `
		SELECT id, user_id, course_id, duration_minutes, notes, created_at
		FROM pomodoro_sessions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	var models []PomodoroSessionModel
	if err := r.db.SelectContext(ctx, &models, query, userID); err != nil {
		return nil, err
	}

	sessions := make([]domain.PomodoroSession, len(models))
	for i, m := range models {
		sessions[i] = m.ToDomain()
	}

	return sessions, nil
}

func (r *pgRepository) FindByCourseID(ctx context.Context, userID, courseID int) ([]domain.PomodoroSession, error) {
	query := `
		SELECT id, user_id, course_id, duration_minutes, notes, created_at
		FROM pomodoro_sessions
		WHERE user_id = $1 AND course_id = $2
		ORDER BY created_at DESC
	`

	var models []PomodoroSessionModel
	if err := r.db.SelectContext(ctx, &models, query, userID, courseID); err != nil {
		return nil, err
	}

	sessions := make([]domain.PomodoroSession, len(models))
	for i, m := range models {
		sessions[i] = m.ToDomain()
	}

	return sessions, nil
}

func (r *pgRepository) GetSettings(ctx context.Context, userID int) (*domain.PomodoroSettings, error) {
	query := `
		SELECT user_id, study_presets, study_color, updated_at
		FROM pomodoro_settings
		WHERE user_id = $1
	`

	var model PomodoroSettingsModel
	if err := r.db.GetContext(ctx, &model, query, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Return default settings
			return &domain.PomodoroSettings{
				UserID:       userID,
				StudyPresets: []int{25, 45},
				StudyColor:   "gray",
			}, nil
		}
		return nil, err
	}

	settings := model.ToDomain()
	return &settings, nil
}

func (r *pgRepository) UpdateSettings(ctx context.Context, settings *domain.PomodoroSettings) error {
	query := `
		INSERT INTO pomodoro_settings (user_id, study_presets, study_color, updated_at)
		VALUES (:user_id, :study_presets, :study_color, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			study_presets = EXCLUDED.study_presets,
			study_color = EXCLUDED.study_color,
			updated_at = NOW()
	`

	presets := make(pq.Int64Array, len(settings.StudyPresets))
	for i, p := range settings.StudyPresets {
		presets[i] = int64(p)
	}

	model := PomodoroSettingsModel{
		UserID:       settings.UserID,
		StudyPresets: presets,
		StudyColor:   settings.StudyColor,
	}

	_, err := r.db.NamedExecContext(ctx, query, model)
	return err
}
