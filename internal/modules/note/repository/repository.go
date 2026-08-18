package repository

import (
	"context"
	"time"

	"github.com/M1ralai/notly-api/internal/modules/note/domain"
)

// NoteRepository defines all data-access operations for the note module.
// The HTTP and Service layers must depend ONLY on this interface.
type NoteRepository interface {
	// ── Core CRUD ──────────────────────────────────────────────────────────────
	Create(ctx context.Context, note *domain.Note) (*domain.Note, error)
	GetByID(ctx context.Context, id int) (*domain.Note, error)
	GetByUserID(ctx context.Context, userID int) ([]*domain.Note, error)
	Update(ctx context.Context, note *domain.Note) error
	Delete(ctx context.Context, id int) error

	// ── Sharing / Public Access ────────────────────────────────────────────────
	SetPublic(ctx context.Context, noteID int, isPublic bool, shareToken *string) error
	GetByShareToken(ctx context.Context, token string) (*domain.Note, error)

	// ── Attachments ───────────────────────────────────────────────────────────
	AddAttachment(ctx context.Context, att *domain.NoteAttachment) (*domain.NoteAttachment, error)
	GetAttachmentByID(ctx context.Context, attachmentID int) (*domain.NoteAttachment, error)
	GetAttachments(ctx context.Context, noteID int) ([]*domain.NoteAttachment, error)
	DeleteAttachment(ctx context.Context, attachmentID int) (*domain.NoteAttachment, error)

	// ── Collaborators ─────────────────────────────────────────────────────────
	AddCollaborator(ctx context.Context, col *domain.NoteCollaborator) (*domain.NoteCollaborator, error)
	GetCollaborators(ctx context.Context, noteID int) ([]*domain.NoteCollaborator, error)
	RemoveCollaborator(ctx context.Context, noteID, userID int) error
	IsCollaborator(ctx context.Context, noteID, userID int) (bool, error)

	// ── Sync ──────────────────────────────────────────────────────────────────
	GetUpdatedSince(ctx context.Context, userID int, since time.Time) ([]interface{}, error)
	GetDeletedSince(ctx context.Context, userID int, since time.Time) ([]int, error)
}
