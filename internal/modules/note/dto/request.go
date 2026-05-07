package dto

// ─────────────────────────────────────────────────────────────────────────────
// Notes – Requests
// ─────────────────────────────────────────────────────────────────────────────

// CreateNoteRequest is the payload for POST /api/notes
type CreateNoteRequest struct {
	Title        string `json:"title"         validate:"required,max=512"`
	Content      string `json:"content"`
	CourseID     *int   `json:"course_id"`
	LifeAreaID   *int   `json:"life_area_id"`
	LinkedTaskID *int   `json:"linked_task_id"`
	ParentNoteID *int   `json:"parent_note_id"`
}

// UpdateNoteRequest is the payload for PUT /api/notes/{id}
type UpdateNoteRequest struct {
	Title   string `json:"title"   validate:"required,max=512"`
	Content string `json:"content"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Sharing – Requests
// ─────────────────────────────────────────────────────────────────────────────

// SetPublicRequest toggles note visibility
type SetPublicRequest struct {
	IsPublic bool `json:"is_public"`
}

// AddCollaboratorRequest adds a user to a note by their user ID
type AddCollaboratorRequest struct {
	UserID int    `json:"user_id" validate:"required,min=1"`
	Role   string `json:"role"    validate:"required,oneof=viewer editor"`
}
