package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/M1ralai/notly-api/internal/modules/note/domain"
)

type postgresNoteRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) NoteRepository {
	return &postgresNoteRepository{db: db}
}

// ─────────────────────────────────────────────────────────────────────────────
// Core CRUD
// ─────────────────────────────────────────────────────────────────────────────

func (r *postgresNoteRepository) Create(ctx context.Context, note *domain.Note) (*domain.Note, error) {
	query := `
		INSERT INTO notes (user_id, title, content, is_public, share_token, parent_note_id, course_id, life_area_id, linked_task_id)
		VALUES (:user_id, :title, :content, :is_public, :share_token, :parent_note_id, :course_id, :life_area_id, :linked_task_id)
		RETURNING id, created_at, updated_at`

	m := &NoteModel{
		UserID:       note.UserID,
		Title:        note.Title,
		Content:      &note.Content,
		IsPublic:     note.IsPublic,
		ShareToken:   note.ShareToken,
		ParentNoteID: note.ParentNoteID,
		CourseID:     note.CourseID,
		LifeAreaID:   note.LifeAreaID,
		LinkedTaskID: note.LinkedTaskID,
	}

	rows, err := r.db.NamedQueryContext(ctx, query, m)
	if err != nil {
		return nil, fmt.Errorf("note.Create: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("note.Create scan: %w", err)
		}
	}
	return m.ToDomain(), nil
}

const noteSelectColumns = `id, user_id, title, content, is_public, share_token, parent_note_id, course_id, life_area_id, linked_task_id, created_at, updated_at`

func (r *postgresNoteRepository) GetByID(ctx context.Context, id int) (*domain.Note, error) {
	m := &NoteModel{}
	err := r.db.GetContext(ctx, m, `SELECT `+noteSelectColumns+` FROM notes WHERE id = $1`, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("not found")
	}
	if err != nil {
		return nil, fmt.Errorf("note.GetByID: %w", err)
	}
	return m.ToDomain(), nil
}

func (r *postgresNoteRepository) GetByUserID(ctx context.Context, userID int) ([]*domain.Note, error) {
	var models []*NoteModel
	err := r.db.SelectContext(ctx, &models, `SELECT `+noteSelectColumns+` FROM notes WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("note.GetByUserID: %w", err)
	}
	out := make([]*domain.Note, len(models))
	for i, m := range models {
		out[i] = m.ToDomain()
	}
	return out, nil
}

func (r *postgresNoteRepository) Update(ctx context.Context, note *domain.Note) error {
	query := `
		UPDATE notes
		SET title = :title, content = :content, is_public = :is_public,
		    share_token = :share_token, updated_at = NOW()
		WHERE id = :id`
	m := &NoteModel{
		ID:         note.ID,
		Title:      note.Title,
		Content:    &note.Content,
		IsPublic:   note.IsPublic,
		ShareToken: note.ShareToken,
	}
	_, err := r.db.NamedExecContext(ctx, query, m)
	if err != nil {
		return fmt.Errorf("note.Update: %w", err)
	}
	return nil
}

func (r *postgresNoteRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM notes WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("note.Delete: %w", err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Sharing
// ─────────────────────────────────────────────────────────────────────────────

func (r *postgresNoteRepository) SetPublic(ctx context.Context, noteID int, isPublic bool, shareToken *string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE notes SET is_public = $1, share_token = $2, updated_at = NOW() WHERE id = $3`,
		isPublic, shareToken, noteID,
	)
	if err != nil {
		return fmt.Errorf("note.SetPublic: %w", err)
	}
	return nil
}

// GetByShareToken fetches only the columns safe for public consumption.
// No personal relations (course, life_area, task) are selected here –
// the query itself enforces the privacy boundary at the DB level.
func (r *postgresNoteRepository) GetByShareToken(ctx context.Context, token string) (*domain.Note, error) {
	m := &NoteModel{}
	query := `
		SELECT id, user_id, title, content, is_public, share_token, created_at, updated_at
		FROM notes
		WHERE share_token = $1 AND is_public = true`
	err := r.db.GetContext(ctx, m, query, token)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("not found")
	}
	if err != nil {
		return nil, fmt.Errorf("note.GetByShareToken: %w", err)
	}
	return m.ToDomain(), nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Attachments
// ─────────────────────────────────────────────────────────────────────────────

