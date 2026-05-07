package service

import (
	"context"
	"mime/multipart"

	"github.com/M1ralai/notly-api/internal/modules/note/dto"
)

// NoteService is the contract that the HTTP layer depends on.
// It returns DTOs – never domain models or raw DB rows.
type NoteService interface {
	// ── Core CRUD ──────────────────────────────────────────────────────────────
	Create(ctx context.Context, req *dto.CreateNoteRequest, userID int) (*dto.NoteOwnerResponse, error)
	GetByID(ctx context.Context, id, userID int) (*dto.NoteOwnerResponse, error)
	GetAll(ctx context.Context, userID int) ([]*dto.NoteOwnerResponse, error)
	Update(ctx context.Context, id int, req *dto.UpdateNoteRequest, userID int) (*dto.NoteOwnerResponse, error)
	Delete(ctx context.Context, id, userID int) error

	// ── Attachments (Kısım 03) ─────────────────────────────────────────────────
	UploadAttachment(ctx context.Context, noteID, userID int, file multipart.File, filename, contentType string, size int64) (*dto.AttachmentResponse, error)
	DeleteAttachment(ctx context.Context, attachmentID, userID int) error

	// ── Public Sharing (Kısım 04) ─────────────────────────────────────────────
	SetPublic(ctx context.Context, noteID, userID int, isPublic bool) (*dto.ShareTokenResponse, error)
	GetByShareToken(ctx context.Context, token string) (*dto.SharedNoteMinimalResponse, error)

	// ── Collaborators (Kısım 04) ──────────────────────────────────────────────
	AddCollaborator(ctx context.Context, noteID, ownerID int, req *dto.AddCollaboratorRequest) (*dto.CollaboratorResponse, error)
	RemoveCollaborator(ctx context.Context, noteID, ownerID, collaboratorUserID int) error
	GetCollaborators(ctx context.Context, noteID, userID int) ([]*dto.CollaboratorResponse, error)
}
