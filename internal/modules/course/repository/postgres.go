package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/M1ralai/notly-api/internal/modules/course/domain"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type postgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) CourseRepository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(ctx context.Context, course *domain.Course) (*domain.Course, error) {
	query := `
		INSERT INTO courses (user_id, name, code, instructor, credits, semester_id, type, color, syllabus_url, final_grade, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at
	`

	now := time.Now()
	model := FromDomain(course)

	err := r.db.QueryRowxContext(
		ctx, query,
		model.UserID,
		model.Name,
		model.Code,
		model.Instructor,
		model.Credits,
		model.SemesterID,
		model.Type,
		model.Color,
		model.SyllabusURL,
		model.FinalGrade,
		model.IsActive,
		now,
		now,
	).Scan(&model.ID, &model.CreatedAt, &model.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return model.ToDomain(), nil
}

func (r *postgresRepository) GetByID(ctx context.Context, id int) (*domain.Course, error) {
	query := `
		SELECT id, user_id, name, code, instructor, credits, semester_id, type, color, syllabus_url, final_grade, is_active, created_at, updated_at
		FROM courses
		WHERE id = $1 AND deleted_at IS NULL
	`

	var model CourseModel
	err := r.db.GetContext(ctx, &model, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

func (r *postgresRepository) GetByUserID(ctx context.Context, userID int) ([]*domain.Course, error) {
	query := `
		SELECT id, user_id, name, code, instructor, credits, semester_id, type, color, syllabus_url, final_grade, is_active, created_at, updated_at
		FROM courses
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	var models []CourseModel
	err := r.db.SelectContext(ctx, &models, query, userID)
	if err != nil {
		return nil, err
	}

	courses := make([]*domain.Course, len(models))
	for i, m := range models {
		courses[i] = m.ToDomain()
	}

	return courses, nil
}

func (r *postgresRepository) GetActiveCourses(ctx context.Context, userID int) ([]*domain.Course, error) {
	query := `
		SELECT id, user_id, name, code, instructor, credits, semester_id, type, color, syllabus_url, final_grade, is_active, created_at, updated_at
		FROM courses
		WHERE user_id = $1 AND is_active = true AND deleted_at IS NULL
		ORDER BY name ASC
	`

	var models []CourseModel
	err := r.db.SelectContext(ctx, &models, query, userID)
	if err != nil {
		return nil, err
	}

	courses := make([]*domain.Course, len(models))
	for i, m := range models {
		courses[i] = m.ToDomain()
	}

	return courses, nil
}

func (r *postgresRepository) Update(ctx context.Context, course *domain.Course) error {
	query := `
		UPDATE courses
		SET name = $1, code = $2, instructor = $3, credits = $4, semester_id = $5, type = $6, color = $7, syllabus_url = $8, final_grade = $9, is_active = $10, updated_at = $11
		WHERE id = $12
	`

	model := FromDomain(course)
	_, err := r.db.ExecContext(
		ctx, query,
		model.Name,
		model.Code,
		model.Instructor,
		model.Credits,
		model.SemesterID,
		model.Type,
		model.Color,
		model.SyllabusURL,
		model.FinalGrade,
		model.IsActive,
		time.Now(),
		model.ID,
	)

	return err
}

func (r *postgresRepository) Delete(ctx context.Context, id int) error {
	query := `UPDATE courses SET deleted_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

func (r *postgresRepository) GetUpdatedSince(ctx context.Context, userID int, since time.Time) ([]*domain.Course, error) {
	query := `
		SELECT id, user_id, name, code, instructor, credits, semester_id, type, color, syllabus_url, final_grade, is_active, created_at, updated_at
		FROM courses
		WHERE user_id = $1 AND updated_at > $2 AND deleted_at IS NULL
		ORDER BY updated_at ASC
	`
	var models []CourseModel
	err := r.db.SelectContext(ctx, &models, query, userID, since)
	if err != nil {
		return nil, err
	}
	courses := make([]*domain.Course, len(models))
	for i, m := range models {
		courses[i] = m.ToDomain()
	}
	return courses, nil
}

func (r *postgresRepository) GetDeletedSince(ctx context.Context, userID int, since time.Time) ([]int, error) {
	query := `SELECT id FROM courses WHERE user_id = $1 AND deleted_at > $2`
	var ids []int
	if err := r.db.SelectContext(ctx, &ids, query, userID, since); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if ids == nil {
		ids = make([]int, 0)
	}
	return ids, nil
}

func (r *postgresRepository) CreateComponent(ctx context.Context, comp *domain.Component) (*domain.Component, error) {
	query := `
		INSERT INTO course_components (course_id, type, name, weight, max_score, achieved_score, due_date, is_completed, notes, display_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`

	now := time.Now()
	model := FromDomainComponent(comp)

	err := r.db.QueryRowxContext(
		ctx, query,
		model.CourseID,
		model.Type,
		model.Name,
		model.Weight,
		model.MaxScore,
		model.AchievedScore,
		model.DueDate,
		model.IsCompleted,
		model.Notes,
		model.DisplayOrder,
		now,
		now,
	).Scan(&model.ID, &model.CreatedAt, &model.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return model.ToDomain(), nil
}

func (r *postgresRepository) UpdateComponent(ctx context.Context, comp *domain.Component) error {
	query := `
		UPDATE course_components
		SET type = $1, name = $2, weight = $3, max_score = $4, achieved_score = $5, due_date = $6, completion_date = $7, is_completed = $8, notes = $9, display_order = $10, updated_at = $11
		WHERE id = $12
	`

	model := FromDomainComponent(comp)
	_, err := r.db.ExecContext(
		ctx, query,
		model.Type,
		model.Name,
		model.Weight,
		model.MaxScore,
		model.AchievedScore,
		model.DueDate,
		model.CompletionDate,
		model.IsCompleted,
		model.Notes,
		model.DisplayOrder,
		time.Now(),
		model.ID,
	)

	return err
}

func (r *postgresRepository) DeleteComponent(ctx context.Context, id int) error {
	query := `UPDATE course_components SET deleted_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

func (r *postgresRepository) GetComponents(ctx context.Context, courseID int) ([]*domain.Component, error) {
	query := `
		SELECT id, course_id, type, name, weight, max_score, achieved_score, due_date, completion_date, is_completed, notes, display_order, created_at, updated_at
		FROM course_components
		WHERE course_id = $1 AND deleted_at IS NULL
		ORDER BY display_order ASC, created_at ASC
	`

	var models []ComponentModel
	err := r.db.SelectContext(ctx, &models, query, courseID)
	if err != nil {
		return nil, err
	}

	components := make([]*domain.Component, len(models))
	for i, m := range models {
		components[i] = m.ToDomain()
	}

	return components, nil
}

func (r *postgresRepository) GetComponentByID(ctx context.Context, id int) (*domain.Component, error) {
	query := `
		SELECT id, course_id, type, name, weight, max_score, achieved_score, due_date, completion_date, is_completed, notes, display_order, created_at, updated_at
		FROM course_components
		WHERE id = $1 AND deleted_at IS NULL
	`

	var model ComponentModel
	err := r.db.GetContext(ctx, &model, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

func (r *postgresRepository) CreateSchedule(ctx context.Context, sched *domain.Schedule) (*domain.Schedule, error) {
	query := `
		INSERT INTO course_schedules (course_id, day_of_week, start_time, end_time, location, notifications_enabled, notification_type, reminder_time, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at
	`

	now := time.Now()
	model := FromDomainSchedule(sched)

	err := r.db.QueryRowxContext(
		ctx, query,
		model.CourseID,
		model.DayOfWeek,
		model.StartTime,
		model.EndTime,
		model.Location,
		model.NotificationsEnabled,
		model.NotificationType,
		model.ReminderTime,
		now,
	).Scan(&model.ID, &model.CreatedAt)

	if err != nil {
		return nil, err
	}

	return model.ToDomain(), nil
}

func (r *postgresRepository) UpdateSchedule(ctx context.Context, sched *domain.Schedule) error {
	query := `
		UPDATE course_schedules
		SET day_of_week = $1, start_time = $2, end_time = $3, location = $4, notifications_enabled = $5, notification_type = $6, reminder_time = $7
		WHERE id = $8
	`

	model := FromDomainSchedule(sched)
	_, err := r.db.ExecContext(
		ctx, query,
		model.DayOfWeek,
		model.StartTime,
		model.EndTime,
		model.Location,
		model.NotificationsEnabled,
		model.NotificationType,
		model.ReminderTime,
		model.ID,
	)

	return err
}

func (r *postgresRepository) DeleteSchedule(ctx context.Context, id int) error {
	query := `UPDATE course_schedules SET deleted_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

func (r *postgresRepository) GetSchedules(ctx context.Context, courseID int) ([]*domain.Schedule, error) {
	query := `
		SELECT
			id,
			course_id,
			day_of_week,
			start_time::text as start_time,
			end_time::text as end_time,
			location,
			notifications_enabled,
			notification_type,
			reminder_time,
			created_at
		FROM course_schedules
		WHERE course_id = $1 AND deleted_at IS NULL
		ORDER BY
			CASE day_of_week
				WHEN 'Monday' THEN 1
				WHEN 'Tuesday' THEN 2
				WHEN 'Wednesday' THEN 3
				WHEN 'Thursday' THEN 4
				WHEN 'Friday' THEN 5
				WHEN 'Saturday' THEN 6
				WHEN 'Sunday' THEN 7
				ELSE 8
			END,
			start_time ASC
	`

	var models []ScheduleModel
	err := r.db.SelectContext(ctx, &models, query, courseID)
	if err != nil {
		return nil, err
	}

	schedules := make([]*domain.Schedule, len(models))
	for i, m := range models {
		schedules[i] = m.ToDomain()
	}

	return schedules, nil
}

func (r *postgresRepository) GetScheduleByID(ctx context.Context, id int) (*domain.Schedule, error) {
	query := `
		SELECT
			id,
			course_id,
			day_of_week,
			start_time::text as start_time,
			end_time::text as end_time,
			location,
			notifications_enabled,
			notification_type,
			reminder_time,
			created_at
		FROM course_schedules
		WHERE id = $1 AND deleted_at IS NULL
	`

	var model ScheduleModel
	err := r.db.GetContext(ctx, &model, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

const resourceSelectColumns = `
	id, course_id, component_id, title, type, url, file_path, description,
	tags, is_primary, file_size_bytes, mime_type, created_at, updated_at`

func (r *postgresRepository) CreateResource(ctx context.Context, resource *domain.Resource) (*domain.Resource, error) {
	query := `
		INSERT INTO course_resources (
			course_id, component_id, title, type, url, file_path, description,
			tags, is_primary, file_size_bytes, mime_type, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at
	`

	now := time.Now()
	model := FromDomainResource(resource)
	err := r.db.QueryRowxContext(
		ctx,
		query,
		model.CourseID,
		model.ComponentID,
		model.Title,
		model.Type,
		model.URL,
		model.FilePath,
		model.Description,
		pq.Array([]string(model.Tags)),
		model.IsPrimary,
		model.FileSizeBytes,
		model.MimeType,
		now,
		now,
	).Scan(&model.ID, &model.CreatedAt, &model.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return model.ToDomain(), nil
}

func (r *postgresRepository) UpdateResource(ctx context.Context, resource *domain.Resource) error {
	query := `
		UPDATE course_resources
		SET component_id = $1, title = $2, type = $3, url = $4, file_path = $5,
		    description = $6, tags = $7, is_primary = $8, file_size_bytes = $9,
		    mime_type = $10, updated_at = $11
		WHERE id = $12
	`

	model := FromDomainResource(resource)
	_, err := r.db.ExecContext(
		ctx,
		query,
		model.ComponentID,
		model.Title,
		model.Type,
		model.URL,
		model.FilePath,
		model.Description,
		pq.Array([]string(model.Tags)),
		model.IsPrimary,
		model.FileSizeBytes,
		model.MimeType,
		time.Now(),
		model.ID,
	)
	return err
}

func (r *postgresRepository) DeleteResource(ctx context.Context, id int) (*domain.Resource, error) {
	var model ResourceModel
	err := r.db.GetContext(
		ctx,
		&model,
		`DELETE FROM course_resources WHERE id = $1 RETURNING `+resourceSelectColumns,
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *postgresRepository) GetResources(ctx context.Context, courseID int) ([]*domain.Resource, error) {
	var models []ResourceModel
	err := r.db.SelectContext(
		ctx,
		&models,
		`SELECT `+resourceSelectColumns+`
		 FROM course_resources
		 WHERE course_id = $1
		 ORDER BY is_primary DESC, created_at DESC`,
		courseID,
	)
	if err != nil {
		return nil, err
	}

	resources := make([]*domain.Resource, len(models))
	for i, m := range models {
		resources[i] = m.ToDomain()
	}
	return resources, nil
}

func (r *postgresRepository) GetResourceByID(ctx context.Context, id int) (*domain.Resource, error) {
	var model ResourceModel
	err := r.db.GetContext(
		ctx,
		&model,
		`SELECT `+resourceSelectColumns+` FROM course_resources WHERE id = $1`,
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return model.ToDomain(), nil
}