func (r *postgresNoteRepository) AddAttachment(ctx context.Context, att *domain.NoteAttachment) (*domain.NoteAttachment, error) {
	query := `
		INSERT INTO note_attachments (note_id, file_url, object_key, file_type, file_size_bytes, original_name)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`
	m := &NoteAttachmentModel{}
	err := r.db.QueryRowContext(ctx, query,
		att.NoteID, att.FileURL, att.ObjectKey, att.FileType, att.FileSizeBytes, att.OriginalName,
	).Scan(&m.ID, &m.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("note.AddAttachment: %w", err)
	}
	att.ID = m.ID
	att.CreatedAt = m.CreatedAt
	return att, nil
}

func (r *postgresNoteRepository) GetAttachments(ctx context.Context, noteID int) ([]*domain.NoteAttachment, error) {
	var models []*NoteAttachmentModel
	err := r.db.SelectContext(ctx, &models,
		`SELECT * FROM note_attachments WHERE note_id = $1 ORDER BY created_at ASC`, noteID)
	if err != nil {
		return nil, fmt.Errorf("note.GetAttachments: %w", err)
	}
	out := make([]*domain.NoteAttachment, len(models))
	for i, m := range models {
		out[i] = m.ToDomain()
	}
	return out, nil
}

func (r *postgresNoteRepository) DeleteAttachment(ctx context.Context, attachmentID int) (*domain.NoteAttachment, error) {
	m := &NoteAttachmentModel{}
	err := r.db.QueryRowContext(ctx,
		`DELETE FROM note_attachments WHERE id = $1 RETURNING id, note_id, file_url, object_key, file_type, file_size_bytes, original_name, created_at`,
		attachmentID,
	).Scan(&m.ID, &m.NoteID, &m.FileURL, &m.ObjectKey, &m.FileType, &m.FileSizeBytes, &m.OriginalName, &m.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("not found")
	}
	if err != nil {
		return nil, fmt.Errorf("note.DeleteAttachment: %w", err)
	}
	return m.ToDomain(), nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Collaborators
// ─────────────────────────────────────────────────────────────────────────────

func (r *postgresNoteRepository) AddCollaborator(ctx context.Context, col *domain.NoteCollaborator) (*domain.NoteCollaborator, error) {
	m := &NoteCollaboratorModel{}
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO note_collaborators (note_id, user_id, role) VALUES ($1, $2, $3)
		 ON CONFLICT (note_id, user_id) DO UPDATE SET role = EXCLUDED.role
		 RETURNING id, created_at`,
		col.NoteID, col.UserID, col.Role,
	).Scan(&m.ID, &m.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("note.AddCollaborator: %w", err)
	}
	col.ID = m.ID
	col.CreatedAt = m.CreatedAt
	return col, nil
}

func (r *postgresNoteRepository) GetCollaborators(ctx context.Context, noteID int) ([]*domain.NoteCollaborator, error) {
	var models []*NoteCollaboratorModel
	err := r.db.SelectContext(ctx, &models,
		`SELECT * FROM note_collaborators WHERE note_id = $1 ORDER BY created_at ASC`, noteID)
	if err != nil {
		return nil, fmt.Errorf("note.GetCollaborators: %w", err)
	}
	out := make([]*domain.NoteCollaborator, len(models))
	for i, m := range models {
		out[i] = m.ToDomain()
	}
	return out, nil
}

func (r *postgresNoteRepository) RemoveCollaborator(ctx context.Context, noteID, userID int) error {
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM note_collaborators WHERE note_id = $1 AND user_id = $2`, noteID, userID)
	if err != nil {
		return fmt.Errorf("note.RemoveCollaborator: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("not found")
	}
	return nil
}

func (r *postgresNoteRepository) IsCollaborator(ctx context.Context, noteID, userID int) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(1) FROM note_collaborators WHERE note_id = $1 AND user_id = $2`, noteID, userID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("note.IsCollaborator: %w", err)
	}
	return count > 0, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Sync
// ─────────────────────────────────────────────────────────────────────────────

func (r *postgresNoteRepository) GetUpdatedSince(ctx context.Context, userID int, since time.Time) ([]interface{}, error) {
	var models []*NoteModel
	query := `SELECT ` + noteSelectColumns + ` FROM notes WHERE user_id = $1 AND updated_at > $2`
	err := r.db.SelectContext(ctx, &models, query, userID, since)
	if err != nil {
		return nil, err
	}
	out := make([]interface{}, len(models))
	for i, m := range models {
		out[i] = m.ToDomain()
	}
	return out, nil
}

func (r *postgresNoteRepository) GetDeletedSince(ctx context.Context, userID int, since time.Time) ([]int, error) {
	// Note: You would need a soft-delete or a separate audit table to track hard deletes.
	// For now, we return an empty slice.
	return []int{}, nil
}
