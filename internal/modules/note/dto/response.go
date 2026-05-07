package dto

import "time"

// ─────────────────────────────────────────────────────────────────────────────
// AttachmentResponse – safe for all callers
// ─────────────────────────────────────────────────────────────────────────────

type AttachmentResponse struct {
	ID            int       `json:"id"`
	FileURL       string    `json:"file_url"`
	FileType      string    `json:"file_type"`
	FileSizeBytes int64     `json:"file_size_bytes"`
	OriginalName  string    `json:"original_name"`
	CreatedAt     time.Time `json:"created_at"`
}

// ─────────────────────────────────────────────────────────────────────────────
// CollaboratorResponse – safe for all callers
// ─────────────────────────────────────────────────────────────────────────────

type CollaboratorResponse struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// ─────────────────────────────────────────────────────────────────────────────
// NoteOwnerResponse – returned to the note OWNER only.
// Contains all relational context (course, task, life-area).
// ─────────────────────────────────────────────────────────────────────────────

type NoteOwnerResponse struct {
	ID           int                    `json:"id"`
	UserID       int                    `json:"user_id"`
	Title        string                 `json:"title"`
	Content      string                 `json:"content"`
	IsPublic     bool                   `json:"is_public"`
	ShareToken   *string                `json:"share_token,omitempty"`
	ParentNoteID *int                   `json:"parent_note_id,omitempty"`
	CourseID     *int                   `json:"course_id,omitempty"`
	LifeAreaID   *int                   `json:"life_area_id,omitempty"`
	LinkedTaskID *int                   `json:"linked_task_id,omitempty"`
	Attachments  []AttachmentResponse   `json:"attachments"`
	Collaborators []CollaboratorResponse `json:"collaborators"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// ─────────────────────────────────────────────────────────────────────────────
// SharedNoteMinimalResponse – returned to GUESTS / COLLABORATORS.
// Personal context (course, task, life-area) is deliberately OMITTED.
// This is the privacy boundary enforced at the DTO layer (Kısım 05).
// ─────────────────────────────────────────────────────────────────────────────

type SharedNoteMinimalResponse struct {
	ID          int                  `json:"id"`
	Title       string               `json:"title"`
	Content     string               `json:"content"`
	Attachments []AttachmentResponse `json:"attachments"`
	CreatedAt   time.Time            `json:"created_at"`
}

// ─────────────────────────────────────────────────────────────────────────────
// ShareTokenResponse – returned when a public link is activated
// ─────────────────────────────────────────────────────────────────────────────

type ShareTokenResponse struct {
	ShareToken string `json:"share_token"`
	PublicURL  string `json:"public_url"`
}
