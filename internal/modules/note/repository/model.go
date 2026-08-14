package repository

import (
	"time"

	"github.com/M1ralai/notly-api/internal/modules/note/domain"
)

// ─────────────────────────────────────────────────────────────────────────────
// DB Models (carry db tags) – never exposed beyond the repository package
// ─────────────────────────────────────────────────────────────────────────────

type NoteModel struct {
	ID           int       `db:"id"`
	UserID       int       `db:"user_id"`
	Title        string    `db:"title"`
	Content      *string   `db:"content"`
	IsPublic     bool      `db:"is_public"`
	ShareToken   *string   `db:"share_token"`
	ParentNoteID *int      `db:"parent_note_id"`
	CourseID     *int      `db:"course_id"`
	LifeAreaID   *int      `db:"life_area_id"`
	LinkedTaskID *int      `db:"linked_task_id"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

func (m *NoteModel) ToDomain() *domain.Note {
	content := ""
	if m.Content != nil {
		content = *m.Content
	}
	return &domain.Note{
		ID:           m.ID,
		UserID:       m.UserID,
		Title:        m.Title,
		Content:      content,
		IsPublic:     m.IsPublic,
		ShareToken:   m.ShareToken,
		ParentNoteID: m.ParentNoteID,
		CourseID:     m.CourseID,
		LifeAreaID:   m.LifeAreaID,
		LinkedTaskID: m.LinkedTaskID,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

// ─────────────────────────────────────────────────────────────────────────────

type NoteAttachmentModel struct {
	ID            int       `db:"id"`
	NoteID        int       `db:"note_id"`
	FileURL       string    `db:"file_url"`
	ObjectKey     string    `db:"object_key"`
	FileType      string    `db:"file_type"`
	FileSizeBytes int64     `db:"file_size_bytes"`
	OriginalName  string    `db:"original_name"`
	CreatedAt     time.Time `db:"created_at"`
}

func (m *NoteAttachmentModel) ToDomain() *domain.NoteAttachment {
	return &domain.NoteAttachment{
		ID:            m.ID,
		NoteID:        m.NoteID,
		FileURL:       m.FileURL,
		ObjectKey:     m.ObjectKey,
		FileType:      m.FileType,
		FileSizeBytes: m.FileSizeBytes,
		OriginalName:  m.OriginalName,
		CreatedAt:     m.CreatedAt,
	}
}

// ─────────────────────────────────────────────────────────────────────────────

type NoteCollaboratorModel struct {
	ID        int                    `db:"id"`
	NoteID    int                    `db:"note_id"`
	UserID    int                    `db:"user_id"`
	Role      domain.CollaboratorRole `db:"role"`
	CreatedAt time.Time              `db:"created_at"`
}

func (m *NoteCollaboratorModel) ToDomain() *domain.NoteCollaborator {
	return &domain.NoteCollaborator{
		ID:        m.ID,
		NoteID:    m.NoteID,
		UserID:    m.UserID,
		Role:      m.Role,
		CreatedAt: m.CreatedAt,
	}
}
