package domain

import "time"

// Note represents a personal (or shared) note entity.
// No json/db tags – this is pure domain.
type Note struct {
	ID           int
	UserID       int
	Title        string
	Content      string
	IsPublic     bool
	ShareToken   *string
	ParentNoteID *int
	CourseID     *int
	LifeAreaID   *int
	LinkedTaskID *int
	CreatedAt    time.Time
	UpdatedAt    time.Time

	// Related entities (populated on demand)
	Attachments   []NoteAttachment
	Collaborators []NoteCollaborator
}

// NoteAttachment represents a media file attached to a note.
type NoteAttachment struct {
	ID           int
	NoteID       int
	FileURL      string
	ObjectKey    string // MinIO object key – needed for deletion
	FileType     string
	FileSizeBytes int64
	OriginalName string
	CreatedAt    time.Time
}

// NoteCollaborator represents a user who has been granted access to a note.
type NoteCollaborator struct {
	ID        int
	NoteID    int
	UserID    int
	Role      CollaboratorRole
	CreatedAt time.Time
}

// CollaboratorRole enumerates the access levels a collaborator can have.
type CollaboratorRole string

const (
	RoleViewer CollaboratorRole = "viewer"
	RoleEditor CollaboratorRole = "editor"
)

// IsValidRole confirms the role string is one of the known values.
func IsValidRole(r CollaboratorRole) bool {
	return r == RoleViewer || r == RoleEditor
}

// IsShared returns true when the note has been made public or has collaborators.
func (n *Note) IsShared() bool {
	return n.IsPublic || len(n.Collaborators) > 0
}

// HasAttachments is a convenience helper used in templates / serializers.
func (n *Note) HasAttachments() bool {
	return len(n.Attachments) > 0
}
